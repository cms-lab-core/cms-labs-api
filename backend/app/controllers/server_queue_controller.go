package controllers

import (
	"github.com/cms-lab-core/cms-labs-api/backend/app/di"
	"github.com/cms-lab-core/cms-labs-api/backend/app/usecases/auth"
	"github.com/cms-lab-core/cms-labs-api/shared/cms_client"
	"github.com/cms-lab-core/cms-labs-api/shared/jsonrpc"
	"github.com/cms-lab-core/cms-labs-api/shared/logs"
)

// ServerQueueUpsert func for update on new queue RoundRobin.
// @Description Create server_queue. Roles: [admin, instructor]
// @Summary create server_queue
// @Tags server_queue
// @Accept json
// @Produce json
// @Param object body server_queue.ServerQueueUpsertRequest true "request"
// @Success 200 {object} server_queue.ServerQueueUpsertResponse
// @Security ApiKeyAuth
// @Router /api/v1/rpc/server_queue.upsert [post]
func ServerQueueUpsert(c *jsonrpc.Ctx) (interface{}, error) {
	diLoggerConf := logs.NewZeroLoggerConf(c)
	if _, err := auth.ExtractTokenMetadata(
		c,
		[]string{cms_client.SSOUsersRoleAdmin, cms_client.SSOUsersRoleInstructor},
	); err != nil {
		return nil, err
	}
	container, err := di.NewDIContainer(diLoggerConf)
	if err != nil {
		return nil, err
	}
	defer container.Close()
	uc := container.ServerChangeDistributionUC()
	err = uc.Execute()
	if err != nil {
		return nil, err
	}
	return true, nil
}

// ServerQueueList func for list distribution queue
// @Description List server_queue. Roles: [admin, instructor]
// @Summary list server_queue
// @Tags server_queue
// @Accept json
// @Produce json
// @Param object body server_queue.ServerQueueListRequest true "request"
// @Success 200 {object} server_queue.ServerQueueListResponse
// @Security ApiKeyAuth
// @Router /api/v1/rpc/server_queue.list [post]
func ServerQueueList(c *jsonrpc.Ctx) (interface{}, error) {
	diLoggerConf := logs.NewZeroLoggerConf(c)
	if _, err := auth.ExtractTokenMetadata(
		c,
		[]string{cms_client.SSOUsersRoleAdmin, cms_client.SSOUsersRoleInstructor},
	); err != nil {
		return nil, err
	}
	container, err := di.NewDIContainer(diLoggerConf)
	if err != nil {
		return nil, err
	}
	defer container.Close()
	uc := container.ServerListDistributionUC()
	output, err := uc.Execute()
	return output, err
}
