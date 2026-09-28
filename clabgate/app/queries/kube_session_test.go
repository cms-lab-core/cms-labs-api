package queries

import (
	"context"
	"strings"
	"testing"

	"github.com/rs/zerolog"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	dynamicfake "k8s.io/client-go/dynamic/fake"
	kubernetesfake "k8s.io/client-go/kubernetes/fake"
)

func TestSessionNamespace(t *testing.T) {
	tests := map[string]string{
		"UUID":      "lab-550e8400-e29b-41d4-a716-446655440000",
		"mixedCase": "lab-sessionone",
	}
	for name, expected := range tests {
		t.Run(name, func(t *testing.T) {
			input := "550e8400-e29b-41d4-a716-446655440000"
			if name == "mixedCase" {
				input = "SessionOne"
			}
			if got := SessionNamespace(input); got != expected {
				t.Fatalf("SessionNamespace(%q) = %q, want %q", input, got, expected)
			}
		})
	}

	longName := SessionNamespace(strings.Repeat("x", 100))
	if len(longName) > 63 || !strings.HasPrefix(longName, "lab-") {
		t.Fatalf("long session namespace is not a DNS label: %q", longName)
	}
}

func TestEnsureSessionSupportsMultiDocumentTopology(t *testing.T) {
	client := kubernetesfake.NewSimpleClientset()
	dynamicClient := dynamicfake.NewSimpleDynamicClientWithCustomListKinds(
		runtime.NewScheme(),
		map[schema.GroupVersionResource]string{topologyGVR: "TopologyList"},
	)
	logger := zerolog.Nop()
	admin := NewKubernetesAdminWithClients(client, dynamicClient, nil, &logger)
	manifest := `apiVersion: v1
kind: ConfigMap
metadata:
  name: startup
data:
  config: test
---
apiVersion: c9s.run/v1alpha1
kind: Topology
metadata:
  name: $NAME
spec:
  definition:
    containerlab: |
      name: $NAME
      topology:
        nodes: {}
`
	params := EnsureSessionParams{
		AttemptID:       "550e8400-e29b-41d4-a716-446655440000",
		OwnerID:         "42",
		Username:        "student",
		Title:           "Lab",
		LabPath:         "modules/lab/notebook.ipynb",
		TopologyYAML:    manifest,
		JupyterImage:    "example.test/jupyter:latest",
		StorageSize:     "1Gi",
		WorkspacePrefix: "/clabgate/workspace",
		TaskRepository:  "https://git.example.test/tasks",
		TaskRef:         "master",
		TaskRevision:    "0123456789abcdef",
	}

	record, err := admin.EnsureSession(context.Background(), params)
	if err != nil {
		t.Fatalf("EnsureSession returned error: %v", err)
	}
	if record.Namespace != "lab-"+params.AttemptID || record.TopologyReady {
		t.Fatalf("unexpected initial session state: %+v", record)
	}

	namespace := record.Namespace
	if _, err := client.CoreV1().ConfigMaps(namespace).Get(context.Background(), "startup", metav1.GetOptions{}); err != nil {
		t.Fatalf("startup ConfigMap was not created: %v", err)
	}
	if _, err := dynamicClient.Resource(topologyGVR).Namespace(namespace).Get(context.Background(), namespace, metav1.GetOptions{}); err != nil {
		t.Fatalf("Topology was not created: %v", err)
	}
	deployment, err := client.AppsV1().Deployments(namespace).Get(context.Background(), workspaceName, metav1.GetOptions{})
	if err != nil {
		t.Fatalf("Jupyter Deployment was not created: %v", err)
	}
	if deployment.Spec.Template.Spec.AutomountServiceAccountToken == nil || *deployment.Spec.Template.Spec.AutomountServiceAccountToken {
		t.Fatal("Jupyter must not receive a Kubernetes service account token")
	}
	if deployment.Spec.Template.Spec.EnableServiceLinks == nil || *deployment.Spec.Template.Spec.EnableServiceLinks {
		t.Fatal("Jupyter service links must be disabled to avoid the JUPYTER_PORT environment collision")
	}
	container := deployment.Spec.Template.Spec.Containers[0]
	if container.Image != params.JupyterImage {
		t.Fatalf("Jupyter image = %q, want %q", container.Image, params.JupyterImage)
	}
	wantArgs := []string{
		"start-notebook.py",
		"--ServerApp.base_url=/clabgate/workspace/" + params.AttemptID,
		"--ServerApp.allow_remote_access=True",
		"--IdentityProvider.token=",
	}
	if strings.Join(container.Args, "\x00") != strings.Join(wantArgs, "\x00") {
		t.Fatalf("Jupyter args = %q, want %q", container.Args, wantArgs)
	}
	if _, err := client.CoreV1().PersistentVolumeClaims(namespace).Get(context.Background(), workspaceName, metav1.GetOptions{}); err != nil {
		t.Fatalf("Jupyter PVC was not created: %v", err)
	}

	if _, err := admin.EnsureSession(context.Background(), params); err != nil {
		t.Fatalf("EnsureSession is not idempotent: %v", err)
	}
}

func TestEnsureTopologyValidatesAllDocumentsBeforeApply(t *testing.T) {
	client := kubernetesfake.NewSimpleClientset()
	dynamicClient := dynamicfake.NewSimpleDynamicClientWithCustomListKinds(
		runtime.NewScheme(),
		map[schema.GroupVersionResource]string{topologyGVR: "TopologyList"},
	)
	logger := zerolog.Nop()
	admin := NewKubernetesAdminWithClients(client, dynamicClient, nil, &logger)
	manifest := `apiVersion: v1
kind: ConfigMap
metadata:
  name: must-not-exist
---
apiVersion: v1
kind: Secret
metadata:
  name: forbidden
`
	if err := admin.ensureTopology(context.Background(), "lab-test", "session", "owner", manifest); err == nil {
		t.Fatal("unsupported manifest was accepted")
	}
	if _, err := client.CoreV1().ConfigMaps("lab-test").Get(context.Background(), "must-not-exist", metav1.GetOptions{}); err == nil {
		t.Fatal("ConfigMap was applied before the complete manifest was validated")
	}
}

func TestBuildWorkspaceURL(t *testing.T) {
	got := buildWorkspaceURL(
		"/clabgate/workspace",
		"550e8400-e29b-41d4-a716-446655440000",
		"https://git.example.test/tasks",
		"main",
		"module/Lab.ipynb",
	)
	for _, expectedPart := range []string{
		"/clabgate/workspace/550e8400-e29b-41d4-a716-446655440000/git-pull?",
		"repo=https%3A%2F%2Fgit.example.test%2Ftasks",
		"urlpath=lab%2Ftree%2Fmodule%2FLab.ipynb",
	} {
		if !strings.Contains(got, expectedPart) {
			t.Fatalf("workspace URL %q does not contain %q", got, expectedPart)
		}
	}
}
