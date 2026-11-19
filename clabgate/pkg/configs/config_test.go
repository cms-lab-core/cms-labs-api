package configs

import "testing"

func TestCheckerPlacementConfiguration(t *testing.T) {
	t.Setenv("CHECKER_NAMESPACE", "")
	t.Setenv("CHECKER_SERVICE_ACCOUNT", "")
	t.Setenv("POD_NAMESPACE", "")
	config := NewAppConfigModel().Session
	if config.CheckerNamespace != "cms-labs" || config.CheckerServiceAccount != "cms-labs-checker" {
		t.Fatalf("unexpected checker defaults: %+v", config)
	}
	t.Setenv("POD_NAMESPACE", "cms-pre")
	if NewAppConfigModel().Session.CheckerNamespace != "cms-pre" {
		t.Fatal("checker must follow the platform namespace")
	}
	t.Setenv("CHECKER_NAMESPACE", "checks")
	t.Setenv("CHECKER_SERVICE_ACCOUNT", "my-checker")
	config = NewAppConfigModel().Session
	if config.CheckerNamespace != "checks" || config.CheckerServiceAccount != "my-checker" {
		t.Fatal("checker placement overrides ignored")
	}
}

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

func TestJupyterExtraAnnotations(t *testing.T) {
	for _, value := range []string{"", "{}", "null"} {
		t.Setenv("JUPYTER_EXTRA_ANNOTATIONS", value)
		if len(NewAppConfigModel().Session.JupyterExtraAnnotations) != 0 {
			t.Fatal("annotations must be empty by default")
		}
	}
	t.Setenv("JUPYTER_EXTRA_ANNOTATIONS", `{"example.org/pool":"jupyter","example.org/empty":""}`)
	got := NewAppConfigModel().Session.JupyterExtraAnnotations
	if got["example.org/pool"] != "jupyter" || len(got) != 2 {
		t.Fatalf("annotations: %#v", got)
	}
}
func TestJupyterExtraAnnotationsRejectInvalidConfig(t *testing.T) {
	for _, value := range []string{`{`, `{"pool":true}`, `["pool"]`} {
		t.Run(value, func(t *testing.T) {
			t.Setenv("JUPYTER_EXTRA_ANNOTATIONS", value)
			defer func() {
				if recover() == nil {
					t.Fatal("invalid annotation config was silently accepted")
				}
			}()
			NewAppConfigModel()
		})
	}
}
