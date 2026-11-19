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

func TestAttemptFiltersCombineAndKeepUnassignedAttempts(t *testing.T) {
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
		`INSERT INTO users VALUES (1,'one@example.test','one'),(2,'two@example.test','two')`,
		`INSERT INTO servers VALUES (7,'oauth-a','A'),(9,'oauth-b','B')`,
		`INSERT INTO lti_routings VALUES (1,'lab','repo','check')`,
		`INSERT INTO lti_attempts (id,attempt_id,status,user_id,server_id,lti_routing_id,created_at) VALUES (1,'one','active',1,7,1,'2026-01-01'),(2,'two','completed',1,9,1,'2026-01-01'),(3,'three','pending',2,7,1,'2026-01-01'),(4,'four','pending',2,NULL,1,'2026-01-01'),(5,'five','active',2,9,1,'2026-01-01')`,
	} {
		if err := db.Exec(sql).Error; err != nil {
			t.Fatal(err)
		}
	}
	q := LTIAttemptQueries{DB: db}
	for _, tc := range []struct {
		filter LTIAttemptSearchParams
		want   []uint
	}{
		{LTIAttemptSearchParams{}, []uint{5, 4, 3, 2, 1}},
		{LTIAttemptSearchParams{Statuses: []string{"pending"}}, []uint{4, 3}},
		{LTIAttemptSearchParams{Statuses: []string{"active", "pending"}}, []uint{5, 4, 3, 1}},
		{LTIAttemptSearchParams{UserIds: []uint{1}}, []uint{2, 1}},
		{LTIAttemptSearchParams{UserIds: []uint{1, 2}, ServerIds: []uint{7, 9}}, []uint{5, 3, 2, 1}},
		{LTIAttemptSearchParams{UserIds: []uint{2}, Statuses: []string{"pending"}}, []uint{4, 3}},
		{LTIAttemptSearchParams{UserIds: []uint{2}, Statuses: []string{"pending"}, ServerIds: []uint{7}}, []uint{3}},
		{LTIAttemptSearchParams{UserIds: []uint{2}, Statuses: []string{"pending"}, ServerClientIds: []string{"oauth-a"}}, []uint{3}},
		{LTIAttemptSearchParams{UserIds: []uint{2}, Statuses: []string{"pending"}, ServerIds: []uint{9}}, nil},
		{LTIAttemptSearchParams{AttemptIds: []string{"one", "three"}}, []uint{3, 1}},
		{LTIAttemptSearchParams{AttemptIds: []string{"one", "three"}, UserIds: []uint{1}}, []uint{1}},
		{LTIAttemptSearchParams{ServerIds: []uint{7}, ServerClientIds: []string{"oauth-b"}}, nil},
	} {
		rows, count, err := q.List(&tc.filter)
		if err != nil {
			t.Fatal(err)
		}
		if count != int64(len(tc.want)) || len(rows) != len(tc.want) {
			t.Fatalf("filter=%+v count=%d rows=%v want=%v", tc.filter, count, rows, tc.want)
		}
		for i, r := range rows {
			if r.ID != tc.want[i] {
				t.Fatalf("filter=%+v row[%d]=%d want=%d", tc.filter, i, r.ID, tc.want[i])
			}
		}
	}
	for offset := 0; offset <= 5; offset++ {
		filter := LTIAttemptSearchParams{Limit: 1, Offset: offset}
		rows, count, err := q.List(&filter)
		if err != nil {
			t.Fatal(err)
		}
		if count != 5 {
			t.Fatal(count)
		}
		if offset == 5 {
			if len(rows) != 0 {
				t.Fatal(rows)
			}
		} else if len(rows) != 1 || rows[0].ID != uint(5-offset) {
			t.Fatalf("offset=%d rows=%v", offset, rows)
		}
	}
}
