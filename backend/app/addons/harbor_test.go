package addons

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"

	"github.com/h2non/gock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cms-lab-core/cms-labs-api/backend/app/addons/vault"
	"github.com/cms-lab-core/cms-labs-api/shared/connection"
)

func TestHarborAddonService_GetType(t *testing.T) {
	cfg := &connection.AddonConfig{Type: connection.HarborAddon}
	svc := NewHarborAddonService(cfg, nil).(*HarborAddonService)
	assert.Equal(t, connection.HarborAddon, svc.GetType())
}

func TestHarborAddonService_generateProjectName(t *testing.T) {
	svc := &HarborAddonService{}
	tests := []struct {
		input string
		want  string
	}{
		{"My Project", "svc_my_project"},
		{"Project 123", "svc_project_123"},
		{"a#b$c", "svc_a_b_c"},
		{strings.Repeat("x", 100), "svc_" + strings.Repeat("x", 59)}, // 63 total
	}
	for _, tt := range tests {
		got := svc.GenerateProjectNameForTest(tt.input)
		assert.Equal(t, tt.want, got)
	}
}

// Для тестирования неэкспортируемых методов можно использовать вспомогательные экспортируемые обёртки
func (s *HarborAddonService) GenerateProjectNameForTest(name string) string {
	return s.generateProjectName(name)
}

func TestHarborAddonService_urlPathEscape(t *testing.T) {
	assert.Equal(t, "repo%2Fname", urlPathEscape("repo/name"))
	assert.Equal(t, "simple", urlPathEscape("simple"))
}

func TestHarborAddonService_Create_Success(t *testing.T) {
	defer gock.Off()

	// Настройка моков HTTP
	gock.New("https://harbor.example.com").
		Post("/api/v2.0/projects").
		Reply(201)

	gock.New("https://harbor.example.com").
		Post("/api/v2.0/robots").
		Reply(201).
		JSON(map[string]interface{}{
			"id":     123,
			"name":   "robot$svc_test_project",
			"secret": "secret123",
		})

	gock.New("https://harbor.example.com").
		Put("/api/v2.0/projects/svc_test_target").
		Reply(200)

	// Мок Vault
	createCalled := false
	mockVault := &vault.MockVaultClient{
		CreateServiceExtensionFunc: func(ctx context.Context, serviceName, addonName string, data map[string]any) error {
			createCalled = true
			// Можно проверить аргументы, если нужно
			assert.Equal(t, "test-target", serviceName)
			assert.Equal(t, "harbor_test", addonName) // формируется как тип_тег
			assert.Equal(t, "secret123", data["HARBOR_REGISTRY_TOKEN"])
			return nil
		},
	}

	cfg := &connection.AddonConfig{
		Type: connection.HarborAddon,
		Tag:  "test",
		Params: map[string]interface{}{
			"apiUrl":        "https://harbor.example.com",
			"adminUsername": "admin",
			"adminPassword": "pass",
			"baseUrl":       "https://harbor.example.com",
		},
	}
	svc := NewHarborAddonService(cfg, mockVault).(*HarborAddonService)
	svc.HTTPClient = &http.Client{Transport: http.DefaultTransport} // gock перехватит

	ctx := context.Background()
	targetName := "test-target"
	result, err := svc.Create(ctx, targetName, nil)

	require.NoError(t, err)
	assert.Equal(t, "svc_test_target", result.Name)
	assert.NotNil(t, result.URL)
	assert.Equal(t, "https://harbor.example.com/harbor/projects?globalSearch=svc_test_target", *result.URL)

	assert.True(t, createCalled)
	assert.True(t, gock.IsDone())
}

func TestHarborAddonService_Create_ProjectExists(t *testing.T) {
	defer gock.Off()

	// Проект уже существует (409)
	gock.New("https://harbor.example.com").
		Post("/api/v2.0/projects").
		Reply(409)

	// Robot создаётся нормально
	gock.New("https://harbor.example.com").
		Post("/api/v2.0/robots").
		Reply(201).
		JSON(map[string]interface{}{"id": 1, "name": "robot$svc_test", "secret": "sec"})

	// Public access
	gock.New("https://harbor.example.com").
		Put("/api/v2.0/projects/svc_test").
		Reply(200)

	createCalled := false
	mockVault := &vault.MockVaultClient{
		CreateServiceExtensionFunc: func(ctx context.Context, serviceName, addonName string, data map[string]any) error {
			createCalled = true
			return nil
		},
	}

	cfg := &connection.AddonConfig{
		Type: connection.HarborAddon,
		Tag:  "test",
		Params: map[string]interface{}{
			"apiUrl":        "https://harbor.example.com",
			"adminUsername": "admin",
			"adminPassword": "pass",
			"baseUrl":       "https://harbor.example.com",
		},
	}
	svc := NewHarborAddonService(cfg, mockVault).(*HarborAddonService)
	svc.HTTPClient = &http.Client{Transport: http.DefaultTransport}

	ctx := context.Background()
	targetName := "test"
	result, err := svc.Create(ctx, targetName, nil)

	require.NoError(t, err)
	assert.Equal(t, "svc_test", result.Name)
	assert.True(t, createCalled)
	assert.True(t, gock.IsDone())
}

func TestHarborAddonService_Create_ProjectFails(t *testing.T) {
	defer gock.Off()

	// Ошибка создания проекта (не 409)
	gock.New("https://harbor.example.com").
		Post("/api/v2.0/projects").
		Reply(500)

	// Vault не должен вызываться
	createCalled := false
	mockVault := &vault.MockVaultClient{
		CreateServiceExtensionFunc: func(ctx context.Context, serviceName, addonName string, data map[string]any) error {
			createCalled = true
			return nil
		},
	}

	cfg := &connection.AddonConfig{
		Type: connection.HarborAddon,
		Tag:  "test",
		Params: map[string]interface{}{
			"apiUrl":        "https://harbor.example.com",
			"adminUsername": "admin",
			"adminPassword": "pass",
			"baseUrl":       "https://harbor.example.com",
		},
	}
	svc := NewHarborAddonService(cfg, mockVault).(*HarborAddonService)
	svc.HTTPClient = &http.Client{Transport: http.DefaultTransport}

	ctx := context.Background()
	_, err := svc.Create(ctx, "test", nil)
	assert.Error(t, err)
	assert.False(t, createCalled, "Vault should not be called")
	assert.True(t, gock.IsDone())
	// Проверяем, что cleanup не вызывался (проект не создан) — нет дополнительных запросов
}

func TestHarborAddonService_Create_RobotFails(t *testing.T) {
	defer gock.Off()

	// Проект создаётся успешно
	gock.New("https://harbor.example.com").
		Post("/api/v2.0/projects").
		Reply(201)

	// Robot fails
	gock.New("https://harbor.example.com").
		Post("/api/v2.0/robots").
		Reply(500)

	// Cleanup: удаление проекта (robot не создавался, поэтому только проект)
	gock.New("https://harbor.example.com").
		Delete("/api/v2.0/projects/svc_test").
		Reply(200)

	createCalled := false
	mockVault := &vault.MockVaultClient{
		CreateServiceExtensionFunc: func(ctx context.Context, serviceName, addonName string, data map[string]any) error {
			createCalled = true
			return nil
		},
	}

	cfg := &connection.AddonConfig{
		Type: connection.HarborAddon,
		Tag:  "test",
		Params: map[string]interface{}{
			"apiUrl":        "https://harbor.example.com",
			"adminUsername": "admin",
			"adminPassword": "pass",
			"baseUrl":       "https://harbor.example.com",
		},
	}
	svc := NewHarborAddonService(cfg, mockVault).(*HarborAddonService)
	svc.HTTPClient = &http.Client{Transport: http.DefaultTransport}

	ctx := context.Background()
	_, err := svc.Create(ctx, "test", nil)
	assert.Error(t, err)
	assert.False(t, createCalled, "Vault should not be called")
	assert.True(t, gock.IsDone())
}

func TestHarborAddonService_Create_SetPublicFails(t *testing.T) {
	defer gock.Off()

	// Проект создаётся успешно
	gock.New("https://harbor.example.com").
		Post("/api/v2.0/projects").
		Reply(201)

	// Robot успешно
	gock.New("https://harbor.example.com").
		Post("/api/v2.0/robots").
		Reply(201).
		JSON(map[string]interface{}{"id": 1, "name": "robot$svc_test", "secret": "sec"})

	// Set public fails
	gock.New("https://harbor.example.com").
		Put("/api/v2.0/projects/svc_test").
		Reply(500)

	// Cleanup: удалить robot и project
	gock.New("https://harbor.example.com").
		Get("/api/v2.0/robots").
		Reply(200).
		JSON([]interface{}{map[string]interface{}{"id": 1}})

	gock.New("https://harbor.example.com").
		Delete("/api/v2.0/robots/1").
		Reply(200)

	gock.New("https://harbor.example.com").
		Delete("/api/v2.0/projects/svc_test").
		Reply(200)

	createCalled := false
	mockVault := &vault.MockVaultClient{
		CreateServiceExtensionFunc: func(ctx context.Context, serviceName, addonName string, data map[string]any) error {
			createCalled = true
			return nil
		},
	}

	cfg := &connection.AddonConfig{
		Type: connection.HarborAddon,
		Tag:  "test",
		Params: map[string]interface{}{
			"apiUrl":        "https://harbor.example.com",
			"adminUsername": "admin",
			"adminPassword": "pass",
			"baseUrl":       "https://harbor.example.com",
		},
	}
	svc := NewHarborAddonService(cfg, mockVault).(*HarborAddonService)
	svc.HTTPClient = &http.Client{Transport: http.DefaultTransport}

	ctx := context.Background()
	_, err := svc.Create(ctx, "test", nil)
	assert.Error(t, err)
	assert.False(t, createCalled, "Vault should not be called")
	assert.True(t, gock.IsDone())
}

func TestHarborAddonService_Create_VaultFails(t *testing.T) {
	defer gock.Off()

	// Все шаги до Vault успешны
	gock.New("https://harbor.example.com").
		Post("/api/v2.0/projects").
		Reply(201)

	gock.New("https://harbor.example.com").
		Post("/api/v2.0/robots").
		Reply(201).
		JSON(map[string]interface{}{"id": 1, "name": "robot$svc_test", "secret": "sec"})

	gock.New("https://harbor.example.com").
		Put("/api/v2.0/projects/svc_test").
		Reply(200)

	// Cleanup
	gock.New("https://harbor.example.com").
		Get("/api/v2.0/robots").
		Reply(200).
		JSON([]interface{}{map[string]interface{}{"id": 1}})

	gock.New("https://harbor.example.com").
		Delete("/api/v2.0/robots/1").
		Reply(200)

	gock.New("https://harbor.example.com").
		Delete("/api/v2.0/projects/svc_test").
		Reply(200)

	// Мок Vault с ошибкой
	createCalled := false
	mockVault := &vault.MockVaultClient{
		CreateServiceExtensionFunc: func(ctx context.Context, serviceName, addonName string, data map[string]any) error {
			createCalled = true
			return errors.New("vault error")
		},
	}

	cfg := &connection.AddonConfig{
		Type: connection.HarborAddon,
		Tag:  "test",
		Params: map[string]interface{}{
			"apiUrl":        "https://harbor.example.com",
			"adminUsername": "admin",
			"adminPassword": "pass",
			"baseUrl":       "https://harbor.example.com",
		},
	}
	svc := NewHarborAddonService(cfg, mockVault).(*HarborAddonService)
	svc.HTTPClient = &http.Client{Transport: http.DefaultTransport}

	ctx := context.Background()
	_, err := svc.Create(ctx, "test", nil)
	assert.Error(t, err)
	assert.True(t, createCalled)
	assert.True(t, gock.IsDone())
}

// -----------------------------------------------------------------------------
// Delete tests
// -----------------------------------------------------------------------------

func TestHarborAddonService_Delete_Success(t *testing.T) {
	defer gock.Off()

	projectName := "svc_test"

	// List repositories
	gock.New("https://harbor.example.com").
		Get("/api/v2.0/projects/" + projectName + "/repositories").
		Reply(200).
		JSON([]interface{}{
			map[string]string{"name": projectName + "/repo1"},
			map[string]string{"name": projectName + "/repo2"},
		})

	// List artifacts for repo1
	gock.New("https://harbor.example.com").
		Get("/api/v2.0/projects/" + projectName + "/repositories/repo1/artifacts").
		Reply(200).
		JSON([]interface{}{
			map[string]string{"digest": "sha256:abc"},
			map[string]string{"digest": "sha256:def"},
		})

	// Delete artifacts
	gock.New("https://harbor.example.com").
		Delete("/api/v2.0/projects/" + projectName + "/repositories/repo1/artifacts/sha256:abc").
		Reply(204)
	gock.New("https://harbor.example.com").
		Delete("/api/v2.0/projects/" + projectName + "/repositories/repo1/artifacts/sha256:def").
		Reply(204)

	// List artifacts for repo2 (empty)
	gock.New("https://harbor.example.com").
		Get("/api/v2.0/projects/" + projectName + "/repositories/repo2/artifacts").
		Reply(200).
		JSON([]interface{}{})

	// Delete robot: сначала поиск
	gock.New("https://harbor.example.com").
		Get("/api/v2.0/robots").
		Reply(200).
		JSON([]interface{}{map[string]interface{}{"id": 42}})
	gock.New("https://harbor.example.com").
		Delete("/api/v2.0/robots/42").
		Reply(200)

	// Delete project
	gock.New("https://harbor.example.com").
		Delete("/api/v2.0/projects/" + projectName).
		Reply(200)

	cfg := &connection.AddonConfig{
		Type: connection.HarborAddon,
		Params: map[string]interface{}{
			"apiUrl":        "https://harbor.example.com",
			"adminUsername": "admin",
			"adminPassword": "pass",
		},
	}
	svc := NewHarborAddonService(cfg, nil).(*HarborAddonService)
	svc.HTTPClient = &http.Client{Transport: http.DefaultTransport}

	ctx := context.Background()
	opCfg := &AddonOperationConfig{Name: projectName}
	result, err := svc.Delete(ctx, "target", opCfg)

	require.NoError(t, err)
	assert.Equal(t, projectName, result.Name)
	assert.True(t, gock.IsDone())
}

func TestHarborAddonService_Delete_ProjectNotFound(t *testing.T) {
	defer gock.Off()

	projectName := "svc_test"

	// List repositories (404)
	gock.New("https://harbor.example.com").
		Get("/api/v2.0/projects/" + projectName + "/repositories").
		Reply(404)

	// Delete robot: поиск (не найден)
	gock.New("https://harbor.example.com").
		Get("/api/v2.0/robots").
		Reply(200).
		JSON([]interface{}{})

	// Delete project (404)
	gock.New("https://harbor.example.com").
		Delete("/api/v2.0/projects/" + projectName).
		Reply(404)

	cfg := &connection.AddonConfig{
		Type: connection.HarborAddon,
		Params: map[string]interface{}{
			"apiUrl":        "https://harbor.example.com",
			"adminUsername": "admin",
			"adminPassword": "pass",
		},
	}
	svc := NewHarborAddonService(cfg, nil).(*HarborAddonService)
	svc.HTTPClient = &http.Client{Transport: http.DefaultTransport}

	ctx := context.Background()
	opCfg := &AddonOperationConfig{Name: projectName}
	_, err := svc.Delete(ctx, "target", opCfg)
	assert.NoError(t, err) // ошибки быть не должно, 404 игнорируются
	assert.True(t, gock.IsDone())
}

func TestHarborAddonService_Delete_ProjectDeleteFails(t *testing.T) {
	defer gock.Off()

	projectName := "svc_test"

	// Repositories list empty
	gock.New("https://harbor.example.com").
		Get("/api/v2.0/projects/" + projectName + "/repositories").
		Reply(200).
		JSON([]interface{}{})

	// Delete robot (not found)
	gock.New("https://harbor.example.com").
		Get("/api/v2.0/robots").
		Reply(200).
		JSON([]interface{}{})

	// Delete project fails with 500
	gock.New("https://harbor.example.com").
		Delete("/api/v2.0/projects/" + projectName).
		Reply(500)

	cfg := &connection.AddonConfig{
		Type: connection.HarborAddon,
		Params: map[string]interface{}{
			"apiUrl":        "https://harbor.example.com",
			"adminUsername": "admin",
			"adminPassword": "pass",
		},
	}
	svc := NewHarborAddonService(cfg, nil).(*HarborAddonService)
	svc.HTTPClient = &http.Client{Transport: http.DefaultTransport}

	ctx := context.Background()
	opCfg := &AddonOperationConfig{Name: projectName}
	_, err := svc.Delete(ctx, "target", opCfg)
	assert.Error(t, err)
	assert.True(t, gock.IsDone())
}

// -----------------------------------------------------------------------------
// Reset tests
// -----------------------------------------------------------------------------

func TestHarborAddonService_Reset_Success(t *testing.T) {
	defer gock.Off()

	projectName := "svc_test"

	// Delete old robot: сначала поиск
	gock.New("https://harbor.example.com").
		Get("/api/v2.0/robots").
		Reply(200).
		JSON([]interface{}{map[string]interface{}{"id": 42}})
	gock.New("https://harbor.example.com").
		Delete("/api/v2.0/robots/42").
		Reply(200)

	// Create new robot
	gock.New("https://harbor.example.com").
		Post("/api/v2.0/robots").
		JSON(map[string]interface{}{
			"name":        projectName,
			"level":       "system",
			"description": "Robot account for project svc_test",
			"duration":    -1,
			"permissions": []interface{}{
				map[string]interface{}{
					"kind":      "project",
					"namespace": projectName,
					"access": []interface{}{
						map[string]string{"resource": "tag", "action": "create"},
						map[string]string{"resource": "tag", "action": "delete"},
						map[string]string{"resource": "repository", "action": "list"},
						map[string]string{"resource": "tag", "action": "list"},
						map[string]string{"resource": "repository", "action": "push"},
						map[string]string{"resource": "repository", "action": "read"},
						map[string]string{"resource": "repository", "action": "update"},
						map[string]string{"resource": "repository", "action": "pull"},
						map[string]string{"resource": "repository", "action": "delete"},
					},
				},
			},
		}).
		Reply(201).
		JSON(map[string]interface{}{
			"id":     43,
			"name":   "robot$svc_test",
			"secret": "newsecret",
		})

	createCalled := false
	mockVault := &vault.MockVaultClient{
		CreateServiceExtensionFunc: func(ctx context.Context, serviceName, addonName string, data map[string]any) error {
			createCalled = true
			assert.Equal(t, "target", serviceName)
			assert.Equal(t, "harbor_test", addonName)
			assert.Equal(t, "newsecret", data["HARBOR_REGISTRY_TOKEN"])
			return nil
		},
	}

	cfg := &connection.AddonConfig{
		Type: connection.HarborAddon,
		Tag:  "test",
		Params: map[string]interface{}{
			"apiUrl":        "https://harbor.example.com",
			"adminUsername": "admin",
			"adminPassword": "pass",
			"baseUrl":       "https://harbor.example.com",
		},
	}
	svc := NewHarborAddonService(cfg, mockVault).(*HarborAddonService)
	svc.HTTPClient = &http.Client{Transport: http.DefaultTransport}

	ctx := context.Background()
	opCfg := &AddonOperationConfig{Name: projectName}
	result, err := svc.Reset(ctx, "target", opCfg)

	require.NoError(t, err)
	assert.Equal(t, projectName, result.Name)
	assert.NotNil(t, result.URL)
	assert.True(t, createCalled)
	assert.True(t, gock.IsDone())
}

func TestHarborAddonService_Reset_RobotCreateFails(t *testing.T) {
	defer gock.Off()

	projectName := "svc_test"

	// Delete old robot
	gock.New("https://harbor.example.com").
		Get("/api/v2.0/robots").
		Reply(200).
		JSON([]interface{}{map[string]interface{}{"id": 42}})
	gock.New("https://harbor.example.com").
		Delete("/api/v2.0/robots/42").
		Reply(200)

	// Create new robot fails
	gock.New("https://harbor.example.com").
		Post("/api/v2.0/robots").
		Reply(500)

	createCalled := false
	mockVault := &vault.MockVaultClient{
		CreateServiceExtensionFunc: func(ctx context.Context, serviceName, addonName string, data map[string]any) error {
			createCalled = true
			return nil
		},
	}

	cfg := &connection.AddonConfig{
		Type: connection.HarborAddon,
		Tag:  "test",
		Params: map[string]interface{}{
			"apiUrl":        "https://harbor.example.com",
			"adminUsername": "admin",
			"adminPassword": "pass",
		},
	}
	svc := NewHarborAddonService(cfg, mockVault).(*HarborAddonService)
	svc.HTTPClient = &http.Client{Transport: http.DefaultTransport}

	ctx := context.Background()
	opCfg := &AddonOperationConfig{Name: projectName}
	_, err := svc.Reset(ctx, "target", opCfg)
	assert.Error(t, err)
	assert.False(t, createCalled, "Vault should not be called")
	assert.True(t, gock.IsDone())
}
