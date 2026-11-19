package configs

import (
	"fmt"
	"os"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/yaml"
)

type JupyterPodConfig struct {
	ExtraAnnotations map[string]string `json:"extraAnnotations"`
	NodeSelector     map[string]string `json:"nodeSelector"`
	Affinity         *corev1.Affinity  `json:"affinity"`
}

func loadJupyterPodConfig(path string) JupyterPodConfig {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return JupyterPodConfig{}
	}
	if err != nil {
		panic(fmt.Sprintf("read Jupyter pod config %s: %v", path, err))
	}
	var config JupyterPodConfig
	if err := yaml.UnmarshalStrict(data, &config); err != nil {
		panic(fmt.Sprintf("invalid Jupyter pod config %s: %v", path, err))
	}
	return config
}
