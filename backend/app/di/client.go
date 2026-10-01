package di

import (
	"github.com/cms-lab-core/cms-labs-api/backend/app/addons/vault"
	"github.com/cms-lab-core/cms-labs-api/backend/pkg/configs"
	"github.com/cms-lab-core/cms-labs-api/shared/logs"
	resty "github.com/go-resty/resty/v2"
)

var (
	NewRestyClient = func() *resty.Client {
		return resty.New()
	}
	NewVaultClient = func() vault.ClientInterface {
		return vault.NewClient(
			configs.AppConfig.Vault,
			logs.NewZeroLogger(&logs.ZeroLoggerConf{}),
		)
	}
)
