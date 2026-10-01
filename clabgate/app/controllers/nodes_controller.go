package controllers

import (
	"github.com/cms-lab-core/cms-labs-api/clabgate/app/di"
	"github.com/cms-lab-core/cms-labs-api/clabgate/app/usecases"
	"github.com/cms-lab-core/cms-labs-api/clabgate/app/usecases/auth"
	"github.com/cms-lab-core/cms-labs-api/shared/jsonrpc"
	"github.com/cms-lab-core/cms-labs-api/shared/logs"
)

// NodeAction func for restart/wipe pods.
// @Description Restart or wipe pods. Roles: [admin, instructor, student]
// @Summary node action
// @Tags Node
// @Accept json
// @Produce json
// @Param form body usecases.NodeActionRequest true "node action"
// @Success 200 {object} usecases.NodeActionResponse
// @Security ApiKeyAuth
// @Router /clabgate/api/v1/rpc/node.action [post]
func NodeAction(c *jsonrpc.Ctx) (interface{}, error) {
	user, err := auth.ExtractTokenMetadata(c, []string{})
	if err != nil {
		return nil, err
	}
	diLoggerConf := logs.NewZeroLoggerConf(c)
	dto := usecases.NodeActionInputDTO{}
	err = jsonrpc.ValidatorBase(c, &dto)
	if err != nil {
		return nil, err
	}
	container, err := di.NewDIContainer(diLoggerConf)
	if err != nil {
		return nil, err
	}
	defer container.Close()
	uc, err := container.NodeActionsUC()
	if err != nil {
		return nil, err
	}
	output, err := uc.SetContext(user).Execute(dto)
	return output, err
}
