package usecases

import (
	"github.com/cms-lab-core/cms-labs-api/backend/app/models"
	"github.com/cms-lab-core/cms-labs-api/backend/app/queries"
)

type LTIAttemptGetUC struct {
	LTIAttemptQueries *queries.LTIAttemptQueries
}

type LTIAttemptGetInputDTO struct {
	ID uint `json:"id" required:"true"`
}

type LTIAttemptGetRequest struct {
	JSONRPC string                `json:"jsonrpc" default:"2.0" validate:"required"`
	Method  string                `json:"method" default:"lti_attempt.get" validate:"required"`
	Params  LTIAttemptGetInputDTO `json:"params,omitempty"`
	ID      string                `json:"id,omitempty" default:"1" validate:"required"`
}

type LTIAttemptGetOutputDTO struct {
	Model models.LTIAttempt `json:"model" required:"true"`
}

type LTIAttemptGetResponse struct {
	JSONRPC string                 `json:"jsonrpc" default:"2.0" validate:"required"`
	Result  LTIAttemptGetOutputDTO `json:"result,omitempty"`
	Error   interface{}            `json:"error,omitempty"`
	ID      string                 `json:"id,omitempty" default:"1" validate:"required"`
}

func (u *LTIAttemptGetUC) Execute(dto LTIAttemptGetInputDTO) (LTIAttemptGetOutputDTO, error) {
	form, err := u.LTIAttemptQueries.Get(dto.ID)
	return LTIAttemptGetOutputDTO{
		Model: form,
	}, err
}
