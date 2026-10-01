package queries

import (
	"fmt"

	"github.com/rs/zerolog"

	"github.com/cms-lab-core/cms-labs-api/backend/app/models"
	"github.com/cms-lab-core/cms-labs-api/shared/jsonrpc"
	"gorm.io/gorm"
)

type TokenAttemptQueries struct {
	DB     *gorm.DB
	Logger *zerolog.Logger
}

var (
	TokenAttemptNotFoundError = jsonrpc.NewRpcError(
		"token_attempt_not_found",
		"token attempt has not found",
	)
)

func (q *TokenAttemptQueries) GetByTokenId(tokenId string) (models.TokenAttempt, error) {
	entity := models.TokenAttempt{}
	q.DB.Where("token = ?", tokenId).Limit(1).Find(&entity)
	if entity.ID == 0 {
		return entity, TokenAttemptNotFoundError
	}
	return entity, nil
}

func (q *TokenAttemptQueries) GetByAuthCode(code string) (models.TokenAttempt, error) {
	entity := models.TokenAttempt{}
	q.DB.Where("authorization_code = ?", code).Limit(1).Find(&entity)
	if entity.ID == 0 {
		return entity, TokenAttemptNotFoundError
	}
	return entity, nil
}

func (q *TokenAttemptQueries) Upsert(entity *models.TokenAttempt) error {
	err := q.DeleteByParams(entity.UserID, entity.ServerID, entity.TargetID)
	if err != nil {
		return nil
	}
	result := q.DB.Create(entity)
	q.Logger.Debug().Msg(fmt.Sprintf("TokenAttemptQueries: entity create: %+v", entity))
	q.Logger.Info().Msg(fmt.Sprintf(
		"TokenAttemptQueries: entity create user_id=%v, target_id=%v",
		entity.UserID,
		entity.TargetID,
	),
	)
	return result.Error
}

func (q *TokenAttemptQueries) GetByParams(userID *uint, serverID *uint, targetID *string) (models.TokenAttempt, error) {
	entity := models.TokenAttempt{}
	if serverID != nil && userID != nil {
		q.DB.Where("user_id = ?", *userID).Where("server_id = ?", *serverID).
			Limit(1).Find(&entity)
	} else if targetID != nil && userID != nil {
		q.DB.Where("user_id = ?", *userID).Where("target_id = ?", *targetID).
			Limit(1).Find(&entity)
	}

	if entity.ID == 0 {
		return entity, TokenAttemptNotFoundError
	}
	return entity, nil
}

func (q *TokenAttemptQueries) DeleteByParams(userID *uint, serverID *uint, targetID *string) error {
	if serverID != nil && userID != nil {
		_ = q.DB.Where("user_id = ?", *userID).Where("server_id = ?", *serverID).Delete(
			&models.TokenAttempt{},
		)
		q.Logger.Info().Msg(
			fmt.Sprintf(
				"TokenAttemptQueries: delete token by user_id=%+v server_id=%+v",
				*userID,
				*serverID,
			),
		)
	}
	if targetID != nil && userID != nil {
		_ = q.DB.Where("user_id = ?", *userID).Where("target_id = ?", *targetID).Delete(
			&models.TokenAttempt{},
		)
		q.Logger.Info().Msg(
			fmt.Sprintf(
				"TokenAttemptQueries: delete token by user_id=%+v target_id=%+v",
				*userID,
				*targetID,
			),
		)
	}
	return nil
}
