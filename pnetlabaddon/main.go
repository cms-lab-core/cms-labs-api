// Дополнение в PNETLab. API для аутентификации с помощью общего компонента SSO
package main

import (
	"github.com/cms-lab-core/cms-labs-api/pnetlabaddon/app/di"
	_ "github.com/cms-lab-core/cms-labs-api/pnetlabaddon/docs" // load API Docs files (Swagger)
	"github.com/cms-lab-core/cms-labs-api/pnetlabaddon/pkg/configs"
	"github.com/cms-lab-core/cms-labs-api/pnetlabaddon/pkg/middleware"
	"github.com/cms-lab-core/cms-labs-api/pnetlabaddon/pkg/routes"
	"github.com/cms-lab-core/cms-labs-api/pnetlabaddon/pkg/utils"
	"github.com/cms-lab-core/cms-labs-api/shared/logs"
	"github.com/cms-lab-core/cms-labs-api/shared/scheduler"
	fiber "github.com/gofiber/fiber/v2"
	_ "github.com/joho/godotenv/autoload" // load .env file automatically
)

// @title API
// @version 1.0
// @description This is an auto-generated API Docs.
// @termsOfService http://swagger.io/terms/
// @contact.name API Support
// @contact.email your@mail.com
// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html
// @BasePath /
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
func main() {
	// Define Fiber config.
	config := configs.FiberConfig()
	logs.ZeroLogInit(configs.AppConfig.Debug)

	// Define a new Fiber app with config.
	app := fiber.New(config)

	// Define scheduler app with config
	schedulerInterval := scheduler.NewScheduler(
		configs.AppConfig.SchedulerInterval,
	).AddTask(&di.PnetServerPingTask{})

	// Middlewares.
	middleware.FiberMiddleware(app) // Register Fiber's middleware for app.
	routes.FiberRoutes(app)         // Register Fiber's routes for app.

	// Start server (with or without graceful shutdown).
	if configs.AppConfig.Server.Layer == "dev" {
		utils.StartServer(app, schedulerInterval)
	} else {
		utils.StartServerWithGracefulShutdown(app, schedulerInterval)
	}
}
