package usecases

import (
	"github.com/cms-lab-core/cms-labs-api/backend/app/models"
	"github.com/cms-lab-core/cms-labs-api/backend/app/queries"
)

type LTIRoutingGetUC struct {
	LTIRoutingQueries *queries.LTIRoutingQueries
}

type LTIRoutingGetInputDTO struct {
	ID uint `json:"id" required:"true"`
}

type LTIRoutingGetRequest struct {
	JSONRPC string                `json:"jsonrpc" default:"2.0" required:"true"`
	Method  string                `json:"method" default:"lti_routing.get" required:"true"`
	Params  LTIRoutingGetInputDTO `json:"params,omitempty"`
	ID      string                `json:"id,omitempty" default:"1" required:"true"`
}

type LTIRoutingGetOutputDTO struct {
	Model models.LTIRouting `json:"model" required:"true"`
}

type LTIRoutingGetResponse struct {
	JSONRPC string                 `json:"jsonrpc" default:"2.0" required:"true"`
	Result  LTIRoutingGetOutputDTO `json:"result,omitempty"`
	Error   interface{}            `json:"error,omitempty"`
	ID      string                 `json:"id,omitempty" default:"1" required:"true"`
}

func (u *LTIRoutingGetUC) Execute(dto LTIRoutingGetInputDTO) (LTIRoutingGetOutputDTO, error) {
	form, err := u.LTIRoutingQueries.Get(dto.ID)
	return LTIRoutingGetOutputDTO{
		Model: form,
	}, err
}
