package usecases

import (
	"context"

	"github.com/cms-lab-core/cms-labs-api/backend/app/addons"
	"github.com/cms-lab-core/cms-labs-api/backend/app/models"
	"github.com/cms-lab-core/cms-labs-api/backend/app/models/types"
	"github.com/cms-lab-core/cms-labs-api/backend/app/queries"
	"github.com/cms-lab-core/cms-labs-api/shared/cms_client"
	"github.com/cms-lab-core/cms-labs-api/shared/jsonrpc"
	"github.com/goccy/go-json"
)

// TargetAddonResetUC – сброс состояния дополнения
type TargetAddonResetUC struct {
	TargetAddonQueries  *queries.TargetAddonQueries
	TargetUserQueries   *queries.TargetUserQueries
	TargetQueries       *queries.TargetQueries
	FactoryAddonService *addons.FactoryAddonService
	User                *cms_client.SSOTokenPublicData
}

// TargetAddonResetInputDTO – параметры сброса
type TargetAddonResetInputDTO struct {
	TargetID string `json:"target_id" validate:"required"`
	AddonID  string `json:"addon_id" validate:"required"`
	IssId    string `json:"iss_id"`
}

// TargetAddonResetOutputDTO – результат
type TargetAddonResetOutputDTO struct {
	ID uint `json:"id"`
}

type TargetAddonResetRequest struct {
	JSONRPC string                   `json:"jsonrpc" default:"2.0" required:"true"`
	Method  string                   `json:"method" default:"service_card.upsert" required:"true"`
	Params  TargetAddonResetInputDTO `json:"params,omitempty"`
	ID      string                   `json:"id,omitempty" default:"1" required:"true"`
}

type TargetAddonResetResponse struct {
	JSONRPC string                    `json:"jsonrpc" default:"2.0" required:"true"`
	Result  TargetAddonResetOutputDTO `json:"result,omitempty"`
	Error   interface{}               `json:"error,omitempty"`
	ID      string                    `json:"id,omitempty" default:"1" required:"true"`
}

func (uc *TargetAddonResetUC) SetContext(user *cms_client.SSOTokenPublicData) *TargetAddonResetUC {
	uc.User = user
	return uc
}

// Execute – выполняет reset дополнения
func (uc *TargetAddonResetUC) Execute(dto TargetAddonResetInputDTO) (*TargetAddonResetOutputDTO, error) {
	uc.FactoryAddonService.SetContext(dto.IssId)
	if !uc.TargetUserQueries.CheckTargetAndUserByRole(dto.TargetID, uc.User.UserID(), models.UserRoleEditor) {
		return nil, ErrPermissionDenied
	}
	addon, err := uc.TargetAddonQueries.GetByTargetAndAddon(dto.TargetID, dto.AddonID)
	if err != nil {
		return nil, err
	}
	target, err := uc.TargetQueries.Get(dto.TargetID)
	if err != nil {
		return nil, err
	}
	addonService := uc.FactoryAddonService.GetAddonServiceByID(dto.AddonID)
	if addonService == nil {
		return nil, ErrAddonConfigNotFound
	}
	data, _ := json.Marshal(addon.Config)
	var configAddon addons.AddonOperationConfig
	_ = json.Unmarshal(data, &configAddon)
	ctx := context.Background()
	addonNewConfig, err := addonService.Reset(ctx, target.Name, &configAddon)
	if err != nil {
		return nil, jsonrpc.NewRpcError("addon_reset_failed", "Failed to reset addon: "+err.Error())
	}
	data, _ = json.Marshal(addonNewConfig)
	var newConfigAddon types.JsonStore
	_ = json.Unmarshal(data, &newConfigAddon)
	addon.Config = newConfigAddon
	if err := uc.TargetAddonQueries.Upsert(&addon); err != nil {
		return nil, err
	}
	return &TargetAddonResetOutputDTO{ID: addon.ID}, nil
}
