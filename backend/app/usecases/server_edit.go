package usecases

import (
	"github.com/cms-lab-core/cms-labs-api/backend/app/models"
	"github.com/cms-lab-core/cms-labs-api/backend/app/queries"
)

type ServerEditUC struct {
	ServerQueries *queries.ServerQueries
	RoleQueries   *queries.RoleQueries
}

type ServerEditInputDTO struct {
	ID                   uint   `json:"id"`
	ClientID             string `json:"client_id"`
	Type                 string `json:"type" validate:"required"`
	Name                 string `json:"name" validate:"required"`
	Url                  string `json:"url" validate:"required"`
	IsActive             bool   `json:"is_active"`
	MinutesForDisconnect int    `json:"minutes_for_disconnect"`
	MaxCountUsersLimit   int    `json:"max_count_users_limit"`
	UnitRate             int    `json:"unit_rate"`
	Roles                []uint `json:"roles"`
	Token                string `json:"token"`
}

type ServerEditRequest struct {
	JSONRPC string             `json:"jsonrpc" default:"2.0" required:"true"`
	Method  string             `json:"method" default:"server.upsert" required:"true"`
	Params  ServerEditInputDTO `json:"params,omitempty"`
	ID      string             `json:"id,omitempty" default:"1" required:"true"`
}

type ServerEditOutputDTO struct {
	ID uint `json:"id" required:"true"`
}

type ServerEditResponse struct {
	JSONRPC string              `json:"jsonrpc" default:"2.0" required:"true"`
	Result  ServerEditOutputDTO `json:"result,omitempty"`
	Error   interface{}         `json:"error,omitempty"`
	ID      string              `json:"id,omitempty" default:"1" required:"true"`
}

func (u *ServerEditUC) Execute(dto ServerEditInputDTO) (ServerEditOutputDTO, error) {
	entity := &models.Server{}
	entity.ID = dto.ID
	entity.ClientID = dto.ClientID
	entity.Type = dto.Type
	entity.Name = dto.Name
	entity.Url = dto.Url
	entity.IsActive = dto.IsActive
	entity.UnitRate = dto.UnitRate
	entity.MinutesForDisconnect = dto.MinutesForDisconnect
	entity.MaxCountUsersLimit = dto.MaxCountUsersLimit
	if dto.Token != "" {
		entity.Token = dto.Token
	}
	err := u.ServerQueries.Upsert(entity)
	if err != nil {
		return ServerEditOutputDTO{}, err
	}
	err = u.RoleQueries.SetByServerId(entity.ID, dto.Roles)
	return ServerEditOutputDTO{ID: entity.ID}, err
}
