package cmd

import (
	"github.com/futurehomeno/cliffhanger/bootstrap"
	"github.com/futurehomeno/cliffhanger/edge"
	"github.com/futurehomeno/cliffhanger/router"
	log "github.com/sirupsen/logrus"

	"github.com/futurehomeno/edge-hue-adapter/internal/config"
	"github.com/futurehomeno/edge-hue-adapter/internal/routing"
)

// Execute loads configuration and then builds and executes the application.
func Execute() {
	// nolint:forcetypeassert
	cfg := getConfigService().Model().(*config.Config)

	bootstrap.InitializeLogger(cfg.LogFile, cfg.LogLevel, cfg.LogFormat)

	app, err := Build(cfg)
	if err != nil {
		log.WithError(err).Fatalf("failed to build the application")
	}

	err = app.Start()
	if err != nil {
		log.WithError(err).Fatalf("failed to start the application")
	}

	bootstrap.WaitForShutdown()

	err = app.Stop()
	if err != nil {
		log.WithError(err).Fatalf("failed to stop the application")
	}
}

// Build builds the application.
func Build(cfg *config.Config) (edge.Edge, error) {
	app, err := edge.NewBuilder().
		WithMQTT(getMQTT(cfg)).
		WithServiceDiscovery(routing.GetDiscoveryResource()).
		WithLifecycle(getLifecycle(cfg)).
		WithTopicSubscription(
			router.TopicPatternAdapter(routing.ResourceName),
			router.TopicPatternDevices(routing.ResourceName),
		).
		WithRouting(newRouting(cfg)...).
		WithTask(newTasks(cfg)...).
		Build()
	if err != nil {
		return nil, err
	}

	return app, nil
}
