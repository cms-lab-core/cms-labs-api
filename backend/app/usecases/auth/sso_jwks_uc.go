package auth

import (
	"crypto/rsa"
	"encoding/base64"
	"math/big"

	"github.com/cms-lab-core/cms-labs-api/backend/pkg/configs"
)

type SSOJWKSOutputDTO struct {
	Keys []JWK `json:"keys"`
}

type JWK struct {
	Kty string `json:"kty"`
	Use string `json:"use,omitempty"`
	Kid string `json:"kid"`
	Alg string `json:"alg"`
	N   string `json:"n"`
	E   string `json:"e"`
}

type SSOJwksInputDTO struct {
	PublicKeys map[string]*rsa.PublicKey // kid -> PublicKey
}

type SSOJwksUC struct {
}

func (u *SSOJwksUC) Execute() (*SSOJWKSOutputDTO, error) {
	inputDTO := &SSOJwksInputDTO{
		PublicKeys: map[string]*rsa.PublicKey{
			"access":  configs.AppConfig.JWT.AccessKey.PublicKey,
			"id":      configs.AppConfig.JWT.AccessKey.PublicKey,
			"refresh": configs.AppConfig.JWT.RefreshKey.PublicKey,
		},
	}
	keys := make([]JWK, 0, len(inputDTO.PublicKeys))

	for kid, publicKey := range inputDTO.PublicKeys {
		jwk := JWK{
			Kty: "RSA",
			Use: "sig",
			Kid: kid,
			Alg: "RS256",
			N:   base64.RawURLEncoding.EncodeToString(publicKey.N.Bytes()),
			E:   base64.RawURLEncoding.EncodeToString(big.NewInt(int64(publicKey.E)).Bytes()),
		}
		keys = append(keys, jwk)
	}

	return &SSOJWKSOutputDTO{
		Keys: keys,
	}, nil
}
