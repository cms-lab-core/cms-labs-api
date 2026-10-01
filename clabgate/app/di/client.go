package di

import (
	"github.com/cms-lab-core/cms-labs-api/clabgate/app/queries"
)

func (di *DIContainer) KubernetesAdmin() (*queries.KubernetesAdminQuery, error) {
	kubeQuery, err := queries.NewKubernetesAdmin(
		di.ZeroLogConf.SetName("queries.KubernetesAdminQuery"),
	)
	return kubeQuery, err
}
