package middleware

import (
	fiber "github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/healthcheck"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/maintainer64/cms-labs-api/clabgate/pkg/configs"
	"github.com/maintainer64/cms-labs-api/shared/jsonrpc"
	"github.com/maintainer64/cms-labs-api/shared/logs"
)

// FiberMiddleware provide Fiber's built-in middlewares.
// See: https://docs.gofiber.io/api/middleware
func FiberMiddleware(a *fiber.App) {
	a.Use(
		// Add CORS to each route.
		cors.New(),
		// Options запросы всегда 204
		logs.NewOptionsFiberNoContent(),
		// Header X-Request-ID
		requestid.New(requestid.Config{
			ContextKey: "x-request-id",
		}),
		// Add simple healthcheck.
		healthcheck.New(),
		// Add simple logger.
		logs.NewFiberZerologLogger(),
		// InternalFormatterException
		jsonrpc.InternalFormatterNew(configs.AppConfig.Debug),
	)
}
