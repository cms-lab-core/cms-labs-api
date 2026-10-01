package di

import (
	"github.com/cms-lab-core/cms-labs-api/backend/app/usecases/tasks"
	"github.com/cms-lab-core/cms-labs-api/backend/pkg/configs"
	"github.com/cms-lab-core/cms-labs-api/shared/logs"
)

func (di *DIContainer) TaskStartup() *tasks.StartupFiberUC {
	return &tasks.StartupFiberUC{
		UserDefaultCreateUC: di.TaskUserDefaultCreateUC(),
		ProxmoxSyncUC:       di.TaskProxmoxSyncUC(),
		LTISyncResultUC:     di.TaskLTISyncResultUC(),
		DemoSeedUC:          &tasks.DemoSeedUC{DB: di.Queries.DB, Logger: logs.NewZeroLogger(di.ZeroLogConf.SetName("tasks.DemoSeedUC"))},
	}
}

func (di *DIContainer) TaskUserDefaultCreateUC() *tasks.UserDefaultCreateUC {
	return &tasks.UserDefaultCreateUC{
		UserQueries:         di.Queries.UserQueries,
		UserPasswordQueries: di.Queries.UserPasswordQueries,
		RoleQueries:         di.Queries.RoleQueries,
		Logger:              logs.NewZeroLogger(di.ZeroLogConf.SetName("tasks.StartupFiberUC")),
	}
}

func (di *DIContainer) TaskProxmoxSyncUC() *tasks.ProxmoxSyncUC {
	return &tasks.ProxmoxSyncUC{
		Debug:                 configs.AppConfig.Debug,
		ProxmoxSyncConfig:     configs.AppConfig.ProxmoxSyncConfig,
		TargetQueries:         di.Queries.TargetQueries,
		TargetRelationQueries: di.Queries.TargetRelationQueries,
		Logger:                logs.NewZeroLogger(di.ZeroLogConf.SetName("tasks.ProxmoxSyncUC")),
	}
}

func (di *DIContainer) TaskLTISyncResultUC() *tasks.LTISyncResultUC {
	return &tasks.LTISyncResultUC{
		LTIAttemptQueries: di.Queries.LTIAttemptQueries,
		LTIConnectorAPI:   di.LTIProtocolDatastoreConfig(),
		Logger:            logs.NewZeroLogger(di.ZeroLogConf.SetName("tasks.LTISyncResultUC")),
	}
}
