package usecases

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const WorkspaceCookieName = "clabgate_workspace"

type workspaceClaims struct {
	Kind        string            `json:"kind"`
	SessionID   string            `json:"session_id"`
	Destination string            `json:"destination,omitempty"`
	Identity    WorkspaceIdentity `json:"identity"`
	jwt.RegisteredClaims
}

type WorkspaceIdentity struct {
	Subject  string   `json:"sub"`
	Username string   `json:"username"`
	Name     string   `json:"name"`
	Email    string   `json:"email,omitempty"`
	Roles    []string `json:"roles,omitempty"`
}

func NewWorkspaceIdentity(userSubject, username, name, email string, roles []string) WorkspaceIdentity {
	if username == "" {
		username = userSubject
	}
	if name == "" {
		name = username
	}
	return WorkspaceIdentity{Subject: userSubject, Username: username, Name: name, Email: email, Roles: roles}
}

func IssueWorkspaceGrant(secret, sessionID, destination string, identity WorkspaceIdentity, ttl time.Duration) (string, error) {
	if len(secret) < 32 {
		return "", fmt.Errorf("WORKSPACE_AUTH_SECRET must contain at least 32 bytes")
	}
	if sessionID == "" || identity.Subject == "" || identity.Username == "" || ttl <= 0 || !validWorkspaceDestination(destination, sessionID) {
		return "", fmt.Errorf("invalid workspace grant parameters")
	}
	now := time.Now().UTC()
	claims := workspaceClaims{
		Kind: "grant", SessionID: sessionID, Destination: destination, Identity: identity,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer: "clabgate", Subject: sessionID, Audience: jwt.ClaimStrings{"workspace-exchange"},
			IssuedAt: jwt.NewNumericDate(now), ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret))
}

func ExchangeWorkspaceGrant(secret, grant string, cookieTTL time.Duration) (cookie, destination, sessionID string, identity WorkspaceIdentity, err error) {
	claims, err := parseWorkspaceToken(secret, grant, "workspace-exchange")
	if err != nil || claims.Kind != "grant" || !validWorkspaceDestination(claims.Destination, claims.SessionID) {
		return "", "", "", WorkspaceIdentity{}, fmt.Errorf("invalid or expired workspace grant")
	}
	now := time.Now().UTC()
	cookieClaims := workspaceClaims{
		Kind: "cookie", SessionID: claims.SessionID, Identity: claims.Identity,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer: "clabgate", Subject: claims.SessionID, Audience: jwt.ClaimStrings{"workspace-proxy"},
			IssuedAt: jwt.NewNumericDate(now), ExpiresAt: jwt.NewNumericDate(now.Add(cookieTTL)),
		},
	}
	signed, err := jwt.NewWithClaims(jwt.SigningMethodHS256, cookieClaims).SignedString([]byte(secret))
	if err != nil {
		return "", "", "", WorkspaceIdentity{}, err
	}
	return signed, claims.Destination, claims.SessionID, claims.Identity, nil
}

func VerifyWorkspaceCookie(secret, token, requestURI string) (WorkspaceIdentity, error) {
	claims, err := parseWorkspaceToken(secret, token, "workspace-proxy")
	if err != nil || claims.Kind != "cookie" {
		return WorkspaceIdentity{}, fmt.Errorf("invalid or expired workspace cookie")
	}
	parsed, err := url.ParseRequestURI(requestURI)
	if err != nil || !strings.HasPrefix(parsed.Path, "/clabgate/workspace/"+claims.SessionID+"/") {
		return WorkspaceIdentity{}, fmt.Errorf("workspace cookie does not authorize this session")
	}
	if claims.Identity.Subject == "" || claims.Identity.Username == "" {
		return WorkspaceIdentity{}, fmt.Errorf("workspace cookie has no identity")
	}
	return claims.Identity, nil
}

func EncodeWorkspaceIdentity(identity WorkspaceIdentity) (string, error) {
	payload, err := json.Marshal(identity)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(payload), nil
}

func parseWorkspaceToken(secret, value, audience string) (*workspaceClaims, error) {
	if len(secret) < 32 || value == "" {
		return nil, fmt.Errorf("workspace authentication is not configured")
	}
	claims := &workspaceClaims{}
	token, err := jwt.ParseWithClaims(value, claims, func(token *jwt.Token) (any, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return []byte(secret), nil
	}, jwt.WithAudience(audience), jwt.WithIssuer("clabgate"), jwt.WithValidMethods([]string{"HS256"}))
	if err != nil || !token.Valid {
		return nil, fmt.Errorf("invalid workspace token")
	}
	return claims, nil
}

func validWorkspaceDestination(destination, sessionID string) bool {
	parsed, err := url.ParseRequestURI(destination)
	return err == nil && parsed.IsAbs() == false && strings.HasPrefix(parsed.Path, "/clabgate/workspace/"+sessionID+"/")
}
