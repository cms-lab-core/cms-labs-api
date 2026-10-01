package lti_query

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"time"

	"github.com/rs/zerolog"

	"github.com/cms-lab-core/cms-labs-api/backend/app/models"
	"github.com/cms-lab-core/cms-labs-api/shared/jsonrpc"
	"gorm.io/gorm"
)

type AuthProviderQueries struct {
	DB     *gorm.DB
	Logger *zerolog.Logger
}

type AuthProviderListInputDTO struct {
	Search string `json:"search"`
	IsAuth bool   `json:"is_auth"`
	Limit  int    `json:"limit"`
	Offset int    `json:"offset"`
}

func ltiGenerateKeys(l *zerolog.Logger) (string, string, error) {
	l.Info().Msg("LTIGenerate RSA keys")
	// Generate RSA private key
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		l.Warn().Msg(fmt.Sprintf("Error generating private key: %+v", err))
		return "", "", err
	}

	// Encode private key to PEM format
	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})

	// Generate RSA public key
	publicKey := &privateKey.PublicKey

	// Encode public key to PEM format
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		l.Warn().Msg(fmt.Sprintf("Error marshaling public key: %+v", err))
		return "", "", err
	}
	publicKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: publicKeyBytes,
	})
	return string(privateKeyPEM), string(publicKeyPEM), nil
}
func (q *AuthProviderQueries) Get(id uint) (models.AuthProvider, error) {
	var entity models.AuthProvider
	result := q.DB.First(&entity, id)
	if result.Error != nil && result.Error.Error() == "record not found" {
		return entity, jsonrpc.NewRpcError("auth_provider_not_found", "provider not found")
	}
	return entity, result.Error
}

func (q *AuthProviderQueries) Upsert(entity *models.AuthProvider) error {
	if entity == nil {
		return nil
	}
	if entity.Type != models.AuthProviderTypeLTI && entity.Type != models.AuthProviderTypeLDAP {
		return jsonrpc.NewRpcError("invalid auth provider type", "Invalid provider type")
	}
	entityDB := models.AuthProvider{}
	q.DB.Where("id = ?", entity.ID).Find(&entityDB)
	if entityDB.ID != 0 {
		// Update
		q.Logger.Debug().Msg(fmt.Sprintf("AuthProviderQueries: entity update: %+v", entityDB))
		q.Logger.Info().Msg(fmt.Sprintf("AuthProviderQueries: entity update id=%+v", entityDB.ID))
		entity.ID = entityDB.ID
		entity.Type = entityDB.Type
		entity.CreatedAt = entityDB.CreatedAt
		entity.PublicKey = entityDB.PublicKey
		entity.PrivateKey = entityDB.PrivateKey
		entity.UpdatedAt = time.Now().UTC()
		result := q.DB.Save(&entity)
		return result.Error
	}
	// Create
	q.Logger.Debug().Msg(fmt.Sprintf("AuthProviderQueries: entity create: %+v", entity))
	q.Logger.Info().Msg(fmt.Sprintf("AuthProviderQueries: entity create name=%+v", entity.Name))
	entity.ID = 0
	entity.CreatedAt = time.Now().UTC()
	entity.UpdatedAt = time.Now().UTC()
	privateKey, publicKey, err := ltiGenerateKeys(q.Logger)
	if err != nil {
		return err
	}
	if entity.Type == models.AuthProviderTypeLTI {
		entity.PublicKey = publicKey
		entity.PrivateKey = privateKey
	}
	result := q.DB.Create(entity)
	return result.Error
}

const MaxLimitCount = 5000

func (q *AuthProviderQueries) List(
	dto *AuthProviderListInputDTO,
) ([]models.AuthProviderListItem, int64, error) {
	var entities []models.AuthProviderListItem
	result := q.listFilter(
		dto,
		q.DB.Limit(MaxLimitCount).Offset(0),
	).Find(&entities)
	count := result.RowsAffected
	q.Logger.Debug().Msg(fmt.Sprintf("AuthProviderQueries list: count %+v", count))
	result = q.listFilter(
		dto,
		q.DB.Limit(dto.Limit).Offset(dto.Offset),
	).Find(&entities)
	q.Logger.Debug().Msg(fmt.Sprintf("AuthProviderQueries list: entities %+v", entities))
	return entities, count, result.Error
}

func (q *AuthProviderQueries) listFilter(dto *AuthProviderListInputDTO, tx *gorm.DB) *gorm.DB {
	tx = tx.Model(&models.AuthProvider{}).Order("created_at desc")

	if dto.IsAuth {
		// (sso_url is not null and sso_url != '' and type = 'lti') or type = 'ldap'
		tx = tx.Where("(sso_url IS NOT NULL AND sso_url != '' AND type = ?) OR type = ?",
			models.AuthProviderTypeLTI, models.AuthProviderTypeLDAP)
	}

	if dto.Search != "" {
		searchCond := tx.Where(
			"base_uri LIKE ?",
			fmt.Sprintf("%%%s%%", dto.Search),
		).Or("id = ?", dto.Search)
		tx = tx.Where(searchCond)
	}

	return tx
}

func (q *AuthProviderQueries) Delete(id uint) error {
	err := q.DB.Where("id = ?", id).Delete(&models.AuthProvider{}).Error
	q.Logger.Debug().Msg(fmt.Sprintf("AuthProviderQueries: delete entity by id: %+v", id))
	return err
}
