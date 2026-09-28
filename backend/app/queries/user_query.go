package queries

import (
	"fmt"
	"time"

	"github.com/maintainer64/cms-labs-api/shared/jsonrpc"

	"github.com/maintainer64/cms-labs-api/backend/app/models/types"

	"github.com/rs/zerolog"

	"github.com/maintainer64/cms-labs-api/backend/app/models"
	"gorm.io/gorm"
)

var (
	UserNotFoundError = jsonrpc.NewRpcError("user_not_found", "user has not found")
	UserNotActive     = jsonrpc.NewRpcError("user_is_deactivated", "user is deactivated")
)

type UserQueries struct {
	DB     *gorm.DB
	Logger *zerolog.Logger
}

func (q *UserQueries) Get(id uint) (models.User, error) {
	var entity models.User
	result := q.DB.First(&entity, id)
	if result.Error != nil && result.Error.Error() == "record not found" {
		return entity, UserNotFoundError
	}
	if !entity.IsActive() {
		return entity, UserNotActive
	}
	return entity, result.Error
}

func (q *UserQueries) Upsert(entity *models.User) error {
	if entity == nil {
		return nil
	}
	entityDB := models.User{}
	q.DB.Where("id = ?", entity.ID).Find(&entityDB)
	if entityDB.ID != 0 {
		// Update
		q.Logger.Debug().Msg(fmt.Sprintf("UserQueries: entity update: %+v", entityDB))
		q.Logger.Info().Msg(fmt.Sprintf("UserQueries: entity update id=%+v", entityDB.ID))
		entity.ID = entityDB.ID
		entity.CreatedAt = entityDB.CreatedAt
		entity.UpdatedAt = time.Now().UTC()
		result := q.DB.Save(&entity)
		return result.Error
	} else {
		// Create
		q.Logger.Debug().Msg(fmt.Sprintf("UserQueries: entity create: %+v", entity))
		q.Logger.Info().Msg(fmt.Sprintf("UserQueries: entity create email=%+v", entity.Email))
		entity.ID = 0
		entity.CreatedAt = time.Now().UTC()
		entity.UpdatedAt = time.Now().UTC()
		result := q.DB.Create(entity)
		return result.Error
	}
}

func (q *UserQueries) GetByEmail(email string) (models.User, error) {
	var entity models.User
	q.Logger.Info().Msg(fmt.Sprintf("UserQueries: get user by email=%+v", email))
	if email == "" {
		return entity, UserNotFoundError
	}
	q.DB.Where("email = ?", email).Limit(1).Find(&entity)
	if entity.Email != email {
		return entity, UserNotFoundError
	}
	if !entity.IsActive() {
		q.Logger.Info().Msg(fmt.Sprintf("UserQueries: user is not active by email=%+v", email))
		return entity, UserNotActive
	}
	return entity, nil
}

func (q *UserQueries) GetByLaunchID(launchID string) (models.User, error) {
	var entity models.User
	q.Logger.Info().Msg(fmt.Sprintf("UserQueries: get user by last_launch_id=%+v", launchID))
	if launchID == "" {
		return entity, UserNotFoundError
	}
	q.DB.Where("last_launch_id = ?", launchID).Limit(1).Find(&entity)
	q.Logger.Info().Msg(fmt.Sprintf(
		"UserQueries: get user by last_launch_id=%+v fetched id=%+v",
		launchID,
		entity.ID,
	))
	if entity.LastLaunchID != launchID {
		return entity, UserNotFoundError
	}
	if !entity.IsActive() {
		q.Logger.Info().Msg(fmt.Sprintf("UserQueries: user is not active by launchID=%+v", launchID))
		return entity, UserNotFoundError
	}
	return entity, nil
}

func (q *UserQueries) List(
	search string,
	ids []uint,
	limit int,
	offset int,
) ([]models.UserListItem, int64, error) {
	var entities []models.UserListItem
	result := q.listFilter(
		search,
		ids,
		q.DB.Limit(MaxLimitCount).Offset(0),
	).Find(&entities)
	count := result.RowsAffected
	q.Logger.Debug().Msg(fmt.Sprintf("UserQueries list: count %+v", count))
	result = q.listFilter(
		search,
		ids,
		q.DB.Limit(limit).Offset(offset),
	).Find(&entities)
	q.Logger.Debug().Msg(fmt.Sprintf("UserQueries list: entities %+v", entities))
	return entities, count, result.Error
}

func (q *UserQueries) listFilter(search string, ids []uint, tx *gorm.DB) *gorm.DB {
	tx = tx.Model(&models.User{})
	// Keep pagination deterministic when multiple users are created within the
	// same database timestamp precision window.
	tx = tx.Order(`created_at desc`).Order(`id desc`)
	if search != "" {
		tx = tx.Or("email LIKE ?", fmt.Sprintf("%%%s%%", search))
		tx = tx.Or("name LIKE ?", fmt.Sprintf("%%%s%%", search))
		tx = tx.Or("id = ?", search)
	}
	if len(ids) > 0 {
		tx = tx.Or("id IN ?", ids)
	}
	return tx
}

func (q *UserQueries) StoreGetByUserId(id uint) (types.JsonStore, error) {
	userDB, err := q.Get(id)
	if err != nil {
		return types.JsonStore{}, err
	}
	return userDB.Store, nil
}

func (q *UserQueries) StoreSetByUserId(id uint, store types.JsonStore) error {
	userDB, err := q.Get(id)
	if err != nil {
		return err
	}
	userDB.Store = store
	err = q.Upsert(&userDB)
	return err
}
