package queries

import (
	"context"

	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/tools/cache"
)

func (k *KubernetesAdminQuery) WatchCheckerResults(ctx context.Context, wake chan<- struct{}) {
	selector := labels.Set{SessionManagedByLabel: SessionManagedByValue, SessionComponentLabel: SessionComponentChecker}.AsSelector().String()
	informer := cache.NewSharedIndexInformer(&cache.ListWatch{
		ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
			options.LabelSelector = selector
			return k.clientset.BatchV1().Jobs("").List(ctx, options)
		},
		WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
			options.LabelSelector = selector
			return k.clientset.BatchV1().Jobs("").Watch(ctx, options)
		},
	}, &batchv1.Job{}, 0, cache.Indexers{})
	notify := func(obj interface{}) {
		if tombstone, ok := obj.(cache.DeletedFinalStateUnknown); ok {
			obj = tombstone.Obj
		}
		job, ok := obj.(*batchv1.Job)
		if !ok || job.Status.Succeeded == 0 && job.Status.Failed == 0 || job.Annotations[CheckerSyncedAnnotation] == "true" {
			return
		}
		select {
		case wake <- struct{}{}:
		default:
		}
	}
	_, err := informer.AddEventHandler(cache.ResourceEventHandlerFuncs{AddFunc: notify, DeleteFunc: notify, UpdateFunc: func(_, current interface{}) { notify(current) }})
	if err != nil {
		k.Logger.Error().Err(err).Msg("watch checker results")
		return
	}
	informer.Run(ctx.Done())
}
