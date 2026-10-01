// Package logs общие компоненты для Fiber логирования
package logs

import (
	"os"

	"github.com/cms-lab-core/cms-labs-api/shared/jsonrpc"

	"github.com/rs/zerolog"
)

type ZeroLoggerConf struct {
	Name      string `json:"logger"`
	RequestID string `json:"x-request-id"`
	UserEmail string `json:"x-user-email"`
	UserRole  string `json:"x-user-role"`
	ServiceID string `json:"x-service-id"`
}

// SetName для ZeroLoggerConf возвращает копию с измененным именем
func (z *ZeroLoggerConf) SetName(name string) *ZeroLoggerConf {
	return &ZeroLoggerConf{
		Name:      name,
		RequestID: z.RequestID,
		UserEmail: z.UserEmail,
		UserRole:  z.UserRole,
		ServiceID: z.ServiceID,
	}
}

// FiberLocalsParseDefault парсит из fiber.Ctx переменные с приведением типов
func FiberLocalsParseDefault(c *jsonrpc.Ctx, key string) string {
	if c == nil {
		return ""
	}
	if value, ok := c.FiberCtx.Locals(key).(string); ok {
		return value
	}
	return ""
}

// NewZeroLoggerConf создаёт контекст для zerolog.Logger
func NewZeroLoggerConf(c *jsonrpc.Ctx) *ZeroLoggerConf {
	return &ZeroLoggerConf{
		Name:      "main",
		RequestID: FiberLocalsParseDefault(c, "x-request-id"),
		UserEmail: FiberLocalsParseDefault(c, "x-user-email"),
		UserRole:  FiberLocalsParseDefault(c, "x-user-role"),
		ServiceID: FiberLocalsParseDefault(c, "x-service-id"),
	}
}

func NewZeroLogger(config *ZeroLoggerConf) *zerolog.Logger {
	logger := zerolog.New(os.Stderr).With().Timestamp().Logger()
	logger.UpdateContext(func(c zerolog.Context) zerolog.Context {
		return c.
			Str("logger", config.Name).
			Str("x-request-id", config.RequestID).
			Str("x-user-email", config.UserEmail).
			Str("x-user-role", config.UserRole).
			Str("x-service-id", config.ServiceID)
	})
	return &logger
}

func ZeroLogInit(debug bool) {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}
}
