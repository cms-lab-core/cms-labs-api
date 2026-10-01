package controllers

import (
	"net/url"

	"github.com/cms-lab-core/cms-labs-api/shared/logs"

	"github.com/cms-lab-core/cms-labs-api/backend/app/di"
	"github.com/cms-lab-core/cms-labs-api/backend/app/usecases/auth"
	"github.com/cms-lab-core/cms-labs-api/shared/jsonrpc"
	fiber "github.com/gofiber/fiber/v2"
)

// SSOAuthorize Получить код авторизации.
// @Description Получение кода авторизации по OpenID.
// @Summary Получение кода авторизации по OpenID.
// @Tags SSO
// @Accept json
// @Produce json
// @Param client_id query string true "Код приложения клиента"
// @Param redirect_uri query string true "Адрес переадресации клиента"
// @Param response_type query string true "Тип ответа. Возможные значения: [code]"  Enums(code)
// @Param scope query string true "Область доступа (значения, разделённые запятой). Возможные значения: [default, openid]"  Enums(default, openid)
// @Param path query string false "Путь, возвращается в параметрах редиректа"
// @Param state query string false "Состояние, возвращается в параметрах редиректа, участвует в генерации кода авторизации"
// @Param extra query string false "Дополнительные параметры вида base64"
// @Success 200 {object} auth.SSOAuthorizeResponse
// @Success 400 {object} auth.SSOError
// @Security ApiKeyAuth
// @Router /api/v1/sso/authorize [get]
func SSOAuthorize(c *fiber.Ctx) error {
	// Получаем QueryParams в виде map
	queryParams := c.Queries()

	// Формируем строку QueryParams
	queryString := "?"
	for key, value := range queryParams {
		// Кодируем ключ и значение, чтобы избежать проблем с символами
		queryString += url.QueryEscape(key) + "=" + url.QueryEscape(value) + "&"
	}

	// Убираем последний символ '&', если он есть
	if len(queryString) > 1 {
		queryString = queryString[:len(queryString)-1]
	}

	// Формируем новый URL
	newURL := "/login" + queryString

	// Выполняем редирект
	return c.Redirect(newURL)
}

// SSOAuthorizePost Получить код авторизации.
// @Description Получение кода авторизации по OpenID.
// @Summary Получение кода авторизации по OpenID.
// @Tags SSO
// @Accept json
// @Produce json
// @Param form body auth.SSOAuthorizeRequest true "renew token form info"
// @Success 200 {object} auth.SSOAuthorizeResponse
// @Security ApiKeyAuth
// @Router /api/v1/rpc/sso.authorize [post]
func SSOAuthorizePost(c *jsonrpc.Ctx) (interface{}, error) {
	diLoggerConf := logs.NewZeroLoggerConf(c)
	claims, err := auth.ExtractTokenMetadata(c, []string{})
	if err != nil {
		return nil, err
	}
	dto := auth.SSOAuthorizeInputDTO{}
	err = jsonrpc.ValidatorBase(c, &dto)
	if err != nil {
		return nil, err
	}
	dto.UserID = claims.UserID()
	container, err := di.NewDIContainer(diLoggerConf)
	if err != nil {
		return nil, err
	}
	defer container.Close()
	uc := container.SSOAuthorizeUC()
	output, err := uc.Execute(dto)
	return output, err
}

// SSOToken Получить токен доступа и токен обновления.
// @Description Получение токена доступа и токена обновления.
// @Summary Получение токена доступа и токена обновления.
// @Tags SSO, EXTERNAL
// @Accept json
// @Produce json
// @Param grant_type formData string true "Тип авторизации [authorization_code, refresh_token]" Enums(authorization_code, refresh_token)
// @Param redirect_uri formData string false "Адрес переадресации клиента"
// @Param code formData string false "Код авторизации"
// @Param refresh_token formData string false "Токен обновления"
// @Param Authorization header string true "Basic-токен, созданный клиентом"
// @Success 200 {object} auth.SwaggerSSOToken
// @Success 400 {object} auth.SSOError
// @Security ApiKeyAuth
// @Router /api/v1/sso/token [post]
func SSOToken(ctx *fiber.Ctx) error {
	c := &jsonrpc.Ctx{FiberCtx: ctx}
	issuer, err := auth.IssuerURLByBaseUrl(c.FiberCtx.BaseURL())
	if err != nil {
		return err
	}
	diLoggerConf := logs.NewZeroLoggerConf(c)
	dto := auth.SSOTokenInputDTO{}
	dto.GrantType = c.FiberCtx.FormValue("grant_type", "")
	dto.RedirectUri = c.FiberCtx.FormValue("redirect_uri", "")
	dto.Code = c.FiberCtx.FormValue("code", "")
	dto.RefreshToken = c.FiberCtx.FormValue("refresh_token", "")
	dto.CodeVerifier = c.FiberCtx.FormValue("code_verifier", "")
	container, err := di.NewDIContainer(diLoggerConf)
	if err != nil {
		return c.FiberCtx.JSON(auth.SSOError{Error: "invalid_request", ErrorDescription: err.Error()})
	}
	defer container.Close()
	uc := container.SSOTokenUC()
	output, err := uc.SetContext(issuer).Execute(dto)
	if err != nil {
		return c.FiberCtx.JSON(auth.SSOError{Error: "invalid_grant", ErrorDescription: err.Error()})
	}
	return c.FiberCtx.JSON(output)
}

// SSOIntrospect Проверка состояния токена.
// @Description Проверка состояния токена.
// @Summary Проверка состояния токена.
// @Tags SSO, EXTERNAL
// @Accept json
// @Produce json
// @Param token formData string true "Токен доступа или токен обновления"
// @Param Authorization header string true "Basic-токен, созданный клиентом"
// @Success 200 {object} auth.SSOTokenIntrospect
// @Success 400 {object} auth.SSOError
// @Security ApiKeyAuth
// @Router /api/v1/sso/introspect [post]
func SSOIntrospect(ctx *fiber.Ctx) error {
	c := &jsonrpc.Ctx{FiberCtx: ctx}
	diLoggerConf := logs.NewZeroLoggerConf(c)
	dto := auth.SSOIntrospectInputDTO{}
	dto.Token = c.FiberCtx.FormValue("token", "")
	container, err := di.NewDIContainer(diLoggerConf)
	if err != nil {
		return c.FiberCtx.JSON(auth.SSOError{Error: "internal_error", ErrorDescription: err.Error()})
	}
	defer container.Close()
	uc := container.SSOIntrospectUC()
	output, err := uc.Execute(dto)
	if err != nil {
		return c.FiberCtx.JSON(auth.SSOError{Error: "invalid_grant", ErrorDescription: err.Error()})
	}
	return c.FiberCtx.JSON(output)
}

// SSOUserInfo Получить информацию о пользователе.
// @Description Получить информацию о пользователе.
// @Summary Получить информацию о пользователе.
// @Tags SSO
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer токен"
// @Success 200 {object} auth.SwaggerSSOTokenPublicData
// @Success 400 {object} auth.SSOError
// @Security ApiKeyAuth
// @Router /api/v1/sso/userinfo [get]
func SSOUserInfo(ctx *fiber.Ctx) error {
	c := &jsonrpc.Ctx{FiberCtx: ctx}
	claims, err := auth.ExtractTokenMetadata(c, []string{})
	if err != nil {
		return c.FiberCtx.JSON(auth.SSOError{Error: `invalid_grant`, ErrorDescription: err.Error()})
	}
	return c.FiberCtx.JSON(claims)
}

// SSOOpenIdConfiguration Получить информацию о спецификации.
// @Description Получить информацию о спецификации.
// @Summary Получить информацию о спецификации.
// @Tags SSO
// @Accept json
// @Produce json
// @Success 200 {object} auth.SSOAuthorizeResponse
// @Success 400 {object} auth.SSOError
// @Router /api/v1/sso/.well-known/openid-configuration [get]
func SSOOpenIdConfiguration(ctx *fiber.Ctx) error {
	c := &jsonrpc.Ctx{FiberCtx: ctx}
	diLoggerConf := logs.NewZeroLoggerConf(c)
	dto := auth.SSOOpenidConfigurationInputDTO{
		BaseURL: c.FiberCtx.BaseURL(),
	}
	container, err := di.NewDIContainer(diLoggerConf)
	if err != nil {
		return c.FiberCtx.JSON(auth.SSOError{Error: "internal_server_error", ErrorDescription: err.Error()})
	}
	defer container.Close()
	uc := container.SSOOpenidConfigurationUC()
	output, err := uc.Execute(dto)
	if err != nil {
		return c.FiberCtx.JSON(auth.SSOError{Error: `internal_server_error`, ErrorDescription: err.Error()})
	}
	return c.FiberCtx.JSON(output)
}

// SSOJwks Получить информацию о публичных ключах для подписи JWT токенов.
// @Description Получить информацию о публичных ключах для подписи JWT токенов.
// @Summary Получить информацию о публичных ключах для подписи JWT токенов.
// @Tags SSO
// @Accept json
// @Produce json
// @Success 200 {object} auth.SSOJWKSOutputDTO
// @Success 500 {object} auth.SSOError
// @Router /api/v1/sso/jwks [get]
func SSOJwks(ctx *fiber.Ctx) error {
	c := &jsonrpc.Ctx{FiberCtx: ctx}
	diLoggerConf := logs.NewZeroLoggerConf(c)
	container, err := di.NewDIContainer(diLoggerConf)
	if err != nil {
		return c.FiberCtx.JSON(auth.SSOError{Error: "internal_server_error", ErrorDescription: err.Error()})
	}
	defer container.Close()
	uc := container.SSOJwksUC()
	output, err := uc.Execute()
	if err != nil {
		return c.FiberCtx.JSON(auth.SSOError{Error: `internal_server_error`, ErrorDescription: err.Error()})
	}
	return c.FiberCtx.JSON(output)
}
