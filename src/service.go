package main

import (
	"flag"
	"fmt"
	"time"

	"github.com/amimof/huego"
	"github.com/futurehomeno/fimpgo"
	"github.com/futurehomeno/fimpgo/discovery"
	log "github.com/sirupsen/logrus"
	"github.com/thingsplex/hue-ad/model"
	"github.com/thingsplex/hue-ad/router"
	"github.com/thingsplex/hue-ad/utils"
)

func main() {
	var workDir string
	flag.StringVar(&workDir, "c", "", "Work dir")
	flag.Parse()
	if workDir == "" {
		workDir = "./"
	} else {
		fmt.Println("Work dir ", workDir)
	}
	appLifecycle := model.NewAppLifecycle()
	configs := model.NewConfigs(workDir)
	err := configs.LoadFromFile()
	if err != nil {
		fmt.Print(err)
		panic("Can't load config file.")
	}

	utils.SetupLog(configs.LogFile, configs.LogLevel, configs.LogFormat)
	log.Info("--------------Starting hue-ad----------------")
	appLifecycle.SetAppState(model.AppStateStarting, nil)
	mqtt := fimpgo.NewMqttTransport(configs.MqttServerURI, configs.MqttClientIdPrefix, configs.MqttUsername, configs.MqttPassword, true, 1, 1)
	err = mqtt.Start()
	responder := discovery.NewServiceDiscoveryResponder(mqtt)
	responder.RegisterResource(model.GetDiscoveryResource())
	responder.Start()
	var bridge **huego.Bridge
	b := &huego.Bridge{}
	bridge = &b

	stateMonitor := router.NewStateMonitor(mqtt, bridge, configs.InstanceAddress)
	stateMonitor.SetPoolingInterval(configs.StatePoolingInterval)
	fimpRouter := router.NewFromFimpRouter(mqtt, appLifecycle, configs, bridge, stateMonitor)
	fimpRouter.Start()

	appLifecycle.SetConnectionState(model.ConnStateDisconnected)
	if configs.IsConfigured() && err == nil {
		appLifecycle.SetConfigState(model.ConfigStateConfigured)
		appLifecycle.SetAppState(model.AppStateRunning, nil)
	} else {
		appLifecycle.SetAppState(model.AppStateNotConfigured, nil)
		appLifecycle.SetConfigState(model.ConfigStateNotConfigured)
	}

	var br []huego.Bridge
	var retryCounter int
	for {
		// At this point application mightn not be configured and the app is waiting for user actions
		appLifecycle.WaitForState("main", model.AppStateRunning)
		if appLifecycle.ConnectionState() == model.ConnStateConnected {
			stateMonitor.Start()
		} else {
			retryCounter = 0
			for {
				appLifecycle.SetConnectionState(model.ConnStateConnecting)
				br, err = huego.DiscoverAll()
				log.Info("DiscoverAll returned br: ", br)
				if err == nil {
					break
				} else {
					log.Error("Can't discover the bridge. retrying... ", err)
					retryCounter++
					if retryCounter > 10 {
						break
					}
					time.Sleep(time.Second * 5 * time.Duration(retryCounter))
				}
			}

			if err == nil {
				for _, b := range br {
					if b.ID == configs.BridgeId {
						log.Info("Found b.ID == configs.BridgeId")
						log.Info("b.ID: ", b.ID)
						log.Info("configs.BridgeId: ", configs.BridgeId)
						*bridge = &b
					}
				}
				if (*bridge).ID != "" {
					log.Info("Bridge discovered on address = %s , id = %s", (*bridge).Host, (*bridge).ID)
					(*bridge).Login(configs.Token)
					err = stateMonitor.TestConnection()
					if err != nil {
						appLifecycle.SetConnectionState(model.ConnStateDisconnected)
						appLifecycle.SetLastError(err.Error())
					} else {
						appLifecycle.SetConnectionState(model.ConnStateConnected)
						stateMonitor.Start()
					}

				} else {
					log.Info("Adapter is not configured")
					appLifecycle.SetAppState(model.AppStateNotConfigured, nil)
				}
			} else {
				appLifecycle.SetAppState(model.AppStateStartupError, nil)
			}
		}
		appLifecycle.WaitForState("main", model.AppStateNotConfigured)
	}

	mqtt.Stop()
	time.Sleep(5 * time.Second)
}
