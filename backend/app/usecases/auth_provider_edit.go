package usecases

import (
	"github.com/cms-lab-core/cms-labs-api/backend/app/models"
	"github.com/cms-lab-core/cms-labs-api/backend/app/queries/lti_query"
)

type AuthProviderEditUC struct {
	AuthProviderQueries *lti_query.AuthProviderQueries
}

type AuthProviderEditInputDTO struct {
	ID            uint    `json:"id"`
	Type          string  `json:"type" validate:"required"`
	ClientID      string  `json:"client_id"`
	DeploymentID  string  `json:"deployment_id"`
	BaseURI       string  `json:"base_uri" validate:"required"`
	AuthTokenURI  string  `json:"auth_token_uri"`
	AuthLoginURI  string  `json:"auth_login_uri"`
	KeySetURI     string  `json:"key_set_uri"`
	TargetLinkURI string  `json:"target_link_uri"`
	Name          string  `json:"name" validate:"required"`
	SSOURL        *string `json:"sso_url"`
}

type AuthProviderEditRequest struct {
	JSONRPC string                   `json:"jsonrpc" default:"2.0" required:"true"`
	Method  string                   `json:"method" default:"lti_form.upsert" required:"true"`
	Params  AuthProviderEditInputDTO `json:"params,omitempty"`
	ID      string                   `json:"id,omitempty" default:"1" required:"true"`
}

type AuthProviderEditOutputDTO struct {
	ID uint `json:"id" required:"true"`
}

type AuthProviderEditResponse struct {
	JSONRPC string                    `json:"jsonrpc" default:"2.0" required:"true"`
	Result  AuthProviderEditOutputDTO `json:"result,omitempty"`
	Error   interface{}               `json:"error,omitempty"`
	ID      string                    `json:"id,omitempty" default:"1" required:"true"`
}

func (u *AuthProviderEditUC) Execute(dto AuthProviderEditInputDTO) (AuthProviderEditOutputDTO, error) {
	entity := &models.AuthProvider{}
	entity.ID = dto.ID
	entity.Type = dto.Type
	entity.Name = dto.Name
	entity.LTIClientID = dto.ClientID
	entity.LTIDeploymentID = dto.DeploymentID
	entity.BaseURI = dto.BaseURI
	entity.LTIAuthTokenURI = dto.AuthTokenURI
	entity.LTIAuthLoginURI = dto.AuthLoginURI
	entity.KeySetURI = dto.KeySetURI
	entity.TargetLinkURI = dto.TargetLinkURI
	entity.SSOURL = dto.SSOURL
	err := u.AuthProviderQueries.Upsert(entity)
	return AuthProviderEditOutputDTO{ID: entity.ID}, err
}
