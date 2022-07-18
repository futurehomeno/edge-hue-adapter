package adapter

import (
	"github.com/futurehomeno/cliffhanger/adapter"
	"github.com/futurehomeno/fimpgo"

	"github.com/futurehomeno/edge-hue-adapter/internal/config"
)

// NewThingFactory creates new instance of a thing factory.
func NewThingFactory(cfgSrv *config.Service) adapter.ThingFactory {
	return &thingFactory{
		cfgSrv: cfgSrv,
	}
}

// thingFactory is a private implementation of a thing factory service.
type thingFactory struct {
	cfgSrv *config.Service
}

// Create creates an instance of a thing using provided state.
func (f *thingFactory) Create(mqtt *fimpgo.MqttTransport, adapter adapter.ExtendedAdapter, thingState adapter.ThingState) (adapter.Thing, error) {
	// TODO: This is where you create things whenever adapter requests it.
	//  Usually it happens upon initialization or when adapter is specifically called to create a particular thing.
	panic("implement me")
}
