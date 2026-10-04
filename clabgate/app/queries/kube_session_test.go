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
		LabManifest:     manifest,
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
	environment := make(map[string]string, len(container.Env))
	for _, variable := range container.Env {
		environment[variable.Name] = variable.Value
	}
	if environment["CMS_LABS_SESSION_ID"] != params.AttemptID ||
		environment["CMS_LABS_CAPTURE_URL"] != "http://cms-labs-capture:8080" {
		t.Fatalf("Jupyter capture environment = %#v", environment)
	}
	if _, err := client.CoreV1().PersistentVolumeClaims(namespace).Get(context.Background(), workspaceName, metav1.GetOptions{}); err != nil {
		t.Fatalf("Jupyter PVC was not created: %v", err)
	}

	if _, err := admin.EnsureSession(context.Background(), params); err != nil {
		t.Fatalf("EnsureSession is not idempotent: %v", err)
	}
}

func TestEnsureLabManifestAppliesNamespacedResources(t *testing.T) {
	client := kubernetesfake.NewSimpleClientset()
	dynamicClient := dynamicfake.NewSimpleDynamicClientWithCustomListKinds(
		runtime.NewScheme(),
		map[schema.GroupVersionResource]string{topologyGVR: "TopologyList"},
	)
	logger := zerolog.Nop()
	admin := NewKubernetesAdminWithClients(client, dynamicClient, nil, &logger)
	manifest := `apiVersion: c9s.run/v1alpha1
kind: Topology
metadata:
  name: $NAME
spec:
  definition:
    containerlab: |
      name: $NAME
      topology:
        nodes: {}
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: lab-helper
automountServiceAccountToken: true
---
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: lab-helper
rules:
  - apiGroups: [""]
    resources: ["pods"]
    verbs: ["get", "list"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: lab-helper
subjects:
  - kind: ServiceAccount
    name: lab-helper
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: lab-helper
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: lab-helper
  finalizers: [task.example.test/retain]
data:
  namespace: $NAME
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: lab-helper
spec:
  replicas: 1
  selector:
    matchLabels: {app: lab-helper}
  template:
    metadata:
      labels: {app: lab-helper}
    spec:
      serviceAccountName: lab-helper
      containers:
        - name: helper
          image: example.test/lab-helper:1.0.0
          securityContext:
            allowPrivilegeEscalation: false
---
apiVersion: v1
kind: Service
metadata:
  name: lab-helper
spec:
  type: ClusterIP
  selector: {app: lab-helper}
  ports:
    - name: ttyd
      port: 7681
      targetPort: 7681
---
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: lab-helper
spec:
  podSelector:
    matchLabels: {app: lab-helper}
  policyTypes: [Ingress]
  ingress:
    - from:
        - namespaceSelector:
            matchLabels:
              kubernetes.io/metadata.name: $WORKSPACE_PROXY_NAMESPACE
      ports:
        - protocol: TCP
          port: 7681
`
	params := EnsureSessionParams{
		AttemptID:               "session",
		OwnerID:                 "owner",
		LabManifest:             manifest,
		WorkspaceProxyNamespace: "cms-labs-system",
	}
	const namespace = "lab-session"
	if err := admin.ensureLabManifest(context.Background(), namespace, params); err != nil {
		t.Fatalf("ensureLabManifest returned error: %v", err)
	}

	configMap, err := client.CoreV1().ConfigMaps(namespace).Get(context.Background(), "lab-helper", metav1.GetOptions{})
	if err != nil {
		t.Fatalf("ConfigMap was not created: %v", err)
	}
	if configMap.Data["namespace"] != namespace || configMap.Labels[SessionIDLabel] != params.AttemptID ||
		len(configMap.Finalizers) != 0 {
		t.Fatalf("ConfigMap substitutions or ownership labels are missing: %#v", configMap)
	}
	if _, err = client.CoreV1().ServiceAccounts(namespace).Get(context.Background(), "lab-helper", metav1.GetOptions{}); err != nil {
		t.Fatalf("ServiceAccount was not created: %v", err)
	}
	if _, err = client.RbacV1().Roles(namespace).Get(context.Background(), "lab-helper", metav1.GetOptions{}); err != nil {
		t.Fatalf("Role was not created: %v", err)
	}
	binding, err := client.RbacV1().RoleBindings(namespace).Get(context.Background(), "lab-helper", metav1.GetOptions{})
	if err != nil {
		t.Fatalf("RoleBinding was not created: %v", err)
	}
	if len(binding.Subjects) != 1 || binding.Subjects[0].Namespace != namespace {
		t.Fatalf("RoleBinding subject namespace was not constrained: %#v", binding.Subjects)
	}
	deployment, err := client.AppsV1().Deployments(namespace).Get(context.Background(), "lab-helper", metav1.GetOptions{})
	if err != nil {
		t.Fatalf("Deployment was not created: %v", err)
	}
	if deployment.Spec.Template.Spec.Containers[0].Image != "example.test/lab-helper:1.0.0" ||
		deployment.Labels[SessionLabResourceLabel] != "true" {
		t.Fatalf("Deployment must come from the manifest and be tracked for readiness: %#v", deployment)
	}
	if _, err = client.CoreV1().Services(namespace).Get(context.Background(), "lab-helper", metav1.GetOptions{}); err != nil {
		t.Fatalf("Service was not created: %v", err)
	}
	policy, err := client.NetworkingV1().NetworkPolicies(namespace).Get(context.Background(), "lab-helper", metav1.GetOptions{})
	if err != nil {
		t.Fatalf("NetworkPolicy was not created: %v", err)
	}
	proxyNamespace := policy.Spec.Ingress[0].From[0].NamespaceSelector.MatchLabels["kubernetes.io/metadata.name"]
	if proxyNamespace != params.WorkspaceProxyNamespace {
		t.Fatalf("proxy namespace = %q, want %q", proxyNamespace, params.WorkspaceProxyNamespace)
	}
	topology, err := dynamicClient.Resource(topologyGVR).Namespace(namespace).Get(
		context.Background(),
		namespace,
		metav1.GetOptions{},
	)
	if err != nil {
		t.Fatalf("Topology was not created: %v", err)
	}
	if topology.GetNamespace() != namespace || topology.GetLabels()[SessionOwnerLabel] != params.OwnerID {
		t.Fatalf("Topology namespace or ownership labels are missing: %#v", topology.Object)
	}

	if err = admin.ensureLabManifest(context.Background(), namespace, params); err != nil {
		t.Fatalf("ensureLabManifest is not idempotent: %v", err)
	}
}

func TestEnsureLabManifestValidatesAllDocumentsBeforeApply(t *testing.T) {
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
	params := EnsureSessionParams{
		AttemptID:   "session",
		OwnerID:     "owner",
		LabManifest: manifest,
	}
	if err := admin.ensureLabManifest(context.Background(), "lab-test", params); err == nil {
		t.Fatal("unsupported manifest was accepted")
	}
	if _, err := client.CoreV1().ConfigMaps("lab-test").Get(context.Background(), "must-not-exist", metav1.GetOptions{}); err == nil {
		t.Fatal("ConfigMap was applied before the complete manifest was validated")
	}
}

func TestSessionWaitsForLabDeployments(t *testing.T) {
	const (
		namespace = "lab-session"
		sessionID = "session"
	)
	replicas := int32(1)
	client := kubernetesfake.NewSimpleClientset(
		&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
			Labels: map[string]string{
				SessionManagedByLabel: SessionManagedByValue,
				SessionIDLabel:        sessionID,
				SessionOwnerLabel:     "owner",
			},
			Annotations: map[string]string{
				SessionAttemptAnnotation:  sessionID,
				SessionTopologyAnnotation: "false",
			},
		}},
		&appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{Name: workspaceName, Namespace: namespace},
			Spec:       appsv1.DeploymentSpec{Replicas: &replicas},
			Status:     appsv1.DeploymentStatus{Replicas: 1, ReadyReplicas: 1},
		},
		&appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "lab-helper",
				Namespace: namespace,
				Labels:    map[string]string{SessionLabResourceLabel: "true"},
			},
			Spec:   appsv1.DeploymentSpec{Replicas: &replicas},
			Status: appsv1.DeploymentStatus{Replicas: 1},
		},
	)
	logger := zerolog.Nop()
	admin := NewKubernetesAdminWithClients(client, nil, nil, &logger)

	record, err := admin.GetSession(context.Background(), sessionID, "/clabgate/workspace")
	if err != nil {
		t.Fatalf("GetSession returned error: %v", err)
	}
	if record.Phase == SessionPhaseReady || record.LabResourcesReady {
		t.Fatalf("session became ready before its lab Deployment: %+v", record)
	}

	deployment, err := client.AppsV1().Deployments(namespace).Get(
		context.Background(),
		"lab-helper",
		metav1.GetOptions{},
	)
	if err != nil {
		t.Fatal(err)
	}
	deployment.Status.ReadyReplicas = 1
	if _, err = client.AppsV1().Deployments(namespace).UpdateStatus(
		context.Background(),
		deployment,
		metav1.UpdateOptions{},
	); err != nil {
		t.Fatal(err)
	}
	record, err = admin.GetSession(context.Background(), sessionID, "/clabgate/workspace")
	if err != nil {
		t.Fatalf("GetSession after readiness returned error: %v", err)
	}
	if record.Phase != SessionPhaseReady || !record.LabResourcesReady {
		t.Fatalf("session did not become ready with its lab Deployment: %+v", record)
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
		"app=lab",
		"branch=main",
		"repo=https%3A%2F%2Fgit.example.test%2Fgroup%2Fcms-labs-simple-task.git",
		"targetpath=task",
	} {
		if !strings.Contains(got, expectedPart) {
			t.Fatalf("workspace URL %q does not contain %q", got, expectedPart)
		}
	}
	// app must stay "lab": nbgitpuller otherwise defaults to its "notebook"
	// app and redirects to /tree/task, serving the classic Notebook browser
	// instead of JupyterLab. A "notebook" value would pin that regression.
	if strings.Contains(got, "app=notebook") {
		t.Fatalf("workspace URL asks nbgitpuller for the classic Notebook UI: %q", got)
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
