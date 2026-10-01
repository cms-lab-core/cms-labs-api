package server_queue

import (
	"github.com/cms-lab-core/cms-labs-api/backend/app/models"
	"github.com/cms-lab-core/cms-labs-api/backend/app/queries"
	"github.com/rs/zerolog"
	funk "github.com/thoas/go-funk"
)

type ServerQueueUpsert struct {
	ServerQueries         *queries.ServerQueries
	RoundQueuePoolQueries *queries.RoundQueuePoolQueries
	*zerolog.Logger
}

type ServerQueueUpsertRequest struct {
	JSONRPC string       `json:"jsonrpc" default:"2.0" required:"true"`
	Method  string       `json:"method" default:"server_queue.upsert" required:"true"`
	Params  *interface{} `json:"params,omitempty"`
	ID      string       `json:"id,omitempty" default:"1" required:"true"`
}

type ServerQueueUpsertOutputDTO struct {
}

type ServerQueueUpsertResponse struct {
	JSONRPC string                     `json:"jsonrpc" default:"2.0" required:"true"`
	Result  ServerQueueUpsertOutputDTO `json:"result,omitempty"`
	Error   interface{}                `json:"error,omitempty"`
	ID      string                     `json:"id,omitempty" default:"1" required:"true"`
}

func (u *ServerQueueUpsert) serverEntityToStats(item models.ServerListItem) ServerStats {
	return ServerStats{
		ID:             item.ID,
		Name:           item.Name,
		UnitRate:       int(item.UnitRate),
		LastCountUsers: int(item.LastCountUsers),
	}
}

func (u *ServerQueueUpsert) Execute() error {
	u.Logger.Info().Msg("ServerQueueUpsert start to generate new distribution servers")
	entities, _, err := u.ServerQueries.List(queries.ServerQueriesListDTO{
		Limit:  queries.MaxLimitCount,
		Types:  []string{models.ServerTypePnet},
		Offset: 0,
	})
	if err != nil {
		return err
	}
	if len(entities) == 0 {
		return nil
	}
	stats := funk.Map(entities, u.serverEntityToStats).([]ServerStats)
	distribute := ByPriority(stats).GenerateSequencePriorityDistribute(u.Logger)
	return u.RoundQueuePoolQueries.UpsertQueueByServerIds(distribute)
}
