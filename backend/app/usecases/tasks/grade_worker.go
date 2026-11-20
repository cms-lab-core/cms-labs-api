package tasks

import (
	"context"
	"sync"
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
		}
	}
}
