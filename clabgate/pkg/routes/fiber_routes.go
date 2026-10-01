package routes

import (
	"github.com/cms-lab-core/cms-labs-api/shared/jsonrpc"
	fiber "github.com/gofiber/fiber/v2"
)

func FiberRoutes(app *fiber.App) {
	// Routes.
	SwaggerRoute(app)
	WorkspaceAuthRoutes(app)
	V1RpcRoute(app)
	jsonrpc.EmptyRoutes(app)
}
