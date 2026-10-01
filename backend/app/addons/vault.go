package addons

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/cms-lab-core/cms-labs-api/backend/app/addons/vault"
	"github.com/cms-lab-core/cms-labs-api/shared/connection"
)

// VaultAddonService implements AddonService for issuing Vault tokens.
// It creates, rotates, and revokes service tokens for a target service.
type VaultAddonService struct {
	Config      *connection.AddonConfig
	VaultClient vault.ClientInterface
	tokenTTL    string // e.g. "87600h" (10 years)
}

// NewVaultAddonService creates a new VaultAddonService.
func NewVaultAddonService(cfg *connection.AddonConfig, vaultClient vault.ClientInterface) AddonService {
	return &VaultAddonService{
		Config:      cfg,
		VaultClient: vaultClient,
		tokenTTL:    "87600h", // 10 years
	}
}

// TokenInfo represents the response from issuing/rotating a service token.
// This structure must match what the Vault client returns.
type TokenInfo struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expiresAt"`
}

// Create issues a new service token and stores it in Vault under the target's secret.
// The optional name parameter is ignored – the token is associated with the targetName.
func (s *VaultAddonService) Create(ctx context.Context, targetName string, _ *string) (*AddonOperationConfig, error) {
	// Issue a new service token
	renewable := false
	tokenInfo, err := s.VaultClient.IssueServiceToken(ctx, targetName, vault.ServiceTokenOptions{
		TTL:        s.tokenTTL,
		Renewable:  &renewable,
		Permission: "read",
	})
	if err != nil {
		return nil, fmt.Errorf("issue service token: %w", err)
	}

	// Store the token and expiry in the target's own Vault extension
	extName := strings.ToLower(string(s.Config.Type))
	err = s.VaultClient.CreateServiceExtension(
		ctx,
		targetName,
		extName,
		map[string]interface{}{
			"VAULT_API_TOKEN":      tokenInfo.Token,
			"VAULT_API_EXPIRED_AT": tokenInfo.ExpiresAt,
		},
	)
	if err != nil {
		// Attempt to clean up the issued token if storing fails
		_ = s.VaultClient.RevokeServiceToken(ctx, targetName)
		return nil, fmt.Errorf("store token in Vault: %w", err)
	}

	url := s.getVaultURL(targetName)
	return &AddonOperationConfig{
		Name: targetName,
		URL:  &url,
	}, nil
}

// Delete revokes the service token for the given configuration.
// cfg.Name must contain the target name that was used during creation.
func (s *VaultAddonService) Delete(ctx context.Context, targetName string, cfg *AddonOperationConfig) (*AddonOperationConfig, error) {
	err := s.VaultClient.RevokeServiceToken(ctx, targetName)
	if err != nil {
		return nil, fmt.Errorf("revoke service token: %w", err)
	}

	return &AddonOperationConfig{Name: targetName}, nil
}

// Reset rotates the service token for the target and updates the stored secret.
// The provided cfg is ignored; rotation always applies to the given targetName.
func (s *VaultAddonService) Reset(ctx context.Context, targetName string, _ *AddonOperationConfig) (*AddonOperationConfig, error) {
	// Rotate the existing service token
	renewable := false
	tokenInfo, err := s.VaultClient.RotateServiceToken(ctx, targetName, vault.ServiceTokenOptions{
		TTL:        s.tokenTTL,
		Renewable:  &renewable,
		Permission: "read",
	})
	if err != nil {
		return nil, fmt.Errorf("rotate service token: %w", err)
	}

	// Update the stored secret with the new token
	extName := strings.ToLower(string(s.Config.Type))
	err = s.VaultClient.CreateServiceExtension(
		ctx,
		targetName,
		extName,
		map[string]interface{}{
			"VAULT_API_TOKEN":      tokenInfo.Token,
			"VAULT_API_EXPIRED_AT": tokenInfo.ExpiresAt,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("update token in Vault: %w", err)
	}

	url := s.getVaultURL(targetName)
	return &AddonOperationConfig{
		Name: targetName,
		URL:  &url,
	}, nil
}

// getVaultURL constructs a deep link to the Vault UI secret location for the target.
// It reads the VAULT_ADDR environment variable; defaults to http://localhost:8200.
func (s *VaultAddonService) getVaultURL(targetName string) string {
	return fmt.Sprintf(
		"%s/ui/vault/secrets/secret/kv/list/services/%s",
		s.VaultClient.GetVaultAddr(),
		targetName,
	)
}

func (s *VaultAddonService) GetType() connection.AddonType {
	return s.Config.Type
}
