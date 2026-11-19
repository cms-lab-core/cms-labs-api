package configs

import (
	"os"
	"path/filepath"
	"testing"
)

func TestJupyterPodConfigFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	if got := loadJupyterPodConfig(path); got.ExtraAnnotations != nil || got.NodeSelector != nil || got.Affinity != nil {
		t.Fatal("missing file must leave the pod unchanged")
	}
	data := []byte(`extraAnnotations:
  example.org/pool: jupyter
  example.org/empty: ""
  example.org/json: '{"key":"value"}'
nodeSelector:
  kubernetes.io/os: linux
affinity:
  nodeAffinity:
    requiredDuringSchedulingIgnoredDuringExecution:
      nodeSelectorTerms:
        - matchExpressions:
            - key: kubernetes.io/hostname
              operator: In
              values:
                - k8s-prod-jup-1-cms
`)
	if err := os.WriteFile(path, data, 0600); err != nil {
		t.Fatal(err)
	}
	got := loadJupyterPodConfig(path)
	if got.ExtraAnnotations["example.org/json"] != `{"key":"value"}` || got.ExtraAnnotations["example.org/empty"] != "" || len(got.ExtraAnnotations) != 3 {
		t.Fatalf("annotations changed: %#v", got.ExtraAnnotations)
	}
	if got.NodeSelector["kubernetes.io/os"] != "linux" {
		t.Fatalf("selector ignored: %#v", got.NodeSelector)
	}
	if got.Affinity == nil || got.Affinity.NodeAffinity == nil {
		t.Fatal("node affinity ignored")
	}
	required := got.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution
	if required == nil || len(required.NodeSelectorTerms) != 1 || required.NodeSelectorTerms[0].MatchExpressions[0].Values[0] != "k8s-prod-jup-1-cms" {
		t.Fatalf("node allowlist ignored: %#v", required)
	}
}

func TestJupyterPodConfigRejectsInvalidFile(t *testing.T) {
	for _, data := range []string{"extraAnnotation: {}", "extraAnnotations: []", "affinity: invalid", "extraAnnotations: {"} {
		t.Run(data, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "config.yaml")
			if err := os.WriteFile(path, []byte(data), 0600); err != nil {
				t.Fatal(err)
			}
			defer func() {
				if recover() == nil {
					t.Fatal("invalid pod config was silently accepted")
				}
			}()
			loadJupyterPodConfig(path)
		})
	}
}
