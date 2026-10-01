package usecases

import (
	"fmt"

	"github.com/cms-lab-core/cms-labs-api/pnetlabaddon/app/models"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/cms-lab-core/cms-labs-api/pnetlabaddon/app/queries"
	"github.com/cms-lab-core/cms-labs-api/shared/cms_client"
)

type LabCreateInputDTO struct {
	Extra   string `json:"extra"`
	UserPod int    `json:"user_pod"`
}

type LabCreateUC struct {
	UserQueries     *queries.UserQueries
	LabSessionQuery *queries.LabSessionQuery
	CMSClient       *cms_client.CMSClient
	*zerolog.Logger
}

func (u *LabCreateUC) CreateOrUpdateLab(
	userPod int,
	unlFilePathServer string,
	attemptId string,
) error {
	user, err := u.UserQueries.Get(userPod)
	if err != nil {
		return err
	}
	lab, _ := u.LabSessionQuery.GetByAttemptId(attemptId)
	if lab.LabSessionID != 0 {
		lab.AddJoinedUser(userPod)
		u.LabSessionQuery.DB.Save(&lab)
		user.LabSession = &lab.LabSessionID
		u.UserQueries.DB.Save(&user)
		log.Info().Msg(
			fmt.Sprintf(
				"UpdateLab by unlFilePathServer: %s, attemptId:%s and userPod: %d",
				unlFilePathServer,
				attemptId,
				userPod,
			),
		)
		return nil
	}
	lab = models.LabSession{
		LabSessionLID:    attemptId,
		LabSessionPod:    userPod,
		LabSessionJoined: fmt.Sprintf("%d", userPod),
		LabSessionPath:   unlFilePathServer,
	}
	u.LabSessionQuery.DB.Save(&lab)
	user.LabSession = &lab.LabSessionID
	u.UserQueries.DB.Save(&user)
	log.Info().Msg(
		fmt.Sprintf(
			"CreateLab by unlFilePathServer: %s, attemptId:%s and userPod: %d",
			unlFilePathServer,
			attemptId,
			userPod,
		),
	)
	return nil
}

func (u *LabCreateUC) Execute(dto LabCreateInputDTO) error {
	extra := cms_client.UnmarshalSSOTokenPublicExtraParams(dto.Extra)
	if extra.AttemptID == "" {
		u.Logger.Info().Msg("LabCreate attemptID is null")
		return nil
	}
	if extra.LabsType == cms_client.PNETLabsTypeDefault && extra.LabsPath != "" {
		u.Logger.Info().Msg(fmt.Sprintf("LabCreate by default path: %s", extra.LabsPath))
		return u.CreateOrUpdateLab(dto.UserPod, extra.LabsPath, extra.AttemptID)
	}
	return nil
}
