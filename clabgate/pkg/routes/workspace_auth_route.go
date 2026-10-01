package routes

import (
	"net/http"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/maintainer64/cms-labs-api/clabgate/app/usecases"
	"github.com/maintainer64/cms-labs-api/clabgate/pkg/configs"
)

func WorkspaceAuthRoutes(app *fiber.App) {
	app.Get("/clabgate/workspace-auth/exchange", workspaceExchange)
	app.Get("/clabgate/workspace-auth/verify", workspaceVerify)
}

func workspaceExchange(c *fiber.Ctx) error {
	config := configs.AppConfig.Session
	cookie, destination, sessionID, _, err := usecases.ExchangeWorkspaceGrant(
		config.WorkspaceSecret,
		c.Query("grant"),
		time.Duration(config.WorkspaceCookieTTL)*time.Second,
	)
	if err != nil {
		return c.Status(http.StatusUnauthorized).SendString("invalid or expired workspace grant")
	}
	maxAge := config.WorkspaceCookieTTL
	if maxAge > 86400 {
		maxAge = 86400
	}
	c.Cookie(&fiber.Cookie{
		Name: usecases.WorkspaceCookieName, Value: cookie,
		Path:   strings.TrimRight(config.WorkspacePrefix, "/") + "/" + sessionID + "/",
		MaxAge: int(maxAge), HTTPOnly: true, Secure: config.WorkspaceCookieSecure, SameSite: "Lax",
	})
	return c.Redirect(destination, http.StatusSeeOther)
}

func workspaceVerify(c *fiber.Ctx) error {
	identity, err := usecases.VerifyWorkspaceCookie(
		configs.AppConfig.Session.WorkspaceSecret,
		c.Cookies(usecases.WorkspaceCookieName),
		c.Get("X-Original-URI"),
	)
	if err != nil {
		return c.Status(http.StatusUnauthorized).SendString("workspace authorization required")
	}
	encoded, err := usecases.EncodeWorkspaceIdentity(identity)
	if err != nil {
		return c.Status(http.StatusInternalServerError).SendString("workspace identity unavailable")
	}
	c.Set("X-CMS-Identity", encoded)
	return c.SendStatus(http.StatusNoContent)
}
