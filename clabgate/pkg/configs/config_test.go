package configs

import "testing"

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
