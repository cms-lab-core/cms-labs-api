package addons

import (
	"context"
	"fmt"
	"time"

	"github.com/cms-lab-core/cms-labs-api/backend/app/addons/vault"
	"github.com/cms-lab-core/cms-labs-api/backend/app/models"
	"github.com/cms-lab-core/cms-labs-api/backend/app/queries"
	"github.com/cms-lab-core/cms-labs-api/backend/app/usecases/auth"
	"github.com/cms-lab-core/cms-labs-api/shared/cms_client"
	"github.com/cms-lab-core/cms-labs-api/shared/connection"
	"github.com/cms-lab-core/cms-labs-api/shared/k8s_utils"
	"github.com/google/uuid"
)

const ExpireServiceDuration = time.Duration(87600) * time.Hour // 10 лет
const ExpireUserDuration = time.Duration(24) * time.Hour       // 1 день

// KubernetesAddonService implements AddonService for k8s.
type KubernetesAddonService struct {
	IssId               string
	Config              *connection.AddonConfig
	VaultClient         vault.ClientInterface
	TokenAttemptQueries *queries.TokenAttemptQueries
	ServerQueries       *queries.ServerQueries
	UserQueries         *queries.UserQueries
	TargetQueries       *queries.TargetQueries
}

func (s *KubernetesAddonService) GetType() connection.AddonType {
	return s.Config.Type
}

func (s *KubernetesAddonService) Create(ctx context.Context, emailOrTargetName string, name *string) (*AddonOperationConfig, error) {
	serverID := s.getServerId()
	if serverID == 0 {
		return nil, fmt.Errorf("not create kubernetes token with %s", emailOrTargetName)
	}
	err := s.CreateByService(ctx, emailOrTargetName)
	if err != nil {
		exp := time.Now().Add(ExpireServiceDuration).Unix()
		return &AddonOperationConfig{
			Name:    k8s_utils.NormalizeK8SEntityName(k8s_utils.UsernameByEmail(emailOrTargetName)),
			Expired: &exp,
		}, nil
	}
	err = s.CreateByEmail(ctx, emailOrTargetName)
	if err != nil {
		exp := time.Now().Add(ExpireUserDuration).Unix()
		return &AddonOperationConfig{
			Name:    k8s_utils.NormalizeK8SEntityName(k8s_utils.UsernameByEmail(emailOrTargetName)),
			Expired: &exp,
		}, nil
	}
	return nil, err
}

func (s *KubernetesAddonService) CreateByService(ctx context.Context, targetName string) error {
	serverID := s.getServerId()
	if serverID == 0 {
		return fmt.Errorf("not create kubernetes token with %s", targetName)
	}
	service, err := s.TargetQueries.GetByName(targetName)
	if err != nil {
		return err
	}
	token := cms_client.SSOTokenPublicData{
		Iss:          s.IssId,
		Sub:          fmt.Sprintf("service-%s", service.ID),
		Aud:          models.ServerTypeKubernetes,
		Azp:          models.ServerTypeKubernetes,
		Nonce:        uuid.New().String(), // Identify on current token
		Email:        service.Name,
		Username:     k8s_utils.NormalizeK8SEntityName(k8s_utils.UsernameByEmail(targetName)),
		Name:         service.Name,
		ServerID:     &serverID,
		Roles:        []string{},
		LastLaunchId: "",
		K8SType:      cms_client.SSOK8STypeService,
	}
	accessToken, err := auth.GenerateAndSignNewAccessToken(&token, ExpireServiceDuration)
	if err != nil {
		return err
	}
	tokenAttempt := &models.TokenAttempt{}
	tokenAttempt.ServerID = &serverID
	tokenAttempt.Token = token.Nonce
	tokenAttempt.Nonce = token.Nonce
	tokenAttempt.TargetID = &service.ID
	err = s.TokenAttemptQueries.Upsert(tokenAttempt)
	if err != nil {
		return err
	}
	err = s.VaultClient.CreateServiceExtension(
		ctx,
		targetName,
		fmt.Sprintf("%s_%s", s.Config.Type, s.Config.Tag),
		map[string]any{
			"K8S_TOKEN": accessToken.Token,
		},
	)
	if err != nil {
		return err
	}
	err = s.VaultClient.CreateKubernetesRole(
		ctx,
		targetName,
	)
	if err != nil {
		return err
	}
	return nil
}

func (s *KubernetesAddonService) CreateByEmail(ctx context.Context, email string) error {
	serverID := s.getServerId()
	if serverID == 0 {
		return fmt.Errorf("not create kubernetes token with %s", email)
	}
	user, err := s.UserQueries.GetByEmail(email)
	var token cms_client.SSOTokenPublicData
	if err != nil {
		return err
	}
	token = cms_client.SSOTokenPublicData{
		Iss:          s.IssId,
		Sub:          fmt.Sprintf("%d", user.ID),
		Aud:          models.ServerTypeKubernetes,
		Azp:          models.ServerTypeKubernetes,
		Nonce:        uuid.New().String(), // Identify on current token
		Email:        user.Email,
		Username:     k8s_utils.NormalizeK8SEntityName(k8s_utils.UsernameByEmail(email)),
		Name:         user.Name,
		ServerID:     &serverID,
		Roles:        []string{},
		LastLaunchId: "",
		K8SType:      cms_client.SSOK8STypeUser,
	}
	accessToken, err := auth.GenerateAndSignNewAccessToken(&token, ExpireUserDuration)
	if err != nil {
		return err
	}
	tokenAttempt := &models.TokenAttempt{}
	tokenAttempt.ServerID = &serverID
	tokenAttempt.UserID = &user.ID
	tokenAttempt.Token = token.Nonce
	tokenAttempt.Nonce = token.Nonce
	err = s.TokenAttemptQueries.Upsert(tokenAttempt)
	if err != nil {
		return err
	}
	err = s.VaultClient.CreateUsernameExtension(
		ctx,
		token.Username,
		fmt.Sprintf("%s_%s", s.Config.Type, s.Config.Tag),
		map[string]any{
			"K8S_TOKEN": accessToken.Token,
		},
	)
	if err != nil {
		return err
	}
	return nil
}

func (s *KubernetesAddonService) getServerId() uint {
	server, err := s.ServerQueries.GetByClientId(models.ServerTypeKubernetes)
	if err == nil && server.ID != 0 {
		return server.ID
	}
	entityDB := models.Server{}
	entityDB.Name = "Внутренний сервис k8s"
	entityDB.Type = models.ServerTypeKubernetes
	entityDB.ClientID = models.ServerTypeKubernetes
	entityDB.CreatedAt = time.Now().UTC()
	entityDB.UpdatedAt = time.Now().UTC()
	_ = s.ServerQueries.Upsert(&entityDB)
	return entityDB.ID
}

// Delete removes the database and the user. No Vault interaction – Vault secret is deleted separately.
func (s *KubernetesAddonService) Delete(ctx context.Context, emailOrTargetName string, cfg *AddonOperationConfig) (*AddonOperationConfig, error) {
	serverID := s.getServerId()
	if serverID == 0 {
		return nil, fmt.Errorf("not delete kubernetes token with %s", emailOrTargetName)
	}
	service, _ := s.TargetQueries.GetByName(emailOrTargetName)
	user, _ := s.UserQueries.GetByEmail(emailOrTargetName)
	if user.ID == 0 && service.ID == "" {
		return nil, fmt.Errorf("not found target for delete kubernetes token with %s", emailOrTargetName)
	}
	err := s.TokenAttemptQueries.DeleteByParams(&user.ID, &serverID, &service.ID)
	if err != nil {
		return nil, err
	}
	err = s.VaultClient.RevokeKubernetesRole(
		ctx,
		emailOrTargetName,
	)
	if err != nil {
		return nil, err
	}
	return &AddonOperationConfig{}, nil
}

// Reset rotates the password, updates Vault, and changes the user's password.
func (s *KubernetesAddonService) Reset(
	ctx context.Context,
	targetName string,
	cfg *AddonOperationConfig,
) (*AddonOperationConfig, error) {
	return s.Create(ctx, targetName, nil)
}
