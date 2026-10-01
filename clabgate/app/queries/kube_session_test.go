package queries

import (
	"context"
	"strings"
	"testing"

	"github.com/rs/zerolog"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	dynamicfake "k8s.io/client-go/dynamic/fake"
	kubernetesfake "k8s.io/client-go/kubernetes/fake"
)

func TestSessionNamespace(t *testing.T) {
	tests := map[string]string{
		"UUID":      "lab-00000000-0000-0000-0000-000000000000",
		"mixedCase": "lab-sessionone",
	}
	for name, expected := range tests {
		t.Run(name, func(t *testing.T) {
			input := "00000000-0000-0000-0000-000000000000"
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
		AttemptID:       "00000000-0000-0000-0000-000000000000",
		OwnerID:         "42",
		Username:        "student",
		Title:           "Lab",
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
		"--ServerApp.identity_provider_class=cms_labs_jupyter.identity.ProxyIdentityProvider",
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
		"00000000-0000-0000-0000-000000000000",
		"https://git.example.test/group/cms-labs-simple-task.git",
		"main",
	)
	for _, expectedPart := range []string{
		"/clabgate/workspace/00000000-0000-0000-0000-000000000000/git-pull?",
		"branch=main",
		"repo=https%3A%2F%2Fgit.example.test%2Fgroup%2Fcms-labs-simple-task.git",
		"targetpath=task",
	} {
		if !strings.Contains(got, expectedPart) {
			t.Fatalf("workspace URL %q does not contain %q", got, expectedPart)
		}
	}
	// No urlpath may pin a directory that does not exist inside the clone, and
	// the only directory in the URL is the fixed task directory.
	if strings.Contains(got, "urlpath") || strings.Contains(got, "tree%2F") {
		t.Fatalf("workspace URL pins a directory inside the clone: %q", got)
	}
}

func TestBuildWorkspaceURLFallsBackToLabWithoutRepository(t *testing.T) {
	for _, tc := range []struct{ repository, ref string }{
		{"", "main"},
		{"https://git.example.test/group/tasks.git", ""},
		{"   ", "   "},
	} {
		got := buildWorkspaceURL("/clabgate/workspace", "attempt", tc.repository, tc.ref)
		if got != "/clabgate/workspace/attempt/lab" {
			t.Fatalf("buildWorkspaceURL(%q, %q) = %q", tc.repository, tc.ref, got)
		}
	}
}

func TestReadySessionUsesRepositoryBranchForNBGitPuller(t *testing.T) {
	replicas := int32(1)
	client := kubernetesfake.NewSimpleClientset(
		&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{
			Name: "lab-00000000-0000-0000-0000-000000000000",
			Labels: map[string]string{
				SessionManagedByLabel: SessionManagedByValue,
				SessionIDLabel:        "00000000-0000-0000-0000-000000000000",
				SessionOwnerLabel:     "42",
			},
			Annotations: map[string]string{
				SessionAttemptAnnotation:      "00000000-0000-0000-0000-000000000000",
				SessionTaskRepoAnnotation:     "https://github.com/cms-lab-core/cms-labs-simple-task",
				SessionTaskRefAnnotation:      "main",
				SessionTaskRevisionAnnotation: "d3136b13c7ae361eda4eaf16efd93d666b9bff0a",
				SessionTopologyAnnotation:     "false",
			},
		}},
		&appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{Name: workspaceName, Namespace: "lab-00000000-0000-0000-0000-000000000000"},
			Spec:       appsv1.DeploymentSpec{Replicas: &replicas},
			Status:     appsv1.DeploymentStatus{Replicas: 1, ReadyReplicas: 1},
		},
	)
	logger := zerolog.Nop()
	admin := NewKubernetesAdminWithClients(client, nil, nil, &logger)

	record, err := admin.GetSession(
		context.Background(),
		"00000000-0000-0000-0000-000000000000",
		"/clabgate/workspace",
	)
	if err != nil {
		t.Fatalf("GetSession returned error: %v", err)
	}
	if !strings.Contains(record.WorkspaceURL, "branch=main") {
		t.Fatalf("workspace URL does not use repository branch: %q", record.WorkspaceURL)
	}
	if strings.Contains(record.WorkspaceURL, "d3136b13c7ae361eda4eaf16efd93d666b9bff0a") {
		t.Fatalf("workspace URL incorrectly uses commit SHA as nbgitpuller branch: %q", record.WorkspaceURL)
	}
}

func TestCheckerResultStatusUsesScoreAndTasks(t *testing.T) {
	passed := &CheckerRun{
		MaxScore: 2, CurrentScore: 2,
		Tasks: []CheckerTaskResult{{Title: "one", Complete: true}, {Title: "two", Complete: true}},
	}
	if got := checkerResultStatus(passed, true); got != "passed" {
		t.Fatalf("checkerResultStatus() = %q, want passed", got)
	}

	partial := &CheckerRun{
		MaxScore: 2, CurrentScore: 1,
		Tasks: []CheckerTaskResult{{Title: "one", Complete: true}, {Title: "two", Complete: false}},
	}
	if got := checkerResultStatus(partial, true); got != "failed" {
		t.Fatalf("checkerResultStatus() = %q, want failed", got)
	}
	if got := checkerResultStatus(passed, false); got != "failed" {
		t.Fatalf("failed Job status = %q, want failed", got)
	}
}
