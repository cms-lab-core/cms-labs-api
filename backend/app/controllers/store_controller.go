package controllers

import (
	"github.com/cms-lab-core/cms-labs-api/backend/app/di"
	"github.com/cms-lab-core/cms-labs-api/backend/app/models/types"
	"github.com/cms-lab-core/cms-labs-api/backend/app/usecases/auth"
	"github.com/cms-lab-core/cms-labs-api/shared/jsonrpc"
	"github.com/cms-lab-core/cms-labs-api/shared/logs"
	"github.com/goccy/go-json"
)

// UserGlobalStoreGet func for get global store.
// @Description Global store of user get
// @Summary get store by user
// @Tags user
// @Accept json
// @Produce json
// @Param object body types.UserStoreGetRequest true "store of get"
// @Success 200 {object} types.UserStoreGetResponse
// @Security ApiKeyAuth
// @Router /api/v1/rpc/user.global_store_get [post]
func UserGlobalStoreGet(c *jsonrpc.Ctx) (interface{}, error) {
	diLoggerConf := logs.NewZeroLoggerConf(c)
	user, err := auth.ExtractTokenMetadata(c, []string{})
	if err != nil {
		return nil, err
	}
	container, err := di.NewDIContainer(diLoggerConf)
	if err != nil {
		return nil, err
	}
	defer container.Close()
	store, err := container.Queries.UserQueries.StoreGetByUserId(user.UserID())
	return store, err
}

// UserGlobalStoreSet func for get global store.
// @Description Global store of user get
// @Summary set store by user
// @Tags user
// @Accept json
// @Produce json
// @Param object body types.UserStoreSetRequest true "store of create"
// @Success 200 {object} types.UserStoreSetResponse
// @Security ApiKeyAuth
// @Router /api/v1/rpc/user.global_store_set [post]
func UserGlobalStoreSet(c *jsonrpc.Ctx) (interface{}, error) {
	diLoggerConf := logs.NewZeroLoggerConf(c)
	user, err := auth.ExtractTokenMetadata(c, []string{})
	if err != nil {
		return nil, err
	}
	store := types.JsonStore{}
	err = json.Unmarshal(c.Params, &store)
	if err != nil {
		return nil, err
	}
	container, err := di.NewDIContainer(diLoggerConf)
	if err != nil {
		return nil, err
	}
	defer container.Close()
	err = container.Queries.UserQueries.StoreSetByUserId(user.UserID(), store)
	return true, err
}
