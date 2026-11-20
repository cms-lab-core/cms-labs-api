package queries

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/rs/zerolog"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
	typedcore "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
	ktesting "k8s.io/client-go/testing"
)

type checkerLogClient struct {
	kubernetes.Interface
	logs typedcore.CoreV1Interface
}

func (c checkerLogClient) CoreV1() typedcore.CoreV1Interface {
	return checkerLogCore{CoreV1Interface: c.Interface.CoreV1(), logs: c.logs}
}

type checkerLogCore struct {
	typedcore.CoreV1Interface
	logs typedcore.CoreV1Interface
}

func (c checkerLogCore) Pods(ns string) typedcore.PodInterface {
	return checkerLogPods{PodInterface: c.CoreV1Interface.Pods(ns), logs: c.logs.Pods(ns)}
}

type checkerLogPods struct {
	typedcore.PodInterface
	logs typedcore.PodInterface
}

func (c checkerLogPods) GetLogs(name string, options *corev1.PodLogOptions) *rest.Request {
	return c.logs.GetLogs(name, options)
}

func checkerOutboxFixture(t *testing.T) (*KubernetesAdminQuery, *fake.Clientset) {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasSuffix(r.URL.Path, "/log") {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(makeResultLog(`{"max_score":2,"current_score":1,"result_display":"1/2"}`)))
	}))
	t.Cleanup(server.Close)
	real, err := kubernetes.NewForConfig(&rest.Config{Host: server.URL})
	if err != nil {
		t.Fatal(err)
	}
	client := fake.NewSimpleClientset(
		&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "lab-attempt", Labels: map[string]string{SessionManagedByLabel: SessionManagedByValue, SessionIDLabel: "attempt", SessionOwnerLabel: "42"}}},
		&batchv1.Job{ObjectMeta: metav1.ObjectMeta{Name: "checker-one", Namespace: "cms-labs", Labels: map[string]string{SessionManagedByLabel: SessionManagedByValue, SessionComponentLabel: SessionComponentChecker, SessionIDLabel: "attempt", SessionOwnerLabel: "42", CheckerTargetNamespaceLabel: "lab-attempt"}, Annotations: map[string]string{CheckerIDAnnotation: "check-one"}, CreationTimestamp: metav1.NewTime(time.Now().UTC())}, Status: batchv1.JobStatus{Succeeded: 1}},
		&corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "checker-pod", Namespace: "cms-labs", Labels: map[string]string{"batch.kubernetes.io/job-name": "checker-one"}}, Status: corev1.PodStatus{ContainerStatuses: []corev1.ContainerStatus{{Name: "checker", State: corev1.ContainerState{Terminated: &corev1.ContainerStateTerminated{ExitCode: 0}}}}}},
	)
	logger := zerolog.Nop()
	admin := NewKubernetesAdminWithClients(checkerLogClient{Interface: client, logs: real.CoreV1()}, nil, nil, &logger)
	admin.checkerNamespace = "cms-labs"
	return admin, client
}

func TestCheckerOutboxSurvivesLabAndJobDeletion(t *testing.T) {
	admin, client := checkerOutboxFixture(t)
	ctx := context.Background()
	if err := admin.DeleteSession(ctx, "attempt", "42", false); err != nil {
		t.Fatal(err)
	}
	if _, err := client.CoreV1().Namespaces().Get(ctx, "lab-attempt", metav1.GetOptions{}); !apierrors.IsNotFound(err) {
		t.Fatalf("lab remains: %v", err)
	}
	if _, err := client.BatchV1().Jobs("cms-labs").Get(ctx, "checker-one", metav1.GetOptions{}); !apierrors.IsNotFound(err) {
		t.Fatalf("checker remains: %v", err)
	}
	logger := zerolog.Nop()
	restarted := NewKubernetesAdminWithClients(client, nil, nil, &logger)
	restarted.checkerNamespace = "cms-labs"
	results, err := restarted.QueuedCheckerResults(ctx)
	if err != nil || len(results) != 1 || results[0].Error != nil || results[0].CheckID != "check-one" || !strings.Contains(results[0].Payload, `"current_score":1`) {
		t.Fatalf("lost queued result: %+v %v", results, err)
	}
	cm, err := client.CoreV1().ConfigMaps("cms-labs").Get(ctx, results[0].OutboxName, metav1.GetOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if len(cm.OwnerReferences) != 0 {
		t.Fatal("outbox must not be garbage collected with a lab or Job")
	}
	if err := restarted.AcknowledgeCheckerResult(ctx, results[0]); err != nil {
		t.Fatal(err)
	}
	results, err = restarted.QueuedCheckerResults(ctx)
	if err != nil || len(results) != 0 {
		t.Fatal("acknowledged result remained queued", err)
	}
}

func TestCheckerCleanupFailsClosed(t *testing.T) {
	for _, kind := range []string{"running", "storage unavailable"} {
		t.Run(kind, func(t *testing.T) {
			admin, client := checkerOutboxFixture(t)
			ctx := context.Background()
			if kind == "running" {
				job, err := client.BatchV1().Jobs("cms-labs").Get(ctx, "checker-one", metav1.GetOptions{})
				if err != nil {
					t.Fatal(err)
				}
				job.Status.Succeeded = 0
				_, err = client.BatchV1().Jobs("cms-labs").Update(ctx, job, metav1.UpdateOptions{})
				if err != nil {
					t.Fatal(err)
				}
			} else {
				client.PrependReactor("create", "configmaps", func(ktesting.Action) (bool, runtime.Object, error) {
					return true, nil, errors.New("outbox unavailable")
				})
			}
			err := admin.DeleteSession(ctx, "attempt", "42", false)
			if err == nil {
				t.Fatal("lab deleted before result was durable")
			}
			if kind == "running" && !errors.Is(err, ErrCheckerRunning) {
				t.Fatal(err)
			}
			if _, err := client.CoreV1().Namespaces().Get(ctx, "lab-attempt", metav1.GetOptions{}); err != nil {
				t.Fatal("lost lab", err)
			}
			if _, err := client.BatchV1().Jobs("cms-labs").Get(ctx, "checker-one", metav1.GetOptions{}); err != nil {
				t.Fatal("lost pending checker", err)
			}
		})
	}
}

func TestCheckerOutboxRetainsUntilAcknowledgmentCompletes(t *testing.T) {
	admin, client := checkerOutboxFixture(t)
	ctx := context.Background()
	if _, err := admin.CaptureCheckerResults(ctx, "lab-attempt"); err != nil {
		t.Fatal(err)
	}
	results, err := admin.QueuedCheckerResults(ctx)
	if err != nil || len(results) != 1 {
		t.Fatal(err)
	}
	fail := true
	client.PrependReactor("update", "jobs", func(ktesting.Action) (bool, runtime.Object, error) {
		if fail {
			return true, nil, errors.New("job update unavailable")
		}
		return false, nil, nil
	})
	if err := admin.AcknowledgeCheckerResult(ctx, results[0]); err == nil {
		t.Fatal("expected failure")
	}
	queued, err := admin.QueuedCheckerResults(ctx)
	if err != nil || len(queued) != 1 {
		t.Fatal("lost unacknowledged result", err)
	}
	fail = false
	if err := admin.AcknowledgeCheckerResult(ctx, results[0]); err != nil {
		t.Fatal(err)
	}
	job, err := client.BatchV1().Jobs("cms-labs").Get(ctx, "checker-one", metav1.GetOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if job.Annotations[CheckerSyncedAnnotation] != "true" || job.Spec.TTLSecondsAfterFinished == nil {
		t.Fatal("ack missing")
	}
}

func TestCheckerOutboxPreservesLargeResults(t *testing.T) {
	logger := zerolog.Nop()
	admin := NewKubernetesAdminWithClients(fake.NewSimpleClientset(), nil, nil, &logger)
	random := make([]byte, 700*1024)
	if _, err := rand.Read(random); err != nil {
		t.Fatal(err)
	}
	payload := `{"report":"` + base64.StdEncoding.EncodeToString(random) + `","max_score":1,"current_score":0,"result_display":"0/1"}`
	result := CheckerResult{AttemptID: "attempt", CheckID: "large", Namespace: "cms-labs", JobName: "job", Payload: payload, Logs: base64.StdEncoding.EncodeToString(random[:48*1024])}
	if err := admin.QueueCheckerResult(context.Background(), result); err != nil {
		t.Fatal(err)
	}
	queued, err := admin.QueuedCheckerResults(context.Background())
	if err != nil || len(queued) != 1 || queued[0].Payload != result.Payload || queued[0].Logs != result.Logs {
		t.Fatal("large result changed", err)
	}
}

func TestCheckerWatchWakesOnCompletion(t *testing.T) {
	for _, existing := range []bool{true, false} {
		t.Run(map[bool]string{true: "existing", false: "new"}[existing], func(t *testing.T) {
			job := &batchv1.Job{ObjectMeta: metav1.ObjectMeta{Name: "job", Namespace: "cms-labs", Labels: map[string]string{SessionManagedByLabel: SessionManagedByValue, SessionComponentLabel: SessionComponentChecker}}, Status: batchv1.JobStatus{Succeeded: 1}}
			client := fake.NewSimpleClientset()
			if existing {
				if _, err := client.BatchV1().Jobs("cms-labs").Create(context.Background(), job, metav1.CreateOptions{}); err != nil {
					t.Fatal(err)
				}
			}
			started := make(chan struct{})
			watcher := watch.NewRaceFreeFake()
			client.PrependWatchReactor("jobs", func(ktesting.Action) (bool, watch.Interface, error) { close(started); return true, watcher, nil })
			logger := zerolog.Nop()
			admin := NewKubernetesAdminWithClients(client, nil, nil, &logger)
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			wake := make(chan struct{}, 1)
			go admin.WatchCheckerResults(ctx, wake)
			if !existing {
				select {
				case <-started:
				case <-time.After(3 * time.Second):
					t.Fatal("watch not started")
				}
				watcher.Add(job)
			}
			select {
			case <-wake:
			case <-time.After(3 * time.Second):
				t.Fatal("completed check did not trigger immediate delivery")
			}
		})
	}
}

func TestDurableResultNotifiesThroughExistingJobWatch(t *testing.T) {
	client := fake.NewSimpleClientset(&batchv1.Job{ObjectMeta: metav1.ObjectMeta{Name: "job", Namespace: "cms-labs", Labels: map[string]string{SessionManagedByLabel: SessionManagedByValue, SessionComponentLabel: SessionComponentChecker}, Annotations: map[string]string{CheckerIDAnnotation: "check"}}, Status: batchv1.JobStatus{Succeeded: 1}})
	started := make(chan struct{})
	var once sync.Once
	client.PrependWatchReactor("jobs", func(ktesting.Action) (bool, watch.Interface, error) {
		once.Do(func() { close(started) })
		return false, nil, nil
	})
	logger := zerolog.Nop()
	admin := NewKubernetesAdminWithClients(client, nil, nil, &logger)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	wake := make(chan struct{}, 1)
	go admin.WatchCheckerResults(ctx, wake)
	select {
	case <-wake:
	case <-time.After(3 * time.Second):
		t.Fatal("initial checker not observed")
	}
	select {
	case <-started:
	case <-time.After(3 * time.Second):
		t.Fatal("job watch not started")
	}
	result := CheckerResult{Namespace: "cms-labs", JobName: "job", AttemptID: "attempt", CheckID: "check", Payload: `{"max_score":1,"current_score":1,"result_display":"1/1"}`}
	if err := admin.QueueCheckerResult(ctx, result); err != nil {
		t.Fatal(err)
	}
	select {
	case <-wake:
	case <-time.After(3 * time.Second):
		t.Fatal("durable result waited for polling")
	}
	if err := admin.QueueCheckerResult(ctx, result); err != nil {
		t.Fatal(err)
	}
	count := 0
	for _, action := range client.Actions() {
		if action.GetVerb() == "update" && action.GetResource().Resource == "jobs" {
			count++
		}
	}
	if count != 1 {
		t.Fatalf("outbox wake must not cause a reconcile loop: %d updates", count)
	}
}
