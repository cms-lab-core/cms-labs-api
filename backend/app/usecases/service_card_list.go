package usecases

import (
	"github.com/cms-lab-core/cms-labs-api/backend/app/models"
	"github.com/cms-lab-core/cms-labs-api/backend/app/queries"
)

type ServiceCardListUC struct {
	ServiceCardQueries *queries.ServiceCardQueries
}

type ServiceCardListInputDTO struct {
	Search string `json:"search"`
	Limit  int    `json:"limit"`
	Offset int    `json:"offset"`
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

type ServiceCardListResponse struct {
	JSONRPC string                   `json:"jsonrpc" default:"2.0" required:"true"`
	Result  ServiceCardListOutputDTO `json:"result,omitempty"`
	Error   interface{}              `json:"error,omitempty"`
	ID      string                   `json:"id,omitempty" default:"1" required:"true"`
}

func (u *ServiceCardListUC) Execute(dto ServiceCardListInputDTO) (ServiceCardListOutputDTO, error) {
	entities, count, err := u.ServiceCardQueries.List(dto.Search, dto.Limit, dto.Offset)
	return ServiceCardListOutputDTO{
		Model:      entities,
		TotalCount: count,
	}, err
}
