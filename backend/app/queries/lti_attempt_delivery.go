package queries

import (
	"errors"
	"time"

	"github.com/cms-lab-core/cms-labs-api/backend/app/models"
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func (q *LTIAttemptQueries) ApplyExternalUpdate(attemptID string, serverID uint, status string, result *datatypes.JSONType[models.LTIAttemptResult]) (accepted, changed bool, err error) {
	err = q.DB.Transaction(func(tx *gorm.DB) error {
		var attempt models.LTIAttempt
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("attempt_id = ?", attemptID).First(&attempt).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil
			}
			return err
		}
		if attempt.ServerID != nil && *attempt.ServerID != serverID {
			return nil
		}
		accepted = true
		previousStatus := attempt.Status
		attempt.SetStatus(status)
		updates := map[string]interface{}{}
		if attempt.Status != previousStatus {
			updates["status"] = attempt.Status
		}
		if result != nil {
			incoming := result.Data()
			current := models.LTIAttemptResult{}
			if attempt.Result != nil {
				current = attempt.Result.Data()
			}
			duplicate := incoming.CheckID != "" && incoming.CheckID == current.CheckID
			older := !incoming.CheckedAt.IsZero() && !current.CheckedAt.IsZero() && (incoming.CheckedAt.Before(current.CheckedAt) || incoming.CheckedAt.Equal(current.CheckedAt) && incoming.CheckID < current.CheckID)
			if !duplicate && !older {
				updates["result"] = result
			}
		}
		if len(updates) == 0 {
			return nil
		}
		if _, resultChanged := updates["result"]; resultChanged {
			updates["synchronized_at"] = nil
		}
		updates["updated_at"] = time.Now().UTC()
		if err := tx.Model(&models.LTIAttempt{}).Where("id = ?", attempt.ID).Updates(updates).Error; err != nil {
			return err
		}
		changed = true
		return nil
	})
	return
}

func (q *LTIAttemptQueries) EditFields(id uint, serverID *uint, status string, result *models.LTIAttemptResult) error {
	updates := map[string]interface{}{"status": status, "updated_at": time.Now().UTC()}
	if serverID != nil {
		updates["server_id"] = *serverID
	}
	if result != nil {
		manual := *result
		manual.Source = models.GradeSourceManual
		updates["result"] = datatypes.NewJSONType(manual)
		updates["synchronized_at"] = nil
	}
	return q.DB.Model(&models.LTIAttempt{}).Where("id = ?", id).Updates(updates).Error
}

func (q *LTIAttemptQueries) IsGradeFromClabgate(attempt models.LTIAttempt) (bool, error) {
	if attempt.Result == nil || !attempt.Result.Data().HasGrade() {
		return false, nil
	}
	result := attempt.Result.Data()
	if result.Source == models.GradeSourceClabgate {
		return true, nil
	}
	if result.Source != "" || result.CheckedAt.IsZero() || attempt.ServerID == nil {
		return false, nil
	}
	var count int64
	err := q.DB.Model(&models.Server{}).Where("id = ? AND type = ?", *attempt.ServerID, models.ServerTypeKubernetes).Count(&count).Error
	return count == 1, err
}
