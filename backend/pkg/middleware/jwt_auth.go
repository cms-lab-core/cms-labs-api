package middleware

import (
	"github.com/cms-lab-core/cms-labs-api/backend/app/usecases/auth"
	"github.com/cms-lab-core/cms-labs-api/shared/jsonrpc"
	fiber "github.com/gofiber/fiber/v2"
)

// NewJWTMiddleware создаёт middleware для извлечения и проверки JWT токена без ошибок
func NewJWTMiddleware() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		c := &jsonrpc.Ctx{FiberCtx: ctx}
		// Извлекаем и парсим JWT токен
		tokenData, _ := auth.ExtractTokenMetadata(c, []string{})
		// Сохраняем распарсенные данные токена в locals
		if tokenData != nil {
			c.FiberCtx.Locals("jwtUser", tokenData)
			c.FiberCtx.Locals("x-user-email", tokenData.Email)
			c.FiberCtx.Locals("x-user-role", tokenData.UserRoleMain())
		}
		// Переходим к следующему middleware или обработчику
		return c.FiberCtx.Next()
	}
}
