package controllers

import (
	"github.com/cms-lab-core/cms-labs-api/backend/app/di"
	"github.com/cms-lab-core/cms-labs-api/backend/app/usecases"
	"github.com/cms-lab-core/cms-labs-api/backend/app/usecases/auth"
	"github.com/cms-lab-core/cms-labs-api/shared/cms_client"
	"github.com/cms-lab-core/cms-labs-api/shared/jsonrpc"
	"github.com/cms-lab-core/cms-labs-api/shared/logs"
)

// RoleUpsert func for create and update a few Role.
// @Description Upset role. Roles[admin]
// @Summary upsert role
// @Tags role
// @Accept json
// @Produce json
// @Param object body usecases.RoleEditRequest true "user form info"
// @Success 200 {object} usecases.RoleEditResponse
// @Security ApiKeyAuth
// @Router /api/v1/rpc/role.upsert [post]
func RoleUpsert(c *jsonrpc.Ctx) (interface{}, error) {
	diLoggerConf := logs.NewZeroLoggerConf(c)
	if _, err := auth.ExtractTokenMetadata(
		c,
		[]string{cms_client.SSOUsersRoleAdmin},
	); err != nil {
		return nil, err
	}
	dto := usecases.RoleEditInputDTO{}
	err := jsonrpc.ValidatorBase(c, &dto)
	if err != nil {
		return nil, err
	}
	container, err := di.NewDIContainer(diLoggerConf)
	if err != nil {
		return nil, err
	}
	defer container.Close()
	uc := container.RoleEditUC()
	output, err := uc.Execute(dto)
	return output, err
}

// RoleList func for view of list Role.
// @Description List user. Roles[admin, instructor, student, any]
// @Summary list role
// @Tags role
// @Accept json
// @Produce json
// @Param object body usecases.RoleListRequest true "user list info"
// @Success 200 {object} usecases.RoleListResponse
// @Router /api/v1/rpc/role.list [post]
func RoleList(c *jsonrpc.Ctx) (interface{}, error) {
	diLoggerConf := logs.NewZeroLoggerConf(c)
	dto := usecases.RoleListInputDTO{}
	err := jsonrpc.ValidatorBase(c, &dto)
	if err != nil {
		return nil, err
	}
	container, err := di.NewDIContainer(diLoggerConf)
	if err != nil {
		return nil, err
	}
	defer container.Close()
	uc := container.RoleListUC()
	output, err := uc.Execute(dto)
	return output, err
}

// RoleDelete func for delete Role.
// @Description Delete role. Roles: [admin]
// @Summary delete role
// @Tags role
// @Accept json
// @Produce json
// @Param object body usecases.RoleDeleteRequest true "pnet_server id"
// @Success 200 {object} usecases.RoleDeleteResponse
// @Security ApiKeyAuth
// @Router /api/v1/rpc/role.delete [post]
func RoleDelete(c *jsonrpc.Ctx) (interface{}, error) {
	diLoggerConf := logs.NewZeroLoggerConf(c)
	if _, err := auth.ExtractTokenMetadata(
		c,
		[]string{cms_client.SSOUsersRoleAdmin},
	); err != nil {
		return nil, err
	}
	dto := usecases.RoleDeleteInputDTO{}
	err := jsonrpc.ValidatorBase(c, &dto)
	if err != nil {
		return nil, err
	}
	container, err := di.NewDIContainer(diLoggerConf)
	if err != nil {
		return nil, err
	}
	defer container.Close()
	uc := container.RoleDeleteUC()
	output, err := uc.Execute(dto)
	return output, err
}
