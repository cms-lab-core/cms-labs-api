package usecases

import (
	"context"

	"github.com/cms-lab-core/cms-labs-api/backend/app/addons"
	"github.com/cms-lab-core/cms-labs-api/backend/app/models"
	"github.com/cms-lab-core/cms-labs-api/backend/app/queries"
	"github.com/cms-lab-core/cms-labs-api/shared/cms_client"
	"github.com/goccy/go-json"
)

// TargetAddonDeleteUC – отключение дополнения от цели
type TargetAddonDeleteUC struct {
	TargetAddonQueries  *queries.TargetAddonQueries
	TargetQueries       *queries.TargetQueries
	TargetUserQueries   *queries.TargetUserQueries
	FactoryAddonService *addons.FactoryAddonService
	User                *cms_client.SSOTokenPublicData
}

// TargetAddonDeleteInputDTO – параметры удаления
type TargetAddonDeleteInputDTO struct {
	TargetID string `json:"target_id" validate:"required"`
	AddonID  string `json:"addon_id" validate:"required"`
	IssId    string `json:"iss_id"`
	Revoke   bool   `json:"revoke"`
}

type TargetAddonDeleteOutputDTO struct {
	PendingConfirmation bool   `json:"pending_confirmation"`
	Message             string `json:"message"`
}

type TargetAddonDeleteRequest struct {
	JSONRPC string                    `json:"jsonrpc" default:"2.0" required:"true"`
	Method  string                    `json:"method" default:"service_card.upsert" required:"true"`
	Params  TargetAddonDeleteInputDTO `json:"params,omitempty"`
	ID      string                    `json:"id,omitempty" default:"1" required:"true"`
}

type TargetAddonDeleteResponse struct {
	JSONRPC string                     `json:"jsonrpc" default:"2.0" required:"true"`
	Result  TargetAddonDeleteOutputDTO `json:"result,omitempty"`
	Error   interface{}                `json:"error,omitempty"`
	ID      string                     `json:"id,omitempty" default:"1" required:"true"`
}

func (uc *TargetAddonDeleteUC) SetContext(user *cms_client.SSOTokenPublicData) *TargetAddonDeleteUC {
	uc.User = user
	return uc
}

// Execute – отключает дополнение
func (uc *TargetAddonDeleteUC) Execute(dto TargetAddonDeleteInputDTO) (*TargetAddonDeleteOutputDTO, error) {
	uc.FactoryAddonService.SetContext(dto.IssId)
	if !uc.TargetUserQueries.CheckTargetAndUserByRole(dto.TargetID, uc.User.UserID(), models.UserRoleEditor) {
		return nil, ErrPermissionDenied
	}
	target, err := uc.TargetQueries.Get(dto.TargetID)
	if err != nil {
		return nil, err
	}
	addon, err := uc.TargetAddonQueries.GetByTargetAndAddon(dto.TargetID, dto.AddonID)
	if err != nil {
		return nil, err
	}

	currentUserID := uc.User.UserID()

	config := addon.Config
	if config == nil {
		config = make(map[string]interface{})
	}

	existingRequestUserID, _ := config["request_deleted_user_id"].(float64)

	if dto.Revoke {
		if existingRequestUserID > 0 {
			delete(config, "request_deleted_user_id")
			addon.Config = config
			err = uc.TargetAddonQueries.Update(&addon)
			if err != nil {
				return nil, err
			}
			return &TargetAddonDeleteOutputDTO{
				PendingConfirmation: false,
				Message:             "Deletion request revoked.",
			}, nil
		}
		return &TargetAddonDeleteOutputDTO{
			PendingConfirmation: false,
			Message:             "No pending deletion request.",
		}, nil
	}

	if existingRequestUserID > 0 {
		if uint(existingRequestUserID) == currentUserID {
			return &TargetAddonDeleteOutputDTO{
				PendingConfirmation: true,
				Message:             "You have already requested deletion. Waiting for another user to confirm.",
			}, nil
		}

		addonService := uc.FactoryAddonService.GetAddonServiceByID(dto.AddonID)
		if addonService == nil {
			return nil, ErrAddonConfigNotFound
		}
		data, _ := json.Marshal(addon.Config)
		var configAddon addons.AddonOperationConfig
		_ = json.Unmarshal(data, &configAddon)
		ctx := context.Background()
		_, err = addonService.Delete(ctx, target.Name, &configAddon)
		if err != nil {
			return nil, err
		}
		return &TargetAddonDeleteOutputDTO{}, uc.TargetAddonQueries.Delete(addon.ID)
	}

	config["request_deleted_user_id"] = float64(currentUserID)
	addon.Config = config
	err = uc.TargetAddonQueries.Update(&addon)
	if err != nil {
		return nil, err
	}

	return &TargetAddonDeleteOutputDTO{
		PendingConfirmation: true,
		Message:             "Deletion requested. Another user must confirm.",
	}, nil
}
