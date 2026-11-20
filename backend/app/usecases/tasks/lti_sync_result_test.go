package tasks

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/cms-lab-core/cms-labs-api/backend/app/models"
	"github.com/cms-lab-core/cms-labs-api/backend/app/queries"
	"github.com/cms-lab-core/cms-labs-api/backend/app/queries/lti_query"
	"github.com/cms-lab-core/cms-labs-api/backend/app/usecases/lti_connector"
	"github.com/cms-lab-core/cms-labs-api/backend/app/usecases/lti_connector/connector"
	"github.com/rs/zerolog"
	"gorm.io/datatypes"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestGradeDeliveryRetriesAfterFailureAndRestart(t *testing.T) {
	for _, score := range []float64{0, 8} {
		t.Run(map[float64]string{0: "zero", 8: "partial"}[score], func(t *testing.T) {
			db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
			if err != nil {
				t.Fatal(err)
			}
			sqlDB, err := db.DB()
			if err != nil {
				t.Fatal(err)
			}
			sqlDB.SetMaxOpenConns(1)
			defer sqlDB.Close()

			for _, fixture := range []struct {
				model   interface{}
				columns string
			}{
				{&models.LTIAttempt{}, `id integer PRIMARY KEY,created_at datetime,updated_at datetime,attempt_id text,status text,result text,user_id integer,server_id integer,lti_routing_id integer,synchronized_at datetime,room_id integer`},
				{&models.AuthProvider{}, `id integer PRIMARY KEY,created_at datetime,updated_at datetime,name text,type text,sso_url text,base_uri text,lti_client_id text,lti_deployment_id text,lti_auth_token_uri text,lti_auth_login_uri text,key_set_uri text,target_link_uri text,public_key text,private_key text`},
				{&models.LTILaunchData{}, `id text PRIMARY KEY,created_at datetime,updated_at datetime,lti_form_id integer,attempt_id text,launch_data text`},
				{&models.LTIAccessToken{}, `id integer PRIMARY KEY,created_at datetime,updated_at datetime,"index" text,payload text`},
			} {
				stmt := &gorm.Statement{DB: db}
				if err := stmt.Parse(fixture.model); err != nil {
					t.Fatal(err)
				}
				if err := db.Exec("CREATE TABLE " + stmt.Schema.Table + " (" + fixture.columns + ")").Error; err != nil {
					t.Fatal(err)
				}
			}

			if err := db.Exec(`CREATE TABLE servers(id integer PRIMARY KEY,type text)`).Error; err != nil {
				t.Fatal(err)
			}
			var mutex sync.Mutex
			var sent []connector.Score
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != "POST" || r.URL.Path != "/lineitem/scores" {
					http.NotFound(w, r)
					return
				}
				var grade connector.Score
				if err := json.NewDecoder(r.Body).Decode(&grade); err != nil {
					t.Error(err)
					w.WriteHeader(400)
					return
				}
				mutex.Lock()
				sent = append(sent, grade)
				count := len(sent)
				mutex.Unlock()
				if count == 1 {
					http.Error(w, "try again", 503)
					return
				}
				w.WriteHeader(http.StatusNoContent)
			}))
			defer server.Close()
			key, err := rsa.GenerateKey(rand.Reader, 2048)
			if err != nil {
				t.Fatal(err)
			}
			private := string(pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)}))
			provider := models.AuthProvider{AuthProviderBase: models.AuthProviderBase{Name: "test", Type: "lti"}, AuthProviderSecret: models.AuthProviderSecret{BaseURI: server.URL, LTIClientID: "client", LTIAuthTokenURI: server.URL + "/token", PrivateKey: private}}
			if err := db.Create(&provider).Error; err != nil {
				t.Fatal(err)
			}
			attemptID := "retry-attempt"
			claims := map[string]interface{}{"iss": server.URL, "aud": []string{"client"}, "sub": "student", "https://purl.imsglobal.org/spec/lti-ags/claim/endpoint": map[string]interface{}{"lineitem": server.URL + "/lineitem", "lineitems": server.URL + "/lineitems", "scope": []string{"https://purl.imsglobal.org/spec/lti-ags/scope/score"}}}
			raw, err := json.Marshal(claims)
			if err != nil {
				t.Fatal(err)
			}
			launch := models.LTILaunchData{LTILaunchDataBase: models.LTILaunchDataBase{ID: "launch", LTIFormID: provider.ID, AttemptID: &attemptID}, LTILaunchDataSecret: models.LTILaunchDataSecret{LaunchData: string(raw)}}
			if err := db.Create(&launch).Error; err != nil {
				t.Fatal(err)
			}
			grade := datatypes.NewJSONType(models.LTIAttemptResult{Source: models.GradeSourceClabgate, CheckID: "check", MaxScore: 10, CurrentScore: score, ResultDisplay: "ready"})
			attempt := models.LTIAttempt{LTIAttemptBase: models.LTIAttemptBase{AttemptID: attemptID, Status: "active", Result: &grade}}
			if err := db.Create(&attempt).Error; err != nil {
				t.Fatal(err)
			}
			logger := zerolog.Nop()
			tokens := &lti_query.LTIAccessTokenQueries{DB: db, Logger: &logger}
			if err := tokens.StoreAccessToken(lti_query.AccessToken{TokenURI: server.URL + "/token", ClientID: "client", Scopes: []string{"https://purl.imsglobal.org/spec/lti-ags/scope/score"}, Token: "test-only", ExpiryTime: time.Now().Add(time.Hour)}); err != nil {
				t.Fatal(err)
			}
			providerQ := &lti_query.AuthProviderQueries{DB: db, Logger: &logger}
			launchQ := &lti_query.LTILaunchDataQueries{DB: db, Logger: &logger}
			api := &lti_connector.LTIConnectorAPI{AuthProviderQueries: providerQ, LTILaunchDataQueries: launchQ, LTIProtocolDatastoreConfig: &lti_query.Config{Registrations: providerQ, LaunchData: launchQ, AccessTokens: tokens}, Logger: &logger}
			newWorker := func() *LTISyncResultUC {
				return &LTISyncResultUC{LTIAttemptQueries: &queries.LTIAttemptQueries{DB: db}, LTIConnectorAPI: api, Logger: &logger}
			}
			ctx, cancel := context.WithCancel(context.Background())
			done := make(chan struct{})
			go func() { defer close(done); newWorker().Run(ctx) }()
			WakeGradeSync(attemptID)
			deadline := time.Now().Add(3 * time.Second)
			for {
				mutex.Lock()
				count := len(sent)
				mutex.Unlock()
				if count == 1 {
					break
				}
				if time.Now().After(deadline) {
					cancel()
					t.Fatal("grade event was not delivered immediately")
				}
				time.Sleep(5 * time.Millisecond)
			}
			cancel()
			<-done
			stored, err := newWorker().LTIAttemptQueries.GetByAttemptID(attemptID)
			if err != nil || stored.SynchronizedAt != nil {
				t.Fatal("failed event delivery removed pending state", err)
			}
			idle, stop := context.WithCancel(context.Background())
			stopped := make(chan struct{})
			go func() { defer close(stopped); newWorker().Run(idle) }()
			time.Sleep(100 * time.Millisecond)
			stop()
			<-stopped
			mutex.Lock()
			countBeforeCron := len(sent)
			mutex.Unlock()
			if countBeforeCron != 1 {
				t.Fatal("backend scanned persisted grades on startup")
			}
			if err := newWorker().Execute(); err != nil {
				t.Fatal(err)
			}
			stored, err = newWorker().LTIAttemptQueries.GetByAttemptID(attemptID)
			if err != nil || stored.SynchronizedAt == nil {
				t.Fatal("cron did not deliver persisted grade", err)
			}
			mutex.Lock()
			got := append([]connector.Score(nil), sent...)
			mutex.Unlock()
			if len(got) != 2 {
				t.Fatalf("sent %d times, want initial failure and retry", len(got))
			}
			if got[0].Timestamp != got[1].Timestamp {
				t.Fatal("retry changed grade identity timestamp")
			}
			for _, g := range got {
				if g.ScoreGiven != score || g.GradingProgress != connector.GradingFullyGraded || g.ActivityProgress != connector.ActivityInProgress || g.UserID != "student" {
					t.Fatalf("wrong ready grade: %+v", g)
				}
			}
			if err := newWorker().SyncGradeToLTI(attemptID); err != nil {
				t.Fatal(err)
			}
			mutex.Lock()
			count := len(sent)
			mutex.Unlock()
			if count != 2 {
				t.Fatal("acknowledged grade was sent again")
			}
		})
	}
}
