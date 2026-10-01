package lti_query

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/rs/zerolog"

	json "github.com/goccy/go-json"

	"github.com/cms-lab-core/cms-labs-api/backend/app/models"
	"github.com/cms-lab-core/cms-labs-api/shared/jsonrpc"
	"gorm.io/gorm"
)

func accessTokenIndex(tokenURI, clientID string, scopes []string) string {
	return tokenURI + ";" + clientID + ";" + strings.Join(scopes[:], " ")
}

type LTIAccessTokenQueries struct {
	DB     *gorm.DB
	Logger *zerolog.Logger
}

func (q *LTIAccessTokenQueries) GetByIndex(index string) (models.LTIAccessToken, error) {
	var entity models.LTIAccessToken
	q.DB.Where("`index` = ?", index).Find(&entity)
	if entity.Index != index {
		return entity, jsonrpc.NewRpcError(
			"has_not_access", "lti access token has not found",
		)
	}
	return entity, nil
}

func (q *LTIAccessTokenQueries) Upsert(entity *models.LTIAccessToken) error {
	if entity == nil {
		return nil
	}
	entityDB, _ := q.GetByIndex(entity.Index)
	if entityDB.ID != 0 {
		// Update
		q.Logger.Debug().Msg(fmt.Sprintf("LTIAccessTokenQueries: entity update: %+v", entityDB))
		entity.ID = entityDB.ID
		entity.CreatedAt = entityDB.CreatedAt
		entity.UpdatedAt = time.Now().UTC()
		result := q.DB.Save(&entity)
		return result.Error
	} else {
		// Create
		q.Logger.Debug().Msg(fmt.Sprintf("LTIAccessTokenQueries: entity create: %+v", entity))
		entity.ID = 0
		entity.CreatedAt = time.Now().UTC()
		entity.UpdatedAt = time.Now().UTC()
		result := q.DB.Create(entity)
		return result.Error
	}
}

// StoreAccessToken stores bearer tokens for potential reuse.
func (q *LTIAccessTokenQueries) StoreAccessToken(token AccessToken) error {
	q.Logger.Info().Msg(fmt.Sprintf("LTIAccessTokenQueries: StoreAccessToken on clientID: %+v", token.ClientID))
	if token.TokenURI == "" {
		return errors.New("received empty tokenURI")
	}
	if token.ClientID == "" {
		return errors.New("received empty clientID")
	}
	if len(token.Scopes) == 0 {
		return errors.New("received empty scopes")
	}
	if token.Token == "" {
		return errors.New("received empty accessToken")
	}
	zeroTime := time.Time{}
	if token.ExpiryTime == zeroTime {
		return errors.New("received empty expiry time")
	}

	sort.Strings(token.Scopes)

	storeValue, err := json.Marshal(token)
	if err != nil {
		return fmt.Errorf("error encoding access token to store: %w", err)
	}
	entity := models.LTIAccessToken{}
	entity.Index = accessTokenIndex(
		token.TokenURI,
		token.ClientID,
		token.Scopes,
	)
	entity.Payload = string(storeValue)
	return q.Upsert(&entity)
}

// FindAccessToken retrieves bearer tokens for potential reuse.
func (q *LTIAccessTokenQueries) FindAccessToken(tokenURI, clientID string, scopes []string) (AccessToken, error) {
	q.Logger.Info().Msg(fmt.Sprintf("LTIAccessTokenQueries: FindAccessToken on clientID: %+v", clientID))
	if tokenURI == "" {
		return AccessToken{}, errors.New("received empty tokenURI")
	}
	if clientID == "" {
		return AccessToken{}, errors.New("received empty clientID")
	}
	if len(scopes) == 0 {
		return AccessToken{}, errors.New("received empty scopes")
	}

	index := accessTokenIndex(tokenURI, clientID, scopes)
	entity, err := q.GetByIndex(index)
	accessToken := AccessToken{}
	if err != nil {
		return accessToken, ErrAccessTokenNotFound
	}
	err = json.Unmarshal([]byte(entity.Payload), &accessToken)
	if err != nil {
		return accessToken, fmt.Errorf("could not decode access token: %w", err)
	}
	if accessToken.ExpiryTime.Before(time.Now()) {
		return AccessToken{}, ErrAccessTokenExpired
	}
	return accessToken, nil
}
