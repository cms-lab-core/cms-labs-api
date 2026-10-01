package queries

import (
	"fmt"
	"time"

	"github.com/rs/zerolog"
	"gorm.io/gorm"

	"github.com/cms-lab-core/cms-labs-api/backend/app/models"
)

type TargetRelationQueries struct {
	DB     *gorm.DB
	Logger *zerolog.Logger
}

// Upsert создаёт или обновляет запись по составному уникальному ключу
func (q *TargetRelationQueries) Upsert(entity *models.TargetRelation) error {
	if entity == nil {
		return nil
	}
	// Поиск по уникальному ключу (from, to, type)
	var existing models.TargetRelation
	err := q.DB.Where("from_target_id = ? AND to_target_id = ? AND relation_type = ?",
		entity.FromTargetID, entity.ToTargetID, entity.RelationType).First(&existing).Error
	if err == nil {
		// Update
		q.Logger.Debug().Msg(fmt.Sprintf("TargetRelationQueries: update %+v", existing))
		q.Logger.Info().Msg(fmt.Sprintf("TargetRelationQueries: update id=%d", existing.ID))
		entity.ID = existing.ID
		entity.CreatedAt = existing.CreatedAt
		entity.UpdatedAt = time.Now().UTC()
		return q.DB.Save(entity).Error
	} else if err == gorm.ErrRecordNotFound {
		// Create
		q.Logger.Debug().Msg(fmt.Sprintf("TargetRelationQueries: create %+v", entity))
		q.Logger.Info().Msg(fmt.Sprintf("TargetRelationQueries: create from=%s to=%s type=%s",
			entity.FromTargetID, entity.ToTargetID, entity.RelationType))
		entity.CreatedAt = time.Now().UTC()
		entity.UpdatedAt = time.Now().UTC()
		return q.DB.Create(entity).Error
	}
	return err
}

// DeleteByFromToType удаляет запись по ID
func (q *TargetRelationQueries) DeleteByFromToType(fromTargetID, toTargetID, relationType string) error {
	q.Logger.Debug().Msg(fmt.Sprintf("TargetRelationQueries: delete from=%s to=%s type=%s",
		fromTargetID, toTargetID, relationType))
	return q.DB.Where("from_target_id = ? AND to_target_id = ? AND relation_type = ?",
		fromTargetID, toTargetID, relationType).Delete(&models.TargetRelation{}).Error
}

func (q *TargetRelationQueries) tableName(object interface{}) string {
	stmt := &gorm.Statement{DB: q.DB}
	_ = stmt.Parse(object)
	return stmt.Schema.Table
}

func (q *TargetRelationQueries) List() ([]models.TargetRelation, error) {
	var entities []models.TargetRelation
	result := q.DB.Table(q.tableName(&models.TargetRelation{}) + " AS target_relations").Order(`created_at desc`).Find(&entities)
	q.Logger.Debug().Msg(fmt.Sprintf("TargetRelationQueries list: entities %+v", entities))
	return entities, result.Error
}
