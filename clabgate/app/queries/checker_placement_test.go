package queries

import (
	"context"
	"fmt"
	"testing"

	"github.com/rs/zerolog"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	ktesting "k8s.io/client-go/testing"
)

func TestSharedCheckerNamespaceKeepsSessionsSeparate(t *testing.T) {
	ctx := context.Background()
	client := fake.NewSimpleClientset(
		&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "lab-one"}},
		&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "lab-two"}},
	)
	generated := 0
	client.PrependReactor("create", "jobs", func(action ktesting.Action) (bool, runtime.Object, error) {
		job := action.(ktesting.CreateAction).GetObject().(*batchv1.Job)
		if job.Name == "" {
			generated++
			job.Name = fmt.Sprintf("checker-%d", generated)
		}
		return false, nil, nil
	})
	logger := zerolog.Nop()
	admin := NewKubernetesAdminWithClients(client, nil, nil, &logger)
	admin.checkerNamespace = "cms-labs"
	admin.checkerServiceAccount = "custom-checker"
	params := RunCheckerParams{Namespace: "lab-one", SessionID: "one", OwnerID: "42", Image: "checker:test", TimeoutSeconds: 60}
	first, err := admin.RunChecker(ctx, params)
	if err != nil {
		t.Fatal(err)
	}
	again, err := admin.RunChecker(ctx, params)
	if err != nil || again != first {
		t.Fatalf("duplicate check for same session: %q %v", again, err)
	}
	params.Namespace, params.SessionID = "lab-two", "two"
	second, err := admin.RunChecker(ctx, params)
	if err != nil || second == first {
		t.Fatalf("another session blocked by active checker: %q %v", second, err)
	}
	jobs, err := client.BatchV1().Jobs("cms-labs").List(ctx, metav1.ListOptions{})
	if err != nil || len(jobs.Items) != 2 {
		t.Fatalf("shared jobs: %v %v", jobs, err)
	}
	for _, job := range jobs.Items {
		pod := job.Spec.Template.Spec
		if pod.ServiceAccountName != "custom-checker" || pod.AutomountServiceAccountToken == nil || !*pod.AutomountServiceAccountToken {
			t.Fatal("configured API identity not used")
		}
		var target string
		for _, env := range pod.Containers[0].Env {
			if env.Name == "SESSION_NAMESPACE" {
				target = env.Value
			}
		}
		if target == "cms-labs" || target != job.Labels[CheckerTargetNamespaceLabel] {
			t.Fatalf("lost target namespace: %+v", job)
		}
	}
	for _, action := range client.Actions() {
		if action.GetVerb() == "create" && action.GetResource().Resource != "jobs" {
			t.Fatalf("clabgate must not provision identities/RBAC: %v", action)
		}
	}
	// Completed pre-upgrade checks remain visible alongside the new shared Jobs.
	_, err = client.BatchV1().Jobs("lab-one").Create(ctx, &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{Name: "legacy", Labels: map[string]string{SessionComponentLabel: SessionComponentChecker}},
	}, metav1.CreateOptions{})
	if err != nil {
		t.Fatal(err)
	}
	one, err := admin.checkerJobsForLab(ctx, "lab-one")
	if err != nil || len(one.Items) != 2 {
		t.Fatalf("lost legacy result or mixed sessions: %+v %v", one, err)
	}
	two, err := admin.checkerJobsForLab(ctx, "lab-two")
	if err != nil || len(two.Items) != 1 {
		t.Fatalf("another session's jobs leaked: %+v %v", two, err)
	}
}

func TestDeleteSessionCleansOnlyItsSharedCheckerJobs(t *testing.T) {
	client := fake.NewSimpleClientset(&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{
		Name: "lab-session", Labels: map[string]string{SessionManagedByLabel: SessionManagedByValue, SessionIDLabel: "session", SessionOwnerLabel: "42"},
	}})
	logger := zerolog.Nop()
	admin := NewKubernetesAdminWithClients(client, nil, nil, &logger)
	admin.checkerNamespace = "cms-labs"
	for _, name := range []string{"session", "another"} {
		_, err := client.BatchV1().Jobs("cms-labs").Create(context.Background(), &batchv1.Job{ObjectMeta: metav1.ObjectMeta{Name: name, Labels: map[string]string{SessionComponentLabel: SessionComponentChecker, CheckerTargetNamespaceLabel: "lab-" + name, SessionIDLabel: name, SessionOwnerLabel: "42"}}, Status: batchv1.JobStatus{Failed: 1}}, metav1.CreateOptions{})
		if err != nil {
			t.Fatal(err)
		}
	}
	if err := admin.DeleteSession(context.Background(), "session", "42", false); err != nil {
		t.Fatal(err)
	}
	jobs, err := client.BatchV1().Jobs("cms-labs").List(context.Background(), metav1.ListOptions{})
	if err != nil || len(jobs.Items) != 1 || jobs.Items[0].Name != "another" {
		t.Fatalf("session cleanup touched another session or retained its own job: %+v %v", jobs, err)
	}
}
