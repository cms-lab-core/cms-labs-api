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
	LTISyncResultUC   *tasks.LTISyncResultUC
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
		attempt, err := u.LTIAttemptQueries.GetByAttemptID(attemptDTO.AttemptID)
		if err != nil {
			log.Warn().Msg(fmt.Sprintf("LTIAttemptEditBulkUC: LTI attempt id %s not found", attemptDTO.AttemptID))
			continue
		}
		if attempt.ServerID != nil && *attempt.ServerID != serverEntity.ID {
			log.Warn().Msg(fmt.Sprintf(
				"LTIAttemptEditBulkUC: LTI attempt id %s is not owned by service %s",
				attemptDTO.AttemptID,
				u.xServiceId,
			))
			continue
		}
		if attemptDTO.Result != nil && attemptDTO.Result.Data().CheckID != "" && attempt.Result != nil &&
			attempt.Result.Data().CheckID == attemptDTO.Result.Data().CheckID {
			processed++
			continue
		}
		attempt.SetStatus(attemptDTO.Status)
		if attemptDTO.Result != nil {
			attempt.Result = attemptDTO.Result
		}
		attempt.SynchronizedAt = nil
		err = u.LTIAttemptQueries.Upsert(&attempt)
		if err != nil {
			log.Warn().Msg(fmt.Sprintf(
				"LTIAttemptEditBulkUC: LTI attempt id %s update not complete %+v",
				attemptDTO.AttemptID,
				err,
			))
			return LTIAttemptEditBulkOutputDTO{}, err
		}
		attemptIds = append(attemptIds, attemptDTO.AttemptID)
		processed++
	}
	activeAttempts, _ := u.LTIAttemptQueries.List(
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
			serverEntity, err := u.ServerQueries.GetByClientId(u.xServiceId)
			if err != nil {
				return err
			}
			now := time.Now().UTC()
			serverEntity.LastOnlineStatus = &now
			serverEntity.LastCountUsers = len(activeAttempts)
			serverEntity.UpdatedAt = now
			tx.Save(serverEntity)
			return nil
		},
	)
	if err != nil {
		return LTIAttemptEditBulkOutputDTO{}, err
	}
	// Тут можно подумать над упрощением
	for _, attemptId := range attemptIds {
		_ = u.LTISyncResultUC.SyncGradeToLTI(
			attemptId,
		)
	}
	return LTIAttemptEditBulkOutputDTO{Count: uint(processed)}, err
}
