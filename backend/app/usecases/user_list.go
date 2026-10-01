package usecases

import (
	"github.com/cms-lab-core/cms-labs-api/backend/app/models"
	"github.com/cms-lab-core/cms-labs-api/backend/app/queries"
)

type UserListUC struct {
	UserQueries *queries.UserQueries
	RoleQueries *queries.RoleQueries
}

type UserListInputDTO struct {
	Search  string `json:"search"`
	UserIds []uint `json:"user_ids"`
	Limit   int    `json:"limit"`
	Offset  int    `json:"offset"`
}

type UserListModel struct {
	Model models.UserListItem `json:"model" validate:"required"`
	Roles []uint              `json:"roles"`
}
type UserListOutputDTO struct {
	Model      []UserListModel `json:"model" validate:"required"`
	TotalCount int64           `json:"total_count" validate:"required"`
}

type UserListRequest struct {
	JSONRPC string           `json:"jsonrpc" default:"2.0" required:"true"`
	Method  string           `json:"method" default:"user.get" required:"true"`
	Params  UserListInputDTO `json:"params,omitempty"`
	ID      string           `json:"id,omitempty" default:"1" required:"true"`
}

type UserListResponse struct {
	JSONRPC string            `json:"jsonrpc" default:"2.0" required:"true"`
	Result  UserListOutputDTO `json:"result,omitempty"`
	Error   interface{}       `json:"error,omitempty"`
	ID      string            `json:"id,omitempty" default:"1" required:"true"`
}

func (u *UserListUC) Execute(dto UserListInputDTO) (UserListOutputDTO, error) {
	entities, count, err := u.UserQueries.List(dto.Search, dto.UserIds, dto.Limit, dto.Offset)
	if err != nil {
		return UserListOutputDTO{}, err
	}
	usersIds := make([]uint, 0)
	for _, entity := range entities {
		usersIds = append(usersIds, entity.ID)
	}
	hmap, err := u.RoleQueries.GetByRelationUsersIds(usersIds)
	if err != nil {
		return UserListOutputDTO{}, err
	}
	result := UserListOutputDTO{TotalCount: count}
	for _, entity := range entities {
		item := UserListModel{
			Model: entity,
		}
		if val, ok := hmap[entity.ID]; ok {
			item.Roles = val
		}
		result.Model = append(result.Model, item)
	}
	return result, err
}
