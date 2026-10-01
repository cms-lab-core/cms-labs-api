package routes

import (
	"github.com/cms-lab-core/cms-labs-api/pnetlabaddon/app/controllers"
	fiber "github.com/gofiber/fiber/v2"
)

// V1TaskRoute func for describe group of task routes.
func V1TaskRoute(a *fiber.App) {
	group := a.Group("/pnet-lab-addon/api/v1/task")
	group.Post("/pnet-server-ping", controllers.PNETServerPing)
}
