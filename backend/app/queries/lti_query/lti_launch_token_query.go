package lti_query

import (
	"fmt"
	"time"

	"github.com/rs/zerolog"

	json "github.com/goccy/go-json"

	"github.com/cms-lab-core/cms-labs-api/shared/jsonrpc"

	"github.com/cms-lab-core/cms-labs-api/backend/app/models"
	"gorm.io/gorm"
)

type LTILaunchDataQueries struct {
	DB     *gorm.DB
	Logger *zerolog.Logger
}

func (q *LTILaunchDataQueries) Get(id string) (models.LTILaunchData, error) {
	var entity models.LTILaunchData
	result := q.DB.First(&entity, "id = ?", id)
	if result.Error != nil && result.Error.Error() == "record not found" {
		return entity, jsonrpc.NewRpcError("user_not_found", "lti launch data has not found")
	}
	return entity, result.Error
}

func (q *LTILaunchDataQueries) GetByAttemptId(id string) (models.LTILaunchData, error) {
	var entity models.LTILaunchData
	result := q.DB.Where("attempt_id = ?", id).Order(`created_at desc`).First(&entity)
	if result.Error != nil && result.Error.Error() == "record not found" {
		return entity, jsonrpc.NewRpcError("user_not_found", "lti launch data has not found")
	}
	return entity, result.Error
}

func (q *LTILaunchDataQueries) Upsert(entity *models.LTILaunchData) error {
	if entity == nil {
		return nil
	}
	entityDB := models.LTILaunchData{}
	if entity.ID != "" {
		q.DB.Where("id = ?", entity.ID).Find(&entityDB)
	}
	if entityDB.ID != "" {
		// Update
		q.Logger.Debug().Msg(fmt.Sprintf("LTILaunchDataQueries: entity update: %+v", entityDB))
		q.Logger.Info().Msg(fmt.Sprintf("LTILaunchDataQueries: entity update launch_id=%+v", entityDB.ID))
		entity.ID = entityDB.ID
		entity.CreatedAt = entityDB.CreatedAt
		entity.UpdatedAt = time.Now().UTC()
		result := q.DB.Save(&entity)
		return result.Error
	}
	// Create
	q.Logger.Debug().Msg(fmt.Sprintf("LTILaunchDataQueries: entity create: %+v", entity))
	q.Logger.Info().Msg(fmt.Sprintf("LTILaunchDataQueries: entity create launch_id=%+v", entity.ID))
	entity.CreatedAt = time.Now().UTC()
	entity.UpdatedAt = time.Now().UTC()
	result := q.DB.Create(entity)
	return result.Error
}

// StoreLaunchData stores the JSON launch data associated with the supplied launch ID.
func (q *LTILaunchDataQueries) StoreLaunchData(launchID string, launchData json.RawMessage, regUid uint) error {
	q.Logger.Info().Msg(fmt.Sprintf("LTILaunchDataQueries: StoreLaunchData by launch_id=%+v", launchID))
	entity := models.LTILaunchData{}
	entity.ID = launchID
	entity.LaunchData = string(launchData)
	entity.LTIFormID = regUid
	return q.Upsert(&entity)
}

// FindLaunchData retrieves previously-stored launch data using the `launchID'. If the launch data cannot be
// found, it returns ErrLaunchDataNotFound.
func (q *LTILaunchDataQueries) FindLaunchData(launchID string) (json.RawMessage, error) {
	q.Logger.Info().Msg(fmt.Sprintf("LTILaunchDataQueries: FindLaunchData by launch_id: %+v", launchID))
	entity, err := q.Get(launchID)
	if err != nil {
		return []byte{}, err
	}
	return []byte(entity.LaunchData), nil
}
