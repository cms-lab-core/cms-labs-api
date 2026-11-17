package queries

import (
	"context"
	"fmt"

	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
)

const CheckerTargetNamespaceLabel = "labs.cmslabs.ru/target-namespace"

func (k *KubernetesAdminQuery) checkerAccountName() string {
	if k.checkerServiceAccount != "" {
		return k.checkerServiceAccount
	}
	return "cms-labs-checker"
}

func (k *KubernetesAdminQuery) checkerJobNamespace(labNamespace string) string {
	if k.checkerNamespace != "" {
		return k.checkerNamespace
	}
	return labNamespace
}

// Shared-namespace Jobs are selected by their target lab. Include pre-upgrade
// Jobs in the lab namespace so an in-flight check/result survives deployment.
// The platform only consumes the pre-provisioned identity; it never manages RBAC.
func (k *KubernetesAdminQuery) checkerJobsForLab(ctx context.Context, labNamespace string) (*batchv1.JobList, error) {
	jobNamespace := k.checkerJobNamespace(labNamespace)
	selector := labels.Set{SessionComponentLabel: SessionComponentChecker}
	if jobNamespace != labNamespace {
		selector[CheckerTargetNamespaceLabel] = labNamespace
	}
	jobs, err := k.clientset.BatchV1().Jobs(jobNamespace).List(ctx, metav1.ListOptions{LabelSelector: selector.AsSelector().String()})
	if err != nil {
		return nil, fmt.Errorf("list checker jobs: %w", err)
	}
	if jobNamespace != labNamespace {
		legacy, err := k.clientset.BatchV1().Jobs(labNamespace).List(ctx, metav1.ListOptions{
			LabelSelector: labels.Set{SessionComponentLabel: SessionComponentChecker}.AsSelector().String(),
		})
		if err != nil {
			return nil, fmt.Errorf("list legacy checker jobs: %w", err)
		}
		jobs.Items = append(jobs.Items, legacy.Items...)
	}
	return jobs, nil
}
