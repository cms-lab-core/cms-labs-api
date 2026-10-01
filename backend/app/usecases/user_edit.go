package usecases

import (
	"time"

	"github.com/cms-lab-core/cms-labs-api/backend/app/models/types"

	"github.com/cms-lab-core/cms-labs-api/backend/app/models"
	"github.com/cms-lab-core/cms-labs-api/backend/app/queries"
	"github.com/ory/go-convenience/stringsx"
)

type UserEditUC struct {
	UserQueries *queries.UserQueries
	RoleQueries *queries.RoleQueries
}

type UserEditInputDTO struct {
	ID        uint            `json:"id"`
	Name      string          `json:"name" validate:"required"`
	Email     string          `json:"email" validate:"required"`
	LTIUserID string          `json:"lti_user_id"`
	Roles     []uint          `json:"roles"`
	GroupName string          `json:"group_name"`
	Store     types.JsonStore `json:"store"`
	IsActive  bool            `json:"is_active"`
}

type UserEditRequest struct {
	JSONRPC string           `json:"jsonrpc" default:"2.0" required:"true"`
	Method  string           `json:"method" default:"user.upsert" required:"true"`
	Params  UserEditInputDTO `json:"params,omitempty"`
	ID      string           `json:"id,omitempty" default:"1" required:"true"`
}

type UserEditOutputDTO struct {
	ID uint `json:"id" required:"true"`
}

type UserEditResponse struct {
	JSONRPC string            `json:"jsonrpc" default:"2.0" required:"true"`
	Result  UserEditOutputDTO `json:"result,omitempty"`
	Error   interface{}       `json:"error,omitempty"`
	ID      string            `json:"id,omitempty" default:"1" required:"true"`
}

func (u *UserEditUC) Execute(dto UserEditInputDTO) (UserEditOutputDTO, error) {
	now := time.Now().UTC()
	userFromDB, _ := u.UserQueries.Get(dto.ID)
	entity := &models.User{}
	entity.ID = dto.ID
	entity.Email = stringsx.Coalesce(dto.Email, userFromDB.Email)
	entity.Name = stringsx.Coalesce(dto.Name, userFromDB.Name)
	entity.LTIUserID = stringsx.Coalesce(dto.LTIUserID, userFromDB.LTIUserID)
	entity.GroupName = stringsx.Coalesce(dto.GroupName, userFromDB.GroupName)
	entity.DeletedAt = userFromDB.DeletedAt
	entity.LastLaunchID = userFromDB.LastLaunchID
	entity.Store = dto.Store
	if dto.IsActive {
		entity.DeletedAt = nil
	}
	if !dto.IsActive && entity.IsActive() {
		entity.DeletedAt = &now
	}
	err := u.UserQueries.Upsert(entity)
	if err != nil {
		return UserEditOutputDTO{}, err
	}
	err = u.RoleQueries.SetByUserId(entity.ID, dto.Roles)
	return UserEditOutputDTO{ID: entity.ID}, err
}
