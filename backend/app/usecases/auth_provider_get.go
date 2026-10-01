package usecases

import (
	"strings"

	"github.com/cms-lab-core/cms-labs-api/backend/app/models"
	"github.com/cms-lab-core/cms-labs-api/backend/app/queries/lti_query"
)

type AuthProviderGetUC struct {
	AuthProviderQueries *lti_query.AuthProviderQueries
}

type AuthProviderGetInputDTO struct {
	ID uint `json:"id" required:"true"`
}

type AuthProviderGetRequest struct {
	JSONRPC string                  `json:"jsonrpc" default:"2.0" required:"true"`
	Method  string                  `json:"method" default:"lti_form.get" required:"true"`
	Params  AuthProviderGetInputDTO `json:"params,omitempty"`
	ID      string                  `json:"id,omitempty" default:"1" required:"true"`
}

type AuthProviderGetOutputDTO struct {
	Model models.AuthProvider `json:"model" required:"true"`
}

type AuthProviderGetResponse struct {
	JSONRPC string                   `json:"jsonrpc" default:"2.0" required:"true"`
	Result  AuthProviderGetOutputDTO `json:"result,omitempty"`
	Error   interface{}              `json:"error,omitempty"`
	ID      string                   `json:"id,omitempty" default:"1" required:"true"`
}

func (u *AuthProviderGetUC) ReplacePublicKey(publicKey string) string {
	return strings.Replace(
		strings.Replace(
			publicKey,
			"BEGIN RSA PUBLIC KEY",
			"BEGIN PUBLIC KEY",
			1,
		),
		"END RSA PUBLIC KEY",
		"END PUBLIC KEY",
		1,
	)
}
func (u *AuthProviderGetUC) Execute(dto AuthProviderGetInputDTO) (AuthProviderGetOutputDTO, error) {
	form, err := u.AuthProviderQueries.Get(dto.ID)
	form.PublicKey = u.ReplacePublicKey(form.PublicKey)
	return AuthProviderGetOutputDTO{
		Model: form,
	}, err
}
