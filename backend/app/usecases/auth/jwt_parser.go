package auth

import (
	"strings"

	"github.com/cms-lab-core/cms-labs-api/shared/cms_client"

	"github.com/cms-lab-core/cms-labs-api/backend/pkg/configs"
	"github.com/cms-lab-core/cms-labs-api/shared/jsonrpc"

	fiber "github.com/gofiber/fiber/v2"
	jwt "github.com/golang-jwt/jwt/v5"
)

// ExtractTokenMetadata func to extract metadata from JWT.
func ExtractTokenMetadata(
	c *jsonrpc.Ctx,
	roles []string,
) (*cms_client.SSOTokenPublicData, error) {
	token, err := verifyToken(extractToken(c))
	if err != nil {
		return nil, jsonrpc.NewRpcError("unauthorized", "unauthorized")
	}
	if !token.Valid {
		return nil, jsonrpc.NewRpcError("unauthorized", "invalid token")
	}
	tokenData, err := cms_client.SSODecodeToken(token)
	if err != nil {
		return nil, jsonrpc.NewRpcError("unauthorized", err.Error())
	}
	if len(roles) != 0 && !cms_client.SSOHasIntersection(roles, tokenData.Roles) {
		return nil, jsonrpc.NewRpcError("forbidden", "user with current role is not allow action")
	}
	return tokenData, nil
}

func extractToken(c *jsonrpc.Ctx) string {
	bearToken := c.FiberCtx.Get("Authorization")

	// Normally Authorization HTTP header.
	onlyToken := strings.Split(bearToken, " ")
	if len(onlyToken) == 2 {
		return onlyToken[1]
	}

	return ""
}

func verifyToken(tokenString string) (*jwt.Token, error) {
	publicKey := configs.AppConfig.JWT.AccessKey.PublicKey
	if publicKey == nil {
		return nil, fiber.NewError(fiber.StatusInternalServerError, "JWT public key not configured")
	}
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Проверяем алгоритм подписи
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fiber.NewError(fiber.StatusUnauthorized, "Unexpected signing method")
		}
		return publicKey, nil
	})

	if err != nil {
		return nil, fiber.NewError(fiber.StatusUnauthorized, "Invalid token: "+err.Error())
	}

	if !token.Valid {
		return nil, fiber.NewError(fiber.StatusUnauthorized, "Invalid token")
	}

	return token, nil
}
