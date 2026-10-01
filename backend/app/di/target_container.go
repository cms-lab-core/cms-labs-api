package di

import (
	"github.com/cms-lab-core/cms-labs-api/backend/app/addons"
	"github.com/cms-lab-core/cms-labs-api/backend/app/addons/vault"
	"github.com/cms-lab-core/cms-labs-api/backend/app/usecases"
	"github.com/cms-lab-core/cms-labs-api/backend/pkg/configs"
	"github.com/cms-lab-core/cms-labs-api/shared/logs"
)

func (di *DIContainer) NewVaultClient() vault.ClientInterface {
	return NewVaultClient()
}
func (di *DIContainer) FactoryAddonService() *addons.FactoryAddonService {
	return &addons.FactoryAddonService{
		Config:              configs.AppConfig.AddonsConfig,
		VaultClient:         di.NewVaultClient(),
		TokenAttemptQueries: di.Queries.TokenAttemptQueries,
		ServerQueries:       di.Queries.ServerQueries,
		UserQueries:         di.Queries.UserQueries,
		TargetQueries:       di.Queries.TargetQueries,
		IssId:               "",
	}
}

func (di *DIContainer) TargetAddonCreateUC() *usecases.TargetAddonCreateUC {
	return &usecases.TargetAddonCreateUC{
		TargetAddonQueries:  di.Queries.TargetAddonQueries,
		TargetQueries:       di.Queries.TargetQueries,
		TargetUserQueries:   di.Queries.TargetUserQueries,
		FactoryAddonService: di.FactoryAddonService(),
	}
}

func (di *DIContainer) TargetAddonDeleteUC() *usecases.TargetAddonDeleteUC {
	return &usecases.TargetAddonDeleteUC{
		TargetAddonQueries:  di.Queries.TargetAddonQueries,
		TargetQueries:       di.Queries.TargetQueries,
		TargetUserQueries:   di.Queries.TargetUserQueries,
		FactoryAddonService: di.FactoryAddonService(),
	}
}

func (di *DIContainer) TargetAddonResetUC() *usecases.TargetAddonResetUC {
	return &usecases.TargetAddonResetUC{
		TargetAddonQueries:  di.Queries.TargetAddonQueries,
		TargetQueries:       di.Queries.TargetQueries,
		TargetUserQueries:   di.Queries.TargetUserQueries,
		FactoryAddonService: di.FactoryAddonService(),
	}
}

func (di *DIContainer) TargetDeleteUC() *usecases.TargetDeleteUC {
	return &usecases.TargetDeleteUC{
		TargetQueries:     di.Queries.TargetQueries,
		TargetUserQueries: di.Queries.TargetUserQueries,
	}
}

func (di *DIContainer) TargetGetUC() *usecases.TargetGetUC {
	return &usecases.TargetGetUC{
		TargetQueries:      di.Queries.TargetQueries,
		TargetAddonQueries: di.Queries.TargetAddonQueries,
		TargetUserQueries:  di.Queries.TargetUserQueries,
		AddonsConfig:       configs.AppConfig.AddonsConfig,
	}
}

func (di *DIContainer) TargetListUC() *usecases.TargetListUC {
	return &usecases.TargetListUC{
		TargetQueries:         di.Queries.TargetQueries,
		TargetUserQueries:     di.Queries.TargetUserQueries,
		TargetRelationQueries: di.Queries.TargetRelationQueries,
	}
}

func (di *DIContainer) TargetRelationCreateUC() *usecases.TargetRelationCreateUC {
	return &usecases.TargetRelationCreateUC{
		TargetRelationQueries: di.Queries.TargetRelationQueries,
		TargetUserQueries:     di.Queries.TargetUserQueries,
	}
}

func (di *DIContainer) TargetRelationDeleteUC() *usecases.TargetRelationDeleteUC {
	return &usecases.TargetRelationDeleteUC{
		TargetRelationQueries: di.Queries.TargetRelationQueries,
		TargetUserQueries:     di.Queries.TargetUserQueries,
	}
}

func (di *DIContainer) TargetUpsertUC() *usecases.TargetUpsertUC {
	return &usecases.TargetUpsertUC{
		TargetQueries:     di.Queries.TargetQueries,
		TargetUserQueries: di.Queries.TargetUserQueries,
	}
}

func (di *DIContainer) TargetUserDeleteUC() *usecases.TargetUserDeleteUC {
	return &usecases.TargetUserDeleteUC{
		TargetQueries:            di.Queries.TargetQueries,
		TargetUserQueries:        di.Queries.TargetUserQueries,
		TargetUserRotateAddonsUC: di.TargetUserRotateAddonsUC(),
	}
}

func (di *DIContainer) TargetUserRotateAddonsUC() *usecases.TargetUserRotateAddonsUC {
	return &usecases.TargetUserRotateAddonsUC{
		TargetUserQueries:   di.Queries.TargetUserQueries,
		TargetQueries:       di.Queries.TargetQueries,
		UserQueries:         di.Queries.UserQueries,
		VaultClient:         di.NewVaultClient(),
		TargetAddonQueries:  di.Queries.TargetAddonQueries,
		FactoryAddonService: di.FactoryAddonService(),
		Logger: logs.NewZeroLogger(
			di.ZeroLogConf.SetName("usecases.TargetUserRotateAddonsUC"),
		),
	}
}

func (di *DIContainer) TargetUserUpsertUC() *usecases.TargetUserUpsertUC {
	return &usecases.TargetUserUpsertUC{
		TargetQueries:            di.Queries.TargetQueries,
		TargetUserQueries:        di.Queries.TargetUserQueries,
		TargetUserRotateAddonsUC: di.TargetUserRotateAddonsUC(),
	}
}
