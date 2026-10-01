package queries

import (
	"errors"
	"fmt"
	"time"

	"github.com/rs/zerolog"
	"gorm.io/gorm"

	"github.com/cms-lab-core/cms-labs-api/backend/app/models"
	"github.com/cms-lab-core/cms-labs-api/shared/jsonrpc"
)

type TargetUserQueries struct {
	DB     *gorm.DB
	Logger *zerolog.Logger
}

var TargetUserNotFoundError = jsonrpc.NewRpcError("target_user_not_found", "Target user not found")

// GetByTargetAndUser ищет запись по составному ключу (target_id, user_id)
func (q *TargetUserQueries) GetByTargetAndUser(targetID string, userID uint) (models.TargetUser, error) {
	var entity models.TargetUser
	err := q.DB.Where("target_id = ? AND user_id = ?", targetID, userID).First(&entity).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return entity, TargetUserNotFoundError
	}
	return entity, err
}

// GetByUserId ищет записи все где есть user_id
func (q *TargetUserQueries) GetByUserId(userID uint) ([]string, error) {
	var targetIDs []string
	err := q.DB.Model(&models.TargetUser{}).
		Where("user_id = ?", userID).
		Pluck("target_id", &targetIDs).Error
	if err != nil {
		return nil, err
	}
	return targetIDs, err
}

// HasEditor проверяет наличие редактора, загружая всех пользователей цели в память.
func (q *TargetUserQueries) HasEditor(targetID string) (bool, error) {
	var users []models.TargetUser
	err := q.DB.Where("target_id = ?", targetID).Find(&users).Error
	if err != nil {
		return false, err
	}
	for _, u := range users {
		// Извлекаем массив ролей из JSON-объекта
		rolesInterface, ok := u.Roles["roles"]
		if !ok {
			continue
		}
		rolesSlice, ok := rolesInterface.([]interface{})
		if !ok {
			continue
		}
		for _, r := range rolesSlice {
			if str, ok := r.(string); ok && str == string(models.UserRoleEditor) {
				return true, nil
			}
		}
	}
	return false, nil
}

// CheckTargetAndUserByRole находит пользователя цели и проверяет наличие указанной роли.
// Возвращает запись TargetUser, если роль найдена, иначе — TargetUserNotFoundError или специализированную ошибку.
func (q *TargetUserQueries) CheckTargetAndUserByRole(targetID string, userID uint, role string) bool {
	var entity models.TargetUser

	// 1. Ищем запись по составному ключу
	err := q.DB.Where("target_id = ? AND user_id = ?", targetID, userID).First(&entity).Error
	if err != nil {
		return false
	}

	// 2. Извлекаем массив ролей из UserStore
	// Ожидаемая структура JSON: {"roles": ["editor", "vault_viewer", ...]}
	rolesMap := entity.Roles // тип map[string]interface{}
	if rolesMap == nil {
		return false
	}

	rolesInterface, ok := rolesMap["roles"]
	if !ok {
		return false
	}

	// 3. Преобразуем []interface{} в []string
	rolesSlice, ok := rolesInterface.([]interface{})
	if !ok {
		return false
	}

	// 4. Проверяем наличие нужной роли
	for _, r := range rolesSlice {
		if str, ok := r.(string); ok && str == role {
			return true
		}
	}

	// 5. Роль не найдена
	return false
}

// Upsert создаёт или обновляет запись по составному ключу
func (q *TargetUserQueries) Upsert(entity *models.TargetUser) error {
	if entity == nil {
		return nil
	}
	// Поиск по target_id и user_id
	var existing models.TargetUser
	err := q.DB.Where("target_id = ? AND user_id = ?", entity.TargetID, entity.UserID).First(&existing).Error
	if err == nil {
		// Update
		q.Logger.Debug().Msg(fmt.Sprintf("TargetUserQueries: update %+v", existing))
		q.Logger.Info().Msg(fmt.Sprintf("TargetUserQueries: update id=%d", existing.ID))
		entity.ID = existing.ID
		entity.CreatedAt = existing.CreatedAt
		entity.UpdatedAt = time.Now().UTC()
		return q.DB.Save(entity).Error
	} else if err == gorm.ErrRecordNotFound {
		// Create
		q.Logger.Debug().Msg(fmt.Sprintf("TargetUserQueries: create %+v", entity))
		q.Logger.Info().Msg(fmt.Sprintf("TargetUserQueries: create target=%s user=%d", entity.TargetID, entity.UserID))
		entity.CreatedAt = time.Now().UTC()
		entity.UpdatedAt = time.Now().UTC()
		return q.DB.Create(entity).Error
	}
	return err
}

// Delete удаляет запись по ID
func (q *TargetUserQueries) Delete(id uint) error {
	q.Logger.Debug().Msg(fmt.Sprintf("TargetUserQueries: delete id=%d", id))
	return q.DB.Where("id = ?", id).Delete(&models.TargetUser{}).Error
}

// DeleteByTargetAndUser удаляет по составному ключу
func (q *TargetUserQueries) DeleteByTargetAndUser(targetID string, userID uint) error {
	q.Logger.Debug().Msg(fmt.Sprintf("TargetUserQueries: delete target=%s user=%d", targetID, userID))
	return q.DB.Where("target_id = ? AND user_id = ?", targetID, userID).Delete(&models.TargetUser{}).Error
}

// GetAllByTarget возвращает всех пользователей для указанного target
func (q *TargetUserQueries) GetAllByTarget(targetID string) ([]models.TargetUser, error) {
	var entities []models.TargetUser
	err := q.DB.Where("target_id = ?", targetID).Find(&entities).Error
	return entities, err
}
