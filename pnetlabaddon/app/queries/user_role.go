package queries

import (
	"github.com/cms-lab-core/cms-labs-api/pnetlabaddon/app/models"
	"gorm.io/gorm"
)

const (
	UserRoleStudent            = "student"
	UserRoleStudentDescription = "LTI Role integration"
)

type UserRoleQueries struct {
	DB *gorm.DB
}

func (q *UserRoleQueries) GetOrCreateDefault() (models.UserRole, error) {
	entityDB := models.UserRole{}
	q.DB.Where("user_role_name = ?", UserRoleStudent).Find(&entityDB)
	if entityDB.UserRoleID != 0 {
		return entityDB, nil
	}
	entityDB.UserRoleName = UserRoleStudent
	entityDB.UserRoleWorkspace = "/"
	entityDB.UserRoleNote = UserRoleStudentDescription
	result := q.DB.Create(&entityDB)
	return entityDB, result.Error
}
