package di

import (
	"github.com/cms-lab-core/cms-labs-api/pnetlabaddon/app/usecases"
	"github.com/cms-lab-core/cms-labs-api/shared/logs"
)

func (di *DIContainer) SSOSecondFactorUC() *usecases.SSOSecondFactorUC {
	return &usecases.SSOSecondFactorUC{
		CMSClient:        di.CMSClient(),
		UserQueries:      di.Queries.UserQueries,
		UserRoleQueries:  di.Queries.UserRoleQueries,
		GuacamoleQueries: di.Queries.GuacamoleQueries,
		GuacamoleClient:  di.GuacamoleClient(),
		Logger:           logs.NewZeroLogger(di.ZeroLogConf.SetName("usecases.SSOSecondFactorUC")),
	}
}

func (di *DIContainer) LabCreateUC() *usecases.LabCreateUC {
	return &usecases.LabCreateUC{
		UserQueries:     di.Queries.UserQueries,
		LabSessionQuery: di.Queries.LabSessionQuery,
		CMSClient:       di.CMSClient(),
		Logger:          logs.NewZeroLogger(di.ZeroLogConf.SetName("usecases.LabCreateUC")),
	}
}

func (di *DIContainer) PnetServerPingUC() *usecases.PnetServerPingUC {
	return &usecases.PnetServerPingUC{
		LabSessionQuery: di.Queries.LabSessionQuery,
		CMSClient:       di.CMSClient(),
		Logger:          logs.NewZeroLogger(di.ZeroLogConf.SetName("usecases.PnetServerPingUC")),
	}
}
