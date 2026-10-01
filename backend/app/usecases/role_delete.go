package usecases

import (
	"github.com/cms-lab-core/cms-labs-api/backend/app/queries"
)

type RoleDeleteUC struct {
	RoleQueries *queries.RoleQueries
}

type RoleDeleteInputDTO struct {
	ID uint `json:"id" validate:"required"`
}

type RoleDeleteRequest struct {
	JSONRPC string             `json:"jsonrpc" default:"2.0" required:"true"`
	Method  string             `json:"method" default:"role.delete" required:"true"`
	Params  RoleDeleteInputDTO `json:"params,omitempty"`
	ID      string             `json:"id,omitempty" default:"1" required:"true"`
}
type RoleDeleteResponse struct {
	JSONRPC string             `json:"jsonrpc" default:"2.0" required:"true"`
	Result  RoleDeleteInputDTO `json:"result,omitempty"`
	Error   interface{}        `json:"error,omitempty"`
	ID      string             `json:"id,omitempty" default:"1" required:"true"`
}

func (u *RoleDeleteUC) Execute(dto RoleDeleteInputDTO) (RoleDeleteInputDTO, error) {
	err := u.RoleQueries.Delete(dto.ID)
	return RoleDeleteInputDTO{ID: dto.ID}, err
}
