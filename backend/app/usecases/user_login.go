package usecases

import (
	"fmt"

	"github.com/cms-lab-core/cms-labs-api/backend/app/models"
	"github.com/cms-lab-core/cms-labs-api/backend/app/queries"
	"github.com/cms-lab-core/cms-labs-api/backend/app/queries/lti_query"
	"github.com/cms-lab-core/cms-labs-api/backend/app/usecases/auth"
	"github.com/cms-lab-core/cms-labs-api/shared/cms_client"
	"github.com/cms-lab-core/cms-labs-api/shared/ldap_client"
	"github.com/rs/zerolog"
)

type UserLoginUC struct {
	TokenManager        *auth.TokenManager
	AuthProviderQueries *lti_query.AuthProviderQueries
	UserQueries         *queries.UserQueries
	*zerolog.Logger
	IssId string
}

func (u *UserLoginUC) SetContext(issId string) *UserLoginUC {
	u.IssId = issId
	return u
}

func (u *UserLoginUC) Execute(dto auth.RenewManagerCredentialsInputDTO) (*cms_client.SSOToken, error) {
	if dto.ProviderId == 0 {
		return u.internalAuth(dto)
	}
	provider, err := u.AuthProviderQueries.Get(dto.ProviderId)
	if err != nil {
		return nil, err
	}
	if provider.Type == models.AuthProviderTypeLDAP {
		return u.ldapAuth(dto, &provider)
	}
	return nil, err
}

func (u *UserLoginUC) ldapAuth(dto auth.RenewManagerCredentialsInputDTO, provider *models.AuthProvider) (
	*cms_client.SSOToken,
	error,
) {
	u.Logger.Info().Msg(
		fmt.Sprintf(
			"UserLoginUC: ldap auth with email: %s",
			dto.Email,
		),
	)
	client := &ldap_client.LDAPClient{
		URL: provider.BaseURI,
		DN:  provider.KeySetURI,
	}
	user, err := client.GetInfo(dto.Email, dto.Password)
	if err != nil {
		u.Logger.Info().Msg(
			fmt.Sprintf(
				"UserLoginUC: ldap user is not found by email: %s, error: %+v",
				dto.Email,
				err,
			),
		)
		return nil, queries.IncorrectPassword
	}
	if len(user.Mails) == 0 {
		u.Logger.Info().Msg(
			fmt.Sprintf(
				"UserLoginUC: ldap user has not login: %s",
				dto.Email,
			),
		)
		return nil, queries.IncorrectLogin
	}
	var userFromDB models.User
	for _, mail := range user.Mails {
		userFromDB, err = u.UserQueries.GetByEmail(mail)
		if err == nil {
			break
		}
	}
	userFromDB.Name = user.FullName
	userFromDB.GroupName = user.Group
	if userFromDB.Email == "" {
		userFromDB.Email = user.Mails[0]
	}
	err = u.UserQueries.Upsert(&userFromDB)
	if err != nil {
		u.Logger.Info().Msg(
			fmt.Sprintf(
				"UserLoginUC: ldap update user error by login: %s. Error: %+v",
				dto.Email,
				err,
			),
		)
		return nil, queries.IncorrectLogin
	}
	token, err := u.TokenManager.NewJWTByUserId(u.IssId, userFromDB.ID, nil, nil)
	if err != nil {
		return nil, err
	}
	return token, nil
}

func (u *UserLoginUC) internalAuth(dto auth.RenewManagerCredentialsInputDTO) (*cms_client.SSOToken, error) {
	u.Logger.Info().Msg(
		fmt.Sprintf(
			"UserLoginUC: internal auth with email: %s",
			dto.Email,
		),
	)
	token, err := u.TokenManager.NewJWTByCredentials(u.IssId, dto.Email, dto.Password)
	if err != nil {
		return nil, err
	}
	return token, nil
}
