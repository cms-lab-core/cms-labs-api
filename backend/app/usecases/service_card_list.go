package usecases

import (
	"github.com/cms-lab-core/cms-labs-api/backend/app/models"
	"github.com/cms-lab-core/cms-labs-api/backend/app/queries"
)

type ServiceCardListUC struct {
	ServiceCardQueries *queries.ServiceCardQueries
}

type ServiceCardListInputDTO struct {
	OnlyActive bool   `json:"-"`
	Search     string `json:"search"`
	Limit      int    `json:"limit"`
	Offset     int    `json:"offset"`
}

type ServiceCardListOutputDTO struct {
	Model      []models.ServiceCardListItem `json:"model" validate:"required"`
	TotalCount int64                        `json:"total_count" validate:"required"`
}

type ServiceCardListRequest struct {
	JSONRPC string                  `json:"jsonrpc" default:"2.0" required:"true"`
	Method  string                  `json:"method" default:"service_card.list" required:"true"`
	Params  ServiceCardListInputDTO `json:"params,omitempty"`
	ID      string                  `json:"id,omitempty" default:"1" required:"true"`
}

type ServiceCardAvailableRequest struct {
	JSONRPC string                  `json:"jsonrpc" default:"2.0" required:"true"`
	Method  string                  `json:"method" default:"service_card.available" required:"true"`
	Params  ServiceCardListInputDTO `json:"params,omitempty"`
	ID      string                  `json:"id,omitempty" default:"1" required:"true"`
}

type ServiceCardListResponse struct {
	JSONRPC string                   `json:"jsonrpc" default:"2.0" required:"true"`
	Result  ServiceCardListOutputDTO `json:"result,omitempty"`
	Error   interface{}              `json:"error,omitempty"`
	ID      string                   `json:"id,omitempty" default:"1" required:"true"`
}

func (u *ServiceCardListUC) Execute(dto ServiceCardListInputDTO) (ServiceCardListOutputDTO, error) {
	if dto.Limit <= 0 || dto.Limit > queries.MaxLimitCount {
		dto.Limit = 100
	}
	if dto.Offset < 0 {
		dto.Offset = 0
	}
	entities, count, err := u.ServiceCardQueries.List(dto.Search, dto.Limit, dto.Offset, dto.OnlyActive)
	return ServiceCardListOutputDTO{
		Model:      entities,
		TotalCount: count,
	}, err
}
