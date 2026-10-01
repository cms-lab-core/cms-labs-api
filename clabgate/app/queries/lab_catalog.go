package queries

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	resty "github.com/go-resty/resty/v2"
	k8syaml "k8s.io/apimachinery/pkg/util/yaml"
	"sigs.k8s.io/yaml"
)

const gitLabPageSize = 100

type repositoryProvider string

const (
	providerGitLab repositoryProvider = "gitlab"
	providerGitHub repositoryProvider = "github"
	providerGit    repositoryProvider = "git"
)

type repositoryLocation struct {
	provider repositoryProvider
	project  string
	apiBase  string
	cloneURL string
}

type LabCatalog struct {
	projectURL string
	branch     string
	token      string
	client     *resty.Client
}

type LabBundle struct {
	Manifest   string
	Revision   string
	Files      []string
	ProjectURL string
}

type gitLabTreeEntry struct {
	Path string `json:"path"`
	Type string `json:"type"`
}

type gitLabCommit struct {
	ID string `json:"id"`
}

type gitHubCommit struct {
	SHA string `json:"sha"`
}

type gitHubTree struct {
	Tree      []gitLabTreeEntry `json:"tree"`
	Truncated bool              `json:"truncated"`
}

func NewLabCatalog(projectURL, branch, token string, client *resty.Client) *LabCatalog {
	return &LabCatalog{
		projectURL: strings.TrimRight(projectURL, "/"),
		branch:     branch,
		token:      token,
		client:     client,
	}
}

// Bundle resolves one directory in a GitLab or GitHub project and downloads every YAML
// file below it in deterministic order. Kubernetes object validation happens
// later, before any document is applied.
func (c *LabCatalog) Bundle(ctx context.Context, labsPath string) (LabBundle, error) {
	return c.BundleAt(ctx, labsPath, "")
}

// BundleFor resolves a lab using repository metadata pinned to the CMS route.
// The configured catalog remains a fallback for legacy attempts.
func (c *LabCatalog) BundleFor(ctx context.Context, projectURL, branch, labsPath, pinnedRevision string) (LabBundle, error) {
	scoped := *c
	if strings.TrimSpace(projectURL) != "" {
		scoped.projectURL = strings.TrimRight(strings.TrimSpace(projectURL), "/")
	}
	if strings.TrimSpace(branch) != "" {
		scoped.branch = strings.TrimSpace(branch)
	}
	return scoped.BundleAt(ctx, labsPath, pinnedRevision)
}

// BundleAt uses an already resolved commit when retrying a partially created
// session, so a moving branch cannot alter that session between retries.
func (c *LabCatalog) BundleAt(ctx context.Context, labsPath, pinnedRevision string) (LabBundle, error) {
	if c.projectURL == "" {
		return LabBundle{}, fmt.Errorf("CMS_TASK_URL is not configured")
	}
	repository, err := parseRepositoryURL(c.projectURL)
	if err != nil {
		return LabBundle{}, err
	}
	cleaned, err := cleanRepositoryPath(labsPath)
	if err != nil {
		return LabBundle{}, err
	}
	// The unauthenticated GitHub REST API is limited per runner IP. Codespaces
	// and hosted CI can therefore receive a 403 even for a public repository.
	// Git's smart HTTP protocol has no such API quota and still gives us an
	// immutable commit to pin retries to.
	if repository.provider == providerGit || (repository.provider == providerGitHub && c.token == "") {
		return c.bundleFromGit(ctx, repository, cleaned, pinnedRevision)
	}

	revision := strings.TrimSpace(pinnedRevision)
	if revision == "" {
		revision, err = c.resolveRevision(ctx, repository)
		if err != nil {
			return LabBundle{}, err
		}
	}
	files, err := c.listYAMLFiles(ctx, repository, cleaned, revision)
	if err != nil {
		return LabBundle{}, err
	}

	documents := make([]string, 0, len(files))
	for _, file := range files {
		response, requestErr := c.downloadFile(ctx, repository, file, revision)
		if requestErr != nil {
			return LabBundle{}, fmt.Errorf("download task manifest %q: %w", file, requestErr)
		}
		if response.StatusCode() < http.StatusOK || response.StatusCode() >= http.StatusMultipleChoices {
			return LabBundle{}, fmt.Errorf("download task manifest %q: HTTP %d", file, response.StatusCode())
		}
		if value := strings.TrimSpace(response.String()); value != "" {
			documents = append(documents, value)
		}
	}

	manifest := strings.Join(filterTopologyDocuments(documents), "\n---\n")
	if strings.TrimSpace(manifest) == "" {
		return LabBundle{}, fmt.Errorf("lab directory %q contains no clabernetes Topology manifests", repositoryRootLabel(cleaned))
	}
	return LabBundle{
		Manifest:   manifest,
		Revision:   revision,
		Files:      files,
		ProjectURL: c.projectURL,
	}, nil
}

func (c *LabCatalog) bundleFromGit(
	ctx context.Context,
	repository repositoryLocation,
	labsPath string,
	pinnedRevision string,
) (LabBundle, error) {
	checkoutRoot, err := os.MkdirTemp("", "cms-labs-task-*")
	if err != nil {
		return LabBundle{}, fmt.Errorf("create task checkout: %w", err)
	}
	defer os.RemoveAll(checkoutRoot)

	repositoryDir := filepath.Join(checkoutRoot, "repository")
	revision := strings.TrimSpace(pinnedRevision)
	if revision == "" {
		branch := strings.TrimSpace(c.branch)
		if branch == "" {
			branch = "master"
		}
		if err = runGit(ctx, "", "clone", "--quiet", "--depth=1", "--single-branch", "--branch", branch,
			"--", repository.cloneURL, repositoryDir); err != nil {
			return LabBundle{}, fmt.Errorf("clone repository ref %q: %w", branch, err)
		}
	} else {
		if !isGitCommit(revision) {
			return LabBundle{}, fmt.Errorf("invalid pinned repository revision %q", revision)
		}
		if err = os.Mkdir(repositoryDir, 0o700); err != nil {
			return LabBundle{}, fmt.Errorf("create task repository directory: %w", err)
		}
		if err = runGit(ctx, repositoryDir, "init", "--quiet"); err != nil {
			return LabBundle{}, fmt.Errorf("initialize task repository: %w", err)
		}
		if err = runGit(ctx, repositoryDir, "remote", "add", "origin", repository.cloneURL); err != nil {
			return LabBundle{}, fmt.Errorf("configure task repository: %w", err)
		}
		if err = runGit(ctx, repositoryDir, "fetch", "--quiet", "--depth=1", "origin", revision); err != nil {
			return LabBundle{}, fmt.Errorf("fetch pinned repository revision %q: %w", revision, err)
		}
		if err = runGit(ctx, repositoryDir, "checkout", "--quiet", "--detach", "FETCH_HEAD"); err != nil {
			return LabBundle{}, fmt.Errorf("checkout pinned repository revision %q: %w", revision, err)
		}
	}

	resolvedRevision, err := gitOutput(ctx, repositoryDir, "rev-parse", "HEAD")
	if err != nil {
		return LabBundle{}, fmt.Errorf("resolve checked out repository revision: %w", err)
	}
	resolvedRevision = strings.TrimSpace(resolvedRevision)
	if revision != "" && !strings.EqualFold(revision, resolvedRevision) {
		return LabBundle{}, fmt.Errorf("pinned repository revision mismatch: expected %s, got %s", revision, resolvedRevision)
	}

	label := repositoryRootLabel(labsPath)
	labDirectory := repositoryDir
	if labsPath != "" {
		labDirectory = filepath.Join(repositoryDir, filepath.FromSlash(labsPath))
	}
	info, err := os.Stat(labDirectory)
	if err != nil {
		if os.IsNotExist(err) {
			return LabBundle{}, fmt.Errorf("lab directory %q was not found in repository", label)
		}
		return LabBundle{}, fmt.Errorf("inspect lab directory %q: %w", label, err)
	}
	if !info.IsDir() {
		return LabBundle{}, fmt.Errorf("lab path %q is not a directory", label)
	}

	files := make([]string, 0)
	err = filepath.WalkDir(labDirectory, func(filePath string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if !entry.Type().IsRegular() {
			return nil
		}
		extension := strings.ToLower(filepath.Ext(entry.Name()))
		if extension != ".yaml" && extension != ".yml" {
			return nil
		}
		relative, relativeErr := filepath.Rel(repositoryDir, filePath)
		if relativeErr != nil {
			return relativeErr
		}
		files = append(files, filepath.ToSlash(relative))
		return nil
	})
	if err != nil {
		return LabBundle{}, fmt.Errorf("list task manifests in %q: %w", label, err)
	}
	if len(files) == 0 {
		return LabBundle{}, fmt.Errorf("lab directory %q contains no YAML manifests", label)
	}
	sort.Strings(files)

	documents := make([]string, 0, len(files))
	for _, file := range files {
		content, readErr := os.ReadFile(filepath.Join(repositoryDir, filepath.FromSlash(file)))
		if readErr != nil {
			return LabBundle{}, fmt.Errorf("read task manifest %q: %w", file, readErr)
		}
		if value := strings.TrimSpace(string(content)); value != "" {
			documents = append(documents, value)
		}
	}

	manifest := strings.Join(filterTopologyDocuments(documents), "\n---\n")
	if strings.TrimSpace(manifest) == "" {
		return LabBundle{}, fmt.Errorf("lab directory %q contains no clabernetes Topology manifests", label)
	}
	return LabBundle{
		Manifest:   manifest,
		Revision:   resolvedRevision,
		Files:      files,
		ProjectURL: c.projectURL,
	}, nil
}

// filterTopologyDocuments keeps only the objects session provisioning accepts:
// v1 ConfigMap and clabernetes Topology. A lab repository also carries CI
// workflows and its own deployment manifests, and those must not reach the
// session namespace. Allowed objects are re-emitted one document per manifest,
// so a mixed file contributes only its lab objects.
func filterTopologyDocuments(documents []string) []string {
	kept := make([]string, 0, len(documents))
	for _, document := range documents {
		decoder := k8syaml.NewYAMLOrJSONDecoder(strings.NewReader(document), 4096)
		for {
			raw := map[string]any{}
			if err := decoder.Decode(&raw); errors.Is(err, io.EOF) {
				break
			} else if err != nil {
				// A file that is not Kubernetes YAML at all is not a manifest.
				break
			}
			if len(raw) == 0 {
				continue
			}
			apiVersion, _ := raw["apiVersion"].(string)
			kind, _ := raw["kind"].(string)
			if (apiVersion != "v1" || kind != "ConfigMap") &&
				(apiVersion != "c9s.run/v1alpha1" || kind != "Topology") {
				continue
			}
			encoded, err := yaml.Marshal(raw)
			if err != nil {
				continue
			}
			kept = append(kept, strings.TrimSpace(string(encoded)))
		}
	}
	return kept
}

func runGit(ctx context.Context, directory string, args ...string) error {
	_, err := gitOutput(ctx, directory, args...)
	return err
}

func gitOutput(ctx context.Context, directory string, args ...string) (string, error) {
	command := exec.CommandContext(ctx, "git", args...)
	command.Dir = directory
	output, err := command.CombinedOutput()
	if err != nil {
		message := strings.TrimSpace(string(output))
		if message == "" {
			return "", err
		}
		return "", fmt.Errorf("%w: %s", err, message)
	}
	return string(output), nil
}

func isGitCommit(value string) bool {
	if len(value) != 40 && len(value) != 64 {
		return false
	}
	for _, character := range value {
		if (character < '0' || character > '9') && (character < 'a' || character > 'f') &&
			(character < 'A' || character > 'F') {
			return false
		}
	}
	return true
}

func (c *LabCatalog) resolveRevision(ctx context.Context, repository repositoryLocation) (string, error) {
	branch := strings.TrimSpace(c.branch)
	if branch == "" {
		branch = "master"
	}
	if repository.provider == providerGitHub {
		commit := gitHubCommit{}
		response, err := c.request(ctx, repository.provider).SetResult(&commit).Get(
			repository.apiBase + "/repos/" + repository.project + "/commits/" + url.PathEscape(branch),
		)
		if err != nil {
			return "", fmt.Errorf("resolve GitHub ref %q: %w", branch, err)
		}
		if response.StatusCode() < http.StatusOK || response.StatusCode() >= http.StatusMultipleChoices {
			return "", fmt.Errorf("resolve GitHub ref %q: HTTP %d", branch, response.StatusCode())
		}
		if commit.SHA == "" {
			return "", fmt.Errorf("resolve GitHub ref %q: empty commit SHA", branch)
		}
		return commit.SHA, nil
	}

	commit := gitLabCommit{}
	response, err := c.request(ctx, repository.provider).SetResult(&commit).Get(
		repository.apiBase + "/projects/" + url.PathEscape(repository.project) + "/repository/commits/" + url.PathEscape(branch),
	)
	if err != nil {
		return "", fmt.Errorf("resolve GitLab ref %q: %w", branch, err)
	}
	if response.StatusCode() < http.StatusOK || response.StatusCode() >= http.StatusMultipleChoices {
		return "", fmt.Errorf("resolve GitLab ref %q: HTTP %d", branch, response.StatusCode())
	}
	if commit.ID == "" {
		return "", fmt.Errorf("resolve GitLab ref %q: empty commit ID", branch)
	}
	return commit.ID, nil
}

func (c *LabCatalog) listYAMLFiles(ctx context.Context, repository repositoryLocation, labsPath, revision string) ([]string, error) {
	label := repositoryRootLabel(labsPath)
	if repository.provider == providerGitHub {
		tree := gitHubTree{}
		response, err := c.request(ctx, repository.provider).
			SetResult(&tree).
			SetQueryParam("recursive", "1").
			Get(repository.apiBase + "/repos/" + repository.project + "/git/trees/" + url.PathEscape(revision))
		if err != nil {
			return nil, fmt.Errorf("list task manifests in %q: %w", label, err)
		}
		if response.StatusCode() == http.StatusNotFound {
			return nil, fmt.Errorf("lab directory %q was not found in GitHub", label)
		}
		if response.StatusCode() < http.StatusOK || response.StatusCode() >= http.StatusMultipleChoices {
			return nil, fmt.Errorf("list task manifests in %q: HTTP %d", label, response.StatusCode())
		}
		if tree.Truncated {
			return nil, fmt.Errorf("GitHub repository tree is truncated; split the task catalog or use GitLab")
		}
		files := filterYAMLFiles(tree.Tree, labsPath)
		if len(files) == 0 {
			return nil, fmt.Errorf("lab directory %q contains no YAML manifests", label)
		}
		return files, nil
	}

	projectAPI := repository.apiBase + "/projects/" + url.PathEscape(repository.project)
	files := make([]string, 0)
	for pageNumber := 1; ; pageNumber++ {
		entries := []gitLabTreeEntry{}
		request := c.request(ctx, repository.provider).
			SetResult(&entries).
			SetQueryParams(map[string]string{
				"ref": revision, "recursive": "true",
				"per_page": strconv.Itoa(gitLabPageSize), "page": strconv.Itoa(pageNumber),
			})
		if labsPath != "" {
			request = request.SetQueryParam("path", labsPath)
		}
		response, err := request.Get(projectAPI + "/repository/tree")
		if err != nil {
			return nil, fmt.Errorf("list task manifests in %q: %w", label, err)
		}
		if response.StatusCode() == http.StatusNotFound {
			return nil, fmt.Errorf("lab directory %q was not found in GitLab", label)
		}
		if response.StatusCode() < http.StatusOK || response.StatusCode() >= http.StatusMultipleChoices {
			return nil, fmt.Errorf("list task manifests in %q: HTTP %d", label, response.StatusCode())
		}
		for _, entry := range entries {
			extension := strings.ToLower(path.Ext(entry.Path))
			if entry.Type == "blob" && (extension == ".yaml" || extension == ".yml") {
				files = append(files, entry.Path)
			}
		}
		if response.Header().Get("X-Next-Page") == "" {
			break
		}
	}
	if len(files) == 0 {
		return nil, fmt.Errorf("lab directory %q contains no YAML manifests", label)
	}
	sort.Strings(files)
	return files, nil
}

func filterYAMLFiles(entries []gitLabTreeEntry, labsPath string) []string {
	prefix := ""
	if labsPath != "" {
		prefix = strings.TrimSuffix(labsPath, "/") + "/"
	}
	files := make([]string, 0)
	for _, entry := range entries {
		extension := strings.ToLower(path.Ext(entry.Path))
		if entry.Type == "blob" && strings.HasPrefix(entry.Path, prefix) && (extension == ".yaml" || extension == ".yml") {
			files = append(files, entry.Path)
		}
	}
	sort.Strings(files)
	return files
}

func (c *LabCatalog) downloadFile(ctx context.Context, repository repositoryLocation, file, revision string) (*resty.Response, error) {
	if repository.provider == providerGitHub {
		return c.request(ctx, repository.provider).
			SetHeader("Accept", "application/vnd.github.raw+json").
			SetQueryParam("ref", revision).
			Get(repository.apiBase + "/repos/" + repository.project + "/contents/" + escapeRepositoryPath(file))
	}
	return c.request(ctx, repository.provider).
		SetQueryParam("ref", revision).
		Get(repository.apiBase + "/projects/" + url.PathEscape(repository.project) + "/repository/files/" + url.PathEscape(file) + "/raw")
}

func (c *LabCatalog) request(ctx context.Context, provider repositoryProvider) *resty.Request {
	request := c.client.R().SetContext(ctx)
	if c.token != "" {
		if provider == providerGitHub {
			request.SetAuthToken(c.token)
		} else {
			request.SetHeader("PRIVATE-TOKEN", c.token)
		}
	}
	if provider == providerGitHub {
		request.SetHeader("X-GitHub-Api-Version", "2022-11-28")
	}
	return request
}

func parseRepositoryURL(raw string) (repositoryLocation, error) {
	parsed, err := url.Parse(raw)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return repositoryLocation{}, fmt.Errorf("CMS_TASK_URL must be a full GitLab or GitHub project URL")
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return repositoryLocation{}, fmt.Errorf("CMS_TASK_URL must use HTTP or HTTPS")
	}
	if parsed.RawQuery != "" || parsed.Fragment != "" || parsed.User != nil {
		return repositoryLocation{}, fmt.Errorf("CMS_TASK_URL must not contain credentials, query or fragment")
	}
	project := strings.Trim(strings.TrimSuffix(parsed.EscapedPath(), ".git"), "/")
	decoded, decodeErr := url.PathUnescape(project)
	if decodeErr != nil || decoded == "" || !strings.Contains(decoded, "/") {
		return repositoryLocation{}, fmt.Errorf("CMS_TASK_URL must include repository owner and project")
	}
	if strings.EqualFold(parsed.Hostname(), "github.com") {
		if len(strings.Split(decoded, "/")) != 2 {
			return repositoryLocation{}, fmt.Errorf("GitHub CMS_TASK_URL must include owner and repository")
		}
		return repositoryLocation{
			provider: providerGitHub,
			project:  escapeRepositoryPath(decoded),
			apiBase:  "https://api.github.com",
			cloneURL: parsed.Scheme + "://" + parsed.Host + "/" + decoded + ".git",
		}, nil
	}
	if strings.HasSuffix(strings.ToLower(parsed.Path), ".git") {
		return repositoryLocation{
			provider: providerGit,
			project:  decoded,
			cloneURL: parsed.String(),
		}, nil
	}
	return repositoryLocation{
		provider: providerGitLab,
		project:  decoded,
		apiBase:  parsed.Scheme + "://" + parsed.Host + "/api/v4",
		cloneURL: parsed.Scheme + "://" + parsed.Host + "/" + decoded + ".git",
	}, nil
}

// ParseRepositoryLink splits a routing Git link into the clone URL and the
// pinned ref. The ref comes from an optional `#ref` fragment, so
// https://git.example.org/course/lab.git#main pins branch main while a link
// without a fragment leaves the branch to the caller.
func ParseRepositoryLink(raw string) (repository, ref string) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", ""
	}
	if index := strings.Index(trimmed, "#"); index >= 0 {
		return strings.TrimSpace(trimmed[:index]), strings.TrimSpace(trimmed[index+1:])
	}
	return trimmed, ""
}

func parseGitLabProjectURL(raw string) (project, apiBase string, err error) {
	repository, err := parseRepositoryURL(raw)
	if err != nil {
		return "", "", err
	}
	if repository.provider != providerGitLab {
		return "", "", fmt.Errorf("CMS_TASK_URL is not a GitLab project URL")
	}
	return repository.project, repository.apiBase, nil
}

func escapeRepositoryPath(value string) string {
	parts := strings.Split(value, "/")
	for index := range parts {
		parts[index] = url.PathEscape(parts[index])
	}
	return strings.Join(parts, "/")
}

func repositoryRootLabel(labsPath string) string {
	if strings.TrimSpace(labsPath) == "" {
		return "repository root"
	}
	return labsPath
}

// cleanRepositoryPath normalizes labs_path. An empty result means the
// repository root, which is the default when a routing does not pin a
// subdirectory.
func cleanRepositoryPath(value string) (string, error) {
	trimmed := strings.TrimSpace(value)
	if strings.Contains(trimmed, "\\") {
		return "", fmt.Errorf("invalid labs_path")
	}
	cleaned := strings.Trim(path.Clean("/"+trimmed), "/")
	if cleaned == "." {
		return "", nil
	}
	return cleaned, nil
}
