package routes

import (
	"fmt"
	"strings"
	"testing"

	"github.com/cms-lab-core/cms-labs-api/backend/app/models"
	"github.com/cms-lab-core/cms-labs-api/backend/app/usecases/auth"
	"github.com/cms-lab-core/cms-labs-api/shared/cms_client"
	json "github.com/goccy/go-json"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestSSOAuthorize(t *testing.T) {
	description := "successful authorization code generation"
	f := NewTestHTTP()
	defer f.Close()

	_, clientID := f.AuthorizationServiceBasic()

	// Создаем тестового пользователя
	user := models.User{}
	user.Email = "test" + uuid.New().String() + "@example.com"
	user.Name = "Name " + uuid.New().String()
	f.DB.Create(&user)
	authHeader := f.AuthorizationUser(user.ID, nil)

	role := models.Role{}
	role.Code = "student" + uuid.New().String()
	role.Name = "Student"
	f.DB.Create(&role)

	roleUser := models.RoleRelation{}
	roleUser.RoleID = role.ID
	roleUser.UserID = &user.ID

	f.DB.Create(&roleUser)

	input := auth.SSOAuthorizeInputDTO{
		ClientID:     clientID,
		RedirectUri:  "https://localhost/callback",
		ResponseType: "code",
		Scope:        "openid",
		UserID:       user.ID,
	}

	expectedCode := 200
	statusCode, body := f.Rpc(&TestRpcRequest{
		Method:        "sso.authorize",
		Params:        input,
		Authorization: authHeader,
	})

	response := auth.SSOAuthorizeResponse{}
	_ = json.Unmarshal([]byte(body), &response)

	assert.Equal(t, expectedCode, statusCode, description)
	assert.NotEmpty(t, response.Result.Code, description)
	assert.Equal(t, input.RedirectUri, response.Result.RedirectUri, description)
	assert.Equal(t, input.ClientID, response.Result.Application, description)
}

func TestSSOAuthorizeInvalidClient(t *testing.T) {
	description := "authorization with invalid client"
	f := NewTestHTTP()
	defer f.Close()

	// Создаем тестового пользователя
	user := models.User{}
	user.Email = "test" + uuid.New().String() + "@example.com"
	user.Name = "Name " + uuid.New().String()
	f.DB.Create(&user)
	role := models.Role{}
	role.Code = "student" + uuid.New().String()
	role.Name = "Student"
	f.DB.Create(&role)

	roleUser := models.RoleRelation{}
	roleUser.RoleID = role.ID
	roleUser.UserID = &user.ID

	f.DB.Create(&roleUser)
	authHeader := f.AuthorizationUser(user.ID, nil)

	input := auth.SSOAuthorizeInputDTO{
		ClientID:     "invalid-client",
		RedirectUri:  "https://invalid.example.com/callback",
		ResponseType: "code",
		Scope:        "openid",
		UserID:       user.ID,
	}

	expectedCode := 500
	statusCode, body := f.Rpc(&TestRpcRequest{
		Method:        "sso.authorize",
		Params:        input,
		Authorization: authHeader,
	})

	assert.Equal(t, expectedCode, statusCode, description)
	assert.Contains(t, body, "server_not_found", description)
}

func TestSSOAuthorizeNotMatchRoles(t *testing.T) {
	description := "authorization with not match roles"
	f := NewTestHTTP()
	defer f.Close()

	// OpenID server
	server := models.Server{}
	server.Type = models.ServerTypeOpenID
	server.Name = uuid.New().String() + "_server"
	server.Url = "https://localhost"
	server.IsActive = true
	server.Token = uuid.New().String()
	server.ClientID = uuid.New().String()
	f.DB.Create(&server)

	customRole := models.Role{}
	customRole.Code = "custom_role_" + uuid.New().String()
	customRole.Name = "CustomRole"
	f.DB.Create(&customRole)

	serverRole := models.RoleRelation{}
	serverRole.RoleID = customRole.ID
	serverRole.ServerID = &server.ID
	f.DB.Create(&serverRole)

	// Создаем тестового пользователя
	user := models.User{}
	user.Email = "test" + uuid.New().String() + "@example.com"
	user.Name = "Name " + uuid.New().String()
	f.DB.Create(&user)
	role := models.Role{}
	role.Code = "student" + uuid.New().String()
	role.Name = "Student"
	f.DB.Create(&role)

	roleUser := models.RoleRelation{}
	roleUser.RoleID = role.ID
	roleUser.UserID = &user.ID

	f.DB.Create(&roleUser)
	authHeader := f.AuthorizationUser(user.ID, nil)

	input := auth.SSOAuthorizeInputDTO{
		ClientID:     server.ClientID,
		RedirectUri:  "https://localhost",
		ResponseType: "code",
		Scope:        "openid",
		UserID:       user.ID,
	}

	expectedCode := 500
	statusCode, body := f.Rpc(&TestRpcRequest{
		Method:        "sso.authorize",
		Params:        input,
		Authorization: authHeader,
	})

	assert.Equal(t, expectedCode, statusCode, description)
	assert.Contains(t, body, "not_access_sso_with_current_role", description)
}

func TestSSOAuthorizeMatchRoles(t *testing.T) {
	description := "authorization with match roles"
	f := NewTestHTTP()
	defer f.Close()

	// OpenID server
	server := models.Server{}
	server.Type = models.ServerTypeOpenID
	server.Name = uuid.New().String() + "_server"
	server.Url = "https://localhost"
	server.IsActive = true
	server.Token = uuid.New().String()
	server.ClientID = uuid.New().String()
	f.DB.Create(&server)

	customRole := models.Role{}
	customRole.Code = "custom_role_" + uuid.New().String()
	customRole.Name = "CustomRole"
	f.DB.Create(&customRole)

	serverRole := models.RoleRelation{}
	serverRole.RoleID = customRole.ID
	serverRole.ServerID = &server.ID
	f.DB.Create(&serverRole)

	// Создаем тестового пользователя
	user := models.User{}
	user.Email = "test" + uuid.New().String() + "@example.com"
	user.Name = "Name " + uuid.New().String()
	f.DB.Create(&user)

	roleUser := models.RoleRelation{}
	roleUser.RoleID = customRole.ID
	roleUser.UserID = &user.ID

	f.DB.Create(&roleUser)
	authHeader := f.AuthorizationUser(user.ID, nil)

	input := auth.SSOAuthorizeInputDTO{
		ClientID:     server.ClientID,
		RedirectUri:  "https://localhost",
		ResponseType: "code",
		Scope:        "openid",
		UserID:       user.ID,
	}

	expectedCode := 200
	statusCode, body := f.Rpc(&TestRpcRequest{
		Method:        "sso.authorize",
		Params:        input,
		Authorization: authHeader,
	})

	response := auth.SSOAuthorizeResponse{}
	_ = json.Unmarshal([]byte(body), &response)

	assert.Equal(t, expectedCode, statusCode, description)
	assert.NotEmpty(t, response.Result.Code, description)
	assert.Equal(t, input.RedirectUri, response.Result.RedirectUri, description)
	assert.Equal(t, input.ClientID, response.Result.Application, description)
}

func TestSSOTokenByAuthCode(t *testing.T) {
	description := "get tokens by authorization code"
	f := NewTestHTTP()
	defer f.Close()

	authHeaderServer, clientID := f.AuthorizationServiceBasic()
	server := models.Server{}
	f.DB.Where("client_id = ?", clientID).Find(&server)

	// Создаем тестового пользователя
	user := models.User{}
	user.Email = "test" + uuid.New().String() + "@example.com"
	user.Name = "Name " + uuid.New().String()
	f.DB.Create(&user)

	role := models.Role{}
	role.Code = "student" + uuid.New().String()
	role.Name = "Student"
	f.DB.Create(&role)

	roleUser := models.RoleRelation{}
	roleUser.RoleID = role.ID
	roleUser.UserID = &user.ID

	f.DB.Create(&roleUser)

	// Создаем authorization code
	attempt := models.TokenAttempt{}
	attempt.UserID = &user.ID
	attempt.ServerID = &server.ID
	attempt.State = uuid.New().String()
	attempt.AuthorizationCode = uuid.New().String()
	f.DB.Create(&attempt)

	input := map[string]string{
		"grant_type":   "authorization_code",
		"code":         attempt.AuthorizationCode,
		"redirect_uri": server.Url + "/callback",
	}

	expectedCode := 200
	statusCode, body := f.Request(&TestHttpRequest{
		Method:        "POST",
		Route:         "/api/v1/sso/token",
		Body:          FiberRequestFormPayload(input),
		Authorization: authHeaderServer,
		ContentType:   "application/x-www-form-urlencoded",
	})

	response := auth.SwaggerSSOToken{}
	_ = json.Unmarshal([]byte(body), &response)

	assert.Equal(t, expectedCode, statusCode, description)
	assert.NotEmpty(t, response.AccessToken, description)
	assert.NotEmpty(t, response.IdToken, description)
	assert.NotEmpty(t, response.RefreshToken, description)
	assert.Equal(t, "bearer", response.TokenType, description)
	assert.Equal(t, fmt.Sprintf("%d", user.ID), response.UserId, description)
	assert.Greater(t, response.ExpiresIn, int64(0), description)
}

func TestSSOAuthorizeWithCodeChallenge(t *testing.T) {
	description := "successful authorization with PKCE code_challenge"
	f := NewTestHTTP()
	defer f.Close()

	_, clientID := f.AuthorizationServiceBasic()

	user := models.User{}
	user.Email = "test" + uuid.New().String() + "@example.com"
	user.Name = "Name " + uuid.New().String()
	f.DB.Create(&user)
	authHeader := f.AuthorizationUser(user.ID, nil)

	role := models.Role{}
	role.Code = "student" + uuid.New().String()
	role.Name = "Student"
	f.DB.Create(&role)

	roleUser := models.RoleRelation{}
	roleUser.RoleID = role.ID
	roleUser.UserID = &user.ID

	f.DB.Create(&roleUser)

	codeChallenge := "dFYxeuLgr2mq8IeJT4C1WqoAH9A7s9e3QzZ9LqK5rXw"
	codeChallengeMethod := "S256"

	input := auth.SSOAuthorizeInputDTO{
		ClientID:            clientID,
		RedirectUri:         "https://localhost/callback",
		ResponseType:        "code",
		Scope:               "openid",
		UserID:              user.ID,
		CodeChallenge:       codeChallenge,
		CodeChallengeMethod: codeChallengeMethod,
	}

	expectedCode := 200
	statusCode, body := f.Rpc(&TestRpcRequest{
		Method:        "sso.authorize",
		Params:        input,
		Authorization: authHeader,
	})

	response := auth.SSOAuthorizeResponse{}
	_ = json.Unmarshal([]byte(body), &response)

	assert.Equal(t, expectedCode, statusCode, description)
	assert.NotEmpty(t, response.Result.Code, description)

	var attempt models.TokenAttempt
	f.DB.Where("authorization_code = ?", response.Result.Code).First(&attempt)
	assert.Equal(t, codeChallenge, *attempt.CodeChallenge, description)
	assert.Equal(t, codeChallengeMethod, *attempt.CodeChallengeMethod, description)
}

func TestSSOTokenByAuthCodeWithPKCE(t *testing.T) {
	description := "get tokens by authorization code with valid PKCE code_verifier"
	f := NewTestHTTP()
	defer f.Close()

	authHeaderServer, clientID := f.AuthorizationServiceBasic()
	server := models.Server{}
	f.DB.Where("client_id = ?", clientID).Find(&server)

	user := models.User{}
	user.Email = "test" + uuid.New().String() + "@example.com"
	user.Name = "Name " + uuid.New().String()
	f.DB.Create(&user)

	role := models.Role{}
	role.Code = "student" + uuid.New().String()
	role.Name = "Student"
	f.DB.Create(&role)

	roleUser := models.RoleRelation{}
	roleUser.RoleID = role.ID
	roleUser.UserID = &user.ID

	f.DB.Create(&roleUser)

	codeVerifier := "dFYxeuLgr2mq8IeJT4C1WqoAH9A7s9e3QzZ9LqK5rXw"
	codeChallenge := "JxTX7oXvq0V2O2kKCFOw3NCuEAxCxhYH7Pxr0ZZEd0U"
	codeChallengeMethod := "S256"

	attempt := models.TokenAttempt{}
	attempt.UserID = &user.ID
	attempt.ServerID = &server.ID
	attempt.State = uuid.New().String()
	attempt.AuthorizationCode = uuid.New().String()
	attempt.CodeChallenge = &codeChallenge
	attempt.CodeChallengeMethod = &codeChallengeMethod
	f.DB.Create(&attempt)

	input := map[string]string{
		"grant_type":    "authorization_code",
		"code":          attempt.AuthorizationCode,
		"redirect_uri":  server.Url + "/callback",
		"code_verifier": codeVerifier,
	}

	expectedCode := 200
	statusCode, body := f.Request(&TestHttpRequest{
		Method:        "POST",
		Route:         "/api/v1/sso/token",
		Body:          FiberRequestFormPayload(input),
		Authorization: authHeaderServer,
		ContentType:   "application/x-www-form-urlencoded",
	})

	response := auth.SwaggerSSOToken{}
	_ = json.Unmarshal([]byte(body), &response)

	assert.Equal(t, expectedCode, statusCode, description)
	assert.NotEmpty(t, response.AccessToken, description)
}

func TestSSOTokenByAuthCodeWithInvalidPKCE(t *testing.T) {
	description := "get tokens by authorization code with invalid PKCE code_verifier"
	f := NewTestHTTP()
	defer f.Close()

	authHeaderServer, clientID := f.AuthorizationServiceBasic()
	server := models.Server{}
	f.DB.Where("client_id = ?", clientID).Find(&server)

	user := models.User{}
	user.Email = "test" + uuid.New().String() + "@example.com"
	user.Name = "Name " + uuid.New().String()
	f.DB.Create(&user)

	role := models.Role{}
	role.Code = "student" + uuid.New().String()
	role.Name = "Student"
	f.DB.Create(&role)

	roleUser := models.RoleRelation{}
	roleUser.RoleID = role.ID
	roleUser.UserID = &user.ID

	f.DB.Create(&roleUser)

	codeChallenge := "JxTX7oXvq0V2O2kKCFOw3NCuEAxCxhYH7Pxr0ZZEd0U"
	codeChallengeMethod := "S256"

	attempt := models.TokenAttempt{}
	attempt.UserID = &user.ID
	attempt.ServerID = &server.ID
	attempt.State = uuid.New().String()
	attempt.AuthorizationCode = uuid.New().String()
	attempt.CodeChallenge = &codeChallenge
	attempt.CodeChallengeMethod = &codeChallengeMethod
	f.DB.Create(&attempt)

	input := map[string]string{
		"grant_type":    "authorization_code",
		"code":          attempt.AuthorizationCode,
		"redirect_uri":  server.Url + "/callback",
		"code_verifier": "invalid_code_verifier",
	}

	expectedCode := 200
	statusCode, body := f.Request(&TestHttpRequest{
		Method:        "POST",
		Route:         "/api/v1/sso/token",
		Body:          FiberRequestFormPayload(input),
		Authorization: authHeaderServer,
		ContentType:   "application/x-www-form-urlencoded",
	})

	assert.Equal(t, expectedCode, statusCode, description)
	assert.Contains(t, body, "invalid code_verifier", description)
}

func TestSSOTokenByAuthCodeWithMissingCodeVerifier(t *testing.T) {
	description := "get tokens by authorization code with missing code_verifier when code_challenge exists"
	f := NewTestHTTP()
	defer f.Close()

	authHeaderServer, clientID := f.AuthorizationServiceBasic()
	server := models.Server{}
	f.DB.Where("client_id = ?", clientID).Find(&server)

	user := models.User{}
	user.Email = "test" + uuid.New().String() + "@example.com"
	user.Name = "Name " + uuid.New().String()
	f.DB.Create(&user)

	role := models.Role{}
	role.Code = "student" + uuid.New().String()
	role.Name = "Student"
	f.DB.Create(&role)

	roleUser := models.RoleRelation{}
	roleUser.RoleID = role.ID
	roleUser.UserID = &user.ID

	f.DB.Create(&roleUser)

	codeChallenge := "JxTX7oXvq0V2O2kKCFOw3NCuEAxCxhYH7Pxr0ZZEd0U"
	codeChallengeMethod := "S256"

	attempt := models.TokenAttempt{}
	attempt.UserID = &user.ID
	attempt.ServerID = &server.ID
	attempt.State = uuid.New().String()
	attempt.AuthorizationCode = uuid.New().String()
	attempt.CodeChallenge = &codeChallenge
	attempt.CodeChallengeMethod = &codeChallengeMethod
	f.DB.Create(&attempt)

	input := map[string]string{
		"grant_type":   "authorization_code",
		"code":         attempt.AuthorizationCode,
		"redirect_uri": server.Url + "/callback",
	}

	expectedCode := 200
	statusCode, body := f.Request(&TestHttpRequest{
		Method:        "POST",
		Route:         "/api/v1/sso/token",
		Body:          FiberRequestFormPayload(input),
		Authorization: authHeaderServer,
		ContentType:   "application/x-www-form-urlencoded",
	})

	assert.Equal(t, expectedCode, statusCode, description)
	assert.Contains(t, body, "code_verifier required", description)
}

func TestSSOIntrospectValidToken(t *testing.T) {
	description := "introspect valid access token"
	f := NewTestHTTP()
	defer f.Close()

	authHeaderClient, clientID := f.AuthorizationServiceBasic()
	server := models.Server{}
	f.DB.Where("client_id = ?", clientID).Find(&server)
	authHeader := f.AuthorizationUser(0, &server.ID)

	input := map[string]string{
		"token": strings.Replace(authHeader, "Bearer ", "", 1),
	}

	expectedCode := 200
	statusCode, body := f.Request(&TestHttpRequest{
		Method:        "POST",
		Route:         "/api/v1/sso/introspect",
		Body:          FiberRequestFormPayload(input),
		Authorization: authHeaderClient,
		ContentType:   "application/x-www-form-urlencoded",
	})

	response := auth.SSOTokenIntrospect{}
	_ = json.Unmarshal([]byte(body), &response)

	assert.Equal(t, expectedCode, statusCode, description)
	assert.True(t, response.Active, description)
	assert.Equal(t, server.ClientID, response.ClientID, description)
	assert.Equal(t, server.Name, response.Username, description)
	assert.Equal(t, "access", response.TokenType, description)
}

func TestSSOUserInfo(t *testing.T) {
	description := "get user info with valid token"
	f := NewTestHTTP()
	defer f.Close()

	_, clientID := f.AuthorizationServiceBasic()
	server := models.Server{}
	f.DB.Where("client_id = ?", clientID).Find(&server)
	user := models.User{}
	user.Email = uuid.New().String() + "@admin.com"
	f.DB.Create(&user)

	role := models.Role{}
	role.Code = "student" + uuid.New().String()
	role.Name = "Student"
	f.DB.Create(&role)

	roleUser := models.RoleRelation{}
	roleUser.RoleID = role.ID
	roleUser.UserID = &user.ID

	f.DB.Create(&roleUser)
	authHeader := f.AuthorizationUser(user.ID, &server.ID)

	expectedCode := 200
	statusCode, body := f.Request(
		&TestHttpRequest{
			Method:        "GET",
			Route:         "/api/v1/sso/userinfo",
			Authorization: authHeader,
		},
	)

	response := cms_client.SSOTokenPublicData{}
	err := json.Unmarshal([]byte(body), &response)
	if err != nil {
		panic(err)
	}

	assert.Equal(t, expectedCode, statusCode, description)
	assert.Equal(t, user.Email, response.Email, description)
	assert.Equal(t, user.Name, response.Name, description)
	assert.Equal(t, role.Code, response.UserRoleMain(), description)
	assert.Equal(t, fmt.Sprintf("%d", user.ID), response.Sub, description)
	assert.Equal(t, server.ClientID, response.Aud, description)
	assert.Equal(t, server.ID, *response.ServerID, description)
}

func TestSSOOpenIdConfiguration(t *testing.T) {
	description := "get OpenID configuration"
	f := NewTestHTTP()
	defer f.Close()

	expectedCode := 200
	statusCode, body := f.Request(&TestHttpRequest{
		Method: "GET",
		Route:  "/api/v1/sso/.well-known/openid-configuration",
	})

	response := auth.SSOOpenidConfigurationOutputDTO{}
	_ = json.Unmarshal([]byte(body), &response)

	assert.Equal(t, expectedCode, statusCode, description)
	assert.Equal(t, response.Issuer, "http://example.com/api/v1/sso", description)
	assert.NotEmpty(t, response.AuthorizationEndpoint, description)
	assert.Contains(t, response.AuthorizationEndpoint, "/api/v1/sso", description)
	assert.NotEmpty(t, response.TokenEndpoint, description)
	assert.NotEmpty(t, response.UserInfoEndpoint, description)
	assert.NotEmpty(t, response.JwksUri, description)
}

func TestSSOJwks(t *testing.T) {
	description := "get JWKS"
	f := NewTestHTTP()
	defer f.Close()

	expectedCode := 200
	statusCode, body := f.Request(&TestHttpRequest{
		Method: "GET",
		Route:  "/api/v1/sso/jwks",
	})

	response := auth.SSOJWKSOutputDTO{}
	_ = json.Unmarshal([]byte(body), &response)

	assert.Equal(t, expectedCode, statusCode, description)
	assert.Len(t, response.Keys, 3, description) // access + id и refresh ключи
	for _, key := range response.Keys {
		assert.Equal(t, "RSA", key.Kty, description)
		assert.Equal(t, "sig", key.Use, description)
		assert.Contains(t, []string{"access", "refresh", "id"}, key.Kid, description)
		assert.Equal(t, "RS256", key.Alg, description)
		assert.NotEmpty(t, key.N, description)
		assert.NotEmpty(t, key.E, description)
	}
}
