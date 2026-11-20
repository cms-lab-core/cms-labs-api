package tasks

import (
	"net"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/cms-lab-core/cms-labs-api/backend/app/models"
	"github.com/cms-lab-core/cms-labs-api/backend/app/queries"
	drivermysql "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

func TestGradeDeliverySerializesAcrossDatabaseConnections(t *testing.T) {
	if os.Getenv("CMS_GRADE_TEST_MYSQL") != "1" {
		t.Skip("enable CMS_GRADE_TEST_MYSQL with a test database")
	}
	cfg := drivermysql.NewConfig()
	cfg.User = os.Getenv("DB_USER")
	cfg.Passwd = os.Getenv("DB_PASSWORD")
	cfg.Net = "tcp"
	cfg.Addr = net.JoinHostPort(os.Getenv("DB_HOST"), os.Getenv("DB_PORT"))
	cfg.DBName = os.Getenv("DB_NAME")
	cfg.ParseTime = true
	prefix := "grade_outbox_" + strings.ReplaceAll(uuid.NewString(), "-", "")[:10] + "_"
	open := func() *gorm.DB {
		db, err := gorm.Open(mysql.Open(cfg.FormatDSN()), &gorm.Config{NamingStrategy: schema.NamingStrategy{TablePrefix: prefix}})
		if err != nil {
			t.Fatal(err)
		}
		pool, err := db.DB()
		if err != nil {
			t.Fatal(err)
		}
		pool.SetMaxOpenConns(1)
		t.Cleanup(func() { pool.Close() })
		return db
	}
	first, second, third := open(), open(), open()
	if err := first.AutoMigrate(&models.LTIAttempt{}); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { first.Migrator().DropTable(&models.LTIAttempt{}) })
	serverID := uint(7)
	at := time.Now().UTC().Add(-time.Minute)
	grade := datatypes.NewJSONType(models.LTIAttemptResult{CheckID: "old", CheckedAt: at, MaxScore: 10, CurrentScore: 1, ResultDisplay: "old"})
	attempt := models.LTIAttempt{LTIAttemptBase: models.LTIAttemptBase{AttemptID: "concurrent", Status: "active", ServerID: &serverID, Result: &grade}}
	if err := first.Create(&attempt).Error; err != nil {
		t.Fatal(err)
	}
	oldStarted, releaseOld, oldDone := make(chan struct{}), make(chan struct{}), make(chan error, 1)
	var releaseOnce sync.Once
	release := func() { releaseOnce.Do(func() { close(releaseOld) }) }
	defer release()
	go func() {
		uc := LTISyncResultUC{LTIAttemptQueries: &queries.LTIAttemptQueries{DB: first}}
		oldDone <- uc.syncGrade("concurrent", func(*gorm.DB, models.LTIAttempt) error { close(oldStarted); <-releaseOld; return nil })
	}()
	select {
	case <-oldStarted:
	case <-time.After(3 * time.Second):
		t.Fatal("first sender did not acquire row")
	}
	competing := make(chan error, 1)
	go func() {
		uc := LTISyncResultUC{LTIAttemptQueries: &queries.LTIAttemptQueries{DB: third}}
		competing <- uc.syncGrade("concurrent", func(*gorm.DB, models.LTIAttempt) error { t.Error("parallel sender bypassed row lock"); return nil })
	}()
	select {
	case err := <-competing:
		if err == nil {
			t.Fatal("parallel sender should defer to the active sender")
		}
	case <-time.After(3 * time.Second):
		t.Fatal("parallel sender blocked instead of leaving grade queued")
	}
	newer := datatypes.NewJSONType(models.LTIAttemptResult{CheckID: "new", CheckedAt: at.Add(time.Second), MaxScore: 10, CurrentScore: 8, ResultDisplay: "latest"})
	written := make(chan error, 1)
	go func() {
		_, _, err := (&queries.LTIAttemptQueries{DB: second}).ApplyExternalUpdate("concurrent", 7, "completed", &newer)
		written <- err
	}()
	select {
	case err := <-written:
		t.Fatalf("new grade bypassed in-flight sender: %v", err)
	case <-time.After(100 * time.Millisecond):
	}
	release()
	select {
	case err := <-oldDone:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("sender stuck")
	}
	select {
	case err := <-written:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("latest grade stuck")
	}
	q := queries.LTIAttemptQueries{DB: second}
	stored, err := q.GetByAttemptID("concurrent")
	if err != nil {
		t.Fatal(err)
	}
	if stored.Result.Data().CheckID != "new" || stored.Status != "completed" || stored.SynchronizedAt != nil {
		t.Fatalf("old ACK lost latest grade: %+v", stored)
	}
	uc := LTISyncResultUC{LTIAttemptQueries: &queries.LTIAttemptQueries{DB: third}}
	if err := uc.syncGrade("concurrent", func(_ *gorm.DB, a models.LTIAttempt) error {
		if a.Result.Data().CurrentScore != 8 || a.Status != "completed" {
			t.Error("latest sender got stale grade")
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	stored, err = q.GetByAttemptID("concurrent")
	if err != nil || stored.SynchronizedAt == nil || stored.Result.Data().CheckID != "new" {
		t.Fatal("final grade did not survive", err)
	}
}
