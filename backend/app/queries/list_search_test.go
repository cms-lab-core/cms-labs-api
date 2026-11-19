package queries_test

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/cms-lab-core/cms-labs-api/backend/app/queries"
	"github.com/cms-lab-core/cms-labs-api/backend/app/queries/lti_query"
	"github.com/cms-lab-core/cms-labs-api/backend/app/usecases"
	"github.com/rs/zerolog"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func searchDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { sqlDB.Close() })
	for _, sql := range []string{
		`CREATE TABLE users (id integer PRIMARY KEY, email text, name text, created_at datetime, updated_at datetime, group_name text, lti_user_id text, deleted_at datetime)`,
		`CREATE TABLE servers (id integer PRIMARY KEY, name text, url text, client_id text, type text, is_active boolean, created_at datetime, unit_rate integer, last_count_users integer)`,
		`CREATE TABLE lti_routings (id integer PRIMARY KEY, name text, lti_title text, lti_description text, labs_path text, test_path text, created_at datetime, updated_at datetime)`,
		`CREATE TABLE service_cards (id integer PRIMARY KEY, name text, description text, url text, is_active boolean, "order" integer, created_at datetime, updated_at datetime, image_url text)`,
		`CREATE TABLE auth_providers (id integer PRIMARY KEY, name text, base_uri text, type text, sso_url text, created_at datetime, updated_at datetime)`,
		`INSERT INTO users (id,email,name,created_at) VALUES (11,'alpha@example.test','Alpha_100%','2026-01-01'),(12,'beta@example.test','Beta!','2026-01-01'),(13,'gamma@example.test','лаборатория','2026-01-01')`,
		`INSERT INTO servers (id,name,url,client_id,type,is_active,created_at,unit_rate,last_count_users) VALUES (11,'Alpha_100%','https://alpha.test','oauth-alpha','k8s',true,'2026-01-01',10,1),(12,'Beta!','https://beta.test','oauth-beta','pnet',false,'2026-01-01',30,2),(13,'лаборатория','https://gamma.test','oauth-gamma','openid',true,'2026-01-01',20,3)`,
		`INSERT INTO lti_routings (id,name,lti_title,lti_description,labs_path,test_path,created_at) VALUES (11,'Alpha_100%','OSPF title','Routing description','https://repo.test/alpha','check-alpha','2026-01-01'),(12,'Beta!','CN title','Switching description','https://repo.test/beta','check-beta','2026-01-01'),(13,'лаборатория','Third','Third description','https://repo.test/gamma','check-gamma','2026-01-01')`,
		`INSERT INTO service_cards (id,name,description,url,is_active,"order",created_at) VALUES (11,'Alpha_100%','Routing service','https://alpha.test',true,0,'2026-01-01'),(12,'Beta!','Switching service','https://beta.test',false,0,'2026-01-01'),(13,'лаборатория','Third service','https://gamma.test',true,0,'2026-01-01')`,
		`INSERT INTO auth_providers (id,name,base_uri,type,sso_url,created_at) VALUES (11,'Alpha_100%','https://alpha.test','lti','https://sso.test','2026-01-01'),(12,'Beta!','https://beta.test','lti','','2026-01-01'),(13,'лаборатория','https://gamma.test','ldap',NULL,'2026-01-01')`,
	} {
		if err := db.Exec(sql).Error; err != nil {
			t.Fatal(err)
		}
	}
	return db
}

func TestListSearchAndPagination(t *testing.T) {
	db := searchDB(t)
	logger := zerolog.Nop()
	user := queries.UserQueries{DB: db, Logger: &logger}
	server := queries.ServerQueries{DB: db, Logger: &logger}
	routing := queries.LTIRoutingQueries{DB: db, Logger: &logger}
	service := queries.ServiceCardQueries{DB: db, Logger: &logger}
	provider := lti_query.AuthProviderQueries{DB: db, Logger: &logger}
	type listFunc func(string, int, int) ([]uint, int64, error)
	lists := map[string]listFunc{
		"users": func(search string, limit, offset int) (ids []uint, count int64, err error) {
			rows, count, err := user.List(search, nil, limit, offset)
			for _, r := range rows {
				ids = append(ids, r.ID)
			}
			return ids, count, err
		},
		"servers": func(search string, limit, offset int) (ids []uint, count int64, err error) {
			rows, count, err := server.List(queries.ServerQueriesListDTO{Search: search, Limit: limit, Offset: offset})
			for _, r := range rows {
				ids = append(ids, r.ID)
			}
			return ids, count, err
		},
		"routings": func(search string, limit, offset int) (ids []uint, count int64, err error) {
			rows, count, err := routing.List(search, limit, offset)
			for _, r := range rows {
				ids = append(ids, r.ID)
			}
			return ids, count, err
		},
		"services": func(search string, limit, offset int) (ids []uint, count int64, err error) {
			rows, count, err := service.List(search, limit, offset)
			for _, r := range rows {
				ids = append(ids, r.ID)
			}
			return ids, count, err
		},
		"providers": func(search string, limit, offset int) (ids []uint, count int64, err error) {
			rows, count, err := provider.List(&lti_query.AuthProviderListInputDTO{Search: search, Limit: limit, Offset: offset})
			for _, r := range rows {
				ids = append(ids, r.ID)
			}
			return ids, count, err
		},
	}
	for name, list := range lists {
		t.Run(name, func(t *testing.T) {
			for _, tc := range []struct {
				search string
				want   []uint
			}{
				{"", []uint{13, 12, 11}}, {"   ", []uint{13, 12, 11}}, {" aLPhA ", []uint{11}}, {"B", []uint{12}}, {"11", []uint{11}}, {"1nope", nil}, {"not-found", nil}, {"%", []uint{11}}, {"_", []uint{11}}, {"!", []uint{12}}, {"ЛАБОРАТОРИЯ", []uint{13}}, {"' OR 1=1 --", nil},
			} {
				t.Run(tc.search, func(t *testing.T) {
					ids, count, err := list(tc.search, 0, -1)
					if err != nil {
						t.Fatal(err)
					}
					if !reflect.DeepEqual(ids, tc.want) || count != int64(len(tc.want)) {
						t.Fatalf("ids=%v count=%d want=%v", ids, count, tc.want)
					}
				})
			}
			for _, offset := range []int{0, 1, 2, 3, 100} {
				ids, count, err := list("", 1, offset)
				if err != nil {
					t.Fatal(err)
				}
				var want []uint
				if offset < 3 {
					want = []uint{uint(13 - offset)}
				}
				if !reflect.DeepEqual(ids, want) || count != 3 {
					t.Fatalf("offset=%d ids=%v count=%d want=%v", offset, ids, count, want)
				}
			}
		})
	}
	for _, tc := range []struct {
		name, search string
		want         uint
	}{{"users", "example.test", 13}, {"servers", "oauth-beta", 12}, {"servers", "https://alpha", 11}, {"routings", "OSPF", 11}, {"routings", "Switching", 12}, {"routings", "check-beta", 12}, {"services", "Routing service", 11}, {"services", "https://beta", 12}, {"providers", "ldap", 13}, {"providers", "sso.test", 11}} {
		ids, _, err := lists[tc.name](tc.search, 100, 0)
		if err != nil {
			t.Fatal(err)
		}
		if len(ids) == 0 || ids[0] != tc.want {
			t.Fatalf("%s %q: %v", tc.name, tc.search, ids)
		}
	}
}

func TestSearchRespectsOtherFilters(t *testing.T) {
	db := searchDB(t)
	logger := zerolog.Nop()
	user := queries.UserQueries{DB: db, Logger: &logger}
	for _, ids := range [][]uint{{11}, {12}, {11, 12}} {
		rows, count, err := user.List("alpha", ids, 100, 0)
		if err != nil {
			t.Fatal(err)
		}
		want := int64(1)
		if len(ids) == 1 && ids[0] == 12 {
			want = 0
		}
		if int64(len(rows)) != want || count != want {
			t.Fatalf("user ids=%v rows=%v count=%d", ids, rows, count)
		}
	}
	server := queries.ServerQueries{DB: db, Logger: &logger}
	for _, tc := range []struct {
		types  []string
		search string
		want   int64
	}{{[]string{"k8s"}, "alpha", 1}, {[]string{"pnet"}, "alpha", 0}, {[]string{"pnet", "k8s"}, "", 2}, {[]string{"openid"}, "gamma", 1}} {
		rows, count, err := server.List(queries.ServerQueriesListDTO{Types: tc.types, Search: tc.search})
		if err != nil {
			t.Fatal(err)
		}
		if int64(len(rows)) != tc.want || count != tc.want {
			t.Fatalf("server filter=%+v count=%d rows=%d", tc, count, len(rows))
		}
	}
	service := queries.ServiceCardQueries{DB: db, Logger: &logger}
	for _, tc := range []struct {
		search string
		want   int64
	}{{"alpha", 1}, {"beta", 0}, {"12", 0}, {"service", 2}} {
		rows, count, err := service.List(tc.search, 100, 0, true)
		if err != nil {
			t.Fatal(err)
		}
		if int64(len(rows)) != tc.want || count != tc.want {
			t.Fatalf("active services search=%q count=%d rows=%d", tc.search, count, len(rows))
		}
	}
	provider := lti_query.AuthProviderQueries{DB: db, Logger: &logger}
	for _, tc := range []struct {
		search string
		want   int64
	}{{"", 2}, {"alpha", 1}, {"beta", 0}, {"12", 0}, {"ldap", 1}, {"test", 2}} {
		rows, count, err := provider.List(&lti_query.AuthProviderListInputDTO{Search: tc.search, IsAuth: true})
		if err != nil {
			t.Fatal(err)
		}
		if int64(len(rows)) != tc.want || count != tc.want {
			t.Fatalf("auth providers search=%q count=%d rows=%d", tc.search, count, len(rows))
		}
	}
	routing := queries.LTIRoutingQueries{DB: db, Logger: &logger}
	for _, search := range []string{" alpha ", "ROUTING", "repo.test/alpha", "11", "%", "_"} {
		rows, err := routing.ListLabs(search)
		if err != nil {
			t.Fatal(err)
		}
		if len(rows) != 1 || rows[0].ID != 11 {
			t.Fatalf("catalog search=%q rows=%v", search, rows)
		}
	}
}

func TestEmptySearchResultsUseArrays(t *testing.T) {
	db := searchDB(t)
	if err := db.Exec("CREATE TABLE role_relations (id integer PRIMARY KEY, user_id integer, server_id integer, role_id integer)").Error; err != nil {
		t.Fatal(err)
	}
	logger := zerolog.Nop()
	roles := queries.RoleQueries{DB: db, Logger: &logger}
	users := usecases.UserListUC{UserQueries: &queries.UserQueries{DB: db, Logger: &logger}, RoleQueries: &roles}
	userResult, err := users.Execute(usecases.UserListInputDTO{Search: "not-found"})
	if err != nil {
		t.Fatal(err)
	}
	servers := usecases.ServerListUC{ServerQueries: &queries.ServerQueries{DB: db, Logger: &logger}, RoleQueries: &roles}
	serverResult, err := servers.Execute(queries.ServerQueriesListDTO{Search: "not-found"})
	if err != nil {
		t.Fatal(err)
	}
	for _, result := range []interface{}{userResult, serverResult} {
		encoded, err := json.Marshal(result)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(encoded), `"model":[]`) || !strings.Contains(string(encoded), `"total_count":0`) {
			t.Fatalf("unexpected empty response: %s", encoded)
		}
	}
}
