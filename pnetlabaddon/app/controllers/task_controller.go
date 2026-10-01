package controllers

import (
	"github.com/cms-lab-core/cms-labs-api/pnetlabaddon/app/di"
	"github.com/cms-lab-core/cms-labs-api/shared/jsonrpc"
	"github.com/cms-lab-core/cms-labs-api/shared/logs"
	fiber "github.com/gofiber/fiber/v2"
)

// PNETServerPing Пинг в core-backend для синхронизации попыток.
// @Description Пинг в core-backend для синхронизации попыток.
// @Summary Пинг в core-backend для синхронизации попыток.
// @Tags Task
// @Accept json
// @Produce json
// @Success 200
// @Router /pnet-lab-addon/api/v1/task/pnet-server-ping [post]
func PNETServerPing(ctx *fiber.Ctx) error {
	c := &jsonrpc.Ctx{FiberCtx: ctx}
	diLoggerConf := logs.NewZeroLoggerConf(c)
	container, err := di.NewDIContainer(diLoggerConf)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err})
	}
	defer container.Close()

	client := container.PnetServerPingUC()
	err = client.Execute()

	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err})
	}
	return ctx.Status(fiber.StatusNoContent).JSON(fiber.Map{"result": true})
}
