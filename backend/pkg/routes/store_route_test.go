package routes

import (
	"testing"

	"github.com/goccy/go-json"

	"github.com/gofiber/fiber/v2"

	"github.com/cms-lab-core/cms-labs-api/backend/app/models"
	"github.com/cms-lab-core/cms-labs-api/backend/app/models/types"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestStoreGet(t *testing.T) {
	description := "get user store"
	f := NewTestHTTP()
	defer f.Close()

	testUser := models.User{
		UserBase: models.UserBase{
			Email: uuid.New().String() + "@user.com",
			Name:  "Test User",
		},
		UserSecret: models.UserSecret{
			Store: types.JsonStore{
				"theme": "dark",
				"prefs": map[string]interface{}{"notifications": true},
				"word":  "Русские символы в БД",
			},
		},
	}
	f.DB.Create(&testUser)

	authHeader := f.AuthorizationUser(testUser.ID, nil)

	expectedCode := 200
	statusCode, body := f.Rpc(&TestRpcRequest{
		Method:        "user.global_store_get",
		Authorization: authHeader,
	})

	var response types.UserStoreGetResponse
	err := json.Unmarshal([]byte(body), &response)
	assert.NoError(t, err, description)

	assert.Equal(t, expectedCode, statusCode, description)
	assert.Equal(t, "dark", response.Result["theme"], description)
	assert.True(t, response.Result["prefs"].(map[string]interface{})["notifications"].(bool), description)
}

func TestStoreGetEmpty(t *testing.T) {
	description := "get empty user store"
	f := NewTestHTTP()
	defer f.Close()

	testUser := models.User{
		UserBase: models.UserBase{
			Email: uuid.New().String() + "@user.com",
			Name:  "Test User",
		},
	}
	f.DB.Create(&testUser)

	authHeader := f.AuthorizationUser(testUser.ID, nil)

	expectedCode := 200
	statusCode, body := f.Rpc(&TestRpcRequest{
		Method:        "user.global_store_get",
		Authorization: authHeader,
	})

	var response types.UserStoreGetResponse
	err := json.Unmarshal([]byte(body), &response)
	assert.NoError(t, err, description)

	assert.Equal(t, expectedCode, statusCode, description)
	assert.Empty(t, response.Result, description)
}

func TestStoreSet(t *testing.T) {
	description := "set user store"
	f := NewTestHTTP()
	defer f.Close()

	testUser := models.User{
		UserBase: models.UserBase{
			Email: uuid.New().String() + "@user.com",
			Name:  "Test User",
		},
	}
	f.DB.Create(&testUser)

	authHeader := f.AuthorizationUser(testUser.ID, nil)

	input := map[string]interface{}{
		"theme": "light",
		"prefs": map[string]interface{}{
			"notifications": false,
			"language":      "en",
		},
	}

	expectedCode := 200
	statusCode, _ := f.Rpc(&TestRpcRequest{
		Method:        "user.global_store_set",
		Params:        input,
		Authorization: authHeader,
	})

	assert.Equal(t, expectedCode, statusCode, description)

	// Verify database record
	var user models.User
	err := f.DB.First(&user, testUser.ID).Error
	assert.NoError(t, err, description)
	assert.Equal(t, "light", user.Store["theme"], description)
	assert.False(t, user.Store["prefs"].(map[string]interface{})["notifications"].(bool), description)
	assert.Equal(t, "en", user.Store["prefs"].(map[string]interface{})["language"].(string), description)
}

func TestStoreSetInvalidJSON(t *testing.T) {
	description := "set user store with invalid JSON"
	f := NewTestHTTP()
	defer f.Close()
	authHeader := f.AuthorizationUser(0, nil)

	expectedCode := 500
	statusCode, body := f.Rpc(&TestRpcRequest{
		Method:        "user.global_store_set",
		Params:        "{invalid json}",
		Authorization: authHeader,
	})

	assert.Equal(t, expectedCode, statusCode, description)
	assert.Contains(t, body, "character for map value", description)
}

func TestStoreUnauthorized(t *testing.T) {
	tests := []struct {
		method       string
		description  string
		expectedCode int
	}{
		{
			method:       "user.global_store_get",
			description:  "get store unauthorized",
			expectedCode: 500,
		},
		{
			method:       "user.global_store_set",
			description:  "set store unauthorized",
			expectedCode: 500,
		},
	}

	f := NewTestHTTP()
	defer f.Close()

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			statusCode, body := f.Rpc(&TestRpcRequest{
				Method: test.method,
				Params: fiber.Map{},
			})

			assert.Equal(t, test.expectedCode, statusCode, test.description)
			assert.Contains(t, body, "unauthorized", test.description)
		})
	}
}
