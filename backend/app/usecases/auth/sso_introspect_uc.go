package auth

import (
	"fmt"
	"time"

	"github.com/cms-lab-core/cms-labs-api/shared/cms_client"

	"github.com/cms-lab-core/cms-labs-api/backend/app/queries"
	"github.com/cms-lab-core/cms-labs-api/shared/jsonrpc"
)

type SSOIntrospectInputDTO struct {
	Token string `json:"token"`
}

type SSOIntrospectUC struct {
	TokenAttemptQueries *queries.TokenAttemptQueries
	ServerQueries       *queries.ServerQueries
}

func (u *SSOIntrospectUC) Execute(inputDTO SSOIntrospectInputDTO) (*SSOTokenIntrospect, error) {
	entity, err := u.ByRefresh(inputDTO)
	if err == nil {
		return entity, err
	}
	entity, err = u.ByAccess(inputDTO)
	return entity, err
}

func (u *SSOIntrospectUC) ByRefresh(inputDTO SSOIntrospectInputDTO) (*SSOTokenIntrospect, error) {
	introspect := &SSOTokenIntrospect{}
	introspect.Active = true
	introspect.Scope = []string{"default"}
	introspect.TokenType = "refresh"
	expiresRefreshToken, jti, err := ParseRefreshToken(inputDTO.Token)
	if err != nil {
		return nil, err
	}
	now := time.Now().Unix()
	introspect.Exp = now
	if now >= expiresRefreshToken {
		introspect.Active = false
	}
	attempt, err := u.TokenAttemptQueries.GetByTokenId(jti)
	if err != nil {
		return introspect, err
	}
	introspect.Sub = fmt.Sprintf("%v", attempt.UserID)
	introspect.Iat = attempt.CreatedAt.Unix()
	if attempt.ServerID == nil {
		return introspect, nil
	}
	server, err := u.ServerQueries.Get(*attempt.ServerID)
	if err != nil {
		return nil, err
	}
	introspect.Username = server.Name
	introspect.ClientID = server.ClientID
	if !server.IsActive {
		introspect.Active = false
	}
	return introspect, nil
}

func (u *SSOIntrospectUC) ByAccess(inputDTO SSOIntrospectInputDTO) (*SSOTokenIntrospect, error) {
	introspect := &SSOTokenIntrospect{}
	introspect.TokenType = "access"
	introspect.Active = true
	introspect.Scope = []string{"default"}
	token, err := verifyToken(inputDTO.Token)
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, jsonrpc.NewRpcError("invalid_token", "token is invalid")
	}
	tokenData, err := cms_client.SSODecodeToken(token)
	if err != nil {
		return nil, jsonrpc.NewRpcError("invalid_token", "token is invalid")
	}
	userID := tokenData.UserID()
	introspect.ClientID = tokenData.Aud
	introspect.Sub = tokenData.Sub
	introspect.Exp = tokenData.Exp
	attempt, err := u.TokenAttemptQueries.GetByParams(&userID, tokenData.ServerID, nil)
	if err != nil {
		introspect.Active = false
		return introspect, nil
	}
	introspect.Iat = tokenData.Iat
	if attempt.ServerID == nil {
		return introspect, nil
	}
	server, err := u.ServerQueries.Get(*attempt.ServerID)
	if err != nil {
		return introspect, nil
	}
	introspect.Username = server.Name
	introspect.ClientID = server.ClientID
	if !server.IsActive {
		introspect.Active = false
	}
	return introspect, nil
}
