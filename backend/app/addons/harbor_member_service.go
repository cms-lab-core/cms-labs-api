package addons

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/cms-lab-core/cms-labs-api/shared/connection"
	"github.com/goccy/go-json"
	"github.com/rs/zerolog/log"
)

// NewHarborMemberService creates a new Harbor service instance.
func NewHarborMemberService(cfg *connection.AddonConfig) *HarborMemberService {
	apiUrl, _ := cfg.Params["apiUrl"].(string)
	username, _ := cfg.Params["adminUsername"].(string)
	password, _ := cfg.Params["adminPassword"].(string)
	baseUrl, _ := cfg.Params["baseUrl"].(string)
	params := HarborConfig{
		ApiURL:        apiUrl,
		AdminUsername: username,
		AdminPassword: password,
		BaseURL:       baseUrl,
	}
	return &HarborMemberService{
		Params: params,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

type HarborMemberService struct {
	Params     HarborConfig
	HTTPClient *http.Client
}

// addProjectMemberRequest – тело запроса для добавления участника
type addProjectMemberRequest struct {
	RoleID     int `json:"role_id"`
	MemberUser struct {
		UserID int `json:"user_id"`
	} `json:"member_user"`
}

// projectMember – минимальное представление участника проекта
type projectMember struct {
	ID         int    `json:"id"`
	EntityName string `json:"entity_name"`
}

type harborUser struct {
	UserID   int    `json:"user_id"`
	Username string `json:"username"`
}

var UserNotFoundError = errors.New("user not found")

// findUser ищет пользователя по имени (возвращает ошибку, если не найден)
func (s *HarborMemberService) findUser(ctx context.Context, username string) (*harborUser, error) {
	var users []harborUser
	path := "/api/v2.0/users/search?username=" + username
	err := s.doRequest(ctx, http.MethodGet, path, nil, &users)
	if err != nil {
		return nil, err
	}
	if len(users) == 0 {
		return nil, UserNotFoundError
	}
	return &users[0], nil
}

// getUserMembership возвращает ID членства пользователя в проекте (0 если не участник)
func (s *HarborMemberService) getUserMembership(ctx context.Context, projectName, username string) (int, error) {
	var members []projectMember
	path := fmt.Sprintf("/api/v2.0/projects/%s/members?entityname=%s", projectName, username)
	err := s.doRequest(ctx, http.MethodGet, path, nil, &members)
	if err != nil {
		// Если проект не существует – вернёт 404, но мы просто вернём 0
		if strings.Contains(err.Error(), "404") {
			return 0, nil
		}
		return 0, err
	}
	for _, m := range members {
		if m.EntityName == username {
			return m.ID, nil
		}
	}
	return 0, nil
}

// AddProjectMember добавляет пользователя в проект как maintainer (role_id = 4)
func (s *HarborMemberService) AddProjectMember(ctx context.Context, projectName, username string) error {
	// 1. Проверяем, существует ли пользователь в Harbor
	userIntoHarbor, err := s.findUser(ctx, username)
	if errors.Is(err, UserNotFoundError) {
		// Пользователь не найден – ничего не делаем (по заданию)
		log.Info().Msg(fmt.Sprintf("User %s not found in Harbor, skipping", username))
		return nil
	}
	if err != nil {
		return err
	}

	// 2. Формируем запрос на добавление с ролью maintainer
	req := addProjectMemberRequest{
		RoleID: 4, // maintainer
	}
	req.MemberUser.UserID = userIntoHarbor.UserID

	path := fmt.Sprintf("/api/v2.0/projects/%s/members", projectName)
	err = s.doRequest(ctx, http.MethodPost, path, req, nil)
	if err != nil {
		// 409 Conflict – пользователь уже участник – игнорируем
		if strings.Contains(err.Error(), "409") {
			log.Info().Msg(fmt.Sprintf("User %s is already a member of project %s", username, projectName))
			return nil
		}
		// 404 – проект не найден – можно игнорировать или вернуть ошибку (по ситуации)
		if strings.Contains(err.Error(), "404") {
			log.Info().Msg(fmt.Sprintf("Project %s not found", projectName))
			return nil
		}
		return fmt.Errorf("add member: %w", err)
	}
	log.Info().Msg(fmt.Sprintf("User %s added to project %s as maintainer", username, projectName))
	return nil
}

// RemoveProjectMember удаляет пользователя из проекта
func (s *HarborMemberService) RemoveProjectMember(ctx context.Context, projectName, username string) error {
	// 1. Получаем ID членства
	memberID, err := s.getUserMembership(ctx, projectName, username)
	if err != nil {
		return fmt.Errorf("get membership: %w", err)
	}
	if memberID == 0 {
		// Пользователь не участник – ничего не делаем
		log.Info().Msg(fmt.Sprintf("User %s is not a member of project %s", username, projectName))
		return nil
	}

	// 2. Удаляем
	path := fmt.Sprintf("/api/v2.0/projects/%s/members/%d", projectName, memberID)
	err = s.doRequest(ctx, http.MethodDelete, path, nil, nil)
	if err != nil {
		// 404 – членство уже удалено – игнорируем
		if strings.Contains(err.Error(), "404") {
			return nil
		}
		return fmt.Errorf("remove member: %w", err)
	}
	log.Info().Msg(fmt.Sprintf("User %s removed from project %s", username, projectName))
	return nil
}

// doRequest performs an HTTP request to Harbor API with basic authentication.
func (s *HarborMemberService) doRequest(ctx context.Context, method, path string, body, out interface{}) error {
	url := s.Params.ApiURL + path
	var reqBody []byte
	var err error
	if body != nil {
		reqBody, err = json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal request body: %w", err)
		}
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewReader(reqBody))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.SetBasicAuth(s.Params.AdminUsername, s.Params.AdminPassword)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	// Accept 2xx and 409 (already exists) – we handle 409 specially in callers.
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("unexpected status: %s", resp.Status)
	}

	if out != nil && resp.StatusCode != http.StatusNoContent {
		if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
			return fmt.Errorf("decode response: %w", err)
		}
	}
	return nil
}
