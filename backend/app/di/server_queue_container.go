package di

import (
	"github.com/cms-lab-core/cms-labs-api/backend/app/usecases/server_queue"
	"github.com/cms-lab-core/cms-labs-api/shared/logs"
)

func (di *DIContainer) ServerChangeDistributionUC() *server_queue.ServerQueueUpsert {
	return &server_queue.ServerQueueUpsert{
		ServerQueries:         di.Queries.ServerQueries,
		RoundQueuePoolQueries: di.Queries.RoundQueuePoolQueries,
		Logger:                logs.NewZeroLogger(di.ZeroLogConf.SetName("server_queue.ServerQueueUpsert")),
	}
}

func (di *DIContainer) ServerListDistributionUC() *server_queue.ServerQueueList {
	return &server_queue.ServerQueueList{
		RoundQueuePoolQueries: di.Queries.RoundQueuePoolQueries,
		Logger:                logs.NewZeroLogger(di.ZeroLogConf.SetName("server_queue.ServerQueueList")),
	}
}
