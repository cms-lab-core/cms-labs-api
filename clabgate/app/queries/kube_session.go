package queries

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/url"
	"path"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/validation"
	k8syaml "k8s.io/apimachinery/pkg/util/yaml"
)

const (
	SessionManagedByLabel = "app.kubernetes.io/managed-by"
	SessionManagedByValue = "clabgate"
	SessionIDLabel        = "labs.cmslabs.ru/session-id"
	SessionOwnerLabel     = "labs.cmslabs.ru/owner-id"
	SessionComponentLabel = "labs.cmslabs.ru/component"

	SessionAttemptAnnotation      = "labs.cmslabs.ru/attempt-id"
	SessionUsernameAnnotation     = "labs.cmslabs.ru/username"
	SessionLabPathAnnotation      = "labs.cmslabs.ru/lab-path"
	SessionTestPathAnnotation     = "labs.cmslabs.ru/test-path"
	SessionTitleAnnotation        = "labs.cmslabs.ru/title"
	SessionTopologyAnnotation     = "labs.cmslabs.ru/topology-required"
	SessionRuntimeAnnotation      = "labs.cmslabs.ru/runtime"
	SessionTaskRepoAnnotation     = "labs.cmslabs.ru/task-repository-url"
	SessionTaskRefAnnotation      = "labs.cmslabs.ru/task-repository-ref"
	SessionTaskRevisionAnnotation = "labs.cmslabs.ru/task-revision"
	SessionPhaseAnnotation        = "labs.cmslabs.ru/session-phase"
	SessionReadyAtAnnotation      = "labs.cmslabs.ru/ready-at"
	SessionProvisionedAnnotation  = "labs.cmslabs.ru/desired-resources-accepted"
	CheckerSyncedAnnotation       = "labs.cmslabs.ru/checker-result-synced"
	CheckerIDAnnotation           = "labs.cmslabs.ru/check-id"

	SessionComponentWorkspace = "workspace"
	SessionComponentChecker   = "checker"

	workspaceName = "jupyter"
)

var topologyGVR = schema.GroupVersionResource{
	Group: "c9s.run", Version: "v1alpha1", Resource: "topologies",
}

type SessionPhase string

const (
	SessionPhasePending      SessionPhase = "pending"
	SessionPhaseProvisioning SessionPhase = "provisioning"
	SessionPhaseReady        SessionPhase = "ready"
	SessionPhaseDegraded     SessionPhase = "degraded"
	SessionPhaseFailed       SessionPhase = "failed"
	SessionPhaseStopping     SessionPhase = "stopping"
	SessionPhaseActive       SessionPhase = "active"
)

type SessionRecord struct {
	ID                string       `json:"id"`
	AttemptID         string       `json:"attempt_id"`
	Namespace         string       `json:"namespace"`
	OwnerID           string       `json:"owner_id"`
	Username          string       `json:"username"`
	Title             string       `json:"title,omitempty"`
	LabPath           string       `json:"lab_path,omitempty"`
	TestPath          string       `json:"test_path,omitempty"`
	TaskRevision      string       `json:"task_revision,omitempty"`
	Phase             SessionPhase `json:"phase"`
	Message           string       `json:"message,omitempty"`
	WorkspaceURL      string       `json:"workspace_url,omitempty"`
	TopologyReady     bool         `json:"topology_ready"`
	WorkspaceReady    bool         `json:"workspace_ready"`
	DesiredAccepted   bool         `json:"desired_resources_accepted"`
	CheckerRunning    bool         `json:"checker_running"`
	CheckerSuccessful *bool        `json:"checker_successful,omitempty"`
	CreatedAt         time.Time    `json:"created_at"`
}

type EnsureSessionParams struct {
	AttemptID       string
	OwnerID         string
	Username        string
	Title           string
	LabPath         string
	TestPath        string
	TopologyYAML    string
	JupyterImage    string
	StorageSize     string
	WorkspacePrefix string
	TaskRepository  string
	TaskRef         string
	TaskRevision    string
}

type RunCheckerParams struct {
	SessionID      string
	Namespace      string
	AttemptID      string
	OwnerID        string
	LabPath        string
	TestPath       string
	Image          string
	TimeoutSeconds int64
}

type CheckerResult struct {
	JobName   string
	Namespace string
	AttemptID string
	CheckID   string
	Payload   string
	Logs      string
}

func SessionNamespace(sessionID string) string {
	normalized := strings.ToLower(sessionID)
	var b strings.Builder
	for _, r := range normalized {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			b.WriteRune(r)
		}
	}
	normalized = strings.Trim(b.String(), "-")
	if normalized == "" || len(validation.IsDNS1123Label("lab-"+normalized)) != 0 {
		sum := sha256.Sum256([]byte(sessionID))
		normalized = hex.EncodeToString(sum[:16])
	}
	if len(normalized) > 59 {
		normalized = strings.TrimRight(normalized[:59], "-")
	}
	return "lab-" + normalized
}

func (k *KubernetesAdminQuery) EnsureSession(ctx context.Context, params EnsureSessionParams) (SessionRecord, error) {
	if params.AttemptID == "" || params.OwnerID == "" || params.Username == "" {
		return SessionRecord{}, fmt.Errorf("attempt ID, owner ID and username are required")
	}
	if params.JupyterImage == "" {
		return SessionRecord{}, fmt.Errorf("JUPYTER_IMAGE is not configured")
	}

	namespaceName := SessionNamespace(params.AttemptID)
	namespace, err := k.ensureNamespace(ctx, namespaceName, params)
	if err != nil {
		return SessionRecord{}, err
	}

	topologyRequired := strings.TrimSpace(params.TopologyYAML) != ""
	if topologyRequired {
		if err := k.ensureTopology(ctx, namespaceName, params.AttemptID, params.OwnerID, params.TopologyYAML); err != nil {
			return SessionRecord{}, err
		}
	}
	if err := k.ensureWorkspace(ctx, namespaceName, params); err != nil {
		return SessionRecord{}, err
	}

	// The annotation is written after all desired resources have been accepted. This lets status
	// distinguish an empty namespace from a session being reconciled by Kubernetes.
	if namespace.Annotations[SessionTopologyAnnotation] != fmt.Sprintf("%t", topologyRequired) ||
		namespace.Annotations[SessionProvisionedAnnotation] != "true" {
		copy := namespace.DeepCopy()
		copy.Annotations[SessionTopologyAnnotation] = fmt.Sprintf("%t", topologyRequired)
		copy.Annotations[SessionProvisionedAnnotation] = "true"
		if _, err := k.clientset.CoreV1().Namespaces().Update(ctx, copy, metav1.UpdateOptions{}); err != nil {
			return SessionRecord{}, fmt.Errorf("mark session topology requirement: %w", err)
		}
	}

	return k.GetSession(ctx, params.AttemptID, params.WorkspacePrefix)
}

func (k *KubernetesAdminQuery) ensureNamespace(
	ctx context.Context,
	name string,
	params EnsureSessionParams,
) (*corev1.Namespace, error) {
	namespace, err := k.clientset.CoreV1().Namespaces().Get(ctx, name, metav1.GetOptions{})
	if err == nil {
		if namespace.Labels[SessionIDLabel] != params.AttemptID || namespace.Labels[SessionOwnerLabel] != params.OwnerID {
			return nil, fmt.Errorf("session namespace %q belongs to another session or owner", name)
		}
		copy := namespace.DeepCopy()
		if copy.Annotations == nil {
			copy.Annotations = map[string]string{}
		}
		desiredAnnotations := map[string]string{
			SessionAttemptAnnotation:      params.AttemptID,
			SessionUsernameAnnotation:     params.Username,
			SessionLabPathAnnotation:      params.LabPath,
			SessionTestPathAnnotation:     params.TestPath,
			SessionTitleAnnotation:        params.Title,
			SessionRuntimeAnnotation:      "jupyter",
			SessionTaskRepoAnnotation:     params.TaskRepository,
			SessionTaskRefAnnotation:      params.TaskRef,
			SessionTaskRevisionAnnotation: params.TaskRevision,
		}
		changed := false
		for key, value := range desiredAnnotations {
			if copy.Annotations[key] != value {
				copy.Annotations[key] = value
				changed = true
			}
		}
		if changed {
			updated, updateErr := k.clientset.CoreV1().Namespaces().Update(ctx, copy, metav1.UpdateOptions{})
			if updateErr != nil {
				return nil, fmt.Errorf("update session namespace metadata: %w", updateErr)
			}
			return updated, nil
		}
		return namespace, nil
	}
	if !apierrors.IsNotFound(err) {
		return nil, fmt.Errorf("get session namespace: %w", err)
	}

	namespace = &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Labels: map[string]string{
				SessionManagedByLabel:                SessionManagedByValue,
				SessionIDLabel:                       params.AttemptID,
				SessionOwnerLabel:                    params.OwnerID,
				"pod-security.kubernetes.io/enforce": "privileged",
				"pod-security.kubernetes.io/warn":    "privileged",
				"pod-security.kubernetes.io/audit":   "privileged",
			},
			Annotations: map[string]string{
				SessionAttemptAnnotation:      params.AttemptID,
				SessionUsernameAnnotation:     params.Username,
				SessionLabPathAnnotation:      params.LabPath,
				SessionTestPathAnnotation:     params.TestPath,
				SessionTitleAnnotation:        params.Title,
				SessionRuntimeAnnotation:      "jupyter",
				SessionTopologyAnnotation:     "unknown",
				SessionTaskRepoAnnotation:     params.TaskRepository,
				SessionTaskRefAnnotation:      params.TaskRef,
				SessionTaskRevisionAnnotation: params.TaskRevision,
				SessionPhaseAnnotation:        string(SessionPhaseProvisioning),
			},
		},
	}
	created, err := k.clientset.CoreV1().Namespaces().Create(ctx, namespace, metav1.CreateOptions{})
	if err != nil {
		return nil, fmt.Errorf("create session namespace: %w", err)
	}
	return created, nil
}

func (k *KubernetesAdminQuery) ensureTopology(
	ctx context.Context,
	namespace, sessionID, ownerID, manifest string,
) error {
	manifest = strings.ReplaceAll(manifest, "$NAME", namespace)
	decoder := k8syaml.NewYAMLOrJSONDecoder(strings.NewReader(manifest), 4096)
	objects := make([]*unstructured.Unstructured, 0)
	topologyCount := 0
	for {
		raw := map[string]any{}
		if err := decoder.Decode(&raw); errors.Is(err, io.EOF) {
			break
		} else if err != nil {
			return fmt.Errorf("decode topology manifest: %w", err)
		}
		if len(raw) == 0 {
			continue
		}
		obj := &unstructured.Unstructured{Object: raw}
		switch {
		case obj.GetAPIVersion() == "v1" && obj.GetKind() == "ConfigMap":
		case obj.GetAPIVersion() == "c9s.run/v1alpha1" && obj.GetKind() == "Topology":
			topologyCount++
		default:
			return fmt.Errorf("unsupported object %s %s in topology manifest; only v1 ConfigMap and clabernetes Topology are allowed", obj.GetAPIVersion(), obj.GetKind())
		}
		objects = append(objects, obj)
	}
	if topologyCount != 1 {
		return fmt.Errorf("topology manifest must contain exactly one clabernetes Topology, got %d", topologyCount)
	}
	for _, obj := range objects {
		if obj.GetKind() == "ConfigMap" {
			if err := k.ensureTopologyConfigMap(ctx, namespace, sessionID, ownerID, obj); err != nil {
				return err
			}
			continue
		}
		if err := k.ensureTopologyObject(ctx, namespace, sessionID, ownerID, obj); err != nil {
			return err
		}
	}
	return nil
}

func (k *KubernetesAdminQuery) ensureTopologyObject(
	ctx context.Context,
	namespace, sessionID, ownerID string,
	obj *unstructured.Unstructured,
) error {
	obj.SetNamespace(namespace)
	if obj.GetName() == "" {
		obj.SetName(namespace)
	}
	objectLabels := obj.GetLabels()
	if objectLabels == nil {
		objectLabels = map[string]string{}
	}
	objectLabels[SessionManagedByLabel] = SessionManagedByValue
	objectLabels[SessionIDLabel] = sessionID
	objectLabels[SessionOwnerLabel] = ownerID
	obj.SetLabels(objectLabels)

	resourceClient := k.dynamicClient.Resource(topologyGVR).Namespace(namespace)
	existing, err := resourceClient.Get(ctx, obj.GetName(), metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		if _, createErr := resourceClient.Create(ctx, obj, metav1.CreateOptions{}); createErr != nil {
			return fmt.Errorf("create topology: %w", createErr)
		}
		return nil
	}
	if err != nil {
		return fmt.Errorf("get topology: %w", err)
	}
	obj.SetResourceVersion(existing.GetResourceVersion())
	if _, err := resourceClient.Update(ctx, obj, metav1.UpdateOptions{}); err != nil {
		return fmt.Errorf("update topology: %w", err)
	}
	return nil
}

func (k *KubernetesAdminQuery) ensureTopologyConfigMap(
	ctx context.Context,
	namespace, sessionID, ownerID string,
	obj *unstructured.Unstructured,
) error {
	configMap := &corev1.ConfigMap{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(obj.Object, configMap); err != nil {
		return fmt.Errorf("decode topology ConfigMap: %w", err)
	}
	if configMap.Name == "" {
		return errors.New("topology ConfigMap name is required")
	}
	configMap.Namespace = namespace
	if configMap.Labels == nil {
		configMap.Labels = map[string]string{}
	}
	configMap.Labels[SessionManagedByLabel] = SessionManagedByValue
	configMap.Labels[SessionIDLabel] = sessionID
	configMap.Labels[SessionOwnerLabel] = ownerID

	existing, err := k.clientset.CoreV1().ConfigMaps(namespace).Get(ctx, configMap.Name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		if _, createErr := k.clientset.CoreV1().ConfigMaps(namespace).Create(ctx, configMap, metav1.CreateOptions{}); createErr != nil {
			return fmt.Errorf("create topology ConfigMap %s: %w", configMap.Name, createErr)
		}
		return nil
	}
	if err != nil {
		return fmt.Errorf("get topology ConfigMap %s: %w", configMap.Name, err)
	}
	configMap.ResourceVersion = existing.ResourceVersion
	if _, err := k.clientset.CoreV1().ConfigMaps(namespace).Update(ctx, configMap, metav1.UpdateOptions{}); err != nil {
		return fmt.Errorf("update topology ConfigMap %s: %w", configMap.Name, err)
	}
	return nil
}

func (k *KubernetesAdminQuery) ensureWorkspace(
	ctx context.Context,
	namespace string,
	params EnsureSessionParams,
) error {
	quantity, err := resource.ParseQuantity(params.StorageSize)
	if err != nil {
		return fmt.Errorf("invalid Jupyter storage size %q: %w", params.StorageSize, err)
	}
	labelsMap := map[string]string{
		SessionManagedByLabel: SessionManagedByValue,
		SessionIDLabel:        params.AttemptID,
		SessionOwnerLabel:     params.OwnerID,
		SessionComponentLabel: SessionComponentWorkspace,
	}

	pvc := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{Name: workspaceName, Namespace: namespace, Labels: labelsMap},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
			Resources:   corev1.VolumeResourceRequirements{Requests: corev1.ResourceList{corev1.ResourceStorage: quantity}},
		},
	}
	if _, err := k.clientset.CoreV1().PersistentVolumeClaims(namespace).Get(ctx, workspaceName, metav1.GetOptions{}); apierrors.IsNotFound(err) {
		if _, err = k.clientset.CoreV1().PersistentVolumeClaims(namespace).Create(ctx, pvc, metav1.CreateOptions{}); err != nil {
			return fmt.Errorf("create Jupyter PVC: %w", err)
		}
	} else if err != nil {
		return fmt.Errorf("get Jupyter PVC: %w", err)
	}

	one := int32(1)
	baseURL := strings.TrimRight(params.WorkspacePrefix, "/") + "/" + params.AttemptID
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{Name: workspaceName, Namespace: namespace, Labels: labelsMap},
		Spec: appsv1.DeploymentSpec{
			Replicas: &one,
			Strategy: appsv1.DeploymentStrategy{Type: appsv1.RecreateDeploymentStrategyType},
			Selector: &metav1.LabelSelector{MatchLabels: labelsMap},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Labels: labelsMap},
				Spec: corev1.PodSpec{
					AutomountServiceAccountToken: boolPtr(false),
					EnableServiceLinks:           boolPtr(false),
					SecurityContext:              &corev1.PodSecurityContext{RunAsUser: int64Ptr(1000), RunAsGroup: int64Ptr(100), FSGroup: int64Ptr(100)},
					Containers: []corev1.Container{{
						Name:  workspaceName,
						Image: params.JupyterImage,
						Args: []string{
							"start-notebook.py",
							"--ServerApp.base_url=" + baseURL,
							"--ServerApp.allow_remote_access=True",
							"--IdentityProvider.token=",
						},
						Env: []corev1.EnvVar{
							{Name: "ATTEMPT_ID", Value: params.AttemptID},
							{Name: "JUPYTER_ENABLE_LAB", Value: "yes"},
						},
						Ports:          []corev1.ContainerPort{{Name: "http", ContainerPort: 8888}},
						ReadinessProbe: &corev1.Probe{ProbeHandler: corev1.ProbeHandler{TCPSocket: &corev1.TCPSocketAction{Port: intstr.FromInt32(8888)}}, InitialDelaySeconds: 2, PeriodSeconds: 3},
						Resources: corev1.ResourceRequirements{
							Requests: corev1.ResourceList{corev1.ResourceCPU: resource.MustParse("500m"), corev1.ResourceMemory: resource.MustParse("1Gi")},
							Limits:   corev1.ResourceList{corev1.ResourceCPU: resource.MustParse("2"), corev1.ResourceMemory: resource.MustParse("4Gi")},
						},
						VolumeMounts: []corev1.VolumeMount{{Name: "workspace", MountPath: "/home/jovyan"}},
					}},
					Volumes: []corev1.Volume{{Name: "workspace", VolumeSource: corev1.VolumeSource{PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{ClaimName: workspaceName}}}},
				},
			},
		},
	}
	deployment.Spec.Template.Spec.Containers[0].Lifecycle = &corev1.Lifecycle{PostStart: &corev1.LifecycleHandler{
		Exec: &corev1.ExecAction{Command: []string{"sh", "-c", "if [ -x /usr/local/bin/notebook.entrypoint.sh ]; then /usr/local/bin/notebook.entrypoint.sh; fi"}},
	}}
	existingDeployment, err := k.clientset.AppsV1().Deployments(namespace).Get(ctx, workspaceName, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		if _, err = k.clientset.AppsV1().Deployments(namespace).Create(ctx, deployment, metav1.CreateOptions{}); err != nil {
			return fmt.Errorf("create Jupyter deployment: %w", err)
		}
	} else if err != nil {
		return fmt.Errorf("get Jupyter deployment: %w", err)
	} else {
		deployment.ResourceVersion = existingDeployment.ResourceVersion
		if _, err = k.clientset.AppsV1().Deployments(namespace).Update(ctx, deployment, metav1.UpdateOptions{}); err != nil {
			return fmt.Errorf("update Jupyter deployment: %w", err)
		}
	}

	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{Name: workspaceName, Namespace: namespace, Labels: labelsMap},
		Spec:       corev1.ServiceSpec{Selector: labelsMap, Ports: []corev1.ServicePort{{Name: "http", Port: 8888, TargetPort: intstr.FromInt32(8888)}}},
	}
	existingService, err := k.clientset.CoreV1().Services(namespace).Get(ctx, workspaceName, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		if _, err = k.clientset.CoreV1().Services(namespace).Create(ctx, service, metav1.CreateOptions{}); err != nil {
			return fmt.Errorf("create Jupyter service: %w", err)
		}
	} else if err != nil {
		return fmt.Errorf("get Jupyter service: %w", err)
	} else {
		service.ResourceVersion = existingService.ResourceVersion
		service.Spec.ClusterIP = existingService.Spec.ClusterIP
		service.Spec.ClusterIPs = existingService.Spec.ClusterIPs
		service.Spec.IPFamilies = existingService.Spec.IPFamilies
		service.Spec.IPFamilyPolicy = existingService.Spec.IPFamilyPolicy
		if _, err = k.clientset.CoreV1().Services(namespace).Update(ctx, service, metav1.UpdateOptions{}); err != nil {
			return fmt.Errorf("update Jupyter service: %w", err)
		}
	}
	return nil
}

func (k *KubernetesAdminQuery) GetSession(ctx context.Context, sessionID, workspacePrefix string) (SessionRecord, error) {
	selector := labels.Set{SessionManagedByLabel: SessionManagedByValue, SessionIDLabel: sessionID}.AsSelector().String()
	namespaces, err := k.clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{LabelSelector: selector})
	if err != nil {
		return SessionRecord{}, fmt.Errorf("list session namespaces: %w", err)
	}
	if len(namespaces.Items) == 0 {
		return SessionRecord{}, apierrors.NewNotFound(schema.GroupResource{Group: "labs.cmslabs.ru", Resource: "sessions"}, sessionID)
	}
	if len(namespaces.Items) > 1 {
		return SessionRecord{}, fmt.Errorf("more than one namespace found for session %q", sessionID)
	}
	return k.sessionFromNamespace(ctx, &namespaces.Items[0], workspacePrefix)
}

func (k *KubernetesAdminQuery) ListSessions(ctx context.Context, ownerID, workspacePrefix string) ([]SessionRecord, error) {
	selectorSet := labels.Set{SessionManagedByLabel: SessionManagedByValue}
	if ownerID != "" {
		selectorSet[SessionOwnerLabel] = ownerID
	}
	namespaces, err := k.clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{LabelSelector: selectorSet.AsSelector().String()})
	if err != nil {
		return nil, fmt.Errorf("list session namespaces: %w", err)
	}
	result := make([]SessionRecord, 0, len(namespaces.Items))
	for index := range namespaces.Items {
		record, recordErr := k.sessionFromNamespace(ctx, &namespaces.Items[index], workspacePrefix)
		if recordErr != nil {
			return nil, recordErr
		}
		result = append(result, record)
	}
	return result, nil
}

func (k *KubernetesAdminQuery) sessionFromNamespace(
	ctx context.Context,
	namespace *corev1.Namespace,
	workspacePrefix string,
) (SessionRecord, error) {
	record := SessionRecord{
		ID:              namespace.Labels[SessionIDLabel],
		AttemptID:       namespace.Annotations[SessionAttemptAnnotation],
		Namespace:       namespace.Name,
		OwnerID:         namespace.Labels[SessionOwnerLabel],
		Username:        namespace.Annotations[SessionUsernameAnnotation],
		Title:           namespace.Annotations[SessionTitleAnnotation],
		LabPath:         namespace.Annotations[SessionLabPathAnnotation],
		TestPath:        namespace.Annotations[SessionTestPathAnnotation],
		TaskRevision:    namespace.Annotations[SessionTaskRevisionAnnotation],
		Phase:           SessionPhaseProvisioning,
		CreatedAt:       namespace.CreationTimestamp.Time,
		DesiredAccepted: namespace.Annotations[SessionProvisionedAnnotation] == "true",
	}
	if namespace.DeletionTimestamp != nil {
		record.Phase = SessionPhaseStopping
		return record, nil
	}

	workspace, err := k.clientset.AppsV1().Deployments(namespace.Name).Get(ctx, workspaceName, metav1.GetOptions{})
	if err == nil {
		record.WorkspaceReady = workspace.Status.ReadyReplicas > 0 && workspace.Status.ReadyReplicas == workspace.Status.Replicas
	} else if !apierrors.IsNotFound(err) {
		return SessionRecord{}, fmt.Errorf("get workspace deployment: %w", err)
	}

	topologyRequired := namespace.Annotations[SessionTopologyAnnotation] == "true"
	record.TopologyReady = !topologyRequired
	if topologyRequired {
		topologies, listErr := k.dynamicClient.Resource(topologyGVR).Namespace(namespace.Name).List(ctx, metav1.ListOptions{})
		if listErr != nil {
			return SessionRecord{}, fmt.Errorf("list session topology: %w", listErr)
		}
		if len(topologies.Items) > 0 {
			ready, _, _ := unstructured.NestedBool(topologies.Items[0].Object, "status", "topologyReady")
			state, _, _ := unstructured.NestedString(topologies.Items[0].Object, "status", "topologyState")
			record.TopologyReady = ready
			switch state {
			case "deployfailed":
				record.Phase = SessionPhaseFailed
				record.Message = "topology deployment failed"
			case "degraded":
				record.Phase = SessionPhaseDegraded
				record.Message = "topology is degraded"
			}
		}
	}

	jobs, err := k.clientset.BatchV1().Jobs(namespace.Name).List(ctx, metav1.ListOptions{LabelSelector: labels.Set{SessionComponentLabel: SessionComponentChecker}.AsSelector().String()})
	if err == nil && len(jobs.Items) > 0 {
		sort.Slice(jobs.Items, func(i, j int) bool {
			return jobs.Items[i].CreationTimestamp.Before(&jobs.Items[j].CreationTimestamp)
		})
		job := jobs.Items[len(jobs.Items)-1]
		record.CheckerRunning = job.Status.Active > 0
		if job.Status.Succeeded > 0 {
			value := true
			record.CheckerSuccessful = &value
		} else if job.Status.Failed > 0 {
			value := false
			record.CheckerSuccessful = &value
		}
	}

	if record.Phase != SessionPhaseFailed && record.Phase != SessionPhaseDegraded {
		if record.WorkspaceReady && record.TopologyReady {
			record.Phase = SessionPhaseReady
		} else if namespace.Annotations[SessionTopologyAnnotation] == "unknown" {
			record.Phase = SessionPhasePending
		}
	}
	if record.WorkspaceReady {
		taskRef := namespace.Annotations[SessionTaskRevisionAnnotation]
		if taskRef == "" {
			taskRef = namespace.Annotations[SessionTaskRefAnnotation]
		}
		record.WorkspaceURL = buildWorkspaceURL(
			workspacePrefix,
			record.ID,
			namespace.Annotations[SessionTaskRepoAnnotation],
			taskRef,
			record.LabPath,
		)
	}
	return record, nil
}

func (k *KubernetesAdminQuery) RunChecker(ctx context.Context, params RunCheckerParams) (string, error) {
	if params.Image == "" {
		return "", errors.New("CHECKER_IMAGE is not configured")
	}
	activeJobs, err := k.clientset.BatchV1().Jobs(params.Namespace).List(ctx, metav1.ListOptions{
		LabelSelector: labels.Set{SessionComponentLabel: SessionComponentChecker}.AsSelector().String(),
	})
	if err != nil {
		return "", fmt.Errorf("list checker jobs: %w", err)
	}
	for _, job := range activeJobs.Items {
		// A newly accepted Job can have all counters at zero until its controller
		// creates a Pod. Treat it as in progress to make repeated clicks idempotent.
		if job.Status.Succeeded == 0 && job.Status.Failed == 0 {
			return job.Name, nil
		}
	}

	backoffLimit := int32(0)
	ttlSeconds := int32(86400)
	checkID := uuid.NewString()
	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "checker-",
			Namespace:    params.Namespace,
			Labels: map[string]string{
				SessionManagedByLabel: SessionManagedByValue,
				SessionIDLabel:        params.SessionID,
				SessionOwnerLabel:     params.OwnerID,
				SessionComponentLabel: SessionComponentChecker,
			},
			Annotations: map[string]string{CheckerSyncedAnnotation: "false", CheckerIDAnnotation: checkID},
		},
		Spec: batchv1.JobSpec{
			BackoffLimit:            &backoffLimit,
			TTLSecondsAfterFinished: &ttlSeconds,
			ActiveDeadlineSeconds:   &params.TimeoutSeconds,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{
					SessionManagedByLabel: SessionManagedByValue,
					SessionIDLabel:        params.SessionID,
					SessionOwnerLabel:     params.OwnerID,
					SessionComponentLabel: SessionComponentChecker,
				}},
				Spec: corev1.PodSpec{
					AutomountServiceAccountToken: boolPtr(false),
					RestartPolicy:                corev1.RestartPolicyNever,
					Containers: []corev1.Container{{
						Name:  "checker",
						Image: params.Image,
						Env: []corev1.EnvVar{
							{Name: "SESSION_ID", Value: params.SessionID},
							{Name: "ATTEMPT_ID", Value: params.AttemptID},
							{Name: "LAB_PATH", Value: params.LabPath},
							{Name: "TEST_PATH", Value: params.TestPath},
							{Name: "SESSION_NAMESPACE", Value: params.Namespace},
						},
						TerminationMessagePath:   "/dev/termination-log",
						TerminationMessagePolicy: corev1.TerminationMessageReadFile,
						Resources: corev1.ResourceRequirements{
							Requests: corev1.ResourceList{corev1.ResourceCPU: resource.MustParse("100m"), corev1.ResourceMemory: resource.MustParse("128Mi")},
							Limits:   corev1.ResourceList{corev1.ResourceCPU: resource.MustParse("1"), corev1.ResourceMemory: resource.MustParse("1Gi")},
						},
					}},
				},
			},
		},
	}
	created, err := k.clientset.BatchV1().Jobs(params.Namespace).Create(ctx, job, metav1.CreateOptions{})
	if err != nil {
		return "", fmt.Errorf("create checker job: %w", err)
	}
	return created.Name, nil
}

func (k *KubernetesAdminQuery) PendingCheckerResults(ctx context.Context, namespace string) ([]CheckerResult, error) {
	jobs, err := k.clientset.BatchV1().Jobs(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: labels.Set{SessionComponentLabel: SessionComponentChecker}.AsSelector().String(),
	})
	if err != nil {
		return nil, fmt.Errorf("list checker jobs: %w", err)
	}
	results := make([]CheckerResult, 0)
	for _, job := range jobs.Items {
		if job.Status.Succeeded == 0 || job.Annotations[CheckerSyncedAnnotation] == "true" {
			continue
		}
		pods, podErr := k.clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
			LabelSelector: labels.Set{"batch.kubernetes.io/job-name": job.Name}.AsSelector().String(),
		})
		if podErr != nil {
			return nil, fmt.Errorf("list checker pods for %s: %w", job.Name, podErr)
		}
		for _, pod := range pods.Items {
			logs := ""
			if rawLogs, logsErr := k.clientset.CoreV1().Pods(namespace).GetLogs(pod.Name, &corev1.PodLogOptions{Container: "checker"}).DoRaw(ctx); logsErr == nil {
				const maxCheckerLogBytes = 64 * 1024
				if len(rawLogs) > maxCheckerLogBytes {
					rawLogs = rawLogs[len(rawLogs)-maxCheckerLogBytes:]
				}
				logs = string(rawLogs)
			}
			for _, status := range pod.Status.ContainerStatuses {
				if status.Name == "checker" && status.State.Terminated != nil && strings.TrimSpace(status.State.Terminated.Message) != "" {
					results = append(results, CheckerResult{
						JobName: job.Name, Namespace: namespace,
						AttemptID: job.Labels[SessionIDLabel], CheckID: job.Annotations[CheckerIDAnnotation],
						Payload: status.State.Terminated.Message, Logs: logs,
					})
				}
			}
		}
	}
	return results, nil
}

func (k *KubernetesAdminQuery) MarkCheckerResultSynced(ctx context.Context, namespace, jobName string) error {
	job, err := k.clientset.BatchV1().Jobs(namespace).Get(ctx, jobName, metav1.GetOptions{})
	if err != nil {
		return err
	}
	copy := job.DeepCopy()
	if copy.Annotations == nil {
		copy.Annotations = map[string]string{}
	}
	copy.Annotations[CheckerSyncedAnnotation] = "true"
	_, err = k.clientset.BatchV1().Jobs(namespace).Update(ctx, copy, metav1.UpdateOptions{})
	return err
}

func (k *KubernetesAdminQuery) DeleteSession(ctx context.Context, sessionID, ownerID string, allowAnyOwner bool) error {
	record, err := k.GetSession(ctx, sessionID, "")
	if err != nil {
		return err
	}
	if !allowAnyOwner && record.OwnerID != ownerID {
		return errors.New("session belongs to another user")
	}
	return k.clientset.CoreV1().Namespaces().Delete(ctx, record.Namespace, metav1.DeleteOptions{})
}

func (k *KubernetesAdminQuery) MarkSessionPhase(ctx context.Context, namespaceName string, phase SessionPhase) error {
	namespace, err := k.clientset.CoreV1().Namespaces().Get(ctx, namespaceName, metav1.GetOptions{})
	if err != nil {
		return err
	}
	copy := namespace.DeepCopy()
	if copy.Annotations == nil {
		copy.Annotations = map[string]string{}
	}
	if copy.Annotations[SessionPhaseAnnotation] == string(phase) {
		return nil
	}
	copy.Annotations[SessionPhaseAnnotation] = string(phase)
	if phase == SessionPhaseReady || phase == SessionPhaseActive {
		if copy.Annotations[SessionReadyAtAnnotation] == "" {
			copy.Annotations[SessionReadyAtAnnotation] = time.Now().UTC().Format(time.RFC3339)
		}
	}
	_, err = k.clientset.CoreV1().Namespaces().Update(ctx, copy, metav1.UpdateOptions{})
	return err
}

func buildWorkspaceURL(prefix, sessionID, repository, ref, labPath string) string {
	base := strings.TrimRight(prefix, "/") + "/" + sessionID
	cleanLabPath := strings.Trim(path.Clean("/"+labPath), "/")
	if repository == "" || ref == "" || cleanLabPath == "" {
		return base + "/lab"
	}
	query := url.Values{
		"repo":    []string{repository},
		"branch":  []string{ref},
		"urlpath": []string{"lab/tree/" + cleanLabPath},
	}
	return base + "/git-pull?" + query.Encode()
}

func int64Ptr(value int64) *int64 { return &value }
func boolPtr(value bool) *bool    { return &value }
