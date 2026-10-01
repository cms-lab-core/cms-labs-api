package di

import (
	"github.com/cms-lab-core/cms-labs-api/backend/app/usecases"
)

func (di *DIContainer) LTIRoutingEditUC() *usecases.LTIRoutingEditUC {
	return &usecases.LTIRoutingEditUC{
		LTIRoutingQueries: di.Queries.LTIRoutingQueries,
	}
}

func (di *DIContainer) LTIRoutingGetUC() *usecases.LTIRoutingGetUC {
	return &usecases.LTIRoutingGetUC{
		LTIRoutingQueries: di.Queries.LTIRoutingQueries,
	}
}

func (di *DIContainer) LTIRoutingListUC() *usecases.LTIRoutingListUC {
	return &usecases.LTIRoutingListUC{
		LTIRoutingQueries: di.Queries.LTIRoutingQueries,
	}
}

func (di *DIContainer) LTIRoutingDeleteUC() *usecases.LTIRoutingDeleteUC {
	return &usecases.LTIRoutingDeleteUC{
		LTIRoutingQueries: di.Queries.LTIRoutingQueries,
	}
}

func (di *DIContainer) LabCatalogUC() *usecases.LabCatalogUC {
	return &usecases.LabCatalogUC{
		LTIRoutingQueries: di.Queries.LTIRoutingQueries,
		LTIAttemptQueries: di.Queries.LTIAttemptQueries,
		AttemptCreateUC:   di.LTIAttemptCreateUC(),
	}
}
