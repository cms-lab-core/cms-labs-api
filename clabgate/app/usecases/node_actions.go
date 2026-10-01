package usecases

import (
	"context"
	"errors"
	"fmt"

	"github.com/cms-lab-core/cms-labs-api/clabgate/app/queries"
	"github.com/cms-lab-core/cms-labs-api/shared/cms_client"
	"github.com/rs/zerolog"
)

type NodeActionsUC struct {
	*zerolog.Logger
	KubernetesAdminQuery *queries.KubernetesAdminQuery
	user                 *cms_client.SSOTokenPublicData
}

type NodeActionItem struct {
	Node string `json:"node"`
	// enum: wipe,restart
	Action string `json:"action"`
}

type NodeActionInputDTO struct {
	Actions   []NodeActionItem `json:"actions"`
	SessionID string           `json:"session_id" validate:"required"`
}

type NodeActionRequest struct {
	JSONRPC string             `json:"jsonrpc" default:"2.0" required:"true"`
	Method  string             `json:"method" default:"nodes.action" required:"true"`
	Params  NodeActionInputDTO `json:"params,omitempty"`
	ID      string             `json:"id,omitempty" default:"1" required:"true"`
}

type NodeActionOutputDTO struct {
	Count int `json:"count"`
}

type NodeActionResponse struct {
	JSONRPC string              `json:"jsonrpc" default:"2.0" required:"true"`
	Result  NodeActionOutputDTO `json:"result,omitempty"`
	Error   interface{}         `json:"error,omitempty"`
	ID      string              `json:"id,omitempty" default:"1" required:"true"`
}

func (u *NodeActionsUC) SetContext(user *cms_client.SSOTokenPublicData) *NodeActionsUC {
	u.user = user
	return u
}

func (u *NodeActionsUC) Execute(dto NodeActionInputDTO) (NodeActionOutputDTO, error) {
	if u.user == nil {
		return NodeActionOutputDTO{}, errors.New("not logged in")
	}
	if dto.SessionID == "" {
		return NodeActionOutputDTO{}, errors.New("session_id is required")
	}
	ctx := context.Background()
	u.Logger.Info().Str("session_id", dto.SessionID).Msg("NodeActionsUC: execute session action")
	session, err := u.KubernetesAdminQuery.GetSession(ctx, dto.SessionID, "")
	if err != nil {
		return NodeActionOutputDTO{}, err
	}
	if !isOperator(u.user) && session.OwnerID != u.user.Sub {
		return NodeActionOutputDTO{}, errors.New("session belongs to another user")
	}
	namespace := session.Namespace

	deployments, _ := u.KubernetesAdminQuery.GetDeploymentsInfo(ctx, namespace)

	var count int
	for _, action := range dto.Actions {
		pod := u.GetPodByNode(&deployments, action.Node)
		switch action.Action {
		case "restart":
			if err := u.KubernetesAdminQuery.RestartPod(ctx, namespace, pod, action.Node); err != nil {
				u.Logger.Error().Err(err).Msg(
					fmt.Sprintf(
						"failed to restart pod %s, container %s with namespace %s",
						pod,
						action.Node,
						namespace,
					),
				)
			} else {
				count++
			}
		case "wipe":
			if err := u.KubernetesAdminQuery.DeletePod(ctx, namespace, pod); err != nil {
				u.Logger.Error().Err(err).Msg(
					fmt.Sprintf("failed to wipe pod %s with namespace %s", pod, namespace),
				)
			} else {
				count++
			}
		default:
			u.Logger.Warn().Msg(fmt.Sprintf("unknown action: %s", action.Action))
		}
	}

	return NodeActionOutputDTO{
		Count: count,
	}, nil
}

func (u *NodeActionsUC) GetPodByNode(deployments *[]queries.DeploymentInfo, node string) string {
	if deployments == nil {
		return ""
	}
	for _, deployment := range *deployments {
		if deployment.Name == node {
			return deployment.PodNameLast
		}
	}
	return ""
}
