package queries

import (
	"context"
	"errors"
	"fmt"
	"io"
	"slices"
	"strings"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation"
	k8syaml "k8s.io/apimachinery/pkg/util/yaml"
)

type labObjectType string

const (
	labConfigMap      labObjectType = "configmap"
	labServiceAccount labObjectType = "serviceaccount"
	labService        labObjectType = "service"
	labDeployment     labObjectType = "deployment"
	labRole           labObjectType = "role"
	labRoleBinding    labObjectType = "rolebinding"
	labNetworkPolicy  labObjectType = "networkpolicy"
	labTopology       labObjectType = "topology"
)

type labObject struct {
	kind   labObjectType
	object *unstructured.Unstructured
}

// ensureLabManifest is intentionally unaware of cms-labs-terminal. It applies a constrained set of
// namespace-scoped Kubernetes resources supplied by a trusted lab repository. Every object is
// validated before the first write, forced into the attempt namespace and marked for ownership and
// readiness tracking.
func (k *KubernetesAdminQuery) ensureLabManifest(
	ctx context.Context,
	namespace string,
	params EnsureSessionParams,
) error {
	manifest := strings.ReplaceAll(params.LabManifest, "$NAME", namespace)
	if strings.Contains(manifest, "$WORKSPACE_PROXY_NAMESPACE") {
		if strings.TrimSpace(params.WorkspaceProxyNamespace) == "" {
			return errors.New("WORKSPACE_PROXY_NAMESPACE or POD_NAMESPACE is required by the lab manifest")
		}
		manifest = strings.ReplaceAll(manifest, "$WORKSPACE_PROXY_NAMESPACE", params.WorkspaceProxyNamespace)
	}
	objects, err := decodeLabManifest(manifest, namespace, params.AttemptID, params.OwnerID)
	if err != nil {
		return err
	}
	for _, object := range objects {
		if err = k.applyLabObject(ctx, namespace, object); err != nil {
			return err
		}
	}
	return nil
}

func decodeLabManifest(manifest, namespace, sessionID, ownerID string) ([]labObject, error) {
	decoder := k8syaml.NewYAMLOrJSONDecoder(strings.NewReader(manifest), 4096)
	objects := make([]labObject, 0)
	topologyCount := 0
	for {
		raw := map[string]any{}
		if err := decoder.Decode(&raw); errors.Is(err, io.EOF) {
			break
		} else if err != nil {
			return nil, fmt.Errorf("decode lab manifest: %w", err)
		}
		if len(raw) == 0 {
			continue
		}
		object := &unstructured.Unstructured{Object: raw}
		kind, ok := allowedLabObject(object.GetAPIVersion(), object.GetKind())
		if !ok {
			return nil, fmt.Errorf("unsupported lab object %s %s", object.GetAPIVersion(), object.GetKind())
		}
		if kind == labTopology {
			topologyCount++
			if object.GetName() == "" {
				object.SetName(namespace)
			}
		}
		if object.GetName() == "" || len(validation.IsDNS1123Subdomain(object.GetName())) > 0 {
			return nil, fmt.Errorf("lab object %s %s has an invalid or empty metadata.name", object.GetKind(), object.GetName())
		}
		// Repository metadata must not retain an object or impersonate API-server state. The attempt
		// namespace is the lifecycle boundary, so task resources do not need their own finalizers or
		// owner references.
		object.SetNamespace(namespace)
		object.SetGenerateName("")
		object.SetResourceVersion("")
		object.SetUID("")
		object.SetGeneration(0)
		object.SetCreationTimestamp(metav1.Time{})
		object.SetDeletionTimestamp(nil)
		object.SetDeletionGracePeriodSeconds(nil)
		object.SetManagedFields(nil)
		object.SetOwnerReferences(nil)
		object.SetFinalizers(nil)
		labelsMap := object.GetLabels()
		if labelsMap == nil {
			labelsMap = map[string]string{}
		}
		labelsMap[SessionManagedByLabel] = SessionManagedByValue
		labelsMap[SessionIDLabel] = sessionID
		labelsMap[SessionOwnerLabel] = ownerID
		labelsMap[SessionLabResourceLabel] = "true"
		object.SetLabels(labelsMap)
		if err := validateLabObject(kind, object, namespace); err != nil {
			return nil, fmt.Errorf("validate %s %s: %w", object.GetKind(), object.GetName(), err)
		}
		objects = append(objects, labObject{kind: kind, object: object})
	}
	if topologyCount != 1 {
		return nil, fmt.Errorf("lab manifest must contain exactly one clabernetes Topology, got %d", topologyCount)
	}
	return objects, nil
}

func allowedLabObject(apiVersion, kind string) (labObjectType, bool) {
	switch apiVersion + "/" + kind {
	case "v1/ConfigMap":
		return labConfigMap, true
	case "v1/ServiceAccount":
		return labServiceAccount, true
	case "v1/Service":
		return labService, true
	case "apps/v1/Deployment":
		return labDeployment, true
	case "rbac.authorization.k8s.io/v1/Role":
		return labRole, true
	case "rbac.authorization.k8s.io/v1/RoleBinding":
		return labRoleBinding, true
	case "networking.k8s.io/v1/NetworkPolicy":
		return labNetworkPolicy, true
	case "c9s.run/v1alpha1/Topology":
		return labTopology, true
	default:
		return "", false
	}
}

func validateLabObject(kind labObjectType, object *unstructured.Unstructured, namespace string) error {
	switch kind {
	case labDeployment:
		deployment := &appsv1.Deployment{}
		if err := runtime.DefaultUnstructuredConverter.FromUnstructured(object.Object, deployment); err != nil {
			return err
		}
		pod := deployment.Spec.Template.Spec
		if pod.HostNetwork || pod.HostPID || pod.HostIPC || pod.NodeName != "" {
			return errors.New("host namespaces and a fixed nodeName are forbidden")
		}
		for _, volume := range pod.Volumes {
			if volume.HostPath != nil {
				return fmt.Errorf("hostPath volume %q is forbidden", volume.Name)
			}
		}
		containers := append(append([]corev1.Container{}, pod.InitContainers...), pod.Containers...)
		if len(pod.Containers) == 0 {
			return errors.New("at least one container is required")
		}
		for _, container := range containers {
			if container.SecurityContext != nil {
				if container.SecurityContext.Privileged != nil && *container.SecurityContext.Privileged {
					return fmt.Errorf("privileged container %q is forbidden", container.Name)
				}
				if container.SecurityContext.AllowPrivilegeEscalation != nil && *container.SecurityContext.AllowPrivilegeEscalation {
					return fmt.Errorf("privilege escalation in container %q is forbidden", container.Name)
				}
				if container.SecurityContext.Capabilities != nil && len(container.SecurityContext.Capabilities.Add) > 0 {
					return fmt.Errorf("added capabilities in container %q are forbidden", container.Name)
				}
			}
			for _, port := range container.Ports {
				if port.HostPort != 0 {
					return fmt.Errorf("hostPort in container %q is forbidden", container.Name)
				}
			}
		}
	case labService:
		service := &corev1.Service{}
		if err := runtime.DefaultUnstructuredConverter.FromUnstructured(object.Object, service); err != nil {
			return err
		}
		if service.Spec.Type != "" && service.Spec.Type != corev1.ServiceTypeClusterIP {
			return fmt.Errorf("only ClusterIP Services are allowed, got %s", service.Spec.Type)
		}
		if len(service.Spec.ExternalIPs) > 0 || service.Spec.ExternalName != "" {
			return errors.New("external Service addresses are forbidden")
		}
	case labRoleBinding:
		binding := &rbacv1.RoleBinding{}
		if err := runtime.DefaultUnstructuredConverter.FromUnstructured(object.Object, binding); err != nil {
			return err
		}
		if binding.RoleRef.Kind != "Role" || binding.RoleRef.APIGroup != rbacv1.GroupName {
			return errors.New("RoleBinding may reference only a Role in the attempt namespace")
		}
		for index := range binding.Subjects {
			subject := &binding.Subjects[index]
			if subject.Kind != "ServiceAccount" || subject.APIGroup != "" {
				return errors.New("RoleBinding subjects must be ServiceAccounts")
			}
			if subject.Namespace == "" {
				subject.Namespace = namespace
			}
			if subject.Namespace != namespace {
				return errors.New("RoleBinding subject is outside the attempt namespace")
			}
		}
		converted, err := runtime.DefaultUnstructuredConverter.ToUnstructured(binding)
		if err != nil {
			return err
		}
		object.Object = converted
	case labRole:
		role := &rbacv1.Role{}
		if err := runtime.DefaultUnstructuredConverter.FromUnstructured(object.Object, role); err != nil {
			return err
		}
		for _, rule := range role.Rules {
			if len(rule.NonResourceURLs) > 0 {
				return errors.New("non-resource RBAC URLs are forbidden")
			}
			if slices.Contains(rule.APIGroups, "*") ||
				slices.Contains(rule.Resources, "*") ||
				slices.Contains(rule.Verbs, "*") {
				return errors.New("wildcard RBAC rules are forbidden")
			}
		}
	case labConfigMap, labServiceAccount, labNetworkPolicy, labTopology:
		return nil
	default:
		return fmt.Errorf("unsupported object type %q", kind)
	}
	return nil
}

func (k *KubernetesAdminQuery) applyLabObject(ctx context.Context, namespace string, object labObject) error {
	switch object.kind {
	case labConfigMap:
		desired := &corev1.ConfigMap{}
		if err := convertLabObject(object.object, desired); err != nil {
			return err
		}
		existing, err := k.clientset.CoreV1().ConfigMaps(namespace).Get(ctx, desired.Name, metav1.GetOptions{})
		if apierrors.IsNotFound(err) {
			_, err = k.clientset.CoreV1().ConfigMaps(namespace).Create(ctx, desired, metav1.CreateOptions{})
		} else if err == nil {
			desired.ResourceVersion = existing.ResourceVersion
			_, err = k.clientset.CoreV1().ConfigMaps(namespace).Update(ctx, desired, metav1.UpdateOptions{})
		}
		return labApplyError("ConfigMap", desired.Name, err)
	case labServiceAccount:
		desired := &corev1.ServiceAccount{}
		if err := convertLabObject(object.object, desired); err != nil {
			return err
		}
		existing, err := k.clientset.CoreV1().ServiceAccounts(namespace).Get(ctx, desired.Name, metav1.GetOptions{})
		if apierrors.IsNotFound(err) {
			_, err = k.clientset.CoreV1().ServiceAccounts(namespace).Create(ctx, desired, metav1.CreateOptions{})
		} else if err == nil {
			desired.ResourceVersion = existing.ResourceVersion
			desired.Secrets = existing.Secrets
			_, err = k.clientset.CoreV1().ServiceAccounts(namespace).Update(ctx, desired, metav1.UpdateOptions{})
		}
		return labApplyError("ServiceAccount", desired.Name, err)
	case labService:
		desired := &corev1.Service{}
		if err := convertLabObject(object.object, desired); err != nil {
			return err
		}
		existing, err := k.clientset.CoreV1().Services(namespace).Get(ctx, desired.Name, metav1.GetOptions{})
		if apierrors.IsNotFound(err) {
			_, err = k.clientset.CoreV1().Services(namespace).Create(ctx, desired, metav1.CreateOptions{})
		} else if err == nil {
			desired.ResourceVersion = existing.ResourceVersion
			desired.Spec.ClusterIP = existing.Spec.ClusterIP
			desired.Spec.ClusterIPs = existing.Spec.ClusterIPs
			desired.Spec.IPFamilies = existing.Spec.IPFamilies
			desired.Spec.IPFamilyPolicy = existing.Spec.IPFamilyPolicy
			_, err = k.clientset.CoreV1().Services(namespace).Update(ctx, desired, metav1.UpdateOptions{})
		}
		return labApplyError("Service", desired.Name, err)
	case labDeployment:
		desired := &appsv1.Deployment{}
		if err := convertLabObject(object.object, desired); err != nil {
			return err
		}
		existing, err := k.clientset.AppsV1().Deployments(namespace).Get(ctx, desired.Name, metav1.GetOptions{})
		if apierrors.IsNotFound(err) {
			_, err = k.clientset.AppsV1().Deployments(namespace).Create(ctx, desired, metav1.CreateOptions{})
		} else if err == nil {
			desired.ResourceVersion = existing.ResourceVersion
			_, err = k.clientset.AppsV1().Deployments(namespace).Update(ctx, desired, metav1.UpdateOptions{})
		}
		return labApplyError("Deployment", desired.Name, err)
	case labRole:
		desired := &rbacv1.Role{}
		if err := convertLabObject(object.object, desired); err != nil {
			return err
		}
		existing, err := k.clientset.RbacV1().Roles(namespace).Get(ctx, desired.Name, metav1.GetOptions{})
		if apierrors.IsNotFound(err) {
			_, err = k.clientset.RbacV1().Roles(namespace).Create(ctx, desired, metav1.CreateOptions{})
		} else if err == nil {
			desired.ResourceVersion = existing.ResourceVersion
			_, err = k.clientset.RbacV1().Roles(namespace).Update(ctx, desired, metav1.UpdateOptions{})
		}
		return labApplyError("Role", desired.Name, err)
	case labRoleBinding:
		desired := &rbacv1.RoleBinding{}
		if err := convertLabObject(object.object, desired); err != nil {
			return err
		}
		existing, err := k.clientset.RbacV1().RoleBindings(namespace).Get(ctx, desired.Name, metav1.GetOptions{})
		if apierrors.IsNotFound(err) {
			_, err = k.clientset.RbacV1().RoleBindings(namespace).Create(ctx, desired, metav1.CreateOptions{})
		} else if err == nil {
			desired.ResourceVersion = existing.ResourceVersion
			_, err = k.clientset.RbacV1().RoleBindings(namespace).Update(ctx, desired, metav1.UpdateOptions{})
		}
		return labApplyError("RoleBinding", desired.Name, err)
	case labNetworkPolicy:
		desired := &networkingv1.NetworkPolicy{}
		if err := convertLabObject(object.object, desired); err != nil {
			return err
		}
		existing, err := k.clientset.NetworkingV1().NetworkPolicies(namespace).Get(ctx, desired.Name, metav1.GetOptions{})
		if apierrors.IsNotFound(err) {
			_, err = k.clientset.NetworkingV1().NetworkPolicies(namespace).Create(ctx, desired, metav1.CreateOptions{})
		} else if err == nil {
			desired.ResourceVersion = existing.ResourceVersion
			_, err = k.clientset.NetworkingV1().NetworkPolicies(namespace).Update(ctx, desired, metav1.UpdateOptions{})
		}
		return labApplyError("NetworkPolicy", desired.Name, err)
	case labTopology:
		resourceClient := k.dynamicClient.Resource(topologyGVR).Namespace(namespace)
		existing, err := resourceClient.Get(ctx, object.object.GetName(), metav1.GetOptions{})
		if apierrors.IsNotFound(err) {
			_, err = resourceClient.Create(ctx, object.object, metav1.CreateOptions{})
		} else if err == nil {
			object.object.SetResourceVersion(existing.GetResourceVersion())
			_, err = resourceClient.Update(ctx, object.object, metav1.UpdateOptions{})
		}
		return labApplyError("Topology", object.object.GetName(), err)
	default:
		return fmt.Errorf("unsupported lab object type %q", object.kind)
	}
}

func convertLabObject(object *unstructured.Unstructured, target any) error {
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(object.Object, target); err != nil {
		return fmt.Errorf("decode %s %s: %w", object.GetKind(), object.GetName(), err)
	}
	return nil
}

func labApplyError(kind, name string, err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("apply lab %s %s: %w", kind, name, err)
}
