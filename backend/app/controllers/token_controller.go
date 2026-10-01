package controllers

import (
	"time"

	"github.com/cms-lab-core/cms-labs-api/backend/app/di"
	"github.com/cms-lab-core/cms-labs-api/backend/app/usecases/auth"
	"github.com/cms-lab-core/cms-labs-api/shared/cms_client"
	"github.com/cms-lab-core/cms-labs-api/shared/jsonrpc"
	"github.com/cms-lab-core/cms-labs-api/shared/logs"
	fiber "github.com/gofiber/fiber/v2"
)

// UserTokenRefresh method for renew access and refresh tokens.
// @Description Renew access and refresh tokens.
// @Summary renew access and refresh tokens
// @Tags user
// @Accept json
// @Produce json
// @Param form body auth.RenewManagerRefreshRequest true "renew token form info"
// @Success 200 {object} auth.SwaggerSSOTokenResponse
// @Router /api/v1/rpc/user.token_refresh [post]
func UserTokenRefresh(c *jsonrpc.Ctx) (interface{}, error) {
	issuer, err := auth.IssuerURLByBaseUrl(c.FiberCtx.BaseURL())
	if err != nil {
		return nil, err
	}
	diLoggerConf := logs.NewZeroLoggerConf(c)
	refreshToken := c.FiberCtx.Cookies(cms_client.SSORefreshTokenName, "")
	if refreshToken == "" {
		dto := auth.RenewManagerInputDTO{}
		err := jsonrpc.ValidatorBase(c, &dto)
		if err != nil {
			return nil, err
		}
		refreshToken = dto.RefreshToken
	}
	container, err := di.NewDIContainer(diLoggerConf)
	if err != nil {
		return nil, err
	}
	defer container.Close()
	uc := container.SSOTokenUC()
	token, err := uc.SetContext(issuer).Execute(
		auth.SSOTokenInputDTO{
			GrantType:    auth.TokenGrantTypeRefreshToken,
			RefreshToken: refreshToken,
		},
	)
	if err != nil {
		return nil, err
	}
	c.FiberCtx.Cookie(&fiber.Cookie{
		Name:     cms_client.SSORefreshTokenName,
		Value:    token.RefreshToken,
		Path:     "/",
		Expires:  auth.ExpiresRefreshCookie(),
		SameSite: fiber.CookieSameSiteNoneMode,
		Secure:   true,
	})
	return token, nil
}

// UserLogin method for grant access by email and password
// @Description Login by email and password.
// @Summary login by email and password
// @Tags user
// @Accept json
// @Produce json
// @Param object body auth.RenewManagerCredentialsRequest true "credentials form info"
// @Success 200 {object} auth.SwaggerSSOTokenResponse
// @Router /api/v1/rpc/user.login [post]
func UserLogin(c *jsonrpc.Ctx) (interface{}, error) {
	issuer, err := auth.IssuerURLByBaseUrl(c.FiberCtx.BaseURL())
	if err != nil {
		return nil, err
	}
	diLoggerConf := logs.NewZeroLoggerConf(c)
	dto := auth.RenewManagerCredentialsInputDTO{}
	err = jsonrpc.ValidatorBase(c, &dto)
	if err != nil {
		return nil, err
	}
	container, err := di.NewDIContainer(diLoggerConf)
	if err != nil {
		return nil, err
	}
	defer container.Close()
	uc := container.UserLoginUC()
	token, err := uc.SetContext(issuer).Execute(dto)
	if err != nil {
		return nil, err
	}
	c.FiberCtx.Cookie(&fiber.Cookie{
		Name:     cms_client.SSORefreshTokenName,
		Value:    token.RefreshToken,
		Path:     "/",
		Expires:  auth.ExpiresRefreshCookie(),
		SameSite: fiber.CookieSameSiteNoneMode,
		Secure:   true,
	})
	return token, nil
}

// UserLogout method for remove access and refresh token
// @Description Logout.
// @Summary logout
// @Tags user
// @Accept json
// @Produce json
// @Param object body auth.UserLogoutRequest true "logout"
// @Success 200 {object} auth.SwaggerSSOTokenResponse
// @Router /api/v1/rpc/user.logout [post]
func UserLogout(c *jsonrpc.Ctx) (interface{}, error) {
	c.FiberCtx.Cookie(&fiber.Cookie{
		Name:     cms_client.SSORefreshTokenName,
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(1, 0),
		MaxAge:   -1,
		SameSite: fiber.CookieSameSiteNoneMode,
		Secure:   true,
	})
	return true, nil
}

// UserPasswordChange method for change password
// @Description Change password.
// @Summary change password
// @Tags user
// @Accept json
// @Produce json
// @Param object body auth.UserPasswordChangeRequest true "credentials form info"
// @Success 200 {object} auth.UserPasswordChangeResponse
// @Router /api/v1/rpc/user.password_change [post]
func UserPasswordChange(c *jsonrpc.Ctx) (interface{}, error) {
	diLoggerConf := logs.NewZeroLoggerConf(c)
	claims, err := auth.ExtractTokenMetadata(c, []string{})
	if err != nil {
		return nil, err
	}
	dto := auth.UserPasswordChangeInputDTO{}
	if err := jsonrpc.ValidatorBase(c, &dto); err != nil {
		return nil, err
	}
	container, err := di.NewDIContainer(diLoggerConf)
	if err != nil {
		return nil, err
	}
	defer container.Close()
	uc := container.UserPasswordRecoverUC()
	response, err := uc.SetContext(claims).Execute(dto)
	return response, err
}
