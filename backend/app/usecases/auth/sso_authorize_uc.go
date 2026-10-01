package auth

import (
	"errors"
	"fmt"

	"github.com/cms-lab-core/cms-labs-api/shared/jsonrpc"

	"github.com/rs/zerolog"

	"github.com/cms-lab-core/cms-labs-api/backend/app/models"
	"github.com/cms-lab-core/cms-labs-api/backend/app/queries"
	"github.com/google/uuid"
)

type SSOAuthorizeInputDTO struct {
	UserID              uint   `json:"user_id"`
	ClientID            string `json:"client_id"`
	RedirectUri         string `json:"redirect_uri"`
	ResponseType        string `json:"response_type"`
	Scope               string `json:"scope"`
	Path                string `json:"path"`
	Nonce               string `json:"nonce"`
	State               string `json:"state"`
	Extra               string `json:"extra"`
	CodeChallenge       string `json:"code_challenge"`
	CodeChallengeMethod string `json:"code_challenge_method"`
}

type SSOAuthorizeRequest struct {
	JSONRPC string               `json:"jsonrpc" default:"2.0" required:"true"`
	Method  string               `json:"method" default:"sso.authorize" required:"true"`
	Params  SSOAuthorizeInputDTO `json:"params,omitempty"`
	ID      string               `json:"id,omitempty" default:"1" required:"true"`
}

type SSOAuthorizeOutputDTO struct {
	RedirectUri string `json:"redirect_uri"`
	Code        string `json:"code"`
	Scope       string `json:"scope"`
	Application string `json:"application"`
	ClientID    string `json:"client_id"`
	Path        string `json:"path"`
	State       string `json:"state"`
	Nonce       string `json:"nonce"`
	Extra       string `json:"extra"`
}

type SSOAuthorizeResponse struct {
	JSONRPC string                `json:"jsonrpc" default:"2.0" required:"true"`
	Result  SSOAuthorizeOutputDTO `json:"result,omitempty"`
	Error   interface{}           `json:"error,omitempty"`
	ID      string                `json:"id,omitempty" default:"1" required:"true"`
}

const (
	SSOAuthorizeResponseType = "code"
)

var (
	SSOAuthorizeUCRoles = jsonrpc.NewRpcError(
		"not_access_sso_with_current_role",
		"you cannot use SSO with this role",
	)
)

type SSOAuthorizeUC struct {
	TokenAttemptQueries *queries.TokenAttemptQueries
	ServerQueries       *queries.ServerQueries
	RoleQueries         *queries.RoleQueries
	*zerolog.Logger
}

func (u *SSOAuthorizeUC) Execute(inputDTO SSOAuthorizeInputDTO) (*SSOAuthorizeOutputDTO, error) {
	if err := u.validate(inputDTO); err != nil {
		return nil, err
	}
	attemptState, attemptNonce := inputDTO.State, inputDTO.Nonce
	if attemptState == "" {
		attemptState = uuid.New().String()
	}
	if attemptNonce == "" {
		attemptNonce = uuid.New().String()
	}
	u.Logger.Info().Msg(fmt.Sprintf("authorize sso authorize with clientID %s and userID %v", inputDTO.ClientID, inputDTO.UserID))
	server, err := u.ServerQueries.GetByClientId(inputDTO.ClientID)
	if err != nil {
		return nil, err
	}
	if !server.IsActive {
		u.Logger.Info().Msg(fmt.Sprintf("server is not active sso authorize with clientID %s and userID %v", inputDTO.ClientID, inputDTO.UserID))
		return nil, errors.New("server is not active")
	}
	if !server.HasPrefixUrl(inputDTO.RedirectUri) {
		u.Logger.Info().Msg(fmt.Sprintf("invalid redirect_uri sso authorize with clientID %s and userID %v", inputDTO.ClientID, inputDTO.UserID))
		return nil, errors.New("invalid redirect_uri")
	}
	err = u.validateRoles(server.ID, inputDTO.UserID)
	if err != nil {
		return nil, err
	}
	entityCreate := &models.TokenAttempt{}
	entityCreate.UserID = &inputDTO.UserID
	entityCreate.ServerID = &server.ID
	entityCreate.Token = ""
	entityCreate.State = attemptState
	entityCreate.Nonce = attemptNonce
	entityCreate.AuthorizationCode = uuid.New().String()
	if inputDTO.CodeChallenge != "" {
		entityCreate.CodeChallenge = &inputDTO.CodeChallenge
	}
	if inputDTO.CodeChallengeMethod != "" {
		entityCreate.CodeChallengeMethod = &inputDTO.CodeChallengeMethod
	}
	err = u.TokenAttemptQueries.Upsert(entityCreate)
	if err != nil {
		return nil, err
	}
	u.Logger.Info().Msg(fmt.Sprintf("response second factor sso authorize with clientID %s and userID %v", inputDTO.ClientID, inputDTO.UserID))
	return &SSOAuthorizeOutputDTO{
		RedirectUri: inputDTO.RedirectUri,
		Code:        entityCreate.AuthorizationCode,
		Scope:       inputDTO.Scope,
		Application: inputDTO.ClientID,
		ClientID:    inputDTO.ClientID,
		Path:        inputDTO.Path,
		State:       attemptState,
		Nonce:       attemptNonce,
		Extra:       inputDTO.Extra,
	}, nil

}

func (u *SSOAuthorizeUC) validateRoles(serverID uint, userID uint) error {
	serverRolesMap, err := u.RoleQueries.GetByRelationServerIds([]uint{serverID})
	if err != nil {
		return err
	}
	serverRoles, ok := serverRolesMap[serverID]
	if !ok {
		u.Logger.Info().Msg(fmt.Sprintf("server sso without whitelist roles serverID %d", serverID))
		return nil
	}
	userRolesMap, err := u.RoleQueries.GetByRelationUsersIds([]uint{userID})
	if err != nil {
		return err
	}
	userRoles, ok := userRolesMap[userID]
	if !ok {
		u.Logger.Info().Msg(fmt.Sprintf("client without roles userID %d", userID))
		return SSOAuthorizeUCRoles
	}
	for _, serverRoleId := range serverRoles {
		for _, userRoleId := range userRoles {
			if userRoleId == serverRoleId {
				u.Logger.Info().Msg(fmt.Sprintf("user and server roles match roleID: %d", userRoleId))
				return nil
			}
		}
	}
	u.Logger.Info().Msg(fmt.Sprintf("userID: %d and server roles not match", userID))
	return SSOAuthorizeUCRoles
}

func (u *SSOAuthorizeUC) validate(inputDTO SSOAuthorizeInputDTO) error {
	if inputDTO.ResponseType == SSOAuthorizeResponseType && inputDTO.ClientID != "" && inputDTO.RedirectUri != "" && inputDTO.Scope != "" {
		return nil
	}
	return errors.New("Invalid params response_type")
}
