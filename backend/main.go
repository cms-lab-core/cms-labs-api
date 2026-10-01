// Package main CMS Labs Core проект LTI <-> SSO компонентов
package main

import (
	"flag"
	"fmt"

	fiber "github.com/gofiber/fiber/v2"
	_ "github.com/joho/godotenv/autoload" // load .env file automatically
	"github.com/maintainer64/cms-labs-api/backend/app/di"
	_ "github.com/maintainer64/cms-labs-api/backend/docs" // load API Docs files (Swagger)
	"github.com/maintainer64/cms-labs-api/backend/pkg/configs"
	"github.com/maintainer64/cms-labs-api/backend/pkg/middleware"
	"github.com/maintainer64/cms-labs-api/backend/pkg/routes"
	"github.com/maintainer64/cms-labs-api/backend/pkg/utils"
	"github.com/maintainer64/cms-labs-api/shared/logs"
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
	task := flag.String("task", "", "Task name")
	demo := flag.Bool("demo", false, "Seed local full-stack demo records and exit")
	flag.Parse()
	taskArgs := flag.Args()
	if *demo {
		*task = "demo_seed"
		if len(taskArgs) > 1 {
			panic(fmt.Errorf("--demo accepts at most one catalog JSON file"))
		}
	}
	config := configs.FiberConfig()
	logs.ZeroLogInit(configs.AppConfig.Debug)

	container, err := di.NewDIContainer(&logs.ZeroLoggerConf{Name: "main"})
	if err != nil {
		panic(err)
	}

	startup := container.TaskStartup()
	taskCompleted, err := startup.Startup(task, taskArgs...)
	if err != nil {
		panic(err)
	}
	if taskCompleted {
		return
	}

	// Define a new Fiber app with config.
	app := fiber.New(config)

	// Middlewares.
	middleware.FiberMiddleware(app) // Register Fiber's middleware for app.
	routes.FiberRoutes(app)         // Register Fiber's routes for app.

	// Start server (with or without graceful shutdown).
	if configs.AppConfig.Server.Layer == "dev" {
		utils.StartServer(app)
	} else {
		utils.StartServerWithGracefulShutdown(app)
	}
}
