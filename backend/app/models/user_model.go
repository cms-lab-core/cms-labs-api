package models

import (
	"time"

	"github.com/cms-lab-core/cms-labs-api/backend/app/models/types"
)

// UserBase struct to describe User object.
type UserBase struct {
	Email     string     `gorm:"type:varchar(255)" json:"email" valid:"required,email"`
	Name      string     `gorm:"type:varchar(255)" json:"name" valid:"required"`
	GroupName string     `gorm:"type:varchar(255)" json:"group_name" validate:"required"`
	LTIUserID string     `gorm:"type:varchar(255)" json:"lti_user_id"`
	DeletedAt *time.Time `gorm:"type:datetime(3)" json:"deleted_at"`
}

type UserSecret struct {
	LastLaunchID string          `gorm:"type:varchar(255)" json:"last_launch_id"`
	Store        types.JsonStore `gorm:"type:json" json:"store"`
}

type UserListItem struct {
	Base
	UserBase
}

// TableName переопределяет название таблицы для UserListItem на `user`
func (UserListItem) TableName() string {
	return "user"
}

type User struct {
	Base
	UserBase
	UserSecret
}

func (c *UserBase) IsActive() bool {
	zeroTime := time.Time{}
	return c.DeletedAt == nil || *c.DeletedAt == zeroTime
}
