package auth

import (
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/cms-lab-core/cms-labs-api/shared/k8s_utils"

	"github.com/cms-lab-core/cms-labs-api/shared/cms_client"

	"github.com/cms-lab-core/cms-labs-api/backend/app/models"
	"github.com/cms-lab-core/cms-labs-api/backend/app/queries"
	"github.com/cms-lab-core/cms-labs-api/backend/pkg/configs"
	"golang.org/x/crypto/bcrypt"
)

type TokenManager struct {
	UserQueries         *queries.UserQueries
	RoleQueries         *queries.RoleQueries
	ServerQueries       *queries.ServerQueries
	UserPasswordQueries *queries.UserPasswordQueries
	TokenAttemptQueries *queries.TokenAttemptQueries
}

type RenewManagerInputDTO struct {
	RefreshToken string `json:"refresh_token"`
}

type RenewManagerRefreshRequest struct {
	JSONRPC string               `json:"jsonrpc" default:"2.0" validate:"required"`
	Method  string               `json:"method" default:"user.token_refresh" validate:"required"`
	Params  RenewManagerInputDTO `json:"params,omitempty"`
	ID      string               `json:"id,omitempty" default:"1" validate:"required"`
}

type RenewManagerCredentialsInputDTO struct {
	Email      string `json:"email" required:"true"`
	Password   string `json:"password" required:"true"`
	ProviderId uint   `json:"provider_id"`
}

type RenewManagerCredentialsRequest struct {
	JSONRPC string                          `json:"jsonrpc" default:"2.0" validate:"required"`
	Method  string                          `json:"method" default:"user.login" validate:"required"`
	Params  RenewManagerCredentialsInputDTO `json:"params,omitempty"`
	ID      string                          `json:"id,omitempty" default:"1" validate:"required"`
}

// NewJWTByCredentials генерирует новый JWT по логину/паролю
func (m *TokenManager) NewJWTByCredentials(issID string, email string, password string) (*cms_client.SSOToken, error) {
	if email == "" {
		return nil, queries.IncorrectPassword
	}
	if password == "" {
		return nil, queries.IncorrectPassword
	}

	entity, err := m.UserQueries.GetByEmail(email)
	if err != nil {
		return nil, queries.IncorrectPassword
	}
	creds, err := m.UserPasswordQueries.Get(entity.ID)
	if err != nil {
		return nil, queries.IncorrectPassword
	}
	err = bcrypt.CompareHashAndPassword([]byte(creds.HashPassword), []byte(password))
	if err != nil {
		return nil, queries.IncorrectPassword
	}
	return m.NewJWTByUserId(issID, entity.ID, nil, nil)
}

func (m *TokenManager) NewJWTByLaunchID(issID string, launchID string) (*cms_client.SSOToken, error) {
	invalidCreds := errors.New("LTI process is incorrect")
	if launchID == "" {
		return nil, invalidCreds
	}
	entity, err := m.UserQueries.GetByLaunchID(launchID)
	if err != nil {
		return nil, invalidCreds
	}
	return m.NewJWTByUserId(issID, entity.ID, nil, nil)
}

func (m *TokenManager) NewJWTByUserId(
	issID string,
	userId uint,
	serverID *uint,
	attempt *models.TokenAttempt,
) (*cms_client.SSOToken, error) {
	attemptState, attemptNonce := "", ""
	if attempt != nil {
		attemptState, attemptNonce = attempt.State, attempt.Nonce
	}
	userModel, err := m.UserQueries.Get(userId)
	if err != nil {
		return nil, err
	}
	var serverModel models.Server
	if serverID != nil {
		serverModel, _ = m.ServerQueries.Get(*serverID)
	}
	rolesJWT, err := m.JWTRolesByUserId(userModel.ID)
	if err != nil {
		return nil, err
	}
	tokens, jti, err := GenerateNewTokens(
		&cms_client.SSOTokenPublicData{
			Iss:          issID,
			Sub:          fmt.Sprintf("%d", userModel.ID),
			Aud:          serverModel.ClientID,
			Azp:          serverModel.ClientID,
			Nonce:        attemptNonce,
			Email:        userModel.Email,
			Username:     k8s_utils.NormalizeK8SEntityName(k8s_utils.UsernameByEmail(userModel.Email)),
			Name:         userModel.Name,
			ServerID:     serverID,
			Roles:        rolesJWT,
			LastLaunchId: userModel.LastLaunchID,
		}, attemptState)
	if err != nil {
		return nil, err
	}
	refreshModel := &models.TokenAttempt{}
	refreshModel.UserID = &userId
	refreshModel.ServerID = serverID
	refreshModel.State = attemptState
	refreshModel.Nonce = attemptNonce
	refreshModel.Token = jti
	if err = m.TokenAttemptQueries.Upsert(refreshModel); err != nil {
		return nil, err
	}
	return tokens, nil
}

func (m *TokenManager) JWTRolesByUserId(
	userId uint,
) ([]string, error) {
	rolesJWT := make([]string, 0)
	var isAdmin, isStudent, isInstructor bool
	roles, err := m.RoleQueries.GetRolesByUserId(userId)
	if err != nil {
		return rolesJWT, err
	}
	// Delete all default roles
	for _, role := range roles {
		if role.Code == cms_client.SSOUsersRoleStudent {
			isStudent = true
			continue
		}
		if role.Code == cms_client.SSOUsersRoleAdmin {
			isAdmin = true
			continue
		}
		if role.Code == cms_client.SSOUsersRoleInstructor {
			isInstructor = true
			continue
		}
		rolesJWT = append(rolesJWT, role.Code)
	}
	// The administrator applies only if he is a instructor
	if isAdmin && isInstructor {
		rolesJWT = slices.Insert(rolesJWT, 0, cms_client.SSOUsersRoleAdmin)
		return rolesJWT, nil
	}
	if isInstructor {
		rolesJWT = slices.Insert(rolesJWT, 0, cms_client.SSOUsersRoleInstructor)
		return rolesJWT, nil
	}
	if isStudent {
		rolesJWT = slices.Insert(rolesJWT, 0, cms_client.SSOUsersRoleStudent)
		return rolesJWT, nil
	}
	return rolesJWT, nil
}

func ExpiresRefreshCookie() time.Time {
	return time.Now().Add(configs.AppConfig.JWT.RefreshKey.Expire)
}
