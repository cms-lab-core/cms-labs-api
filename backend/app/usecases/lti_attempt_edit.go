package usecases

import (
	"github.com/cms-lab-core/cms-labs-api/backend/app/models"
	"github.com/cms-lab-core/cms-labs-api/backend/app/queries"
	"github.com/cms-lab-core/cms-labs-api/backend/app/usecases/tasks"
)

type LTIAttemptEditUC struct {
	LTIAttemptQueries *queries.LTIAttemptQueries
}

type LTIAttemptEditInputDTO struct {
	ID       uint                     `json:"id"`
	ServerID *uint                    `json:"server_id,omitempty"`
	Status   string                   `json:"status" validate:"required"`
	Result   *models.LTIAttemptResult `json:"result" swaggertype:"object"`
}

type LTIAttemptEditRequest struct {
	JSONRPC string                 `json:"jsonrpc" default:"2.0" validate:"required"`
	Method  string                 `json:"method" default:"lti_attempt.update" validate:"required"`
	Params  LTIAttemptEditInputDTO `json:"params,omitempty"`
	ID      string                 `json:"id,omitempty" default:"1" validate:"required"`
}

type LTIAttemptEditOutputDTO struct {
	ID uint `json:"id" required:"true"`
}

type LTIAttemptEditResponse struct {
	JSONRPC string                  `json:"jsonrpc" default:"2.0" validate:"required"`
	Result  LTIAttemptEditOutputDTO `json:"result,omitempty"`
	Error   interface{}             `json:"error,omitempty"`
	ID      string                  `json:"id,omitempty" default:"1" validate:"required"`
}

func (u *LTIAttemptEditUC) Execute(dto LTIAttemptEditInputDTO) (LTIAttemptEditOutputDTO, error) {
	entity, err := u.LTIAttemptQueries.Get(dto.ID)
	if err != nil {
		return LTIAttemptEditOutputDTO{}, err
	}
	err = u.LTIAttemptQueries.EditFields(dto.ID, dto.ServerID, dto.Status, dto.Result)
	if err != nil {
		return LTIAttemptEditOutputDTO{ID: entity.ID}, err
	}
	tasks.WakeGradeSync(entity.AttemptID)
	return LTIAttemptEditOutputDTO{ID: entity.ID}, err
}
