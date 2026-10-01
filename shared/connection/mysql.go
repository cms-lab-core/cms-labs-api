package connection

import (
	"fmt"
	"time"

	"github.com/cms-lab-core/cms-labs-api/shared/jsonrpc"

	"github.com/rs/zerolog"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

type DBLogger struct {
	*zerolog.Logger
}

func (l *DBLogger) Printf(format string, ctx ...interface{}) {
	l.Logger.Info().Msg(fmt.Sprintf(format, ctx...))
}

// MysqlConnection func for connection to Mysql database.
func MysqlConnection(c *DBConfig, l *zerolog.Logger) (*gorm.DB, error) {

	// Build Mysql connection URL.
	mysqlConnURL := UrlBuilderMySql(c)
	l.Debug().Msg(fmt.Sprintf("Table prefix MysqlConnection %+v", c.TablePrefix))
	l.Debug().Msg(fmt.Sprintf("Mysql DSN %+v", mysqlConnURL))

	// Define database connection for Mysql.
	db, err := gorm.Open(
		mysql.Open(mysqlConnURL),
		&gorm.Config{
			QueryFields: true,
			Logger: logger.New(
				&DBLogger{Logger: l},
				logger.Config{
					SlowThreshold:             200 * time.Millisecond,
					LogLevel:                  logger.Warn,
					IgnoreRecordNotFoundError: false,
					Colorful:                  false,
				}),
			NamingStrategy: schema.NamingStrategy{
				TablePrefix: c.TablePrefix,
			},
		},
	)
	if err != nil {
		errText := fmt.Sprintf("error, not connected to database, %+v", err)
		l.Error().Msg(errText)
		return nil, jsonrpc.NewRpcError("db_not_connected", errText)
	}

	sqlDB, err := db.DB()
	if err != nil {
		errText := fmt.Sprintf("error, not opened connected to database, %+v", err)
		l.Error().Msg(errText)
		return nil, jsonrpc.NewRpcError("db_not_connected", errText)
	}

	sqlDB.SetMaxIdleConns(c.MaxIdleConnections)
	sqlDB.SetMaxOpenConns(c.MaxConnections)
	sqlDB.SetConnMaxLifetime(time.Duration(c.MaxLifetimeConnections) * time.Minute)
	return db, nil
}
