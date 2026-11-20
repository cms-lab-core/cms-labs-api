package queries

import (
	"testing"
	"time"

	"github.com/cms-lab-core/cms-labs-api/backend/app/models"
	"gorm.io/datatypes"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestPendingSyncOnlyIncludesClabgateGrades(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	pool, _ := db.DB()
	defer pool.Close()
	for _, s := range []string{`CREATE TABLE lti_attempts(id integer PRIMARY KEY,created_at datetime,updated_at datetime,attempt_id text,status text,result text,user_id integer,server_id integer,lti_routing_id integer,synchronized_at datetime,room_id integer)`, `CREATE TABLE servers(id integer PRIMARY KEY,type text)`, `INSERT INTO servers(id,type) VALUES(7,'k8s'),(8,'pnet')`} {
		if err := db.Exec(s).Error; err != nil {
			t.Fatal(err)
		}
	}
	k8s, pnet := uint(7), uint(8)
	now := time.Now().UTC()
	cases := []struct {
		name, source string
		server       *uint
		checked      bool
		max          float64
		want         bool
	}{
		{"clabgate", models.GradeSourceClabgate, &k8s, true, 10, true},
		{"manual", models.GradeSourceManual, &k8s, true, 10, false},
		{"external", models.GradeSourceExternal, &k8s, true, 10, false},
		{"empty", models.GradeSourceClabgate, &k8s, true, 0, false},
		{"legacy-k8s", "", &k8s, true, 10, true},
		{"legacy-pnet", "", &pnet, true, 10, false},
		{"legacy-no-check-time", "", &k8s, false, 10, false},
	}
	for _, c := range cases {
		g := models.LTIAttemptResult{Source: c.source, CheckID: c.name, MaxScore: c.max, ResultDisplay: "ready"}
		if c.checked {
			g.CheckedAt = now
		}
		value := datatypes.NewJSONType(g)
		a := models.LTIAttempt{LTIAttemptBase: models.LTIAttemptBase{AttemptID: c.name, Status: "completed", ServerID: c.server, Result: &value}}
		if err := db.Create(&a).Error; err != nil {
			t.Fatal(err)
		}
		ok, err := (&LTIAttemptQueries{DB: db}).IsGradeFromClabgate(a)
		if err != nil || ok != c.want {
			t.Fatalf("%s eligibility=%v error=%v", c.name, ok, err)
		}
	}
	q := LTIAttemptQueries{DB: db}
	pending, err := q.ListPendingSync()
	if err != nil {
		t.Fatal(err)
	}
	if len(pending) != 2 {
		t.Fatalf("pending=%d", len(pending))
	}
	for _, a := range pending {
		if a.AttemptID != "clabgate" && a.AttemptID != "legacy-k8s" {
			t.Fatal("unrelated grade queued", a.AttemptID)
		}
	}
	var a models.LTIAttempt
	if err := db.Where("attempt_id = ?", "clabgate").First(&a).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Model(&a).UpdateColumn("synchronized_at", now).Error; err != nil {
		t.Fatal(err)
	}
	if _, _, err := q.ApplyExternalUpdate(a.AttemptID, k8s, "completed", nil); err != nil {
		t.Fatal(err)
	}
	if err := q.EditFields(a.ID, nil, "terminating", nil); err != nil {
		t.Fatal(err)
	}
	stored, err := q.Get(a.ID)
	if err != nil || stored.SynchronizedAt == nil {
		t.Fatal("status-only update requeued delivered grade", err)
	}
	manual := stored.Result.Data()
	manual.Source = models.GradeSourceClabgate
	if err := q.EditFields(a.ID, nil, "completed", &manual); err != nil {
		t.Fatal(err)
	}
	stored, err = q.Get(a.ID)
	if err != nil || stored.Result.Data().Source != models.GradeSourceManual {
		t.Fatal("manual edit forged clabgate provenance", err)
	}
}
