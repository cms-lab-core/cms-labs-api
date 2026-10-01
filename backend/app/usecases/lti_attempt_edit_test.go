package usecases

import (
	"testing"

	"github.com/maintainer64/cms-labs-api/backend/app/models"
)

func TestApplyLTIAttemptEditPreservesServerWhenOmitted(t *testing.T) {
	serverID := uint(17)
	entity := models.LTIAttempt{}
	entity.ServerID = &serverID

	applyLTIAttemptEdit(&entity, LTIAttemptEditInputDTO{
		ID:     5,
		Status: models.AttemptStatusTerminating,
	})

	if entity.ID != 5 || entity.Status != models.AttemptStatusTerminating {
		t.Fatalf("unexpected edited attempt: %#v", entity)
	}
	if entity.ServerID == nil || *entity.ServerID != serverID {
		t.Fatalf("server ID changed: %#v", entity.ServerID)
	}
}

func TestApplyLTIAttemptEditUpdatesExplicitServer(t *testing.T) {
	originalServerID := uint(17)
	newServerID := uint(23)
	entity := models.LTIAttempt{}
	entity.ServerID = &originalServerID

	applyLTIAttemptEdit(&entity, LTIAttemptEditInputDTO{
		ID:       5,
		ServerID: &newServerID,
		Status:   models.AttemptStatusActive,
	})

	if entity.ServerID == nil || *entity.ServerID != newServerID {
		t.Fatalf("server ID was not updated: %#v", entity.ServerID)
	}
}
