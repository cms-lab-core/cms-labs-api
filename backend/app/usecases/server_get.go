package usecases

import (
	"github.com/cms-lab-core/cms-labs-api/backend/app/models"
	"github.com/cms-lab-core/cms-labs-api/backend/app/queries"
)

type ServerGetUC struct {
	ServerQueries *queries.ServerQueries
	RoleQueries   *queries.RoleQueries
}

type ServerGetInputDTO struct {
	ID uint `json:"id" required:"true"`
}

type ServerGetRequest struct {
	JSONRPC string            `json:"jsonrpc" default:"2.0" required:"true"`
	Method  string            `json:"method" default:"server.get" required:"true"`
	Params  ServerGetInputDTO `json:"params,omitempty"`
	ID      string            `json:"id,omitempty" default:"1" required:"true"`
}

type ServerGetOutputDTO struct {
	Model models.Server `json:"model" required:"true"`
	Roles []uint        `json:"roles"`
}

type ServerGetResponse struct {
	JSONRPC string             `json:"jsonrpc" default:"2.0" required:"true"`
	Result  ServerGetOutputDTO `json:"result,omitempty"`
	Error   interface{}        `json:"error,omitempty"`
	ID      string             `json:"id,omitempty" default:"1" required:"true"`
}

func (u *ServerGetUC) Execute(dto ServerGetInputDTO) (ServerGetOutputDTO, error) {
	form, err := u.ServerQueries.Get(dto.ID)
	if err != nil {
		return ServerGetOutputDTO{}, err
	}
	result := ServerGetOutputDTO{
		Model: form,
	}
	roles, err := u.RoleQueries.GetByRelationServerIds([]uint{dto.ID})
	if err != nil {
		return result, err
	}
	if val, ok := roles[dto.ID]; ok {
		result.Roles = val
	}
	return result, err
}
