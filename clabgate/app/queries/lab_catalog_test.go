package queries

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	resty "github.com/go-resty/resty/v2"
)

func TestLabCatalogBundleUsesGitLabAPIAndStableOrder(t *testing.T) {
	const revision = "0123456789abcdef"
	client := resty.New().SetTransport(roundTripFunc(func(r *http.Request) (*http.Response, error) {
		if r.Header.Get("PRIVATE-TOKEN") != "private-token" {
			t.Fatalf("GitLab token was not forwarded")
		}
		status := http.StatusOK
		var body []byte
		switch {
		case strings.Contains(r.URL.Path, "/repository/commits/main"):
			body, _ = json.Marshal(map[string]string{"id": revision})
		case strings.HasSuffix(r.URL.Path, "/repository/tree"):
			if r.URL.Query().Get("path") != "course/lab 1" || r.URL.Query().Get("ref") != revision || r.URL.Query().Get("recursive") != "true" {
				t.Fatalf("unexpected tree query: %s", r.URL.RawQuery)
			}
			body, _ = json.Marshal([]map[string]string{
				{"path": "course/lab 1/z.yml", "type": "blob"},
				{"path": "course/lab 1/readme.md", "type": "blob"},
				{"path": "course/lab 1/a.yaml", "type": "blob"},
			})
		case strings.Contains(r.URL.Path, "/repository/files/"):
			file, _ := url.PathUnescape(strings.TrimSuffix(strings.TrimPrefix(r.URL.Path[strings.Index(r.URL.Path, "/repository/files/"):], "/repository/files/"), "/raw"))
			body = []byte("kind: ConfigMap\nmetadata:\n  name: " + strings.TrimSuffix(file[strings.LastIndex(file, "/")+1:], ".yaml"))
		default:
			status = http.StatusNotFound
		}
		header := make(http.Header)
		header.Set("Content-Type", "application/json")
		return &http.Response{StatusCode: status, Header: header, Body: io.NopCloser(strings.NewReader(string(body))), Request: r}, nil
	}))

	catalog := NewLabCatalog("https://gitlab.test/group/task-project", "main", "private-token", client)
	bundle, err := catalog.Bundle(context.Background(), "/course/lab 1/")
	if err != nil {
		t.Fatal(err)
	}
	if bundle.Revision != revision {
		t.Fatalf("revision = %q", bundle.Revision)
	}
	if len(bundle.Files) != 2 || bundle.Files[0] != "course/lab 1/a.yaml" || bundle.Files[1] != "course/lab 1/z.yml" {
		t.Fatalf("unexpected files: %#v", bundle.Files)
	}
	if strings.Index(bundle.Manifest, "name: a") > strings.Index(bundle.Manifest, "name: z.yml") {
		t.Fatalf("manifests are not stable-sorted: %s", bundle.Manifest)
	}
}

func TestLabCatalogBundleUsesGitHubAPIAndStableOrder(t *testing.T) {
	const revision = "fedcba9876543210"
	client := resty.New().SetTransport(roundTripFunc(func(r *http.Request) (*http.Response, error) {
		if r.Header.Get("Authorization") != "Bearer github-token" {
			t.Fatalf("GitHub token was not forwarded as a bearer token")
		}
		if r.Header.Get("X-GitHub-Api-Version") == "" {
			t.Fatal("GitHub API version header is missing")
		}
		status := http.StatusOK
		contentType := "application/json"
		var body []byte
		switch {
		case strings.HasSuffix(r.URL.Path, "/commits/main"):
			body, _ = json.Marshal(map[string]string{"sha": revision})
		case strings.Contains(r.URL.Path, "/git/trees/"):
			if r.URL.Query().Get("recursive") != "1" {
				t.Fatalf("unexpected tree query: %s", r.URL.RawQuery)
			}
			body, _ = json.Marshal(map[string]any{"tree": []map[string]string{
				{"path": "course/lab-1/z.yml", "type": "blob"},
				{"path": "course/other/ignored.yaml", "type": "blob"},
				{"path": "course/lab-1/a.yaml", "type": "blob"},
			}})
		case strings.Contains(r.URL.Path, "/contents/course/lab-1/"):
			if r.Header.Get("Accept") != "application/vnd.github.raw+json" || r.URL.Query().Get("ref") != revision {
				t.Fatalf("unexpected contents request headers or query")
			}
			contentType = "application/yaml"
			body = []byte("kind: ConfigMap\nmetadata:\n  name: " + pathBaseWithoutYAML(r.URL.Path))
		default:
			status = http.StatusNotFound
		}
		header := make(http.Header)
		header.Set("Content-Type", contentType)
		return &http.Response{StatusCode: status, Header: header, Body: io.NopCloser(strings.NewReader(string(body))), Request: r}, nil
	}))

	catalog := NewLabCatalog("https://github.com/maintainer64/lab-tasks.git", "main", "github-token", client)
	bundle, err := catalog.Bundle(context.Background(), "course/lab-1")
	if err != nil {
		t.Fatal(err)
	}
	if bundle.Revision != revision {
		t.Fatalf("revision = %q", bundle.Revision)
	}
	if len(bundle.Files) != 2 || bundle.Files[0] != "course/lab-1/a.yaml" || bundle.Files[1] != "course/lab-1/z.yml" {
		t.Fatalf("unexpected files: %#v", bundle.Files)
	}
	if strings.Index(bundle.Manifest, "name: a") > strings.Index(bundle.Manifest, "name: z") {
		t.Fatalf("manifests are not stable-sorted: %s", bundle.Manifest)
	}
}

func TestLabCatalogBundleUsesGitForPublicGitHubAndPinsRevision(t *testing.T) {
	repositoryDir := t.TempDir()
	runTestGit(t, repositoryDir, "init", "--initial-branch=main")
	runTestGit(t, repositoryDir, "config", "user.name", "CMS Labs test")
	runTestGit(t, repositoryDir, "config", "user.email", "cms-labs@example.com")

	labDirectory := filepath.Join(repositoryDir, "task", "nested")
	if err := os.MkdirAll(labDirectory, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(repositoryDir, "task", "z.yml"), []byte("kind: ConfigMap\nmetadata:\n  name: z\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(labDirectory, "a.yaml"), []byte("kind: ConfigMap\nmetadata:\n  name: a\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(repositoryDir, "task", "README.md"), []byte("ignored"), 0o600); err != nil {
		t.Fatal(err)
	}
	runTestGit(t, repositoryDir, "add", ".")
	runTestGit(t, repositoryDir, "commit", "-m", "initial task")
	firstRevision := strings.TrimSpace(runTestGit(t, repositoryDir, "rev-parse", "HEAD"))

	catalog := NewLabCatalog("https://github.com/maintainer64/cms-labs-simple-task.git", "main", "", resty.New())
	repository := repositoryLocation{provider: providerGitHub, cloneURL: repositoryDir}
	bundle, err := catalog.bundleFromGit(context.Background(), repository, "task", "")
	if err != nil {
		t.Fatal(err)
	}
	if bundle.Revision != firstRevision {
		t.Fatalf("revision = %q, want %q", bundle.Revision, firstRevision)
	}
	if len(bundle.Files) != 2 || bundle.Files[0] != "task/nested/a.yaml" || bundle.Files[1] != "task/z.yml" {
		t.Fatalf("unexpected files: %#v", bundle.Files)
	}
	if strings.Index(bundle.Manifest, "name: a") > strings.Index(bundle.Manifest, "name: z") {
		t.Fatalf("manifests are not stable-sorted: %s", bundle.Manifest)
	}

	if err := os.WriteFile(filepath.Join(repositoryDir, "task", "z.yml"), []byte("kind: ConfigMap\nmetadata:\n  name: changed\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	runTestGit(t, repositoryDir, "add", ".")
	runTestGit(t, repositoryDir, "commit", "-m", "change task")
	pinned, err := catalog.bundleFromGit(context.Background(), repository, "task", firstRevision)
	if err != nil {
		t.Fatal(err)
	}
	if pinned.Revision != firstRevision || strings.Contains(pinned.Manifest, "name: changed") {
		t.Fatalf("pinned bundle moved from %s: %#v", firstRevision, pinned)
	}
}

func runTestGit(t *testing.T, directory string, args ...string) string {
	t.Helper()
	command := exec.Command("git", args...)
	command.Dir = directory
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("git %s: %v: %s", strings.Join(args, " "), err, output)
	}
	return string(output)
}

func pathBaseWithoutYAML(value string) string {
	base := value[strings.LastIndex(value, "/")+1:]
	return strings.TrimSuffix(strings.TrimSuffix(base, ".yaml"), ".yml")
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (function roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return function(request)
}

func TestLabCatalogRejectsInvalidInput(t *testing.T) {
	if _, _, err := parseGitLabProjectURL("gitlab.com/project"); err == nil {
		t.Fatal("expected full project URL validation error")
	}
	if _, err := cleanRepositoryPath(""); err == nil {
		t.Fatal("expected empty labs_path validation error")
	}
	if _, err := cleanRepositoryPath(`course\\lab`); err == nil {
		t.Fatal("expected backslash validation error")
	}
	if _, err := parseRepositoryURL("https://github.com/owner/group/project"); err == nil {
		t.Fatal("expected GitHub owner/repository validation error")
	}
}
