package routes

import (
	"fmt"
	rand "math/rand/v2"
	"testing"

	"github.com/cms-lab-core/cms-labs-api/pnetlabaddon/app/queries"
	"github.com/stretchr/testify/assert"

	"github.com/cms-lab-core/cms-labs-api/pnetlabaddon/app/models"
	"github.com/google/uuid"
	"github.com/h2non/gock"
)

func externalSSOMock() (int, string, string) {
	externalUserId := rand.IntN(99999)
	externalUserName := uuid.New().String()
	externalUserEmail := externalUserName + "@admin.com"

	cmsCoreTokenResponse := map[string]interface{}{
		"access_token":  "access_token",
		"refresh_token": "refresh_token",
		"token_type":    "bearer",
		"expires_in":    -1,
		"state":         "77c02576-43b0-4aeb-a69c-d8baa61e0e3b",
		"user_id":       fmt.Sprintf("%d", externalUserId),
	}

	gock.New("https://cms-core.local").
		MatchHeader("Authorization", "Basic Y2xpZW50LWlkOnBuZXRsYWI=").
		MatchHeader("X-Client-ID", "client-id").
		Post("/api/v1/sso/token").
		Reply(200).
		JSON(cmsCoreTokenResponse)

	cmsCoreUserResponse := map[string]interface{}{
		"iss":            "",
		"sub":            fmt.Sprintf("%d", externalUserId),
		"aud":            "d777dba8-e5f1-4967-883f-8c7465940e07",
		"exp":            -1,
		"iat":            -1,
		"nonce":          "24d9b73d-f0b9-4435-92e2-6c6ee7b7c9b5",
		"email":          externalUserEmail,
		"name":           externalUserName,
		"username":       externalUserEmail,
		"server_id":      2,
		"role":           "student",
		"last_launch_id": "1",
	}

	gock.New("https://cms-core.local").
		MatchHeader("Authorization", "Bearer access_token").
		MatchHeader("X-Client-ID", "client-id").
		Get("/api/v1/sso/userinfo").
		Reply(200).
		JSON(cmsCoreUserResponse)

	gock.New("https://guacamole.local").
		Post("/html5/api/tokens").
		Reply(200).
		JSON(map[string]interface{}{"authToken": "guacamoleToken"})
	return externalUserId, externalUserName, externalUserEmail
}

func TestSSOSecondFactorSuccess(t *testing.T) {
	description := "second factor set cookies. "
	defer gock.Off()
	externalUserId, externalUserName, externalUserEmail := externalSSOMock()
	f := NewFiberTestHTTP()

	expectedCode := 307
	statusCode, _ := f.Request(&FiberTestHttpRequest{
		Method: "GET",
		Route:  "/pnet-lab-addon/api/v1/sso/openid?code=b4f0cf67-e403-41d0-8afc-4dd252b43e38&application=client-id&path=/&state=77c02576-43b0-4aeb-a69c-d8baa61e0e3b&extra=",
		Body:   FiberRequestPayload(nil),
	})
	assert.Equal(t, expectedCode, statusCode, description+"Status code")

	var entityUser models.User
	f.DB.Where("email = ?", externalUserEmail).Find(&entityUser)
	assert.Equal(t, entityUser.Email, externalUserEmail, description+"User email")
	assert.Equal(t, entityUser.Username, externalUserName, description+"User name")
	assert.Equal(t, entityUser.Name, fmt.Sprintf("%d", externalUserId), description+"User extId")

	var entityRole models.UserRole
	f.DB.Where("user_role_name = ?", queries.UserRoleStudent).Find(&entityRole)
	assert.Equal(t, entityRole.UserRoleName, queries.UserRoleStudent, description+"User role")
	assert.Equal(t, entityUser.Role, fmt.Sprintf("%d", entityRole.UserRoleID), description+"User role")

	var entityTokenGuacamole models.HTML5
	f.DB.Where("pod = ?", entityUser.Pod).Find(&entityTokenGuacamole)
	assert.Equal(t, entityTokenGuacamole.Username, externalUserEmail, description+"User guacamole")

	var existsGuacamoleEntity bool
	_ = f.DB.Raw(
		"SELECT 1 FROM `guacdb`.`guacamole_entity` WHERE entity_id = ? AND name = ? AND type = ?",
		entityUser.Pod+1000,
		entityUser.Email,
		queries.GuacamoleEntityTypeUser,
	).Row().Scan(&existsGuacamoleEntity)
	assert.True(t, existsGuacamoleEntity, description+"Guacamole entity")

	var existsGuacamoleUserEntity bool
	_ = f.DB.Raw(
		"SELECT 1 FROM `guacdb`.`guacamole_user` WHERE user_id = ? AND entity_id = ?",
		entityUser.Pod+1000,
		entityUser.Pod+1000,
	).Row().Scan(&existsGuacamoleUserEntity)
	assert.True(t, existsGuacamoleUserEntity, description+"Guacamole user")

	var existsGuacamolePermissionEntity bool
	_ = f.DB.Raw(
		"SELECT 1 FROM `guacdb`.`guacamole_user_permission` WHERE entity_id = ? AND affected_user_id = ? AND permission = ?",
		entityUser.Pod+1000,
		entityUser.Pod+1000,
		queries.GuacamolePermissionTypeRead,
	).Row().Scan(&existsGuacamolePermissionEntity)
	assert.True(t, existsGuacamolePermissionEntity, description+"Guacamole permission")
}

func TestSSOSecondFactorNotDefaultRole(t *testing.T) {
	description := "second factor set cookies"
	defer gock.Off()
	externalUserId, externalUserName, externalUserEmail := externalSSOMock()
	f := NewFiberTestHTTP()

	createEntityRole := models.UserRole{
		UserRoleName: "some_role_" + uuid.New().String(),
	}
	f.DB.Create(&createEntityRole)
	createEntityUser := &models.User{
		Username: externalUserName,
		Email:    externalUserEmail,
		Name:     fmt.Sprintf("%d", externalUserId),
		Role:     fmt.Sprintf("%d", createEntityRole.UserRoleID),
	}
	f.DB.Create(&createEntityUser)

	expectedCode := 307
	statusCode, _ := f.Request(&FiberTestHttpRequest{
		Method: "GET",
		Route:  "/pnet-lab-addon/api/v1/sso/openid?code=b4f0cf67-e403-41d0-8afc-4dd252b43e38&application=client-id&path=/&state=77c02576-43b0-4aeb-a69c-d8baa61e0e3b&extra=",
		Body:   FiberRequestPayload(nil),
	})
	assert.Equal(t, expectedCode, statusCode, description)

	var entityUser models.User
	f.DB.Where("email = ?", externalUserEmail).Find(&entityUser)
	assert.Equal(t, entityUser.Email, externalUserEmail, description)
	assert.Equal(t, entityUser.Role, fmt.Sprintf("%d", createEntityRole.UserRoleID), description)
}
