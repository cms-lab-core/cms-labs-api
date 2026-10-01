package controllers

import (
	"github.com/cms-lab-core/cms-labs-api/backend/app/di"
	"github.com/cms-lab-core/cms-labs-api/backend/app/queries/lti_query"
	"github.com/cms-lab-core/cms-labs-api/backend/app/usecases"
	"github.com/cms-lab-core/cms-labs-api/backend/app/usecases/auth"
	"github.com/cms-lab-core/cms-labs-api/shared/cms_client"
	"github.com/cms-lab-core/cms-labs-api/shared/jsonrpc"
	"github.com/cms-lab-core/cms-labs-api/shared/logs"
)

// AuthProviderUpsert func for creates a new AuthProvider.
// @Description Create lti_form. Roles: [admin]
// @Summary create lti_form
// @Tags lti_form
// @Accept json
// @Produce json
// @Param object body usecases.AuthProviderEditRequest true "lti_form form info"
// @Success 200 {object} usecases.AuthProviderEditResponse
// @Security ApiKeyAuth
// @Router /api/v1/rpc/auth_provider.upsert [post]
func AuthProviderUpsert(c *jsonrpc.Ctx) (interface{}, error) {
	diLoggerConf := logs.NewZeroLoggerConf(c)
	if _, err := auth.ExtractTokenMetadata(
		c,
		[]string{cms_client.SSOUsersRoleAdmin},
	); err != nil {
		return nil, err
	}
	dto := usecases.AuthProviderEditInputDTO{}
	err := jsonrpc.ValidatorBase(c, &dto)
	if err != nil {
		return nil, err
	}
	container, err := di.NewDIContainer(diLoggerConf)
	if err != nil {
		return nil, err
	}
	uc := container.AuthProviderEditUC()
	output, err := uc.Execute(dto)
	return output, err
}

// AuthProviderList func for view of list AuthProvider.
// @Description List lti_form. Roles: [*]
// @Summary list lti_form
// @Tags lti_form
// @Accept json
// @Produce json
// @Param object body usecases.AuthProviderListRequest true "lti_form list info"
// @Success 200 {object} usecases.AuthProviderListResponse
// @Router /api/v1/rpc/auth_provider.list [post]
func AuthProviderList(c *jsonrpc.Ctx) (interface{}, error) {
	diLoggerConf := logs.NewZeroLoggerConf(c)
	dto := lti_query.AuthProviderListInputDTO{}
	err := jsonrpc.ValidatorBase(c, &dto)
	if err != nil {
		return nil, err
	}
	container, err := di.NewDIContainer(diLoggerConf)
	if err != nil {
		return nil, err
	}
	uc := container.AuthProviderListUC()
	output, err := uc.Execute(dto)
	return output, err
}

// AuthProviderDelete func for delete AuthProvider.
// @Description Delete lti_form. Roles: [admin]
// @Summary delete lti_form
// @Tags lti_form
// @Accept json
// @Produce json
// @Param object body usecases.AuthProviderDeleteRequest true "lti_form id"
// @Success 200 {object} usecases.AuthProviderDeleteResponse
// @Security ApiKeyAuth
// @Router /api/v1/rpc/auth_provider.delete [post]
func AuthProviderDelete(c *jsonrpc.Ctx) (interface{}, error) {
	diLoggerConf := logs.NewZeroLoggerConf(c)
	if _, err := auth.ExtractTokenMetadata(
		c,
		[]string{cms_client.SSOUsersRoleAdmin},
	); err != nil {
		return nil, err
	}
	dto := usecases.AuthProviderDeleteInputDTO{}
	err := jsonrpc.ValidatorBase(c, &dto)
	if err != nil {
		return nil, err
	}
	container, err := di.NewDIContainer(diLoggerConf)
	if err != nil {
		return nil, err
	}
	defer container.Close()
	uc := container.AuthProviderDeleteUC()
	output, err := uc.Execute(dto)
	return output, err
}

// AuthProviderGet func for full model AuthProvider.
// @Description get lti_form. Roles [admin]
// @Summary get lti_form
// @Tags lti_form
// @Accept json
// @Produce json
// @Param object body usecases.AuthProviderGetRequest true "lti_form id"
// @Success 200 {object} usecases.AuthProviderGetResponse
// @Security ApiKeyAuth
// @Router /api/v1/rpc/auth_provider.get [post]
func AuthProviderGet(c *jsonrpc.Ctx) (interface{}, error) {
	diLoggerConf := logs.NewZeroLoggerConf(c)
	if _, err := auth.ExtractTokenMetadata(
		c,
		[]string{cms_client.SSOUsersRoleAdmin},
	); err != nil {
		return nil, err
	}
	dto := usecases.AuthProviderGetInputDTO{}
	err := jsonrpc.ValidatorBase(c, &dto)
	if err != nil {
		return nil, err
	}
	container, err := di.NewDIContainer(diLoggerConf)
	if err != nil {
		return nil, err
	}
	defer container.Close()
	uc := container.AuthProviderGetUC()
	output, err := uc.Execute(dto)
	return output, err
}
