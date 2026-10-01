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

type TargetAddonQueries struct {
	DB     *gorm.DB
	Logger *zerolog.Logger
}

var TargetAddonNotFoundError = jsonrpc.NewRpcError("target_addon_not_found", "Target addon not found")

// Get возвращает запись по ID
func (q *TargetAddonQueries) Get(id uint) (models.TargetAddon, error) {
	var entity models.TargetAddon
	err := q.DB.First(&entity, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return entity, TargetAddonNotFoundError
	}
	return entity, err
}

// GetByTargetAndAddon ищет запись по составному ключу (target_id, addon_id)
func (q *TargetAddonQueries) GetByTargetAndAddon(targetID string, addonID string) (models.TargetAddon, error) {
	var entity models.TargetAddon
	err := q.DB.Where("target_id = ? AND addon_id = ?", targetID, addonID).First(&entity).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return entity, TargetAddonNotFoundError
	}
	return entity, err
}

// Upsert создаёт или обновляет запись по составному ключу
func (q *TargetAddonQueries) Upsert(entity *models.TargetAddon) error {
	if entity == nil {
		return nil
	}
	// Поиск по target_id и addon_id
	var existing models.TargetAddon
	err := q.DB.Where("target_id = ? AND addon_id = ?", entity.TargetID, entity.AddonID).First(&existing).Error
	if err == nil {
		// Update
		q.Logger.Debug().Msg(fmt.Sprintf("TargetAddonQueries: update %+v", existing))
		q.Logger.Info().Msg(fmt.Sprintf("TargetAddonQueries: update id=%d", existing.ID))
		entity.ID = existing.ID
		entity.CreatedAt = existing.CreatedAt
		entity.UpdatedAt = time.Now().UTC()
		return q.DB.Save(entity).Error
	} else if err == gorm.ErrRecordNotFound {
		// Create
		q.Logger.Debug().Msg(fmt.Sprintf("TargetAddonQueries: create %+v", entity))
		q.Logger.Info().Msg(fmt.Sprintf("TargetAddonQueries: create target=%s addon=%s", entity.TargetID, entity.AddonID))
		entity.CreatedAt = time.Now().UTC()
		entity.UpdatedAt = time.Now().UTC()
		return q.DB.Create(entity).Error
	}
	return err
}

// TargetAddonListDTO — параметры фильтрации
type TargetAddonListDTO struct {
	TargetID  string `json:"target_id"`
	AddonType string `json:"addon_type"`
	AddonID   string `json:"addon_id"`
	Limit     int    `json:"limit"`
	Offset    int    `json:"offset"`
}

// Delete удаляет запись по ID
func (q *TargetAddonQueries) Delete(id uint) error {
	q.Logger.Debug().Msg(fmt.Sprintf("TargetAddonQueries: delete id=%d", id))
	return q.DB.Where("id = ?", id).Delete(&models.TargetAddon{}).Error
}

// DeleteByTargetAndAddon удаляет по составному ключу
func (q *TargetAddonQueries) DeleteByTargetAndAddon(targetID string, addonID string) error {
	q.Logger.Debug().Msg(fmt.Sprintf("TargetAddonQueries: delete target=%s addon=%s", targetID, addonID))
	return q.DB.Where("target_id = ? AND addon_id = ?", targetID, addonID).Delete(&models.TargetAddon{}).Error
}

// GetAllByTarget возвращает все аддоны для указанного target
func (q *TargetAddonQueries) GetAllByTarget(targetID string) ([]models.TargetAddon, error) {
	var entities []models.TargetAddon
	result := q.DB.Where("target_id = ?", targetID).Find(&entities)
	if result.Error != nil {
		return nil, result.Error
	}
	return entities, nil
}

// Update обновляет запись
func (q *TargetAddonQueries) Update(entity *models.TargetAddon) error {
	if entity == nil {
		return nil
	}
	q.Logger.Debug().Msg(fmt.Sprintf("TargetAddonQueries: update id=%d", entity.ID))
	entity.UpdatedAt = time.Now().UTC()
	return q.DB.Save(entity).Error
}
