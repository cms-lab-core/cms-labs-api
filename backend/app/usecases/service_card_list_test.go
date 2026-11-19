package usecases

import (
	"testing"

	"github.com/cms-lab-core/cms-labs-api/backend/app/models"
	"github.com/cms-lab-core/cms-labs-api/backend/app/queries"
	"github.com/rs/zerolog"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestServiceCardListsDefaultLimitAndActiveVisibility(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	defer sqlDB.Close()
	if err := db.Exec(`CREATE TABLE service_cards (id integer PRIMARY KEY, created_at datetime, updated_at datetime, image_url text, url text, name text, description text, "order" integer, is_active boolean)`).Error; err != nil {
		t.Fatal(err)
	}
	for i, active := range []bool{true, false, true} {
		card := models.ServiceCard{}
		card.Name = "service"
		card.Url = "https://example.test"
		card.IsActive = active
		card.Order = uint(i)
		if err := db.Create(&card).Error; err != nil {
			t.Fatal(err)
		}
	}
	logger := zerolog.Nop()
	uc := ServiceCardListUC{ServiceCardQueries: &queries.ServiceCardQueries{DB: db, Logger: &logger}}
	all, err := uc.Execute(ServiceCardListInputDTO{})
	if err != nil {
		t.Fatal(err)
	}
	if len(all.Model) != 3 || all.TotalCount != 3 {
		t.Fatalf("default list: %+v", all)
	}
	active, err := uc.Execute(ServiceCardListInputDTO{OnlyActive: true, Limit: 1})
	if err != nil {
		t.Fatal(err)
	}
	if len(active.Model) != 1 || active.TotalCount != 2 || !active.Model[0].IsActive || active.Model[0].Order != 2 {
		t.Fatalf("active list: %+v", active)
	}
}
