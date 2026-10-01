package controllers

import (
	"github.com/cms-lab-core/cms-labs-api/backend/app/di"
	"github.com/cms-lab-core/cms-labs-api/backend/app/usecases"
	"github.com/cms-lab-core/cms-labs-api/backend/app/usecases/auth"
	"github.com/cms-lab-core/cms-labs-api/shared/jsonrpc"
	"github.com/cms-lab-core/cms-labs-api/shared/logs"
)

// TargetAddonCreate func for creates a new Addon with Target.
// @Description Create target_addon. Roles: [admin, instructor, student]
// @Summary create target_addon
// @Tags target_addon
// @Accept json
// @Produce json
// @Param object body usecases.TargetAddonCreateRequest true "target_addon form info"
// @Success 200 {object} usecases.TargetAddonCreateResponse
// @Security ApiKeyAuth
// @Router /api/v1/rpc/target.addon_create [post]
func TargetAddonCreate(c *jsonrpc.Ctx) (interface{}, error) {
	issuer, err := auth.IssuerURLByBaseUrl(c.FiberCtx.BaseURL())
	if err != nil {
		return nil, err
	}
	claims, err := auth.ExtractTokenMetadata(c, []string{})
	if err != nil {
		return nil, err
	}
	diLoggerConf := logs.NewZeroLoggerConf(c)
	dto := usecases.TargetAddonCreateInputDTO{}
	if err := jsonrpc.ValidatorBase(c, &dto); err != nil {
		return nil, err
	}
	container, err := di.NewDIContainer(diLoggerConf)
	if err != nil {
		return nil, err
	}
	defer container.Close()
	uc := container.TargetAddonCreateUC()
	dto.IssId = issuer
	output, err := uc.SetContext(claims).Execute(dto)
	return output, err
}

// TargetAddonDelete deletes a target addon.
// @Description Delete target_addon. Roles: [admin, instructor, student]
// @Summary delete target_addon
// @Tags target_addon
// @Accept json
// @Produce json
// @Param object body usecases.TargetAddonDeleteInputDTO true "target_addon delete info"
// @Success 200 {object} usecases.TargetAddonDeleteOutputDTO
// @Security ApiKeyAuth
// @Router /api/v1/rpc/target.addon_delete [post]
func TargetAddonDelete(c *jsonrpc.Ctx) (interface{}, error) {
	issuer, err := auth.IssuerURLByBaseUrl(c.FiberCtx.BaseURL())
	if err != nil {
		return nil, err
	}
	claims, err := auth.ExtractTokenMetadata(c, []string{})
	if err != nil {
		return nil, err
	}
	loggerConf := logs.NewZeroLoggerConf(c)
	dto := usecases.TargetAddonDeleteInputDTO{}
	if err := jsonrpc.ValidatorBase(c, &dto); err != nil {
		return nil, err
	}
	container, err := di.NewDIContainer(loggerConf)
	if err != nil {
		return nil, err
	}
	defer container.Close()
	uc := container.TargetAddonDeleteUC()
	dto.IssId = issuer
	output, err := uc.SetContext(claims).Execute(dto)
	return output, err
}

// TargetAddonReset resets a target addon.
// @Description Reset target_addon. Roles: [admin, instructor, student]
// @Summary reset target_addon
// @Tags target_addon
// @Accept json
// @Produce json
// @Param object body usecases.TargetAddonResetInputDTO true "target_addon reset info"
// @Success 200 {object} usecases.TargetAddonResetOutputDTO
// @Security ApiKeyAuth
// @Router /api/v1/rpc/target.addon_reset [post]
func TargetAddonReset(c *jsonrpc.Ctx) (interface{}, error) {
	issuer, err := auth.IssuerURLByBaseUrl(c.FiberCtx.BaseURL())
	if err != nil {
		return nil, err
	}
	claims, err := auth.ExtractTokenMetadata(c, []string{})
	if err != nil {
		return nil, err
	}
	loggerConf := logs.NewZeroLoggerConf(c)
	dto := usecases.TargetAddonResetInputDTO{}
	if err := jsonrpc.ValidatorBase(c, &dto); err != nil {
		return nil, err
	}
	container, err := di.NewDIContainer(loggerConf)
	if err != nil {
		return nil, err
	}
	dto.IssId = issuer
	defer container.Close()
	uc := container.TargetAddonResetUC()
	output, err := uc.SetContext(claims).Execute(dto)
	return output, err
}

// TargetDelete deletes a target.
// @Description Delete target. Roles: [admin, instructor, student]
// @Summary delete target
// @Tags target
// @Accept json
// @Produce json
// @Param object body usecases.TargetDeleteInputDTO true "target delete info"
// @Success 200 {object} usecases.TargetDeleteOutputDTO
// @Security ApiKeyAuth
// @Router /api/v1/rpc/target.delete [post]
func TargetDelete(c *jsonrpc.Ctx) (interface{}, error) {
	claims, err := auth.ExtractTokenMetadata(c, []string{})
	if err != nil {
		return nil, err
	}
	loggerConf := logs.NewZeroLoggerConf(c)
	dto := usecases.TargetDeleteInputDTO{}
	if err := jsonrpc.ValidatorBase(c, &dto); err != nil {
		return nil, err
	}
	container, err := di.NewDIContainer(loggerConf)
	if err != nil {
		return nil, err
	}
	defer container.Close()
	uc := container.TargetDeleteUC()
	output, err := uc.SetContext(claims).Execute(dto)
	return output, err
}

// TargetGet retrieves a target.
// @Description Get target. Roles: [admin, instructor, student]
// @Summary get target
// @Tags target
// @Accept json
// @Produce json
// @Param object body usecases.TargetGetRequest true "target get info"
// @Success 200 {object} usecases.TargetGetResponse
// @Security ApiKeyAuth
// @Router /api/v1/rpc/target.get [post]
func TargetGet(c *jsonrpc.Ctx) (interface{}, error) {
	_, err := auth.ExtractTokenMetadata(c, []string{})
	if err != nil {
		return nil, err
	}
	loggerConf := logs.NewZeroLoggerConf(c)
	dto := usecases.TargetGetInputDTO{}
	if err := jsonrpc.ValidatorBase(c, &dto); err != nil {
		return nil, err
	}
	container, err := di.NewDIContainer(loggerConf)
	if err != nil {
		return nil, err
	}
	defer container.Close()
	uc := container.TargetGetUC()
	output, err := uc.Execute(dto)
	return output, err
}

// TargetList list of target.
// @Description list target. Roles: [admin, instructor, student]
// @Summary list target
// @Tags target
// @Accept json
// @Produce json
// @Param object body usecases.TargetListRequest true "target list info"
// @Success 200 {object} usecases.TargetListResponse
// @Security ApiKeyAuth
// @Router /api/v1/rpc/target.list [post]
func TargetList(c *jsonrpc.Ctx) (interface{}, error) {
	claims, err := auth.ExtractTokenMetadata(c, []string{})
	if err != nil {
		return nil, err
	}
	loggerConf := logs.NewZeroLoggerConf(c)
	dto := usecases.TargetListInputDTO{}
	if err := jsonrpc.ValidatorBase(c, &dto); err != nil {
		return nil, err
	}
	container, err := di.NewDIContainer(loggerConf)
	if err != nil {
		return nil, err
	}
	defer container.Close()
	uc := container.TargetListUC()
	output, err := uc.SetContext(claims).Execute(dto)
	return output, err
}

// TargetRelationCreate creates a target relation.
// @Description Create target_relation. Roles: [admin, instructor, student]
// @Summary create target_relation
// @Tags target_relation
// @Accept json
// @Produce json
// @Param object body usecases.TargetRelationCreateRequest true "target_relation create info"
// @Success 200 {object} usecases.TargetRelationCreateResponse
// @Security ApiKeyAuth
// @Router /api/v1/rpc/target.relation_create [post]
func TargetRelationCreate(c *jsonrpc.Ctx) (interface{}, error) {
	claims, err := auth.ExtractTokenMetadata(c, []string{})
	if err != nil {
		return nil, err
	}
	loggerConf := logs.NewZeroLoggerConf(c)
	dto := usecases.TargetRelationCreateInputDTO{}
	if err := jsonrpc.ValidatorBase(c, &dto); err != nil {
		return nil, err
	}
	container, err := di.NewDIContainer(loggerConf)
	if err != nil {
		return nil, err
	}
	defer container.Close()
	uc := container.TargetRelationCreateUC()
	output, err := uc.SetContext(claims).Execute(dto)
	return output, err
}

// TargetRelationDelete deletes a target relation.
// @Description Delete target_relation. Roles: [admin, instructor, student]
// @Summary delete target_relation
// @Tags target_relation
// @Accept json
// @Produce json
// @Param object body usecases.TargetRelationDeleteRequest true "target_relation delete info"
// @Success 200 {object} usecases.TargetRelationDeleteResponse
// @Security ApiKeyAuth
// @Router /api/v1/rpc/target.relation_delete [post]
func TargetRelationDelete(c *jsonrpc.Ctx) (interface{}, error) {
	claims, err := auth.ExtractTokenMetadata(c, []string{})
	if err != nil {
		return nil, err
	}
	loggerConf := logs.NewZeroLoggerConf(c)
	dto := usecases.TargetRelationDeleteInputDTO{}
	if err := jsonrpc.ValidatorBase(c, &dto); err != nil {
		return nil, err
	}
	container, err := di.NewDIContainer(loggerConf)
	if err != nil {
		return nil, err
	}
	defer container.Close()
	uc := container.TargetRelationDeleteUC()
	output, err := uc.SetContext(claims).Execute(dto)
	return output, err
}

// TargetUpsert creates or updates a target.
// @Description Upsert target. Roles: [admin, instructor, student]
// @Summary upsert target
// @Tags target
// @Accept json
// @Produce json
// @Param object body usecases.TargetUpsertRequest true "target upsert info"
// @Success 200 {object} usecases.TargetUpsertResponse
// @Security ApiKeyAuth
// @Router /api/v1/rpc/target.upsert [post]
func TargetUpsert(c *jsonrpc.Ctx) (interface{}, error) {
	claims, err := auth.ExtractTokenMetadata(c, []string{})
	if err != nil {
		return nil, err
	}
	loggerConf := logs.NewZeroLoggerConf(c)
	dto := usecases.TargetUpsertInputDTO{}
	if err := jsonrpc.ValidatorBase(c, &dto); err != nil {
		return nil, err
	}
	container, err := di.NewDIContainer(loggerConf)
	if err != nil {
		return nil, err
	}
	defer container.Close()
	uc := container.TargetUpsertUC()
	output, err := uc.SetContext(claims).Execute(dto)
	return output, err
}

// TargetUserDelete deletes a target-user association.
// @Description Delete target_user. Roles: [admin, instructor, student]
// @Summary delete target_user
// @Tags target_user
// @Accept json
// @Produce json
// @Param object body usecases.TargetUserDeleteRequest true "target_user delete info"
// @Success 200 {object} usecases.TargetUserDeleteResponse
// @Security ApiKeyAuth
// @Router /api/v1/rpc/target.user_delete [post]
func TargetUserDelete(c *jsonrpc.Ctx) (interface{}, error) {
	claims, err := auth.ExtractTokenMetadata(c, []string{})
	if err != nil {
		return nil, err
	}
	loggerConf := logs.NewZeroLoggerConf(c)
	dto := usecases.TargetUserDeleteInputDTO{}
	if err := jsonrpc.ValidatorBase(c, &dto); err != nil {
		return nil, err
	}
	container, err := di.NewDIContainer(loggerConf)
	if err != nil {
		return nil, err
	}
	defer container.Close()
	uc := container.TargetUserDeleteUC()
	output, err := uc.SetContext(claims).Execute(dto)
	return output, err
}

// TargetUserUpsert creates or updates a target-user association.
// @Description Upsert target_user. Roles: [admin, instructor, student]
// @Summary upsert target_user
// @Tags target_user
// @Accept json
// @Produce json
// @Param object body usecases.TargetUserUpsertRequest true "target_user upsert info"
// @Success 200 {object} usecases.TargetUserUpsertResponse
// @Security ApiKeyAuth
// @Router /api/v1/rpc/target.user_upsert [post]
func TargetUserUpsert(c *jsonrpc.Ctx) (interface{}, error) {
	claims, err := auth.ExtractTokenMetadata(c, []string{})
	if err != nil {
		return nil, err
	}
	loggerConf := logs.NewZeroLoggerConf(c)
	dto := usecases.TargetUserUpsertInputDTO{}
	if err := jsonrpc.ValidatorBase(c, &dto); err != nil {
		return nil, err
	}
	container, err := di.NewDIContainer(loggerConf)
	if err != nil {
		return nil, err
	}
	defer container.Close()
	uc := container.TargetUserUpsertUC()
	output, err := uc.SetContext(claims).Execute(dto)
	return output, err
}
