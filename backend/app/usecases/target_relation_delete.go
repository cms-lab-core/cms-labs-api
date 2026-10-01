package usecases

import (
	"github.com/cms-lab-core/cms-labs-api/backend/app/models"
	"github.com/cms-lab-core/cms-labs-api/backend/app/queries"
	"github.com/cms-lab-core/cms-labs-api/shared/cms_client"
	"github.com/cms-lab-core/cms-labs-api/shared/jsonrpc"
)

// TargetRelationDeleteUC – удаление связи между целями
type TargetRelationDeleteUC struct {
	TargetRelationQueries *queries.TargetRelationQueries
	TargetUserQueries     *queries.TargetUserQueries
	User                  *cms_client.SSOTokenPublicData
}

// TargetRelationDeleteInputDTO – параметры удаления
type TargetRelationDeleteInputDTO struct {
	FromTargetID string `json:"from_target_id"`
	ToTargetID   string `json:"to_target_id"`
	RelationType string `json:"relation_type"`
}

type TargetRelationDeleteRequest struct {
	JSONRPC string                       `json:"jsonrpc" default:"2.0" required:"true"`
	Method  string                       `json:"method" default:"service_card.upsert" required:"true"`
	Params  TargetRelationDeleteInputDTO `json:"params,omitempty"`
	ID      string                       `json:"id,omitempty" default:"1" required:"true"`
}

type TargetRelationDeleteResponse struct {
	JSONRPC string                       `json:"jsonrpc" default:"2.0" required:"true"`
	Result  TargetRelationDeleteInputDTO `json:"result,omitempty"`
	Error   interface{}                  `json:"error,omitempty"`
	ID      string                       `json:"id,omitempty" default:"1" required:"true"`
}

func (uc *TargetRelationDeleteUC) SetContext(user *cms_client.SSOTokenPublicData) *TargetRelationDeleteUC {
	uc.User = user
	return uc
}

// Execute – удаляет связь
func (uc *TargetRelationDeleteUC) Execute(dto TargetRelationDeleteInputDTO) (*TargetRelationDeleteInputDTO, error) {
	if !uc.TargetUserQueries.CheckTargetAndUserByRole(dto.FromTargetID, uc.User.UserID(), models.UserRoleEditor) {
		return nil, ErrPermissionDenied
	}
	if dto.FromTargetID == "" || dto.ToTargetID == "" || dto.RelationType == "" {
		return nil, jsonrpc.NewRpcError("invalid_arguments", "Either id or (from,to,type) must be provided")
	}
	if dto.FromTargetID == dto.ToTargetID {
		return nil, jsonrpc.NewRpcError("invalid_arguments", "Either (from,to) must be not equal")
	}
	err := uc.TargetRelationQueries.DeleteByFromToType(dto.FromTargetID, dto.ToTargetID, dto.RelationType)
	if err != nil {
		return nil, err
	}
	return &dto, nil
}
