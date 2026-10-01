package usecases

import (
	"github.com/cms-lab-core/cms-labs-api/backend/app/models"
	"github.com/cms-lab-core/cms-labs-api/backend/app/queries"
)

type RoleListUC struct {
	RoleQueries *queries.RoleQueries
}

type RoleListInputDTO struct {
}

type RoleListRequest struct {
	JSONRPC string           `json:"jsonrpc" default:"2.0" required:"true"`
	Method  string           `json:"method" default:"role.list" required:"true"`
	Params  RoleListInputDTO `json:"params,omitempty"`
	ID      string           `json:"id,omitempty" default:"1" required:"true"`
}

type RoleListOutputDTO struct {
	Model      []models.Role `json:"model" validate:"required"`
	TotalCount int           `json:"total_count" validate:"required"`
}

type RoleListResponse struct {
	JSONRPC string            `json:"jsonrpc" default:"2.0" required:"true"`
	Result  RoleListOutputDTO `json:"result,omitempty"`
	Error   interface{}       `json:"error,omitempty"`
	ID      string            `json:"id,omitempty" default:"1" required:"true"`
}

func (u *RoleListUC) Execute(dto RoleListInputDTO) (RoleListOutputDTO, error) {
	entities, err := u.RoleQueries.List()
	return RoleListOutputDTO{
		Model:      entities,
		TotalCount: len(entities),
	}, err
}
