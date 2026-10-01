package queries

import (
	"errors"

	"github.com/cms-lab-core/cms-labs-api/pnetlabaddon/app/models"
	"gorm.io/gorm"
)

var (
	UserNotFoundError = errors.New("User not found")
)

type UserQueries struct {
	DB *gorm.DB
}

func (q *UserQueries) Get(id int) (models.User, error) {
	entityDB := models.User{}
	q.DB.Where("pod = ?", id).Find(&entityDB)
	if entityDB.Pod != 0 {
		return entityDB, nil
	}
	return entityDB, UserNotFoundError
}

func (q *UserQueries) GetOrCreateByEmail(user models.User) (models.User, error) {
	entityDB := models.User{}
	q.DB.Where("email = ?", user.Email).Find(&entityDB)
	if entityDB.Pod != 0 {
		return entityDB, nil
	}
	result := q.DB.Create(&user)
	if result.Error != nil {
		return user, result.Error
	}
	return user, nil
}
