package controllers

import (
	"github.com/cms-lab-core/cms-labs-api/clabgate/app/di"
	"github.com/cms-lab-core/cms-labs-api/clabgate/app/usecases"
	"github.com/cms-lab-core/cms-labs-api/clabgate/app/usecases/auth"
	"github.com/cms-lab-core/cms-labs-api/shared/jsonrpc"
	"github.com/cms-lab-core/cms-labs-api/shared/logs"
)

// TopologyGet func for view of list tasks.
// @Description List tasks. Roles: [student, admin, instructor]
// @Summary list tasks
// @Tags Topology
// @Accept json
// @Produce json
// @Param form body usecases.TopologiesGetRequest true "topology namespace"
// @Success 200 {object} usecases.TopologiesGetResponse
// @Security ApiKeyAuth
// @Router /clabgate/api/v1/rpc/topology.get [post]
func TopologyGet(c *jsonrpc.Ctx) (interface{}, error) {
	user, err := auth.ExtractTokenMetadata(c, []string{})
	if err != nil {
		return nil, err
	}
	diLoggerConf := logs.NewZeroLoggerConf(c)
	dto := usecases.TopologiesGetInputDTO{}
	err = jsonrpc.ValidatorBase(c, &dto)
	if err != nil {
		return nil, err
	}
	container, err := di.NewDIContainer(diLoggerConf)
	if err != nil {
		return nil, err
	}
	defer container.Close()
	uc, err := container.TopologiesGetUC()
	if err != nil {
		return nil, err
	}
	output, err := uc.SetContext(user).Execute(dto)
	return output, err
}
