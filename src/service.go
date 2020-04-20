package main

import (
	"flag"
	"fmt"
	"github.com/amimof/huego"
	"github.com/futurehomeno/fimpgo"
	"github.com/futurehomeno/fimpgo/discovery"
	log "github.com/sirupsen/logrus"
	"github.com/thingsplex/hue-ad/model"
	"github.com/thingsplex/hue-ad/router"
	"github.com/thingsplex/hue-ad/utils"
	"time"
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
	appLifecycle.PublishEvent(model.EventConfiguring, "main", nil)
	appLifecycle.SetConnectionState(model.ConnStateDisconnected)

	mqtt := fimpgo.NewMqttTransport(configs.MqttServerURI, configs.MqttClientIdPrefix, configs.MqttUsername, configs.MqttPassword, true, 1, 1)
	err = mqtt.Start()
	responder := discovery.NewServiceDiscoveryResponder(mqtt)
	responder.RegisterResource(model.GetDiscoveryResource())
	responder.Start()
	var bridge **huego.Bridge
	b := &huego.Bridge{}
	bridge = &b

	stateMonitor := router.NewStateMonitor(mqtt,bridge,configs.InstanceAddress)
    stateMonitor.SetPoolingInterval(configs.StatePoolingInterval)
	fimpRouter := router.NewFromFimpRouter(mqtt,appLifecycle,configs,bridge,stateMonitor)
	fimpRouter.Start()

	if configs.IsConfigured() && err == nil {
		appLifecycle.SetConfigState(model.ConfigStateConfigured)
		appLifecycle.PublishEvent(model.EventConfigured,"service",nil)
	}else {
		appLifecycle.SetAppState(model.AppStateNotConfigured,nil)
	}
    var br []huego.Bridge
	var retryCounter int
	for {
		appLifecycle.WaitForState("main", model.StateRunning)
		retryCounter = 0
		for {
			appLifecycle.SetConnectionState(model.ConnStateConnecting)
			br , err = huego.DiscoverAll()
			if err == nil {
				break
			}else {
				log.Error("Can't discover the bridge. retrying... ",err)
				retryCounter++
				if retryCounter > 10 {
					break
				}
				time.Sleep(time.Second*5*time.Duration(retryCounter))
			}
		}

		if err == nil {
			for _,b := range br {
				if b.ID == configs.BridgeId {
					*bridge = &b
				}
			}
			if (*bridge).ID != "" {
				log.Infof("Bridge discovered on address = %s , id = %s", (*bridge).Host,(*bridge).ID)
				(*bridge).Login(configs.Token)
				err = stateMonitor.TestConnection()
				if err != nil {
					appLifecycle.SetConnectionState(model.ConnStateDisconnected)
					appLifecycle.SetLastError(err.Error())
				}else {
					appLifecycle.SetConnectionState(model.ConnStateConnected)
				}
				stateMonitor.Start()
			}else {
				log.Info("Adapter is not configured")
				appLifecycle.PublishEvent(model.EventConfiguring, "main", nil)
			}
		}else {

		}
		appLifecycle.WaitForState("main", model.StateConfiguring)
	}

	mqtt.Stop()
	time.Sleep(5 * time.Second)
}
