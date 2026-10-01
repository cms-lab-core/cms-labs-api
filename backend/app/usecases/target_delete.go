package usecases

import (
	"github.com/cms-lab-core/cms-labs-api/backend/app/models"
	"github.com/cms-lab-core/cms-labs-api/backend/app/queries"
	"github.com/cms-lab-core/cms-labs-api/shared/cms_client"
)

// TargetDeleteUC – удаление цели
type TargetDeleteUC struct {
	TargetQueries     *queries.TargetQueries
	TargetUserQueries *queries.TargetUserQueries
	User              *cms_client.SSOTokenPublicData
}

// TargetDeleteInputDTO – параметры удаления
type TargetDeleteInputDTO struct {
	ID string `json:"id" validate:"required"`
}

type TargetDeleteOutputDTO struct {
}

type TargetDeleteRequest struct {
	JSONRPC string               `json:"jsonrpc" default:"2.0" required:"true"`
	Method  string               `json:"method" default:"service_card.upsert" required:"true"`
	Params  TargetDeleteInputDTO `json:"params,omitempty"`
	ID      string               `json:"id,omitempty" default:"1" required:"true"`
}

type TargetDeleteResponse struct {
	JSONRPC string                `json:"jsonrpc" default:"2.0" required:"true"`
	Result  TargetDeleteOutputDTO `json:"result,omitempty"`
	Error   interface{}           `json:"error,omitempty"`
	ID      string                `json:"id,omitempty" default:"1" required:"true"`
}

func (uc *TargetDeleteUC) SetContext(user *cms_client.SSOTokenPublicData) *TargetDeleteUC {
	uc.User = user
	return uc
}

// Execute – основной метод
func (uc *TargetDeleteUC) Execute(dto TargetDeleteInputDTO) (*TargetDeleteOutputDTO, error) {
	// Проверяем, является ли пользователь редактором цели
	if !uc.TargetUserQueries.CheckTargetAndUserByRole(dto.ID, uc.User.UserID(), models.UserRoleEditor) {
		return nil, ErrPermissionDenied
	}
	err := uc.TargetQueries.Delete(dto.ID)
	if err != nil {
		return nil, err
	}
	return &TargetDeleteOutputDTO{}, nil
}
