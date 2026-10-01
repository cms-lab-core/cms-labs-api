package usecases

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/cms-lab-core/cms-labs-api/clabgate/app/queries"
	"github.com/cms-lab-core/cms-labs-api/shared/cms_client"
	"github.com/cms-lab-core/cms-labs-api/shared/jsonrpc"
	"github.com/rs/zerolog"
	"gopkg.in/yaml.v3"
)

type TopologiesGetUC struct {
	*zerolog.Logger
	KubernetesAdminQuery *queries.KubernetesAdminQuery
	WorkspaceSecret      string
	WorkspaceGrantTTL    time.Duration
	user                 *cms_client.SSOTokenPublicData
}

type TopologiesGetInputDTO struct {
	SessionID string `json:"session_id" validate:"required"`
}

type TopologiesGetRequest struct {
	JSONRPC string                `json:"jsonrpc" default:"2.0" required:"true"`
	Method  string                `json:"method" default:"topology.get" required:"true"`
	Params  TopologiesGetInputDTO `json:"params,omitempty"`
	ID      string                `json:"id,omitempty" default:"1" required:"true"`
}

type TopologiesGetOutputDTO struct {
	Topology    string                   `json:"topology"`
	Deployments []queries.DeploymentInfo `json:"deployments,omitempty"`
	Services    []queries.ServiceInfo    `json:"services,omitempty"`
	TTYD        []queries.TTYDInfo       `json:"ttyd,omitempty"`
}

type TopologiesGetResponse struct {
	JSONRPC string                 `json:"jsonrpc" default:"2.0" required:"true"`
	Result  TopologiesGetOutputDTO `json:"result,omitempty"`
	Error   interface{}            `json:"error,omitempty"`
	ID      string                 `json:"id,omitempty" default:"1" required:"true"`
}

func (u *TopologiesGetUC) SetContext(user *cms_client.SSOTokenPublicData) *TopologiesGetUC {
	u.user = user
	return u
}

func (u *TopologiesGetUC) Execute(dto TopologiesGetInputDTO) (TopologiesGetOutputDTO, error) {
	if u.user == nil {
		return TopologiesGetOutputDTO{}, errors.New("not logged in")
	}
	ctx := context.Background()
	if dto.SessionID == "" {
		return TopologiesGetOutputDTO{}, jsonrpc.NewRpcError("session_required", "session_id is required")
	}
	u.Logger.Info().Str("session_id", dto.SessionID).Msg("TopologiesGetUC: read session topology")
	session, err := u.KubernetesAdminQuery.GetSession(ctx, dto.SessionID, "")
	if err != nil {
		return TopologiesGetOutputDTO{}, jsonrpc.NewRpcError("session_not_found", err.Error())
	}
	if !isOperator(u.user) && session.OwnerID != u.user.Sub {
		return TopologiesGetOutputDTO{}, jsonrpc.NewRpcError("session_forbidden", "session belongs to another user")
	}
	namespace := session.Namespace
	yamlContent, err := u.KubernetesAdminQuery.GetTopologyYAML(ctx, namespace)

	var containerlabContent string
	if err == nil && len(yamlContent) > 0 {
		var topoData map[string]interface{}
		if err := yaml.Unmarshal(yamlContent, &topoData); err == nil {
			if spec, ok := topoData["spec"].(map[string]interface{}); ok {
				if definition, ok := spec["definition"].(map[string]interface{}); ok {
					if clab, ok := definition["containerlab"].(string); ok {
						containerlabContent = clab
					}
				}
			}
		}
	} else {
		u.Logger.Error().Msg(fmt.Sprintf("TopologiesGetUC: GetTopologyYAML error: %v", err))
	}

	u.Logger.Debug().Msg(fmt.Sprintf("TopologiesGetUC: containerlab content length: %d", len(containerlabContent)))

	deployments, _ := u.KubernetesAdminQuery.GetDeploymentsInfo(ctx, namespace)
	services, _ := u.KubernetesAdminQuery.GetServicesInfo(ctx, namespace)
	ttyd, _ := u.KubernetesAdminQuery.GetTTYDInfoInNamespace(ctx, namespace)
	for index := range ttyd {
		destination := fmt.Sprintf("/clabgate/workspace/%s/terminal/%s/", dto.SessionID, ttyd[index].Name)
		grant, grantErr := IssueWorkspaceGrant(
			u.WorkspaceSecret,
			dto.SessionID,
			destination,
			NewWorkspaceIdentity(u.user.Sub, u.user.Username, u.user.Name, u.user.Email, u.user.Roles),
			u.WorkspaceGrantTTL,
		)
		if grantErr != nil {
			return TopologiesGetOutputDTO{}, jsonrpc.NewRpcError("workspace_auth_unavailable", grantErr.Error())
		}
		ttyd[index].Url = "/clabgate/workspace-auth/exchange?grant=" + url.QueryEscape(grant)
	}

	return TopologiesGetOutputDTO{
		Topology:    containerlabContent,
		Deployments: deployments,
		Services:    services,
		TTYD:        ttyd,
	}, nil
}
