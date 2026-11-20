package queries

import (
	"testing"
	"time"

	"github.com/cms-lab-core/cms-labs-api/backend/app/models"
	"gorm.io/datatypes"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestGradeUpdatesAreOrderedIdempotentAndIndependentOfStatus(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	defer sqlDB.Close()
	if err := db.Exec(`CREATE TABLE lti_attempts(id integer PRIMARY KEY,created_at datetime,updated_at datetime,attempt_id text,status text,result text,user_id integer,server_id integer,lti_routing_id integer,synchronized_at datetime,room_id integer)`).Error; err != nil {
		t.Fatal(err)
	}
	serverID := uint(7)
	base := time.Now().UTC().Add(-time.Hour)
	makeGrade := func(id string, score float64, at time.Time) *datatypes.JSONType[models.LTIAttemptResult] {
		g := datatypes.NewJSONType(models.LTIAttemptResult{CheckID: id, CheckedAt: at, MaxScore: 10, CurrentScore: score, ResultDisplay: id})
		return &g
	}
	attempt := models.LTIAttempt{LTIAttemptBase: models.LTIAttemptBase{AttemptID: "attempt", ServerID: &serverID, Status: "active", Result: makeGrade("old", 1, base)}}
	if err := db.Create(&attempt).Error; err != nil {
		t.Fatal(err)
	}
	q := LTIAttemptQueries{DB: db}
	if accepted, _, err := q.ApplyExternalUpdate("attempt", 9, "", makeGrade("foreign", 10, base.Add(time.Hour))); err != nil || accepted {
		t.Fatal("foreign service updated grade", err)
	}
	newer := makeGrade("new", 8, base.Add(time.Minute))
	if accepted, changed, err := q.ApplyExternalUpdate("attempt", 7, "", newer); err != nil || !accepted || !changed {
		t.Fatal("new grade rejected", err)
	}
	ack := time.Now().UTC()
	if err := db.Model(&models.LTIAttempt{}).Where("id = ?", attempt.ID).UpdateColumn("synchronized_at", ack).Error; err != nil {
		t.Fatal(err)
	}
	for _, grade := range []*datatypes.JSONType[models.LTIAttemptResult]{makeGrade("old", 1, base), makeGrade("new", 0, base.Add(time.Minute))} {
		if accepted, changed, err := q.ApplyExternalUpdate("attempt", 7, "", grade); err != nil || !accepted || changed {
			t.Fatal("replay must be acknowledged without changing grade", err)
		}
	}
	stored, err := q.GetByAttemptID("attempt")
	if err != nil {
		t.Fatal(err)
	}
	if stored.Result.Data().CheckID != "new" || stored.Result.Data().CurrentScore != 8 || stored.SynchronizedAt == nil {
		t.Fatal("replay changed grade or delivery state")
	}
	if err := q.EditFields(attempt.ID, nil, "completed", nil); err != nil {
		t.Fatal(err)
	}
	if accepted, changed, err := q.ApplyExternalUpdate("attempt", 7, "active", makeGrade("last", 0, base.Add(2*time.Minute))); err != nil || !accepted || !changed {
		t.Fatal("late final grade rejected", err)
	}
	stored, err = q.GetByAttemptID("attempt")
	if err != nil {
		t.Fatal(err)
	}
	if stored.Status != "completed" || stored.Result.Data().CheckID != "last" || stored.Result.Data().CurrentScore != 0 || stored.SynchronizedAt != nil {
		t.Fatalf("late grade lost or reopened lab: %+v", stored)
	}
	pending, err := q.ListPendingSync()
	if err != nil || len(pending) != 1 || pending[0].Status != "completed" {
		t.Fatal("completed attempts must remain in delivery outbox", err)
	}
	if err := q.EditFields(attempt.ID, nil, "terminating", nil); err != nil {
		t.Fatal(err)
	}
	stored, err = q.GetByAttemptID("attempt")
	if err != nil || stored.Result.Data().CheckID != "last" {
		t.Fatal("status-only edit erased grade", err)
	}
}
