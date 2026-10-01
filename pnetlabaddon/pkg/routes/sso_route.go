package routes

import (
	"github.com/cms-lab-core/cms-labs-api/pnetlabaddon/app/controllers"
	fiber "github.com/gofiber/fiber/v2"
)

// V1SSORoute func for describe group of openid protocol auth.
func V1SSORoute(a *fiber.App) {
	group := a.Group("/pnet-lab-addon/api/v1/sso")
	group.Get("/login", controllers.SSOFirstFactor)
	group.Get("/openid", controllers.SSOSecondFactor)
}
