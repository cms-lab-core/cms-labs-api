package usecases

import (
	"context"
	"errors"
	"fmt"

	"github.com/cms-lab-core/cms-labs-api/backend/app/addons"
	"github.com/cms-lab-core/cms-labs-api/backend/app/addons/vault"
	"github.com/cms-lab-core/cms-labs-api/backend/app/models"
	"github.com/cms-lab-core/cms-labs-api/backend/app/queries"
	"github.com/cms-lab-core/cms-labs-api/shared/connection"
	"github.com/cms-lab-core/cms-labs-api/shared/k8s_utils"
	"github.com/goccy/go-json"
	"github.com/rs/zerolog"
)

// TargetUserRotateAddonsUC – перевыпуск и обновление всех прав для пользователя
type TargetUserRotateAddonsUC struct {
	TargetUserQueries   *queries.TargetUserQueries
	TargetQueries       *queries.TargetQueries
	UserQueries         *queries.UserQueries
	VaultClient         vault.ClientInterface
	TargetAddonQueries  *queries.TargetAddonQueries
	FactoryAddonService *addons.FactoryAddonService
	Logger              *zerolog.Logger
}

type HarborMembers struct {
	Username string
	TargetID string
	Access   bool
}

// Execute – Перевыпускает и меняет правила работы с пользователем
func (uc *TargetUserRotateAddonsUC) Execute(userID uint, currentTargetID string) error {
	uc.Logger.Error().Msg(fmt.Sprintf(
		"TargetUserRotateAddonsUC rotate permissions on by users %d",
		userID,
	))
	user, err := uc.UserQueries.Get(userID)
	if err != nil && !errors.Is(err, queries.UserNotActive) && !errors.Is(err, queries.UserNotFoundError) {
		uc.Logger.Info().Msg(fmt.Sprintf(
			"TargetUserRotateAddonsUC get by userId %d %+v", userID, err,
		))
		return err
	}
	if errors.Is(err, queries.UserNotFoundError) {
		uc.Logger.Info().Msg(fmt.Sprintf(
			"TargetUserRotateAddonsUC get by userId %d user has not found",
			userID,
		))
		return nil
	}
	// Для некоторых сервисов нужно bulk роли по одной сущности
	targetsIds := make([]string, 0)
	if user.IsActive() {
		targetsIds, err = uc.TargetUserQueries.GetByUserId(userID)
	} else {
		uc.Logger.Info().Msg(fmt.Sprintf(
			"TargetUserRotateAddonsUC user %d email %s is not active",
			userID,
			user.Email,
		))
	}
	if err != nil {
		uc.Logger.Info().Msg(fmt.Sprintf(
			"TargetUserRotateAddonsUC error get targets by userId %d %+v", userID, err,
		))
		return err
	}
	uc.Logger.Info().Msg(fmt.Sprintf(
		"TargetUserRotateAddonsUC get targets %d count by userId %d",
		len(targetsIds),
		userID,
	))
	// Обязательно чтобы у пользователя обновилась текущая
	targetsIds = append(targetsIds, currentTargetID)
	vaultBind := vault.UsersAndServices{
		UserEmail: user.Email,
		Services:  make([]vault.ServicesBindUser, 0),
	}
	harborBind := make([]HarborMembers, 0)
	ctx := context.Background()
	for _, targetID := range targetsIds {
		uc.Logger.Info().Msg(fmt.Sprintf(
			"TargetUserRotateAddonsUC get target %s and userId %d",
			targetID,
			userID,
		))
		target, err := uc.TargetQueries.Get(targetID)
		if err != nil {
			uc.Logger.Error().Msg(fmt.Sprintf(
				"TargetUserRotateAddonsUC error target %s and userId %d %+v", targetID, userID, err,
			))
			continue
		}
		vaultBindServices := vault.ServicesBindUser{
			ServiceName: target.Name,
			Policies:    make([]string, 0),
		}
		if uc.TargetUserQueries.CheckTargetAndUserByRole(
			targetID,
			userID,
			models.UserRoleVaultViewer,
		) {
			vaultBindServices.Policies = append(vaultBindServices.Policies, vault.Read)
		}
		if uc.TargetUserQueries.CheckTargetAndUserByRole(
			targetID,
			userID,
			models.UserRoleVaultWriter,
		) {
			vaultBindServices.Policies = append(vaultBindServices.Policies, vault.Write)
		}
		accessHarbor := uc.TargetUserQueries.CheckTargetAndUserByRole(
			targetID,
			userID,
			models.UserRoleHarborAccess,
		)
		vaultBind.Services = append(vaultBind.Services, vaultBindServices)
		harborBind = append(harborBind, HarborMembers{
			Username: k8s_utils.NormalizeK8SEntityName(k8s_utils.UsernameByEmail(user.Email)),
			TargetID: targetID,
			Access:   accessHarbor,
		})
	}
	err = uc.VaultClient.UserBindAccessServices(ctx, []vault.UsersAndServices{vaultBind})
	if err != nil {
		uc.Logger.Error().Msg(
			fmt.Sprintf(
				"TargetUserRotateAddonsUC error with vault by userId %d %+v",
				userID,
				err,
			),
		)
		return err
	}
	err = uc.HarborMembersUpdate(ctx, harborBind)
	if err != nil {
		uc.Logger.Error().Msg(
			fmt.Sprintf(
				"TargetUserRotateAddonsUC error with harbor by userId %d %+v",
				userID,
				err,
			),
		)
		return err
	}
	return nil
}

// HarborMembersUpdate – обновляет пользователей в harbor
func (uc *TargetUserRotateAddonsUC) HarborMembersUpdate(ctx context.Context, members []HarborMembers) error {
	for _, member := range members {
		uc.Logger.Info().Msg(fmt.Sprintf("HarborMembersUpdate: get target by id %s", member.TargetID))
		connectedAddons, err := uc.TargetAddonQueries.GetAllByTarget(member.TargetID)
		if err != nil {
			uc.Logger.Error().Msg(
				fmt.Sprintf(
					"HarborMembersUpdate: not found connected addons target by id %s %+v",
					member.TargetID,
					err))
			return err
		}
		for _, connectedAddon := range connectedAddons {
			if connectedAddon.AddonType != connection.HarborAddon {
				continue
			}
			data, _ := json.Marshal(connectedAddon.Config)
			var configAddon addons.AddonOperationConfig
			_ = json.Unmarshal(data, &configAddon)
			if configAddon.Name == "" {
				return errors.New("config addon name is empty by target")
			}
			uc.Logger.Error().Msg(
				fmt.Sprintf(
					"HarborMembersUpdate: found connected harbor type addon on taget id %s",
					member.TargetID,
				),
			)
			harborConfig := uc.FactoryAddonService.GetAddonConfigByID(connectedAddon.AddonID)
			harborService := addons.NewHarborMemberService(harborConfig)
			if member.Access {
				err = harborService.AddProjectMember(ctx, configAddon.Name, member.Username)
			} else {
				err = harborService.RemoveProjectMember(ctx, configAddon.Name, member.Username)
			}
			if err != nil {
				uc.Logger.Error().Msg(fmt.Sprintf(
					"HarborMembersUpdate: error by change access by project %s, %s, %+v",
					configAddon.Name,
					member.Username,
					err,
				))
				return err
			}
		}
	}
	return nil
}
