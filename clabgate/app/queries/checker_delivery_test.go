package queries

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/rs/zerolog"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	kubernetesfake "k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/rest"
)

func TestPendingCheckerResultsIsolatesBrokenJob(t *testing.T) {
	for _, failure := range []string{"missing result", "bad record", "pod list unavailable", "logs unavailable"} {
		t.Run(failure, func(t *testing.T) {
			payload := `{"max_score":2,"current_score":1,"result_display":"1/2"}`
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if strings.HasSuffix(r.URL.Path, "/log") {
					w.Header().Set("Content-Type", "text/plain")
					if strings.Contains(r.URL.Path, "/broken/") {
						switch failure {
						case "logs unavailable":
							http.Error(w, "forbidden", http.StatusForbidden)
						case "bad record":
							_, _ = w.Write([]byte("CMS_LABS_CHECKER_RESULT_V1 bad\nSSH failed\n"))
						default:
							_, _ = w.Write([]byte("SSH failed\n"))
						}
					} else {
						_, _ = w.Write([]byte(makeResultLog(payload)))
					}
					return
				}
				w.Header().Set("Content-Type", "application/json")
				switch r.URL.Path {
				case "/apis/batch/v1/namespaces/lab-test/jobs":
					jobs := batchv1.JobList{}
					for _, name := range []string{"broken", "healthy"} {
						jobs.Items = append(jobs.Items, batchv1.Job{
							ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: "lab-test",
								Labels:      map[string]string{SessionIDLabel: "attempt"},
								Annotations: map[string]string{CheckerIDAnnotation: name}},
							Status: batchv1.JobStatus{Succeeded: 1},
						})
					}
					_ = json.NewEncoder(w).Encode(jobs)
				case "/api/v1/namespaces/lab-test/pods":
					name := "healthy"
					if strings.Contains(r.URL.Query().Get("labelSelector"), "broken") {
						name = "broken"
						if failure == "pod list unavailable" {
							http.Error(w, "unavailable", http.StatusServiceUnavailable)
							return
						}
					}
					_ = json.NewEncoder(w).Encode(corev1.PodList{Items: []corev1.Pod{{
						ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: "lab-test"},
						Status: corev1.PodStatus{ContainerStatuses: []corev1.ContainerStatus{{
							Name: "checker", State: corev1.ContainerState{
								Terminated: &corev1.ContainerStateTerminated{ExitCode: 0},
							},
						}}},
					}}})
				default:
					http.NotFound(w, r)
				}
			}))
			defer server.Close()
			client, err := kubernetes.NewForConfig(&rest.Config{Host: server.URL})
			if err != nil {
				t.Fatal(err)
			}
			logger := zerolog.Nop()
			admin := NewKubernetesAdminWithClients(client, nil, nil, &logger)
			results, err := admin.PendingCheckerResults(context.Background(), "lab-test")
			if err != nil || len(results) != 2 {
				t.Fatalf("one bad job blocked delivery: count=%d error=%v", len(results), err)
			}
			if results[0].Error == nil || results[1].Error != nil || results[1].Payload != payload {
				t.Fatalf("unexpected isolated results: %+v", results)
			}
			if failure == "missing result" && !strings.Contains(results[0].Logs, "SSH failed") {
				t.Fatal("failed checker lost diagnostics")
			}
		})
	}
}

func TestCheckerJobRetainedUntilResultAcknowledged(t *testing.T) {
	ctx := context.Background()
	client := kubernetesfake.NewSimpleClientset()
	logger := zerolog.Nop()
	admin := NewKubernetesAdminWithClients(client, nil, nil, &logger)
	_, err := admin.RunChecker(ctx, RunCheckerParams{
		Namespace: "lab-test", SessionID: "attempt", AttemptID: "attempt",
		OwnerID: "student", TestPath: "smoke", Image: "checker:test", TimeoutSeconds: 60,
	})
	if err != nil {
		t.Fatal(err)
	}
	jobs, err := client.BatchV1().Jobs("lab-test").List(ctx, metav1.ListOptions{})
	if err != nil || len(jobs.Items) != 1 || jobs.Items[0].Spec.TTLSecondsAfterFinished != nil {
		t.Fatal("unsaved checker result has automatic TTL cleanup")
	}
	if err := admin.MarkCheckerResultSynced(ctx, "lab-test", jobs.Items[0].Name); err != nil {
		t.Fatal(err)
	}
	jobs, err = client.BatchV1().Jobs("lab-test").List(ctx, metav1.ListOptions{})
	if err != nil || len(jobs.Items) != 1 {
		t.Fatal("acknowledged checker job disappeared")
	}
	job := jobs.Items[0]
	if job.Annotations[CheckerSyncedAnnotation] != "true" ||
		job.Spec.TTLSecondsAfterFinished == nil || *job.Spec.TTLSecondsAfterFinished != 86400 {
		t.Fatal("acknowledged job must enable TTL cleanup")
	}
}
