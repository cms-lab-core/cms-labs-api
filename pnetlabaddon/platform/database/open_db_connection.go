package database

import (
	"github.com/cms-lab-core/cms-labs-api/pnetlabaddon/pkg/configs"
	"github.com/cms-lab-core/cms-labs-api/shared/connection"
	"github.com/rs/zerolog"

	"gorm.io/gorm"

	"github.com/cms-lab-core/cms-labs-api/pnetlabaddon/app/queries"
)

// Queries struct for collect all app queries.
type Queries struct {
	*gorm.DB
	*queries.UserQueries
	*queries.UserRoleQueries
	*queries.GuacamoleQueries
	*queries.LabSessionQuery
}

// OpenDBConnection func for opening database connection.
func OpenDBConnection(l *zerolog.Logger) (*Queries, error) {
	// Define Database connection variables.
	var (
		db  *gorm.DB
		err error
	)

	db, err = connection.MysqlConnection(configs.AppConfig.DB, l)

	if err != nil {
		return nil, err
	}

	return &Queries{
		// Set queries from models:
		DB:               db,
		UserQueries:      &queries.UserQueries{DB: db},
		UserRoleQueries:  &queries.UserRoleQueries{DB: db},
		GuacamoleQueries: &queries.GuacamoleQueries{DB: db, Logger: l},
		LabSessionQuery:  &queries.LabSessionQuery{DB: db},
	}, nil
}

func (q *Queries) Close() error {
	dbInstance, err := q.DB.DB()
	if err != nil {
		return err
	}
	return dbInstance.Close()
}
