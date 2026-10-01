package di

import (
	"github.com/cms-lab-core/cms-labs-api/backend/app/usecases"
)

func (di *DIContainer) ServiceCardEditUC() *usecases.ServiceCardEditUC {
	return &usecases.ServiceCardEditUC{
		ServiceCardQueries: di.Queries.ServiceCardQueries,
	}
}

func (di *DIContainer) ServiceCardGetUC() *usecases.ServiceCardGetUC {
	return &usecases.ServiceCardGetUC{
		ServiceCardQueries: di.Queries.ServiceCardQueries,
	}
}

func (di *DIContainer) ServiceCardListUC() *usecases.ServiceCardListUC {
	return &usecases.ServiceCardListUC{
		ServiceCardQueries: di.Queries.ServiceCardQueries,
	}
}

func (di *DIContainer) ServiceCardDeleteUC() *usecases.ServiceCardDeleteUC {
	return &usecases.ServiceCardDeleteUC{
		ServiceCardQueries: di.Queries.ServiceCardQueries,
	}
}
