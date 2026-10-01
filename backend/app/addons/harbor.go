package addons

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/goccy/go-json"

	"github.com/cms-lab-core/cms-labs-api/backend/app/addons/vault"
	"github.com/cms-lab-core/cms-labs-api/shared/connection"
)

// HarborConfig contains the parameters for Harbor integration.
type HarborConfig struct {
	ApiURL        string
	AdminUsername string
	AdminPassword string
	BaseURL       string
}

// HarborProject is the payload for creating a project.
type HarborProject struct {
	ProjectName string `json:"project_name"`
	Public      bool   `json:"public"`
}

// HarborPermission defines a permission entry for a robot account.
type HarborPermission struct {
	Kind      string         `json:"kind"`
	Namespace string         `json:"namespace"`
	Access    []HarborAccess `json:"access"`
}

// HarborAccess defines a specific action on a resource.
type HarborAccess struct {
	Resource string `json:"resource"`
	Action   string `json:"action"`
}

// HarborRobotCreate is the payload for creating a robot account.
type HarborRobotCreate struct {
	Name        string             `json:"name"`
	Level       string             `json:"level"`
	Description string             `json:"description"`
	Duration    int                `json:"duration"`
	Permissions []HarborPermission `json:"permissions"`
}

// HarborRobotResponse is the response after successful robot creation.
type HarborRobotResponse struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Secret string `json:"secret"`
}

// HarborRepository is a minimal representation of a repository.
type HarborRepository struct {
	Name string `json:"name"`
}

// HarborArtifact is a minimal representation of an artifact.
type HarborArtifact struct {
	Digest string `json:"digest"`
}

// HarborAddonService implements AddonService for Harbor container registry.
type HarborAddonService struct {
	Config      *connection.AddonConfig
	Params      HarborConfig
	VaultClient vault.ClientInterface
	HTTPClient  *http.Client
}

// NewHarborAddonService creates a new Harbor service instance.
func NewHarborAddonService(cfg *connection.AddonConfig, vaultClient vault.ClientInterface) AddonService {
	apiUrl, _ := cfg.Params["apiUrl"].(string)
	username, _ := cfg.Params["adminUsername"].(string)
	password, _ := cfg.Params["adminPassword"].(string)
	baseUrl, _ := cfg.Params["baseUrl"].(string)
	params := HarborConfig{
		ApiURL:        apiUrl,
		AdminUsername: username,
		AdminPassword: password,
		BaseURL:       baseUrl,
	}
	return &HarborAddonService{
		Config:      cfg,
		Params:      params,
		VaultClient: vaultClient,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (s *HarborAddonService) GetType() connection.AddonType {
	return s.Config.Type
}

// generateProjectName creates a valid Harbor project name from the target name.
func (s *HarborAddonService) generateProjectName(targetName string) string {
	// Replace any non‑alphanumeric or underscore with underscore.
	re := regexp.MustCompile(`[^a-z0-9_]`)
	name := "svc_" + re.ReplaceAllString(strings.ToLower(targetName), "_")
	// Harbor project name length limit is 63 characters.
	if len(name) > 63 {
		name = name[:63]
	}
	return name
}

// doRequest performs an HTTP request to Harbor API with basic authentication.
func (s *HarborAddonService) doRequest(ctx context.Context, method, path string, body, out interface{}) error {
	url := s.Params.ApiURL + path
	var reqBody []byte
	var err error
	if body != nil {
		reqBody, err = json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal request body: %w", err)
		}
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewReader(reqBody))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.SetBasicAuth(s.Params.AdminUsername, s.Params.AdminPassword)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	// Accept 2xx and 409 (already exists) – we handle 409 specially in callers.
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("unexpected status: %s", resp.Status)
	}

	if out != nil && resp.StatusCode != http.StatusNoContent {
		if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
			return fmt.Errorf("decode response: %w", err)
		}
	}
	return nil
}

// createProject creates a new Harbor project. If it already exists (409), it logs a warning and returns nil.
func (s *HarborAddonService) createProject(ctx context.Context, projectName string, public bool) error {
	project := HarborProject{
		ProjectName: projectName,
		Public:      public,
	}
	err := s.doRequest(ctx, http.MethodPost, "/api/v2.0/projects", project, nil)
	if err != nil {
		// 409 Conflict means project already exists
		if strings.Contains(err.Error(), "409") {
			log.Printf("Project %s already exists", projectName)
			return nil
		}
		return fmt.Errorf("create project: %w", err)
	}
	return nil
}

// createRobot creates a robot account with push/pull permissions on the project.
func (s *HarborAddonService) createRobot(ctx context.Context, projectName string) (*HarborRobotResponse, error) {
	// Define the permissions exactly as in TypeScript.
	permissions := []HarborPermission{
		{
			Kind:      "project",
			Namespace: projectName,
			Access: []HarborAccess{
				{Resource: "tag", Action: "create"},
				{Resource: "tag", Action: "delete"},
				{Resource: "repository", Action: "list"},
				{Resource: "tag", Action: "list"},
				{Resource: "repository", Action: "push"},
				{Resource: "repository", Action: "read"},
				{Resource: "repository", Action: "update"},
				{Resource: "repository", Action: "pull"},
				{Resource: "repository", Action: "delete"},
			},
		},
	}

	robot := HarborRobotCreate{
		Name:        projectName,
		Level:       "system",
		Description: fmt.Sprintf("Robot account for project %s", projectName),
		Duration:    -1, // never expires
		Permissions: permissions,
	}

	var resp HarborRobotResponse
	err := s.doRequest(ctx, http.MethodPost, "/api/v2.0/robots", robot, &resp)
	if err != nil {
		return nil, fmt.Errorf("create robot: %w", err)
	}
	return &resp, nil
}

// setProjectPublicAccess updates the public flag of a project.
func (s *HarborAddonService) setProjectPublicAccess(ctx context.Context, projectName string, public bool) error {
	payload := map[string]bool{"public": public}
	err := s.doRequest(ctx, http.MethodPut, fmt.Sprintf("/api/v2.0/projects/%s", projectName), payload, nil)
	if err != nil {
		return fmt.Errorf("set project public access: %w", err)
	}
	return nil
}

// deleteAllArtifacts removes every artifact from every repository in the project.
func (s *HarborAddonService) deleteAllArtifacts(ctx context.Context, projectName string) error {
	// List repositories
	var repos []HarborRepository
	err := s.doRequest(ctx, http.MethodGet, fmt.Sprintf("/api/v2.0/projects/%s/repositories", projectName), nil, &repos)
	if err != nil {
		// If the project is already gone, we can ignore.
		log.Printf("Could not list repositories for %s: %v", projectName, err)
		return nil
	}

	for _, repo := range repos {
		// Extract short name (without project prefix)
		repoName := strings.TrimPrefix(repo.Name, projectName+"/")
		if err := s.deleteRepositoryArtifacts(ctx, projectName, repoName); err != nil {
			log.Printf("Error deleting artifacts from %s: %v", repoName, err)
		}
	}
	return nil
}

// deleteRepositoryArtifacts deletes all artifacts in a specific repository.
func (s *HarborAddonService) deleteRepositoryArtifacts(ctx context.Context, projectName, repoName string) error {
	// List artifacts
	path := fmt.Sprintf("/api/v2.0/projects/%s/repositories/%s/artifacts",
		projectName, urlPathEscape(repoName))
	var artifacts []HarborArtifact
	err := s.doRequest(ctx, http.MethodGet, path, nil, &artifacts)
	if err != nil {
		// Repository may be empty or already deleted.
		return nil
	}

	for _, art := range artifacts {
		delPath := fmt.Sprintf("/api/v2.0/projects/%s/repositories/%s/artifacts/%s",
			projectName, urlPathEscape(repoName), art.Digest)
		_ = s.doRequest(ctx, http.MethodDelete, delPath, nil, nil) // ignore errors
	}
	return nil
}

// deleteProject removes a Harbor project.
func (s *HarborAddonService) deleteProject(ctx context.Context, projectName string) error {
	err := s.doRequest(ctx, http.MethodDelete, fmt.Sprintf("/api/v2.0/projects/%s", projectName), nil, nil)
	if err != nil {
		if strings.Contains(err.Error(), "404") {
			log.Printf("Project %s not found, ignoring", projectName)
			return nil
		}
		return fmt.Errorf("delete project: %w", err)
	}
	return nil
}

// deleteRobot removes the robot account associated with the project.
func (s *HarborAddonService) deleteRobot(ctx context.Context, projectName string) error {
	// Query robot by name (Harbor returns list of robots matching the query)
	var robots []struct {
		ID int `json:"id"`
	}
	query := fmt.Sprintf("/api/v2.0/robots?q=name=%s", projectName)
	err := s.doRequest(ctx, http.MethodGet, query, nil, &robots)
	if err != nil {
		log.Printf("Could not query robots for %s: %v", projectName, err)
		return nil
	}
	if len(robots) == 0 {
		return nil
	}
	robotID := robots[0].ID
	err = s.doRequest(ctx, http.MethodDelete, fmt.Sprintf("/api/v2.0/robots/%d", robotID), nil, nil)
	if err != nil {
		log.Printf("Error deleting robot %d: %v", robotID, err)
	}
	return nil
}

// cleanupOnError attempts to delete the project and the robot, ignoring any errors.
func (s *HarborAddonService) cleanupOnError(ctx context.Context, projectName string) {
	_ = s.deleteProject(ctx, projectName)
	_ = s.deleteRobot(ctx, projectName)
}

// Create provisions a new Harbor project and robot account.
func (s *HarborAddonService) Create(ctx context.Context, targetName string, name *string) (*AddonOperationConfig, error) {
	projectName := s.generateProjectName(targetName)

	// 1. Create project (if it already exists, it's okay)
	if err := s.createProject(ctx, projectName, true); err != nil {
		return nil, err
	}

	// 2. Create robot account
	robot, err := s.createRobot(ctx, projectName)
	if err != nil {
		s.cleanupOnError(ctx, projectName)
		return nil, err
	}

	// 3. Ensure project is public
	if err := s.setProjectPublicAccess(ctx, projectName, true); err != nil {
		s.cleanupOnError(ctx, projectName)
		return nil, err
	}

	// 4. Store credentials in Vault
	extName := fmt.Sprintf("%s_%s", s.Config.Type, s.Config.Tag)
	err = s.VaultClient.CreateServiceExtension(
		ctx,
		targetName,
		strings.ToLower(extName),
		map[string]interface{}{
			"HARBOR_REGISTRY_URL":   s.Params.BaseURL,
			"HARBOR_REGISTRY_LOGIN": robot.Name,
			"HARBOR_REGISTRY_TOKEN": robot.Secret,
			"HARBOR_PROJECT":        projectName,
		},
	)
	if err != nil {
		s.cleanupOnError(ctx, projectName)
		return nil, err
	}

	// Return configuration with direct link to the project
	url := fmt.Sprintf("%s/harbor/projects?globalSearch=%s", s.Params.BaseURL, projectName)
	return &AddonOperationConfig{
		Name: projectName,
		URL:  &url,
	}, nil
}

// Delete removes the project and its associated robot account.
func (s *HarborAddonService) Delete(ctx context.Context, targetName string, cfg *AddonOperationConfig) (*AddonOperationConfig, error) {
	if cfg == nil || cfg.Name == "" {
		return nil, fmt.Errorf("project name is required for deletion")
	}
	projectName := cfg.Name

	// 1. Delete all artifacts (best effort)
	_ = s.deleteAllArtifacts(ctx, projectName)

	// 2. Delete robot account
	_ = s.deleteRobot(ctx, projectName)

	// 3. Delete project
	if err := s.deleteProject(ctx, projectName); err != nil {
		return nil, err
	}

	return &AddonOperationConfig{Name: projectName}, nil
}

// Reset rotates the robot account secret.
func (s *HarborAddonService) Reset(ctx context.Context, targetName string, cfg *AddonOperationConfig) (*AddonOperationConfig, error) {
	if cfg == nil || cfg.Name == "" {
		return nil, fmt.Errorf("project name is required for reset")
	}
	projectName := cfg.Name

	// 1. Delete existing robot
	_ = s.deleteRobot(ctx, projectName)

	// 2. Create new robot (with new secret)
	robot, err := s.createRobot(ctx, projectName)
	if err != nil {
		return nil, err
	}

	// 3. Update Vault with new credentials
	extName := fmt.Sprintf("%s_%s", s.Config.Type, s.Config.Tag)
	err = s.VaultClient.CreateServiceExtension(
		ctx,
		targetName,
		strings.ToLower(extName),
		map[string]interface{}{
			"HARBOR_REGISTRY_URL":   s.Params.BaseURL,
			"HARBOR_REGISTRY_LOGIN": robot.Name,
			"HARBOR_REGISTRY_TOKEN": robot.Secret,
			"HARBOR_PROJECT":        projectName,
		},
	)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/harbor/projects?globalSearch=%s", s.Params.BaseURL, projectName)
	return &AddonOperationConfig{
		Name: projectName,
		URL:  &url,
	}, nil
}

// urlPathEscape is a simple replacement for url.PathEscape to avoid import conflicts.
func urlPathEscape(s string) string {
	return strings.ReplaceAll(s, "/", "%2F")
}
