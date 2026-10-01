package controllers

import (
	"github.com/cms-lab-core/cms-labs-api/backend/app/di"
	"github.com/cms-lab-core/cms-labs-api/backend/app/usecases"
	"github.com/cms-lab-core/cms-labs-api/backend/app/usecases/auth"
	"github.com/cms-lab-core/cms-labs-api/shared/cms_client"
	"github.com/cms-lab-core/cms-labs-api/shared/jsonrpc"
	"github.com/cms-lab-core/cms-labs-api/shared/logs"
)

// LTIRoutingUpsert func for creates a new LTIRouting.
// @Description Create lti_routing. Roles [admin, instructor]
// @Summary create lti_routing
// @Tags lti_routing
// @Accept json
// @Produce json
// @Param object body usecases.LTIRoutingEditRequest true "lti_routing form info"
// @Success 200 {object} usecases.LTIRoutingEditResponse
// @Security ApiKeyAuth
// @Router /api/v1/rpc/lti_routing.upsert [post]
func LTIRoutingUpsert(c *jsonrpc.Ctx) (interface{}, error) {
	diLoggerConf := logs.NewZeroLoggerConf(c)
	if _, err := auth.ExtractTokenMetadata(
		c,
		[]string{cms_client.SSOUsersRoleAdmin, cms_client.SSOUsersRoleInstructor},
	); err != nil {
		return nil, err
	}
	dto := usecases.LTIRoutingEditInputDTO{}
	err := jsonrpc.ValidatorBase(c, &dto)
	if err != nil {
		return nil, err
	}
	container, err := di.NewDIContainer(diLoggerConf)
	if err != nil {
		return nil, err
	}
	defer container.Close()
	uc := container.LTIRoutingEditUC()
	output, err := uc.Execute(dto)
	return output, err
}

// LTIRoutingList func for view of list LTIRouting.
// @Description List lti_routing. Roles [admin, instructor]
// @Summary list lti_routing
// @Tags lti_routing
// @Accept json
// @Produce json
// @Param form body usecases.LTIRoutingListRequest true "lti_routing list info"
// @Success 200 {object} usecases.LTIRoutingListResponse
// @Security ApiKeyAuth
// @Router /api/v1/rpc/lti_routing.list [post]
func LTIRoutingList(c *jsonrpc.Ctx) (interface{}, error) {
	diLoggerConf := logs.NewZeroLoggerConf(c)
	if _, err := auth.ExtractTokenMetadata(
		c,
		[]string{cms_client.SSOUsersRoleAdmin, cms_client.SSOUsersRoleInstructor},
	); err != nil {
		return nil, err
	}
	dto := usecases.LTIRoutingListInputDTO{}
	err := jsonrpc.ValidatorBase(c, &dto)
	if err != nil {
		return nil, err
	}
	container, err := di.NewDIContainer(diLoggerConf)
	if err != nil {
		return nil, err
	}
	defer container.Close()
	uc := container.LTIRoutingListUC()
	output, err := uc.Execute(dto)
	return output, err
}

// LTIRoutingDelete func for delete LTIRouting.
// @Description Delete lti_routing. Roles [admin, instructor]
// @Summary delete lti_routing
// @Tags lti_routing
// @Accept json
// @Produce json
// @Param object body usecases.LTIRoutingDeleteRequest true "lti_routing id"
// @Success 200 {object} usecases.LTIRoutingDeleteResponse
// @Security ApiKeyAuth
// @Router /api/v1/rpc/lti_routing.delete [post]
func LTIRoutingDelete(c *jsonrpc.Ctx) (interface{}, error) {
	diLoggerConf := logs.NewZeroLoggerConf(c)
	if _, err := auth.ExtractTokenMetadata(
		c,
		[]string{cms_client.SSOUsersRoleAdmin, cms_client.SSOUsersRoleInstructor},
	); err != nil {
		return nil, err
	}
	dto := usecases.LTIRoutingDeleteInputDTO{}
	err := jsonrpc.ValidatorBase(c, &dto)
	if err != nil {
		return nil, err
	}
	container, err := di.NewDIContainer(diLoggerConf)
	if err != nil {
		return nil, err
	}
	defer container.Close()
	uc := container.LTIRoutingDeleteUC()
	output, err := uc.Execute(dto)
	return output, err
}

// LTIRoutingGet func for full model LTIRouting.
// @Description get lti_routing. Roles [admin, instructor]
// @Summary get lti_routing
// @Tags lti_routing
// @Accept json
// @Produce json
// @Param object body usecases.LTIRoutingGetRequest true "lti_routing id"
// @Success 200 {object} usecases.LTIRoutingGetResponse
// @Security ApiKeyAuth
// @Router /api/v1/rpc/lti_routing.get [post]
func LTIRoutingGet(c *jsonrpc.Ctx) (interface{}, error) {
	diLoggerConf := logs.NewZeroLoggerConf(c)
	if _, err := auth.ExtractTokenMetadata(
		c,
		[]string{cms_client.SSOUsersRoleAdmin, cms_client.SSOUsersRoleInstructor},
	); err != nil {
		return nil, err
	}
	dto := usecases.LTIRoutingGetInputDTO{}
	err := jsonrpc.ValidatorBase(c, &dto)
	if err != nil {
		return nil, err
	}
	container, err := di.NewDIContainer(diLoggerConf)
	if err != nil {
		return nil, err
	}
	defer container.Close()
	uc := container.LTIRoutingGetUC()
	output, err := uc.Execute(dto)
	return output, err
}
