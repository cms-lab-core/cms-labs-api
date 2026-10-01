package usecases

import (
	"github.com/maintainer64/cms-labs-api/backend/app/models"
	"github.com/maintainer64/cms-labs-api/backend/app/queries"
	"github.com/maintainer64/cms-labs-api/backend/app/usecases/tasks"
	"gorm.io/datatypes"
)

type LTIAttemptEditUC struct {
	LTIAttemptQueries *queries.LTIAttemptQueries
	LTISyncResultUC   *tasks.LTISyncResultUC
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
	applyLTIAttemptEdit(&entity, dto)
	entity.SynchronizedAt = nil
	err = u.LTIAttemptQueries.Upsert(&entity)
	if err != nil {
		return LTIAttemptEditOutputDTO{ID: entity.ID}, err
	}
	_ = u.LTISyncResultUC.SyncGradeToLTI(entity.AttemptID)
	return LTIAttemptEditOutputDTO{ID: entity.ID}, err
}

func applyLTIAttemptEdit(entity *models.LTIAttempt, dto LTIAttemptEditInputDTO) {
	entity.ID = dto.ID
	if dto.ServerID != nil {
		entity.ServerID = dto.ServerID
	}
	entity.Status = dto.Status
	if dto.Result != nil {
		ltiResult := datatypes.NewJSONType(*dto.Result)
		entity.Result = &ltiResult
	}
}
