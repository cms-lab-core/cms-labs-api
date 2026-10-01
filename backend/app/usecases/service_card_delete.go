package usecases

import (
	"github.com/cms-lab-core/cms-labs-api/backend/app/queries"
)

type ServiceCardDeleteUC struct {
	ServiceCardQueries *queries.ServiceCardQueries
}

type ServiceCardDeleteInputDTO struct {
	ID uint `json:"id" required:"true"`
}

type ServiceCardDeleteRequest struct {
	JSONRPC string                    `json:"jsonrpc" default:"2.0" required:"true"`
	Method  string                    `json:"method" default:"service_card.list" required:"true"`
	Params  ServiceCardDeleteInputDTO `json:"params,omitempty"`
	ID      string                    `json:"id,omitempty" default:"1" required:"true"`
}

type ServiceCardDeleteResponse struct {
	JSONRPC string                    `json:"jsonrpc" default:"2.0" required:"true"`
	Result  ServiceCardDeleteInputDTO `json:"result,omitempty"`
	Error   interface{}               `json:"error,omitempty"`
	ID      string                    `json:"id,omitempty" default:"1" required:"true"`
}

func (u *ServiceCardDeleteUC) Execute(dto ServiceCardDeleteInputDTO) (ServiceCardDeleteInputDTO, error) {
	err := u.ServiceCardQueries.Delete(dto.ID)
	return dto, err
}
