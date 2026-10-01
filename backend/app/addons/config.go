package addons

import (
	"context"

	"github.com/cms-lab-core/cms-labs-api/backend/app/addons/vault"
	"github.com/cms-lab-core/cms-labs-api/backend/app/queries"
	"github.com/cms-lab-core/cms-labs-api/shared/connection"
)

type AddonOperationConfig struct {
	Name           string  `json:"name"`
	AdditionalName *string `json:"additionalName,omitempty"`
	URL            *string `json:"url,omitempty"`
	CurrentSizeMb  *int    `json:"currentSizeMb,omitempty"`
	MaxSizeMb      *int    `json:"maxSizeMb,omitempty"`
	Expired        *int64  `json:"expired,omitempty"`
}

type AddonService interface {
	Create(ctx context.Context, targetName string, name *string) (*AddonOperationConfig, error)
	Delete(ctx context.Context, targetName string, cfg *AddonOperationConfig) (*AddonOperationConfig, error)
	Reset(ctx context.Context, targetName string, cfg *AddonOperationConfig) (*AddonOperationConfig, error)
	GetType() connection.AddonType
}

type FactoryAddonService struct {
	Config              *connection.AddonsConfig
	VaultClient         vault.ClientInterface
	TokenAttemptQueries *queries.TokenAttemptQueries
	ServerQueries       *queries.ServerQueries
	UserQueries         *queries.UserQueries
	TargetQueries       *queries.TargetQueries
	IssId               string
}

func (f *FactoryAddonService) SetContext(issId string) *FactoryAddonService {
	f.IssId = issId
	return f
}

func (f *FactoryAddonService) GetAddonServiceByID(id string) AddonService {
	cfg := f.GetAddonConfigByID(id)
	return f.NewAddonService(cfg)
}

func (f *FactoryAddonService) GetAddonConfigByID(id string) *connection.AddonConfig {
	for _, addon := range f.Config.Addons {
		if addon.ID == id {
			return &addon
		}
	}
	return nil
}

func (f *FactoryAddonService) NewAddonService(cfg *connection.AddonConfig) AddonService {
	switch cfg.Type {
	case connection.MySQLAddon:
		return NewMySQLAddonService(cfg, f.VaultClient)
	case connection.PostgreSQLAddon:
		return NewPostgreSQLAddonService(cfg, f.VaultClient)
	case connection.S3Addon:
		return NewSeaweedFSAddonService(cfg, f.VaultClient)
	case connection.HarborAddon:
		return NewHarborAddonService(cfg, f.VaultClient)
	case connection.KubernetesAddon:
		kubernetes := KubernetesAddonService{
			IssId:               f.IssId,
			Config:              cfg,
			VaultClient:         f.VaultClient,
			TokenAttemptQueries: f.TokenAttemptQueries,
			ServerQueries:       f.ServerQueries,
			UserQueries:         f.UserQueries,
			TargetQueries:       f.TargetQueries,
		}
		return &kubernetes
	case connection.VaultAddon:
		return NewVaultAddonService(cfg, f.VaultClient)
	default:
		return nil
	}
}
