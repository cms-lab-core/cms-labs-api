package routes

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/cms-lab-core/cms-labs-api/clabgate/app/usecases"
	"github.com/cms-lab-core/cms-labs-api/clabgate/pkg/configs"
	"github.com/cms-lab-core/cms-labs-api/shared/jsonrpc"
	"github.com/gofiber/fiber/v2"
)

func TestWorkspaceAuthFailuresRemainUnauthorizedWithJSONRPCMiddleware(t *testing.T) {
	app := fiber.New()
	app.Use(jsonrpc.InternalFormatterNew(false))
	WorkspaceAuthRoutes(app)

	for _, path := range []string{
		"/clabgate/workspace-auth/exchange",
		"/clabgate/workspace-auth/verify",
	} {
		response, err := app.Test(httptest.NewRequest(http.MethodGet, path, nil))
		if err != nil {
			t.Fatalf("GET %s: %v", path, err)
		}
		if response.StatusCode != http.StatusUnauthorized {
			t.Fatalf("GET %s returned %d, want %d", path, response.StatusCode, http.StatusUnauthorized)
		}
	}
}

func TestWorkspaceVerifyReturnsSeparateTTYDIdentity(t *testing.T) {
	secret := strings.Repeat("s", 32)
	sessionID := "00000000-0000-0000-0000-000000000000"
	destination := "/clabgate/workspace/" + sessionID + "/terminal/r1-terminal/ws"
	identity := usecases.NewWorkspaceIdentity("user-42", strings.Repeat("student", 8), "Student", "", nil)
	grant, err := usecases.IssueWorkspaceGrant(secret, sessionID, destination, identity, time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	cookie, _, _, _, err := usecases.ExchangeWorkspaceGrant(secret, grant, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	previousConfig := configs.AppConfig
	configs.AppConfig = &configs.AppConfigModel{Session: &configs.SessionConfig{WorkspaceSecret: secret}}
	t.Cleanup(func() { configs.AppConfig = previousConfig })

	app := fiber.New()
	WorkspaceAuthRoutes(app)
	request := httptest.NewRequest(http.MethodGet, "/clabgate/workspace-auth/verify", nil)
	request.Header.Set("Cookie", usecases.WorkspaceCookieName+"="+cookie)
	request.Header.Set("X-Original-URI", destination)
	response, err := app.Test(request)
	if err != nil {
		t.Fatal(err)
	}
	if response.StatusCode != http.StatusNoContent {
		t.Fatalf("verify returned %d, want %d", response.StatusCode, http.StatusNoContent)
	}
	fullIdentity, err := usecases.EncodeWorkspaceIdentity(identity)
	if err != nil {
		t.Fatal(err)
	}
	if got := response.Header.Get("X-CMS-Identity"); got != fullIdentity {
		t.Fatalf("full identity = %q, want %q", got, fullIdentity)
	}
	compact := response.Header.Get("X-CMS-Terminal-Identity")
	if compact == "" || len(compact) >= 30 {
		t.Fatalf("terminal identity must contain 1..29 bytes, got %q (%d)", compact, len(compact))
	}
}
