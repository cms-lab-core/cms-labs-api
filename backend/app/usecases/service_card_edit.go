package usecases

import (
	"github.com/cms-lab-core/cms-labs-api/backend/app/models"
	"github.com/cms-lab-core/cms-labs-api/backend/app/queries"
)

type ServiceCardEditUC struct {
	ServiceCardQueries *queries.ServiceCardQueries
}

type ServiceCardEditInputDTO struct {
	ID          uint   `json:"id"`
	ImageUrl    string `json:"image_url"`
	Url         string `json:"url"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Order       uint   `json:"order"`
	IsActive    bool   `json:"is_active"`
}

type ServiceCardEditRequest struct {
	JSONRPC string                  `json:"jsonrpc" default:"2.0" required:"true"`
	Method  string                  `json:"method" default:"service_card.upsert" required:"true"`
	Params  ServiceCardEditInputDTO `json:"params,omitempty"`
	ID      string                  `json:"id,omitempty" default:"1" required:"true"`
}

type ServiceCardEditOutputDTO struct {
	ID uint `json:"id" required:"true"`
}

type ServiceCardEditResponse struct {
	JSONRPC string                    `json:"jsonrpc" default:"2.0" required:"true"`
	Result  ServiceCardDeleteInputDTO `json:"result,omitempty"`
	Error   interface{}               `json:"error,omitempty"`
	ID      string                    `json:"id,omitempty" default:"1" required:"true"`
}

func (u *ServiceCardEditUC) Execute(dto ServiceCardEditInputDTO) (ServiceCardEditOutputDTO, error) {
	entity := &models.ServiceCard{}
	entity.ID = dto.ID
	entity.ImageUrl = dto.ImageUrl
	entity.Url = dto.Url
	entity.Name = dto.Name
	entity.Description = dto.Description
	entity.Order = dto.Order
	entity.IsActive = dto.IsActive
	err := u.ServiceCardQueries.Upsert(entity)
	return ServiceCardEditOutputDTO{ID: entity.ID}, err
}
