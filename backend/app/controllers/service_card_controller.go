package controllers

import (
	"github.com/cms-lab-core/cms-labs-api/backend/app/di"
	"github.com/cms-lab-core/cms-labs-api/backend/app/usecases"
	"github.com/cms-lab-core/cms-labs-api/backend/app/usecases/auth"
	"github.com/cms-lab-core/cms-labs-api/shared/cms_client"
	"github.com/cms-lab-core/cms-labs-api/shared/jsonrpc"
	"github.com/cms-lab-core/cms-labs-api/shared/logs"
)

// ServiceCardUpsert func for creates a new ServiceCard.
// @Description Create service_card. Roles: [admin, instructor]
// @Summary create service_card
// @Tags service_card
// @Accept json
// @Produce json
// @Param object body usecases.ServiceCardEditRequest true "service_card form info"
// @Success 200 {object} usecases.ServiceCardEditResponse
// @Security ApiKeyAuth
// @Router /api/v1/rpc/service_card.upsert [post]
func ServiceCardUpsert(c *jsonrpc.Ctx) (interface{}, error) {
	if _, err := auth.ExtractTokenMetadata(
		c,
		[]string{cms_client.SSOUsersRoleAdmin, cms_client.SSOUsersRoleInstructor},
	); err != nil {
		return nil, err
	}
	diLoggerConf := logs.NewZeroLoggerConf(c)
	dto := usecases.ServiceCardEditInputDTO{}
	err := jsonrpc.ValidatorBase(c, &dto)
	if err != nil {
		return nil, err
	}
	container, err := di.NewDIContainer(diLoggerConf)
	if err != nil {
		return nil, err
	}
	defer container.Close()
	uc := container.ServiceCardEditUC()
	output, err := uc.Execute(dto)
	return output, err
}

// ServiceCardList func for view of list ServiceCard.
// @Description List service_card. Roles: [admin, instructor]
// @Summary list service_card
// @Tags service_card
// @Accept json
// @Produce json
// @Param form body usecases.ServiceCardListRequest true "service_card list info"
// @Success 200 {object} usecases.ServiceCardListResponse
// @Security ApiKeyAuth
// @Router /api/v1/rpc/service_card.list [post]
func ServiceCardList(c *jsonrpc.Ctx) (interface{}, error) {
	if _, err := auth.ExtractTokenMetadata(
		c,
		[]string{cms_client.SSOUsersRoleAdmin, cms_client.SSOUsersRoleInstructor},
	); err != nil {
		return nil, err
	}
	diLoggerConf := logs.NewZeroLoggerConf(c)
	dto := usecases.ServiceCardListInputDTO{}
	err := jsonrpc.ValidatorBase(c, &dto)
	if err != nil {
		return nil, err
	}
	container, err := di.NewDIContainer(diLoggerConf)
	if err != nil {
		return nil, err
	}
	defer container.Close()
	uc := container.ServiceCardListUC()
	output, err := uc.Execute(dto)
	return output, err
}

// ServiceCardDelete func for delete ServiceCard.
// @Description Delete service_card. Roles: [admin, instructor]
// @Summary delete service_card
// @Tags service_card
// @Accept json
// @Produce json
// @Param object body usecases.ServiceCardDeleteRequest true "service_card id"
// @Success 200 {object} usecases.ServiceCardDeleteResponse
// @Security ApiKeyAuth
// @Router /api/v1/rpc/service_card.delete [post]
func ServiceCardDelete(c *jsonrpc.Ctx) (interface{}, error) {
	if _, err := auth.ExtractTokenMetadata(
		c,
		[]string{cms_client.SSOUsersRoleAdmin, cms_client.SSOUsersRoleInstructor},
	); err != nil {
		return nil, err
	}
	diLoggerConf := logs.NewZeroLoggerConf(c)
	dto := usecases.ServiceCardDeleteInputDTO{}
	err := jsonrpc.ValidatorBase(c, &dto)
	if err != nil {
		return nil, err
	}
	container, err := di.NewDIContainer(diLoggerConf)
	if err != nil {
		return nil, err
	}
	defer container.Close()
	uc := container.ServiceCardDeleteUC()
	output, err := uc.Execute(dto)
	return output, err
}

// ServiceCardGet func for full model ServiceCard.
// @Description get service_card. Roles: [admin, instructor]
// @Summary get service_card
// @Tags service_card
// @Accept json
// @Produce json
// @Param object body usecases.ServiceCardGetRequest true "service_card id"
// @Success 200 {object} usecases.ServiceCardGetResponse
// @Security ApiKeyAuth
// @Router /api/v1/rpc/service_card.get [post]
func ServiceCardGet(c *jsonrpc.Ctx) (interface{}, error) {
	if _, err := auth.ExtractTokenMetadata(
		c,
		[]string{cms_client.SSOUsersRoleAdmin, cms_client.SSOUsersRoleInstructor},
	); err != nil {
		return nil, err
	}
	diLoggerConf := logs.NewZeroLoggerConf(c)
	dto := usecases.ServiceCardGetInputDTO{}
	err := jsonrpc.ValidatorBase(c, &dto)
	if err != nil {
		return nil, err
	}
	container, err := di.NewDIContainer(diLoggerConf)
	if err != nil {
		return nil, err
	}
	defer container.Close()
	uc := container.ServiceCardGetUC()
	output, err := uc.Execute(dto)
	return output, err
}
