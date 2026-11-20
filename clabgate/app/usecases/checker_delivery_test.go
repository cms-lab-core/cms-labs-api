package usecases

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/cms-lab-core/cms-labs-api/clabgate/app/queries"
	"github.com/cms-lab-core/cms-labs-api/shared/cms_client"
	"github.com/go-resty/resty/v2"
	"github.com/rs/zerolog"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func deliveryUC(client *fake.Clientset, url string) *SessionsUC {
	logger := zerolog.Nop()
	return &SessionsUC{Logger: &logger, KubernetesAdminQuery: queries.NewKubernetesAdminWithClients(client, nil, nil, &logger), CMSClient: cms_client.NewCMSClient(&cms_client.CMSClientConfig{BaseUrl: url, ClientID: "test", Token: "test-only", MaxTimeoutSeconds: 2}, resty.New())}
}

func TestQueuedCheckerResultRetriesWithoutLabAfterFailureAndRestart(t *testing.T) {
	var mode atomic.Int32
	var mutex sync.Mutex
	var grades []checkerGrade
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var request struct {
			Method string `json:"method"`
			Params struct {
				Models []struct {
					Status string       `json:"status"`
					Result checkerGrade `json:"result"`
				} `json:"models"`
			} `json:"params"`
		}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Error(err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		if request.Method == "lti_attempt.list_external" {
			_, _ = w.Write([]byte(`{"result":{"model":[]}}`))
			return
		}
		if len(request.Params.Models) != 1 {
			t.Error("unexpected update")
			return
		}
		update := request.Params.Models[0]
		if update.Status != "" {
			t.Error("late result must not reopen a completed lab")
		}
		mutex.Lock()
		grades = append(grades, update.Result)
		mutex.Unlock()
		if mode.Load() == 0 {
			http.Error(w, "unavailable", 503)
			return
		}
		count := 0
		if mode.Load() == 2 {
			count = 1
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"result": map[string]int{"count": count}})
	}))
	defer server.Close()
	client := fake.NewSimpleClientset()
	ctx := context.Background()
	u := deliveryUC(client, server.URL)
	at := time.Now().UTC()
	result := queries.CheckerResult{Namespace: "cms-labs", JobName: "gone", AttemptID: "attempt", CheckID: "check", CreatedAt: at, Payload: `{"max_score":1,"current_score":0,"result_display":"0/1"}`}
	if err := u.KubernetesAdminQuery.QueueCheckerResult(ctx, result); err != nil {
		t.Fatal(err)
	}
	for _, state := range []int32{0, 1, 2} {
		mode.Store(state)
		restarted := deliveryUC(client, server.URL)
		changes, err := restarted.Reconcile(ctx)
		if err != nil {
			t.Fatal(err)
		}
		queued, err := restarted.KubernetesAdminQuery.QueuedCheckerResults(ctx)
		if err != nil {
			t.Fatal(err)
		}
		if state < 2 {
			if changes != 0 || len(queued) != 1 {
				t.Fatal("unacknowledged result disappeared")
			}
		} else if changes != 1 || len(queued) != 0 {
			t.Fatal("acknowledged result not drained")
		}
	}
	mutex.Lock()
	defer mutex.Unlock()
	if len(grades) != 3 {
		t.Fatal("result not retried")
	}
	for _, grade := range grades {
		if grade.CheckID != "check" || !grade.CheckedAt.Equal(at) || grade.CurrentScore != 0 {
			t.Fatalf("retry changed identity/grade: %+v", grade)
		}
	}
}

func TestStopWaitsForCheckerAndDeliversBeforeCompletion(t *testing.T) {
	var mutex sync.Mutex
	status := "active"
	saved := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var request struct {
			Method string `json:"method"`
			Params struct {
				Models []struct {
					Status string        `json:"status"`
					Result *checkerGrade `json:"result"`
				} `json:"models"`
			} `json:"params"`
		}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Error(err)
			return
		}
		mutex.Lock()
		defer mutex.Unlock()
		w.Header().Set("Content-Type", "application/json")
		if request.Method == "lti_attempt.list_external" {
			items := []map[string]interface{}{}
			if status != "completed" {
				items = append(items, map[string]interface{}{"attempt_id": "attempt", "status": status, "user_id": 42})
			}
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"result": map[string]interface{}{"model": items}})
			return
		}
		for _, update := range request.Params.Models {
			if update.Result != nil {
				saved = true
			}
			if update.Status == "completed" && !saved {
				t.Error("lab completed before durable grade")
			}
			if update.Status != "" {
				status = update.Status
			}
		}
		_, _ = w.Write([]byte(`{"result":{"count":1}}`))
	}))
	defer server.Close()
	labels := map[string]string{queries.SessionManagedByLabel: queries.SessionManagedByValue, queries.SessionComponentLabel: queries.SessionComponentChecker, queries.SessionIDLabel: "attempt", queries.SessionOwnerLabel: "42"}
	client := fake.NewSimpleClientset(&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "lab-attempt", Annotations: map[string]string{queries.SessionAttemptAnnotation: "attempt"}, Labels: map[string]string{queries.SessionManagedByLabel: queries.SessionManagedByValue, queries.SessionIDLabel: "attempt", queries.SessionOwnerLabel: "42"}}}, &batchv1.Job{ObjectMeta: metav1.ObjectMeta{Name: "checker", Namespace: "lab-attempt", Labels: labels, Annotations: map[string]string{queries.CheckerIDAnnotation: "check"}}})
	u := deliveryUC(client, server.URL).SetContext(&cms_client.SSOTokenPublicData{Sub: "42", Roles: []string{"student"}})
	ctx := context.Background()
	stopped, err := u.Stop(ctx, SessionStopInputDTO{SessionID: "attempt"})
	if err != nil || !stopped.Stopped {
		t.Fatal("stop not accepted", err)
	}
	if _, err := client.CoreV1().Namespaces().Get(ctx, "lab-attempt", metav1.GetOptions{}); err != nil {
		t.Fatal("running checker lost lab", err)
	}
	job, err := client.BatchV1().Jobs("lab-attempt").Get(ctx, "checker", metav1.GetOptions{})
	if err != nil {
		t.Fatal(err)
	}
	job.Status.Succeeded = 1
	if _, err := client.BatchV1().Jobs("lab-attempt").Update(ctx, job, metav1.UpdateOptions{}); err != nil {
		t.Fatal(err)
	}
	if err := u.KubernetesAdminQuery.QueueCheckerResult(ctx, queries.CheckerResult{JobName: "checker", Namespace: "lab-attempt", AttemptID: "attempt", CheckID: "check", CreatedAt: time.Now().UTC(), Payload: `{"max_score":1,"current_score":1,"result_display":"1/1"}`}); err != nil {
		t.Fatal(err)
	}
	if _, err := u.Reconcile(ctx); err != nil {
		t.Fatal(err)
	}
	if _, err := client.CoreV1().Namespaces().Get(ctx, "lab-attempt", metav1.GetOptions{}); !apierrors.IsNotFound(err) {
		t.Fatal("finished lab not cleaned", err)
	}
	if _, err := u.Reconcile(ctx); err != nil {
		t.Fatal(err)
	}
	mutex.Lock()
	defer mutex.Unlock()
	if status != "completed" || !saved {
		t.Fatal("final grade or completion lost")
	}
}
