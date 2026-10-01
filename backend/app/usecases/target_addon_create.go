package usecases

import (
	"context"
	"errors"

	"github.com/cms-lab-core/cms-labs-api/backend/app/addons"
	"github.com/cms-lab-core/cms-labs-api/backend/app/models"
	"github.com/cms-lab-core/cms-labs-api/backend/app/models/types"
	"github.com/cms-lab-core/cms-labs-api/backend/app/queries"
	"github.com/cms-lab-core/cms-labs-api/shared/cms_client"
	"github.com/cms-lab-core/cms-labs-api/shared/jsonrpc"
	"github.com/goccy/go-json"
)

// TargetAddonCreateUC – подключение дополнения к цели
type TargetAddonCreateUC struct {
	TargetAddonQueries  *queries.TargetAddonQueries
	TargetQueries       *queries.TargetQueries
	TargetUserQueries   *queries.TargetUserQueries
	FactoryAddonService *addons.FactoryAddonService
	User                *cms_client.SSOTokenPublicData
}

// TargetAddonCreateInputDTO – входные данные
type TargetAddonCreateInputDTO struct {
	TargetID string  `json:"target_id" validate:"required"`
	AddonID  string  `json:"addon_id" validate:"required"`
	Name     *string `json:"name"`
	IssId    string  `json:"iss_id"`
}

// TargetAddonCreateOutputDTO – результат
type TargetAddonCreateOutputDTO struct {
	ID uint `json:"id"`
}

type TargetAddonCreateRequest struct {
	JSONRPC string                    `json:"jsonrpc" default:"2.0" required:"true"`
	Method  string                    `json:"method" default:"service_card.upsert" required:"true"`
	Params  TargetAddonCreateInputDTO `json:"params,omitempty"`
	ID      string                    `json:"id,omitempty" default:"1" required:"true"`
}

type TargetAddonCreateResponse struct {
	JSONRPC string                     `json:"jsonrpc" default:"2.0" required:"true"`
	Result  TargetAddonCreateOutputDTO `json:"result,omitempty"`
	Error   interface{}                `json:"error,omitempty"`
	ID      string                     `json:"id,omitempty" default:"1" required:"true"`
}

// Ошибки
var (
	ErrAddonConfigNotFound = jsonrpc.NewRpcError("addon_config_not_found", "Addon configuration not found")
	ErrAddonAlreadyExists  = jsonrpc.NewRpcError("addon_already_exists", "Addon already attached to this target")
)

func (u *TargetAddonCreateUC) SetContext(user *cms_client.SSOTokenPublicData) *TargetAddonCreateUC {
	u.User = user
	return u
}

// Execute – создание и подключение дополнения
func (u *TargetAddonCreateUC) Execute(dto TargetAddonCreateInputDTO) (*TargetAddonCreateOutputDTO, error) {
	u.FactoryAddonService.SetContext(dto.IssId)
	// 1. Проверка прав редактора
	if !u.TargetUserQueries.CheckTargetAndUserByRole(dto.TargetID, u.User.UserID(), models.UserRoleEditor) {
		return nil, ErrPermissionDenied
	}

	// 2. Проверяем существование цели
	target, err := u.TargetQueries.Get(dto.TargetID)
	if err != nil {
		return nil, err
	}

	// 3. Получаем конфигурацию дополнения
	addonService := u.FactoryAddonService.GetAddonServiceByID(dto.AddonID)
	if addonService == nil {
		return nil, ErrAddonConfigNotFound
	}

	// 4. Проверяем, не подключено ли уже
	_, err = u.TargetAddonQueries.GetByTargetAndAddon(dto.TargetID, dto.AddonID)
	if err == nil {
		return nil, ErrAddonAlreadyExists
	}
	if !errors.Is(err, queries.TargetAddonNotFoundError) && err != nil {
		return nil, err
	}

	// 6. Вызываем метод Create
	ctx := context.Background()
	createdConfig, err := addonService.Create(ctx, target.Name, dto.Name)
	if err != nil {
		return nil, jsonrpc.NewRpcError("addon_create_failed", "Failed to create addon instance: "+err.Error())
	}

	data, _ := json.Marshal(createdConfig)
	var configAddon types.JsonStore
	_ = json.Unmarshal(data, &configAddon)

	// 8. Сохраняем в БД
	entity := &models.TargetAddon{
		TargetAddonBase: models.TargetAddonBase{
			TargetID:  dto.TargetID,
			AddonType: addonService.GetType(),
			AddonID:   dto.AddonID,
			Config:    configAddon,
		},
	}
	if err := u.TargetAddonQueries.Upsert(entity); err != nil {
		// Попытка откатить создание во внешней системе
		_, _ = addonService.Delete(ctx, target.Name, createdConfig)
		return nil, err
	}

	return &TargetAddonCreateOutputDTO{ID: entity.ID}, nil
}
