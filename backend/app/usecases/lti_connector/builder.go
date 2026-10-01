package lti_connector

import (
	"github.com/cms-lab-core/cms-labs-api/backend/app/queries"
	datastore "github.com/cms-lab-core/cms-labs-api/backend/app/queries/lti_query"
	"github.com/cms-lab-core/cms-labs-api/backend/app/usecases/lti_connector/connector"
	"github.com/rs/zerolog"
)

// LTIConnectorAPI упрощает взаимодействие с бекендом LTI протокола
type LTIConnectorAPI struct {
	AuthProviderQueries        *datastore.AuthProviderQueries
	LTILaunchDataQueries       *datastore.LTILaunchDataQueries
	UserQueries                *queries.UserQueries
	LTIProtocolDatastoreConfig *datastore.Config
	*zerolog.Logger
}

func (c *LTIConnectorAPI) ConnectorByAttemptID(attemptId string) (*connector.Connector, error) {
	launchEntity, err := c.LTILaunchDataQueries.GetByAttemptId(attemptId)
	if err != nil {
		return nil, err
	}
	form, err := c.AuthProviderQueries.Get(launchEntity.LTIFormID)
	if err != nil {
		return nil, err
	}
	conn, err := connector.New(
		c.LTIProtocolDatastoreConfig,
		launchEntity.ID,
		form.LTIClientID,
		c.Logger,
	)
	if err != nil {
		return nil, err
	}
	err = conn.SetSigningKey(form.PrivateKey)
	if err != nil {
		return nil, err
	}
	return conn, nil
}
