package queries

import (
	"fmt"
	"time"

	"github.com/cms-lab-core/cms-labs-api/shared/jsonrpc"

	"github.com/cms-lab-core/cms-labs-api/backend/app/models"
	"github.com/rs/zerolog"
	gorm "gorm.io/gorm"
)

type RoleQueries struct {
	DB     *gorm.DB
	Logger *zerolog.Logger
}

var (
	RoleNotFoundError = jsonrpc.NewRpcError("role_not_found", "role not found")
)

func (q *RoleQueries) GetRolesByUserId(userId uint) ([]models.Role, error) {
	var entities []models.Role
	subquery := q.DB.Model(&models.RoleRelation{}).Where("user_id = ?", userId).Select("role_id")
	result := q.DB.Where("id in (?)", subquery).Find(&entities)
	q.Logger.Info().Msg(
		fmt.Sprintf("RoleQueries: get roles by user_id = %d, count = %d", userId, len(entities)),
	)
	return entities, result.Error
}

func (q *RoleQueries) GetByRelationUsersIds(usersIds []uint) (map[uint][]uint, error) {
	// Возвращает список связей между userID и RoleId
	var entities []models.RoleRelation
	var hmap = make(map[uint][]uint)
	result := q.DB.Where("user_id in (?)", usersIds).Find(&entities)
	q.Logger.Info().Msg(
		fmt.Sprintf("RoleQueries: get roles by usersIds: %d, count: %d", len(usersIds), len(entities)),
	)
	if result.Error != nil {
		return hmap, result.Error
	}
	for _, entity := range entities {
		if entity.UserID == nil || entity.RoleID == 0 {
			continue
		}
		if val, ok := hmap[*entity.UserID]; ok {
			hmap[*entity.UserID] = append(val, entity.RoleID)
		} else {
			hmap[*entity.UserID] = []uint{entity.RoleID}
		}
	}
	return hmap, nil
}

func (q *RoleQueries) GetByRelationServerIds(serverIds []uint) (map[uint][]uint, error) {
	// Возвращает список связей между serverId и RoleId
	var entities []models.RoleRelation
	var hmap = make(map[uint][]uint)
	result := q.DB.Where("server_id in (?)", serverIds).Find(&entities)
	q.Logger.Info().Msg(
		fmt.Sprintf("RoleQueries: get roles by server_ids: %d, count: %d", len(serverIds), len(entities)),
	)
	if result.Error != nil {
		return hmap, result.Error
	}
	for _, entity := range entities {
		if entity.ServerID == nil || entity.RoleID == 0 {
			continue
		}
		if val, ok := hmap[*entity.ServerID]; ok {
			hmap[*entity.ServerID] = append(val, entity.RoleID)
		} else {
			hmap[*entity.ServerID] = []uint{entity.RoleID}
		}
	}
	return hmap, result.Error
}

func (q *RoleQueries) SetByUserId(userId uint, roleIds []uint) error {
	q.Logger.Info().Msg(fmt.Sprintf("RoleQueries: set roles %d count by user_id = %d", len(roleIds), userId))
	// Начинаем транзакцию
	return q.DB.Transaction(func(tx *gorm.DB) error {
		// Удаляем старые роли пользователя
		if err := tx.Where("user_id = ?", userId).Delete(&models.RoleRelation{}).Error; err != nil {
			return err
		}
		// Добавляем новые роли
		for _, roleId := range roleIds {
			relation := models.RoleRelation{}
			relation.UserID = &userId
			relation.RoleID = roleId
			if err := tx.Create(&relation).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (q *RoleQueries) SetByServerId(serverId uint, roleIds []uint) error {
	q.Logger.Info().Msg(fmt.Sprintf("RoleQueries: set roles %d count by server_id = %d", roleIds, serverId))
	// Начинаем транзакцию
	return q.DB.Transaction(func(tx *gorm.DB) error {
		// Удаляем старые роли сервера
		if err := tx.Where("server_id = ?", serverId).Delete(&models.RoleRelation{}).Error; err != nil {
			return err
		}
		// Добавляем новые роли
		for _, roleId := range roleIds {
			relation := models.RoleRelation{}
			relation.ServerID = &serverId
			relation.RoleID = roleId
			if err := tx.Create(&relation).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (q *RoleQueries) GetByCode(code string) (models.Role, error) {
	var entity models.Role
	q.Logger.Info().Msg(fmt.Sprintf("RoleQueries: get role by code=%+v", code))
	if code == "" {
		return entity, RoleNotFoundError
	}
	q.DB.Where("code = ?", code).Limit(1).Find(&entity)
	if entity.Code != code {
		return entity, RoleNotFoundError
	}
	return entity, nil
}

func (q *RoleQueries) Upsert(entity *models.Role) error {
	if entity == nil {
		return nil
	}
	entityDB := models.Role{}
	q.DB.Where("id = ? or code = ?", entity.ID, entity.Code).Find(&entityDB)
	if entityDB.ID != 0 {
		// Update
		q.Logger.Debug().Msg(fmt.Sprintf("ServiceCardQueries: entity update: %+v", entityDB))
		q.Logger.Info().Msg(fmt.Sprintf("ServiceCardQueries: entity update id=%+v", entityDB.ID))
		entity.ID = entityDB.ID
		entity.CreatedAt = entityDB.CreatedAt
		entity.UpdatedAt = time.Now().UTC()
		result := q.DB.Save(&entity)
		return result.Error
	} else {
		// Create
		q.Logger.Debug().Msg(fmt.Sprintf("ServiceCardQueries: entity create: %+v", entity))
		q.Logger.Info().Msg(fmt.Sprintf("ServiceCardQueries: entity create name=%+v", entity.Name))
		entity.ID = 0
		entity.CreatedAt = time.Now().UTC()
		entity.UpdatedAt = time.Now().UTC()
		result := q.DB.Create(entity)
		return result.Error
	}
}

func (q *RoleQueries) List() ([]models.Role, error) {
	var entities []models.Role
	result := q.DB.Offset(0).Find(&entities)
	q.Logger.Debug().Msg("RoleQueries: get list entities")
	return entities, result.Error
}

func (q *RoleQueries) Delete(id uint) error {
	tx := q.DB.Where("id = ?", id).Delete(&models.Role{})
	q.Logger.Debug().Msg(fmt.Sprintf("RoleQueries: delete entity by id: %+v", id))
	return tx.Error
}
