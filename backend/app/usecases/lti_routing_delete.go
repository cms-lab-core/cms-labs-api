package usecases

import (
	"github.com/cms-lab-core/cms-labs-api/backend/app/queries"
)

type LTIRoutingDeleteUC struct {
	LTIRoutingQueries *queries.LTIRoutingQueries
}

type LTIRoutingDeleteInputDTO struct {
	ID uint `json:"id" required:"true"`
}

type LTIRoutingDeleteRequest struct {
	JSONRPC string                   `json:"jsonrpc" default:"2.0" required:"true"`
	Method  string                   `json:"method" default:"lti_routing.delete" required:"true"`
	Params  LTIRoutingDeleteInputDTO `json:"params,omitempty"`
	ID      string                   `json:"id,omitempty" default:"1" required:"true"`
}

type LTIRoutingDeleteResponse struct {
	JSONRPC string                   `json:"jsonrpc" default:"2.0" required:"true"`
	Result  LTIRoutingDeleteInputDTO `json:"result,omitempty"`
	Error   interface{}              `json:"error,omitempty"`
	ID      string                   `json:"id,omitempty" default:"1" required:"true"`
}

func (u *LTIRoutingDeleteUC) Execute(dto LTIRoutingDeleteInputDTO) (LTIRoutingDeleteInputDTO, error) {
	err := u.LTIRoutingQueries.Delete(dto.ID)
	return dto, err
}
