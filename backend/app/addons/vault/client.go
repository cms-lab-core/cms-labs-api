package vault

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/cms-lab-core/cms-labs-api/shared/connection"
	"github.com/hashicorp/vault/api"
	"github.com/rs/zerolog"
)

func NewClient(cfg *connection.Vault, logger *zerolog.Logger) *Client {
	apiConfig := api.DefaultConfig()
	apiConfig.Address = cfg.VaultAddr
	c, _ := api.NewClient(apiConfig)
	c.SetToken(cfg.VaultToken)
	client := &Client{
		Client:    c,
		VaultAddr: cfg.VaultAddr,
		Logger:    logger,
	}
	client.Logger.Info().Msg(fmt.Sprintf("NewClient: Vault client created with address %s", cfg.VaultAddr))
	return client
}

type Client struct {
	Client    *api.Client
	VaultAddr string
	Logger    *zerolog.Logger
}

const (
	Read  = "read"
	Write = "write"
	None  = "none"
)

type ServiceTokenOptions struct {
	TTL        string
	MaxTTL     string
	Renewable  *bool
	Policies   []string
	Metadata   map[string]string
	NumUses    int
	Permission string
}

type ServiceTokenInfo struct {
	Token       string
	Accessor    string
	ServiceName string
	Policies    []string
	TTL         int
	Renewable   bool
	CreatedAt   string
	ExpiresAt   string
}

type UsersAndServices struct {
	UserEmail string
	Services  []ServicesBindUser
}

type ServicesBindUser struct {
	ServiceName string
	Policies    []string
}

func (v *Client) CreateServiceExtension(ctx context.Context, serviceName, addonName string, data map[string]any) error {
	v.Logger.Info().Msg(fmt.Sprintf("CreateServiceExtension: creating extension for service %s, addon %s", serviceName, addonName))
	p := fmt.Sprintf("secret/data/services/%s/.infra/%s", sanitize(serviceName), sanitize(addonName))

	payload := map[string]any{
		"data": map[string]any{
			"_metadata": map[string]string{
				"created": time.Now().Format(time.RFC3339),
				"updated": time.Now().Format(time.RFC3339),
				"service": serviceName,
				"addon":   addonName,
			},
		},
	}
	for k, val := range data {
		payload["data"].(map[string]any)[k] = val
	}

	_, err := v.Client.Logical().WriteWithContext(ctx, p, payload)
	if err == nil {
		v.Logger.Info().Msg(fmt.Sprintf("CreateServiceExtension: extension created for service %s, addon %s", serviceName, addonName))
	}
	return err
}

func (v *Client) CreateKubernetesRole(ctx context.Context, targetName string) error {
	v.Logger.Info().Msg(fmt.Sprintf("CreateKubernetesRole: creating role for target %s", targetName))
	p := fmt.Sprintf("auth/kubernetes/role/%s", targetName)

	payload := map[string]any{
		"bound_service_account_names":      "*",
		"bound_service_account_namespaces": sanitize(targetName),
		"policies":                         fmt.Sprintf("services-%s-read", sanitize(targetName)),
		"ttl":                              "1h",
	}
	_, err := v.Client.Logical().WriteWithContext(ctx, p, payload)
	if err == nil {
		v.Logger.Info().Msg(fmt.Sprintf("CreateKubernetesRole: role created for target %s", targetName))
	}
	return err
}

func (v *Client) RevokeKubernetesRole(ctx context.Context, targetName string) error {
	v.Logger.Info().Msg(fmt.Sprintf("RevokeKubernetesRole: revoking role for target %s", targetName))
	p := fmt.Sprintf("auth/kubernetes/role/%s", sanitize(targetName))
	_, err := v.Client.Logical().DeleteWithContext(ctx, p)
	if err == nil {
		v.Logger.Info().Msg(fmt.Sprintf("RevokeKubernetesRole: role revoked for target %s", targetName))
	}
	return err
}

func (v *Client) CreateUsernameExtension(ctx context.Context, username, addonName string, data map[string]any) error {
	v.Logger.Info().Msg(fmt.Sprintf("CreateUsernameExtension: creating extension for user %s, addon %s", username, addonName))
	p := fmt.Sprintf("secret/data/users/%s/.infra/%s", sanitize(username), sanitize(addonName))

	payload := map[string]any{
		"data": map[string]any{
			"_metadata": map[string]string{
				"created":  time.Now().Format(time.RFC3339),
				"updated":  time.Now().Format(time.RFC3339),
				"username": username,
				"addon":    addonName,
			},
		},
	}
	for k, val := range data {
		payload["data"].(map[string]any)[k] = val
	}

	_, err := v.Client.Logical().WriteWithContext(ctx, p, payload)
	if err == nil {
		v.Logger.Info().Msg(fmt.Sprintf("CreateUsernameExtension: extension created for user %s, addon %s", username, addonName))
	}
	return err
}

// IssueServiceToken / RotateServiceToken / RevokeServiceToken — полностью аналог TS-версии
func (v *Client) IssueServiceToken(ctx context.Context, targetName string, opts ServiceTokenOptions) (*ServiceTokenInfo, error) {
	v.Logger.Info().Msg(fmt.Sprintf("IssueServiceToken: issuing token for target %s", targetName))
	serviceName := "svc-" + sanitize(targetName)
	servicePath := "services/" + sanitize(targetName)

	perm := opts.Permission
	if perm == "" {
		perm = Write
	}
	policyName, err := v.createOrUpdatePolicy(ctx, servicePath, perm)
	if err != nil {
		return nil, err
	}
	v.Logger.Info().Msg(fmt.Sprintf("IssueServiceToken: policy %s created/updated for %s", policyName, targetName))

	policies := append(opts.Policies, policyName)

	ttl := opts.TTL
	if ttl == "" {
		ttl = "720h" // 30 дней по умолчанию
	}

	req := &api.TokenCreateRequest{
		DisplayName:    serviceName,
		Policies:       policies,
		TTL:            ttl,
		ExplicitMaxTTL: opts.MaxTTL,
		Renewable:      opts.Renewable,
		NumUses:        opts.NumUses,
		Metadata:       opts.Metadata,
		NoParent:       true,
	}

	secret, err := v.Client.Auth().Token().CreateWithContext(ctx, req)
	if err != nil {
		return nil, err
	}

	auth := secret.Auth

	info := &ServiceTokenInfo{
		Token:       auth.ClientToken,
		Accessor:    auth.Accessor,
		ServiceName: serviceName,
		Policies:    auth.Policies,
		TTL:         auth.LeaseDuration,
		Renewable:   auth.Renewable,
		CreatedAt:   time.Now().Format(time.RFC3339),
		ExpiresAt:   time.Now().Add(time.Duration(auth.LeaseDuration) * time.Second).Format(time.RFC3339),
	}

	// сохраняем accessor для revoke
	_ = v.CreateServiceExtension(ctx, ".service-tokens", serviceName, map[string]any{
		"accessor":    auth.Accessor,
		"serviceName": serviceName,
		"targetName":  targetName,
	})
	v.Logger.Info().Msg(fmt.Sprintf("IssueServiceToken: token issued for %s, accessor: %s", targetName, auth.Accessor))

	return info, nil
}

func (v *Client) RevokeServiceToken(ctx context.Context, targetName string) error {
	v.Logger.Info().Msg(fmt.Sprintf("RevokeServiceToken: revoking token for target %s", targetName))
	serviceName := "svc-" + sanitize(targetName)

	// читаем accessor
	secret, err := v.Client.Logical().ReadWithContext(ctx, "secret/data/.service-tokens/"+serviceName)
	if err != nil || secret == nil {
		return err
	}
	accessor, _ := secret.Data["data"].(map[string]any)["accessor"].(string)
	if accessor != "" {
		_ = v.Client.Sys().RevokeWithContext(ctx, accessor)
		v.Logger.Info().Msg(fmt.Sprintf("RevokeServiceToken: revoked token with accessor %s", accessor))
	}

	// удаляем политику и запись
	_ = v.Client.Sys().DeletePolicyWithContext(ctx, strings.ReplaceAll(sanitize("services/"+targetName), "/", "-")+"-write")
	_, _ = v.Client.Logical().DeleteWithContext(ctx, "secret/metadata/.service-tokens/"+serviceName)
	v.Logger.Info().Msg(fmt.Sprintf("RevokeServiceToken: token data and policy cleaned up for %s", targetName))
	return nil
}

func (v *Client) RotateServiceToken(ctx context.Context, targetName string, opts ServiceTokenOptions) (*ServiceTokenInfo, error) {
	v.Logger.Info().Msg(fmt.Sprintf("RotateServiceToken: rotating token for target %s", targetName))
	_ = v.RevokeServiceToken(ctx, targetName) // best-effort
	tokenInfo, err := v.IssueServiceToken(ctx, targetName, opts)
	if err == nil {
		v.Logger.Info().Msg(fmt.Sprintf("RotateServiceToken: token rotated for %s", targetName))
	}
	return tokenInfo, err
}

func (v *Client) UserBindAccessServices(ctx context.Context, binds []UsersAndServices) error {
	v.Logger.Info().Msg(fmt.Sprintf("UserBindAccessServices: binding access for %d users", len(binds)))

	// 1. Get the OIDC mount accessor (cache this in Client struct if possible)
	authMethods, err := v.Client.Sys().ListAuthWithContext(ctx)
	if err != nil {
		return fmt.Errorf("failed to list auth methods: %w", err)
	}
	oidcMethod, ok := authMethods["oidc/"]
	if !ok {
		return fmt.Errorf("oidc auth method not found")
	}
	accessor := oidcMethod.Accessor

	for _, bind := range binds {
		var policies []string
		for _, opts := range bind.Services {
			servicePath := "services/" + sanitize(opts.ServiceName)
			for _, perm := range opts.Policies {
				policyName, err := v.createOrUpdatePolicy(ctx, servicePath, perm)
				if err != nil {
					return err
				}
				policies = append(policies, policyName)
			}
			policyName, err := v.createOrUpdateDefaultPolicy(ctx)
			if err != nil {
				return err
			}
			policies = append(policies, policyName)
		}

		// 2. Lookup existing alias by name and mount accessor
		v.Logger.Info().Msg(fmt.Sprintf(
			"UserBindAccessServices: looking up alias for user %s and accessor %s",
			bind.UserEmail,
			accessor,
		))
		lookupData := map[string]interface{}{
			"alias_name":           bind.UserEmail,
			"alias_mount_accessor": accessor,
		}
		entityData, err := v.Client.Logical().WriteWithContext(ctx, "identity/lookup/entity", lookupData)
		if err != nil {
			// If not found (404), we'll create new entity and alias
			v.Logger.Warn().Msgf("alias not found for %s, will create new", bind.UserEmail)
		}
		var entityID string
		if entityData != nil && entityData.Data != nil {
			v.Logger.Debug().Msg(fmt.Sprintf(
				"UserBindAccessServices: alias %s with oidc/ %+v",
				bind.UserEmail,
				entityData.Data,
			))
			entityID, _ = entityData.Data["id"].(string)
		}
		v.Logger.Info().Msg(fmt.Sprintf(
			"UserBindAccessServices: alias %s and entityID %s",
			bind.UserEmail,
			entityID,
		))
		if entityID == "" {
			// 3a. Create new entity
			v.Logger.Info().Msg(fmt.Sprintf("UserBindAccessServices: creating new entity for user %s", bind.UserEmail))
			entityReq := map[string]interface{}{
				"name":     bind.UserEmail,
				"policies": policies,
			}
			entitySecret, err := v.Client.Logical().WriteWithContext(ctx, "identity/entity", entityReq)
			if err != nil {
				v.Logger.Error().Msg(fmt.Sprintf("failed to write entity for %s: %+v", bind.UserEmail, err))
				continue
			}
			entityID = entitySecret.Data["id"].(string)
			v.Logger.Info().Msg(fmt.Sprintf("UserBindAccessServices: entity created for %s with ID %s", bind.UserEmail, entityID))

			// 3b. Create alias for the new entity
			aliasReq := map[string]interface{}{
				"name":           bind.UserEmail,
				"canonical_id":   entityID,
				"mount_accessor": accessor,
			}
			_, err = v.Client.Logical().WriteWithContext(ctx, "identity/entity-alias", aliasReq)
			if err != nil {
				v.Logger.Error().Msg(fmt.Sprintf("failed to create entity-alias for %s: %+v", bind.UserEmail, err))
				continue
			}
			v.Logger.Info().Msg(fmt.Sprintf("UserBindAccessServices: alias created for user %s", bind.UserEmail))
		} else {
			// 4. Entity exists – update its policies
			v.Logger.Info().Msg(fmt.Sprintf("UserBindAccessServices: updating existing entity %s for user %s", entityID, bind.UserEmail))
			updateReq := map[string]interface{}{
				"policies": policies,
			}
			_, err = v.Client.Logical().WriteWithContext(ctx, "identity/entity/id/"+entityID, updateReq)
			if err != nil {
				v.Logger.Error().Msg(fmt.Sprintf("failed to update entity %s: %+v", entityID, err))
				continue
			}
			v.Logger.Info().Msg(fmt.Sprintf("UserBindAccessServices: entity %s updated for user %s", entityID, bind.UserEmail))
		}
	}
	v.Logger.Info().Msg("UserBindAccessServices: completed")
	return nil
}

func (v *Client) GetVaultAddr() string {
	return v.VaultAddr
}

// ---------------------------------------------------------------------
// внутренние хелперы
// ---------------------------------------------------------------------

func (v *Client) createOrUpdatePolicy(ctx context.Context, servicePath string, perm string) (string, error) {
	v.Logger.Info().Msg(fmt.Sprintf("createOrUpdatePolicy: path %s, permission %s", servicePath, perm))
	name := strings.ReplaceAll(sanitize(servicePath), "/", "-") + "-" + perm
	hcl := generatePolicyHCL(servicePath, perm)
	err := v.Client.Sys().PutPolicyWithContext(ctx, name, hcl)
	if err == nil {
		v.Logger.Info().Msg(fmt.Sprintf("createOrUpdatePolicy: policy %s created/updated", name))
	}
	return name, err
}

func (v *Client) createOrUpdateDefaultPolicy(ctx context.Context) (string, error) {
	err := v.Client.Sys().PutPolicyWithContext(
		ctx,
		"default",
		`
path "secret/metadata/services/" {
  capabilities = ["list"]
}

path "secret/metadata/" {
  capabilities = ["list"]
}
`)
	v.Logger.Info().Msg("createOrUpdateDefaultPolicy: default policy created")
	return "default", err
}

func generatePolicyHCL(servicePath string, perm string) string {
	p := sanitize(servicePath)
	caps := ""
	switch perm {
	case Read:
		caps = `capabilities = ["read","list"]`
	case Write:
		caps = `capabilities = ["create","read","update","delete","list"]`
	case None:
		caps = `capabilities = []`
	}
	return fmt.Sprintf(`
path "secret/data/%s/*" { %s }
path "secret/metadata/%s/*" { %s }
`, p, caps, p, caps)
}

var sanitizeRe = regexp.MustCompile(`[^a-zA-Z0-9\-_\./]`)

func sanitize(s string) string {
	s = sanitizeRe.ReplaceAllString(s, "-")
	s = strings.ReplaceAll(s, "..", "")
	return strings.Trim(s, "/")
}
