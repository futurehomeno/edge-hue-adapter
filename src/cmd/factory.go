package cmd

import (
	cliffAdapter "github.com/futurehomeno/cliffhanger/adapter"
	cliffApp "github.com/futurehomeno/cliffhanger/app"
	"github.com/futurehomeno/cliffhanger/bootstrap"
	"github.com/futurehomeno/cliffhanger/lifecycle"
	"github.com/futurehomeno/cliffhanger/manifest"
	"github.com/futurehomeno/cliffhanger/router"
	"github.com/futurehomeno/cliffhanger/task"
	"github.com/futurehomeno/fimpgo"
	log "github.com/sirupsen/logrus"

	"github.com/futurehomeno/edge-hue-adapter/internal/adapter"
	"github.com/futurehomeno/edge-hue-adapter/internal/app"
	"github.com/futurehomeno/edge-hue-adapter/internal/config"
	"github.com/futurehomeno/edge-hue-adapter/internal/routing"
	"github.com/futurehomeno/edge-hue-adapter/internal/tasks"
)

// services is a container for services that are common dependencies.
var services = &serviceContainer{}

// serviceContainer is a type representing a dependency injection container to be used during bootstrap of the application.
type serviceContainer struct {
	configService *config.Service
	mqtt          *fimpgo.MqttTransport
	lifecycle     *lifecycle.Lifecycle

	application         cliffApp.App
	configurationLocker router.MessageHandlerLocker
	manifestLoader      manifest.Loader
	adapter             cliffAdapter.ExtendedAdapter
	thingFactory        cliffAdapter.ThingFactory
	// TODO: You may add any additional dependency that has to be injected, e.g.: API client.
	adapterState cliffAdapter.State
}

// getConfigService initiates a configuration service and loads the config.
func getConfigService() *config.Service {
	if services.configService == nil {
		services.configService = config.NewConfigService(
			bootstrap.GetWorkingDirectory(),
		)

		err := services.configService.Load()
		if err != nil {
			log.WithError(err).Fatal("failed to load configuration")
		}
	}

	return services.configService
}

// getMQTT creates or returns existing MQTT broker service.
func getMQTT(cfg *config.Config) *fimpgo.MqttTransport {
	if services.mqtt == nil {
		services.mqtt = fimpgo.NewMqttTransport(cfg.MQTTServerURI, cfg.MQTTClientIDPrefix, cfg.MQTTUsername, cfg.MQTTPassword, true, 1, 1)
		services.mqtt.SetDefaultSource(routing.ResourceName)
		services.mqtt.SetAutoDecompression(true)
	}

	return services.mqtt
}

// getLifecycle creates or returns existing lifecycle service.
func getLifecycle(_ *config.Config) *lifecycle.Lifecycle {
	if services.lifecycle == nil {
		services.lifecycle = lifecycle.New()
	}

	return services.lifecycle
}

// getManifestLoader creates or returns existing manifest loader service.
func getManifestLoader(cfg *config.Config) manifest.Loader {
	if services.manifestLoader == nil {
		services.manifestLoader = manifest.NewLoader(cfg.WorkDir)
	}

	return services.manifestLoader
}

// getConfigurationLocker creates or returns existing configuration locker.
func getConfigurationLocker(_ *config.Config) router.MessageHandlerLocker {
	if services.configurationLocker == nil {
		services.configurationLocker = router.NewMessageHandlerLocker()
	}

	return services.configurationLocker
}

// getApplication creates or returns existing application.
func getApplication(cfg *config.Config) cliffApp.App {
	if services.application == nil {
		services.application = app.New(
			getConfigService(),
			getLifecycle(cfg),
			getManifestLoader(cfg),
			getAdapter(cfg),
		)
	}

	return services.application
}

// getAdapter creates or returns existing adapter service.
func getAdapter(cfg *config.Config) cliffAdapter.ExtendedAdapter {
	if services.adapter == nil {
		adapterState, err := cliffAdapter.NewState(cfg.WorkDir)
		if err != nil {
			log.WithError(err).Fatal("failed to load adapter state")
		}

		services.adapter = cliffAdapter.NewExtendedAdapter(
			getMQTT(cfg),
			getThingFactory(cfg),
			adapterState,
			routing.ResourceName,
			"1",
		)
	}

	return services.adapter
}

// getThingFactory creates or returns existing thing factory service.
func getThingFactory(cfg *config.Config) cliffAdapter.ThingFactory {
	if services.thingFactory == nil {
		services.thingFactory = adapter.NewThingFactory(getConfigService())
	}

	return services.thingFactory
}

// TODO: getHueClient?

// newRouting creates new set of routing.
func newRouting(cfg *config.Config) []*router.Routing {
	return routing.New(
		getConfigService(),
		getLifecycle(cfg),
		getConfigurationLocker(cfg),
		getApplication(cfg),
		getAdapter(cfg),
	)
}

// newTasks creates new set of tasks.
func newTasks(cfg *config.Config) []*task.Task {
	return tasks.New(
		getConfigService(),
		getLifecycle(cfg),
		getApplication(cfg),
		getAdapter(cfg),
	)
}
