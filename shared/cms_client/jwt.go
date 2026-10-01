package cms_client

import (
	"github.com/cms-lab-core/cms-labs-api/shared/jsonrpc"
	jwt "github.com/golang-jwt/jwt/v5"
)

func SSOTokenGetStringSlice(claims map[string]interface{}, key string) []string {
	var result []string
	if val, ok := claims[key]; ok {
		if slice, ok := val.([]interface{}); ok {
			for _, item := range slice {
				if str, ok := item.(string); ok {
					result = append(result, str)
				}
			}
		}
	}
	return result
}

func SSODecodeToken(jwtToken *jwt.Token) (*SSOTokenPublicData, error) {
	// Setting and checking token and credentials.
	claims, ok := jwtToken.Claims.(jwt.MapClaims)
	if !ok {
		return nil, jsonrpc.NewRpcError("invalid_token", "token is invalid")
	}
	var serverID *uint = nil
	k8sAccessType := ""
	if val, ok := claims["server_id"].(float64); ok {
		uintServerID := uint(val)
		serverID = &uintServerID
	}
	if val, ok := claims["k8s:access_type"].(string); ok {
		k8sAccessType = val
	}
	tokenData := SSOTokenPublicData{
		Iss:          claims["iss"].(string),
		Sub:          claims["sub"].(string),
		Aud:          claims["aud"].(string),
		Azp:          claims["azp"].(string),
		Exp:          int64(claims["exp"].(float64)),
		Iat:          int64(claims["iat"].(float64)),
		Nonce:        claims["nonce"].(string),
		Email:        claims["email"].(string),
		Name:         claims["name"].(string),
		Username:     claims["username"].(string),
		ServerID:     serverID,
		Roles:        SSOTokenGetStringSlice(claims, "roles"),
		LastLaunchId: claims["last_launch_id"].(string),
		K8SType:      k8sAccessType,
	}
	return &tokenData, nil
}

// SSOHasIntersection Вспомогательная функция для проверки пересечения ролей
func SSOHasIntersection(allowedRoles, userRoles []string) bool {
	for _, userRole := range userRoles {
		for _, allowedRole := range allowedRoles {
			if userRole == allowedRole {
				return true
			}
		}
	}
	return false
}
