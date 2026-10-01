package usecases

import (
	"github.com/cms-lab-core/cms-labs-api/backend/app/models"
	"github.com/cms-lab-core/cms-labs-api/backend/app/queries"
)

type RoleEditUC struct {
	RoleQueries *queries.RoleQueries
}

type RoleEditInputDTO struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
	Code string `json:"code"`
}

type RoleEditRequest struct {
	JSONRPC string           `json:"jsonrpc" default:"2.0" required:"true"`
	Method  string           `json:"method" default:"role.upsert" required:"true"`
	Params  RoleEditInputDTO `json:"params,omitempty"`
	ID      string           `json:"id,omitempty" default:"1" required:"true"`
}

type RoleEditOutputDTO struct {
	ID uint `json:"id" required:"true"`
}

type RoleEditResponse struct {
	JSONRPC string            `json:"jsonrpc" default:"2.0" required:"true"`
	Result  RoleEditOutputDTO `json:"result,omitempty"`
	Error   interface{}       `json:"error,omitempty"`
	ID      string            `json:"id,omitempty" default:"1" required:"true"`
}

func (u *RoleEditUC) Execute(dto RoleEditInputDTO) (RoleEditOutputDTO, error) {
	entity := &models.Role{}
	entity.ID = dto.ID
	entity.Name = dto.Name
	entity.Code = dto.Code
	err := u.RoleQueries.Upsert(entity)
	return RoleEditOutputDTO{ID: entity.ID}, err
}
