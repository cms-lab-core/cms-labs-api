package routes

import (
	"testing"

	json "github.com/goccy/go-json"
	"github.com/stretchr/testify/assert"
	"gorm.io/datatypes"

	"github.com/cms-lab-core/cms-labs-api/backend/app/models"
	"github.com/cms-lab-core/cms-labs-api/backend/app/queries"
	"github.com/cms-lab-core/cms-labs-api/backend/app/usecases"
)

type ltiAttemptFixture struct {
	f              *TestHTTP
	authHeader     string
	serverClientID string
	server         models.Server
	user           models.User
	routing        models.LTIRouting
	attempts       []models.LTIAttempt
}

func setupLTIAttemptFixture(f *TestHTTP) *ltiAttemptFixture {
	f.DB.Where("id > ?", 0).Delete(&models.LTIAttempt{})
	f.DB.Where("id > ?", 0).Delete(&models.User{})
	f.DB.Where("id > ?", 0).Delete(&models.LTIRouting{})
	f.DB.Where("id > ?", 0).Delete(&models.Server{})

	authHeader, serverClientID := f.AuthorizationServiceBasic()

	var server models.Server
	_ = f.DB.Where("client_id = ?", serverClientID).Find(&server)

	user := models.User{}
	user.Email = "test@example.com"
	user.Name = "Test User"
	f.DB.Create(&user)

	routing := models.LTIRouting{}
	routing.Name = "Test Routing"
	routing.LTITitle = "Test Title"
	f.DB.Create(&routing)

	f.DB.Model(&models.Server{}).Where("id > ?", 0)

	attempts := []models.LTIAttempt{
		{},
		{},
		{},
	}
	attempts[0].AttemptID = "attempt-active-1"
	attempts[0].Status = models.AttemptStatusActive
	attempts[0].UserID = user.ID
	attempts[0].ServerID = &server.ID
	attempts[0].LTIRoutingID = routing.ID

	attempts[1].AttemptID = "attempt-pending-1"
	attempts[1].Status = models.AttemptStatusPending
	attempts[1].UserID = user.ID
	attempts[1].ServerID = &server.ID
	attempts[1].LTIRoutingID = routing.ID

	attempts[2].AttemptID = "attempt-completed-1"
	attempts[2].Status = models.AttemptStatusCompleted
	attempts[2].UserID = user.ID
	attempts[2].ServerID = &server.ID
	attempts[2].LTIRoutingID = routing.ID

	for i := range attempts {
		f.DB.Create(&attempts[i])
	}

	return &ltiAttemptFixture{
		f:              f,
		authHeader:     authHeader,
		serverClientID: serverClientID,
		server:         server,
		user:           user,
		routing:        routing,
		attempts:       attempts,
	}
}

func TestV1LTIAttemptListExternal(t *testing.T) {
	f := NewTestHTTP()
	defer f.Close()

	fix := setupLTIAttemptFixture(f)

	input := queries.LTIAttemptSearchParams{
		Limit: 10,
	}

	expectedCode := 200
	statusCode, body := f.Rpc(&TestRpcRequest{
		Method:        "lti_attempt.list_external",
		Params:        input,
		Authorization: fix.authHeader,
	})
	bodyModel := usecases.LTIAttemptListResponse{}
	_ = json.Unmarshal([]byte(body), &bodyModel)

	assert.Equal(t, expectedCode, statusCode)
	assert.GreaterOrEqual(t, len(bodyModel.Result.Model), 1)
}

func TestV1LTIAttemptUpdateStatusPreservesServerWhenOmitted(t *testing.T) {
	f := NewTestHTTP()
	defer f.Close()

	fix := setupLTIAttemptFixture(f)
	authHeader := f.AuthorizationUser(0, nil)
	attempt := fix.attempts[0]

	statusCode, _ := f.Rpc(&TestRpcRequest{
		Method: "lti_attempt.update",
		Params: map[string]interface{}{
			"id":     attempt.ID,
			"status": models.AttemptStatusTerminating,
		},
		Authorization: authHeader,
	})

	updated := models.LTIAttempt{}
	f.DB.First(&updated, attempt.ID)
	assert.Equal(t, 200, statusCode)
	assert.Equal(t, models.AttemptStatusTerminating, updated.Status)
	if assert.NotNil(t, updated.ServerID) {
		assert.Equal(t, fix.server.ID, *updated.ServerID)
	}
}

func TestV1LTIAttemptListExternalFilterByAttemptIDs(t *testing.T) {
	f := NewTestHTTP()
	defer f.Close()

	fix := setupLTIAttemptFixture(f)

	input := queries.LTIAttemptSearchParams{
		AttemptIds: []string{fix.attempts[0].AttemptID},
		Limit:      10,
	}

	expectedCode := 200
	statusCode, body := f.Rpc(&TestRpcRequest{
		Method:        "lti_attempt.list_external",
		Params:        input,
		Authorization: fix.authHeader,
	})
	bodyModel := usecases.LTIAttemptListResponse{}
	_ = json.Unmarshal([]byte(body), &bodyModel)

	assert.Equal(t, expectedCode, statusCode)
	assert.Equal(t, 1, len(bodyModel.Result.Model))
	assert.Equal(t, fix.attempts[0].AttemptID, bodyModel.Result.Model[0].AttemptID)
}

func TestV1LTIAttemptListExternalFilterByUserIDs(t *testing.T) {
	f := NewTestHTTP()
	defer f.Close()

	fix := setupLTIAttemptFixture(f)

	input := queries.LTIAttemptSearchParams{
		UserIds: []uint{fix.user.ID},
		Limit:   10,
	}

	expectedCode := 200
	statusCode, body := f.Rpc(&TestRpcRequest{
		Method:        "lti_attempt.list_external",
		Params:        input,
		Authorization: fix.authHeader,
	})
	bodyModel := usecases.LTIAttemptListResponse{}
	_ = json.Unmarshal([]byte(body), &bodyModel)

	assert.Equal(t, expectedCode, statusCode)
	assert.GreaterOrEqual(t, len(bodyModel.Result.Model), 1)
}

func TestV1LTIAttemptListExternalFilterByStatuses(t *testing.T) {
	f := NewTestHTTP()
	defer f.Close()

	fix := setupLTIAttemptFixture(f)

	input := queries.LTIAttemptSearchParams{
		Statuses: []string{models.AttemptStatusActive},
		Limit:    10,
	}

	expectedCode := 200
	statusCode, body := f.Rpc(&TestRpcRequest{
		Method:        "lti_attempt.list_external",
		Params:        input,
		Authorization: fix.authHeader,
	})
	bodyModel := usecases.LTIAttemptListResponse{}
	_ = json.Unmarshal([]byte(body), &bodyModel)

	assert.Equal(t, expectedCode, statusCode)
	assert.Equal(t, 1, len(bodyModel.Result.Model))
	assert.Equal(t, int64(1), bodyModel.Result.TotalCount)
	assert.Equal(t, models.AttemptStatusActive, bodyModel.Result.Model[0].Status)
}

func TestV1LTIAttemptListExternalFilterByMultipleStatuses(t *testing.T) {
	f := NewTestHTTP()
	defer f.Close()

	fix := setupLTIAttemptFixture(f)

	input := queries.LTIAttemptSearchParams{
		Statuses: []string{models.AttemptStatusActive, models.AttemptStatusPending},
		Limit:    10,
	}

	expectedCode := 200
	statusCode, body := f.Rpc(&TestRpcRequest{
		Method:        "lti_attempt.list_external",
		Params:        input,
		Authorization: fix.authHeader,
	})
	bodyModel := usecases.LTIAttemptListResponse{}
	_ = json.Unmarshal([]byte(body), &bodyModel)

	assert.Equal(t, expectedCode, statusCode)
	assert.Equal(t, 2, len(bodyModel.Result.Model))
	assert.Equal(t, int64(2), bodyModel.Result.TotalCount)
}

func TestV1LTIAttemptListExternalFilterByServerClientIDs(t *testing.T) {
	f := NewTestHTTP()
	defer f.Close()

	fix := setupLTIAttemptFixture(f)

	input := queries.LTIAttemptSearchParams{
		ServerClientIds: []string{fix.serverClientID},
		Limit:           10,
	}

	expectedCode := 200
	statusCode, body := f.Rpc(&TestRpcRequest{
		Method:        "lti_attempt.list_external",
		Params:        input,
		Authorization: fix.authHeader,
	})
	bodyModel := usecases.LTIAttemptListResponse{}
	_ = json.Unmarshal([]byte(body), &bodyModel)

	assert.Equal(t, expectedCode, statusCode)
	assert.GreaterOrEqual(t, len(bodyModel.Result.Model), 1)
}

func TestV1LTIAttemptListExternalFilterByLimit(t *testing.T) {
	f := NewTestHTTP()
	defer f.Close()

	fix := setupLTIAttemptFixture(f)

	input := queries.LTIAttemptSearchParams{
		Limit: 1,
	}

	expectedCode := 200
	statusCode, body := f.Rpc(&TestRpcRequest{
		Method:        "lti_attempt.list_external",
		Params:        input,
		Authorization: fix.authHeader,
	})
	bodyModel := usecases.LTIAttemptListResponse{}
	_ = json.Unmarshal([]byte(body), &bodyModel)

	assert.Equal(t, expectedCode, statusCode)
	assert.LessOrEqual(t, len(bodyModel.Result.Model), 1)
	assert.Equal(t, int64(len(fix.attempts)), bodyModel.Result.TotalCount)
}

func TestV1LTIAttemptListExternalFilterByOffset(t *testing.T) {
	f := NewTestHTTP()
	defer f.Close()

	fix := setupLTIAttemptFixture(f)

	inputList := queries.LTIAttemptSearchParams{
		Limit: 10,
	}
	_, body := f.Rpc(&TestRpcRequest{
		Method:        "lti_attempt.list_external",
		Params:        inputList,
		Authorization: fix.authHeader,
	})
	bodyModelList := usecases.LTIAttemptListResponse{}
	_ = json.Unmarshal([]byte(body), &bodyModelList)
	totalCount := len(bodyModelList.Result.Model)

	input := queries.LTIAttemptSearchParams{
		Limit:  1,
		Offset: 1,
	}

	expectedCode := 200
	statusCode, body := f.Rpc(&TestRpcRequest{
		Method:        "lti_attempt.list_external",
		Params:        input,
		Authorization: fix.authHeader,
	})
	bodyModel := usecases.LTIAttemptListResponse{}
	_ = json.Unmarshal([]byte(body), &bodyModel)

	assert.Equal(t, expectedCode, statusCode)
	if totalCount > 1 {
		assert.LessOrEqual(t, len(bodyModel.Result.Model), 1)
	}
}

func TestV1LTIAttemptListExternalFilterCombined(t *testing.T) {
	f := NewTestHTTP()
	defer f.Close()

	fix := setupLTIAttemptFixture(f)

	input := queries.LTIAttemptSearchParams{
		Statuses:        []string{models.AttemptStatusActive},
		UserIds:         []uint{fix.user.ID},
		ServerClientIds: []string{fix.serverClientID},
		Limit:           10,
	}

	expectedCode := 200
	statusCode, body := f.Rpc(&TestRpcRequest{
		Method:        "lti_attempt.list_external",
		Params:        input,
		Authorization: fix.authHeader,
	})
	bodyModel := usecases.LTIAttemptListResponse{}
	_ = json.Unmarshal([]byte(body), &bodyModel)

	assert.Equal(t, expectedCode, statusCode)
	assert.Equal(t, 1, len(bodyModel.Result.Model))
	assert.Equal(t, models.AttemptStatusActive, bodyModel.Result.Model[0].Status)
}

func TestV1LTIAttemptUpdateExternal(t *testing.T) {
	f := NewTestHTTP()
	defer f.Close()

	fix := setupLTIAttemptFixture(f)

	input := usecases.LTIAttemptEditBulkInputDTO{
		Models: []usecases.LTIAttemptEditBulkInput{
			{
				AttemptID: fix.attempts[0].AttemptID,
				Status:    models.AttemptStatusCompleted,
			},
		},
	}

	expectedCode := 200
	statusCode, body := f.Rpc(&TestRpcRequest{
		Method:        "lti_attempt.update_external",
		Params:        input,
		Authorization: fix.authHeader,
	})
	bodyModel := usecases.LTIAttemptEditBulkResponse{}
	_ = json.Unmarshal([]byte(body), &bodyModel)

	assert.Equal(t, expectedCode, statusCode)
	assert.Equal(t, uint(1), bodyModel.Result.Count)

	var updatedAttempt models.LTIAttempt
	f.DB.Where("attempt_id = ?", fix.attempts[0].AttemptID).First(&updatedAttempt)
	assert.Equal(t, models.AttemptStatusCompleted, updatedAttempt.Status)
}

func TestV1LTIAttemptUpdateExternalBulk(t *testing.T) {
	f := NewTestHTTP()
	defer f.Close()

	fix := setupLTIAttemptFixture(f)
	result2 := datatypes.NewJSONType(models.LTIAttemptResult{
		MaxScore:      100.0,
		CurrentScore:  10.0,
		ResultDisplay: "https://localhost",
	})

	input := usecases.LTIAttemptEditBulkInputDTO{
		Models: []usecases.LTIAttemptEditBulkInput{
			{
				AttemptID: fix.attempts[0].AttemptID,
				Status:    models.AttemptStatusCompleted,
			},
			{
				AttemptID: fix.attempts[1].AttemptID,
				Status:    models.AttemptStatusActive,
				Result:    &result2,
			},
		},
	}

	expectedCode := 200
	statusCode, body := f.Rpc(&TestRpcRequest{
		Method:        "lti_attempt.update_external",
		Params:        input,
		Authorization: fix.authHeader,
	})
	bodyModel := usecases.LTIAttemptEditBulkResponse{}
	_ = json.Unmarshal([]byte(body), &bodyModel)

	assert.Equal(t, expectedCode, statusCode)
	assert.Equal(t, uint(2), bodyModel.Result.Count)

	attemptDB1 := models.LTIAttempt{}
	f.DB.Where("attempt_id = ?", fix.attempts[0].AttemptID).First(&attemptDB1)
	assert.Equal(t, input.Models[0].Status, attemptDB1.Status)

	attemptDB2 := models.LTIAttempt{}
	f.DB.Where("attempt_id = ?", fix.attempts[1].AttemptID).First(&attemptDB2)
	assert.Equal(t, input.Models[1].Status, attemptDB2.Status)
	assert.Equal(t, input.Models[1].Result.Data().MaxScore, attemptDB2.Result.Data().MaxScore)
	assert.Equal(t, input.Models[1].Result.Data().CurrentScore, attemptDB2.Result.Data().CurrentScore)
	assert.Equal(t, input.Models[1].Result.Data().ResultDisplay, attemptDB2.Result.Data().ResultDisplay)
}

func TestV1LTIAttemptUpdateExternalIsIdempotentByCheckID(t *testing.T) {
	f := NewTestHTTP()
	defer f.Close()

	fix := setupLTIAttemptFixture(f)
	first := datatypes.NewJSONType(models.LTIAttemptResult{
		MaxScore: 10, CurrentScore: 8, ResultDisplay: "8/10", CheckID: "check-1", Report: "first report",
	})
	input := usecases.LTIAttemptEditBulkInputDTO{Models: []usecases.LTIAttemptEditBulkInput{{
		AttemptID: fix.attempts[1].AttemptID, Status: models.AttemptStatusActive, Result: &first,
	}}}

	statusCode, _ := f.Rpc(&TestRpcRequest{Method: "lti_attempt.update_external", Params: input, Authorization: fix.authHeader})
	assert.Equal(t, 200, statusCode)

	replay := datatypes.NewJSONType(models.LTIAttemptResult{
		MaxScore: 10, CurrentScore: 1, ResultDisplay: "must be ignored", CheckID: "check-1", Report: "replayed report",
	})
	input.Models[0].Result = &replay
	statusCode, body := f.Rpc(&TestRpcRequest{Method: "lti_attempt.update_external", Params: input, Authorization: fix.authHeader})
	response := usecases.LTIAttemptEditBulkResponse{}
	_ = json.Unmarshal([]byte(body), &response)
	assert.Equal(t, 200, statusCode)
	assert.Equal(t, uint(1), response.Result.Count)

	stored := models.LTIAttempt{}
	f.DB.Where("attempt_id = ?", fix.attempts[1].AttemptID).First(&stored)
	assert.Equal(t, "check-1", stored.Result.Data().CheckID)
	assert.Equal(t, 8.0, stored.Result.Data().CurrentScore)
	assert.Equal(t, "first report", stored.Result.Data().Report)
}

func TestV1LTIAttemptUpdateExternalNotOwnedAttempt(t *testing.T) {
	f := NewTestHTTP()
	defer f.Close()

	fix := setupLTIAttemptFixture(f)

	otherServer := models.Server{}
	otherServer.Type = models.ServerTypeOpenID
	otherServer.Name = "Other Server"
	otherServer.Url = "https://other-localhost"
	otherServer.IsActive = true
	otherServer.Token = "other-token"
	otherServer.ClientID = "other-client-id"
	f.DB.Create(&otherServer)

	otherAttempt := models.LTIAttempt{}
	otherAttempt.AttemptID = "attempt-other-server"
	otherAttempt.Status = models.AttemptStatusActive
	otherAttempt.UserID = fix.user.ID
	otherAttempt.ServerID = &otherServer.ID
	otherAttempt.LTIRoutingID = fix.routing.ID
	f.DB.Create(&otherAttempt)

	input := usecases.LTIAttemptEditBulkInputDTO{
		Models: []usecases.LTIAttemptEditBulkInput{
			{
				AttemptID: "attempt-other-server",
				Status:    models.AttemptStatusCompleted,
			},
		},
	}

	expectedCode := 200
	statusCode, body := f.Rpc(&TestRpcRequest{
		Method:        "lti_attempt.update_external",
		Params:        input,
		Authorization: fix.authHeader,
	})
	bodyModel := usecases.LTIAttemptEditBulkResponse{}
	_ = json.Unmarshal([]byte(body), &bodyModel)

	assert.Equal(t, expectedCode, statusCode)
	assert.Equal(t, uint(0), bodyModel.Result.Count)
}

func TestV1LTIAttemptUpdateExternalNotFound(t *testing.T) {
	f := NewTestHTTP()
	defer f.Close()

	fix := setupLTIAttemptFixture(f)

	input := usecases.LTIAttemptEditBulkInputDTO{
		Models: []usecases.LTIAttemptEditBulkInput{
			{
				AttemptID: "non-existent-attempt-id",
				Status:    models.AttemptStatusCompleted,
			},
		},
	}

	expectedCode := 200
	statusCode, body := f.Rpc(&TestRpcRequest{
		Method:        "lti_attempt.update_external",
		Params:        input,
		Authorization: fix.authHeader,
	})
	bodyModel := usecases.LTIAttemptEditBulkResponse{}
	_ = json.Unmarshal([]byte(body), &bodyModel)

	assert.Equal(t, expectedCode, statusCode)
	assert.Equal(t, uint(0), bodyModel.Result.Count)
}

func TestV1LTIAttemptUpdateExternalStatusTransition(t *testing.T) {
	f := NewTestHTTP()
	defer f.Close()

	fix := setupLTIAttemptFixture(f)

	pendingAttempt := models.LTIAttempt{}
	pendingAttempt.AttemptID = "attempt-pending-transition"
	pendingAttempt.Status = models.AttemptStatusPending
	pendingAttempt.UserID = fix.user.ID
	pendingAttempt.ServerID = &fix.server.ID
	pendingAttempt.LTIRoutingID = fix.routing.ID
	f.DB.Create(&pendingAttempt)

	input := usecases.LTIAttemptEditBulkInputDTO{
		Models: []usecases.LTIAttemptEditBulkInput{
			{
				AttemptID: "attempt-pending-transition",
				Status:    models.AttemptStatusActive,
			},
		},
	}

	expectedCode := 200
	statusCode, body := f.Rpc(&TestRpcRequest{
		Method:        "lti_attempt.update_external",
		Params:        input,
		Authorization: fix.authHeader,
	})
	bodyModel := usecases.LTIAttemptEditBulkResponse{}
	_ = json.Unmarshal([]byte(body), &bodyModel)

	assert.Equal(t, expectedCode, statusCode)
	assert.Equal(t, uint(1), bodyModel.Result.Count)

	var updatedAttempt models.LTIAttempt
	f.DB.Where("attempt_id = ?", "attempt-pending-transition").First(&updatedAttempt)
	assert.Equal(t, models.AttemptStatusActive, updatedAttempt.Status)
}
