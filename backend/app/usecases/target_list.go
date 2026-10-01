package usecases

import (
	"github.com/cms-lab-core/cms-labs-api/backend/app/models"
	"github.com/cms-lab-core/cms-labs-api/backend/app/queries"
	"github.com/cms-lab-core/cms-labs-api/shared/cms_client"
)

// TargetListUC – получение списка со связями
type TargetListUC struct {
	TargetQueries         *queries.TargetQueries
	TargetUserQueries     *queries.TargetUserQueries
	TargetRelationQueries *queries.TargetRelationQueries
	User                  *cms_client.SSOTokenPublicData
}

// TargetListInputDTO – входные данные
type TargetListInputDTO struct{}

type TargetItem struct {
	Target *models.Target `json:"taget" validate:"required"`
	IsMine bool           `json:"is_mine" validate:"required"`
}

type TargetRelationItem struct {
	Relation *models.TargetRelation `json:"relation"`
}

type TargetModel struct {
	Tagets    []*TargetItem         `json:"targets"`
	Relations []*TargetRelationItem `json:"relations"`
}

type TargetListOutputDTO struct {
	Model *TargetModel `json:"model"`
}

type TargetListRequest struct {
	JSONRPC string             `json:"jsonrpc" default:"2.0" required:"true"`
	Method  string             `json:"method" default:"service_card.upsert" required:"true"`
	Params  TargetListInputDTO `json:"params,omitempty"`
	ID      string             `json:"id,omitempty" default:"1" required:"true"`
}

type TargetListResponse struct {
	JSONRPC string              `json:"jsonrpc" default:"2.0" required:"true"`
	Result  TargetListOutputDTO `json:"result,omitempty"`
	Error   interface{}         `json:"error,omitempty"`
	ID      string              `json:"id,omitempty" default:"1" required:"true"`
}

func (uc *TargetListUC) SetContext(user *cms_client.SSOTokenPublicData) *TargetListUC {
	uc.User = user
	return uc
}

// Execute – возвращает полную модель Target
func (uc *TargetListUC) Execute(dto TargetListInputDTO) (*TargetListOutputDTO, error) {
	targets, err := uc.TargetQueries.List()
	if err != nil {
		return nil, err
	}
	targetsMineIds, err := uc.TargetUserQueries.GetByUserId(uc.User.UserID())
	targetsMine := make(map[string]bool)
	for _, targetMine := range targetsMineIds {
		targetsMine[targetMine] = true
	}
	if err != nil {
		return nil, err
	}
	relations, err := uc.TargetRelationQueries.List()
	if err != nil {
		return nil, err
	}
	model := &TargetModel{}
	for _, target := range targets {
		model.Tagets = append(
			model.Tagets,
			&TargetItem{
				Target: &target,
				IsMine: targetsMine[target.ID],
			},
		)
	}
	for _, relation := range relations {
		model.Relations = append(
			model.Relations,
			&TargetRelationItem{
				Relation: &relation,
			},
		)
	}
	return &TargetListOutputDTO{Model: model}, nil
}
