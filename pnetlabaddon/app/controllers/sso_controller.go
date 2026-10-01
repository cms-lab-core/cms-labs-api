package controllers

import (
	"strings"

	"github.com/cms-lab-core/cms-labs-api/pnetlabaddon/app/di"
	"github.com/cms-lab-core/cms-labs-api/pnetlabaddon/app/usecases"
	"github.com/cms-lab-core/cms-labs-api/shared/cms_client"
	"github.com/cms-lab-core/cms-labs-api/shared/jsonrpc"
	"github.com/cms-lab-core/cms-labs-api/shared/logs"
	fiber "github.com/gofiber/fiber/v2"
)

// SSOFirstFactor Переадресация пользователя на сервер аутентификации.
// @Description Отправка пользователя с параметрами на сервер аутентификации.
// @Summary Отправка пользователя с параметрами на сервер аутентификации.
// @Tags SSO
// @Accept json
// @Produce json
// @Param extra query string false "Дополнительные параметры в base64"
// @Param path query string false "Путь с которого выполнился редирект"
// @Success 307
// @Router /pnet-lab-addon/api/v1/sso/login [get]
func SSOFirstFactor(ctx *fiber.Ctx) error {
	c := &jsonrpc.Ctx{FiberCtx: ctx}
	diLoggerConf := logs.NewZeroLoggerConf(c)
	extra := ctx.Query("extra", "")
	path := ctx.Query("path", "/")
	container, err := di.NewDIContainer(diLoggerConf)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err})
	}
	defer container.Close()

	client := container.CMSClient()

	newURL := client.SSOAuthorizeURI(
		ctx.BaseURL()+"/pnet-lab-addon/api/v1/sso/openid",
		"default",
		path,
		extra,
	)
	// Удаляем cookies все
	ctx.ClearCookie()
	return ctx.Redirect(newURL, fiber.StatusTemporaryRedirect)
}

// SSOSecondFactor Получение токена пользователя второй фактор OpenID.
// @Description Дополнительные параметры:
// Если не пришёл access, refresh токен - ошибка при выдаче доступа.
// Проставление cookies происходит автоматически по запросу.
// @Summary Второй фактор openid.
// @Tags SSO
// @Accept json
// @Produce json
// @Param state query string false "ID запроса по всему пути аутентификации"
// @Param path query string false "Оригинальный путь по всему пути аутентификации"
// @Param code query string false "Код OTP обмена авторизации"
// @Param application query string false "ClientID запрашиваемого из аутентификации"
// @Param extra query string false "Дополнительные данные с backend base64 по всему пути аутентификации"
// @Success 307
// @Router /pnet-lab-addon/api/v1/sso/openid [get]
func SSOSecondFactor(ctx *fiber.Ctx) error {
	c := &jsonrpc.Ctx{FiberCtx: ctx}
	diLoggerConf := logs.NewZeroLoggerConf(c)
	container, err := di.NewDIContainer(diLoggerConf)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err})
	}
	defer container.Close()
	// Второй фактор аутентификации
	uc := container.SSOSecondFactorUC()
	dto := usecases.SSOSecondFactorInputDTO{
		Code:        ctx.Query("code", ""),
		Application: ctx.Query("application", ""),
		Path:        ctx.Query("path", ""),
		State:       ctx.Query("state", ""),
		RedirectURI: ctx.BaseURL() + "/pnet-lab-addon/api/v1/sso/openid",
	}
	output, err := uc.Execute(dto)
	if err != nil {
		return err
	}
	// Создание лабораторной работы по extraArgs
	uc2 := container.LabCreateUC()
	err = uc2.Execute(
		usecases.LabCreateInputDTO{
			Extra:   ctx.Query("extra", ""),
			UserPod: output.UserPod,
		},
	)
	if err != nil {
		return err
	}
	// Получаем hostname без порта
	host := ctx.Hostname()
	if strings.Contains(host, ":") {
		host = strings.Split(host, ":")[0]
	}
	// Выполняем редирект
	ctx.Cookie(&fiber.Cookie{
		Name:     "token",
		Value:    output.CookieToken,
		Path:     "/",
		MaxAge:   output.CookieMaxAge,
		Expires:  output.CookieAge,
		HTTPOnly: true,
		Domain:   host,
	})
	ctx.Cookie(&fiber.Cookie{
		Name:     cms_client.SSORefreshTokenName,
		Value:    output.RefreshToken,
		MaxAge:   output.RefreshTokenMaxAge,
		Path:     "/",
		Expires:  output.RefreshTokenAge,
		HTTPOnly: true,
	})
	return ctx.Redirect("/", fiber.StatusTemporaryRedirect)
}
