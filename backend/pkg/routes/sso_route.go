package routes

import (
	"github.com/cms-lab-core/cms-labs-api/backend/app/controllers"
	fiber "github.com/gofiber/fiber/v2"
)

// V1SSORoute func for describe group of openid protocol auth.
func V1SSORoute(a *fiber.App) {
	group := a.Group("/api/v1/sso")
	group.Get("/authorize", controllers.SSOAuthorize)
	group.Post("/token", controllers.SSOToken)
	group.Post("/introspect", controllers.SSOIntrospect)
	group.Get("/userinfo", controllers.SSOUserInfo)
	group.Get("/jwks", controllers.SSOJwks)
	group.Get("/.well-known/openid-configuration", controllers.SSOOpenIdConfiguration)
}
