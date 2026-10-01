package addons

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cms-lab-core/cms-labs-api/backend/app/addons/vault"
	"github.com/cms-lab-core/cms-labs-api/shared/connection"
)

func TestVaultAddonService_GetType(t *testing.T) {
	cfg := &connection.AddonConfig{Type: connection.VaultAddon}
	svc := NewVaultAddonService(cfg, nil).(*VaultAddonService)
	assert.Equal(t, connection.VaultAddon, svc.GetType())
}

func TestVaultAddonService_getVaultURL(t *testing.T) {
	mockVault := &vault.MockVaultClient{
		GetVaultAddrFunc: func() string {
			return "https://vault.example.com:8200"
		},
	}
	svc := &VaultAddonService{
		VaultClient: mockVault,
	}
	url := svc.getVaultURL("test-service")
	assert.Equal(t, "https://vault.example.com:8200/ui/vault/secrets/secret/kv/list/services/test-service", url)
}

func TestVaultAddonService_Create_Success(t *testing.T) {
	ctx := context.Background()
	targetName := "test-service"
	fixedTime := time.Now().UTC()

	issueCalled := false
	createExtCalled := false

	mockVault := &vault.MockVaultClient{
		IssueServiceTokenFunc: func(ctx context.Context, tn string, opts vault.ServiceTokenOptions) (*vault.ServiceTokenInfo, error) {
			issueCalled = true
			assert.Equal(t, targetName, tn)
			assert.Equal(t, "87600h", opts.TTL)
			assert.False(t, *opts.Renewable)
			assert.Equal(t, "read", opts.Permission)
			return &vault.ServiceTokenInfo{
				Token:       "s.abc123",
				Accessor:    "accessor123",
				ServiceName: "svc-" + tn,
				Policies:    []string{"policy"},
				TTL:         3600,
				Renewable:   false,
				CreatedAt:   fixedTime.Format(time.RFC3339),
				ExpiresAt:   fixedTime.Add(3600 * time.Second).Format(time.RFC3339),
			}, nil
		},
		CreateServiceExtensionFunc: func(ctx context.Context, serviceName, addonName string, data map[string]any) error {
			createExtCalled = true
			assert.Equal(t, targetName, serviceName)
			assert.Equal(t, "vault", addonName)
			assert.Equal(t, "s.abc123", data["VAULT_API_TOKEN"])
			assert.NotEmpty(t, data["VAULT_API_EXPIRED_AT"])
			return nil
		},
		GetVaultAddrFunc: func() string {
			return "http://vault:8200"
		},
	}

	cfg := &connection.AddonConfig{Type: connection.VaultAddon}
	svc := NewVaultAddonService(cfg, mockVault).(*VaultAddonService)

	result, err := svc.Create(ctx, targetName, nil)

	require.NoError(t, err)
	assert.Equal(t, targetName, result.Name)
	assert.NotNil(t, result.URL)
	assert.Equal(t, "http://vault:8200/ui/vault/secrets/secret/kv/list/services/"+targetName, *result.URL)
	assert.True(t, issueCalled)
	assert.True(t, createExtCalled)
}

func TestVaultAddonService_Create_IssueTokenFails(t *testing.T) {
	ctx := context.Background()
	targetName := "test-service"

	mockVault := &vault.MockVaultClient{
		IssueServiceTokenFunc: func(ctx context.Context, tn string, opts vault.ServiceTokenOptions) (*vault.ServiceTokenInfo, error) {
			return nil, errors.New("vault token issue failed")
		},
		// CreateServiceExtension не должен вызываться
		CreateServiceExtensionFunc: func(ctx context.Context, serviceName, addonName string, data map[string]any) error {
			t.Fatal("should not be called")
			return nil
		},
	}

	cfg := &connection.AddonConfig{Type: connection.VaultAddon}
	svc := NewVaultAddonService(cfg, mockVault).(*VaultAddonService)

	_, err := svc.Create(ctx, targetName, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "issue service token")
}

func TestVaultAddonService_Create_StoreTokenFails_Revokes(t *testing.T) {
	ctx := context.Background()
	targetName := "test-service"
	fixedTime := time.Now().UTC()

	revokeCalled := false
	mockVault := &vault.MockVaultClient{
		IssueServiceTokenFunc: func(ctx context.Context, tn string, opts vault.ServiceTokenOptions) (*vault.ServiceTokenInfo, error) {
			return &vault.ServiceTokenInfo{
				Token:     "s.abc123",
				Accessor:  "accessor123",
				ExpiresAt: fixedTime.Add(3600 * time.Second).Format(time.RFC3339),
			}, nil
		},
		CreateServiceExtensionFunc: func(ctx context.Context, serviceName, addonName string, data map[string]any) error {
			return errors.New("vault storage error")
		},
		RevokeServiceTokenFunc: func(ctx context.Context, tn string) error {
			revokeCalled = true
			assert.Equal(t, targetName, tn)
			return nil
		},
	}

	cfg := &connection.AddonConfig{Type: connection.VaultAddon}
	svc := NewVaultAddonService(cfg, mockVault).(*VaultAddonService)

	_, err := svc.Create(ctx, targetName, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "store token in Vault")
	assert.True(t, revokeCalled, "token should be revoked after storage failure")
}

func TestVaultAddonService_Delete_Success(t *testing.T) {
	ctx := context.Background()
	targetName := "test-service"

	revokeCalled := false
	mockVault := &vault.MockVaultClient{
		RevokeServiceTokenFunc: func(ctx context.Context, tn string) error {
			revokeCalled = true
			assert.Equal(t, targetName, tn)
			return nil
		},
	}

	cfg := &connection.AddonConfig{Type: connection.VaultAddon}
	svc := NewVaultAddonService(cfg, mockVault).(*VaultAddonService)

	result, err := svc.Delete(ctx, targetName, &AddonOperationConfig{Name: targetName})

	require.NoError(t, err)
	assert.Equal(t, targetName, result.Name)
	assert.True(t, revokeCalled)
}

func TestVaultAddonService_Delete_RevokeFails(t *testing.T) {
	ctx := context.Background()
	targetName := "test-service"

	mockVault := &vault.MockVaultClient{
		RevokeServiceTokenFunc: func(ctx context.Context, tn string) error {
			return errors.New("revoke failed")
		},
	}

	cfg := &connection.AddonConfig{Type: connection.VaultAddon}
	svc := NewVaultAddonService(cfg, mockVault).(*VaultAddonService)

	_, err := svc.Delete(ctx, targetName, &AddonOperationConfig{Name: targetName})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "revoke service token")
}

func TestVaultAddonService_Reset_Success(t *testing.T) {
	ctx := context.Background()
	targetName := "test-service"
	fixedTime := time.Now().UTC()

	rotateCalled := false
	createExtCalled := false

	mockVault := &vault.MockVaultClient{
		RotateServiceTokenFunc: func(ctx context.Context, tn string, opts vault.ServiceTokenOptions) (*vault.ServiceTokenInfo, error) {
			rotateCalled = true
			assert.Equal(t, targetName, tn)
			assert.Equal(t, "87600h", opts.TTL)
			assert.False(t, *opts.Renewable)
			assert.Equal(t, "read", opts.Permission)
			return &vault.ServiceTokenInfo{
				Token:       "s.newtoken",
				Accessor:    "newaccessor",
				ServiceName: "svc-" + tn,
				Policies:    []string{"policy"},
				TTL:         3600,
				Renewable:   false,
				CreatedAt:   fixedTime.Format(time.RFC3339),
				ExpiresAt:   fixedTime.Add(3600 * time.Second).Format(time.RFC3339),
			}, nil
		},
		CreateServiceExtensionFunc: func(ctx context.Context, serviceName, addonName string, data map[string]any) error {
			createExtCalled = true
			assert.Equal(t, targetName, serviceName)
			assert.Equal(t, "vault", addonName)
			assert.Equal(t, "s.newtoken", data["VAULT_API_TOKEN"])
			assert.NotEmpty(t, data["VAULT_API_EXPIRED_AT"])
			return nil
		},
		GetVaultAddrFunc: func() string {
			return "http://vault:8200"
		},
	}

	cfg := &connection.AddonConfig{Type: connection.VaultAddon}
	svc := NewVaultAddonService(cfg, mockVault).(*VaultAddonService)

	result, err := svc.Reset(ctx, targetName, nil)

	require.NoError(t, err)
	assert.Equal(t, targetName, result.Name)
	assert.NotNil(t, result.URL)
	assert.True(t, rotateCalled)
	assert.True(t, createExtCalled)
}

func TestVaultAddonService_Reset_RotateFails(t *testing.T) {
	ctx := context.Background()
	targetName := "test-service"

	mockVault := &vault.MockVaultClient{
		RotateServiceTokenFunc: func(ctx context.Context, tn string, opts vault.ServiceTokenOptions) (*vault.ServiceTokenInfo, error) {
			return nil, errors.New("rotate failed")
		},
		// CreateServiceExtension не должен вызываться
		CreateServiceExtensionFunc: func(ctx context.Context, serviceName, addonName string, data map[string]any) error {
			t.Fatal("should not be called")
			return nil
		},
	}

	cfg := &connection.AddonConfig{Type: connection.VaultAddon}
	svc := NewVaultAddonService(cfg, mockVault).(*VaultAddonService)

	_, err := svc.Reset(ctx, targetName, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "rotate service token")
}

func TestVaultAddonService_Reset_StoreFails(t *testing.T) {
	ctx := context.Background()
	targetName := "test-service"
	fixedTime := time.Now().UTC()

	mockVault := &vault.MockVaultClient{
		RotateServiceTokenFunc: func(ctx context.Context, tn string, opts vault.ServiceTokenOptions) (*vault.ServiceTokenInfo, error) {
			return &vault.ServiceTokenInfo{
				Token:     "s.newtoken",
				Accessor:  "newaccessor",
				ExpiresAt: fixedTime.Add(3600 * time.Second).Format(time.RFC3339),
			}, nil
		},
		CreateServiceExtensionFunc: func(ctx context.Context, serviceName, addonName string, data map[string]any) error {
			return errors.New("storage error")
		},
		// Примечание: при ошибке сохранения токен НЕ отзывается, так как это reset
	}

	cfg := &connection.AddonConfig{Type: connection.VaultAddon}
	svc := NewVaultAddonService(cfg, mockVault).(*VaultAddonService)

	_, err := svc.Reset(ctx, targetName, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "update token in Vault")
}
