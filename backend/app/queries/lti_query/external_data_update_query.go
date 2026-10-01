package lti_query

import (
	"fmt"
	"strings"

	"github.com/cms-lab-core/cms-labs-api/shared/cms_client"

	"github.com/cms-lab-core/cms-labs-api/backend/app/models"
	"github.com/cms-lab-core/cms-labs-api/backend/app/queries"
	json "github.com/goccy/go-json"
	"github.com/ory/go-convenience/mapx"
	"github.com/ory/go-convenience/stringsx"
	"github.com/rs/zerolog"
)

type ExternalDataUpdateQuery struct {
	*queries.UserQueries
	*queries.RoleQueries
	*zerolog.Logger
}

func (q *ExternalDataUpdateQuery) UpdateByLaunchData(
	launchID string,
	launchData json.RawMessage,
) error {
	q.Logger.Info().Msg(fmt.Sprintf("UserQueries: update launch data by id: %+v", launchID))
	var jwtTokenPayload map[interface{}]interface{}
	if err := json.Unmarshal(launchData, &jwtTokenPayload); err != nil {
		return err
	}
	email := strings.ToLower(mapx.GetStringDefault(jwtTokenPayload, "email", ""))
	userFromDB, _ := q.UserQueries.GetByEmail(email)
	user := models.User{}
	user.ID = userFromDB.ID
	user.Email = email
	user.Name = mapx.GetStringDefault(jwtTokenPayload, "name", "")
	user.GroupName = stringsx.Coalesce(user.GroupName, userFromDB.GroupName)
	user.LTIUserID = mapx.GetStringDefault(jwtTokenPayload, "sub", "")
	user.DeletedAt = userFromDB.DeletedAt
	user.LastLaunchID = launchID
	err := q.UserQueries.Upsert(&user)
	if err != nil {
		return err
	}

	q.Logger.Info().Msg(fmt.Sprintf("UserQueries: update roles by user id: %+v", user.ID))
	userRoles, err := q.RoleQueries.GetRolesByUserId(user.ID)
	if err != nil {
		return err
	}
	rolesLTI := mapx.GetStringSliceDefault(jwtTokenPayload, "https://purl.imsglobal.org/spec/lti/claim/roles", []string{})
	mainRoleCode := MapRoleCoreByLTIRoleCode(rolesLTI)
	mainRole, err := q.RoleQueries.GetByCode(mainRoleCode)
	if err != nil {
		return err
	}
	setRoles := make([]uint, 0)
	// Delete student, instructor role from userRoles
	for _, userRole := range userRoles {
		if userRole.Code == cms_client.SSOUsersRoleStudent || userRole.Code == cms_client.SSOUsersRoleAdmin {
			continue
		}
		setRoles = append(setRoles, userRole.ID)
	}
	// Add mainRole
	setRoles = append(setRoles, mainRole.ID)
	err = q.RoleQueries.SetByUserId(user.ID, setRoles)
	return err
}

// MapRoleCoreByLTIRoleCode функция маппинга роли в
func MapRoleCoreByLTIRoleCode(rolesLTI []string) string {
	for _, roleLTI := range rolesLTI {
		switch roleLTI {
		case "http://purl.imsglobal.org/vocab/lis/v2/membership#Administrator":
			return cms_client.SSOUsersRoleAdmin
		case "Administrator":
			return cms_client.SSOUsersRoleAdmin
		case "http://purl.imsglobal.org/vocab/lis/v2/membership#ContentDeveloper":
			continue
		case "ContentDeveloper":
			continue
		case "http://purl.imsglobal.org/vocab/lis/v2/membership#Instructor":
			return cms_client.SSOUsersRoleInstructor
		case "Instructor":
			return cms_client.SSOUsersRoleInstructor
		case "http://purl.imsglobal.org/vocab/lis/v2/membership#Learner":
			continue
		case "Learner":
			continue
		case "http://purl.imsglobal.org/vocab/lis/v2/membership#Mentor":
			return cms_client.SSOUsersRoleInstructor
		case "Mentor":
			return cms_client.SSOUsersRoleInstructor
		default:
			continue
		}
	}
	return cms_client.SSOUsersRoleStudent
}
