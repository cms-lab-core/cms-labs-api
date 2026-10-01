package controllers

import (
	"github.com/cms-lab-core/cms-labs-api/backend/app/di"
	"github.com/cms-lab-core/cms-labs-api/backend/app/queries"
	"github.com/cms-lab-core/cms-labs-api/backend/app/usecases"
	"github.com/cms-lab-core/cms-labs-api/backend/app/usecases/auth"
	"github.com/cms-lab-core/cms-labs-api/shared/cms_client"
	"github.com/cms-lab-core/cms-labs-api/shared/jsonrpc"
	"github.com/cms-lab-core/cms-labs-api/shared/logs"
)

// ServerUpsert func for creates a new Server.
// @Description Create server. Roles [admin]
// @Summary create server
// @Tags server
// @Accept json
// @Produce json
// @Param object body usecases.ServerEditRequest true "server form info"
// @Success 200 {object} usecases.ServerEditResponse
// @Security ApiKeyAuth
// @Router /api/v1/rpc/server.upsert [post]
func ServerUpsert(c *jsonrpc.Ctx) (interface{}, error) {
	diLoggerConf := logs.NewZeroLoggerConf(c)
	if _, err := auth.ExtractTokenMetadata(
		c,
		[]string{cms_client.SSOUsersRoleAdmin},
	); err != nil {
		return nil, err
	}
	dto := usecases.ServerEditInputDTO{}
	err := jsonrpc.ValidatorBase(c, &dto)
	if err != nil {
		return nil, err
	}
	container, err := di.NewDIContainer(diLoggerConf)
	if err != nil {
		return nil, err
	}
	defer container.Close()
	uc := container.ServerEditUC()
	output, err := uc.Execute(dto)
	if err != nil {
		return nil, err
	}
	ucDistribution := container.ServerChangeDistributionUC()
	err = ucDistribution.Execute()
	if err != nil {
		return nil, err
	}
	return output, nil
}

// ServerList func for view of list Server.
// @Description List server. Roles: [admin, instructor]
// @Summary list server
// @Tags server
// @Accept json
// @Produce json
// @Param object body usecases.ServerListRequest true "server list info"
// @Success 200 {object} usecases.ServerListResponse
// @Security ApiKeyAuth
// @Router /api/v1/rpc/server.list [post]
func ServerList(c *jsonrpc.Ctx) (interface{}, error) {
	diLoggerConf := logs.NewZeroLoggerConf(c)
	if _, err := auth.ExtractTokenMetadata(
		c,
		[]string{cms_client.SSOUsersRoleAdmin, cms_client.SSOUsersRoleInstructor},
	); err != nil {
		return nil, err
	}
	dto := queries.ServerQueriesListDTO{}
	err := jsonrpc.ValidatorBase(c, &dto)
	if err != nil {
		return nil, err
	}
	container, err := di.NewDIContainer(diLoggerConf)
	if err != nil {
		return nil, err
	}
	defer container.Close()
	uc := container.ServerListUC()
	output, err := uc.Execute(dto)
	return output, err
}

// ServerDelete func for delete Server.
// @Description Delete server. Roles: [admin]
// @Summary delete server
// @Tags server
// @Accept json
// @Produce json
// @Param object body usecases.ServerDeleteRequest true "server id"
// @Success 200 {object} usecases.ServerDeleteResponse
// @Security ApiKeyAuth
// @Router /api/v1/rpc/server.delete [post]
func ServerDelete(c *jsonrpc.Ctx) (interface{}, error) {
	diLoggerConf := logs.NewZeroLoggerConf(c)
	if _, err := auth.ExtractTokenMetadata(
		c,
		[]string{cms_client.SSOUsersRoleAdmin},
	); err != nil {
		return nil, err
	}
	dto := usecases.ServerDeleteInputDTO{}
	err := jsonrpc.ValidatorBase(c, &dto)
	if err != nil {
		return nil, err
	}
	container, err := di.NewDIContainer(diLoggerConf)
	if err != nil {
		return nil, err
	}
	defer container.Close()
	uc := container.ServerDeleteUC()
	output, err := uc.Execute(dto)
	if err != nil {
		return nil, err
	}
	ucDistribution := container.ServerChangeDistributionUC()
	err = ucDistribution.Execute()
	if err != nil {
		return nil, err
	}
	return output, nil
}

// ServerGet func for full model Server.
// @Description get server. Roles: [admin, instructor]
// @Summary get server
// @Tags server
// @Accept json
// @Produce json
// @Param object body usecases.ServerGetRequest true "server id"
// @Success 200 {object} usecases.ServerGetResponse
// @Security ApiKeyAuth
// @Router /api/v1/rpc/server.get [post]
func ServerGet(c *jsonrpc.Ctx) (interface{}, error) {
	diLoggerConf := logs.NewZeroLoggerConf(c)
	if _, err := auth.ExtractTokenMetadata(
		c,
		[]string{cms_client.SSOUsersRoleAdmin, cms_client.SSOUsersRoleInstructor},
	); err != nil {
		return nil, err
	}
	dto := usecases.ServerGetInputDTO{}
	err := jsonrpc.ValidatorBase(c, &dto)
	if err != nil {
		return nil, err
	}
	container, err := di.NewDIContainer(diLoggerConf)
	if err != nil {
		return nil, err
	}
	defer container.Close()
	uc := container.ServerGetUC()
	output, err := uc.Execute(dto)
	return output, err
}
