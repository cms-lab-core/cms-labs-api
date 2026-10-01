package routes

import (
	"github.com/cms-lab-core/cms-labs-api/clabgate/app/controllers"
	"github.com/cms-lab-core/cms-labs-api/shared/jsonrpc"
	"github.com/gofiber/fiber/v2"
)

// V1RpcRoute func for describe group of jsonrpc 2.0 protocol.
func V1RpcRoute(a *fiber.App) {
	rpc := jsonrpc.NewJsonRPCServer("/clabgate/api/v1/rpc", a)
	rpc.Method("topology.get", controllers.TopologyGet)
	rpc.Method("node.action", controllers.NodeAction)
	rpc.Method("session.ensure", controllers.SessionEnsure)
	rpc.Method("session.get", controllers.SessionGet)
	rpc.Method("session.list", controllers.SessionList)
	rpc.Method("session.stop", controllers.SessionStop)
	rpc.Method("session.check", controllers.SessionCheck)
	rpc.Method("session.open", controllers.SessionOpen)
}
