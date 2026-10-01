package auth

import (
	"errors"

	"github.com/cms-lab-core/cms-labs-api/shared/cms_client"

	"github.com/cms-lab-core/cms-labs-api/backend/app/queries"
	"golang.org/x/crypto/bcrypt"
)

type UserPasswordRecoverUC struct {
	UserPasswordQueries *queries.UserPasswordQueries
	User                *cms_client.SSOTokenPublicData
}

type UserPasswordChangeInputDTO struct {
	OldPassword   string `json:"old_password"`
	NewPassword   string `json:"new_password"`
	AgainPassword string `json:"again_password"`
}

type UserPasswordChangeOutputDTO struct {
	Id uint `json:"id"`
}

type UserPasswordChangeRequest struct {
	JSONRPC string                     `json:"jsonrpc" default:"2.0" required:"true"`
	Method  string                     `json:"method" default:"user.password_change" required:"true"`
	Params  UserPasswordChangeInputDTO `json:"params,omitempty"`
	ID      string                     `json:"id,omitempty" default:"1" required:"true"`
}

type UserPasswordChangeResponse struct {
	JSONRPC string                      `json:"jsonrpc" default:"2.0" required:"true"`
	Result  UserPasswordChangeOutputDTO `json:"result,omitempty"`
	Error   interface{}                 `json:"error,omitempty"`
	ID      string                      `json:"id,omitempty" default:"1" required:"true"`
}

func (u *UserPasswordRecoverUC) SetContext(user *cms_client.SSOTokenPublicData) *UserPasswordRecoverUC {
	u.User = user
	return u
}

func (u *UserPasswordRecoverUC) Execute(dto UserPasswordChangeInputDTO) (UserPasswordChangeOutputDTO, error) {
	wrongPassword := errors.New("wrong password")
	outputDTO := UserPasswordChangeOutputDTO{
		Id: u.User.UserID(),
	}
	if dto.NewPassword != dto.AgainPassword {
		return outputDTO, wrongPassword
	}
	creds, err := u.UserPasswordQueries.Get(u.User.UserID())
	if err != nil {
		return outputDTO, wrongPassword
	}
	err = bcrypt.CompareHashAndPassword([]byte(creds.HashPassword), []byte(dto.OldPassword))
	if err != nil {
		return outputDTO, wrongPassword
	}
	newHashPassword, err := bcrypt.GenerateFromPassword(
		[]byte(dto.NewPassword),
		14,
	)
	if err != nil {
		return outputDTO, wrongPassword
	}
	creds.HashPassword = string(newHashPassword)
	_ = u.UserPasswordQueries.Upsert(&creds)
	outputDTO.Id = u.User.UserID()
	return outputDTO, err
}
