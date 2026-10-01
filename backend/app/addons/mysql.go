package addons

import (
	"context"
	"crypto/rand"
	"database/sql"
	"fmt"
	"regexp"
	"strings"

	"github.com/cms-lab-core/cms-labs-api/backend/app/addons/vault"
	"github.com/cms-lab-core/cms-labs-api/shared/connection"

	_ "github.com/go-sql-driver/mysql"
)

type MySQLAddonService struct {
	Config      *connection.AddonConfig `json:"config"`
	URL         string                  `json:"url"`
	Host        string                  `json:"host"`
	Port        string                  `json:"port"`
	VaultClient vault.ClientInterface
}

func NewMySQLAddonService(cfg *connection.AddonConfig, vaultClient vault.ClientInterface) AddonService {
	url, _ := cfg.Params["url"].(string)
	host, _ := cfg.Params["host"].(string)
	port, _ := cfg.Params["port"].(string)
	return &MySQLAddonService{
		Config:      cfg,
		URL:         url,
		Host:        host,
		Port:        port,
		VaultClient: vaultClient,
	}
}

func quoteIdentifierMysql(id string) string {
	return "`" + strings.ReplaceAll(id, "`", "``") + "`"
}

var nameRe = regexp.MustCompile(`[^a-zA-Z0-9_]`)

func generateName(targetName string, name *string) string {
	full := targetName
	if name != nil {
		full = targetName + "_" + *name
	}
	return "svc_" + strings.ToLower(nameRe.ReplaceAllString(full, "_"))
}

func (s *MySQLAddonService) Create(ctx context.Context, targetName string, name *string) (*AddonOperationConfig, error) {
	dbName := generateName(targetName, name)
	password := randomPassword(20)

	err := s.VaultClient.CreateServiceExtension(
		ctx,
		targetName,
		fmt.Sprintf("%s_%s", s.Config.Type, s.Config.Tag),
		map[string]any{
			"MYSQL_HOST":     s.Config.Params["host"],
			"MYSQL_PORT":     s.Config.Params["port"],
			"MYSQL_USER":     dbName,
			"MYSQL_DB":       dbName,
			"MYSQL_PASSWORD": password,
		},
	)
	if err != nil {
		return nil, err
	}
	dbNameQuoted := quoteIdentifierMysql(dbName)
	text := fmt.Sprintf(
		`CREATE DATABASE IF NOT EXISTS %s;
        CREATE USER IF NOT EXISTS %s@'%%' IDENTIFIED BY '%s';
        GRANT ALL PRIVILEGES ON %s.* TO %s@'%%';
    	ALTER USER %s@'%%' IDENTIFIED BY '%s';
    	FLUSH PRIVILEGES;`,
		dbNameQuoted, dbNameQuoted, password, dbNameQuoted, dbNameQuoted, dbNameQuoted, password,
	)

	if err := s.execRoot(ctx, text); err != nil {
		return nil, err
	}

	return &AddonOperationConfig{Name: dbName, CurrentSizeMb: intPtr(0), MaxSizeMb: intPtr(1024)}, nil
}

// Delete removes the database and the user. No Vault interaction – Vault secret is deleted separately.
func (s *MySQLAddonService) Delete(ctx context.Context, targetName string, cfg *AddonOperationConfig) (*AddonOperationConfig, error) {
	if cfg == nil || cfg.Name == "" {
		return nil, fmt.Errorf("database name is required for deletion")
	}

	dbNameQuoted := quoteIdentifierMysql(cfg.Name)
	text := fmt.Sprintf(
		`DROP DATABASE IF EXISTS %s;
        DROP USER IF EXISTS %s@'%%';
        FLUSH PRIVILEGES;`,
		dbNameQuoted, dbNameQuoted,
	)

	if err := s.execRoot(ctx, text); err != nil {
		return nil, err
	}

	// Return minimal config (only name) as in TypeScript version
	return &AddonOperationConfig{Name: cfg.Name}, nil
}

// Reset rotates the password, updates Vault, and changes the user's password.
func (s *MySQLAddonService) Reset(ctx context.Context, targetName string, cfg *AddonOperationConfig) (*AddonOperationConfig, error) {
	if cfg == nil || cfg.Name == "" {
		return nil, fmt.Errorf("database name is required for reset")
	}

	dbName := cfg.Name
	newPassword := randomPassword(20)

	// 1. Update Vault with the new credentials
	err := s.VaultClient.CreateServiceExtension(
		ctx,
		targetName,
		fmt.Sprintf("%s_%s", s.Config.Type, s.Config.Tag),
		map[string]any{
			"MYSQL_HOST":     s.Host,
			"MYSQL_PORT":     s.Port,
			"MYSQL_USER":     dbName,
			"MYSQL_DB":       dbName,
			"MYSQL_PASSWORD": newPassword,
		},
	)
	if err != nil {
		return nil, err
	}

	// 2. Change the user's password in MySQL
	dbNameQuoted := quoteIdentifierMysql(dbName)
	text := fmt.Sprintf(
		`CREATE DATABASE IF NOT EXISTS %s;
        CREATE USER IF NOT EXISTS %s@'%%' IDENTIFIED BY '%s';
        GRANT ALL PRIVILEGES ON %s.* TO %s@'%%';
    	ALTER USER %s@'%%' IDENTIFIED BY '%s';
    	FLUSH PRIVILEGES;`,
		dbNameQuoted, dbNameQuoted, newPassword, dbNameQuoted, dbNameQuoted, dbNameQuoted, newPassword,
	)

	if err := s.execRoot(ctx, text); err != nil {
		return nil, err
	}

	// Return the updated config (preserve existing size metadata)
	return &AddonOperationConfig{
		Name:          dbName,
		CurrentSizeMb: cfg.CurrentSizeMb,
		MaxSizeMb:     cfg.MaxSizeMb,
	}, nil
}

func (s *MySQLAddonService) execRoot(ctx context.Context, query string) error {
	db, err := sql.Open("mysql", s.URL)
	if err != nil {
		return err
	}
	defer db.Close()

	for _, stmt := range strings.Split(query, ";") {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			continue
		}
		if _, err := db.ExecContext(ctx, stmt); err != nil {
			return err
		}
	}
	return nil
}

func (s *MySQLAddonService) GetType() connection.AddonType {
	return s.Config.Type
}

func randomPassword(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	_, _ = rand.Read(b)
	for i := range b {
		b[i] = letters[b[i]%byte(len(letters))]
	}
	return string(b)
}

func intPtr(i int) *int { return &i }
