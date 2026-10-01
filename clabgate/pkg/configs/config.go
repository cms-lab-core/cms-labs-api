package configs

import (
	"os"
	"strconv"

	"github.com/cms-lab-core/cms-labs-api/shared/connection"
)

type AppConfigModel struct {
	Debug   bool
	Server  *connection.ServerConfig
	CMS     *CMSConfig
	Session *SessionConfig
}

type CMSConfig struct {
	BaseURL           string
	ClientID          string
	Token             string
	MaxTimeoutSeconds int64
}

type SessionConfig struct {
	TaskRepositoryURL     string
	TaskBranch            string
	TaskRepositoryToken   string
	JupyterImage          string
	JupyterStorage        string
	WorkspacePrefix       string
	WorkspaceSecret       string
	WorkspaceGrantTTL     int64
	WorkspaceCookieTTL    int64
	WorkspaceCookieSecure bool
	ReconcileSeconds      int64
	ReconcileMaxSeconds   int64
	CheckerImage          string
	CheckerTimeout        int64
}

func (c *AppConfigModel) Reload() {
	c.Debug = os.Getenv("DEBUG") == "true"
	c.Server = &connection.ServerConfig{
		Host:              os.Getenv("SERVER_HOST"),
		Port:              os.Getenv("SERVER_PORT"),
		ServerReadTimeout: getEnvInt("SERVER_READ_TIMEOUT"),
		Layer:             os.Getenv("STAGE_STATUS"),
	}
	c.CMS = &CMSConfig{
		BaseURL:           os.Getenv("CMS_URL"),
		ClientID:          os.Getenv("CMS_LOGIN"),
		Token:             os.Getenv("CMS_PASSWORD"),
		MaxTimeoutSeconds: getEnvInt64Default("CMS_TIMEOUT_SECONDS", 30),
	}
	c.Session = &SessionConfig{
		TaskRepositoryURL:     os.Getenv("CMS_TASK_URL"),
		TaskBranch:            getEnvDefault("CMS_TASK_BRANCH", "main"),
		TaskRepositoryToken:   getEnvDefault("TASK_REPOSITORY_TOKEN", os.Getenv("GITLAB_TOKEN")),
		JupyterImage:          os.Getenv("JUPYTER_IMAGE"),
		JupyterStorage:        getEnvDefault("JUPYTER_STORAGE_SIZE", "1Gi"),
		WorkspacePrefix:       getEnvDefault("WORKSPACE_PROXY_PREFIX", "/clabgate/workspace"),
		WorkspaceSecret:       os.Getenv("WORKSPACE_AUTH_SECRET"),
		WorkspaceGrantTTL:     getEnvInt64Default("WORKSPACE_GRANT_TTL_SECONDS", 60),
		WorkspaceCookieTTL:    getEnvInt64Default("WORKSPACE_COOKIE_TTL_SECONDS", 3600),
		WorkspaceCookieSecure: getEnvBoolDefault("WORKSPACE_COOKIE_SECURE", true),
		ReconcileSeconds:      getEnvInt64Default("SESSION_RECONCILE_INTERVAL_SECONDS", 30),
		ReconcileMaxSeconds:   getEnvInt64Default("SESSION_RECONCILE_MAX_INTERVAL_SECONDS", 120),
		CheckerImage:          os.Getenv("CHECKER_IMAGE"),
		CheckerTimeout:        getEnvInt64Default("CHECKER_TIMEOUT_SECONDS", 600),
	}
}

func getEnvInt(name string) int {
	digit, _ := strconv.Atoi(os.Getenv(name))
	return digit
}

func getEnvInt64Default(name string, fallback int64) int64 {
	digit, err := strconv.ParseInt(os.Getenv(name), 10, 64)
	if err != nil || digit <= 0 {
		return fallback
	}
	return digit
}

func getEnvDefault(name, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
}

func getEnvBoolDefault(name string, fallback bool) bool {
	value := os.Getenv(name)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func NewAppConfigModel() *AppConfigModel {
	model := &AppConfigModel{}
	model.Reload()
	return model
}

var AppConfig = NewAppConfigModel()
