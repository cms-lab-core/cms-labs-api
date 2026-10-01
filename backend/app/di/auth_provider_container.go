package di

import (
	"github.com/cms-lab-core/cms-labs-api/backend/app/usecases"
)

func (di *DIContainer) AuthProviderEditUC() *usecases.AuthProviderEditUC {
	return &usecases.AuthProviderEditUC{
		AuthProviderQueries: di.Queries.AuthProviderQueries,
	}
}

func (di *DIContainer) AuthProviderGetUC() *usecases.AuthProviderGetUC {
	return &usecases.AuthProviderGetUC{
		AuthProviderQueries: di.Queries.AuthProviderQueries,
	}
}

func (di *DIContainer) AuthProviderListUC() *usecases.AuthProviderListUC {
	return &usecases.AuthProviderListUC{
		AuthProviderQueries: di.Queries.AuthProviderQueries,
	}
}

func (di *DIContainer) AuthProviderDeleteUC() *usecases.AuthProviderDeleteUC {
	return &usecases.AuthProviderDeleteUC{
		AuthProviderQueries: di.Queries.AuthProviderQueries,
	}
}
