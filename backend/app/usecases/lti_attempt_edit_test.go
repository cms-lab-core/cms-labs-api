package usecases

import (
	"testing"

	"github.com/cms-lab-core/cms-labs-api/backend/app/models"
	"github.com/cms-lab-core/cms-labs-api/backend/app/queries"
	"gorm.io/datatypes"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestLTIAttemptEditPreservesGradeAndOptionalServer(t *testing.T) {
	for _, explicitServer := range []bool{false, true} {
		t.Run(map[bool]string{true: "explicit server", false: "omitted server"}[explicitServer], func(t *testing.T) {
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
			serverID := uint(17)
			grade := datatypes.NewJSONType(models.LTIAttemptResult{CheckID: "last", MaxScore: 10, CurrentScore: 8, ResultDisplay: "8/10"})
			attempt := models.LTIAttempt{LTIAttemptBase: models.LTIAttemptBase{AttemptID: "attempt", ServerID: &serverID, Status: "active", Result: &grade}}
			if err := db.Create(&attempt).Error; err != nil {
				t.Fatal(err)
			}
			q := &queries.LTIAttemptQueries{DB: db}
			u := LTIAttemptEditUC{LTIAttemptQueries: q}
			dto := LTIAttemptEditInputDTO{ID: attempt.ID, Status: models.AttemptStatusTerminating}
			want := serverID
			if explicitServer {
				want = 23
				dto.ServerID = &want
			}
			if _, err := u.Execute(dto); err != nil {
				t.Fatal(err)
			}
			stored, err := q.Get(attempt.ID)
			if err != nil {
				t.Fatal(err)
			}
			if stored.ServerID == nil || *stored.ServerID != want || stored.Status != models.AttemptStatusTerminating || stored.Result.Data().CheckID != "last" {
				t.Fatalf("status edit changed grade/server: %+v", stored)
			}
		})
	}
}
