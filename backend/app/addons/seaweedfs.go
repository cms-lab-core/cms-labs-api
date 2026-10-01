package addons

import (
	"context"
	"crypto/rand"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/goccy/go-json"

	"github.com/cms-lab-core/cms-labs-api/backend/app/addons/vault"
	"github.com/cms-lab-core/cms-labs-api/shared/connection"
)

// SeaweedFSConfig holds the parameters specific to SeaweedFS.
type SeaweedFSConfig struct {
	Endpoint       string
	ServiceInVault string
	GitURL         string
	GitRepo        string
	GitToken       string
	GitBranch      string
}

// SeaweedFSAddonService implements AddonService for SeaweedFS S3-compatible buckets.
type SeaweedFSAddonService struct {
	Config      *connection.AddonConfig
	Params      SeaweedFSConfig
	VaultClient vault.ClientInterface
	HTTPClient  *http.Client
}

// NewSeaweedFSAddonService creates a new SeaweedFS service instance.
func NewSeaweedFSAddonService(cfg *connection.AddonConfig, vaultClient vault.ClientInterface) AddonService {
	endpoint, _ := cfg.Params["endpoint"].(string)
	serviceInVault, _ := cfg.Params["service_in_vault"].(string)
	gitUrl, _ := cfg.Params["git_url"].(string)
	gitRepo, _ := cfg.Params["git_repo"].(string)
	gitToken, _ := cfg.Params["git_token"].(string)
	gitBranch, _ := cfg.Params["git_branch"].(string)
	params := SeaweedFSConfig{
		Endpoint:       endpoint,
		ServiceInVault: serviceInVault,
		GitURL:         gitUrl,
		GitRepo:        gitRepo,
		GitToken:       gitToken,
		GitBranch:      gitBranch,
	}
	return &SeaweedFSAddonService{
		Config:      cfg,
		Params:      params,
		VaultClient: vaultClient,
		HTTPClient:  &http.Client{Timeout: 30 * time.Second},
	}
}

// aimPolicyTemplate is the base IAM-like policy for SeaweedFS.
var aimPolicyTemplate = map[string]interface{}{
	"identities": []interface{}{
		map[string]interface{}{
			"name": "{{bucketName}}",
			"credentials": []interface{}{
				map[string]interface{}{
					"accessKey": "{{accessKey}}",
					"secretKey": "{{secretKey}}",
				},
			},
			"actions": []string{
				"Read:{{bucketName}}",
				"Write:{{bucketName}}",
				"List:{{bucketName}}",
				"Tagging:{{bucketName}}",
				"Admin:{{bucketName}}",
			},
		},
	},
}

// generateAIMPolicy replaces placeholders and returns the policy as a map.
func generateAIMPolicy(bucketName, accessKey, secretKey string) (map[string]interface{}, error) {
	// Serialise to JSON, replace placeholders, then deserialise.
	raw, err := json.Marshal(aimPolicyTemplate)
	if err != nil {
		return nil, err
	}
	s := string(raw)
	s = strings.ReplaceAll(s, "{{bucketName}}", bucketName)
	s = strings.ReplaceAll(s, "{{accessKey}}", accessKey)
	s = strings.ReplaceAll(s, "{{secretKey}}", secretKey)

	var policy map[string]interface{}
	if err := json.Unmarshal([]byte(s), &policy); err != nil {
		return nil, err
	}
	return policy, nil
}

// generateBucketName creates a safe S3 bucket name from the target name.
func (s *SeaweedFSAddonService) generateBucketName(targetName string) string {
	// Replace non‑alphanumeric with hyphen, lowercase.
	re := regexp.MustCompile(`[^a-z0-9-]`)
	prefix := re.ReplaceAllString(strings.ToLower(targetName), "-")
	bucketName := "svc-" + prefix
	// Limit to 63 characters and trim leading/trailing hyphens.
	if len(bucketName) > 63 {
		bucketName = bucketName[:63]
	}
	bucketName = strings.Trim(bucketName, "-")
	return bucketName
}

// generateAccessKey creates an S3‑compatible access key.
// Format: SW + last 4 chars of base36 timestamp + 16 random uppercase/digits → max 20 chars.
func (s *SeaweedFSAddonService) generateAccessKey() string {
	const letters = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	timestamp := time.Now().UnixMilli()
	tsPart := fmt.Sprintf("%x", timestamp) // hex is shorter than base36; we take last 4
	if len(tsPart) > 4 {
		tsPart = tsPart[len(tsPart)-4:]
	}
	randomPart := randomStringFromCharset(16, letters)
	key := "SW" + tsPart + randomPart
	if len(key) > 20 {
		key = key[:20]
	}
	return strings.ToUpper(key)
}

// generateSecretKey creates a 40‑character secret key.
// Character set: A-Z a-z 0-9 + /
func (s *SeaweedFSAddonService) generateSecretKey() string {
	const letters = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"
	return randomStringFromCharset(40, letters)
}

// randomStringFromCharset returns a random string of length n using the given charset.
func randomStringFromCharset(n int, charset string) string {
	b := make([]byte, n)
	_, _ = rand.Read(b)
	for i := range b {
		b[i] = charset[b[i]%byte(len(charset))]
	}
	return string(b)
}

// Create provisions a new SeaweedFS bucket and associated credentials.
func (s *SeaweedFSAddonService) Create(ctx context.Context, targetName string, name *string) (*AddonOperationConfig, error) {
	bucketName := s.generateBucketName(targetName)
	accessKey := s.generateAccessKey()
	secretKey := s.generateSecretKey()

	// 1. Create AIM policy in Vault.
	policy, err := generateAIMPolicy(bucketName, accessKey, secretKey)
	if err != nil {
		return nil, fmt.Errorf("failed to generate AIM policy: %w", err)
	}
	err = s.VaultClient.CreateServiceExtension(
		ctx,
		s.Params.ServiceInVault,
		bucketName,
		policy,
	)
	if err != nil {
		return nil, err
	}

	// 2. Store bucket credentials in Vault under the target.
	extName := fmt.Sprintf("%s_%s", s.Config.Type, s.Config.Tag)
	err = s.VaultClient.CreateServiceExtension(
		ctx,
		targetName,
		strings.ToLower(extName),
		map[string]interface{}{
			"SEAWEEDFS_ENDPOINT":    s.Params.Endpoint,
			"SEAWEEDFS_REGION":      "us-east-1",
			"SEAWEEDFS_BUCKET":      bucketName,
			"AWS_ACCESS_KEY_ID":     accessKey,
			"AWS_SECRET_ACCESS_KEY": secretKey,
			"AWS_USER_NAME":         bucketName,
		},
	)
	if err != nil {
		return nil, err
	}

	// 3. Trigger GitLab pipeline to update SeaweedFS configuration.
	if err := s.update(ctx); err != nil {
		// Log but do not fail the creation.
		log.Printf("SeaweedFS update pipeline trigger failed (non‑fatal): %v", err)
	}

	return &AddonOperationConfig{
		Name:          bucketName,
		CurrentSizeMb: intPtr(0),
		MaxSizeMb:     intPtr(1024),
	}, nil
}

// Delete removes the bucket, its AIM policy, and its credentials.
func (s *SeaweedFSAddonService) Delete(ctx context.Context, targetName string, cfg *AddonOperationConfig) (*AddonOperationConfig, error) {
	if cfg == nil || cfg.Name == "" {
		return nil, fmt.Errorf("bucket name is required for deletion")
	}
	bucketName := cfg.Name

	// 1. Delete AIM policy (set to empty).
	err := s.VaultClient.CreateServiceExtension(
		ctx,
		s.Params.ServiceInVault,
		bucketName,
		map[string]interface{}{}, // empty object
	)
	if err != nil {
		return nil, err
	}

	// 2. Delete credentials under the target.
	extName := fmt.Sprintf("%s_%s", s.Config.Type, s.Config.Tag)
	err = s.VaultClient.CreateServiceExtension(
		ctx,
		bucketName, // In TS it uses `config.name` as targetName? Actually TS: await vaultClient.createServiceExtension(config.name, ...)
		strings.ToLower(extName),
		map[string]interface{}{},
	)
	if err != nil {
		return nil, err
	}

	// 3. Trigger GitLab pipeline.
	if err := s.update(ctx); err != nil {
		log.Printf("SeaweedFS update pipeline trigger failed (non‑fatal): %v", err)
	}

	return &AddonOperationConfig{Name: bucketName}, nil
}

// Reset recreates the bucket and credentials (essentially a fresh provision).
func (s *SeaweedFSAddonService) Reset(ctx context.Context, targetName string, cfg *AddonOperationConfig) (*AddonOperationConfig, error) {
	// Reset simply calls Create again, ignoring the old config.
	return s.Create(ctx, targetName, nil)
}

// update triggers a GitLab pipeline to notify SeaweedFS configuration changes.
func (s *SeaweedFSAddonService) update(ctx context.Context) error {
	if s.Params.GitURL == "" || s.Params.GitRepo == "" || s.Params.GitToken == "" {
		// Not configured – skip.
		return nil
	}

	// Build URL: <gitUrl>/api/v4/projects/<gitRepo>/trigger/pipeline
	u := fmt.Sprintf("%s/api/v4/projects/%s/trigger/pipeline",
		strings.TrimRight(s.Params.GitURL, "/"),
		strings.Trim(s.Params.GitRepo, "/"),
	)

	// Prepare form data.
	form := url.Values{}
	form.Set("token", s.Params.GitToken)
	form.Set("ref", s.Params.GitBranch)
	form.Set("variables[SEAWEEDFS_UPDATE]", "true")

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := s.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("gitlab pipeline trigger failed with status: %s", resp.Status)
	}
	return nil
}

func (s *SeaweedFSAddonService) GetType() connection.AddonType {
	return s.Config.Type
}
