package queries

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sort"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
)

const checkerOutboxComponent = "checker-outbox"
const checkerQueuedAnnotation = "labs.cmslabs.ru/checker-result-queued"

var ErrCheckerRunning = errors.New("checker is still running")

func (k *KubernetesAdminQuery) checkerOutboxNamespace() string {
	if k.checkerNamespace != "" {
		return k.checkerNamespace
	}
	return "cms-labs"
}

func checkerOutboxName(result CheckerResult) string {
	sum := sha256.Sum256([]byte(result.Namespace + "/" + result.JobName + "/" + result.CheckID))
	return "checker-result-" + hex.EncodeToString(sum[:20])
}

func (k *KubernetesAdminQuery) readCheckerOutbox(ctx context.Context, result CheckerResult) (CheckerResult, error) {
	cm, err := k.clientset.CoreV1().ConfigMaps(k.checkerOutboxNamespace()).Get(ctx, checkerOutboxName(result), metav1.GetOptions{})
	if err != nil {
		return CheckerResult{}, err
	}
	return decodeCheckerOutbox(cm)
}

func decodeCheckerOutbox(cm *corev1.ConfigMap) (CheckerResult, error) {
	result := CheckerResult{OutboxName: cm.Name}
	reader, err := gzip.NewReader(bytes.NewReader(cm.BinaryData["result.gz"]))
	if err != nil {
		return result, fmt.Errorf("open queued checker result %s: %w", cm.Name, err)
	}
	defer reader.Close()
	data, err := io.ReadAll(io.LimitReader(reader, 4*maxCheckerResultBytes))
	if err != nil {
		return result, fmt.Errorf("read queued checker result %s: %w", cm.Name, err)
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return result, fmt.Errorf("decode queued checker result %s: %w", cm.Name, err)
	}
	result.OutboxName = cm.Name
	if result.AttemptID == "" || result.CheckID == "" || !json.Valid([]byte(result.Payload)) {
		return result, fmt.Errorf("invalid queued checker result %s", cm.Name)
	}
	return result, nil
}

func (k *KubernetesAdminQuery) QueueCheckerResult(ctx context.Context, result CheckerResult) error {
	if result.AttemptID == "" || result.CheckID == "" || !json.Valid([]byte(result.Payload)) {
		return errors.New("cannot queue invalid checker result")
	}
	data, err := json.Marshal(result)
	if err != nil {
		return err
	}
	var buf bytes.Buffer
	writer := gzip.NewWriter(&buf)
	if _, err = writer.Write(data); err != nil {
		return err
	}
	if err = writer.Close(); err != nil {
		return err
	}
	if buf.Len() > maxCheckerResultBytes {
		return errors.New("compressed checker result exceeds ConfigMap size limit")
	}
	cm := &corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: checkerOutboxName(result), Namespace: k.checkerOutboxNamespace(), Labels: map[string]string{SessionManagedByLabel: SessionManagedByValue, SessionComponentLabel: checkerOutboxComponent, SessionIDLabel: result.AttemptID}}, Immutable: boolPtr(true), BinaryData: map[string][]byte{"result.gz": buf.Bytes()}}
	_, err = k.clientset.CoreV1().ConfigMaps(cm.Namespace).Create(ctx, cm, metav1.CreateOptions{})
	if apierrors.IsAlreadyExists(err) {
		stored, getErr := k.readCheckerOutbox(ctx, result)
		if getErr != nil {
			return getErr
		}
		if stored.AttemptID != result.AttemptID || stored.CheckID != result.CheckID {
			return errors.New("queued checker identity mismatch")
		}
		err = nil
	}
	if err != nil {
		return err
	}
	if notifyErr := k.notifyCheckerOutbox(ctx, result); notifyErr != nil {
		k.Logger.Error().Err(notifyErr).Str("job", result.JobName).Msg("checker result is durable; notify will retry")
	}
	return nil
}

func (k *KubernetesAdminQuery) notifyCheckerOutbox(ctx context.Context, result CheckerResult) error {
	job, err := k.clientset.BatchV1().Jobs(result.Namespace).Get(ctx, result.JobName, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		return nil
	}
	if err != nil {
		return err
	}
	name := checkerOutboxName(result)
	if job.Annotations[checkerQueuedAnnotation] == name {
		return nil
	}
	copy := job.DeepCopy()
	if copy.Annotations == nil {
		copy.Annotations = map[string]string{}
	}
	copy.Annotations[checkerQueuedAnnotation] = name
	_, err = k.clientset.BatchV1().Jobs(result.Namespace).Update(ctx, copy, metav1.UpdateOptions{})
	return err
}

func (k *KubernetesAdminQuery) QueuedCheckerResults(ctx context.Context) ([]CheckerResult, error) {
	cms, err := k.clientset.CoreV1().ConfigMaps(k.checkerOutboxNamespace()).List(ctx, metav1.ListOptions{LabelSelector: labels.Set{SessionManagedByLabel: SessionManagedByValue, SessionComponentLabel: checkerOutboxComponent}.AsSelector().String()})
	if err != nil {
		return nil, err
	}
	results := make([]CheckerResult, 0, len(cms.Items))
	for i := range cms.Items {
		result, decodeErr := decodeCheckerOutbox(&cms.Items[i])
		result.Error = decodeErr
		results = append(results, result)
	}
	sort.SliceStable(results, func(i, j int) bool {
		if results[i].CreatedAt.Equal(results[j].CreatedAt) {
			return results[i].CheckID < results[j].CheckID
		}
		return results[i].CreatedAt.Before(results[j].CreatedAt)
	})
	return results, nil
}

func (k *KubernetesAdminQuery) CaptureCheckerResults(ctx context.Context, namespace string) (int, error) {
	results, err := k.PendingCheckerResults(ctx, namespace)
	if err != nil {
		return 0, err
	}
	count := 0
	var failures []error
	for _, result := range results {
		if result.Error != nil {
			failures = append(failures, result.Error)
			continue
		}
		if err := k.QueueCheckerResult(ctx, result); err != nil {
			failures = append(failures, fmt.Errorf("queue checker job %s: %w", result.JobName, err))
			continue
		}
		count++
	}
	return count, errors.Join(failures...)
}

func (k *KubernetesAdminQuery) AcknowledgeCheckerResult(ctx context.Context, result CheckerResult) error {
	if err := k.MarkCheckerResultSynced(ctx, result.Namespace, result.JobName); err != nil && !apierrors.IsNotFound(err) {
		return err
	}
	err := k.clientset.CoreV1().ConfigMaps(k.checkerOutboxNamespace()).Delete(ctx, result.OutboxName, metav1.DeleteOptions{})
	if apierrors.IsNotFound(err) {
		return nil
	}
	return err
}
