package controllers

import (
	"github.com/maintainer64/cms-labs-api/backend/app/di"
	"github.com/maintainer64/cms-labs-api/backend/app/usecases"
	"github.com/maintainer64/cms-labs-api/backend/app/usecases/auth"
	"github.com/maintainer64/cms-labs-api/shared/jsonrpc"
	"github.com/maintainer64/cms-labs-api/shared/logs"
)

func LabCatalogList(c *jsonrpc.Ctx) (interface{}, error) {
	user, err := auth.ExtractTokenMetadata(c, nil)
	if err != nil {
		return nil, err
	}
	dto := usecases.LabCatalogListInputDTO{}
	if err := jsonrpc.ValidatorBase(c, &dto); err != nil {
		return nil, err
	}
	container, err := di.NewDIContainer(logs.NewZeroLoggerConf(c))
	if err != nil {
		return nil, err
	}
	defer container.Close()
	return container.LabCatalogUC().SetContext(user).List(dto)
}

func LabCatalogStart(c *jsonrpc.Ctx) (interface{}, error) {
	user, err := auth.ExtractTokenMetadata(c, nil)
	if err != nil {
		return nil, err
	}
	dto := usecases.LabCatalogStartInputDTO{}
	if err := jsonrpc.ValidatorBase(c, &dto); err != nil {
		return nil, err
	}
	container, err := di.NewDIContainer(logs.NewZeroLoggerConf(c))
	if err != nil {
		return nil, err
	}
	defer container.Close()
	return container.LabCatalogUC().SetContext(user).Start(dto)
}
