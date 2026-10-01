package routes

import (
	"github.com/cms-lab-core/cms-labs-api/shared/jsonrpc"
	fiber "github.com/gofiber/fiber/v2"
)

func FiberRoutes(app *fiber.App) {
	// Routes.
	SwaggerRoute(app)
	V1SSORoute(app)
	V1RpcRoute(app)
	V2LTIRoutes(app)
	jsonrpc.EmptyRoutes(app)
}
