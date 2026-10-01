package di

import (
	"github.com/cms-lab-core/cms-labs-api/backend/app/usecases"
	"github.com/cms-lab-core/cms-labs-api/backend/app/usecases/auth"
	"github.com/cms-lab-core/cms-labs-api/backend/platform/database"
	"github.com/cms-lab-core/cms-labs-api/shared/logs"
)

func (di *DIContainer) authTokenManager(db *database.Queries) *auth.TokenManager {
	return &auth.TokenManager{
		UserQueries:         db.UserQueries,
		RoleQueries:         db.RoleQueries,
		ServerQueries:       db.ServerQueries,
		UserPasswordQueries: db.UserPasswordQueries,
		TokenAttemptQueries: db.TokenAttemptQueries,
	}
}

func (di *DIContainer) AuthTokenManager() *auth.TokenManager {
	return di.authTokenManager(di.Queries)
}

func (di *DIContainer) SSOTokenUC() *auth.SSOTokenUC {
	return &auth.SSOTokenUC{
		TokenAttemptQueries: di.Queries.TokenAttemptQueries,
		TokenManager:        di.authTokenManager(di.Queries),
		ServerQueries:       di.Queries.ServerQueries,
		Logger:              logs.NewZeroLogger(di.ZeroLogConf.SetName("auth.SSOTokenUC")),
	}
}

func (di *DIContainer) SSOIntrospectUC() *auth.SSOIntrospectUC {
	return &auth.SSOIntrospectUC{
		TokenAttemptQueries: di.Queries.TokenAttemptQueries,
		ServerQueries:       di.Queries.ServerQueries,
	}
}

func (di *DIContainer) SSOAuthorizeUC() *auth.SSOAuthorizeUC {
	return &auth.SSOAuthorizeUC{
		TokenAttemptQueries: di.Queries.TokenAttemptQueries,
		ServerQueries:       di.Queries.ServerQueries,
		RoleQueries:         di.Queries.RoleQueries,
		Logger:              logs.NewZeroLogger(di.ZeroLogConf.SetName("auth.SSOAuthorizeUC")),
	}
}

func (di *DIContainer) SSOOpenidConfigurationUC() *auth.SSOOpenidConfiguration {
	return &auth.SSOOpenidConfiguration{}
}

func (di *DIContainer) SSOJwksUC() *auth.SSOJwksUC {
	return &auth.SSOJwksUC{}
}

func (di *DIContainer) ServiceAuthorizeUC() *auth.ServiceAuthorizeUC {
	return &auth.ServiceAuthorizeUC{
		ServerQueries: di.Queries.ServerQueries,
		Logger:        logs.NewZeroLogger(di.ZeroLogConf.SetName("auth.ServiceAuthorizeUC")),
	}
}

func (di *DIContainer) UserLoginUC() *usecases.UserLoginUC {
	return &usecases.UserLoginUC{
		TokenManager:        di.AuthTokenManager(),
		AuthProviderQueries: di.Queries.AuthProviderQueries,
		UserQueries:         di.Queries.UserQueries,
		IssId:               "",
		Logger:              logs.NewZeroLogger(di.ZeroLogConf.SetName("usecases.UserLoginUC")),
	}
}
