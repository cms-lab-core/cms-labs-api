package server_queue

import (
	"github.com/cms-lab-core/cms-labs-api/backend/app/queries"
	"github.com/rs/zerolog"
)

type ServerQueueList struct {
	RoundQueuePoolQueries *queries.RoundQueuePoolQueries
	*zerolog.Logger
}

type ServerQueueListOutputDTO struct {
	Model []queries.RoundQueuePoolServerListItem `json:"model" validate:"required"`
}

type ServerQueueListRequest struct {
	JSONRPC string       `json:"jsonrpc" default:"2.0" required:"true"`
	Method  string       `json:"method" default:"server_queue.list" required:"true"`
	Params  *interface{} `json:"params,omitempty"`
	ID      string       `json:"id,omitempty" default:"1" required:"true"`
}

type ServerQueueListResponse struct {
	JSONRPC string                   `json:"jsonrpc" default:"2.0" required:"true"`
	Result  ServerQueueListOutputDTO `json:"result,omitempty"`
	Error   interface{}              `json:"error,omitempty"`
	ID      string                   `json:"id,omitempty" default:"1" required:"true"`
}

func (u *ServerQueueList) Execute() (ServerQueueListOutputDTO, error) {
	u.Logger.Info().Msg("ServerQueueList execute")
	entities, err := u.RoundQueuePoolQueries.List()
	return ServerQueueListOutputDTO{
		Model: entities,
	}, err
}
