package kubeconfig

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

func TestLoadPrefersInClusterConfig(t *testing.T) {
	expected := &rest.Config{Host: "https://kubernetes.default.svc"}
	actual, err := load(func() (*rest.Config, error) {
		return expected, nil
	}, filepath.Join(t.TempDir(), "missing"))

	require.NoError(t, err)
	assert.Same(t, expected, actual)
}

func TestLoadFallsBackToExplicitKubeconfig(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config")
	config := clientcmdapi.NewConfig()
	config.Clusters["local"] = &clientcmdapi.Cluster{
		Server:                "https://127.0.0.1:16443",
		InsecureSkipTLSVerify: true,
	}
	config.AuthInfos["developer"] = &clientcmdapi.AuthInfo{Token: "test-token"}
	config.Contexts["local"] = &clientcmdapi.Context{Cluster: "local", AuthInfo: "developer"}
	config.CurrentContext = "local"
	require.NoError(t, clientcmd.WriteToFile(*config, path))

	actual, err := load(func() (*rest.Config, error) {
		return nil, errors.New("not running in a pod")
	}, path)

	require.NoError(t, err)
	assert.Equal(t, "https://127.0.0.1:16443", actual.Host)
	assert.Equal(t, "test-token", actual.BearerToken)
}

func TestLoadReportsBothConfigErrors(t *testing.T) {
	_, err := load(func() (*rest.Config, error) {
		return nil, errors.New("not running in a pod")
	}, filepath.Join(t.TempDir(), "missing"))

	require.Error(t, err)
	assert.ErrorContains(t, err, "not running in a pod")
	assert.ErrorContains(t, err, "missing")
}
