package tasks

import (
	"errors"
	"fmt"
	"time"

	"github.com/cms-lab-core/cms-labs-api/backend/app/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/cms-lab-core/cms-labs-api/backend/app/queries"
	"github.com/cms-lab-core/cms-labs-api/backend/app/usecases/lti_connector"
	"github.com/cms-lab-core/cms-labs-api/backend/app/usecases/lti_connector/connector"
	"github.com/rs/zerolog"
)

type LTISyncResultUC struct {
	LTIAttemptQueries *queries.LTIAttemptQueries
	LTIConnectorAPI   *lti_connector.LTIConnectorAPI
	Logger            *zerolog.Logger
}

func (u *LTISyncResultUC) Execute() error {
	attempts, err := u.LTIAttemptQueries.ListPendingSync()
	if err != nil {
		u.Logger.Error().Err(err).Msg("LTISyncResultUC: failed to get pending attempts")
		return err
	}

	count := uint(0)
	for _, attempt := range attempts {
		err := u.SyncGradeToLTI(attempt.AttemptID)
		if err != nil {
			u.Logger.Error().Err(err).Str("attempt_id", attempt.AttemptID).Msg("LTISyncResultUC: failed to sync grade")
			continue
		}

		count++
	}

	u.Logger.Info().Msgf("LTISyncResultUC: synced %d grades", count)
	return nil
}

func (u *LTISyncResultUC) SyncGradeToLTI(attemptID string) error {
	return u.syncGrade(attemptID, func(tx *gorm.DB, attempt models.LTIAttempt) error {
		conn, err := u.LTIConnectorAPI.WithDB(tx).ConnectorByAttemptID(attempt.AttemptID)
		if err != nil {
			return err
		}
		ags, err := conn.UpgradeAGS()
		if err != nil {
			return err
		}
		result := attempt.Result.Data()
		return ags.PutScore(connector.Score{
			Timestamp:  attempt.UpdatedAt.UTC().Format(time.RFC3339Nano),
			ScoreGiven: result.CurrentScore, ScoreMaximum: result.MaxScore, Comment: result.ResultDisplay,
			ActivityProgress: ags.CastStatusToActivity(attempt.Status), GradingProgress: connector.GradingFullyGraded,
		}, true)
	})
}

func (u *LTISyncResultUC) syncGrade(attemptID string, send func(*gorm.DB, models.LTIAttempt) error) error {
	return u.LTIAttemptQueries.DB.Transaction(func(tx *gorm.DB) error {
		var attempt models.LTIAttempt
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE", Options: "NOWAIT"}).Where("attempt_id = ?", attemptID).First(&attempt).Error; err != nil {
			return err
		}
		if attempt.Result == nil || attempt.SynchronizedAt != nil {
			return nil
		}
		result := attempt.Result.Data()
		if !(result.CurrentScore >= 0 && result.CurrentScore <= result.MaxScore && result.MaxScore > 0 && result.ResultDisplay != "") {
			return errors.New("invalid grade remains pending")
		}
		if err := send(tx, attempt); err != nil {
			return fmt.Errorf("deliver grade %s: %w", attemptID, err)
		}
		return tx.Model(&models.LTIAttempt{}).Where("id = ?", attempt.ID).UpdateColumn("synchronized_at", time.Now().UTC()).Error
	})
}
