package controllers

import (
	"github.com/cms-lab-core/cms-labs-api/backend/app/di"
	"github.com/cms-lab-core/cms-labs-api/backend/app/usecases"
	"github.com/cms-lab-core/cms-labs-api/backend/app/usecases/auth"
	"github.com/cms-lab-core/cms-labs-api/shared/cms_client"
	"github.com/cms-lab-core/cms-labs-api/shared/jsonrpc"
	"github.com/cms-lab-core/cms-labs-api/shared/logs"
)

// UserUpsert func for creates a new User.
// @Description Create user. Roles[admin, instructor]
// @Summary create user
// @Tags user
// @Accept json
// @Produce json
// @Param object body usecases.UserEditRequest true "user form info"
// @Success 200 {object} usecases.UserEditResponse
// @Security ApiKeyAuth
// @Router /api/v1/rpc/user.upsert [post]
func UserUpsert(c *jsonrpc.Ctx) (interface{}, error) {
	diLoggerConf := logs.NewZeroLoggerConf(c)
	if _, err := auth.ExtractTokenMetadata(
		c,
		[]string{cms_client.SSOUsersRoleAdmin, cms_client.SSOUsersRoleInstructor},
	); err != nil {
		return nil, err
	}
	dto := usecases.UserEditInputDTO{}
	err := jsonrpc.ValidatorBase(c, &dto)
	if err != nil {
		return nil, err
	}
	container, err := di.NewDIContainer(diLoggerConf)
	if err != nil {
		return nil, err
	}
	defer container.Close()
	uc := container.UserEditUC()
	output, err := uc.Execute(dto)
	return output, err
}

// UserList func for view of list User.
// @Description List user. Roles[admin, instructor]
// @Summary list user
// @Tags user
// @Accept json
// @Produce json
// @Param object body usecases.UserListRequest true "user list info"
// @Success 200 {object} usecases.UserListResponse
// @Security ApiKeyAuth
// @Router /api/v1/rpc/user.list [post]
func UserList(c *jsonrpc.Ctx) (interface{}, error) {
	diLoggerConf := logs.NewZeroLoggerConf(c)
	if _, err := auth.ExtractTokenMetadata(
		c,
		[]string{cms_client.SSOUsersRoleAdmin, cms_client.SSOUsersRoleInstructor},
	); err != nil {
		return nil, err
	}
	dto := usecases.UserListInputDTO{}
	err := jsonrpc.ValidatorBase(c, &dto)
	if err != nil {
		return nil, err
	}
	container, err := di.NewDIContainer(diLoggerConf)
	if err != nil {
		return nil, err
	}
	defer container.Close()
	uc := container.UserListUC()
	output, err := uc.Execute(dto)
	return output, err
}

// UserGet func for full model User.
// @Description get user. Roles[admin, instructor]
// @Summary get user
// @Tags user
// @Accept json
// @Produce json
// @Param object body usecases.UserGetRequest true "user id"
// @Success 200 {object} usecases.UserGetResponse
// @Security ApiKeyAuth
// @Router /api/v1/rpc/user.get [post]
func UserGet(c *jsonrpc.Ctx) (interface{}, error) {
	diLoggerConf := logs.NewZeroLoggerConf(c)
	if _, err := auth.ExtractTokenMetadata(
		c,
		[]string{cms_client.SSOUsersRoleAdmin, cms_client.SSOUsersRoleInstructor},
	); err != nil {
		return nil, err
	}
	dto := usecases.UserGetInputDTO{}
	err := jsonrpc.ValidatorBase(c, &dto)
	if err != nil {
		return nil, err
	}
	container, err := di.NewDIContainer(diLoggerConf)
	if err != nil {
		return nil, err
	}
	defer container.Close()
	uc := container.UserGetUC()
	output, err := uc.Execute(dto)
	return output, err
}
