package usecases

import (
	"github.com/cms-lab-core/cms-labs-api/backend/app/models"
	"github.com/cms-lab-core/cms-labs-api/backend/app/queries"
	"github.com/cms-lab-core/cms-labs-api/shared/cms_client"
)

// TargetUserDeleteUC – удаление пользователя из цели
type TargetUserDeleteUC struct {
	TargetUserQueries        *queries.TargetUserQueries
	TargetQueries            *queries.TargetQueries
	TargetUserRotateAddonsUC *TargetUserRotateAddonsUC
	User                     *cms_client.SSOTokenPublicData
}

// TargetUserDeleteInputDTO – параметры удаления
type TargetUserDeleteInputDTO struct {
	TargetID string `json:"target_id" validate:"required"`
	UserID   uint   `json:"user_id" validate:"required"`
}

type TargetUserDeleteOutputDTO struct {
}

type TargetUserDeleteRequest struct {
	JSONRPC string                   `json:"jsonrpc" default:"2.0" required:"true"`
	Method  string                   `json:"method" default:"service_card.upsert" required:"true"`
	Params  TargetUserDeleteInputDTO `json:"params,omitempty"`
	ID      string                   `json:"id,omitempty" default:"1" required:"true"`
}

type TargetUserDeleteResponse struct {
	JSONRPC string                    `json:"jsonrpc" default:"2.0" required:"true"`
	Result  TargetUserDeleteOutputDTO `json:"result,omitempty"`
	Error   interface{}               `json:"error,omitempty"`
	ID      string                    `json:"id,omitempty" default:"1" required:"true"`
}

func (uc *TargetUserDeleteUC) SetContext(user *cms_client.SSOTokenPublicData) *TargetUserDeleteUC {
	uc.User = user
	return uc
}

// Execute – удаляет пользователя из цели
func (uc *TargetUserDeleteUC) Execute(dto TargetUserDeleteInputDTO) (*TargetUserDeleteOutputDTO, error) {
	// 1. Проверка прав: удаляющий должен быть редактором цели
	if !uc.TargetUserQueries.CheckTargetAndUserByRole(dto.TargetID, uc.User.UserID(), models.UserRoleEditor) {
		return nil, ErrPermissionDenied
	}
	// 4. Удаляем из БД
	err := uc.TargetUserQueries.DeleteByTargetAndUser(dto.TargetID, dto.UserID)
	if err != nil {
		return nil, err
	}
	err = uc.TargetUserRotateAddonsUC.Execute(dto.UserID, dto.TargetID)
	if err != nil {
		return nil, err
	}
	return &TargetUserDeleteOutputDTO{}, nil
}
