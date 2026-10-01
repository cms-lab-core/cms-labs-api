package usecases

import (
	"strings"
	"testing"
	"time"
)

func TestWorkspaceGrantExchangeAndCookieScope(t *testing.T) {
	secret := strings.Repeat("s", 32)
	sessionID := "00000000-0000-0000-0000-000000000000"
	destination := "/clabgate/workspace/" + sessionID + "/lab?path=Lab.ipynb"
	identity := NewWorkspaceIdentity("42", "student", "Student Name", "student@example.test", []string{"student"})
	grant, err := IssueWorkspaceGrant(secret, sessionID, destination, identity, time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	cookie, gotDestination, gotSessionID, gotIdentity, err := ExchangeWorkspaceGrant(secret, grant, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	if gotDestination != destination || gotSessionID != sessionID {
		t.Fatalf("unexpected exchange result: %q %q", gotDestination, gotSessionID)
	}
	if gotIdentity.Subject != identity.Subject || gotIdentity.Name != identity.Name {
		t.Fatalf("unexpected identity: %#v", gotIdentity)
	}
	verifiedIdentity, err := VerifyWorkspaceCookie(secret, cookie, destination)
	if err != nil {
		t.Fatalf("valid cookie rejected: %v", err)
	}
	if verifiedIdentity.Username != "student" {
		t.Fatalf("unexpected verified identity: %#v", verifiedIdentity)
	}
	if _, err := VerifyWorkspaceCookie(secret, cookie, "/clabgate/workspace/aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa/lab"); err == nil {
		t.Fatal("cookie authorized another session")
	}
}

func TestWorkspaceGrantRejectsOpenRedirectAndWeakSecret(t *testing.T) {
	identity := NewWorkspaceIdentity("42", "student", "Student", "", nil)
	if _, err := IssueWorkspaceGrant("short", "session", "/clabgate/workspace/session/lab", identity, time.Minute); err == nil {
		t.Fatal("weak secret was accepted")
	}
	if _, err := IssueWorkspaceGrant(strings.Repeat("s", 32), "session", "https://example.test/steal", identity, time.Minute); err == nil {
		t.Fatal("absolute redirect was accepted")
	}
}
