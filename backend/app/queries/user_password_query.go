package queries

import (
	"fmt"
	"time"

	"github.com/cms-lab-core/cms-labs-api/shared/jsonrpc"

	"github.com/rs/zerolog"

	"github.com/cms-lab-core/cms-labs-api/backend/app/models"
	"github.com/ory/go-convenience/stringsx"
	"gorm.io/gorm"
)

type UserPasswordQueries struct {
	DB     *gorm.DB
	Logger *zerolog.Logger
}

var IncorrectPassword = jsonrpc.NewRpcError("incorrect_password", "Incorrect password")
var IncorrectLogin = jsonrpc.NewRpcError("incorrect_login", "Incorrect login")

func (q *UserPasswordQueries) Get(userID uint) (models.UserPassword, error) {
	var entity models.UserPassword
	q.DB.Where("`user_id` = ?", userID).Find(&entity)
	if entity.UserID != userID {
		return entity, IncorrectPassword
	}
	return entity, nil
}

func (q *UserPasswordQueries) Upsert(entity *models.UserPassword) error {
	if entity == nil {
		return nil
	}
	entityDB := models.UserPassword{}
	q.DB.Where("`user_id` = ?", entity.UserID).Find(&entityDB)
	if entityDB.UserID == entity.UserID {
		// Update
		q.Logger.Info().Msg(fmt.Sprintf("UserPasswordQueries: entity update user_id=%+v", entityDB.UserID))
		entity.HashPassword = stringsx.Coalesce(entity.HashPassword, entityDB.HashPassword)
		entity.CreatedAt = entityDB.CreatedAt
		entity.UpdatedAt = time.Now().UTC()
		result := q.DB.Save(&entity)
		return result.Error
	} else {
		// Create
		q.Logger.Info().Msg(fmt.Sprintf("UserPasswordQueries: entity create user_id=%+v", entity.UserID))
		entity.CreatedAt = time.Now().UTC()
		entity.UpdatedAt = time.Now().UTC()
		result := q.DB.Create(entity)
		return result.Error
	}
}
