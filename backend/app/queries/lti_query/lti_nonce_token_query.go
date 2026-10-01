package lti_query

import (
	"fmt"
	"time"

	"github.com/rs/zerolog"

	"github.com/cms-lab-core/cms-labs-api/backend/app/models"
	"gorm.io/gorm"
)

type LTINonceTokenQueries struct {
	DB     *gorm.DB
	Logger *zerolog.Logger
}

func (q *LTINonceTokenQueries) ClearOldValues() error {
	timeThreshold := time.Now().UTC().Add(-5 * time.Minute)
	result := q.DB.Where("created_at <= ?", timeThreshold).Delete(&models.LTINonceToken{})
	if result.Error != nil {
		return result.Error
	}
	q.Logger.Info().Msg(fmt.Sprintf("LTINonceTokenQueries: cleared old values %+v", timeThreshold))
	return nil
}

func (q *LTINonceTokenQueries) StoreNonce(nonce string, targetLinkURI string) error {
	if err := q.ClearOldValues(); err != nil {
		return err
	}
	entity := &models.LTINonceToken{}
	entity.ID = 0
	entity.Nonce = nonce
	entity.TargetLinkURI = targetLinkURI
	entity.CreatedAt = time.Now().UTC()
	entity.UpdatedAt = time.Now().UTC()
	q.Logger.Debug().Msg(fmt.Sprintf("LTINonceTokenQueries: store nonce %+v", entity))
	result := q.DB.Create(entity)
	q.Logger.Info().Msg(fmt.Sprintf("LTINonceTokenQueries: stored nonce: %+v", entity.Nonce))
	return result.Error
}

func (q *LTINonceTokenQueries) TestAndClearNonce(
	nonce string,
	targetLinkURI string,
) error {
	if err := q.ClearOldValues(); err != nil {
		return err
	}
	entityDB := &models.LTINonceToken{}
	q.Logger.Debug().Msg(fmt.Sprintf(
		"LTINonceTokenQueries: check nonce: %+v and targetLinkURI: %+v",
		nonce,
		targetLinkURI,
	))
	result := q.DB.Where("nonce = ?", nonce).Where("target_link_uri = ?", targetLinkURI).Find(&entityDB)
	q.Logger.Info().Msg(fmt.Sprintf(
		"LTINonceTokenQueries: checked nonce token: %+v",
		nonce,
	))
	if result.Error != nil {
		return result.Error
	}
	q.DB.Where("id = ?", entityDB.ID).Delete(&entityDB)
	return nil
}
