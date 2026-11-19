package queries

import (
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestAttemptServerFilterUsesDatabaseIDsAndKeepsExternalScope(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	defer sqlDB.Close()
	for _, sql := range []string{
		`CREATE TABLE users (id integer PRIMARY KEY, email text, name text)`,
		`CREATE TABLE servers (id integer PRIMARY KEY, client_id text, name text)`,
		`CREATE TABLE lti_routings (id integer PRIMARY KEY, name text, labs_path text, test_path text)`,
		`CREATE TABLE lti_attempts (id integer PRIMARY KEY, attempt_id text, status text, result text, user_id integer, server_id integer, lti_routing_id integer, synchronized_at datetime, created_at datetime, updated_at datetime)`,
		`INSERT INTO users VALUES (1, 'student@example.test', 'student')`,
		`INSERT INTO servers VALUES (7, 'server-oauth-a', 'A'), (9, 'server-oauth-b', 'B')`,
		`INSERT INTO lti_routings VALUES (1, 'lab', 'repo', 'check')`,
		`INSERT INTO lti_attempts (id,attempt_id,status,user_id,server_id,lti_routing_id) VALUES (1,'one','active',1,7,1),(2,'two','active',1,9,1)`,
	} {
		if err := db.Exec(sql).Error; err != nil {
			t.Fatal(err)
		}
	}
	q := LTIAttemptQueries{DB: db}
	for _, tc := range []struct {
		filter LTIAttemptSearchParams
		want   int64
	}{
		{LTIAttemptSearchParams{}, 2},
		{LTIAttemptSearchParams{ServerIds: []uint{7}}, 1},
		{LTIAttemptSearchParams{ServerClientIds: []string{"server-oauth-b"}}, 1},
		{LTIAttemptSearchParams{ServerIds: []uint{7}, ServerClientIds: []string{"server-oauth-b"}}, 0},
	} {
		tc.filter.Limit = 10
		rows, count, err := q.List(&tc.filter)
		if err != nil {
			t.Fatal(err)
		}
		if count != tc.want || int64(len(rows)) != tc.want {
			t.Fatalf("filter %+v: count=%d rows=%d", tc.filter, count, len(rows))
		}
		if len(tc.filter.ServerIds) > 0 {
			for _, row := range rows {
				if row.ServerID == nil || *row.ServerID != 7 {
					t.Fatal("other server leaked into filter")
				}
			}
		}
	}
}
