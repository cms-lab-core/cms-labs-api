package usecases

import (
	"github.com/cms-lab-core/cms-labs-api/backend/app/models"
	"github.com/cms-lab-core/cms-labs-api/backend/app/queries"
	"github.com/cms-lab-core/cms-labs-api/shared/connection"
)

// TargetGetUC – получение цели по ID с полной информацией (аддоны, пользователи)
type TargetGetUC struct {
	TargetQueries      *queries.TargetQueries
	TargetAddonQueries *queries.TargetAddonQueries
	TargetUserQueries  *queries.TargetUserQueries
	AddonsConfig       *connection.AddonsConfig
}

// TargetGetInputDTO – входные данные
type TargetGetInputDTO struct {
	ID string `json:"id" validate:"required"`
}

type TargetGetRequest struct {
	JSONRPC string            `json:"jsonrpc" default:"2.0" required:"true"`
	Method  string            `json:"method" default:"service_card.upsert" required:"true"`
	Params  TargetGetInputDTO `json:"params,omitempty"`
	ID      string            `json:"id,omitempty" default:"1" required:"true"`
}

// AvailableAddonInfo – информация о доступном аддоне
type AvailableAddonInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
}

// TargetUserInfo – информация о пользователе target
type TargetUserInfo struct {
	UserID uint     `json:"user_id"`
	Roles  []string `json:"roles"`
}

// ConnectedAddonInfo – информация о подключенном аддоне
type ConnectedAddonInfo struct {
	ID                   uint                   `json:"id"`
	AddonID              string                 `json:"addon_id"`
	Type                 string                 `json:"type"`
	RequestDeletedUserID *uint                  `json:"request_deleted_user_id"`
	Config               map[string]interface{} `json:"config"`
}

// TargetGetOutputDTO – выходные данные (расширенная модель)
type TargetGetOutputDTO struct {
	models.Target
	AvailableAddons []AvailableAddonInfo `json:"availableAddons"`
	ConnectedAddons []ConnectedAddonInfo `json:"connectedAddons"`
	TargetUsers     []TargetUserInfo     `json:"targetUsers"`
}

type TargetGetResponse struct {
	JSONRPC string             `json:"jsonrpc" default:"2.0" required:"true"`
	Result  TargetGetOutputDTO `json:"result,omitempty"`
	Error   interface{}        `json:"error,omitempty"`
	ID      string             `json:"id,omitempty" default:"1" required:"true"`
}

// Execute – возвращает полную модель Target с аддонами и пользователями
func (uc *TargetGetUC) Execute(dto TargetGetInputDTO) (TargetGetOutputDTO, error) {
	// Получаем target
	target, err := uc.TargetQueries.Get(dto.ID)
	if err != nil {
		return TargetGetOutputDTO{}, err
	}

	// Получаем подключенные аддоны
	connectedAddons, err := uc.TargetAddonQueries.GetAllByTarget(dto.ID)
	if err != nil {
		return TargetGetOutputDTO{}, err
	}

	// Получаем пользователей target
	targetUsers, err := uc.TargetUserQueries.GetAllByTarget(dto.ID)
	if err != nil {
		return TargetGetOutputDTO{}, err
	}

	// Формируем список доступных аддонов из конфига
	availableAddons := make([]AvailableAddonInfo, 0)
	if uc.AddonsConfig != nil {
		for _, addon := range uc.AddonsConfig.Addons {
			availableAddons = append(availableAddons, AvailableAddonInfo{
				ID:   addon.ID,
				Name: addon.Name,
				Type: string(addon.Type),
			})
		}
	}

	// Формируем список подключенных аддонов
	connectedAddonsResp := make([]ConnectedAddonInfo, 0)
	for _, ca := range connectedAddons {
		var requestDeletedUserID *uint
		if ca.Config != nil {
			if uid, ok := ca.Config["request_deleted_user_id"].(float64); ok && uid > 0 {
				uidUint := uint(uid)
				requestDeletedUserID = &uidUint
			}
		}
		connectedAddonsResp = append(connectedAddonsResp, ConnectedAddonInfo{
			ID:                   ca.ID,
			AddonID:              ca.AddonID,
			Type:                 string(ca.AddonType),
			RequestDeletedUserID: requestDeletedUserID,
			Config:               ca.Config,
		})
	}

	// Формируем список пользователей
	usersResp := make([]TargetUserInfo, 0)
	for _, tu := range targetUsers {
		roles := make([]string, 0)
		if rolesInterface, ok := tu.Roles["roles"]; ok {
			if rolesSlice, ok := rolesInterface.([]interface{}); ok {
				for _, r := range rolesSlice {
					if roleStr, ok := r.(string); ok {
						roles = append(roles, roleStr)
					}
				}
			}
		}
		usersResp = append(usersResp, TargetUserInfo{
			UserID: tu.UserID,
			Roles:  roles,
		})
	}

	return TargetGetOutputDTO{
		Target:          target,
		AvailableAddons: availableAddons,
		ConnectedAddons: connectedAddonsResp,
		TargetUsers:     usersResp,
	}, nil
}
