package usecases

import (
	"github.com/cms-lab-core/cms-labs-api/backend/app/models"
	"github.com/cms-lab-core/cms-labs-api/backend/app/queries"
)

type LTIAttemptListUC struct {
	LTIAttemptQueries *queries.LTIAttemptQueries
}

type LTIAttemptListRequest struct {
	JSONRPC string                         `json:"jsonrpc" default:"2.0" validate:"required"`
	Method  string                         `json:"method" default:"lti_attempt.list" validate:"required"`
	Params  queries.LTIAttemptSearchParams `json:"params,omitempty"`
	ID      string                         `json:"id,omitempty" default:"1" validate:"required"`
}

type LTIAttemptListOutputDTO struct {
	Model []models.LTIAttemptListItem `json:"model" validate:"required"`
}

type LTIAttemptListResponse struct {
	JSONRPC string                  `json:"jsonrpc" default:"2.0" validate:"required"`
	Result  LTIAttemptListOutputDTO `json:"result,omitempty"`
	Error   interface{}             `json:"error,omitempty"`
	ID      string                  `json:"id,omitempty" default:"1" validate:"required"`
}

func (u *LTIAttemptListUC) Execute(dto queries.LTIAttemptSearchParams) (LTIAttemptListOutputDTO, error) {
	entities, err := u.LTIAttemptQueries.List(&dto)
	return LTIAttemptListOutputDTO{
		Model: entities,
	}, err
}
