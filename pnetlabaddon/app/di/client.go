package di

import (
	"github.com/cms-lab-core/cms-labs-api/pnetlabaddon/pkg/configs"
	"github.com/cms-lab-core/cms-labs-api/shared/cms_client"
	"github.com/cms-lab-core/cms-labs-api/shared/guacamole_client"
	resty "github.com/go-resty/resty/v2"
)

var (
	NewRestyClient = func() *resty.Client {
		return resty.New()
	}
)

func (di *DIContainer) CMSClient() *cms_client.CMSClient {
	return cms_client.NewCMSClient(
		configs.AppConfig.CMSClient,
		NewRestyClient(),
	)
}

func (di *DIContainer) GuacamoleClient() *guacamole_client.GuacamoleClient {
	return guacamole_client.NewGuacamoleClient(
		configs.AppConfig.GuacamoleClient,
		NewRestyClient(),
	)
}
