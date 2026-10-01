package usecases

import (
	"github.com/cms-lab-core/cms-labs-api/backend/app/models"
	"github.com/cms-lab-core/cms-labs-api/backend/app/queries"
)

type ServiceCardGetUC struct {
	ServiceCardQueries *queries.ServiceCardQueries
}

type ServiceCardGetInputDTO struct {
	ID uint `json:"id" required:"true"`
}

type ServiceCardGetOutputDTO struct {
	Model models.ServiceCard `json:"model" required:"true"`
}

type ServiceCardGetRequest struct {
	JSONRPC string                 `json:"jsonrpc" default:"2.0" required:"true"`
	Method  string                 `json:"method" default:"service_card.get" required:"true"`
	Params  ServiceCardGetInputDTO `json:"params,omitempty"`
	ID      string                 `json:"id,omitempty" default:"1" required:"true"`
}

type ServiceCardGetResponse struct {
	JSONRPC string                  `json:"jsonrpc" default:"2.0" required:"true"`
	Result  ServiceCardGetOutputDTO `json:"result,omitempty"`
	Error   interface{}             `json:"error,omitempty"`
	ID      string                  `json:"id,omitempty" default:"1" required:"true"`
}

func (u *ServiceCardGetUC) Execute(dto ServiceCardGetInputDTO) (ServiceCardGetOutputDTO, error) {
	form, err := u.ServiceCardQueries.Get(dto.ID)
	return ServiceCardGetOutputDTO{
		Model: form,
	}, err
}
