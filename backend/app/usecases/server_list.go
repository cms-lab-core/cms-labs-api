package usecases

import (
	"github.com/cms-lab-core/cms-labs-api/backend/app/models"
	"github.com/cms-lab-core/cms-labs-api/backend/app/queries"
)

type ServerListUC struct {
	ServerQueries *queries.ServerQueries
	RoleQueries   *queries.RoleQueries
}

type ServerListModel struct {
	Model models.ServerListItem `json:"model" validate:"required"`
	Roles []uint                `json:"roles"`
}

type ServerListOutputDTO struct {
	Model      []ServerListModel `json:"model" validate:"required"`
	TotalCount int64             `json:"total_count" validate:"required"`
}

type ServerListRequest struct {
	JSONRPC string                       `json:"jsonrpc" default:"2.0" required:"true"`
	Method  string                       `json:"method" default:"server.list" required:"true"`
	Params  queries.ServerQueriesListDTO `json:"params,omitempty"`
	ID      string                       `json:"id,omitempty" default:"1" required:"true"`
}

type ServerListResponse struct {
	JSONRPC string              `json:"jsonrpc" default:"2.0" required:"true"`
	Result  ServerListOutputDTO `json:"result,omitempty"`
	Error   interface{}         `json:"error,omitempty"`
	ID      string              `json:"id,omitempty" default:"1" required:"true"`
}

func (u *ServerListUC) Execute(dto queries.ServerQueriesListDTO) (ServerListOutputDTO, error) {
	entities, count, err := u.ServerQueries.List(dto)
	if err != nil {
		return ServerListOutputDTO{}, err
	}
	serverIDS := make([]uint, 0)
	for _, entity := range entities {
		serverIDS = append(serverIDS, entity.ID)
	}
	hmap, err := u.RoleQueries.GetByRelationServerIds(serverIDS)
	if err != nil {
		return ServerListOutputDTO{}, err
	}
	result := ServerListOutputDTO{TotalCount: count}
	for _, entity := range entities {
		item := ServerListModel{
			Model: entity,
		}
		if val, ok := hmap[entity.ID]; ok {
			item.Roles = val
		}
		result.Model = append(result.Model, item)
	}
	return result, err
}
