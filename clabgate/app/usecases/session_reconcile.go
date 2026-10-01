package usecases

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/cms-lab-core/cms-labs-api/clabgate/app/queries"
	"github.com/cms-lab-core/cms-labs-api/clabgate/pkg/configs"
	"github.com/cms-lab-core/cms-labs-api/shared/cms_client"
)

type checkerGrade struct {
	MaxScore      float64       `json:"max_score"`
	CurrentScore  float64       `json:"current_score"`
	ResultDisplay string        `json:"result_display"`
	Report        string        `json:"report,omitempty"`
	CheckID       string        `json:"check_id"`
	Logs          string        `json:"logs,omitempty"`
	Tasks         []checkerTask `json:"tasks,omitempty"`
}

type checkerTask struct {
	Title       string       `json:"title"`
	Description string       `json:"description,omitempty"`
	Logs        []checkerLog `json:"logs,omitempty"`
	Complete    bool         `json:"complete"`
}

type checkerLog struct {
	Node      string `json:"node,omitempty"`
	Namespace string `json:"namespace,omitempty"`
	Message   string `json:"message"`
}

func decodeCheckerGrade(payload, checkID, logs string) (checkerGrade, error) {
	grade := checkerGrade{}
	if err := json.Unmarshal([]byte(payload), &grade); err != nil {
		return checkerGrade{}, fmt.Errorf("decode checker result: %w", err)
	}
	if grade.MaxScore <= 0 || grade.CurrentScore < 0 || grade.CurrentScore > grade.MaxScore || strings.TrimSpace(grade.ResultDisplay) == "" {
		return checkerGrade{}, fmt.Errorf("invalid checker grade: score=%g/%g result_display=%q", grade.CurrentScore, grade.MaxScore, grade.ResultDisplay)
	}
	grade.CheckID = checkID
	grade.Logs = logs
	return grade, nil
}

// Reconcile synchronizes the CMS attempt lifecycle with Kubernetes, which is the
// source of truth for runtime sessions. It replaces the old JupyterHub idle checker.
func (u *SessionsUC) Reconcile(ctx context.Context) (int, error) {
	changes := 0
	sessions, err := u.KubernetesAdminQuery.ListSessions(ctx, "", configs.AppConfig.Session.WorkspacePrefix)
	if err != nil {
		return 0, fmt.Errorf("list Kubernetes sessions: %w", err)
	}
	sessionsByAttempt := make(map[string]queries.SessionRecord, len(sessions))
	for _, session := range sessions {
		sessionsByAttempt[session.AttemptID] = session
		results, resultErr := u.KubernetesAdminQuery.PendingCheckerResults(ctx, session.Namespace)
		if resultErr != nil {
			return 0, fmt.Errorf("read checker results for %s: %w", session.AttemptID, resultErr)
		}
		for _, result := range results {
			grade, decodeErr := decodeCheckerGrade(result.Payload, result.CheckID, result.Logs)
			if decodeErr != nil {
				u.Logger.Error().Err(decodeErr).Str("job", result.JobName).Msg("decode checker result")
				continue
			}
			count, updateErr := u.CMSClient.UpdateAttempts([]cms_client.UpdateAttemptParams{{
				AttemptID: result.AttemptID,
				Status:    "active",
				Result:    grade,
			}})
			if updateErr != nil {
				return 0, fmt.Errorf("publish checker result for %s: %w", result.AttemptID, updateErr)
			}
			if count != 1 {
				return 0, fmt.Errorf("publish checker result for %s: CMS updated %d attempts", result.AttemptID, count)
			}
			if markErr := u.KubernetesAdminQuery.MarkCheckerResultSynced(ctx, result.Namespace, result.JobName); markErr != nil {
				return 0, fmt.Errorf("mark checker result %s synced: %w", result.JobName, markErr)
			}
			changes++
		}
	}

	attempts, err := u.CMSClient.ListAttempts(cms_client.ListAttemptsParams{
		Statuses: []string{"pending", "active", "terminating"},
		Limit:    5000,
	})
	if err != nil {
		return 0, fmt.Errorf("list CMS attempts: %w", err)
	}

	for _, attempt := range attempts {
		session, sessionExists := sessionsByAttempt[attempt.AttemptID]
		switch attempt.Status {
		case "pending":
			if !sessionExists || session.Phase == queries.SessionPhasePending || session.Phase == queries.SessionPhaseProvisioning {
				provisioned, provisionErr := u.ensureAttemptSession(
					ctx, attempt, strconv.FormatInt(attempt.UserID, 10), attempt.UserName,
				)
				if provisionErr != nil {
					u.Logger.Error().Err(provisionErr).Str("attempt_id", attempt.AttemptID).Msg("provision pending session")
					continue
				}
				changes++
				session = provisioned
				sessionExists = true
				sessionsByAttempt[attempt.AttemptID] = provisioned
			}
			if sessionExists && session.Phase == queries.SessionPhaseReady {
				if phaseErr := u.KubernetesAdminQuery.MarkSessionPhase(ctx, session.Namespace, queries.SessionPhaseReady); phaseErr != nil {
					return changes, fmt.Errorf("mark session %s ready: %w", attempt.AttemptID, phaseErr)
				}
				count, updateErr := u.CMSClient.UpdateAttempts([]cms_client.UpdateAttemptParams{{AttemptID: attempt.AttemptID, Status: "active"}})
				if updateErr != nil {
					return changes, fmt.Errorf("activate ready attempt %s: %w", attempt.AttemptID, updateErr)
				}
				if count != 1 {
					return changes, fmt.Errorf("activate ready attempt %s: CMS updated %d attempts", attempt.AttemptID, count)
				}
				if phaseErr := u.KubernetesAdminQuery.MarkSessionPhase(ctx, session.Namespace, queries.SessionPhaseActive); phaseErr != nil {
					return changes, fmt.Errorf("mark session %s active: %w", attempt.AttemptID, phaseErr)
				}
				changes++
			}
		case "active":
			if !sessionExists {
				count, updateErr := u.CMSClient.UpdateAttempts([]cms_client.UpdateAttemptParams{{AttemptID: attempt.AttemptID, Status: "completed"}})
				if updateErr != nil {
					return changes, fmt.Errorf("complete missing session %s: %w", attempt.AttemptID, updateErr)
				}
				changes += count
			} else if session.Phase == queries.SessionPhasePending || session.Phase == queries.SessionPhaseProvisioning {
				if _, provisionErr := u.ensureAttemptSession(ctx, attempt, strconv.FormatInt(attempt.UserID, 10), attempt.UserName); provisionErr != nil {
					u.Logger.Error().Err(provisionErr).Str("attempt_id", attempt.AttemptID).Msg("repair active session")
				} else {
					changes++
				}
			} else if phaseErr := u.KubernetesAdminQuery.MarkSessionPhase(ctx, session.Namespace, queries.SessionPhaseActive); phaseErr != nil {
				return changes, fmt.Errorf("mark session %s active: %w", attempt.AttemptID, phaseErr)
			}
		case "terminating":
			if sessionExists {
				if deleteErr := u.KubernetesAdminQuery.DeleteSession(ctx, attempt.AttemptID, "", true); deleteErr != nil {
					u.Logger.Error().Err(deleteErr).Str("attempt_id", attempt.AttemptID).Msg("delete terminating session")
				} else {
					changes++
				}
				// Namespace deletion is asynchronous. Keep the CMS attempt terminating
				// until a later list confirms that all namespaced resources are gone.
				continue
			}
			count, updateErr := u.CMSClient.UpdateAttempts([]cms_client.UpdateAttemptParams{{AttemptID: attempt.AttemptID, Status: "completed"}})
			if updateErr != nil {
				return changes, fmt.Errorf("complete terminating attempt %s: %w", attempt.AttemptID, updateErr)
			}
			changes += count
		}
	}
	return changes, nil
}
