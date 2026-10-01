package usecases

import (
	"github.com/cms-lab-core/cms-labs-api/backend/app/models"
	"github.com/cms-lab-core/cms-labs-api/backend/app/queries"
	"github.com/cms-lab-core/cms-labs-api/shared/cms_client"
	"github.com/cms-lab-core/cms-labs-api/shared/jsonrpc"
)

// TargetRelationCreateUC – создание связи между двумя целями
type TargetRelationCreateUC struct {
	TargetRelationQueries *queries.TargetRelationQueries
	TargetUserQueries     *queries.TargetUserQueries
	User                  *cms_client.SSOTokenPublicData
}

// TargetRelationCreateInputDTO – параметры связи
type TargetRelationCreateInputDTO struct {
	FromTargetID string `json:"from_target_id" validate:"required"`
	ToTargetID   string `json:"to_target_id" validate:"required"`
	RelationType string `json:"relation_type" validate:"required"`
}

// TargetRelationCreateOutputDTO – результат
type TargetRelationCreateOutputDTO struct {
	ID uint `json:"id"`
}

type TargetRelationCreateRequest struct {
	JSONRPC string                       `json:"jsonrpc" default:"2.0" required:"true"`
	Method  string                       `json:"method" default:"service_card.upsert" required:"true"`
	Params  TargetRelationCreateInputDTO `json:"params,omitempty"`
	ID      string                       `json:"id,omitempty" default:"1" required:"true"`
}

type TargetRelationCreateResponse struct {
	JSONRPC string                        `json:"jsonrpc" default:"2.0" required:"true"`
	Result  TargetRelationCreateOutputDTO `json:"result,omitempty"`
	Error   interface{}                   `json:"error,omitempty"`
	ID      string                        `json:"id,omitempty" default:"1" required:"true"`
}

func (uc *TargetRelationCreateUC) SetContext(user *cms_client.SSOTokenPublicData) *TargetRelationCreateUC {
	uc.User = user
	return uc
}

// Execute – основной метод
func (uc *TargetRelationCreateUC) Execute(dto TargetRelationCreateInputDTO) (*TargetRelationCreateOutputDTO, error) {
	if !uc.TargetUserQueries.CheckTargetAndUserByRole(dto.FromTargetID, uc.User.UserID(), models.UserRoleEditor) {
		return nil, ErrPermissionDenied
	}
	if dto.FromTargetID == dto.ToTargetID {
		return nil, jsonrpc.NewRpcError("invalid_arguments", "Either (from,to) must be not equal")
	}
	entity := &models.TargetRelation{
		FromTargetID: dto.FromTargetID,
		ToTargetID:   dto.ToTargetID,
		RelationType: dto.RelationType,
	}

	if err := uc.TargetRelationQueries.Upsert(entity); err != nil {
		return nil, err
	}
	return &TargetRelationCreateOutputDTO{ID: entity.ID}, nil
}
