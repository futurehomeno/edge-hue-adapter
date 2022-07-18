package routing

import (
	cliffAdapter "github.com/futurehomeno/cliffhanger/adapter"
	"github.com/futurehomeno/cliffhanger/app"
	cliffConfig "github.com/futurehomeno/cliffhanger/config"
	"github.com/futurehomeno/cliffhanger/lifecycle"
	"github.com/futurehomeno/cliffhanger/router"

	"github.com/futurehomeno/edge-hue-adapter/internal/config"
)

const (
	ResourceName = "hue" // ResourceName is the name of the application.
	ServiceName  = "hue" // ServiceName is the name of the main service of the application.
)

// New creates a new routing table with all message handlers and their voters.
func New(
	cfgSrv *config.Service,
	appLifecycle *lifecycle.Lifecycle,
	configurationLocker router.MessageHandlerLocker,
	application app.App,
	adapter cliffAdapter.Adapter,
) []*router.Routing {
	return router.Combine(
		// TODO: Add here any routing specific for your application or its internal services.
		[]*router.Routing{
			cliffConfig.RouteCmdLogSetLevel(ServiceName, cfgSrv.SetLogLevel),
			cliffConfig.RouteCmdLogGetLevel(ServiceName, cfgSrv.GetLogLevel),
		},
		app.RouteApp(ServiceName, appLifecycle, cfgSrv, config.Factory, configurationLocker, application),
		cliffAdapter.RouteAdapter(adapter, nil),
		// TODO: You should add here entire routing for an adapter, such as listeners for commands for devices.
		//  You can create your own routing or use predefined routes from cliffhanger, e.g.: thing.RouteBoiler(), meterelec.Route().
	)
}
