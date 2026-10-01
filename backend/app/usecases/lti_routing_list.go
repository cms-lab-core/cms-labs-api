package usecases

import (
	"github.com/cms-lab-core/cms-labs-api/backend/app/models"
	"github.com/cms-lab-core/cms-labs-api/backend/app/queries"
)

type LTIRoutingListUC struct {
	LTIRoutingQueries *queries.LTIRoutingQueries
}

type LTIRoutingListInputDTO struct {
	Search string `json:"search"`
	Limit  int    `json:"limit"`
	Offset int    `json:"offset"`
}

type LTIRoutingListRequest struct {
	JSONRPC string                 `json:"jsonrpc" default:"2.0" required:"true"`
	Method  string                 `json:"method" default:"lti_routing.list" required:"true"`
	Params  LTIRoutingListInputDTO `json:"params,omitempty"`
	ID      string                 `json:"id,omitempty" default:"1" required:"true"`
}

type LTIRoutingListOutputDTO struct {
	Model      []models.LTIRoutingListItem `json:"model" validate:"required"`
	TotalCount int64                       `json:"total_count" validate:"required"`
}

type LTIRoutingListResponse struct {
	JSONRPC string                  `json:"jsonrpc" default:"2.0" required:"true"`
	Result  LTIRoutingListOutputDTO `json:"result,omitempty"`
	Error   interface{}             `json:"error,omitempty"`
	ID      string                  `json:"id,omitempty" default:"1" required:"true"`
}

func (u *LTIRoutingListUC) Execute(dto LTIRoutingListInputDTO) (LTIRoutingListOutputDTO, error) {
	entities, count, err := u.LTIRoutingQueries.List(dto.Search, dto.Limit, dto.Offset)
	return LTIRoutingListOutputDTO{
		Model:      entities,
		TotalCount: count,
	}, err
}
