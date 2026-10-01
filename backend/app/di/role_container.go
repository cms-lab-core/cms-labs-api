package di

import (
	"github.com/cms-lab-core/cms-labs-api/backend/app/usecases"
)

func (di *DIContainer) RoleEditUC() *usecases.RoleEditUC {
	return &usecases.RoleEditUC{
		RoleQueries: di.Queries.RoleQueries,
	}
}

func (di *DIContainer) RoleListUC() *usecases.RoleListUC {
	return &usecases.RoleListUC{
		RoleQueries: di.Queries.RoleQueries,
	}
}

func (di *DIContainer) RoleDeleteUC() *usecases.RoleDeleteUC {
	return &usecases.RoleDeleteUC{
		RoleQueries: di.Queries.RoleQueries,
	}
}
