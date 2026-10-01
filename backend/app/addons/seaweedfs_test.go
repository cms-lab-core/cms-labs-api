package addons

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"

	"github.com/goccy/go-json"

	"github.com/h2non/gock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cms-lab-core/cms-labs-api/backend/app/addons/vault"
	"github.com/cms-lab-core/cms-labs-api/shared/connection"
)

// -----------------------------------------------------------------------------
// Внутренние функции (не требуют моков)
// -----------------------------------------------------------------------------

func TestGenerateAIMPolicy(t *testing.T) {
	policy, err := generateAIMPolicy("my-bucket", "AK123", "SK456")
	require.NoError(t, err)

	// Проверим структуру через JSON
	raw, _ := json.Marshal(policy)
	assert.Contains(t, string(raw), "my-bucket")
	assert.Contains(t, string(raw), "AK123")
	assert.Contains(t, string(raw), "SK456")
	assert.Contains(t, string(raw), "Read:my-bucket")
	assert.Contains(t, string(raw), "Write:my-bucket")
}

func TestGenerateBucketName(t *testing.T) {
	svc := &SeaweedFSAddonService{}
	tests := []struct {
		in   string
		want string
	}{
		{"MyProject", "svc-myproject"},
		{"Project 123", "svc-project-123"},
		{"a#b$c", "svc-a-b-c"},
		{strings.Repeat("x", 100), "svc-" + strings.Repeat("x", 59)}, // 63 total
		{"leading-trailing-", "svc-leading-trailing"},
	}
	for _, tt := range tests {
		got := svc.generateBucketName(tt.in)
		assert.Equal(t, tt.want, got)
	}
}

func TestGenerateAccessKeyFormat(t *testing.T) {
	svc := &SeaweedFSAddonService{}
	key := svc.generateAccessKey()
	// Должен начинаться с "SW"
	assert.True(t, strings.HasPrefix(key, "SW"))
	// Длина не больше 20
	assert.LessOrEqual(t, len(key), 20)
	// Содержит только допустимые символы
	assert.Regexp(t, "^[A-Z0-9]+$", key)
}

func TestGenerateSecretKeyFormat(t *testing.T) {
	svc := &SeaweedFSAddonService{}
	key := svc.generateSecretKey()
	assert.Len(t, key, 40)
	// Допустимые символы: A-Z a-z 0-9 + /
	assert.Regexp(t, "^[A-Za-z0-9+/]+$", key)
}

// -----------------------------------------------------------------------------
// Вспомогательная функция для настройки моков
// -----------------------------------------------------------------------------

// setupMockSeaweedFS создаёт сервис с мокнутым Vault и HTTP-клиентом для gock.
func setupMockSeaweedFS(t *testing.T, mockVault vault.ClientInterface) *SeaweedFSAddonService {
	cfg := &connection.AddonConfig{
		Type: connection.S3Addon,
		Tag:  "test",
		Params: map[string]interface{}{
			"endpoint":         "https://seaweed.example.com",
			"service_in_vault": "seaweed-service",
			"git_url":          "https://gitlab.example.com",
			"git_repo":         "myrepo",
			"git_token":        "glpat-xxx",
			"git_branch":       "main",
		},
	}
	svc := NewSeaweedFSAddonService(cfg, mockVault).(*SeaweedFSAddonService)
	svc.HTTPClient = &http.Client{Transport: gock.DefaultTransport}
	t.Cleanup(func() {
		gock.Off()
	})
	return svc
}

// -----------------------------------------------------------------------------
// Create
// -----------------------------------------------------------------------------

func TestSeaweedFSAddonService_Create_Success(t *testing.T) {
	defer gock.Off()

	// Мок Vault
	var (
		createPolicyCalled bool
		storeCredsCalled   bool
	)
	mockVault := &vault.MockVaultClient{
		CreateServiceExtensionFunc: func(ctx context.Context, serviceName, addonName string, data map[string]any) error {
			if serviceName == "seaweed-service" && addonName == "svc-test-target" {
				// Создание политики
				createPolicyCalled = true
				// Проверим, что политика содержит правильные поля
				policy, ok := data["identities"].([]interface{})
				if !ok {
					t.Error("policy missing identities")
				}
				assert.Len(t, policy, 1)
				return nil
			}
			if serviceName == "test-target" && addonName == "s3_test" {
				// Сохранение учётных данных
				storeCredsCalled = true
				assert.Equal(t, "https://seaweed.example.com", data["SEAWEEDFS_ENDPOINT"])
				assert.Equal(t, "us-east-1", data["SEAWEEDFS_REGION"])
				assert.Equal(t, "svc-test-target", data["SEAWEEDFS_BUCKET"])
				assert.Contains(t, data, "AWS_ACCESS_KEY_ID")
				assert.Contains(t, data, "AWS_SECRET_ACCESS_KEY")
				return nil
			}
			t.Errorf("unexpected call to CreateServiceExtension(%q, %q)", serviceName, addonName)
			return nil
		},
		// Другие методы не используются
	}

	// Мок GitLab pipeline
	gock.New("https://gitlab.example.com").
		Post("/api/v4/projects/myrepo/trigger/pipeline").
		MatchParam("token", "glpat-xxx").
		MatchParam("ref", "main").
		MatchParam("variables[SEAWEEDFS_UPDATE]", "true").
		Reply(201)

	svc := setupMockSeaweedFS(t, mockVault)
	ctx := context.Background()
	result, err := svc.Create(ctx, "test-target", nil)

	require.NoError(t, err)
	assert.Equal(t, "svc-test-target", result.Name)
	assert.Equal(t, 0, *result.CurrentSizeMb)
	assert.Equal(t, 1024, *result.MaxSizeMb)
	assert.True(t, createPolicyCalled)
	assert.True(t, storeCredsCalled)
}

func TestSeaweedFSAddonService_Create_PolicyFails(t *testing.T) {
	mockVault := &vault.MockVaultClient{
		CreateServiceExtensionFunc: func(ctx context.Context, serviceName, addonName string, data map[string]any) error {
			if serviceName == "seaweed-service" {
				return errors.New("vault policy error")
			}
			return nil
		},
	}
	svc := setupMockSeaweedFS(t, mockVault)
	_, err := svc.Create(context.Background(), "test-target", nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "vault policy error")
}

func TestSeaweedFSAddonService_Create_StoreCredsFails(t *testing.T) {
	called := false
	mockVault := &vault.MockVaultClient{
		CreateServiceExtensionFunc: func(ctx context.Context, serviceName, addonName string, data map[string]any) error {
			if serviceName == "seaweed-service" {
				return nil // policy ok
			}
			if serviceName == "test-target" {
				called = true
				return errors.New("vault store error")
			}
			return nil
		},
	}
	svc := setupMockSeaweedFS(t, mockVault)
	_, err := svc.Create(context.Background(), "test-target", nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "vault store error")
	assert.True(t, called)
	// GitLab не должен вызываться, так как ошибка раньше
}

func TestSeaweedFSAddonService_Create_UpdateFails(t *testing.T) {
	defer gock.Off()

	mockVault := &vault.MockVaultClient{
		CreateServiceExtensionFunc: func(ctx context.Context, serviceName, addonName string, data map[string]any) error {
			return nil // всегда успех
		},
	}

	// GitLab pipeline вернёт ошибку
	gock.New("https://gitlab.example.com").
		Post("/api/v4/projects/myrepo/trigger/pipeline").
		Reply(500)

	svc := setupMockSeaweedFS(t, mockVault)
	ctx := context.Background()
	result, err := svc.Create(ctx, "test-target", nil)

	// Ошибка update не должна прерывать создание, только логироваться
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, gock.IsDone())
}

// -----------------------------------------------------------------------------
// Delete
// -----------------------------------------------------------------------------

func TestSeaweedFSAddonService_Delete_Success(t *testing.T) {
	defer gock.Off()

	var (
		deletePolicyCalled bool
		deleteCredsCalled  bool
	)
	mockVault := &vault.MockVaultClient{
		CreateServiceExtensionFunc: func(ctx context.Context, serviceName, addonName string, data map[string]any) error {
			if serviceName == "seaweed-service" && addonName == "svc-test-target" {
				deletePolicyCalled = true
				assert.Empty(t, data) // пустой объект
				return nil
			}
			if serviceName == "svc-test-target" && addonName == "s3_test" {
				deleteCredsCalled = true
				assert.Empty(t, data)
				return nil
			}
			t.Errorf("unexpected call: %s/%s", serviceName, addonName)
			return nil
		},
	}

	gock.New("https://gitlab.example.com").
		Post("/api/v4/projects/myrepo/trigger/pipeline").
		MatchParam("variables[SEAWEEDFS_UPDATE]", "true").
		Reply(201)

	svc := setupMockSeaweedFS(t, mockVault)
	cfg := &AddonOperationConfig{Name: "svc-test-target"}
	result, err := svc.Delete(context.Background(), "ignored-target", cfg)

	require.NoError(t, err)
	assert.Equal(t, "svc-test-target", result.Name)
	assert.True(t, deletePolicyCalled)
	assert.True(t, deleteCredsCalled)
}

func TestSeaweedFSAddonService_Delete_NoBucketName(t *testing.T) {
	svc := setupMockSeaweedFS(t, nil)
	_, err := svc.Delete(context.Background(), "target", nil)
	assert.Error(t, err)

	_, err = svc.Delete(context.Background(), "target", &AddonOperationConfig{Name: ""})
	assert.Error(t, err)
}

func TestSeaweedFSAddonService_Delete_PolicyFails(t *testing.T) {
	mockVault := &vault.MockVaultClient{
		CreateServiceExtensionFunc: func(ctx context.Context, serviceName, addonName string, data map[string]any) error {
			if serviceName == "seaweed-service" {
				return errors.New("delete policy error")
			}
			return nil
		},
	}
	svc := setupMockSeaweedFS(t, mockVault)
	_, err := svc.Delete(context.Background(), "target", &AddonOperationConfig{Name: "bucket"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "delete policy error")
}

// -----------------------------------------------------------------------------
// Reset
// -----------------------------------------------------------------------------

func TestSeaweedFSAddonService_Reset(t *testing.T) {
	// Reset должен просто вызвать Create. Проверим это.
	mockVault := &vault.MockVaultClient{
		CreateServiceExtensionFunc: func(ctx context.Context, serviceName, addonName string, data map[string]any) error {
			return nil // любое создание успешно
		},
	}
	defer gock.Off()
	gock.New("https://gitlab.example.com").
		Post("/api/v4/projects/myrepo/trigger/pipeline").
		Reply(201)

	svc := setupMockSeaweedFS(t, mockVault)
	ctx := context.Background()

	// Reset с nil-конфигом
	result, err := svc.Reset(ctx, "test-target", nil)
	require.NoError(t, err)
	assert.Equal(t, "svc-test-target", result.Name)

	// Reset с конфигом (игнорируется) – результат должен быть таким же
	result2, err := svc.Reset(ctx, "test-target", &AddonOperationConfig{Name: "old-bucket"})
	require.NoError(t, err)
	assert.Equal(t, "svc-test-target", result2.Name) // имя генерируется заново
}

// -----------------------------------------------------------------------------
// GetType
// -----------------------------------------------------------------------------

func TestSeaweedFSAddonService_GetType(t *testing.T) {
	cfg := &connection.AddonConfig{Type: connection.S3Addon}
	svc := NewSeaweedFSAddonService(cfg, nil).(*SeaweedFSAddonService)
	assert.Equal(t, connection.S3Addon, svc.GetType())
}
