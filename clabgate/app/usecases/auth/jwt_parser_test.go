package auth

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cms-lab-core/cms-labs-api/shared/cms_client"
	"github.com/cms-lab-core/cms-labs-api/shared/jsonrpc"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeUserInfoProvider struct {
	user        *cms_client.SSOTokenPublicData
	err         error
	accessToken string
}

func (f *fakeUserInfoProvider) SSOUserInfoContext(
	_ context.Context,
	accessToken string,
) (*cms_client.SSOTokenPublicData, error) {
	f.accessToken = accessToken
	return f.user, f.err
}

func TestExtractTokenMetadataUsesCMSUserInfo(t *testing.T) {
	provider := &fakeUserInfoProvider{user: &cms_client.SSOTokenPublicData{
		Sub: "42", Email: "student@example.com", Roles: []string{cms_client.SSOUsersRoleStudent},
	}}

	user, err := extractThroughFiber(t, "Bearer cms-access-token", []string{cms_client.SSOUsersRoleStudent}, provider)

	require.NoError(t, err)
	require.NotNil(t, user)
	assert.Equal(t, "cms-access-token", provider.accessToken)
	assert.Equal(t, "42", user.Sub)
	assert.Equal(t, "student@example.com", user.Email)
}

func TestExtractTokenMetadataRejectsMissingBearerToken(t *testing.T) {
	provider := &fakeUserInfoProvider{err: errors.New("must not be called")}

	user, err := extractThroughFiber(t, "", nil, provider)

	require.Error(t, err)
	assert.Nil(t, user)
	assert.Empty(t, provider.accessToken)
}

func TestExtractTokenMetadataRejectsCMSFailure(t *testing.T) {
	provider := &fakeUserInfoProvider{err: errors.New("CMS rejected token")}

	user, err := extractThroughFiber(t, "Bearer invalid", nil, provider)

	require.Error(t, err)
	assert.Nil(t, user)
}

func TestExtractTokenMetadataEnforcesRolesFromCMS(t *testing.T) {
	provider := &fakeUserInfoProvider{user: &cms_client.SSOTokenPublicData{
		Sub: "42", Roles: []string{cms_client.SSOUsersRoleStudent},
	}}

	user, err := extractThroughFiber(t, "Bearer cms-access-token", []string{cms_client.SSOUsersRoleAdmin}, provider)

	require.Error(t, err)
	assert.Nil(t, user)
}

func extractThroughFiber(
	t *testing.T,
	authorization string,
	roles []string,
	provider userInfoProvider,
) (*cms_client.SSOTokenPublicData, error) {
	t.Helper()
	var user *cms_client.SSOTokenPublicData
	var authErr error
	app := fiber.New()
	app.Get("/", func(ctx *fiber.Ctx) error {
		user, authErr = extractTokenMetadata(&jsonrpc.Ctx{FiberCtx: ctx}, roles, provider)
		return ctx.SendStatus(http.StatusNoContent)
	})
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	if authorization != "" {
		request.Header.Set("Authorization", authorization)
	}
	response, err := app.Test(request)
	require.NoError(t, err)
	require.NoError(t, response.Body.Close())
	return user, authErr
}
