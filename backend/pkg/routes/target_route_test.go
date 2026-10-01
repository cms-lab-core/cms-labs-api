package routes

import (
	"testing"
	"time"

	"github.com/cms-lab-core/cms-labs-api/backend/app/models"
	"github.com/cms-lab-core/cms-labs-api/backend/app/models/types"
	"github.com/cms-lab-core/cms-labs-api/backend/app/usecases"
	json "github.com/goccy/go-json"
	"github.com/google/uuid"
	"github.com/h2non/gock"
	"github.com/stretchr/testify/assert"
)

func TestV1TargetUpsertCreate(t *testing.T) {
	desc := "create target"
	f := NewTestHTTP()
	defer f.Close()
	authAdmin := f.AuthorizationUser(0, nil)
	links := []models.TargetLink{
		{
			Value: "https://google.com",
			Type:  "doc",
		},
	}

	tags := []string{"prod"}

	input := usecases.TargetUpsertInputDTO{
		ID:          nil, // создание
		Name:        "Test Target " + uuid.New().String(),
		Description: ptrString("Test Description"),
		Type:        "service",
		Links:       &links,
		Tags:        &tags,
	}

	code, body := f.Rpc(&TestRpcRequest{
		Method:        "target.upsert",
		Params:        input,
		Authorization: authAdmin,
	})
	var resp usecases.TargetUpsertResponse
	err := json.Unmarshal([]byte(body), &resp)
	assert.NoError(t, err, desc)
	assert.Equal(t, 200, code, desc)
	assert.NotEmpty(t, resp.Result.ID, desc)

	// Проверка в БД
	var target models.Target
	err = f.DB.First(&target, "id = ?", resp.Result.ID).Error
	assert.NoError(t, err, desc)
	assert.Equal(t, input.Name, target.Name, desc)
	assert.Equal(t, *input.Description, *target.Description, desc)
	assert.Equal(t, (*input.Links)[0].Value, (*target.Links)[0].Value, desc)
	assert.Equal(t, (*input.Tags)[0], (*target.Tags)[0], desc)

	// Проверка, что создатель стал редактором
	var tu models.TargetUser
	err = f.DB.First(&tu, "target_id = ?", resp.Result.ID).Error
	assert.NoError(t, err, desc)
	roles, ok := tu.Roles["roles"].([]interface{})
	assert.True(t, ok, desc)
	assert.Contains(t, roles, models.UserRoleEditor, desc)
}

func TestV1TargetUpsertUpdate(t *testing.T) {
	desc := "update target"
	f := NewTestHTTP()
	defer f.Close()
	authAdmin := f.AuthorizationUser(0, nil)

	// Сначала создадим цель
	createInput := usecases.TargetUpsertInputDTO{
		ID:          nil,
		Name:        "Original" + uuid.New().String(),
		Description: ptrString("Original desc"),
		Type:        "service",
		Links:       nil,
		Tags:        nil,
	}
	_, body := f.Rpc(&TestRpcRequest{
		Method:        "target.upsert",
		Params:        createInput,
		Authorization: authAdmin,
	})
	var createResp usecases.TargetUpsertResponse
	_ = json.Unmarshal([]byte(body), &createResp)
	targetID := createResp.Result.ID

	links := []models.TargetLink{
		{
			Value: "https://google.com",
			Type:  "doc",
		},
	}

	tags := []string{"prod"}

	// Обновляем
	updateInput := usecases.TargetUpsertInputDTO{
		ID:          &targetID,
		Name:        "Updated" + uuid.New().String(),
		Description: ptrString("Updated desc"),
		Type:        "server",
		Links:       &links,
		Tags:        &tags,
	}

	code, body := f.Rpc(&TestRpcRequest{
		Method:        "target.upsert",
		Params:        updateInput,
		Authorization: authAdmin,
	})
	var updateResp usecases.TargetUpsertResponse
	err := json.Unmarshal([]byte(body), &updateResp)
	assert.NoError(t, err, desc)
	assert.Equal(t, 200, code, desc)
	assert.Equal(t, targetID, updateResp.Result.ID, desc)

	// Проверка в БД
	var target models.Target
	f.DB.First(&target, "id = ?", targetID)
	// Имя не является изменяемым
	assert.Equal(t, createInput.Name, target.Name, desc)
	assert.Equal(t, "Updated desc", *target.Description, desc)
}

func TestV1TargetUpsertUpdate_NotEditor(t *testing.T) {
	desc := "update target by non-editor should fail"
	f := NewTestHTTP()
	defer f.Close()
	authAdmin := f.AuthorizationUser(0, nil)
	regularUser := models.User{}
	regularUser.Email = "user" + uuid.New().String() + "@example.com"
	f.DB.Create(&regularUser)
	authStudent := f.AuthorizationUser(regularUser.ID, nil)

	// Создаём цель от админа
	createInput := usecases.TargetUpsertInputDTO{
		ID:          nil,
		Name:        "Target" + uuid.New().String(),
		Description: ptrString("desc"),
		Type:        "service",
	}
	_, body := f.Rpc(&TestRpcRequest{
		Method:        "target.upsert",
		Params:        createInput,
		Authorization: authAdmin,
	})
	var createResp usecases.TargetUpsertResponse
	_ = json.Unmarshal([]byte(body), &createResp)
	targetID := createResp.Result.ID

	// Пытаемся обновить от студента (не редактор)
	updateInput := usecases.TargetUpsertInputDTO{
		ID:          &targetID,
		Name:        "Hacked" + uuid.New().String(),
		Description: ptrString("evil"),
		Type:        "service",
	}
	code, body := f.Rpc(&TestRpcRequest{
		Method:        "target.upsert",
		Params:        updateInput,
		Authorization: authStudent,
	})
	assert.Equal(t, 500, code, desc) // код ответа 500, ошибка в JSON-RPC
	var rpcError struct {
		Error struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
			Data    struct {
				Code string `json:"code"`
			} `json:"data"`
		} `json:"error"`
	}
	_ = json.Unmarshal([]byte(body), &rpcError)
	assert.Equal(t, "permission_denied", rpcError.Error.Data.Code, desc)
}

func TestV1TargetUpsertCreate_ForbiddenRole(t *testing.T) {
	desc := "create target by student should be forbidden"
	f := NewTestHTTP()
	defer f.Close()
	regularUser := models.User{}
	regularUser.Email = "user" + uuid.New().String() + "@example.com"
	f.DB.Create(&regularUser)
	authStudent := f.AuthorizationUser(regularUser.ID, nil) // student

	input := usecases.TargetUpsertInputDTO{
		ID:          nil,
		Name:        "Target" + uuid.New().String(),
		Description: ptrString("desc"),
		Type:        "service",
	}
	code, body := f.Rpc(&TestRpcRequest{
		Method:        "target.upsert",
		Params:        input,
		Authorization: authStudent,
	})
	assert.Equal(t, 500, code, desc)
	var rpcError struct {
		Error struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
			Data    struct {
				Code string `json:"code"`
			} `json:"data"`
		} `json:"error"`
	}
	_ = json.Unmarshal([]byte(body), &rpcError)
	assert.Equal(t, "permission_denied", rpcError.Error.Data.Code, desc)
}

func TestV1TargetGet(t *testing.T) {
	desc := "get target by ID"
	f := NewTestHTTP()
	defer f.Close()
	auth := f.AuthorizationUser(0, nil)

	target := models.Target{
		ID:             uuid.New().String(),
		Name:           "Test Target" + uuid.New().String(),
		Description:    ptrString("Test Description"),
		Type:           "service",
		Links:          nil,
		Tags:           nil,
		SynchronizedAt: time.Now().UTC(),
		UpdatedAt:      time.Now().UTC(),
		CreatedAt:      time.Now().UTC(),
	}
	err := f.DB.Create(&target).Error
	assert.NoError(t, err, desc)

	input := usecases.TargetGetInputDTO{ID: target.ID}
	code, body := f.Rpc(&TestRpcRequest{
		Method:        "target.get",
		Params:        input,
		Authorization: auth,
	})
	var resp usecases.TargetGetResponse
	err = json.Unmarshal([]byte(body), &resp)
	assert.NoError(t, err, desc)
	assert.Equal(t, 200, code, desc)
	assert.Equal(t, target.ID, resp.Result.ID, desc)
	assert.Equal(t, target.Name, resp.Result.Name, desc)
	assert.Equal(t, *target.Description, *resp.Result.Description, desc)
}

func TestV1TargetDelete(t *testing.T) {
	desc := "delete target"
	f := NewTestHTTP()
	defer f.Close()
	regularUser := models.User{}
	regularUser.Email = "user" + uuid.New().String() + "@example.com"
	f.DB.Create(&regularUser)
	authUser := f.AuthorizationUser(regularUser.ID, nil)

	// Создаём цель
	target := models.Target{
		ID:             uuid.New().String(),
		Name:           "ToDelete" + uuid.New().String(),
		Description:    ptrString("desc"),
		Type:           "service",
		SynchronizedAt: time.Now().UTC(),
		CreatedAt:      time.Now().UTC(),
		UpdatedAt:      time.Now().UTC(),
	}
	f.DB.Create(&target)

	targetUser := models.TargetUser{
		TargetUserBase: models.TargetUserBase{
			TargetID: target.ID,
			UserID:   regularUser.ID,
			Roles: types.JsonStore{
				"roles": []string{models.UserRoleEditor},
			},
		},
	}
	f.DB.Create(&targetUser)

	input := usecases.TargetDeleteInputDTO{ID: target.ID}
	code, body := f.Rpc(&TestRpcRequest{
		Method:        "target.delete",
		Params:        input,
		Authorization: authUser,
	})
	var resp usecases.TargetDeleteResponse
	err := json.Unmarshal([]byte(body), &resp)
	assert.NoError(t, err, desc)
	assert.Equal(t, 200, code, desc)

	// Проверка удаления
	var count int64
	f.DB.Model(&models.Target{}).Where("id = ?", target.ID).Count(&count)
	assert.Equal(t, int64(0), count, desc)
}

func TestV1TargetDelete_NotEditor(t *testing.T) {
	desc := "delete target by non-editor should fail"
	f := NewTestHTTP()
	defer f.Close()
	authAdmin := f.AuthorizationUser(0, nil)
	regularUser := models.User{}
	regularUser.Email = "user" + uuid.New().String() + "@example.com"
	f.DB.Create(&regularUser)
	authStudent := f.AuthorizationUser(regularUser.ID, nil)

	// Создаём цель от админа (он становится редактором)
	createInput := usecases.TargetUpsertInputDTO{
		ID:          nil,
		Name:        "Target" + uuid.New().String(),
		Description: ptrString("desc"),
		Type:        "service",
	}
	_, body := f.Rpc(&TestRpcRequest{
		Method:        "target.upsert",
		Params:        createInput,
		Authorization: authAdmin,
	})
	var createResp usecases.TargetUpsertResponse
	_ = json.Unmarshal([]byte(body), &createResp)
	targetID := createResp.Result.ID

	// Пытаемся удалить от студента
	input := usecases.TargetDeleteInputDTO{ID: targetID}
	code, body := f.Rpc(&TestRpcRequest{
		Method:        "target.delete",
		Params:        input,
		Authorization: authStudent,
	})
	assert.Equal(t, 500, code, desc)
	var rpcError struct {
		Error struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
			Data    struct {
				Code string `json:"code"`
			} `json:"data"`
		} `json:"error"`
	}
	_ = json.Unmarshal([]byte(body), &rpcError)
	assert.Equal(t, "permission_denied", rpcError.Error.Data.Code, desc)
}

func TestV1TargetRelationCreate_Success(t *testing.T) {
	desc := "create target relation successfully"
	f := NewTestHTTP()
	defer f.Close()
	authAdmin := f.AuthorizationUser(0, nil)

	// Создаём две цели через usecase (редактор автоматически назначается)
	createTarget1 := usecases.TargetUpsertInputDTO{
		ID:          nil,
		Name:        "FromTarget " + uuid.New().String(),
		Description: ptrString("desc"),
		Type:        "service",
	}
	code1, body1 := f.Rpc(&TestRpcRequest{
		Method:        "target.upsert",
		Params:        createTarget1,
		Authorization: authAdmin,
	})
	var resp1 usecases.TargetUpsertResponse
	_ = json.Unmarshal([]byte(body1), &resp1)
	assert.Equal(t, 200, code1, desc)
	fromID := resp1.Result.ID

	createTarget2 := usecases.TargetUpsertInputDTO{
		ID:          nil,
		Name:        "ToTarget " + uuid.New().String(),
		Description: ptrString("desc"),
		Type:        "service",
	}
	code2, body2 := f.Rpc(&TestRpcRequest{
		Method:        "target.upsert",
		Params:        createTarget2,
		Authorization: authAdmin,
	})
	var resp2 usecases.TargetUpsertResponse
	_ = json.Unmarshal([]byte(body2), &resp2)
	assert.Equal(t, 200, code2, desc)
	toID := resp2.Result.ID

	// Создаём связь
	input := usecases.TargetRelationCreateInputDTO{
		FromTargetID: fromID,
		ToTargetID:   toID,
		RelationType: "depends_on",
	}
	code, body := f.Rpc(&TestRpcRequest{
		Method:        "target.relation_create",
		Params:        input,
		Authorization: authAdmin,
	})
	var resp usecases.TargetRelationCreateResponse
	err := json.Unmarshal([]byte(body), &resp)
	assert.NoError(t, err, desc)
	assert.Equal(t, 200, code, desc)
	assert.NotZero(t, resp.Result.ID, desc)

	// Проверяем в БД
	var relation models.TargetRelation
	err = f.DB.First(&relation, "id = ?", resp.Result.ID).Error
	assert.NoError(t, err, desc)
	assert.Equal(t, fromID, relation.FromTargetID, desc)
	assert.Equal(t, toID, relation.ToTargetID, desc)
	assert.Equal(t, "depends_on", relation.RelationType, desc)
}

func TestV1TargetRelationCreate_NotEditor(t *testing.T) {
	desc := "create target relation by non-editor should fail"
	f := NewTestHTTP()
	defer f.Close()
	authAdmin := f.AuthorizationUser(0, nil)

	// Создаём обычного пользователя
	regularUser := models.User{
		UserBase: models.UserBase{
			Email: "user" + uuid.New().String() + "@example.com",
		},
	}
	f.DB.Create(&regularUser)
	authRegular := f.AuthorizationUser(regularUser.ID, nil)

	// Админ создаёт цель (становится редактором)
	createTarget := usecases.TargetUpsertInputDTO{
		ID:          nil,
		Name:        "FromTarget " + uuid.New().String(),
		Description: ptrString("desc"),
		Type:        "service",
	}
	_, body := f.Rpc(&TestRpcRequest{
		Method:        "target.upsert",
		Params:        createTarget,
		Authorization: authAdmin,
	})
	var createResp usecases.TargetUpsertResponse
	_ = json.Unmarshal([]byte(body), &createResp)
	fromID := createResp.Result.ID

	// Вторая цель (можно создать через БД для простоты)
	toTarget := models.Target{
		ID:   uuid.New().String(),
		Name: "ToTarget " + uuid.New().String(),
		Type: "service",
	}
	f.DB.Create(&toTarget)

	// Обычный пользователь пытается создать связь
	input := usecases.TargetRelationCreateInputDTO{
		FromTargetID: fromID,
		ToTargetID:   toTarget.ID,
		RelationType: "depends_on",
	}
	code, body := f.Rpc(&TestRpcRequest{
		Method:        "target.relation_create",
		Params:        input,
		Authorization: authRegular,
	})
	assert.Equal(t, 500, code, desc)
	var rpcError struct {
		Error struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
			Data    struct {
				Code string `json:"code"`
			} `json:"data"`
		} `json:"error"`
	}
	_ = json.Unmarshal([]byte(body), &rpcError)
	assert.Equal(t, "permission_denied", rpcError.Error.Data.Code, desc)
}

func TestV1TargetRelationCreate_SelfRelation(t *testing.T) {
	desc := "create relation with from = to should fail"
	f := NewTestHTTP()
	defer f.Close()
	authAdmin := f.AuthorizationUser(0, nil)

	// Создаём цель
	createTarget := usecases.TargetUpsertInputDTO{
		ID:          nil,
		Name:        "Target " + uuid.New().String(),
		Description: ptrString("desc"),
		Type:        "service",
	}
	_, body := f.Rpc(&TestRpcRequest{
		Method:        "target.upsert",
		Params:        createTarget,
		Authorization: authAdmin,
	})
	var createResp usecases.TargetUpsertResponse
	_ = json.Unmarshal([]byte(body), &createResp)
	targetID := createResp.Result.ID

	// Пытаемся создать связь саму на себя
	input := usecases.TargetRelationCreateInputDTO{
		FromTargetID: targetID,
		ToTargetID:   targetID,
		RelationType: "depends_on",
	}
	code, body := f.Rpc(&TestRpcRequest{
		Method:        "target.relation_create",
		Params:        input,
		Authorization: authAdmin,
	})
	assert.Equal(t, 500, code, desc)
	var rpcError struct {
		Error struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
			Data    struct {
				Code string `json:"code"`
			} `json:"data"`
		} `json:"error"`
	}
	_ = json.Unmarshal([]byte(body), &rpcError)
	assert.Equal(t, "invalid_arguments", rpcError.Error.Data.Code, desc)
}

func TestV1TargetRelationCreate_Duplicate(t *testing.T) {
	desc := "create duplicate relation (unique constraint) should fail or update"
	f := NewTestHTTP()
	defer f.Close()
	authAdmin := f.AuthorizationUser(0, nil)

	// Создаём две цели
	createTarget1 := usecases.TargetUpsertInputDTO{
		ID:          nil,
		Name:        "FromTarget " + uuid.New().String(),
		Description: ptrString("desc"),
		Type:        "service",
	}
	_, body1 := f.Rpc(&TestRpcRequest{
		Method:        "target.upsert",
		Params:        createTarget1,
		Authorization: authAdmin,
	})
	var resp1 usecases.TargetUpsertResponse
	_ = json.Unmarshal([]byte(body1), &resp1)
	fromID := resp1.Result.ID

	createTarget2 := usecases.TargetUpsertInputDTO{
		ID:          nil,
		Name:        "ToTarget " + uuid.New().String(),
		Description: ptrString("desc"),
		Type:        "service",
	}
	_, body2 := f.Rpc(&TestRpcRequest{
		Method:        "target.upsert",
		Params:        createTarget2,
		Authorization: authAdmin,
	})
	var resp2 usecases.TargetUpsertResponse
	_ = json.Unmarshal([]byte(body2), &resp2)
	toID := resp2.Result.ID

	// Создаём связь первый раз
	input := usecases.TargetRelationCreateInputDTO{
		FromTargetID: fromID,
		ToTargetID:   toID,
		RelationType: "depends_on",
	}
	code, body := f.Rpc(&TestRpcRequest{
		Method:        "target.relation_create",
		Params:        input,
		Authorization: authAdmin,
	})
	var resp usecases.TargetRelationCreateResponse
	err := json.Unmarshal([]byte(body), &resp)
	assert.NoError(t, err, desc)
	assert.Equal(t, 200, code, desc)

	// Пытаемся создать точно такую же связь (дубликат)
	code, _ = f.Rpc(&TestRpcRequest{
		Method:        "target.relation_create",
		Params:        input,
		Authorization: authAdmin,
	})
	// Идемпотентая операция
	assert.Equal(t, 200, code, desc)
	var count int64
	f.DB.Model(&models.TargetRelation{}).Where("from_target_id = ? AND to_target_id = ? AND relation_type = ?", fromID, toID, "depends_on").Count(&count)
	assert.Equal(t, int64(1), count, desc+": duplicate should not create second record")
}

func TestV1TargetRelationDelete_Success(t *testing.T) {
	desc := "delete target relation successfully"
	f := NewTestHTTP()
	defer f.Close()
	regularUser := models.User{}
	regularUser.Email = "user" + uuid.New().String() + "@example.com"
	f.DB.Create(&regularUser)
	authHeader := f.AuthorizationUser(regularUser.ID, nil)

	// Создаём две цели и связь
	// Для упрощения создадим цели через БД, но добавим пользователя как редактора
	fromTarget := models.Target{
		ID:             uuid.New().String(),
		Name:           "FromTarget " + uuid.New().String(),
		Type:           "service",
		SynchronizedAt: time.Now().UTC(),
		CreatedAt:      time.Now().UTC(),
		UpdatedAt:      time.Now().UTC(),
	}
	f.DB.Create(&fromTarget)
	toTarget := models.Target{
		ID:             uuid.New().String(),
		Name:           "ToTarget " + uuid.New().String(),
		Type:           "service",
		SynchronizedAt: time.Now().UTC(),
		CreatedAt:      time.Now().UTC(),
		UpdatedAt:      time.Now().UTC(),
	}
	f.DB.Create(&toTarget)

	// Добавляем пользователя как редактора fromTarget
	targetUser := models.TargetUser{
		TargetUserBase: models.TargetUserBase{
			TargetID: fromTarget.ID,
			UserID:   regularUser.ID,
			Roles: types.JsonStore{
				"roles": []string{models.UserRoleEditor},
			},
		},
	}
	f.DB.Create(&targetUser)

	// Создаём связь
	relation := models.TargetRelation{
		FromTargetID: fromTarget.ID,
		ToTargetID:   toTarget.ID,
		RelationType: "depends_on",
	}
	f.DB.Create(&relation)

	// Удаляем
	input := usecases.TargetRelationDeleteInputDTO{
		FromTargetID: fromTarget.ID,
		ToTargetID:   toTarget.ID,
		RelationType: "depends_on",
	}
	code, body := f.Rpc(&TestRpcRequest{
		Method:        "target.relation_delete",
		Params:        input,
		Authorization: authHeader,
	})
	var resp usecases.TargetRelationDeleteResponse
	err := json.Unmarshal([]byte(body), &resp)
	assert.NoError(t, err, desc)
	assert.Equal(t, 200, code, desc)
	assert.Equal(t, fromTarget.ID, resp.Result.FromTargetID, desc)
	assert.Equal(t, toTarget.ID, resp.Result.ToTargetID, desc)
	assert.Equal(t, "depends_on", resp.Result.RelationType, desc)

	// Проверяем, что связь удалена
	var count int64
	f.DB.Model(&models.TargetRelation{}).Where("id = ?", relation.ID).Count(&count)
	assert.Equal(t, int64(0), count, desc)
}

func TestV1TargetRelationDelete_NotEditor(t *testing.T) {
	desc := "delete target relation by non-editor should fail"
	f := NewTestHTTP()
	defer f.Close()

	// Создаём обычного пользователя
	regularUser := models.User{
		UserBase: models.UserBase{
			Email: "user" + uuid.New().String() + "@example.com",
		},
	}
	f.DB.Create(&regularUser)
	authRegular := f.AuthorizationUser(regularUser.ID, nil)

	// Создаём цели через БД
	fromTarget := models.Target{
		ID:             uuid.New().String(),
		Name:           "FromTarget " + uuid.New().String(),
		Type:           "service",
		SynchronizedAt: time.Now().UTC(),
		CreatedAt:      time.Now().UTC(),
		UpdatedAt:      time.Now().UTC(),
	}
	f.DB.Create(&fromTarget)
	toTarget := models.Target{
		ID:             uuid.New().String(),
		Name:           "ToTarget " + uuid.New().String(),
		Type:           "service",
		SynchronizedAt: time.Now().UTC(),
		CreatedAt:      time.Now().UTC(),
		UpdatedAt:      time.Now().UTC(),
	}
	f.DB.Create(&toTarget)

	// Админ становится редактором fromTarget
	targetUser := models.TargetUser{
		TargetUserBase: models.TargetUserBase{
			TargetID: fromTarget.ID,
			UserID:   0, // ID админа
			Roles: types.JsonStore{
				"roles": []string{models.UserRoleEditor},
			},
		},
	}
	f.DB.Create(&targetUser)

	// Создаём связь
	relation := models.TargetRelation{
		FromTargetID: fromTarget.ID,
		ToTargetID:   toTarget.ID,
		RelationType: "depends_on",
	}
	f.DB.Create(&relation)

	// Обычный пользователь пытается удалить
	input := usecases.TargetRelationDeleteInputDTO{
		FromTargetID: fromTarget.ID,
		ToTargetID:   toTarget.ID,
		RelationType: "depends_on",
	}
	code, body := f.Rpc(&TestRpcRequest{
		Method:        "target.relation_delete",
		Params:        input,
		Authorization: authRegular,
	})
	assert.Equal(t, 500, code, desc)
	var rpcError struct {
		Error struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
			Data    struct {
				Code string `json:"code"`
			} `json:"data"`
		} `json:"error"`
	}
	_ = json.Unmarshal([]byte(body), &rpcError)
	assert.Equal(t, "permission_denied", rpcError.Error.Data.Code, desc)

	// Проверяем, что связь не удалена
	var count int64
	f.DB.Model(&models.TargetRelation{}).Where("id = ?", relation.ID).Count(&count)
	assert.Equal(t, int64(1), count, desc)
}

func TestV1TargetRelationDelete_PermissionDenied(t *testing.T) {
	desc := "delete with empty fields should fail validation"
	f := NewTestHTTP()
	defer f.Close()
	authAdmin := f.AuthorizationUser(0, nil)

	// Пустой FromTargetID
	input := usecases.TargetRelationDeleteInputDTO{
		FromTargetID: "",
		ToTargetID:   "some-id",
		RelationType: "depends_on",
	}
	code, body := f.Rpc(&TestRpcRequest{
		Method:        "target.relation_delete",
		Params:        input,
		Authorization: authAdmin,
	})
	assert.Equal(t, 500, code, desc) // ожидаем ошибку валидации
	var rpcError struct {
		Error struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
			Data    struct {
				Code string `json:"code"`
			} `json:"data"`
		} `json:"error"`
	}
	_ = json.Unmarshal([]byte(body), &rpcError)
	assert.Equal(t, "permission_denied", rpcError.Error.Data.Code, desc)
}

// -----------------------------------------------------------------------------
// TargetUserUpsert
// -----------------------------------------------------------------------------

func TestV1TargetUserUpsert_CreateFirstEditor(t *testing.T) {
	desc := "create first target_user (editor) when no editors exist"
	f := NewTestHTTP()
	defer f.Close()
	regularUser := models.User{}
	regularUser.Email = "user" + uuid.New().String() + "@example.com"
	f.DB.Create(&regularUser)
	authHeader := f.AuthorizationUser(regularUser.ID, nil)

	// Создаём цель (через БД, чтобы не создавать редактора автоматически)
	target := models.Target{
		ID:             uuid.New().String(),
		Name:           "Target " + uuid.New().String(),
		Type:           "service",
		SynchronizedAt: time.Now().UTC(),
		CreatedAt:      time.Now().UTC(),
		UpdatedAt:      time.Now().UTC(),
	}
	f.DB.Create(&target)

	// У цели пока нет редакторов
	input := usecases.TargetUserUpsertInputDTO{
		TargetID: target.ID,
		UserID:   regularUser.ID,
		Roles:    []string{models.UserRoleEditor},
	}
	code, body := f.Rpc(&TestRpcRequest{
		Method:        "target.user_upsert",
		Params:        input,
		Authorization: authHeader,
	})
	var resp usecases.TargetUserUpsertResponse
	err := json.Unmarshal([]byte(body), &resp)
	assert.NoError(t, err, desc)
	assert.Equal(t, 200, code, desc)
	assert.NotZero(t, resp.Result.ID, desc)

	// Проверяем в БД
	var tu models.TargetUser
	err = f.DB.First(&tu, "id = ?", resp.Result.ID).Error
	assert.NoError(t, err, desc)
	assert.Equal(t, target.ID, tu.TargetID, desc)
	assert.Equal(t, regularUser.ID, tu.UserID, desc)
	roles, ok := tu.Roles["roles"].([]interface{})
	assert.True(t, ok, desc)
	assert.Contains(t, roles, models.UserRoleEditor, desc)
}

func TestV1TargetUserUpsert_CreateAnotherEditor_AsEditor(t *testing.T) {
	desc := "create another editor when current user is an editor"
	defer gock.Off()
	gock.New("http://localhost:8200").
		Get("/v1/sys/auth").
		Reply(200).
		JSON(map[string]interface{}{
			"oidc/": map[string]interface{}{
				"type":        "oidc",
				"accessor":    "auth_oidc_abc123",
				"config":      map[string]interface{}{},
				"description": "OIDC auth method",
			},
		})
	gock.New("http://localhost:8200").
		Post("/v1/identity/lookup/entity").
		Reply(404).
		JSON(map[string]interface{}{})
	gock.New("http://localhost:8200").
		Post("/v1/identity/entity").
		Reply(200).
		JSON(map[string]interface{}{
			"data": map[string]interface{}{
				"id": "entity-123",
			},
		})
	gock.New("http://localhost:8200").
		Post("/v1/identity/entity-alias").
		Reply(200).
		JSON(map[string]interface{}{
			"data": map[string]interface{}{
				"id": "alias-123",
			},
		})

	f := NewTestHTTP()
	defer f.Close()
	authUser := f.AuthorizationUser(0, nil) // редактор

	// Создаём цель
	target := models.Target{
		ID:             uuid.New().String(),
		Name:           "Target " + uuid.New().String(),
		Type:           "service",
		SynchronizedAt: time.Now().UTC(),
		CreatedAt:      time.Now().UTC(),
		UpdatedAt:      time.Now().UTC(),
	}
	f.DB.Create(&target)

	// Добавляем текущего пользователя как редактора
	editor := models.TargetUser{
		TargetUserBase: models.TargetUserBase{
			TargetID: target.ID,
			UserID:   0,
			Roles: types.JsonStore{
				"roles": []string{models.UserRoleEditor},
			},
		},
	}
	f.DB.Create(&editor)

	// Создаём другого пользователя
	otherUser := models.User{}
	otherUser.Email = "other" + uuid.New().String() + "@example.com"
	f.DB.Create(&otherUser)

	input := usecases.TargetUserUpsertInputDTO{
		TargetID: target.ID,
		UserID:   otherUser.ID,
		Roles:    []string{models.UserRoleEditor},
	}
	code, body := f.Rpc(&TestRpcRequest{
		Method:        "target.user_upsert",
		Params:        input,
		Authorization: authUser,
	})
	var resp usecases.TargetUserUpsertResponse
	err := json.Unmarshal([]byte(body), &resp)
	assert.NoError(t, err, desc)
	assert.Equal(t, 200, code, desc)
	assert.NotZero(t, resp.Result.ID, desc)

	// Проверяем
	var tu models.TargetUser
	err = f.DB.First(&tu, "id = ?", resp.Result.ID).Error
	assert.NoError(t, err, desc)
	assert.Equal(t, otherUser.ID, tu.UserID, desc)
}

func TestV1TargetUserUpsert_CreateAnotherEditor_NotEditor(t *testing.T) {
	desc := "create another editor when current user is not an editor -> should fail"
	f := NewTestHTTP()
	defer f.Close()
	firstUser := models.User{}
	firstUser.Email = "notEditor" + uuid.New().String() + "@example.com"
	f.DB.Create(&firstUser)
	authUser := f.AuthorizationUser(firstUser.ID, nil) // этот пользователь не редактор

	// Создаём цель
	target := models.Target{
		ID:             uuid.New().String(),
		Name:           "Target " + uuid.New().String(),
		Type:           "service",
		SynchronizedAt: time.Now().UTC(),
		CreatedAt:      time.Now().UTC(),
		UpdatedAt:      time.Now().UTC(),
	}
	f.DB.Create(&target)

	// Добавляем другого редактора (не текущего пользователя)
	otherUser := models.User{}
	otherUser.Email = "editor" + uuid.New().String() + "@example.com"
	f.DB.Create(&otherUser)
	editor := models.TargetUser{
		TargetUserBase: models.TargetUserBase{
			TargetID: target.ID,
			UserID:   otherUser.ID,
			Roles: types.JsonStore{
				"roles": []string{models.UserRoleEditor},
			},
		},
	}
	f.DB.Create(&editor)

	// Текущий пользователь пытается добавить ещё одного редактора
	input := usecases.TargetUserUpsertInputDTO{
		TargetID: target.ID,
		UserID:   firstUser.ID, // пробуем добавить самого себя? или другого
		Roles:    []string{models.UserRoleEditor},
	}
	code, body := f.Rpc(&TestRpcRequest{
		Method:        "target.user_upsert",
		Params:        input,
		Authorization: authUser,
	})
	assert.Equal(t, 500, code, desc)
	var rpcError struct {
		Error struct {
			Data struct {
				Code string `json:"code"`
			} `json:"data"`
		} `json:"error"`
	}
	_ = json.Unmarshal([]byte(body), &rpcError)
	assert.Equal(t, "permission_denied", rpcError.Error.Data.Code, desc)
}

func TestV1TargetUserUpsert_UpdateOwnRoles_AsEditor(t *testing.T) {
	desc := "editor updates own roles"
	defer gock.Off()
	gock.New("http://localhost:8200").
		Get("/v1/sys/auth").
		Reply(200).
		JSON(map[string]interface{}{
			"oidc/": map[string]interface{}{
				"type":        "oidc",
				"accessor":    "auth_oidc_abc123",
				"config":      map[string]interface{}{},
				"description": "OIDC auth method",
			},
		})
	gock.New("http://localhost:8200").
		Post("/v1/identity/lookup/entity").
		Reply(200).
		JSON(map[string]interface{}{
			"data": map[string]interface{}{
				"id": "entity-123",
			},
		})
	gock.New("http://localhost:8200").
		Post("/v1/identity/entity/id/entity-123").
		Reply(200).
		JSON(map[string]interface{}{
			"data": map[string]interface{}{
				"id": "entity-123",
			},
		})

	f := NewTestHTTP()
	defer f.Close()
	regularUser := models.User{}
	regularUser.Email = "user" + uuid.New().String() + "@example.com"
	f.DB.Create(&regularUser)
	authHeader := f.AuthorizationUser(regularUser.ID, nil)

	// Создаём цель
	target := models.Target{
		ID:             uuid.New().String(),
		Name:           "Target " + uuid.New().String(),
		Type:           "service",
		SynchronizedAt: time.Now().UTC(),
		CreatedAt:      time.Now().UTC(),
		UpdatedAt:      time.Now().UTC(),
	}
	f.DB.Create(&target)

	// Добавляем текущего пользователя как редактора
	editor := models.TargetUser{
		TargetUserBase: models.TargetUserBase{
			TargetID: target.ID,
			UserID:   regularUser.ID,
			Roles: types.JsonStore{
				"roles": []string{models.UserRoleEditor},
			},
		},
	}
	f.DB.Create(&editor)

	// Обновляем роли (добавляем viewer)
	input := usecases.TargetUserUpsertInputDTO{
		TargetID: target.ID,
		UserID:   regularUser.ID,
		Roles:    []string{models.UserRoleEditor, models.UserRoleNominal},
	}
	code, body := f.Rpc(&TestRpcRequest{
		Method:        "target.user_upsert",
		Params:        input,
		Authorization: authHeader,
	})
	var resp usecases.TargetUserUpsertResponse
	err := json.Unmarshal([]byte(body), &resp)
	assert.NoError(t, err, desc)
	assert.Equal(t, 200, code, desc)
	assert.Equal(t, editor.ID, resp.Result.ID, desc)

	// Проверяем обновлённые роли
	var tu models.TargetUser
	f.DB.First(&tu, editor.ID)
	roles, ok := tu.Roles["roles"].([]interface{})
	assert.True(t, ok, desc)
	assert.ElementsMatch(t, []interface{}{models.UserRoleEditor, models.UserRoleNominal}, roles, desc)
}

func TestV1TargetUserUpsert_InvalidInput(t *testing.T) {
	desc := "upsert with missing fields should fail validation"
	f := NewTestHTTP()
	defer f.Close()
	authUser := f.AuthorizationUser(0, nil)

	// Пропускаем обязательное поле UserID
	input := usecases.TargetUserUpsertInputDTO{
		TargetID: "some-id",
		Roles:    []string{models.UserRoleEditor},
	}
	code, body := f.Rpc(&TestRpcRequest{
		Method:        "target.user_upsert",
		Params:        input,
		Authorization: authUser,
	})
	assert.Equal(t, 500, code, desc)
	// Проверяем, что ошибка валидации (скорее всего, jsonrpc.ValidatorBase вернёт ошибку)
	var rpcError struct {
		Error struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	_ = json.Unmarshal([]byte(body), &rpcError)
	assert.Contains(t, rpcError.Error.Message, "validation", desc)
}

// -----------------------------------------------------------------------------
// TargetUserDelete
// -----------------------------------------------------------------------------

func TestV1TargetUserDelete_Success(t *testing.T) {
	desc := "editor deletes target_user"
	defer gock.Off()
	gock.New("http://localhost:8200").
		Get("/v1/sys/auth").
		Reply(200).
		JSON(map[string]interface{}{
			"oidc/": map[string]interface{}{
				"type":        "oidc",
				"accessor":    "auth_oidc_abc123",
				"config":      map[string]interface{}{},
				"description": "OIDC auth method",
			},
		})
	gock.New("http://localhost:8200").
		Post("/v1/identity/lookup/entity").
		Reply(200).
		JSON(map[string]interface{}{
			"data": map[string]interface{}{
				"id": "entity-123",
			},
		})
	gock.New("http://localhost:8200").
		Post("/v1/identity/entity/id/entity-123").
		Reply(200).
		JSON(map[string]interface{}{
			"data": map[string]interface{}{
				"id": "entity-123",
			},
		})

	f := NewTestHTTP()
	defer f.Close()
	regularUser := models.User{}
	regularUser.Email = "user" + uuid.New().String() + "@example.com"
	f.DB.Create(&regularUser)
	authHeader := f.AuthorizationUser(regularUser.ID, nil)

	// Создаём цель
	target := models.Target{
		ID:             uuid.New().String(),
		Name:           "Target " + uuid.New().String(),
		Type:           "service",
		SynchronizedAt: time.Now().UTC(),
		CreatedAt:      time.Now().UTC(),
		UpdatedAt:      time.Now().UTC(),
	}
	f.DB.Create(&target)

	// Добавляем текущего пользователя как редактора
	editor := models.TargetUser{
		TargetUserBase: models.TargetUserBase{
			TargetID: target.ID,
			UserID:   regularUser.ID,
			Roles: types.JsonStore{
				"roles": []string{models.UserRoleEditor},
			},
		},
	}
	f.DB.Create(&editor)

	// Создаём другого пользователя, которого будем удалять
	otherUser := models.User{}
	otherUser.Email = "other" + uuid.New().String() + "@example.com"
	f.DB.Create(&otherUser)
	otherTU := models.TargetUser{
		TargetUserBase: models.TargetUserBase{
			TargetID: target.ID,
			UserID:   otherUser.ID,
			Roles: types.JsonStore{
				"roles": []string{models.UserRoleNominal},
			},
		},
	}
	f.DB.Create(&otherTU)

	// Удаляем otherTU
	input := usecases.TargetUserDeleteInputDTO{
		TargetID: target.ID,
		UserID:   otherUser.ID,
	}
	code, body := f.Rpc(&TestRpcRequest{
		Method:        "target.user_delete",
		Params:        input,
		Authorization: authHeader,
	})
	var resp usecases.TargetUserDeleteResponse
	err := json.Unmarshal([]byte(body), &resp)
	assert.NoError(t, err, desc)
	assert.Equal(t, 200, code, desc)

	// Проверяем, что запись удалена
	var count int64
	f.DB.Model(&models.TargetUser{}).Where("id = ?", otherTU.ID).Count(&count)
	assert.Equal(t, int64(0), count, desc)
}

func TestV1TargetUserDelete_NotEditor(t *testing.T) {
	desc := "non-editor tries to delete target_user -> should fail"
	f := NewTestHTTP()
	defer f.Close()
	authUser := f.AuthorizationUser(0, nil) // этот пользователь не редактор

	// Создаём цель
	target := models.Target{
		ID:             uuid.New().String(),
		Name:           "Target " + uuid.New().String(),
		Type:           "service",
		SynchronizedAt: time.Now().UTC(),
		CreatedAt:      time.Now().UTC(),
		UpdatedAt:      time.Now().UTC(),
	}
	f.DB.Create(&target)

	// Добавляем другого редактора
	editorUser := models.User{}
	editorUser.Email = "editor" + uuid.New().String() + "@example.com"
	f.DB.Create(&editorUser)
	editor := models.TargetUser{
		TargetUserBase: models.TargetUserBase{
			TargetID: target.ID,
			UserID:   editorUser.ID,
			Roles: types.JsonStore{
				"roles": []string{models.UserRoleEditor},
			},
		},
	}
	f.DB.Create(&editor)

	// Создаём запись для удаления (принадлежит тому же редактору)
	toDelete := models.TargetUser{
		TargetUserBase: models.TargetUserBase{
			TargetID: target.ID,
			UserID:   editorUser.ID, // самого редактора
			Roles:    types.JsonStore{"roles": []string{models.UserRoleEditor}},
		},
	}
	f.DB.Create(&toDelete)

	// Текущий пользователь пытается удалить
	input := usecases.TargetUserDeleteInputDTO{
		TargetID: target.ID,
		UserID:   editorUser.ID,
	}
	code, body := f.Rpc(&TestRpcRequest{
		Method:        "target.user_delete",
		Params:        input,
		Authorization: authUser,
	})
	assert.Equal(t, 500, code, desc)
	var rpcError struct {
		Error struct {
			Data struct {
				Code string `json:"code"`
			} `json:"data"`
		} `json:"error"`
	}
	_ = json.Unmarshal([]byte(body), &rpcError)
	assert.Equal(t, "permission_denied", rpcError.Error.Data.Code, desc)

	// Запись должна остаться
	var count int64
	f.DB.Model(&models.TargetUser{}).Where("id = ?", toDelete.ID).Count(&count)
	assert.Equal(t, int64(1), count, desc)
}

func TestV1TargetUserDelete_NonExistent(t *testing.T) {
	desc := "delete non-existent target_user should succeed (idempotent)"
	defer gock.Off()
	gock.New("http://localhost:8200").
		Get("/v1/sys/auth").
		Reply(200).
		JSON(map[string]interface{}{
			"oidc/": map[string]interface{}{
				"type":        "oidc",
				"accessor":    "auth_oidc_abc123",
				"config":      map[string]interface{}{},
				"description": "OIDC auth method",
			},
		})
	gock.New("http://localhost:8200").
		Post("/v1/identity/lookup/entity").
		Reply(200).
		JSON(map[string]interface{}{
			"data": map[string]interface{}{
				"id": "entity-123",
			},
		})
	gock.New("http://localhost:8200").
		Post("/v1/identity/entity/id/entity-123").
		Reply(200).
		JSON(map[string]interface{}{
			"data": map[string]interface{}{
				"id": "entity-123",
			},
		})

	f := NewTestHTTP()
	defer f.Close()
	regularUser := models.User{}
	regularUser.Email = "user" + uuid.New().String() + "@example.com"
	f.DB.Create(&regularUser)
	authHeader := f.AuthorizationUser(regularUser.ID, nil)

	// Создаём цель и делаем текущего пользователя редактором
	target := models.Target{
		ID:             uuid.New().String(),
		Name:           "Target " + uuid.New().String(),
		Type:           "service",
		SynchronizedAt: time.Now().UTC(),
		CreatedAt:      time.Now().UTC(),
		UpdatedAt:      time.Now().UTC(),
	}
	f.DB.Create(&target)
	editor := models.TargetUser{
		TargetUserBase: models.TargetUserBase{
			TargetID: target.ID,
			UserID:   regularUser.ID,
			Roles:    types.JsonStore{"roles": []string{models.UserRoleEditor}},
		},
	}
	f.DB.Create(&editor)

	// Пытаемся удалить несуществующего пользователя
	nonExistentUserID := uint(999999)
	input := usecases.TargetUserDeleteInputDTO{
		TargetID: target.ID,
		UserID:   nonExistentUserID,
	}
	code, body := f.Rpc(&TestRpcRequest{
		Method:        "target.user_delete",
		Params:        input,
		Authorization: authHeader,
	})
	assert.Equal(t, 200, code, desc)
	var resp usecases.TargetUserDeleteResponse
	err := json.Unmarshal([]byte(body), &resp)
	assert.NoError(t, err, desc)
}

func TestV1TargetUserDelete_DeleteLastEditor(t *testing.T) {
	desc := "delete the last editor of a target - should be allowed?"
	defer gock.Off()
	gock.New("http://localhost:8200").
		Get("/v1/sys/auth").
		Reply(200).
		JSON(map[string]interface{}{
			"oidc/": map[string]interface{}{
				"type":        "oidc",
				"accessor":    "auth_oidc_abc123",
				"config":      map[string]interface{}{},
				"description": "OIDC auth method",
			},
		})
	gock.New("http://localhost:8200").
		Post("/v1/identity/lookup/entity").
		Reply(200).
		JSON(map[string]interface{}{
			"data": map[string]interface{}{
				"id": "entity-123",
			},
		})
	gock.New("http://localhost:8200").
		Post("/v1/identity/entity/id/entity-123").
		Reply(200).
		JSON(map[string]interface{}{
			"data": map[string]interface{}{
				"id": "entity-123",
			},
		})

	f := NewTestHTTP()
	defer f.Close()
	regularUser := models.User{}
	regularUser.Email = "user" + uuid.New().String() + "@example.com"
	f.DB.Create(&regularUser)
	authHeader := f.AuthorizationUser(regularUser.ID, nil)

	// Создаём цель и единственного редактора (текущий пользователь)
	target := models.Target{
		ID:             uuid.New().String(),
		Name:           "Target " + uuid.New().String(),
		Type:           "service",
		SynchronizedAt: time.Now().UTC(),
		CreatedAt:      time.Now().UTC(),
		UpdatedAt:      time.Now().UTC(),
	}
	f.DB.Create(&target)
	editor := models.TargetUser{
		TargetUserBase: models.TargetUserBase{
			TargetID: target.ID,
			UserID:   regularUser.ID,
			Roles:    types.JsonStore{"roles": []string{models.UserRoleEditor}},
		},
	}
	f.DB.Create(&editor)

	// Удаляем самого себя (последнего редактора)
	input := usecases.TargetUserDeleteInputDTO{
		TargetID: target.ID,
		UserID:   regularUser.ID,
	}
	code, _ := f.Rpc(&TestRpcRequest{
		Method:        "target.user_delete",
		Params:        input,
		Authorization: authHeader,
	})
	assert.Equal(t, 200, code, desc)
	var count int64
	f.DB.Model(&models.TargetUser{}).Where("target_id = ?", target.ID).Count(&count)
	assert.Equal(t, int64(0), count, desc)
	// После удаления цель остаётся без редакторов (проверка на уровне бизнес-логики не требуется)
}

func TestV1TargetUserDelete_InvalidInput(t *testing.T) {
	desc := "delete with missing fields should fail validation"
	f := NewTestHTTP()
	defer f.Close()
	authUser := f.AuthorizationUser(0, nil)

	input := usecases.TargetUserDeleteInputDTO{
		TargetID: "some-id",
		// UserID отсутствует
	}
	code, body := f.Rpc(&TestRpcRequest{
		Method:        "target.user_delete",
		Params:        input,
		Authorization: authUser,
	})
	assert.Equal(t, 500, code, desc)
	var rpcError struct {
		Error struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	_ = json.Unmarshal([]byte(body), &rpcError)
	assert.Contains(t, rpcError.Error.Message, "validation", desc)
}

func ptrString(s string) *string {
	return &s
}
