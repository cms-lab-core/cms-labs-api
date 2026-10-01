package queries

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/rs/zerolog"

	"github.com/google/uuid"
	"github.com/ory/go-convenience/stringsx"
	"golang.org/x/crypto/bcrypt"

	"github.com/cms-lab-core/cms-labs-api/backend/app/models"
	"github.com/cms-lab-core/cms-labs-api/shared/jsonrpc"
	"gorm.io/gorm"
)

type ServerQueries struct {
	DB     *gorm.DB
	Logger *zerolog.Logger
}

type ServerQueriesListDTO struct {
	Search string   `json:"search"`
	Limit  int      `json:"limit"`
	Offset int      `json:"offset"`
	Types  []string `json:"types"`
	// The type of status
	// enum: all,active
	Status string `json:"status"`
	// The type of orderBy
	// enum: createdAt,unitRate,lastCountUsers
	OrderBy string `json:"order_by"`
}

const (
	ServerListInputDTOStatusAll    = "all"
	ServerListInputDTOStatusActive = "active"
)

const (
	ServerListInputDTOOrderByCreatedAt      = "createdAt"
	ServerListInputDTOOrderByUnitRate       = "unitRate"
	ServerListInputDTOOrderByLastCountUsers = "lastCountUsers"
)

var ServerNotFoundError = jsonrpc.NewRpcError("server_not_found", "Server has not found")

func (q *ServerQueries) tableName(object interface{}) string {
	stmt := &gorm.Statement{DB: q.DB}
	_ = stmt.Parse(object)
	return stmt.Schema.Table
}

func (q *ServerQueries) Get(id uint) (models.Server, error) {
	var entity models.Server
	result := q.DB.First(&entity, id)
	if result.Error != nil && result.Error.Error() == "record not found" {
		return entity, ServerNotFoundError
	}
	return entity, result.Error
}

func (q *ServerQueries) securityTokenGenerate() string {
	uid := uuid.New().String()
	hash, err := bcrypt.GenerateFromPassword([]byte(uid), bcrypt.DefaultCost)
	if err != nil {
		q.Logger.Warn().Msg(fmt.Sprintf("ServerQueries: securityTokenGenerate error: %+v", err))
		return ""
	}
	hasher := sha256.New()
	hasher.Write(hash)
	return hex.EncodeToString(hasher.Sum(nil))
}

func (q *ServerQueries) Upsert(entity *models.Server) error {
	if entity == nil {
		return nil
	}
	entityDB := models.Server{}
	q.DB.Where("id = ?", entity.ID).Find(&entityDB)
	if entityDB.ID != 0 {
		// Update
		q.Logger.Debug().Msg(fmt.Sprintf("ServerQueries: entity update: %+v", entityDB))
		q.Logger.Info().Msg(fmt.Sprintf("ServerQueries: entity update id=%+v", entityDB.ID))
		entity.ID = entityDB.ID
		entity.CreatedAt = entityDB.CreatedAt
		entity.UpdatedAt = time.Now().UTC()
		entity.LastOnlineStatus = entityDB.LastOnlineStatus
		entity.Token = stringsx.Coalesce(entityDB.Token, q.securityTokenGenerate())
		result := q.DB.Save(&entity)
		return result.Error
	} else {
		// Create
		q.Logger.Debug().Msg(fmt.Sprintf("ServerQueries: entity create: %+v", entity))
		q.Logger.Info().Msg(fmt.Sprintf("ServerQueries: entity create name=%+v", entity.Name))
		entity.ID = 0
		entity.Token = q.securityTokenGenerate()
		entity.CreatedAt = time.Now().UTC()
		entity.UpdatedAt = time.Now().UTC()
		result := q.DB.Create(entity)
		return result.Error
	}
}

func (q *ServerQueries) List(filter ServerQueriesListDTO) ([]models.ServerListItem, int64, error) {
	var entities []models.ServerListItem
	result := q.listFilter(
		filter,
		q.DB.Limit(MaxLimitCount).Offset(0),
	).Find(&entities)
	count := result.RowsAffected
	q.Logger.Debug().Msg(fmt.Sprintf("ServerQueries list: count %+v", count))
	result = q.listFilter(
		filter,
		q.DB.Limit(filter.Limit).Offset(filter.Offset),
	).Find(&entities)
	q.Logger.Debug().Msg(fmt.Sprintf("ServerQueries list: entities %+v", entities))
	return entities, count, result.Error
}

func (q *ServerQueries) listFilter(filter ServerQueriesListDTO, tx *gorm.DB) *gorm.DB {
	tx = tx.Table(q.tableName(&models.Server{}) + " AS servers")

	switch filter.OrderBy {
	case ServerListInputDTOOrderByLastCountUsers:
		tx = tx.Order("last_count_users DESC")
	case ServerListInputDTOOrderByUnitRate:
		tx = tx.Order("unit_rate DESC")
	default:
		tx = tx.Order("created_at DESC")
	}

	if filter.Status == ServerListInputDTOStatusActive {
		tx = models.ServerIsRealActive(tx)
	}

	if len(filter.Types) > 0 {
		tx = tx.Where("type IN (?)", filter.Types)
	}

	if filter.Search != "" {
		searchPattern := "%" + filter.Search + "%"
		tx = tx.Where("name LIKE ? OR url LIKE ? OR id = ? OR client_id = ?", searchPattern, searchPattern, filter.Search, filter.Search)
	}

	return tx
}

func (q *ServerQueries) Delete(id uint) error {
	tx := q.DB.Where("id = ?", id).Delete(&models.Server{})
	q.Logger.Debug().Msg(fmt.Sprintf("ServerQueries: delete entity by id: %+v", id))
	return tx.Error
}

func (q *ServerQueries) GetByClientId(clientID string) (models.Server, error) {
	var entity models.Server
	result := q.DB.Where("client_id = ?", clientID).Limit(1).Find(&entity)
	if entity.ID == 0 {
		return entity, ServerNotFoundError
	}
	return entity, result.Error
}
