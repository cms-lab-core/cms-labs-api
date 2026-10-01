package routes

import (
	"fmt"
	"testing"
	"time"

	json "github.com/goccy/go-json"
	"github.com/stretchr/testify/assert"

	"github.com/cms-lab-core/cms-labs-api/shared/cms_client"

	"github.com/cms-lab-core/cms-labs-api/backend/app/models"
	"github.com/cms-lab-core/cms-labs-api/backend/app/usecases"
)

func TestV1LTIRoutingCreate(t *testing.T) {
	description := "create lti routing"
	f := NewTestHTTP()
	defer f.Close()
	authHeader := f.AuthorizationUser(0, nil)
	// Clear Table
	f.DB.Where("id > ?", 0).Delete(&models.LTIRouting{})

	input := usecases.LTIRoutingEditInputDTO{
		Name:           "Test LTIRouting",
		LTITitle:       "Test Title",
		LTIDescription: "Test Description",
		LTITaskID:      "12345",
		LTIParamsTask:  "param1,param2",
		Collaboration:  1,
		LabsType:       cms_client.PNETLabsTypeDefault,
		LabsPath:       "/path/to/labs",
		TestPath:       "/path/to/test",
		IsDefault:      false,
	}

	expectedCode := 200
	statusCode, body := f.Rpc(&TestRpcRequest{
		Method:        "lti_routing.upsert",
		Params:        input,
		Authorization: authHeader,
	})
	bodyModel := usecases.LTIRoutingEditResponse{}
	_ = json.Unmarshal([]byte(body), &bodyModel)

	assert.Equal(t, expectedCode, statusCode, description)
	assert.NotZero(t, bodyModel.Result.ID, description)

	// Verify database record
	var entity models.LTIRouting
	err := f.DB.First(&entity, bodyModel.Result.ID).Error
	assert.NoError(t, err, description)
	assert.Equal(t, input.Name, entity.Name, description)
	assert.Equal(t, input.LTITitle, entity.LTITitle, description)
	assert.Equal(t, input.LTIDescription, entity.LTIDescription, description)
}

func TestV1LTIRoutingList(t *testing.T) {
	description := "list lti routings"
	f := NewTestHTTP()
	defer f.Close()
	authHeader := f.AuthorizationUser(0, nil)
	// Clear Table
	f.DB.Where("id > ?", 0).Delete(&models.LTIRouting{})

	entities := make([]models.LTIRouting, 0)
	for i := 1; i <= 3; i++ {
		entity := models.LTIRouting{}
		entity.Name = fmt.Sprintf("LTIRouting %d", i)
		entity.LTITitle = fmt.Sprintf("Title %d", i)
		entity.LTIDescription = fmt.Sprintf("Description %d", i)
		entity.CreatedAt = time.Now().UTC().Add(time.Duration(i) * time.Minute)
		entity.UpdatedAt = time.Now().UTC().Add(time.Duration(i) * time.Minute)
		f.DB.Create(&entity)
		entities = append(entities, entity)
	}

	input := usecases.LTIRoutingListInputDTO{
		Limit:  10,
		Offset: 0,
	}

	expectedCode := 200
	statusCode, body := f.Rpc(&TestRpcRequest{
		Method:        "lti_routing.list",
		Params:        input,
		Authorization: authHeader,
	})
	bodyModel := usecases.LTIRoutingListResponse{}
	_ = json.Unmarshal([]byte(body), &bodyModel)

	assert.Equal(t, expectedCode, statusCode, description)
	assert.Equal(t, int64(3), bodyModel.Result.TotalCount, description)
	assert.Len(t, bodyModel.Result.Model, 3, description)

	// Verify response content
	for i, item := range bodyModel.Result.Model {
		assert.Equal(t, entities[len(entities)-i-1].Name, item.Name, description)
		assert.Equal(t, entities[len(entities)-i-1].ID, item.ID, description)
	}
}

func TestV1LTIRoutingDelete(t *testing.T) {
	description := "delete lti routing"
	f := NewTestHTTP()
	defer f.Close()
	authHeader := f.AuthorizationUser(0, nil)

	// Clear Table
	f.DB.Where("id > ?", 0).Delete(&models.LTIRouting{})

	// Create a test lti routing
	entity := models.LTIRouting{}
	entity.Name = "Test LTIRouting"
	entity.LTITitle = "Test Title"
	entity.LTIDescription = "Test Description"
	f.DB.Create(&entity)

	input := usecases.LTIRoutingDeleteInputDTO{
		ID: entity.ID, // Use the ID of the created entity
	}

	expectedCode := 200
	statusCode, body := f.Rpc(&TestRpcRequest{
		Method:        "lti_routing.delete",
		Params:        input,
		Authorization: authHeader,
	})
	bodyModel := usecases.LTIRoutingDeleteResponse{}
	_ = json.Unmarshal([]byte(body), &bodyModel)

	assert.Equal(t, expectedCode, statusCode, description)
	assert.Equal(t, entity.ID, bodyModel.Result.ID, description)

	var entityDB models.LTIRouting
	result := f.DB.First(&entityDB, entity.ID)
	assert.Equal(t, result.Error.Error(), "record not found", description)
}

func TestV1LTIRoutingGet(t *testing.T) {
	description := "get lti routing"
	f := NewTestHTTP()
	defer f.Close()
	authHeader := f.AuthorizationUser(0, nil)

	// Clear Table
	f.DB.Where("id > ?", 0).Delete(&models.LTIRouting{})

	// Create a test lti routing
	entity := models.LTIRouting{}
	entity.Name = "Test LTIRouting"
	entity.LTITitle = "Test Title"
	entity.LTIDescription = "Test Description"
	f.DB.Create(&entity)

	input := usecases.LTIRoutingGetInputDTO{
		ID: entity.ID, // Use the ID of the created entity
	}

	expectedCode := 200
	statusCode, body := f.Rpc(&TestRpcRequest{
		Method:        "lti_routing.get",
		Params:        input,
		Authorization: authHeader,
	})
	bodyModel := usecases.LTIRoutingGetResponse{}
	_ = json.Unmarshal([]byte(body), &bodyModel)

	assert.Equal(t, expectedCode, statusCode, description)
	assert.Equal(t, entity.ID, bodyModel.Result.Model.ID, description)
	assert.Equal(t, entity.Name, bodyModel.Result.Model.Name, description)
	assert.Equal(t, entity.LTITitle, bodyModel.Result.Model.LTITitle, description)
	assert.Equal(t, entity.LTIDescription, bodyModel.Result.Model.LTIDescription, description)
}
