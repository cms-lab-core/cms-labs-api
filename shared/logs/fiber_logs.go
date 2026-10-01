// Package logs общие компоненты для Fiber логирования
package logs

import (
	"strings"

	"github.com/cms-lab-core/cms-labs-api/shared/jsonrpc"
	"github.com/gofiber/contrib/fiberzerolog"
	fiber "github.com/gofiber/fiber/v2"
)

func NewFiberZerologLogger() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		c := &jsonrpc.Ctx{FiberCtx: ctx}
		logger := NewZeroLogger(NewZeroLoggerConf(c))
		return fiberzerolog.New(
			fiberzerolog.Config{
				Logger: logger,
				Next: func(c *fiber.Ctx) bool {
					path := c.Path()
					return strings.HasPrefix(path, "/api/docs") ||
						strings.HasPrefix(path, "/clabgate/api/docs")
				},
			},
		)(c.FiberCtx)
	}
}
