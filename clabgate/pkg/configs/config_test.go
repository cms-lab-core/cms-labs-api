package configs

import "testing"

func TestWorkspaceGrantTTLConfiguration(t *testing.T) {
	t.Setenv("WORKSPACE_GRANT_TTL_SECONDS", "")
	if got := NewAppConfigModel().Session.WorkspaceGrantTTL; got != 300 {
		t.Fatalf("default workspace grant TTL = %d, want 300", got)
	}

	t.Setenv("WORKSPACE_GRANT_TTL_SECONDS", "600")
	if got := NewAppConfigModel().Session.WorkspaceGrantTTL; got != 600 {
		t.Fatalf("configured workspace grant TTL = %d, want 600", got)
	}

	t.Setenv("WORKSPACE_GRANT_TTL_SECONDS", "invalid")
	if got := NewAppConfigModel().Session.WorkspaceGrantTTL; got != 300 {
		t.Fatalf("invalid workspace grant TTL fallback = %d, want 300", got)
	}
}

func TestWorkspaceCookieSecureConfiguration(t *testing.T) {
	t.Setenv("WORKSPACE_COOKIE_SECURE", "false")
	if NewAppConfigModel().Session.WorkspaceCookieSecure {
		t.Fatal("WORKSPACE_COOKIE_SECURE=false must allow the loopback HTTP demo")
	}

	t.Setenv("WORKSPACE_COOKIE_SECURE", "true")
	if !NewAppConfigModel().Session.WorkspaceCookieSecure {
		t.Fatal("WORKSPACE_COOKIE_SECURE=true must keep production cookies HTTPS-only")
	}

	t.Setenv("WORKSPACE_COOKIE_SECURE", "invalid")
	if !NewAppConfigModel().Session.WorkspaceCookieSecure {
		t.Fatal("invalid WORKSPACE_COOKIE_SECURE must use the secure default")
	}
}
