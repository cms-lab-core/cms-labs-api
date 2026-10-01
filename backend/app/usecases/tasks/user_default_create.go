package tasks

import (
	"github.com/cms-lab-core/cms-labs-api/backend/app/models"
	"github.com/cms-lab-core/cms-labs-api/backend/app/queries"
	"github.com/cms-lab-core/cms-labs-api/shared/cms_client"
	"github.com/rs/zerolog"
	"golang.org/x/crypto/bcrypt"
)

type UserDefaultCreateUC struct {
	UserQueries         *queries.UserQueries
	UserPasswordQueries *queries.UserPasswordQueries
	RoleQueries         *queries.RoleQueries
	*zerolog.Logger
}

const UserDefaultEmail = "admin@admin.com"
const UserDefaultPassword = "admin"
const UserDefaultName = "admin"

func (u *UserDefaultCreateUC) userDefaultCreate() (uint, error) {
	entity, _ := u.UserQueries.GetByEmail(UserDefaultEmail)
	if entity.Email != UserDefaultEmail {
		entity.Name = UserDefaultName
		entity.Email = UserDefaultEmail
		_ = u.UserQueries.Upsert(&entity)
	}
	creds, _ := u.UserPasswordQueries.Get(entity.ID)
	if creds.UserID == entity.ID {
		return entity.ID, nil
	}
	hashPassword, err := bcrypt.GenerateFromPassword(
		[]byte(UserDefaultPassword),
		14,
	)
	if err != nil {
		return entity.ID, err
	}
	err = u.UserPasswordQueries.Upsert(&models.UserPassword{
		UserID:       entity.ID,
		HashPassword: string(hashPassword),
	})
	u.Logger.Info().Msg("User default created")
	return entity.ID, err
}

func (u *UserDefaultCreateUC) createRoles() []uint {
	student := &models.Role{}
	student.Code = cms_client.SSOUsersRoleStudent
	student.Name = "Student"

	_ = u.RoleQueries.Upsert(student)

	instructor := &models.Role{}
	instructor.Code = cms_client.SSOUsersRoleInstructor
	instructor.Name = "Instructor"

	_ = u.RoleQueries.Upsert(instructor)

	admin := &models.Role{}
	admin.Code = cms_client.SSOUsersRoleAdmin
	admin.Name = "Admin"

	_ = u.RoleQueries.Upsert(admin)

	u.Logger.Info().Msg("Roles default created")
	return []uint{student.ID, instructor.ID, admin.ID}
}

func (u *UserDefaultCreateUC) setUserRoles(userId uint, rolesIds []uint) error {
	err := u.RoleQueries.SetByUserId(userId, rolesIds)
	u.Logger.Info().Msg("Roles admin user set")
	return err
}

func (u *UserDefaultCreateUC) Execute() error {
	userId, _ := u.userDefaultCreate()
	rolesIds := u.createRoles()
	_ = u.setUserRoles(userId, rolesIds)
	return nil
}
