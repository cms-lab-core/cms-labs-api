package queries

import (
	"bytes"
	"context"
	"fmt"

	"github.com/cms-lab-core/cms-labs-api/clabgate/pkg/kubeconfig"
	"github.com/cms-lab-core/cms-labs-api/shared/logs"
	"github.com/rs/zerolog"
	"gopkg.in/yaml.v3"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
)

type DeploymentInfo struct {
	Name          string `json:"name"`
	PodNameLast   string `json:"-"`
	Status        string `json:"status"`
	Restarts      int32  `json:"restarts"`
	ReadyReplicas int32  `json:"ready_replicas"`
	Replicas      *int32 `json:"replicas"`
}

type ServiceInfo struct {
	Name       string   `json:"name"`
	Type       string   `json:"type"`
	ExternalIP []string `json:"external_ip,omitempty"`
	ClusterIP  string   `json:"cluster_ip"`
}

type TTYDInfo struct {
	Name string `json:"name"`
	Url  string `json:"url"`
}

type KubernetesAdminQuery struct {
	clientset     kubernetes.Interface
	dynamicClient dynamic.Interface
	config        *rest.Config
	*zerolog.Logger
}

func NewKubernetesAdminWithClients(
	clientset kubernetes.Interface,
	dynamicClient dynamic.Interface,
	config *rest.Config,
	logger *zerolog.Logger,
) *KubernetesAdminQuery {
	return &KubernetesAdminQuery{
		clientset:     clientset,
		dynamicClient: dynamicClient,
		config:        config,
		Logger:        logger,
	}
}

// NewKubernetesAdmin creates a client from a pod ServiceAccount or local kubeconfig.
func NewKubernetesAdmin(zeroLogConf *logs.ZeroLoggerConf) (*KubernetesAdminQuery, error) {
	config, err := kubeconfig.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load Kubernetes config: %v", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes clientset: %v", err)
	}
	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes dynamicClient: %v", err)
	}

	return &KubernetesAdminQuery{
		clientset:     clientset,
		dynamicClient: dynamicClient,
		config:        config,
		Logger:        logs.NewZeroLogger(zeroLogConf),
	}, nil
}

// GetTopologyYAML возвращает YAML-представление конкретной топологии (CRD)
func (k *KubernetesAdminQuery) GetTopologyYAML(ctx context.Context, namespace string) ([]byte, error) {
	if namespace == "" {
		return []byte{}, fmt.Errorf("namespace is required")
	}

	gvr := schema.GroupVersionResource{
		Group:    "c9s.run",
		Version:  "v1alpha1",
		Resource: "topologies",
	}

	list, err := k.dynamicClient.Resource(gvr).Namespace(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return []byte{}, fmt.Errorf("failed to list Topologies in namespace %s: %v", namespace, err)
	}

	if len(list.Items) == 0 {
		return []byte{}, fmt.Errorf("no Topology found in namespace %s", namespace)
	}

	for _, unstructuredObj := range list.Items {
		yamlData, err := yaml.Marshal(unstructuredObj.Object)
		if err == nil {
			return yamlData, nil
		}
	}
	return []byte{}, fmt.Errorf("failed to marshal Topology to YAML")
}

// GetDeploymentsInfo возвращает информацию о deployments в namespace
func (k *KubernetesAdminQuery) GetDeploymentsInfo(ctx context.Context, namespace string) ([]DeploymentInfo, error) {
	deployments, err := k.clientset.AppsV1().Deployments(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list deployments in namespace %s: %v", namespace, err)
	}

	var result []DeploymentInfo
	for _, deploy := range deployments.Items {
		status := "Ready"
		if deploy.Status.ReadyReplicas != deploy.Status.Replicas {
			status = "NotReady"
		}

		var restarts int32 = 0
		var podNameLast string = ""
		pods, err := k.clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
			LabelSelector: metav1.FormatLabelSelector(deploy.Spec.Selector),
		})
		if err == nil {
			for _, pod := range pods.Items {
				for _, cs := range pod.Status.ContainerStatuses {
					restarts += cs.RestartCount
					podNameLast = pod.Name
				}
			}
		}

		result = append(result, DeploymentInfo{
			Name:          deploy.Name,
			PodNameLast:   podNameLast,
			Status:        status,
			Restarts:      restarts,
			ReadyReplicas: deploy.Status.ReadyReplicas,
			Replicas:      deploy.Spec.Replicas,
		})
	}
	return result, nil
}

// GetServicesInfo возвращает информацию о services типа LoadBalancer в namespace
func (k *KubernetesAdminQuery) GetServicesInfo(ctx context.Context, namespace string) ([]ServiceInfo, error) {
	services, err := k.clientset.CoreV1().Services(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list services in namespace %s: %v", namespace, err)
	}

	var result []ServiceInfo
	for _, svc := range services.Items {
		if svc.Spec.Type != "LoadBalancer" {
			continue
		}

		externalIP := make([]string, 0)
		if svc.Status.LoadBalancer.Ingress != nil && len(svc.Status.LoadBalancer.Ingress) > 0 {
			for _, ingress := range svc.Status.LoadBalancer.Ingress {
				if ingress.IP != "" {
					externalIP = append(externalIP, ingress.IP)
				}
			}
		}

		result = append(result, ServiceInfo{
			Name:       svc.Name,
			Type:       string(svc.Spec.Type),
			ExternalIP: externalIP,
			ClusterIP:  svc.Spec.ClusterIP,
		})
	}
	return result, nil
}

// GetTTYDInfoInNamespace discovers any namespace-local terminal Service that exposes the stable
// ttyd port contract. The Service may be supplied by the task manifest or another lab runtime.
// The caller is responsible for assigning an externally reachable, authorized URL.
func (k *KubernetesAdminQuery) GetTTYDInfoInNamespace(ctx context.Context, namespace string) ([]TTYDInfo, error) {
	services, err := k.clientset.CoreV1().Services(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list services in namespace %s: %v", namespace, err)
	}

	var result []TTYDInfo
	for _, svc := range services.Items {
		if !hasTTYDPort(svc.Spec.Ports) {
			continue
		}

		result = append(result, TTYDInfo{
			Name: svc.Name,
		})
	}

	return result, nil
}

// hasTTYDPort проверяет наличие порта с именем "ttyd" и номером 7681
func hasTTYDPort(ports []corev1.ServicePort) bool {
	for _, port := range ports {
		if port.Name == "ttyd" && port.Port == 7681 {
			return true
		}
	}
	return false
}

// RestartPod выполняет exec kill 1 в контейнере пода
func (k *KubernetesAdminQuery) RestartPod(ctx context.Context, namespace, podName string, containerName string) error {
	if namespace == "" || podName == "" || containerName == "" {
		return fmt.Errorf("namespace, podName and containerName are required")
	}

	var stdout, stderr bytes.Buffer

	req := k.clientset.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(podName).
		Namespace(namespace).
		SubResource("exec").
		VersionedParams(&corev1.PodExecOptions{
			Command:   []string{"/bin/sh", "-c", "kill 1"},
			Container: containerName,
			Stdin:     false,
			Stdout:    true,
			Stderr:    true,
			TTY:       false,
		}, scheme.ParameterCodec)

	executor, err := remotecommand.NewSPDYExecutor(k.config, "POST", req.URL())
	if err != nil {
		return fmt.Errorf("failed to create executor: %w", err)
	}

	_ = executor.StreamWithContext(ctx, remotecommand.StreamOptions{
		Stdin:  nil,
		Stdout: &stdout,
		Stderr: &stderr,
	})
	return nil
}

// DeletePod удаляет под
func (k *KubernetesAdminQuery) DeletePod(ctx context.Context, namespace, podName string) error {
	if namespace == "" || podName == "" {
		return fmt.Errorf("namespace and podName are required")
	}
	return k.clientset.CoreV1().Pods(namespace).Delete(ctx, podName, metav1.DeleteOptions{})
}
