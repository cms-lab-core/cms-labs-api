package database

import (
	"errors"
	"fmt"
	"time"

	"github.com/cms-lab-core/cms-labs-api/shared/connection"
	"github.com/rs/zerolog"

	"gorm.io/gorm/schema"

	_ "github.com/cms-lab-core/cms-labs-api/shared/logs"

	"github.com/cms-lab-core/cms-labs-api/backend/pkg/configs"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	_ "github.com/go-sql-driver/mysql"
)

type DBLogger struct {
	*zerolog.Logger
}

func (l *DBLogger) Printf(format string, ctx ...interface{}) {
	l.Logger.Info().Msg(fmt.Sprintf(format, ctx...))
}

// MysqlConnection func for connection to Mysql database.
func MysqlConnection(l *zerolog.Logger) (*gorm.DB, error) {

	// Build Mysql connection URL.
	mysqlConnURL := connection.UrlBuilderMySql(configs.AppConfig.DB)

	l.Debug().Msg(fmt.Sprintf("Table prefix MysqlConnection %+v", configs.AppConfig.DB.TablePrefix))
	l.Debug().Msg(fmt.Sprintf("Mysql DSN %+v", mysqlConnURL))

	// Define database connection for Mysql.
	db, err := gorm.Open(
		mysql.Open(mysqlConnURL),
		&gorm.Config{
			QueryFields: true,
			Logger: logger.New(
				&DBLogger{
					Logger: l,
				},
				logger.Config{
					SlowThreshold:             200 * time.Millisecond,
					LogLevel:                  logger.Warn,
					IgnoreRecordNotFoundError: false,
					Colorful:                  false,
				}),
			NamingStrategy: schema.NamingStrategy{
				TablePrefix: configs.AppConfig.DB.TablePrefix,
			},
		},
	)
	if err != nil {
		errText := fmt.Sprintf("error, not connected to database, %+v", err)
		l.Error().Msg(errText)
		return nil, errors.New(errText)
	}

	sqlDB, err := db.DB()
	if err != nil {
		errText := fmt.Sprintf("error, not opened connected to database, %+v", err)
		l.Error().Msg(errText)
		return nil, errors.New(errText)
	}

	sqlDB.SetMaxIdleConns(configs.AppConfig.DB.MaxIdleConnections)
	sqlDB.SetMaxOpenConns(configs.AppConfig.DB.MaxConnections)
	sqlDB.SetConnMaxLifetime(time.Duration(configs.AppConfig.DB.MaxLifetimeConnections) * time.Minute)
	return db, nil
}
