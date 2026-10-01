package usecases

import (
	"fmt"

	"github.com/rs/zerolog/log"

	"github.com/cms-lab-core/cms-labs-api/pnetlabaddon/app/queries"
	"github.com/cms-lab-core/cms-labs-api/shared/cms_client"
	"github.com/rs/zerolog"
)

type PnetServerPingUC struct {
	LabSessionQuery *queries.LabSessionQuery
	CMSClient       *cms_client.CMSClient
	*zerolog.Logger
}

// Execute - функция отправки активных сессий лабораторных работ
func (u *PnetServerPingUC) Execute() error {
	labs, err := u.LabSessionQuery.GetRunningLabs()
	if err != nil {
		return fmt.Errorf("get running labs: %w", err)
	}

	// Преобразуем в словарь по номеру попытки
	labsByAttempt := make(map[string]queries.LabSessionRunningLabs, len(labs))
	for _, lab := range labs {
		labsByAttempt[lab.LabSessionLID] = lab
	}
	log.Info().Msgf("Found %d running lab PNET", len(labsByAttempt))

	allAttempts, err := u.CMSClient.ListAttempts(
		cms_client.ListAttemptsParams{
			Limit:    5000,
			Offset:   0,
			Statuses: []string{"pending", "active", "terminating"},
		},
	)
	if err != nil {
		return fmt.Errorf("list attempts: %w", err)
	}
	log.Info().Msgf("Found %d attempts CMS", len(allAttempts))

	var updateList []cms_client.UpdateAttemptParams

	for _, attempt := range allAttempts {
		cmsRequest := cms_client.UpdateAttemptParams{
			AttemptID: attempt.AttemptID,
			Status:    "active",
		}
		// Если статус completed - ничего не делаем
		if attempt.Status == "completed" {
			log.Info().Msgf(
				"Attempt %s id=%d completed, skipping",
				attempt.AttemptID,
				attempt.AttemptNumber,
			)
			continue
		}
		_, exists := labsByAttempt[attempt.AttemptID]
		// Нет сервера с таким номером попытки
		if !exists {
			// Если pending — запрос ещё мог не дойти
			if attempt.Status == "pending" {
				log.Info().Msgf(
					"No server found for attempt %s (id=%d), but pending",
					attempt.AttemptID,
					attempt.AttemptNumber,
				)
				continue
			}
			log.Info().Msgf(
				"No server found for attempt %s (id=%d)",
				attempt.AttemptID,
				attempt.AttemptNumber,
			)
			cmsRequest.Status = "completed"
			updateList = append(updateList, cmsRequest)
			continue
		}
		// Если terminating — останавливаем и удаляем сервер
		if attempt.Status == "terminating" {
			log.Info().Msgf(
				"Stopping server for terminating attempt %s (id=%d)",
				attempt.AttemptID,
				attempt.AttemptNumber,
			)
			// TODO: Здесь должно быть удаление сервера
			delete(labsByAttempt, attempt.AttemptID)
			updateList = append(updateList, cmsRequest)
			continue
		}
		// Для всех остальных статусов — отправляем активное состояние
		log.Info().Msgf("Sending active state for attempt %s (status=%s)", attempt.AttemptID, attempt.Status)
		updateList = append(updateList, cmsRequest)
	}

	if len(updateList) == 0 {
		log.Info().Msg("Nothing to update")
		return nil
	}

	log.Info().Msgf("Updating %d attempts", len(updateList))
	_, err = u.CMSClient.UpdateAttempts(updateList)
	if err != nil {
		return fmt.Errorf("update attempts: %w", err)
	}
	return nil
}
