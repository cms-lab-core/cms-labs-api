package cms_client

import (
	"context"
	"fmt"
	"net/url"

	"github.com/google/uuid"
)

func (c *CMSClient) SSOToken(grantType string, redirectUri string, code string, refreshToken string) (*SSOToken, error) {
	const path = "/api/v1/sso/token"
	var ssoToken SSOToken
	response, err := c.client.R().SetFormData(map[string]string{
		"grant_type":    grantType,
		"redirect_uri":  redirectUri,
		"code":          code,
		"refresh_token": refreshToken,
	}).SetBasicAuth(c.Config.ClientID, c.Config.Token).SetResult(&ssoToken).Post(path)
	if response == nil {
		return nil, NewCMSError("", 0)
	}
	if response.IsError() || ssoToken.UserId == "" {
		return nil, NewCMSError("invalid_granted", response.StatusCode())
	}
	return &ssoToken, err
}

func (c *CMSClient) SSOUserInfo(accessToken string) (*SSOTokenPublicData, error) {
	return c.SSOUserInfoContext(context.Background(), accessToken)
}

func (c *CMSClient) SSOUserInfoContext(ctx context.Context, accessToken string) (*SSOTokenPublicData, error) {
	const path = "/api/v1/sso/userinfo"
	var ssoToken SSOTokenPublicData
	response, err := c.client.R().SetContext(ctx).SetHeader(
		"Authorization",
		fmt.Sprintf("Bearer %s", accessToken),
	).SetResult(&ssoToken).Get(path)
	if response == nil {
		return nil, NewCMSError("", 0)
	}
	if response.IsError() || ssoToken.Sub == "" {
		return nil, NewCMSError("invalid_user", response.StatusCode())
	}
	return &ssoToken, err
}

func (c *CMSClient) DownloadUNLFile(
	curlRequestID uint,
	overrides map[string]string,
) ([]byte, error) {
	const path = "/api/v1/curl-request/execute"
	payload := map[string]interface{}{
		"curl_request_id": curlRequestID,
		"override":        overrides,
	}
	response, err := c.client.R().SetBasicAuth(c.Config.ClientID, c.Config.Token).SetBody(payload).Post(path)
	if response == nil {
		return nil, NewCMSError("", 0)
	}
	if response.IsError() {
		return nil, NewCMSError("external error", response.StatusCode())
	}
	return response.Body(), err
}

func (c *CMSClient) SSOAuthorizeURI(
	redirectUri string,
	scope string,
	path string,
	extra string,
) string {
	const (
		cClientId     = "client_id"
		cRedirectUri  = "redirect_uri"
		cResponseType = "response_type"
		cScope        = "scope"
		cPath         = "path"
		cState        = "state"
		cExtra        = "extra"
	)
	var query = make(url.Values)
	state := uuid.New().String()
	if path == "" {
		path = "/"
	}
	query.Set(cClientId, c.Config.ClientID)
	query.Set(cRedirectUri, redirectUri)
	query.Set(cResponseType, "code")
	query.Set(cScope, scope)
	query.Set(cState, state)
	query.Set(cPath, path)
	query.Set(cExtra, extra)
	query.Set(cExtra, extra)

	var uri = url.URL{
		Scheme:   c.BaseURL.Scheme,
		Host:     c.BaseURL.Host,
		Path:     "/api/v1/sso/authorize",
		RawQuery: query.Encode(),
	}
	return uri.String()
}
