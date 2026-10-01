package configs

import (
	"os"
	"strconv"
	"time"

	"github.com/maintainer64/cms-labs-api/shared/connection"
	"github.com/rs/zerolog/log"
)

type AppConfigModel struct {
	Debug bool
	// LabCatalog gates the laboratory catalog API. It is off unless
	// LAB_CATALOG_ENABLED=true, so a deployment without imported laboratories
	// serves an empty catalog instead of exposing the section.
	LabCatalog        bool
	Server            *connection.ServerConfig
	DB                *connection.DBConfig
	Vault             *connection.Vault
	JWT               *JWTConfig
	AddonsConfig      *connection.AddonsConfig
	ProxmoxSyncConfig *connection.ProxmoxSyncConfig
}

func (c *AppConfigModel) Reload() {
	c.Debug = os.Getenv("DEBUG") == "true"
	c.LabCatalog = os.Getenv("LAB_CATALOG_ENABLED") == "true"
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
	jwtAccess, err := NewJWTKeyConfig(
		os.Getenv("JWT_SECRET_KEY_PRIVATE"),
		os.Getenv("JWT_SECRET_KEY_PUBLIC"),
		time.Duration(getEnvInt("JWT_SECRET_KEY_EXPIRE_MINUTES_COUNT"))*time.Minute,
	)
	if err != nil {
		log.Warn().Err(err).Msg("JWT Access Key")
	}
	jwtRefresh, err := NewJWTKeyConfig(
		os.Getenv("JWT_REFRESH_KEY_PRIVATE"),
		os.Getenv("JWT_REFRESH_KEY_PUBLIC"),
		time.Duration(getEnvInt("JWT_REFRESH_KEY_EXPIRE_HOURS_COUNT"))*time.Hour,
	)
	if err != nil {
		log.Warn().Err(err).Msg("JWT Refresh Key")
	}
	c.JWT = &JWTConfig{
		AccessKey:  jwtAccess,
		RefreshKey: jwtRefresh,
	}
	c.Vault = &connection.Vault{
		VaultAddr:  os.Getenv("VAULT_ADDR"),
		VaultToken: os.Getenv("VAULT_TOKEN"),
	}
	c.AddonsConfig = connection.GetAddonsConfig(
		os.Getenv("ADDONS_PATH_CONFIG"),
	)
	c.ProxmoxSyncConfig = connection.GetProxmoxConfig(
		os.Getenv("PROXMOX_PATH_CONFIG"),
	)
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
