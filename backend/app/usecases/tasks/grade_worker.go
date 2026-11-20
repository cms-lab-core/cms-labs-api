package tasks

import (
	"context"
	"sync"
	"time"
)

var gradeSyncWake = make(chan string, 1024)
var gradeSyncPending = make(map[string]bool)
var gradeSyncMutex sync.Mutex

func WakeGradeSync(attemptID string) {
	gradeSyncMutex.Lock()
	defer gradeSyncMutex.Unlock()
	if gradeSyncPending[attemptID] {
		return
	}
	select {
	case gradeSyncWake <- attemptID:
		gradeSyncPending[attemptID] = true
	default:
	}
}

func (u *LTISyncResultUC) Run(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	scan := func() {
		attempts, err := u.LTIAttemptQueries.ListPendingSync()
		if err != nil {
			u.Logger.Error().Err(err).Msg("read grade outbox")
			return
		}
		for _, attempt := range attempts {
			WakeGradeSync(attempt.AttemptID)
		}
	}
	scan()
	for {
		select {
		case <-ctx.Done():
			return
		case attemptID := <-gradeSyncWake:
			gradeSyncMutex.Lock()
			delete(gradeSyncPending, attemptID)
			gradeSyncMutex.Unlock()
			if err := u.SyncGradeToLTI(attemptID); err != nil {
				u.Logger.Error().Err(err).Str("attempt_id", attemptID).Msg("grade remains queued for retry")
			}
		case <-ticker.C:
			scan()
		}
	}
}
