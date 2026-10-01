package di

import (
	"github.com/cms-lab-core/cms-labs-api/shared/logs"
)

type DIContainer struct {
	ZeroLogConf *logs.ZeroLoggerConf
}

func (di *DIContainer) Close() {
}

func NewDIContainer(zeroLogConf *logs.ZeroLoggerConf) (*DIContainer, error) {
	di := &DIContainer{
		ZeroLogConf: zeroLogConf,
	}
	return di, nil
}
