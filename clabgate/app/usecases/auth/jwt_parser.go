package auth

import (
	"context"
	"strings"

	resty "github.com/go-resty/resty/v2"
	"github.com/maintainer64/cms-labs-api/clabgate/pkg/configs"
	"github.com/maintainer64/cms-labs-api/shared/cms_client"
	"github.com/maintainer64/cms-labs-api/shared/jsonrpc"
)

type userInfoProvider interface {
	SSOUserInfoContext(context.Context, string) (*cms_client.SSOTokenPublicData, error)
}

// ExtractTokenMetadata validates the bearer token against CMS and returns the
// authoritative user profile. Clabgate intentionally does not keep CMS signing
// keys or duplicate token validation policy.
func ExtractTokenMetadata(
	c *jsonrpc.Ctx,
	roles []string,
) (*cms_client.SSOTokenPublicData, error) {
	cmsConfig := configs.AppConfig.CMS
	if cmsConfig.BaseURL == "" {
		return nil, jsonrpc.NewRpcError("unauthorized", "CMS userinfo endpoint is not configured")
	}
	client := cms_client.NewCMSClient(&cms_client.CMSClientConfig{
		Debug:             configs.AppConfig.Debug,
		MaxTimeoutSeconds: cmsConfig.MaxTimeoutSeconds,
		ClientID:          cmsConfig.ClientID,
		Token:             cmsConfig.Token,
		BaseUrl:           cmsConfig.BaseURL,
	}, resty.New())
	return extractTokenMetadata(c, roles, client)
}

func extractTokenMetadata(
	c *jsonrpc.Ctx,
	roles []string,
	client userInfoProvider,
) (*cms_client.SSOTokenPublicData, error) {
	accessToken := extractToken(c)
	if accessToken == "" {
		return nil, jsonrpc.NewRpcError("unauthorized", "bearer token is required")
	}
	tokenData, err := client.SSOUserInfoContext(c.FiberCtx.UserContext(), accessToken)
	if err != nil {
		return nil, jsonrpc.NewRpcError("unauthorized", "unauthorized")
	}
	if len(roles) != 0 && !cms_client.SSOHasIntersection(roles, tokenData.Roles) {
		return nil, jsonrpc.NewRpcError("forbidden", "user with current role is not allow action")
	}
	return tokenData, nil
}

func extractToken(c *jsonrpc.Ctx) string {
	parts := strings.Fields(c.FiberCtx.Get("Authorization"))
	if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
		return parts[1]
	}
	return ""
}
