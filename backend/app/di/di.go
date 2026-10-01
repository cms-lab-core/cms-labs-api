package di

import (
	"github.com/cms-lab-core/cms-labs-api/backend/platform/database"
	"github.com/cms-lab-core/cms-labs-api/shared/logs"
)

type DIContainer struct {
	Queries     *database.Queries
	ZeroLogConf *logs.ZeroLoggerConf
}

func (di *DIContainer) Close() {
	_ = di.Queries.Close()
}

func NewDIContainer(zeroLogConf *logs.ZeroLoggerConf) (*DIContainer, error) {
	dbLogger := logs.NewZeroLogger(zeroLogConf.SetName("db"))
	queries, err := database.OpenDBConnection(dbLogger)
	di := &DIContainer{
		Queries:     queries,
		ZeroLogConf: zeroLogConf,
	}
	if err != nil {
		dbLogger.Warn().Err(err).Msg("failed to open database connection")
		return di, err
	}
	return di, nil
}
