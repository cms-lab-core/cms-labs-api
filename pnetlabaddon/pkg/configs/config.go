package configs

import (
	"os"
	"strconv"
	"time"

	"github.com/cms-lab-core/cms-labs-api/shared/connection"

	"github.com/cms-lab-core/cms-labs-api/shared/cms_client"
	"github.com/cms-lab-core/cms-labs-api/shared/guacamole_client"
)

type ServerConfig struct {
	Host              string
	Port              string
	ServerReadTimeout int
	Layer             string
}

type DBConfig struct {
	Type                   string
	User                   string
	Password               string
	Host                   string
	Port                   string
	Name                   string
	SSL                    string
	MaxConnections         int
	MaxIdleConnections     int
	MaxLifetimeConnections int
	TablePrefix            string
}

type AppConfigModel struct {
	Debug             bool
	Server            *connection.ServerConfig
	DB                *connection.DBConfig
	CMSClient         *cms_client.CMSClientConfig
	GuacamoleClient   *guacamole_client.GuacamoleClientConfig
	SchedulerInterval time.Duration
}

func (c *AppConfigModel) Reload() {
	c.Debug = os.Getenv("DEBUG") == "true"
	c.Server = &connection.ServerConfig{
		Host:              os.Getenv("SERVER_HOST"),
		Port:              os.Getenv("SERVER_PORT"),
		ServerReadTimeout: getEnvInt("SERVER_READ_TIMEOUT"),
		Layer:             os.Getenv("STAGE_STATUS"),
	}
	c.DB = &connection.DBConfig{
		Type:                   os.Getenv("DB_TYPE"),
		User:                   os.Getenv("DB_USER"),
		Password:               os.Getenv("DB_PASSWORD"),
		Host:                   os.Getenv("DB_HOST"),
		Port:                   os.Getenv("DB_PORT"),
		Name:                   os.Getenv("DB_NAME"),
		SSL:                    os.Getenv("DB_SSL_MODE"),
		MaxConnections:         getEnvInt("DB_MAX_CONNECTIONS"),
		MaxIdleConnections:     getEnvInt("DB_MAX_IDLE_CONNECTIONS"),
		MaxLifetimeConnections: getEnvInt("DB_MAX_LIFETIME_CONNECTIONS"),
		TablePrefix:            os.Getenv("DB_TABLE_PREFIX"),
	}
	c.CMSClient = &cms_client.CMSClientConfig{
		Debug:             c.Debug,
		MaxTimeoutSeconds: int64(getEnvInt("CMS_MAX_TIMEOUT")),
		ClientID:          os.Getenv("CMS_CLIENT_ID"),
		Token:             os.Getenv("CMS_TOKEN"),
		BaseUrl:           os.Getenv("CMS_BASE_URL"),
	}
	c.SchedulerInterval = time.Duration(int64(getEnvInt("CMS_PING_MINUTES"))) * time.Minute
	c.GuacamoleClient = &guacamole_client.GuacamoleClientConfig{
		Debug:             c.Debug,
		MaxTimeoutSeconds: int64(getEnvInt("GUACAMOLE_MAX_TIMEOUT")),
		BaseUrl:           os.Getenv("GUACAMOLE_BASE_URL"),
	}
}

func getEnvInt(name string) int {
	digit, _ := strconv.Atoi(os.Getenv(name))
	return digit
}

func NewAppConfigModel() *AppConfigModel {
	model := &AppConfigModel{}
	model.Reload()
	return model
}

var AppConfig = NewAppConfigModel()
