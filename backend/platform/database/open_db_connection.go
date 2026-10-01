package database

import (
	"github.com/cms-lab-core/cms-labs-api/backend/pkg/configs"
	"github.com/cms-lab-core/cms-labs-api/shared/connection"
	"github.com/rs/zerolog"

	"gorm.io/gorm"

	"github.com/cms-lab-core/cms-labs-api/backend/app/queries"
	"github.com/cms-lab-core/cms-labs-api/backend/app/queries/lti_query"
)

// Queries struct for collect all app queries.
type Queries struct {
	*gorm.DB
	*lti_query.AuthProviderQueries
	*lti_query.LTINonceTokenQueries
	*lti_query.LTIAccessTokenQueries
	*lti_query.LTILaunchDataQueries
	*queries.UserQueries
	*queries.UserPasswordQueries
	*queries.ServerQueries
	*queries.RoundQueuePoolQueries
	*queries.LTIRoutingQueries
	*queries.LTIAttemptQueries
	*queries.LTIRoomQueries
	*queries.ServiceCardQueries
	*queries.TokenAttemptQueries
	*queries.RoleQueries
	*queries.TargetUserQueries
	*queries.TargetQueries
	*queries.TargetAddonQueries
	*queries.TargetRelationQueries
}

// OpenDBConnection func for opening database connection.
func OpenDBConnection(l *zerolog.Logger) (*Queries, error) {
	db, err := connection.MysqlConnection(configs.AppConfig.DB, l)

	if err != nil {
		return nil, err
	}

	return &Queries{
		// Set queries from models:
		DB:                    db,
		AuthProviderQueries:   &lti_query.AuthProviderQueries{DB: db, Logger: l},
		LTINonceTokenQueries:  &lti_query.LTINonceTokenQueries{DB: db, Logger: l},
		LTIAccessTokenQueries: &lti_query.LTIAccessTokenQueries{DB: db, Logger: l},
		LTILaunchDataQueries:  &lti_query.LTILaunchDataQueries{DB: db, Logger: l},
		UserQueries:           &queries.UserQueries{DB: db, Logger: l},
		UserPasswordQueries:   &queries.UserPasswordQueries{DB: db, Logger: l},
		ServerQueries:         &queries.ServerQueries{DB: db, Logger: l},
		RoundQueuePoolQueries: &queries.RoundQueuePoolQueries{DB: db, Logger: l},
		LTIRoutingQueries:     &queries.LTIRoutingQueries{DB: db, Logger: l},
		LTIAttemptQueries:     &queries.LTIAttemptQueries{DB: db, Logger: l},
		LTIRoomQueries:        &queries.LTIRoomQueries{DB: db, Logger: l},
		ServiceCardQueries:    &queries.ServiceCardQueries{DB: db, Logger: l},
		TokenAttemptQueries:   &queries.TokenAttemptQueries{DB: db, Logger: l},
		RoleQueries:           &queries.RoleQueries{DB: db, Logger: l},
		TargetUserQueries:     &queries.TargetUserQueries{DB: db, Logger: l},
		TargetQueries:         &queries.TargetQueries{DB: db, Logger: l},
		TargetAddonQueries:    &queries.TargetAddonQueries{DB: db, Logger: l},
		TargetRelationQueries: &queries.TargetRelationQueries{DB: db, Logger: l},
	}, nil
}

func (q *Queries) Close() error {
	dbInstance, err := q.DB.DB()
	if err != nil {
		return err
	}
	return dbInstance.Close()
}
