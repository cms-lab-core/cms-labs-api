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

// LTIAttemptCreate func for create or get exists LTIAttempt.
// @Description Create or get exists lti_attempt. Roles: [admin, instructor, student]
// @Summary create lti_attempt
// @Tags lti_attempt
// @Accept json
// @Produce json
// @Param form body usecases.LTIAttemptCreateRequest true "lti_attempt form info"
// @Success 200 {object} usecases.LTIAttemptCreateResponse
// @Security ApiKeyAuth
// @Router /api/v1/rpc/lti_attempt.create [post]
func LTIAttemptCreate(c *jsonrpc.Ctx) (interface{}, error) {
	diLoggerConf := logs.NewZeroLoggerConf(c)
	token, err := auth.ExtractTokenMetadata(c, []string{})
	if err != nil {
		return nil, err
	}
	dto := usecases.LTIAttemptCreateInputDTO{}
	err = jsonrpc.ValidatorBase(c, &dto)
	if err != nil {
		return nil, err
	}
	container, err := di.NewDIContainer(diLoggerConf)
	if err != nil {
		return nil, err
	}
	defer container.Close()
	uc := container.LTIAttemptCreateUC()
	output, err := uc.SetContext(token).Execute(dto)
	return output, err
}

// LTIAttemptUpdate func for edit LTIAttempt.
// @Description Edit lti_attempt. Roles: [admin, instructor]
// @Summary edit lti_attempt
// @Tags lti_attempt
// @Accept json
// @Produce json
// @Param form body usecases.LTIAttemptEditRequest true "lti_attempt form info"
// @Success 200 {object} usecases.LTIAttemptEditResponse
// @Security ApiKeyAuth
// @Router /api/v1/rpc/lti_attempt.update [post]
func LTIAttemptUpdate(c *jsonrpc.Ctx) (interface{}, error) {
	diLoggerConf := logs.NewZeroLoggerConf(c)
	if _, err := auth.ExtractTokenMetadata(
		c,
		[]string{cms_client.SSOUsersRoleAdmin, cms_client.SSOUsersRoleInstructor},
	); err != nil {
		return nil, err
	}
	dto := usecases.LTIAttemptEditInputDTO{}
	err := jsonrpc.ValidatorBase(c, &dto)
	if err != nil {
		return nil, err
	}
	container, err := di.NewDIContainer(diLoggerConf)
	if err != nil {
		return nil, err
	}
	defer container.Close()
	uc := container.LTIAttemptEditUC()
	output, err := uc.Execute(dto)
	return output, err
}

// LTIAttemptList func for view of list LTIAttempt.
// @Description List lti_attempt. Roles: [admin, instructor]
// @Summary list lti_attempt
// @Tags lti_attempt
// @Accept json
// @Produce json
// @Param form body usecases.LTIAttemptListRequest true "lti_attempt list info"
// @Success 200 {object} usecases.LTIAttemptListResponse
// @Security ApiKeyAuth
// @Router /api/v1/rpc/lti_attempt.list [post]
func LTIAttemptList(c *jsonrpc.Ctx) (interface{}, error) {
	diLoggerConf := logs.NewZeroLoggerConf(c)
	if _, err := auth.ExtractTokenMetadata(
		c,
		[]string{cms_client.SSOUsersRoleAdmin, cms_client.SSOUsersRoleInstructor},
	); err != nil {
		return nil, err
	}
	dto := queries.LTIAttemptSearchParams{}
	err := jsonrpc.ValidatorBase(c, &dto)
	if err != nil {
		return nil, err
	}
	container, err := di.NewDIContainer(diLoggerConf)
	if err != nil {
		return nil, err
	}
	defer container.Close()
	uc := container.LTIAttemptListUC()
	output, err := uc.Execute(dto)
	return output, err
}

// LTIAttemptDelete func for delete LTIAttempt.
// @Description Delete lti_attempt.
// @Summary delete lti_attempt
// @Tags lti_attempt
// @Accept json
// @Produce json
// @Param form body usecases.LTIAttemptDeleteRequest true "lti_attempt id"
// @Success 200 {object} usecases.LTIAttemptDeleteResponse
// @Security ApiKeyAuth
// @Router /api/v1/rpc/lti_attempt.delete [post]
func LTIAttemptDelete(c *jsonrpc.Ctx) (interface{}, error) {
	diLoggerConf := logs.NewZeroLoggerConf(c)
	if _, err := auth.ExtractTokenMetadata(
		c,
		[]string{cms_client.SSOUsersRoleAdmin, cms_client.SSOUsersRoleInstructor},
	); err != nil {
		return nil, err
	}
	dto := usecases.LTIAttemptDeleteInputDTO{}
	err := jsonrpc.ValidatorBase(c, &dto)
	if err != nil {
		return nil, err
	}
	container, err := di.NewDIContainer(diLoggerConf)
	if err != nil {
		return nil, err
	}
	defer container.Close()
	uc := container.LTIAttemptDeleteUC()
	output, err := uc.Execute(dto)
	return output, err
}

// LTIAttemptGet func for full model LTIAttempt.
// @Description get lti_attempt.
// @Summary get lti_attempt
// @Tags lti_attempt
// @Accept json
// @Produce json
// @Param form body usecases.LTIAttemptGetRequest true "lti_attempt id"
// @Success 200 {object} usecases.LTIAttemptGetResponse
// @Security ApiKeyAuth
// @Router /api/v1/rpc/lti_attempt.get [post]
func LTIAttemptGet(c *jsonrpc.Ctx) (interface{}, error) {
	diLoggerConf := logs.NewZeroLoggerConf(c)
	dto := usecases.LTIAttemptGetInputDTO{}
	err := jsonrpc.ValidatorBase(c, &dto)
	if err != nil {
		return nil, err
	}
	container, err := di.NewDIContainer(diLoggerConf)
	if err != nil {
		return nil, err
	}
	defer container.Close()
	uc := container.LTIAttemptGetUC()
	output, err := uc.Execute(dto)
	return output, err
}

// LTIAttemptListExternal func for view of list LTIAttempt by external.
// @Description List lti_attempt. Only external servers
// @Summary list lti_attempt
// @Tags lti_attempt, EXTERNAL
// @Accept json
// @Produce json
// @Param form body usecases.LTIAttemptListRequest true "lti_attempt list info"
// @Success 200 {object} usecases.LTIAttemptListResponse
// @Param Authorization header string true "Basic-токен, созданный клиентом"
// @Router /api/v1/rpc/lti_attempt.list_external [post]
func LTIAttemptListExternal(c *jsonrpc.Ctx) (interface{}, error) {
	diLoggerConf := logs.NewZeroLoggerConf(c)
	dto := queries.LTIAttemptSearchParams{}
	err := jsonrpc.ValidatorBase(c, &dto)
	if err != nil {
		return nil, err
	}
	container, err := di.NewDIContainer(diLoggerConf)
	if err != nil {
		return nil, err
	}
	defer container.Close()
	err = container.ServiceAuthorizeUC().Execute(c)
	if err != nil {
		return nil, err
	}
	uc := container.LTIAttemptListUC()
	dto.ServerClientIds = []string{
		c.FiberCtx.Locals("x-service-id").(string),
	}
	output, err := uc.Execute(dto)
	return output, err
}

// LTIAttemptUpdateExternal update attempts from external servers.
// @Description attempt update from external servers.
// @Summary attempt update from external servers
// @Tags lti_attempt, EXTERNAL
// @Accept json
// @Produce json
// @Param object body usecases.LTIAttemptEditBulkRequest true "attempts"
// @Success 200 {object} usecases.LTIAttemptEditBulkResponse
// @Param Authorization header string true "Basic-токен, созданный клиентом"
// @Router /api/v1/rpc/lti_attempt.update_external [post]
func LTIAttemptUpdateExternal(c *jsonrpc.Ctx) (interface{}, error) {
	diLoggerConf := logs.NewZeroLoggerConf(c)
	dto := usecases.LTIAttemptEditBulkInputDTO{}
	err := jsonrpc.ValidatorBase(c, &dto)
	if err != nil {
		return nil, err
	}
	container, err := di.NewDIContainer(diLoggerConf)
	if err != nil {
		return nil, err
	}
	defer container.Close()
	err = container.ServiceAuthorizeUC().Execute(c)
	if err != nil {
		return nil, err
	}
	uc := container.LTIAttemptEditBulkUC()
	output, err := uc.SetContext(c.FiberCtx.Locals("x-service-id").(string)).Execute(dto)
	return output, err
}
