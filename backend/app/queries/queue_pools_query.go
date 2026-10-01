package queries

import (
	"fmt"
	"time"

	"github.com/rs/zerolog"

	"github.com/cms-lab-core/cms-labs-api/backend/app/models"
	"github.com/cms-lab-core/cms-labs-api/shared/jsonrpc"
	gorm "gorm.io/gorm"
)

type RoundQueuePoolQueries struct {
	DB     *gorm.DB
	Logger *zerolog.Logger
}

func (q *RoundQueuePoolQueries) tableName(object interface{}) string {
	stmt := &gorm.Statement{DB: q.DB}
	_ = stmt.Parse(object)
	return stmt.Schema.Table
}

type RoundQueuePoolServerListItem struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"type:varchar(255)" json:"name"`
	Type        string    `gorm:"type:varchar(255)" json:"type" validate:"required"`
	ConnectedAt time.Time `gorm:"type:datetime(3)" json:"connected_at" validate:"required"`
	LastUsed    bool      `json:"last_used" validate:"required"`
}

func (q *RoundQueuePoolQueries) Get(id uint) (models.RoundQueuePool, error) {
	var entity models.RoundQueuePool
	result := q.DB.First(&entity, id)
	if result.Error != nil && result.Error.Error() == "record not found" {
		return entity, jsonrpc.NewRpcError("round_queue_not_found", "RoundQueuePool not found")
	}
	return entity, result.Error
}

func (q *RoundQueuePoolQueries) UpsertQueueByServerIds(ids []uint) error {
	q.Logger.Info().Msg(fmt.Sprintf("RoundQueuePoolQueries upsert queue by server-ids: count %+v", len(ids)))
	var entities []models.RoundQueuePool
	for index, serverId := range ids {
		entity := models.RoundQueuePool{}
		entity.CreatedAt = time.Now().UTC()
		entity.UpdatedAt = time.Now().UTC()
		entity.ConnectedAt = time.Now().UTC().Add(time.Duration(index-len(ids)) * time.Minute)
		entity.LastUsed = false
		entity.IsActive = true
		entity.ServerID = serverId
		entity.Type = models.RoundQueuePoolTypePNET
		entities = append(entities, entity)
	}
	err := q.DB.Transaction(
		func(tx *gorm.DB) error {
			if err := tx.Where("type = ?", models.RoundQueuePoolTypePNET).Delete(&models.RoundQueuePool{}).Error; err != nil {
				return err
			}
			if err := tx.CreateInBatches(entities, 10).Error; err != nil {
				return err
			}
			return nil
		},
	)
	return err
}

func (q *RoundQueuePoolQueries) List() ([]RoundQueuePoolServerListItem, error) {
	var entities []RoundQueuePoolServerListItem
	query := q.DB.Table(
		q.tableName(&models.RoundQueuePool{})+" AS round_queue_pools",
	).Select(
		"servers.id, servers.name, round_queue_pools.type, round_queue_pools.last_used, round_queue_pools.connected_at",
	).Joins(
		"join "+q.tableName(&models.Server{})+" servers on servers.id = round_queue_pools.server_id",
	).Where(
		"round_queue_pools.type = ?",
		models.RoundQueuePoolTypePNET,
	)
	query = models.ServerIsRealActive(query)
	query = query.Limit(MaxLimitCount).Offset(0).Order("round_queue_pools.connected_at asc")
	if err := query.Scan(&entities).Error; err != nil {
		return entities, err
	}
	return entities, nil
}

func (q *RoundQueuePoolQueries) GetNextByType(poolType string) (models.RoundQueuePool, error) {
	var entityID uint
	err := q.DB.Transaction(
		func(tx *gorm.DB) error {
			entity, err := q.nextPoolItemByType(tx, poolType)
			if err != nil {
				return err
			}
			entityID = entity.ID
			return q.finishDistribution(tx, entity.Type, entity.ID)
		},
	)
	q.Logger.Info().Msg(fmt.Sprintf("RoundQueuePoolQueries: get by id: %+v", entityID))
	if err != nil {
		return models.RoundQueuePool{}, err
	}
	return q.Get(entityID)
}

func (q *RoundQueuePoolQueries) nextPoolItemByType(tx *gorm.DB, poolType string) (models.RoundQueuePool, error) {
	var entities []models.RoundQueuePool
	q.Logger.Info().Msg(fmt.Sprintf("RoundQueuePoolQueries get next pool item from queue by type: %+v", poolType))
	selectCurrentDistribution := tx.Table(
		q.tableName(&models.RoundQueuePool{})+" AS round_queue_pools",
	).Select(
		"round_queue_pools.connected_at",
	).Where(
		"round_queue_pools.type = ?",
		poolType,
	).Where(
		"round_queue_pools.last_used = ?",
		true,
	).Limit(1)
	selectNextDistribution := tx.Table(
		q.tableName(&models.RoundQueuePool{})+" AS round_queue_pools",
	).Select(
		"round_queue_pools.id, "+
			"round_queue_pools.created_at, "+
			"round_queue_pools.updated_at, "+
			"round_queue_pools.type, "+
			"round_queue_pools.connected_at, "+
			"round_queue_pools.last_used, "+
			"round_queue_pools.is_active, "+
			"round_queue_pools.server_id",
	).Joins(
		"join "+q.tableName(&models.Server{})+" servers on servers.id = round_queue_pools.server_id",
		q.tableName(&models.Server{}),
	).Where(
		"round_queue_pools.type = ?",
		poolType,
	).Where(
		"round_queue_pools.is_active = ?",
		true,
	).Where(
		"round_queue_pools.last_used = ?",
		false,
	).Where(
		"round_queue_pools.connected_at > (?)",
		selectCurrentDistribution,
	)
	selectNextDistribution = models.ServerIsRealActive(selectNextDistribution)
	selectNextDistribution = selectNextDistribution.Limit(1).Offset(0).Order(
		"round_queue_pools.connected_at asc",
	)
	if err := selectNextDistribution.Scan(&entities).Error; err != nil {
		q.Logger.Info().Msg(
			fmt.Sprintf(
				"RoundQueuePoolQueries exception for get next pool item from queue by type: %+v",
				poolType,
			),
		)
		return models.RoundQueuePool{}, err
	}
	if len(entities) > 0 {
		entity := entities[0]
		q.Logger.Info().Msg(
			fmt.Sprintf(
				"RoundQueuePoolQueries success get id: %+v, serverId: %+v, pool item from queue by type: %+v",
				entity.ID,
				entity.ServerID,
				entity.Type,
			),
		)
		return entity, nil
	}
	selectFirstDistribution := tx.Table(q.tableName(&models.RoundQueuePool{})+" AS round_queue_pools").Select(
		"round_queue_pools.id, "+
			"round_queue_pools.created_at, "+
			"round_queue_pools.updated_at, "+
			"round_queue_pools.type, "+
			"round_queue_pools.connected_at, "+
			"round_queue_pools.last_used, "+
			"round_queue_pools.is_active, "+
			"round_queue_pools.server_id",
	).Joins(
		"join "+q.tableName(&models.Server{})+" servers on servers.id = round_queue_pools.server_id",
	).Where(
		"round_queue_pools.type = ?",
		poolType,
	).Where(
		"round_queue_pools.is_active = ?",
		true,
	)
	selectFirstDistribution = models.ServerIsRealActive(selectFirstDistribution)
	selectFirstDistribution = selectFirstDistribution.Limit(1).Offset(0).Order(
		"round_queue_pools.connected_at asc",
	)
	if err := selectFirstDistribution.Scan(&entities).Error; err != nil {
		q.Logger.Info().Msg(
			fmt.Sprintf(
				"RoundQueuePoolQueries exception for get first pool item from queue by type: %+v",
				poolType,
			),
		)
		return models.RoundQueuePool{}, err
	}
	if len(entities) > 0 {
		entity := entities[0]
		q.Logger.Info().Msg(
			fmt.Sprintf(
				"RoundQueuePoolQueries success get first entity id: %+v, serverId: %+v, pool item from queue by type: %+v",
				entity.ID,
				entity.ServerID,
				entity.Type,
			),
		)
		return entity, nil
	}
	q.Logger.Info().Msg(
		fmt.Sprintf(
			"RoundQueuePoolQueries not success get entity from queue by type: %+v",
			poolType,
		),
	)
	return models.RoundQueuePool{}, jsonrpc.NewRpcError("pool_item_not_found", "no pool item found")
}

func (q *RoundQueuePoolQueries) finishDistribution(tx *gorm.DB, poolType string, id uint) error {
	q.Logger.Info().Msg(
		fmt.Sprintf(
			"RoundQueuePoolQueries: finish distribution item from queue by type: %+v, id: %+v",
			poolType,
			id,
		),
	)
	resetLastDistribution := tx.Table(
		q.tableName(&models.RoundQueuePool{})+" AS round_queue_pools",
	).Where("round_queue_pools.type = ?", poolType).Updates(
		map[string]interface{}{
			"updated_at": time.Now().UTC(),
			"last_used":  false,
		},
	)
	if resetLastDistribution.Error != nil {
		q.Logger.Info().Msg(
			fmt.Sprintf(
				"RoundQueuePoolQueries: error on change reset last distribution by type: %+v, id: %+v",
				poolType,
				id,
			),
		)
		return resetLastDistribution.Error
	}
	setLastDistribution := tx.Table(
		q.tableName(&models.RoundQueuePool{})+" AS round_queue_pools",
	).Where("round_queue_pools.id = ?", id).Updates(
		map[string]interface{}{
			"updated_at": time.Now().UTC(),
			"last_used":  true,
		},
	)
	if setLastDistribution.Error != nil {
		q.Logger.Info().Msg(
			fmt.Sprintf(
				"RoundQueuePoolQueries: error on change set last distribution by type: %+v, id: %+v",
				poolType,
				id,
			),
		)
		return setLastDistribution.Error
	}
	return nil
}
