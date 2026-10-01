package queries

import (
	"crypto/rand"
	"math/big"

	"github.com/rs/zerolog"

	"github.com/cms-lab-core/cms-labs-api/backend/app/models"
	"gorm.io/gorm"
)

type LTIRoomQueries struct {
	DB     *gorm.DB
	Logger *zerolog.Logger
}

func (q *LTIRoomQueries) Get(id uint) (models.LTIRoom, error) {
	var entity models.LTIRoom
	result := q.DB.Where("id = ?", id).Find(&entity)
	if entity.ID == 0 {
		return entity, LTIAttemptNotFoundError
	}
	return entity, result.Error
}

func (q *LTIRoomQueries) GetByRoomNumber(roomNumber int64) (models.LTIRoom, error) {
	var entity models.LTIRoom
	result := q.DB.Where("room_number = ?", roomNumber).Order(
		`id desc`,
	).Limit(1).Offset(0).Find(&entity)
	if entity.ID == 0 {
		return entity, LTIAttemptNotFoundError
	}
	return entity, result.Error
}

func (q *LTIRoomQueries) Create() (*models.LTIRoom, error) {
	var minRand int64 = 10000
	var maxRand int64 = 99999
	roomNumberFrom, _ := rand.Int(rand.Reader, big.NewInt(maxRand-minRand+1))
	roomNumber := roomNumberFrom.Int64() + minRand
	entity := models.LTIRoom{}
	entity.RoomNumber = roomNumber
	err := q.DB.Model(&models.LTIRoom{}).Where("room_number = ?", roomNumber).Update("room_number", nil).Error
	if err != nil {
		return &entity, err
	}
	err = q.DB.Save(&entity).Error
	return &entity, err
}
