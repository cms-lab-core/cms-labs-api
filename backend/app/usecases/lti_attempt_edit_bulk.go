package usecases

import (
	"errors"
	"fmt"
	"time"

	"github.com/cms-lab-core/cms-labs-api/backend/app/models"
	"github.com/cms-lab-core/cms-labs-api/backend/app/queries"
	"github.com/cms-lab-core/cms-labs-api/backend/app/usecases/tasks"
	"github.com/rs/zerolog/log"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type LTIAttemptEditBulkUC struct {
	LTIAttemptQueries *queries.LTIAttemptQueries
	ServerQueries     *queries.ServerQueries
	xServiceId        string
}

type LTIAttemptEditBulkInput struct {
	AttemptID string                                       `json:"attempt_id" validate:"required"`
	Status    string                                       `json:"status"`
	Result    *datatypes.JSONType[models.LTIAttemptResult] `json:"result" swaggertype:"object"`
}

type LTIAttemptEditBulkInputDTO struct {
	Models []LTIAttemptEditBulkInput `json:"models"`
}

type LTIAttemptEditBulkRequest struct {
	JSONRPC string                     `json:"jsonrpc" default:"2.0" validate:"required"`
	Method  string                     `json:"method" default:"lti_attempt.update" validate:"required"`
	Params  LTIAttemptEditBulkInputDTO `json:"params,omitempty"`
	ID      string                     `json:"id,omitempty" default:"1" validate:"required"`
}

type LTIAttemptEditBulkOutputDTO struct {
	Count uint `json:"count"`
}

type LTIAttemptEditBulkResponse struct {
	JSONRPC string                      `json:"jsonrpc" default:"2.0" validate:"required"`
	Result  LTIAttemptEditBulkOutputDTO `json:"result,omitempty"`
	Error   interface{}                 `json:"error,omitempty"`
	ID      string                      `json:"id,omitempty" default:"1" validate:"required"`
}

func (u *LTIAttemptEditBulkUC) SetContext(xServiceId string) *LTIAttemptEditBulkUC {
	u.xServiceId = xServiceId
	return u
}

func (u *LTIAttemptEditBulkUC) Execute(dto LTIAttemptEditBulkInputDTO) (LTIAttemptEditBulkOutputDTO, error) {
	if u.xServiceId == "" {
		return LTIAttemptEditBulkOutputDTO{}, errors.New("not authorized service")
	}
	serverEntity, err := u.ServerQueries.GetByClientId(u.xServiceId)
	if err != nil {
		return LTIAttemptEditBulkOutputDTO{}, err
	}
	attemptIds := make([]string, 0, len(dto.Models))
	processed := 0
	for _, attemptDTO := range dto.Models {
		gradeResult := attemptDTO.Result
		if gradeResult != nil {
			grade := gradeResult.Data()
			grade.Source = models.GradeSourceExternal
			if serverEntity.Type == models.ServerTypeKubernetes && grade.CheckID != "" && !grade.CheckedAt.IsZero() {
				grade.Source = models.GradeSourceClabgate
			}
			stored := datatypes.NewJSONType(grade)
			gradeResult = &stored
		}
		accepted, _, err := u.LTIAttemptQueries.ApplyExternalUpdate(attemptDTO.AttemptID, serverEntity.ID, attemptDTO.Status, gradeResult)
		if err != nil {
			return LTIAttemptEditBulkOutputDTO{}, err
		}
		if !accepted {
			continue
		}
		if gradeResult != nil && gradeResult.Data().Source == models.GradeSourceClabgate && gradeResult.Data().HasGrade() {
			attemptIds = append(attemptIds, attemptDTO.AttemptID)
		}
		processed++
	}
	activeAttempts, _, _ := u.LTIAttemptQueries.List(
		&queries.LTIAttemptSearchParams{
			Statuses:        []string{models.AttemptStatusActive},
			ServerClientIds: []string{serverEntity.ClientID},
			Limit:           queries.MaxLimitCount,
			Offset:          0,
		},
	)
	log.Info().Msg(fmt.Sprintf("LTIAttemptEditBulkUC: Update %s last online status and count attempts/user %d", u.xServiceId, len(activeAttempts)))
	err = u.ServerQueries.DB.Transaction(
		func(tx *gorm.DB) error {
			now := time.Now().UTC()
			serverEntity.LastOnlineStatus = &now
			serverEntity.LastCountUsers = len(activeAttempts)
			serverEntity.UpdatedAt = now
			return tx.Model(&models.Server{}).Where("id = ?", serverEntity.ID).Updates(map[string]interface{}{"last_online_status": now, "last_count_users": len(activeAttempts), "updated_at": now}).Error
		},
	)
	if err != nil {
		return LTIAttemptEditBulkOutputDTO{}, err
	}
	for _, attemptID := range attemptIds {
		tasks.WakeGradeSync(attemptID)
	}
	return LTIAttemptEditBulkOutputDTO{Count: uint(processed)}, err
}
