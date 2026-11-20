package usecases

import (
	"testing"
	"time"

	"github.com/cms-lab-core/cms-labs-api/backend/app/models"
	"github.com/cms-lab-core/cms-labs-api/backend/app/queries"
	"gorm.io/datatypes"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestBulkGradeProvenanceComesFromAuthenticatedServer(t *testing.T) {
	for _, kind := range []string{models.ServerTypeKubernetes, models.ServerTypePnet} {
		t.Run(kind, func(t *testing.T) {
			db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
			if err != nil {
				t.Fatal(err)
			}
			pool, _ := db.DB()
			defer pool.Close()
			for _, s := range []string{`CREATE TABLE lti_attempts(id integer PRIMARY KEY,created_at datetime,updated_at datetime,attempt_id text,status text,result text,user_id integer,server_id integer,lti_routing_id integer,synchronized_at datetime,room_id integer)`, `CREATE TABLE servers(id integer PRIMARY KEY,client_id text,type text,name text,last_online_status datetime,last_count_users integer,updated_at datetime)`, `CREATE TABLE users(id integer PRIMARY KEY,email text,name text)`, `CREATE TABLE lti_routings(id integer PRIMARY KEY,name text,labs_path text,test_path text)`} {
				if err := db.Exec(s).Error; err != nil {
					t.Fatal(err)
				}
			}
			if err := db.Exec("INSERT INTO servers(id,client_id,type) VALUES(7,?,?)", "authenticated", kind).Error; err != nil {
				t.Fatal(err)
			}
			server := uint(7)
			attempt := models.LTIAttempt{LTIAttemptBase: models.LTIAttemptBase{AttemptID: "attempt", ServerID: &server, Status: "active"}}
			if err := db.Create(&attempt).Error; err != nil {
				t.Fatal(err)
			}
			q := &queries.LTIAttemptQueries{DB: db}
			uc := (&LTIAttemptEditBulkUC{LTIAttemptQueries: q, ServerQueries: &queries.ServerQueries{DB: db}}).SetContext("authenticated")
			grade := datatypes.NewJSONType(models.LTIAttemptResult{Source: models.GradeSourceClabgate, CheckID: "check", CheckedAt: time.Now().UTC(), MaxScore: 10, CurrentScore: 0, ResultDisplay: "ready"})
			out, err := uc.Execute(LTIAttemptEditBulkInputDTO{Models: []LTIAttemptEditBulkInput{{AttemptID: attempt.AttemptID, Result: &grade}}})
			if err != nil || out.Count != 1 {
				t.Fatal(out, err)
			}
			stored, err := q.Get(attempt.ID)
			if err != nil {
				t.Fatal(err)
			}
			want := models.GradeSourceExternal
			if kind == models.ServerTypeKubernetes {
				want = models.GradeSourceClabgate
			}
			if stored.Result.Data().Source != want {
				t.Fatal("client forged result source")
			}
			now := time.Now().UTC()
			if err := db.Model(&models.LTIAttempt{}).Where("id = ?", attempt.ID).UpdateColumn("synchronized_at", now).Error; err != nil {
				t.Fatal(err)
			}
			if _, err := uc.Execute(LTIAttemptEditBulkInputDTO{Models: []LTIAttemptEditBulkInput{{AttemptID: attempt.AttemptID, Status: "completed"}}}); err != nil {
				t.Fatal(err)
			}
			stored, err = q.Get(attempt.ID)
			if err != nil || stored.SynchronizedAt == nil {
				t.Fatal("heartbeat/status requeued score", err)
			}
		})
	}
}
