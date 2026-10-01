package usecases

import (
	"github.com/cms-lab-core/cms-labs-api/backend/app/models"
	"github.com/cms-lab-core/cms-labs-api/backend/app/queries"
)

type UserGetUC struct {
	UserQueries *queries.UserQueries
	RoleQueries *queries.RoleQueries
}

type UserGetInputDTO struct {
	ID uint `json:"id" required:"true"`
}

type UserGetOutputDTO struct {
	Model models.User `json:"model" required:"true"`
	Roles []uint      `json:"roles"`
}

type UserGetRequest struct {
	JSONRPC string          `json:"jsonrpc" default:"2.0" required:"true"`
	Method  string          `json:"method" default:"user.get" required:"true"`
	Params  UserGetInputDTO `json:"params,omitempty"`
	ID      string          `json:"id,omitempty" default:"1" required:"true"`
}

type UserGetResponse struct {
	JSONRPC string           `json:"jsonrpc" default:"2.0" required:"true"`
	Result  UserGetOutputDTO `json:"result,omitempty"`
	Error   interface{}      `json:"error,omitempty"`
	ID      string           `json:"id,omitempty" default:"1" required:"true"`
}

func (u *UserGetUC) Execute(dto UserGetInputDTO) (UserGetOutputDTO, error) {
	form, err := u.UserQueries.Get(dto.ID)
	if err != nil && err.Error() == queries.UserNotActive.Error() {
		err = nil
	}
	if err != nil {
		return UserGetOutputDTO{}, err
	}
	result := UserGetOutputDTO{
		Model: form,
	}
	roles, err := u.RoleQueries.GetByRelationUsersIds([]uint{dto.ID})
	if err != nil {
		return result, err
	}
	if val, ok := roles[dto.ID]; ok {
		result.Roles = val
	}
	return result, err
}
