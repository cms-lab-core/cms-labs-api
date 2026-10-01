package models

import (
	"time"

	"github.com/cms-lab-core/cms-labs-api/backend/app/models/types"
	"github.com/cms-lab-core/cms-labs-api/shared/connection"
	"gorm.io/datatypes"
)

// роль пользователя в TargetUser

const (
	UserRoleVaultViewer  = "vault_viewer"
	UserRoleVaultWriter  = "vault_writer"
	UserRoleHarborAccess = "harbor_access"
	UserRoleEditor       = "editor"
	UserRoleNominal      = "nominal"
)

// TargetType — тип целевого объекта (сервис, сервер, модуль, виртуальный)
const (
	TargetTypeService = "service"
	TargetTypeServer  = "server"
	TargetTypeModule  = "module"
	TargetTypeVirtual = "virtual"
)

// Target представляет собой сервис, сервер, модуль или виртуальный объект

type TargetLink struct {
	Value string `json:"value"`
	Type  string `json:"type"`
}
type Target struct {
	ID             string                           `gorm:"type:varchar(255);primaryKey" json:"id"`
	Type           string                           `gorm:"type:varchar(50);index;not null" json:"type"`
	Name           string                           `gorm:"type:varchar(255);index;not null" json:"name"`
	Description    *string                          `gorm:"type:text" json:"description"`
	Links          *datatypes.JSONSlice[TargetLink] `gorm:"type:json" json:"links" swaggertype:"array,object"`          // список ссылок
	Tags           *datatypes.JSONSlice[string]     `gorm:"type:json" json:"tags" swaggertype:"array,string"`           // массив тегов
	InternalLinks  *datatypes.JSONSlice[TargetLink] `gorm:"type:json" json:"internal_links" swaggertype:"array,object"` // список ссылок
	InternalTags   *datatypes.JSONSlice[string]     `gorm:"type:json" json:"internal_tags" swaggertype:"array,string"`  // массив внутренних тегов
	SynchronizedAt time.Time                        `gorm:"type:datetime(3)" json:"synchronized_at" validate:"required"`
	CreatedAt      time.Time                        `gorm:"type:datetime(3)" json:"created_at" validate:"required"`
	UpdatedAt      time.Time                        `gorm:"type:datetime(3)" json:"updated_at" validate:"required"`
}

// TargetRelation представляет связь «многие‑ко‑многим» между двумя Target
type TargetRelation struct {
	ID           uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	FromTargetID string    `gorm:"column:from_target_id;type:varchar(255);index;not null" json:"from_target_id"`
	ToTargetID   string    `gorm:"column:to_target_id;type:varchar(255);index;not null" json:"to_target_id"`
	RelationType string    `gorm:"column:relation_type;type:varchar(255);not null" json:"relation_type"`
	CreatedAt    time.Time `gorm:"type:datetime(3)" json:"created_at" validate:"required"`
	UpdatedAt    time.Time `gorm:"type:datetime(3)" json:"updated_at" validate:"required"`
}

type TargetUserBase struct {
	TargetID string          `gorm:"type:varchar(255);index;not null" json:"target_id"`
	UserID   uint            `gorm:"type:int" json:"user_id"`
	Roles    types.JsonStore `gorm:"type:json;not null" json:"roles"` // массив строк UserRole
}

type TargetAddonBase struct {
	TargetID  string               `gorm:"column:target_id;type:varchar(255);index;not null" json:"target_id"`
	AddonType connection.AddonType `gorm:"column:addon_type;type:varchar(50);not null" json:"addon_type"`
	AddonID   string               `gorm:"column:addon_id;type:varchar(36);not null" json:"addon_id"`
	Config    types.JsonStore      `gorm:"type:json;not null" json:"config"`
}

type TargetAddon struct {
	Base
	TargetAddonBase
}

type TargetUser struct {
	Base
	TargetUserBase
}
