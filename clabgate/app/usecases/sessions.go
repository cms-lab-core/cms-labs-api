package usecases

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"time"

	"github.com/cms-lab-core/cms-labs-api/clabgate/app/queries"
	"github.com/cms-lab-core/cms-labs-api/clabgate/pkg/configs"
	"github.com/cms-lab-core/cms-labs-api/shared/cms_client"
	"github.com/cms-lab-core/cms-labs-api/shared/jsonrpc"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
)

type SessionsUC struct {
	*zerolog.Logger
	KubernetesAdminQuery *queries.KubernetesAdminQuery
	CMSClient            *cms_client.CMSClient
	LabCatalog           *queries.LabCatalog
	user                 *cms_client.SSOTokenPublicData
}

type SessionEnsureInputDTO struct {
	AttemptID string `json:"attempt_id" validate:"required"`
}

type SessionGetInputDTO struct {
	SessionID string `json:"session_id" validate:"required"`
}

type SessionListInputDTO struct{}

type SessionStopInputDTO struct {
	SessionID string `json:"session_id" validate:"required"`
}

type SessionCheckInputDTO struct {
	SessionID string `json:"session_id" validate:"required"`
}

type SessionOpenInputDTO struct {
	SessionID string `json:"session_id" validate:"required"`
}

type SessionOutputDTO struct {
	Session queries.SessionRecord `json:"session"`
}

type SessionListOutputDTO struct {
	Sessions []queries.SessionRecord `json:"sessions"`
}

type SessionStopOutputDTO struct {
	Stopped bool `json:"stopped"`
}

type SessionCheckOutputDTO struct {
	JobName string `json:"job_name"`
}

type SessionOpenOutputDTO struct {
	URL string `json:"url"`
}

type SessionEnsureRequest struct {
	JSONRPC string                `json:"jsonrpc" default:"2.0" validate:"required"`
	Method  string                `json:"method" default:"session.ensure" validate:"required"`
	Params  SessionEnsureInputDTO `json:"params,omitempty"`
	ID      string                `json:"id,omitempty" default:"1" validate:"required"`
}

type SessionGetRequest struct {
	JSONRPC string             `json:"jsonrpc" default:"2.0" validate:"required"`
	Method  string             `json:"method" default:"session.get" validate:"required"`
	Params  SessionGetInputDTO `json:"params,omitempty"`
	ID      string             `json:"id,omitempty" default:"1" validate:"required"`
}

type SessionListRequest struct {
	JSONRPC string              `json:"jsonrpc" default:"2.0" validate:"required"`
	Method  string              `json:"method" default:"session.list" validate:"required"`
	Params  SessionListInputDTO `json:"params,omitempty"`
	ID      string              `json:"id,omitempty" default:"1" validate:"required"`
}

type SessionStopRequest struct {
	JSONRPC string              `json:"jsonrpc" default:"2.0" validate:"required"`
	Method  string              `json:"method" default:"session.stop" validate:"required"`
	Params  SessionStopInputDTO `json:"params,omitempty"`
	ID      string              `json:"id,omitempty" default:"1" validate:"required"`
}

type SessionCheckRequest struct {
	JSONRPC string               `json:"jsonrpc" default:"2.0" validate:"required"`
	Method  string               `json:"method" default:"session.check" validate:"required"`
	Params  SessionCheckInputDTO `json:"params,omitempty"`
	ID      string               `json:"id,omitempty" default:"1" validate:"required"`
}

func (u *SessionsUC) SetContext(user *cms_client.SSOTokenPublicData) *SessionsUC {
	u.user = user
	return u
}

func (u *SessionsUC) Ensure(ctx context.Context, dto SessionEnsureInputDTO) (SessionOutputDTO, error) {
	if u.user == nil {
		return SessionOutputDTO{}, errors.New("not logged in")
	}
	if _, err := uuid.Parse(dto.AttemptID); err != nil {
		return SessionOutputDTO{}, jsonrpc.NewRpcError("invalid_attempt_id", "attempt_id must be UUID")
	}
	attempts, err := u.CMSClient.ListAttempts(cms_client.ListAttemptsParams{
		AttemptIDs: []string{dto.AttemptID},
		UserIDs:    sessionAttemptUserFilter(u.user),
		Statuses:   []string{"pending", "active"},
		Limit:      1,
	})
	if err != nil {
		return SessionOutputDTO{}, jsonrpc.NewRpcError("attempt_lookup_failed", err.Error())
	}
	if len(attempts) != 1 || attempts[0].AttemptID != dto.AttemptID {
		return SessionOutputDTO{}, jsonrpc.NewRpcError("attempt_not_found", "active attempt was not found for current user")
	}
	ownerID, username := sessionAttemptOwner(u.user, attempts[0])
	record, err := u.ensureAttemptSession(ctx, attempts[0], ownerID, username)
	if err != nil {
		u.Logger.Error().Err(err).Str("attempt_id", dto.AttemptID).Msg("ensure session")
		return SessionOutputDTO{}, jsonrpc.NewRpcError("session_provision_failed", err.Error())
	}
	return SessionOutputDTO{Session: record}, nil
}

func sessionAttemptUserFilter(user *cms_client.SSOTokenPublicData) []int64 {
	if isOperator(user) {
		return nil
	}
	return []int64{int64(user.UserID())}
}

func sessionAttemptOwner(user *cms_client.SSOTokenPublicData, attempt cms_client.ListAttemptsModel) (string, string) {
	if isOperator(user) {
		return strconv.FormatInt(attempt.UserID, 10), attempt.UserName
	}
	return user.Sub, user.Username
}

func (u *SessionsUC) ensureAttemptSession(
	ctx context.Context,
	attempt cms_client.ListAttemptsModel,
	ownerID, username string,
) (queries.SessionRecord, error) {
	if username == "" {
		username = "user-" + ownerID
	}
	pinnedRevision := ""
	existing, existingErr := u.KubernetesAdminQuery.GetSession(ctx, attempt.AttemptID, configs.AppConfig.Session.WorkspacePrefix)
	if existingErr == nil {
		if existing.OwnerID != ownerID {
			return queries.SessionRecord{}, fmt.Errorf("session belongs to another user")
		}
		if existing.Phase == queries.SessionPhaseReady {
			return existing, nil
		}
		pinnedRevision = existing.TaskRevision
	} else if !apierrors.IsNotFound(existingErr) {
		return queries.SessionRecord{}, fmt.Errorf("lookup existing session: %w", existingErr)
	}
	// The routing carries the whole Git link in `repository`, with an optional
	// `#ref` fragment. CMS_TASK_URL and CMS_TASK_BRANCH are the fallback for
	// routes that do not pin a repository of their own.
	repository, repositoryRef := queries.ParseRepositoryLink(attempt.Repository)
	repositoryURL := repository
	if repositoryURL == "" {
		repositoryURL = configs.AppConfig.Session.TaskRepositoryURL
	}
	if repositoryRef == "" {
		repositoryRef = configs.AppConfig.Session.TaskBranch
	}
	bundle, err := u.LabCatalog.BundleFor(ctx, repositoryURL, repositoryRef, "", pinnedRevision)
	if err != nil {
		return queries.SessionRecord{}, fmt.Errorf("download task: %w", err)
	}
	record, err := u.KubernetesAdminQuery.EnsureSession(ctx, queries.EnsureSessionParams{
		AttemptID:       attempt.AttemptID,
		OwnerID:         ownerID,
		Username:        username,
		Title:           attempt.RoutingName,
		TestPath:        attempt.TestPath,
		TopologyYAML:    bundle.Manifest,
		JupyterImage:    configs.AppConfig.Session.JupyterImage,
		StorageSize:     configs.AppConfig.Session.JupyterStorage,
		WorkspacePrefix: configs.AppConfig.Session.WorkspacePrefix,
		TaskRepository:  repositoryURL,
		TaskRef:         repositoryRef,
		TaskRevision:    bundle.Revision,
	})
	if err != nil {
		return queries.SessionRecord{}, err
	}
	return record, nil
}

func (u *SessionsUC) Get(ctx context.Context, dto SessionGetInputDTO) (SessionOutputDTO, error) {
	if u.user == nil {
		return SessionOutputDTO{}, errors.New("not logged in")
	}
	record, err := u.KubernetesAdminQuery.GetSession(ctx, dto.SessionID, configs.AppConfig.Session.WorkspacePrefix)
	if err != nil {
		return SessionOutputDTO{}, jsonrpc.NewRpcError("session_not_found", err.Error())
	}
	if !isOperator(u.user) && record.OwnerID != u.user.Sub {
		return SessionOutputDTO{}, jsonrpc.NewRpcError("session_forbidden", "session belongs to another user")
	}
	return SessionOutputDTO{Session: record}, nil
}

func (u *SessionsUC) List(ctx context.Context, _ SessionListInputDTO) (SessionListOutputDTO, error) {
	if u.user == nil {
		return SessionListOutputDTO{}, errors.New("not logged in")
	}
	ownerID := u.user.Sub
	if isOperator(u.user) {
		ownerID = ""
	}
	records, err := u.KubernetesAdminQuery.ListSessions(ctx, ownerID, configs.AppConfig.Session.WorkspacePrefix)
	if err != nil {
		return SessionListOutputDTO{}, jsonrpc.NewRpcError("session_list_failed", err.Error())
	}
	return SessionListOutputDTO{Sessions: records}, nil
}

func (u *SessionsUC) Stop(ctx context.Context, dto SessionStopInputDTO) (SessionStopOutputDTO, error) {
	if u.user == nil {
		return SessionStopOutputDTO{}, errors.New("not logged in")
	}
	record, err := u.KubernetesAdminQuery.GetSession(ctx, dto.SessionID, configs.AppConfig.Session.WorkspacePrefix)
	if err != nil {
		return SessionStopOutputDTO{}, jsonrpc.NewRpcError("session_not_found", err.Error())
	}
	if !isOperator(u.user) && record.OwnerID != u.user.Sub {
		return SessionStopOutputDTO{}, jsonrpc.NewRpcError("session_forbidden", "session belongs to another user")
	}
	count, err := u.CMSClient.UpdateAttempts([]cms_client.UpdateAttemptParams{{AttemptID: record.AttemptID, Status: "terminating"}})
	if err != nil || count != 1 {
		return SessionStopOutputDTO{}, jsonrpc.NewRpcError("session_stop_failed", "CMS did not accept the terminating state")
	}
	_ = u.KubernetesAdminQuery.MarkSessionPhase(ctx, record.Namespace, queries.SessionPhaseStopping)
	if err := u.KubernetesAdminQuery.DeleteSession(ctx, dto.SessionID, u.user.Sub, isOperator(u.user)); err != nil {
		return SessionStopOutputDTO{}, jsonrpc.NewRpcError("session_stop_failed", fmt.Sprintf("stop session: %v", err))
	}
	return SessionStopOutputDTO{Stopped: true}, nil
}

func (u *SessionsUC) Check(ctx context.Context, dto SessionCheckInputDTO) (SessionCheckOutputDTO, error) {
	if u.user == nil {
		return SessionCheckOutputDTO{}, errors.New("not logged in")
	}
	record, err := u.KubernetesAdminQuery.GetSession(ctx, dto.SessionID, configs.AppConfig.Session.WorkspacePrefix)
	if err != nil {
		return SessionCheckOutputDTO{}, jsonrpc.NewRpcError("session_not_found", err.Error())
	}
	if !isOperator(u.user) && record.OwnerID != u.user.Sub {
		return SessionCheckOutputDTO{}, jsonrpc.NewRpcError("session_forbidden", "session belongs to another user")
	}
	if !record.TopologyReady || !record.WorkspaceReady {
		return SessionCheckOutputDTO{}, jsonrpc.NewRpcError("session_not_ready", "topology and workspace must be ready before checking")
	}
	jobName, err := u.KubernetesAdminQuery.RunChecker(ctx, queries.RunCheckerParams{
		SessionID:      record.ID,
		Namespace:      record.Namespace,
		AttemptID:      record.AttemptID,
		OwnerID:        record.OwnerID,
		TestPath:       record.TestPath,
		Image:          configs.AppConfig.Session.CheckerImage,
		TimeoutSeconds: configs.AppConfig.Session.CheckerTimeout,
	})
	if err != nil {
		return SessionCheckOutputDTO{}, jsonrpc.NewRpcError("checker_start_failed", err.Error())
	}
	return SessionCheckOutputDTO{JobName: jobName}, nil
}

func (u *SessionsUC) Open(ctx context.Context, dto SessionOpenInputDTO) (SessionOpenOutputDTO, error) {
	if u.user == nil {
		return SessionOpenOutputDTO{}, errors.New("not logged in")
	}
	record, err := u.KubernetesAdminQuery.GetSession(ctx, dto.SessionID, configs.AppConfig.Session.WorkspacePrefix)
	if err != nil {
		return SessionOpenOutputDTO{}, jsonrpc.NewRpcError("session_not_found", err.Error())
	}
	if !isOperator(u.user) && record.OwnerID != u.user.Sub {
		return SessionOpenOutputDTO{}, jsonrpc.NewRpcError("session_forbidden", "session belongs to another user")
	}
	if !record.WorkspaceReady {
		return SessionOpenOutputDTO{}, jsonrpc.NewRpcError("session_not_ready", "JupyterLab is not ready")
	}
	grant, err := IssueWorkspaceGrant(
		configs.AppConfig.Session.WorkspaceSecret,
		record.ID,
		record.WorkspaceURL,
		NewWorkspaceIdentity(u.user.Sub, u.user.Username, u.user.Name, u.user.Email, u.user.Roles),
		time.Duration(configs.AppConfig.Session.WorkspaceGrantTTL)*time.Second,
	)
	if err != nil {
		return SessionOpenOutputDTO{}, jsonrpc.NewRpcError("workspace_auth_unavailable", err.Error())
	}
	return SessionOpenOutputDTO{URL: "/clabgate/workspace-auth/exchange?grant=" + url.QueryEscape(grant)}, nil
}

func isOperator(user *cms_client.SSOTokenPublicData) bool {
	return cms_client.SSOHasIntersection(
		[]string{cms_client.SSOUsersRoleAdmin, cms_client.SSOUsersRoleInstructor},
		user.Roles,
	)
}
