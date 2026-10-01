package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/cms-lab-core/cms-labs-api/shared/cms_client"

	"github.com/cms-lab-core/cms-labs-api/backend/pkg/configs"

	jwt "github.com/golang-jwt/jwt/v5"
)

type TokenDataWithExp struct {
	Token string `json:"token"`
	Exp   int64  `json:"exp"`
}

// NewWithClaims creates a new [Token] with the specified signing method and
// claims. Additional options can be specified, but are currently unused.
func NewWithClaims(kid string, method jwt.SigningMethod, claims jwt.Claims, opts ...jwt.TokenOption) *jwt.Token {
	return &jwt.Token{
		Header: map[string]interface{}{
			"typ": "JWT",
			"alg": method.Alg(),
			"kid": kid,
		},
		Claims: claims,
		Method: method,
	}
}

// GenerateNewTokens func for generate a new Access & Refresh tokens. Return jti (unique id on refresh token)
func GenerateNewTokens(entity *cms_client.SSOTokenPublicData, state string) (*cms_client.SSOToken, string, error) {
	// Generate JWT Access token.
	jwtConfig := configs.AppConfig.JWT.AccessKey
	accessToken, err := GenerateAndSignNewAccessToken(entity, jwtConfig.Expire)
	if err != nil {
		// Return token generation error.
		return nil, "", err
	}

	// Generate JWT Refresh token.
	refreshToken, jti, err := generateNewRefreshToken()
	if err != nil {
		// Return token generation error.
		return nil, "", err
	}

	return &cms_client.SSOToken{
		AccessToken:  accessToken.Token,
		RefreshToken: refreshToken.Token,
		IdToken:      accessToken.Token,
		TokenType:    "bearer",
		ExpiresIn:    accessToken.Exp,
		State:        state,
		UserId:       entity.Sub,
	}, jti, nil
}

func GenerateAndSignNewAccessToken(
	entity *cms_client.SSOTokenPublicData,
	expire time.Duration,
) (*TokenDataWithExp, error) {
	// Получаем конфигурацию JWT
	jwtConfig := configs.AppConfig.JWT.AccessKey

	// Проверяем наличие приватного ключа
	if jwtConfig.PrivateKey == nil {
		return nil, errors.New("access token private key not configured")
	}

	// Устанавливаем срок действия токена
	entity.Iat = time.Now().Unix()
	entity.Exp = time.Now().Add(expire).Unix()

	// Создаем claims
	claims := entity.JWTClaims()

	// Создаем токен с RSA алгоритмом
	token := NewWithClaims("access", jwt.SigningMethodRS256, claims)

	// Подписываем токен с использованием приватного ключа
	t, err := token.SignedString(jwtConfig.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to sign access token: %w", err)
	}

	return &TokenDataWithExp{
		Token: t,
		Exp:   entity.Exp,
	}, nil
}

func generateNewRefreshToken() (*TokenDataWithExp, string, error) {
	// Получаем конфигурацию JWT
	jwtConfig := configs.AppConfig.JWT.RefreshKey

	// Проверяем наличие приватного ключа
	if jwtConfig.PrivateKey == nil {
		return nil, "", errors.New("refresh token private key not configured")
	}

	// Устанавливаем срок действия токена
	expireTime := time.Now().Add(jwtConfig.Expire)
	expireUnix := expireTime.Unix()

	// Генерируем уникальный идентификатор для refresh токена
	uniquePart := uuid.New().String()

	// Создаем claims с минимальным набором данных
	claims := jwt.MapClaims{
		"jti": uniquePart,        // Уникальный идентификатор токена
		"exp": expireUnix,        // Время истечения
		"iat": time.Now().Unix(), // Время выпуска
	}

	// Создаем токен с RSA алгоритмом
	token := NewWithClaims("refresh", jwt.SigningMethodRS256, claims)

	// Подписываем токен с использованием приватного ключа
	t, err := token.SignedString(jwtConfig.PrivateKey)
	if err != nil {
		return nil, "", fmt.Errorf("failed to sign refresh token: %w", err)
	}

	return &TokenDataWithExp{
		Token: t,
		Exp:   expireUnix,
	}, uniquePart, nil
}

// ParseRefreshToken func for parse second argument from refresh token.
// ParseRefreshToken проверяет валидность refresh токена и возвращает время его истечения
func ParseRefreshToken(refreshToken string) (int64, string, error) {
	// Проверяем, что токен не пустой
	if refreshToken == "" {
		return 0, "", errors.New("empty refresh token")
	}

	// Получаем публичный ключ для проверки подписи
	publicKey := configs.AppConfig.JWT.RefreshKey.PublicKey
	if publicKey == nil {
		return 0, "", errors.New("refresh token public key not configured")
	}

	// Парсим токен с проверкой подписи
	token, err := jwt.Parse(refreshToken, func(token *jwt.Token) (interface{}, error) {
		// Проверяем, что используется ожидаемый алгоритм подписи
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return publicKey, nil
	})

	if err != nil {
		return 0, "", fmt.Errorf("invalid refresh token: %w", err)
	}

	// Проверяем, что токен валиден
	if !token.Valid {
		return 0, "", errors.New("invalid refresh token")
	}

	// Извлекаем claims и проверяем наличие времени истечения
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return 0, "", errors.New("invalid token claims")
	}

	exp, ok := claims["exp"].(float64)
	if !ok {
		return 0, "", errors.New("token expiration time (exp) not found")
	}

	jti, ok := claims["jti"].(string)
	if !ok {
		return 0, "", errors.New("token identifier not found")
	}

	return int64(exp), jti, nil
}
