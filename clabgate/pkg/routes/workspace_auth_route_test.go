package routes

import (
	"net/http"
	"net/http/httptest"
	"testing"

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
