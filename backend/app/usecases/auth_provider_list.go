package usecases

import (
	"github.com/cms-lab-core/cms-labs-api/backend/app/models"
	"github.com/cms-lab-core/cms-labs-api/backend/app/queries/lti_query"
)

type AuthProviderListUC struct {
	AuthProviderQueries *lti_query.AuthProviderQueries
}

type AuthProviderListRequest struct {
	JSONRPC string                             `json:"jsonrpc" default:"2.0" required:"true"`
	Method  string                             `json:"method" default:"lti_form.list" required:"true"`
	Params  lti_query.AuthProviderListInputDTO `json:"params,omitempty"`
	ID      string                             `json:"id,omitempty" default:"1" required:"true"`
}

type AuthProviderListOutputDTO struct {
	Model      []models.AuthProviderListItem `json:"model" validate:"required"`
	TotalCount int64                         `json:"total_count" validate:"required"`
}

type AuthProviderListResponse struct {
	JSONRPC string                    `json:"jsonrpc" default:"2.0" required:"true"`
	Result  AuthProviderListOutputDTO `json:"result,omitempty"`
	Error   interface{}               `json:"error,omitempty"`
	ID      string                    `json:"id,omitempty" default:"1" required:"true"`
}

func (u *AuthProviderListUC) Execute(dto lti_query.AuthProviderListInputDTO) (AuthProviderListOutputDTO, error) {
	entities, count, err := u.AuthProviderQueries.List(&dto)
	return AuthProviderListOutputDTO{
		Model:      entities,
		TotalCount: count,
	}, err
}
