package routes

import (
	"github.com/cms-lab-core/cms-labs-api/backend/app/controllers"
	fiber "github.com/gofiber/fiber/v2"
)

// V2LTIRoutes func for describe group of LTI routes.
func V2LTIRoutes(a *fiber.App) {
	group := a.Group("/api/v2/lti")
	group.Post("/login", controllers.LTILogin)
	group.Get("/login", controllers.LTILogin)
	group.Post("/launch", controllers.LTILaunch)
	group.Get("/launch", controllers.LTILaunch)
}
