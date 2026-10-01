package addons

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/cms-lab-core/cms-labs-api/backend/app/addons/vault"
	"github.com/cms-lab-core/cms-labs-api/shared/connection"

	_ "github.com/lib/pq"
)

// PostgreSQLAddonService implements AddonService for PostgreSQL databases.
type PostgreSQLAddonService struct {
	Config      *connection.AddonConfig `json:"config"`
	URL         string                  `json:"url"`
	Host        string                  `json:"host"`
	Port        string                  `json:"port"`
	VaultClient vault.ClientInterface
}

// NewPostgreSQLAddonService creates a new PostgreSQL service instance.
func NewPostgreSQLAddonService(cfg *connection.AddonConfig, vaultClient vault.ClientInterface) AddonService {
	url, _ := cfg.Params["url"].(string)
	host, _ := cfg.Params["host"].(string)
	port, _ := cfg.Params["port"].(string)
	return &PostgreSQLAddonService{
		Config:      cfg,
		URL:         url,
		Host:        host,
		Port:        port,
		VaultClient: vaultClient,
	}
}

func quoteIdentifierPostgresql(id string) string {
	return `"` + strings.ReplaceAll(id, `"`, `""`) + `"`
}

// Create provisions a new PostgreSQL database and user, stores credentials in Vault.
func (s *PostgreSQLAddonService) Create(ctx context.Context, targetName string, name *string) (*AddonOperationConfig, error) {
	dbName := generateName(targetName, name)
	password := randomPassword(20)

	// Store credentials in Vault
	err := s.VaultClient.CreateServiceExtension(
		ctx,
		targetName,
		fmt.Sprintf("%s_%s", s.Config.Type, s.Config.Tag),
		map[string]any{
			"PG_HOST":     s.Config.Params["host"],
			"PG_PORT":     s.Config.Params["port"],
			"PG_USER":     dbName,
			"PG_DB":       dbName,
			"PG_PASSWORD": password,
		},
	)
	if err != nil {
		return nil, err
	}

	dbQuoted := quoteIdentifierPostgresql(dbName)
	// Execute three separate statements – we handle “already exists” errors gracefully.
	statements := []string{
		// 1. Создаем пользователя (если не существует)
		fmt.Sprintf(`CREATE USER %s WITH PASSWORD '%s';`, dbQuoted, password),
		// 2. На случай, если пользователь уже был, обновляем пароль
		fmt.Sprintf(`ALTER USER %s WITH PASSWORD '%s';`, dbQuoted, password),
		// 3. Создаем базу и СРАЗУ назначаем пользователя её владельцем
		fmt.Sprintf(`CREATE DATABASE %s OWNER %s;`, dbQuoted, dbQuoted),
		// 4. Дополнительно подтверждаем права (на случай существующих баз)
		fmt.Sprintf(`GRANT ALL PRIVILEGES ON DATABASE %s TO %s;`, dbQuoted, dbQuoted),
		// 5. Чтобы подстраховаться для уже существующих баз
		fmt.Sprintf(`ALTER DATABASE %s OWNER TO %s;`, dbQuoted, dbQuoted),
	}

	for _, stmt := range statements {
		if err := s.execRoot(ctx, stmt); err != nil {
			// Ignore “already exists” errors for CREATE DATABASE and CREATE USER
			if strings.Contains(err.Error(), "already exists") {
				if strings.Contains(stmt, "CREATE DATABASE") {
					// log? we can ignore
					continue
				}
				if strings.Contains(stmt, "CREATE USER") {
					// log? we can ignore
					continue
				}
			}
			return nil, err
		}
	}

	return &AddonOperationConfig{
		Name:          dbName,
		CurrentSizeMb: intPtr(0),
		MaxSizeMb:     intPtr(1024),
	}, nil
}

// Delete removes the database and its dedicated user, after terminating all active connections.
func (s *PostgreSQLAddonService) Delete(ctx context.Context, targetName string, cfg *AddonOperationConfig) (*AddonOperationConfig, error) {
	if cfg == nil || cfg.Name == "" {
		return nil, fmt.Errorf("database name is required for deletion")
	}

	dbName := cfg.Name
	dbQuoted := quoteIdentifierPostgresql(dbName)

	// Terminate existing connections to the database
	terminateSQL := fmt.Sprintf(
		`SELECT pg_terminate_backend(pid)
		 FROM pg_stat_activity
		 WHERE datname = '%s' AND pid <> pg_backend_pid();`,
		dbName,
	)
	_ = s.execRoot(ctx, terminateSQL) // best effort – ignore errors

	// Drop database and user
	statements := []string{
		fmt.Sprintf(`DROP DATABASE IF EXISTS %s;`, dbQuoted),
		fmt.Sprintf(`DROP USER IF EXISTS %s;`, dbQuoted),
	}

	for _, stmt := range statements {
		if err := s.execRoot(ctx, stmt); err != nil {
			return nil, err
		}
	}

	return &AddonOperationConfig{Name: dbName}, nil
}

// Reset rotates the password for the existing user, updates Vault.
func (s *PostgreSQLAddonService) Reset(ctx context.Context, targetName string, cfg *AddonOperationConfig) (*AddonOperationConfig, error) {
	if cfg == nil || cfg.Name == "" {
		return nil, fmt.Errorf("database name is required for reset")
	}

	dbName := cfg.Name
	newPassword := randomPassword(20)

	// Update Vault with the new password
	err := s.VaultClient.CreateServiceExtension(
		ctx,
		targetName,
		fmt.Sprintf("%s_%s", s.Config.Type, s.Config.Tag),
		map[string]any{
			"PG_HOST":     s.Config.Params["host"],
			"PG_PORT":     s.Config.Params["port"],
			"PG_USER":     dbName,
			"PG_DB":       dbName,
			"PG_PASSWORD": newPassword,
		},
	)
	if err != nil {
		return nil, err
	}

	dbQuoted := quoteIdentifierPostgresql(dbName)
	// Execute three separate statements – we handle “already exists” errors gracefully.
	statements := []string{
		// 1. Создаем пользователя (если не существует)
		fmt.Sprintf(`CREATE USER %s WITH PASSWORD '%s';`, dbQuoted, newPassword),
		// 2. На случай, если пользователь уже был, обновляем пароль
		fmt.Sprintf(`ALTER USER %s WITH PASSWORD '%s';`, dbQuoted, newPassword),
		// 3. Создаем базу и СРАЗУ назначаем пользователя её владельцем
		fmt.Sprintf(`CREATE DATABASE %s OWNER %s;`, dbQuoted, dbQuoted),
		// 4. Дополнительно подтверждаем права (на случай существующих баз)
		fmt.Sprintf(`GRANT ALL PRIVILEGES ON DATABASE %s TO %s;`, dbQuoted, dbQuoted),
		// 5. Чтобы подстраховаться для уже существующих баз
		fmt.Sprintf(`ALTER DATABASE %s OWNER TO %s;`, dbQuoted, dbQuoted),
	}

	for _, stmt := range statements {
		if err := s.execRoot(ctx, stmt); err != nil {
			// Ignore “already exists” errors for CREATE DATABASE and CREATE USER
			if strings.Contains(err.Error(), "already exists") {
				if strings.Contains(stmt, "CREATE DATABASE") {
					// log? we can ignore
					continue
				}
				if strings.Contains(stmt, "CREATE USER") {
					// log? we can ignore
					continue
				}
			}
			return nil, err
		}
	}

	// Preserve existing size metadata
	return &AddonOperationConfig{
		Name:          dbName,
		CurrentSizeMb: cfg.CurrentSizeMb,
		MaxSizeMb:     cfg.MaxSizeMb,
	}, nil
}

func (s *PostgreSQLAddonService) GetType() connection.AddonType {
	return s.Config.Type
}

// execRoot executes a single SQL statement using the root connection.
// (It does not split on semicolon – each statement is passed separately.)
func (s *PostgreSQLAddonService) execRoot(ctx context.Context, query string) error {
	db, err := sql.Open("postgres", s.URL)
	if err != nil {
		return err
	}
	defer db.Close()

	if _, err := db.ExecContext(ctx, query); err != nil {
		return err
	}
	return nil
}
