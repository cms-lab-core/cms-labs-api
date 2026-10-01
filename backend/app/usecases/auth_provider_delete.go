package usecases

import (
	"github.com/cms-lab-core/cms-labs-api/backend/app/queries/lti_query"
)

type AuthProviderDeleteUC struct {
	AuthProviderQueries *lti_query.AuthProviderQueries
}

type AuthProviderDeleteInputDTO struct {
	ID uint `json:"id" required:"true"`
}

type AuthProviderDeleteRequest struct {
	JSONRPC string                `json:"jsonrpc" default:"2.0" required:"true"`
	Method  string                `json:"method" default:"lti_form.delete" required:"true"`
	Params  LTIAttemptGetInputDTO `json:"params,omitempty"`
	ID      string                `json:"id,omitempty" default:"1" required:"true"`
}

type AuthProviderDeleteResponse struct {
	JSONRPC string                     `json:"jsonrpc" default:"2.0" required:"true"`
	Result  AuthProviderDeleteInputDTO `json:"result,omitempty"`
	Error   interface{}                `json:"error,omitempty"`
	ID      string                     `json:"id,omitempty" default:"1" required:"true"`
}

func (u *AuthProviderDeleteUC) Execute(dto AuthProviderDeleteInputDTO) (AuthProviderDeleteInputDTO, error) {
	err := u.AuthProviderQueries.Delete(dto.ID)
	return dto, err
}
