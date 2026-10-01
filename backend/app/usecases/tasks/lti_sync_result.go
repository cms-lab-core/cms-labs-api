package tasks

import (
	"time"

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

func (u *LTISyncResultUC) SyncGradeToLTI(attemptId string) error {
	attempt, err := u.LTIAttemptQueries.GetByAttemptID(attemptId)
	if err != nil {
		return err
	}
	if attempt.Result == nil {
		u.Logger.Info().Msgf("LTISyncResultUC: syncing grade for attempt_id=%s is empty", attempt.AttemptID)
		return nil
	}
	result := attempt.Result.Data()
	// A zero score is still a valid grade and must be delivered to the LMS.
	if !(result.CurrentScore >= 0 && result.CurrentScore <= result.MaxScore && result.MaxScore > 0 && result.ResultDisplay != "") {
		u.Logger.Info().Msgf("LTISyncResultUC: syncing grade for attempt_id=%s is empty model", attempt.AttemptID)
		return nil
	}
	u.Logger.Info().Msgf("LTISyncResultUC: syncing grade for attempt_id=%s, maxScore=%f, currentScore=%f, comment=%s",
		attempt.AttemptID, result.MaxScore, result.CurrentScore, result.ResultDisplay)
	conn, err := u.LTIConnectorAPI.ConnectorByAttemptID(attempt.AttemptID)
	if err != nil {
		return err
	}
	ags, err := conn.UpgradeAGS()
	if err != nil {
		return err
	}
	now := time.Now().UTC()
	err = ags.PutScore(
		connector.Score{
			Timestamp:        now.Format(time.RFC3339),
			ScoreGiven:       result.CurrentScore,
			ScoreMaximum:     result.MaxScore,
			Comment:          result.ResultDisplay,
			ActivityProgress: ags.CastStatusToActivity(attempt.Status),
			GradingProgress:  ags.CastStatusToGrade(attempt.Status),
		},
		true,
	)
	if err != nil {
		return err
	}
	attempt.SynchronizedAt = &now
	err = u.LTIAttemptQueries.Upsert(&attempt)
	if err != nil {
		u.Logger.Error().Err(err).Str("attempt_id", attempt.AttemptID).Msg("LTISyncResultUC: failed to update synchronized_at")
		return err
	}
	u.Logger.Info().Msgf("LTISyncResultUC: successfully synced grade for attempt_id=%s", attempt.AttemptID)
	return nil
}
