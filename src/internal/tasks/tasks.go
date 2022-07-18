package tasks

import (
	"github.com/futurehomeno/cliffhanger/adapter"
	"github.com/futurehomeno/cliffhanger/app"
	"github.com/futurehomeno/cliffhanger/lifecycle"
	"github.com/futurehomeno/cliffhanger/task"

	"github.com/futurehomeno/edge-hue-adapter/internal/config"
)

// New creates a new definition of background tasks to be performed by the application.
func New(
	cfgSrv *config.Service,
	appLifecycle *lifecycle.Lifecycle,
	application app.App,
	adapter adapter.Adapter,
) []*task.Task {
	return task.Combine(
		// TODO: You should add here all initialization or recurring tasks for your application.
		app.TaskApp(application, appLifecycle),
		// TODO: You should add here all recurring tasks for an adapter, such as periodic reporting of readings.
		//  You can create your own tasks or use predefined tasks from cliffhanger, e.g.: thing.TaskBoiler(), meterelec.TaskReporting().
	)
}
