package routes

import (
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"

	json "github.com/goccy/go-json"
	"github.com/stretchr/testify/assert"

	"github.com/cms-lab-core/cms-labs-api/backend/app/usecases"
	"github.com/google/uuid"

	"github.com/cms-lab-core/cms-labs-api/backend/app/models"
)

func TestV1ServerRouteGet(t *testing.T) {
	description := "get server"
	f := NewTestHTTP()
	defer f.Close()
	entity := models.Server{}
	entity.Name = "Server"
	entity.Url = "https://localhost"
	entity.Type = models.ServerTypePnet
	entity.IsActive = true
	entity.MinutesForDisconnect = 10
	entity.MaxCountUsersLimit = 20
	entity.UnitRate = 1
	entity.Token = uuid.New().String()
	entity.ClientID = uuid.New().String()
	entity.CreatedAt = time.Now().UTC()
	entity.UpdatedAt = time.Now().UTC()
	f.DB.Create(&entity)
	authHeader := f.AuthorizationUser(0, nil)
	expectedCode := 200
	statusCode, body := f.Rpc(&TestRpcRequest{
		Method: "server.get",
		Params: fiber.Map{
			"id": entity.ID,
		},
		Authorization: authHeader,
	})
	bodyModel := usecases.ServerGetResponse{}
	_ = json.Unmarshal([]byte(body), &bodyModel)

	assert.Equal(t, expectedCode, statusCode, description)
	assert.Equal(t, entity.ID, bodyModel.Result.Model.ID, description)
}

func TestV1ServerRouteGetNotFound(t *testing.T) {
	description := "not found pnet"
	f := NewTestHTTP()
	defer f.Close()
	authHeader := f.AuthorizationUser(0, nil)
	expectedCode := 500
	r := &TestRpcRequest{
		Method: "server.get",
		Params: fiber.Map{
			"id": 9999999,
		},
		Authorization: authHeader,
	}
	statusCode, body := f.Rpc(r)
	expectedBody := map[string]interface{}{
		"error": map[string]interface{}{
			"code": -32010,
			"data": map[string]interface{}{
				"code":    "server_not_found",
				"message": "Server has not found",
			},
			"message": "Validation error",
		},
		"id":      r.ID,
		"jsonrpc": "2.0",
	}

	assert.Equal(t, expectedCode, statusCode, description)
	assert.Equal(t, FiberJSON(expectedBody), body, description)
}

func TestV1ServerRouteSearch(t *testing.T) {
	f := NewTestHTTP()
	defer f.Close()
	now := time.Now().UTC()
	entity := models.Server{}
	entity.Name = "Server " + uuid.New().String()
	entity.Url = "https://localhost"
	entity.Type = models.ServerTypePnet
	entity.IsActive = true
	entity.MinutesForDisconnect = 10
	entity.MaxCountUsersLimit = 20
	entity.LastOnlineStatus = &now
	entity.UnitRate = 20
	entity.Token = uuid.New().String()
	entity.ClientID = uuid.New().String()
	f.DB.Create(&entity)
	entity2 := models.Server{}
	entity2.Name = entity.Name
	entity2.UnitRate = 10
	entity2.Url = "https://anotherhost"
	entity2.Type = models.ServerTypeOpenID
	entity2.IsActive = false
	entity2.Token = uuid.New().String()
	entity2.ClientID = uuid.New().String()
	f.DB.Create(&entity2)
	authHeader := f.AuthorizationUser(0, nil)
	tests := []struct {
		description string
		body        map[string]any
		ids         []uint
	}{
		{
			description: "Search entity order from ALL",
			body: map[string]any{
				"search":   entity.Name,
				"limit":    100,
				"offset":   0,
				"status":   "all",
				"order_by": "createdAt",
			},
			ids: []uint{entity2.ID, entity.ID},
		},
		{
			description: "Search entity 2 from ALL",
			body: map[string]any{
				"search":   entity.Name,
				"limit":    1,
				"offset":   0,
				"status":   "all",
				"order_by": "createdAt",
			},
			ids: []uint{entity2.ID},
		},
		{
			description: "Search entity 1 from ALL",
			body: map[string]any{
				"search":   entity.Name,
				"limit":    1,
				"offset":   1,
				"status":   "all",
				"order_by": "createdAt",
			},
			ids: []uint{entity.ID},
		},
		{
			description: "Search entity 1 from active",
			body: map[string]any{
				"search":   entity.Name,
				"limit":    100,
				"offset":   0,
				"status":   "active",
				"order_by": "createdAt",
			},
			ids: []uint{entity.ID},
		},
		{
			description: "Order from unitRate",
			body: map[string]any{
				"search":   entity.Name,
				"limit":    100,
				"offset":   0,
				"status":   "all",
				"order_by": "unitRate",
			},
			ids: []uint{entity.ID, entity2.ID},
		},
	}
	for _, test := range tests {
		statusCode, body := f.Rpc(&TestRpcRequest{
			Method:        "server.list",
			Params:        test.body,
			Authorization: authHeader,
		})
		bodyModel := usecases.ServerListResponse{}
		_ = json.Unmarshal([]byte(body), &bodyModel)
		ids := make([]uint, 0)
		for _, model := range bodyModel.Result.Model {
			ids = append(ids, model.Model.ID)
		}
		assert.Equal(t, statusCode, 200, test.description)
		assert.Equal(t, test.ids, ids, test.description)
	}
}

func TestV1ServerRouteDelete(t *testing.T) {
	f := NewTestHTTP()
	defer f.Close()
	now := time.Now().UTC()
	entity := models.Server{}
	entity.Name = "Server " + uuid.New().String()
	entity.Url = "https://localhost"
	entity.Type = models.ServerTypePnet
	entity.IsActive = true
	entity.MinutesForDisconnect = 10
	entity.MaxCountUsersLimit = 20
	entity.LastOnlineStatus = &now
	entity.UnitRate = 20
	entity.Token = uuid.New().String()
	entity.ClientID = uuid.New().String()
	f.DB.Create(&entity)

	authHeader := f.AuthorizationUser(0, nil)
	tests := []struct {
		description string
		id          uint
		statusCode  int
		expectedErr string
	}{
		{
			description: "Successfully delete Server",
			id:          entity.ID,
			statusCode:  200,
		},
		{
			description: "Fail to delete non-existent Server",
			id:          999999,
		},
	}

	for _, test := range tests {
		statusCode, body := f.Rpc(
			&TestRpcRequest{
				Method: "server.delete",
				Params: fiber.Map{
					"id": test.id,
				},
				Authorization: authHeader,
			},
		)
		response := usecases.ServerDeleteResponse{}
		_ = json.Unmarshal([]byte(body), &response)
		assert.Equal(t, 200, statusCode, test.description)
		assert.Equal(t, test.id, response.Result.ID, test.description)

		// Проверяем, что сервер успешно удален
		var deletedEntity models.Server
		f.DB.First(&deletedEntity, test.id)
		var idNotFound uint = 0
		assert.Equal(t, idNotFound, deletedEntity.ID, test.description)
	}
}

func TestV1ServerRouteCreate(t *testing.T) {
	description := "Create new server"
	f := NewTestHTTP()
	defer f.Close()
	authHeader := f.AuthorizationUser(0, nil)

	// Тестовые данные
	now := time.Now().UTC()
	entity := models.Server{}

	entity.Name = "Server " + uuid.New().String()
	entity.Url = "https://localhost"
	entity.Type = models.ServerTypePnet
	entity.IsActive = true
	entity.MinutesForDisconnect = 10
	entity.MaxCountUsersLimit = 20
	entity.LastOnlineStatus = &now
	entity.UnitRate = 20
	entity.ClientID = uuid.New().String()

	statusCode, body := f.Rpc(&TestRpcRequest{
		Method: "server.upsert",
		Params: usecases.ServerEditInputDTO{
			ClientID:             entity.ClientID,
			Type:                 entity.Type,
			Name:                 entity.Name,
			Url:                  entity.Url,
			IsActive:             entity.IsActive,
			MinutesForDisconnect: entity.MinutesForDisconnect,
			MaxCountUsersLimit:   entity.MaxCountUsersLimit,
			UnitRate:             entity.UnitRate,
		},
		Authorization: authHeader,
	})

	bodyModel := usecases.ServerEditResponse{}
	_ = json.Unmarshal([]byte(body), &bodyModel)

	assert.Equal(t, statusCode, 200, description)
	assert.Equal(t, true, bodyModel.Result.ID > 0, description)

	// Проверка, что сервер действительно создан в базе данных
	var createdEntity models.Server
	f.DB.First(&createdEntity, bodyModel.Result.ID)
	assert.Equal(t, entity.Name, createdEntity.Name, description)
	assert.Equal(t, entity.Url, createdEntity.Url, description)
	assert.Equal(t, entity.Type, createdEntity.Type, description)
	assert.Equal(t, entity.IsActive, createdEntity.IsActive, description)
	assert.Equal(t, entity.MinutesForDisconnect, createdEntity.MinutesForDisconnect, description)
	assert.Equal(t, entity.MaxCountUsersLimit, createdEntity.MaxCountUsersLimit, description)
	assert.Equal(t, entity.UnitRate, createdEntity.UnitRate, description)
}
