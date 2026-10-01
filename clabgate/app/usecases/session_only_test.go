package usecases

import (
	"testing"

	"github.com/cms-lab-core/cms-labs-api/shared/cms_client"
	"github.com/cms-lab-core/cms-labs-api/shared/jsonrpc"
)

func TestSessionAttemptAccessForStudentAndOperator(t *testing.T) {
	attempt := cms_client.ListAttemptsModel{UserID: 42, UserName: "Student"}
	student := &cms_client.SSOTokenPublicData{Sub: "7", Username: "Current student", Roles: []string{"student"}}
	operator := &cms_client.SSOTokenPublicData{Sub: "1", Username: "Teacher", Roles: []string{"instructor"}}

	studentFilter := sessionAttemptUserFilter(student)
	if len(studentFilter) != 1 || studentFilter[0] != 7 {
		t.Fatalf("student filter = %#v, want [7]", studentFilter)
	}
	if operatorFilter := sessionAttemptUserFilter(operator); len(operatorFilter) != 0 {
		t.Fatalf("operator filter = %#v, want no user restriction", operatorFilter)
	}

	studentOwner, studentName := sessionAttemptOwner(student, attempt)
	if studentOwner != "7" || studentName != "Current student" {
		t.Fatalf("student identity = %q/%q", studentOwner, studentName)
	}
	operatorOwner, operatorName := sessionAttemptOwner(operator, attempt)
	if operatorOwner != "42" || operatorName != "Student" {
		t.Fatalf("operator identity = %q/%q", operatorOwner, operatorName)
	}
}

func TestTopologyGetRequiresSessionID(t *testing.T) {
	_, err := (&TopologiesGetUC{}).
		SetContext(&cms_client.SSOTokenPublicData{Sub: "42"}).
		Execute(TopologiesGetInputDTO{})
	rpcErr, ok := err.(jsonrpc.RpcError)
	if !ok || rpcErr.Code != "session_required" {
		t.Fatalf("expected session_required, got %#v", err)
	}
}

func TestNodeActionRequiresSessionID(t *testing.T) {
	_, err := (&NodeActionsUC{}).
		SetContext(&cms_client.SSOTokenPublicData{Sub: "42"}).
		Execute(NodeActionInputDTO{Actions: []NodeActionItem{{Node: "r1", Action: "restart"}}})
	if err == nil || err.Error() != "session_id is required" {
		t.Fatalf("expected session_id validation error, got %#v", err)
	}
}
