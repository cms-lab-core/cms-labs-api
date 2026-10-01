package di

import (
	"time"

	"github.com/cms-lab-core/cms-labs-api/clabgate/app/queries"
	"github.com/cms-lab-core/cms-labs-api/clabgate/app/usecases"
	"github.com/cms-lab-core/cms-labs-api/clabgate/pkg/configs"
	"github.com/cms-lab-core/cms-labs-api/shared/cms_client"
	"github.com/cms-lab-core/cms-labs-api/shared/logs"
	resty "github.com/go-resty/resty/v2"
)

func (di *DIContainer) SessionsUC() (*usecases.SessionsUC, error) {
	kubeQuery, err := di.KubernetesAdmin()
	if err != nil {
		return nil, err
	}
	cmsConfig := configs.AppConfig.CMS
	cmsClient := cms_client.NewCMSClient(&cms_client.CMSClientConfig{
		Debug:             configs.AppConfig.Debug,
		MaxTimeoutSeconds: cmsConfig.MaxTimeoutSeconds,
		ClientID:          cmsConfig.ClientID,
		Token:             cmsConfig.Token,
		BaseUrl:           cmsConfig.BaseURL,
	}, resty.New())
	repositoryClient := resty.New().
		SetTimeout(time.Duration(cmsConfig.MaxTimeoutSeconds) * time.Second).
		SetRedirectPolicy(resty.NoRedirectPolicy())
	catalog := queries.NewLabCatalog(
		configs.AppConfig.Session.TaskRepositoryURL,
		configs.AppConfig.Session.TaskBranch,
		configs.AppConfig.Session.TaskRepositoryToken,
		repositoryClient,
	)
	return &usecases.SessionsUC{
		Logger:               logs.NewZeroLogger(di.ZeroLogConf.SetName("usecases.SessionsUC")),
		KubernetesAdminQuery: kubeQuery,
		CMSClient:            cmsClient,
		LabCatalog:           catalog,
	}, nil
}
