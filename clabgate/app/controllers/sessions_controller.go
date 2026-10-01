package controllers

import (
	"github.com/cms-lab-core/cms-labs-api/clabgate/app/di"
	"github.com/cms-lab-core/cms-labs-api/clabgate/app/usecases"
	"github.com/cms-lab-core/cms-labs-api/clabgate/app/usecases/auth"
	"github.com/cms-lab-core/cms-labs-api/shared/jsonrpc"
	"github.com/cms-lab-core/cms-labs-api/shared/logs"
)

func sessionsUC(c *jsonrpc.Ctx) (*usecases.SessionsUC, error) {
	user, err := auth.ExtractTokenMetadata(c, []string{})
	if err != nil {
		return nil, err
	}
	container, err := di.NewDIContainer(logs.NewZeroLoggerConf(c))
	if err != nil {
		return nil, err
	}
	uc, err := container.SessionsUC()
	if err != nil {
		container.Close()
		return nil, err
	}
	return uc.SetContext(user), nil
}

// SessionEnsure validates the active CMS attempt and idempotently creates desired Kubernetes resources.
func SessionEnsure(c *jsonrpc.Ctx) (interface{}, error) {
	dto := usecases.SessionEnsureInputDTO{}
	if err := jsonrpc.ValidatorBase(c, &dto); err != nil {
		return nil, err
	}
	uc, err := sessionsUC(c)
	if err != nil {
		return nil, err
	}
	return uc.Ensure(c.FiberCtx.UserContext(), dto)
}

// SessionGet returns a session derived exclusively from Kubernetes resources.
func SessionGet(c *jsonrpc.Ctx) (interface{}, error) {
	dto := usecases.SessionGetInputDTO{}
	if err := jsonrpc.ValidatorBase(c, &dto); err != nil {
		return nil, err
	}
	uc, err := sessionsUC(c)
	if err != nil {
		return nil, err
	}
	return uc.Get(c.FiberCtx.UserContext(), dto)
}

// SessionList lists Kubernetes-backed sessions visible to the current user.
func SessionList(c *jsonrpc.Ctx) (interface{}, error) {
	dto := usecases.SessionListInputDTO{}
	if err := jsonrpc.ValidatorBase(c, &dto); err != nil {
		return nil, err
	}
	uc, err := sessionsUC(c)
	if err != nil {
		return nil, err
	}
	return uc.List(c.FiberCtx.UserContext(), dto)
}

// SessionStop deletes the session namespace. Kubernetes garbage collection removes session resources.
func SessionStop(c *jsonrpc.Ctx) (interface{}, error) {
	dto := usecases.SessionStopInputDTO{}
	if err := jsonrpc.ValidatorBase(c, &dto); err != nil {
		return nil, err
	}
	uc, err := sessionsUC(c)
	if err != nil {
		return nil, err
	}
	return uc.Stop(c.FiberCtx.UserContext(), dto)
}

// SessionCheck starts the configured checker as a Job inside the session namespace.
func SessionCheck(c *jsonrpc.Ctx) (interface{}, error) {
	dto := usecases.SessionCheckInputDTO{}
	if err := jsonrpc.ValidatorBase(c, &dto); err != nil {
		return nil, err
	}
	uc, err := sessionsUC(c)
	if err != nil {
		return nil, err
	}
	return uc.Check(c.FiberCtx.UserContext(), dto)
}

// SessionOpen returns a short-lived exchange URL for an authorized JupyterLab.
func SessionOpen(c *jsonrpc.Ctx) (interface{}, error) {
	dto := usecases.SessionOpenInputDTO{}
	if err := jsonrpc.ValidatorBase(c, &dto); err != nil {
		return nil, err
	}
	uc, err := sessionsUC(c)
	if err != nil {
		return nil, err
	}
	return uc.Open(c.FiberCtx.UserContext(), dto)
}
