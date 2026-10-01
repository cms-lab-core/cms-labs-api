package usecases

import (
	"time"

	"github.com/cms-lab-core/cms-labs-api/backend/app/models"
	"github.com/cms-lab-core/cms-labs-api/backend/app/models/types"
	"github.com/cms-lab-core/cms-labs-api/backend/app/queries"
	"github.com/cms-lab-core/cms-labs-api/shared/cms_client"
	"github.com/cms-lab-core/cms-labs-api/shared/jsonrpc"
	"github.com/google/uuid"
	"gorm.io/datatypes"
)

// TargetUpsertUC – создание или обновление цели (Target)
type TargetUpsertUC struct {
	TargetQueries     *queries.TargetQueries
	TargetUserQueries *queries.TargetUserQueries
	User              *cms_client.SSOTokenPublicData
}

// TargetUpsertInputDTO – входные данные для создания/обновления
type TargetUpsertInputDTO struct {
	ID          *string              `json:"id"` // nil – создание, иначе – обновление
	Name        string               `json:"name" validate:"required"`
	Description *string              `json:"description"`
	Type        string               `json:"type" validate:"required"`
	Links       *[]models.TargetLink `json:"links"`
	Tags        *[]string            `json:"tags"`
}

// TargetUpsertOutputDTO – результат
type TargetUpsertOutputDTO struct {
	ID string `json:"id"`
}

type TargetUpsertRequest struct {
	JSONRPC string               `json:"jsonrpc" default:"2.0" required:"true"`
	Method  string               `json:"method" default:"service_card.upsert" required:"true"`
	Params  TargetUpsertInputDTO `json:"params,omitempty"`
	ID      string               `json:"id,omitempty" default:"1" required:"true"`
}

type TargetUpsertResponse struct {
	JSONRPC string                `json:"jsonrpc" default:"2.0" required:"true"`
	Result  TargetUpsertOutputDTO `json:"result,omitempty"`
	Error   interface{}           `json:"error,omitempty"`
	ID      string                `json:"id,omitempty" default:"1" required:"true"`
}

var (
	ErrPermissionDenied = jsonrpc.NewRpcError("permission_denied", "You must be an editor of this target")
)

func (uc *TargetUpsertUC) SetContext(user *cms_client.SSOTokenPublicData) *TargetUpsertUC {
	uc.User = user
	return uc
}

// Execute – основной метод
func (uc *TargetUpsertUC) Execute(dto TargetUpsertInputDTO) (*TargetUpsertOutputDTO, error) {
	entity := &models.Target{
		Name:        dto.Name,
		Description: dto.Description,
		Type:        dto.Type,
	}
	if dto.Links != nil {
		links := datatypes.NewJSONSlice(*dto.Links)
		entity.Links = &links
	}

	if dto.Tags != nil {
		tags := datatypes.NewJSONSlice(*dto.Tags)
		entity.Tags = &tags
	}

	// Обновление существующей цели
	if dto.ID != nil {
		entity.ID = *dto.ID

		// Проверяем, существует ли цель
		existing, err := uc.TargetQueries.Get(*dto.ID)
		if err != nil {
			return nil, err
		}

		// Проверяем, является ли текущий пользователь редактором
		if !uc.TargetUserQueries.CheckTargetAndUserByRole(existing.ID, uc.User.UserID(), models.UserRoleEditor) {
			return nil, ErrPermissionDenied
		}

		entity.Name = existing.Name
		entity.InternalTags = existing.InternalTags
		entity.InternalLinks = existing.InternalLinks
		entity.CreatedAt = existing.CreatedAt
		entity.SynchronizedAt = existing.SynchronizedAt
		entity.Type = existing.Type
		entity.UpdatedAt = time.Now().UTC()

		if err := uc.TargetQueries.Upsert(entity); err != nil {
			return nil, err
		}
		return &TargetUpsertOutputDTO{ID: entity.ID}, nil
	}
	// Только пользователь с ролью admin или instructor может создавать <Target>
	if !cms_client.SSOHasIntersection(
		[]string{
			cms_client.SSOUsersRoleInstructor,
			cms_client.SSOUsersRoleAdmin,
		},
		uc.User.Roles,
	) {
		return nil, ErrPermissionDenied
	}

	// Создание новой цели
	entity.ID = uuid.New().String()
	entity.CreatedAt = time.Now().UTC()
	entity.UpdatedAt = time.Now().UTC()
	entity.SynchronizedAt = time.Now().UTC()

	if err := uc.TargetQueries.Upsert(entity); err != nil {
		return nil, err
	}

	// Добавляем создателя как редактора
	owner := &models.TargetUser{
		TargetUserBase: models.TargetUserBase{
			TargetID: entity.ID,
			UserID:   uc.User.UserID(),
			Roles: types.JsonStore{
				"roles": []string{models.UserRoleEditor},
			},
		},
	}
	if err := uc.TargetUserQueries.Upsert(owner); err != nil {
		// Роллбэк: удаляем только что созданную цель
		_ = uc.TargetQueries.Delete(entity.ID)
		return nil, err
	}

	return &TargetUpsertOutputDTO{ID: entity.ID}, nil
}
