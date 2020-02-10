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
	"gopkg.in/natefinch/lumberjack.v2"
	"time"
)

func SetupLog(logfile string, level string, logFormat string) {
	if logFormat == "json" {
		log.SetFormatter(&log.JSONFormatter{TimestampFormat: "2006-01-02 15:04:05.999"})
	} else {
		log.SetFormatter(&log.TextFormatter{FullTimestamp: true, ForceColors: true, TimestampFormat: "2006-01-02T15:04:05.999"})
	}

	logLevel, err := log.ParseLevel(level)
	if err == nil {
		log.SetLevel(logLevel)
	} else {
		log.SetLevel(log.DebugLevel)
	}

	if logfile != "" {
		l := lumberjack.Logger{
			Filename:   logfile,
			MaxSize:    5, // megabytes
			MaxBackups: 2,
		}
		log.SetOutput(&l)
	}

}

func main() {
	var configFile string
	flag.StringVar(&configFile, "c", "", "Config file")
	flag.Parse()
	if configFile == "" {
		configFile = "./config.json"
	} else {
		fmt.Println("Loading configs from file ", configFile)
	}
	appLifecycle := model.NewAppLifecycle()
	configs := model.NewConfigs(configFile)
	err := configs.LoadFromFile()
	if err != nil {
		fmt.Print(err)
		panic("Can't load config file.")
	}

	SetupLog(configs.LogFile, configs.LogLevel, configs.LogFormat)
	log.Info("--------------Starting hue-ad----------------")
	appLifecycle.PublishEvent(model.EventConfiguring, "main", nil)

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
		appLifecycle.PublishEvent(model.EventConfigured,"service",nil)
	}
    var br []huego.Bridge
	var retryCounter int
	for {
		appLifecycle.WaitForState("main", model.StateRunning)
		retryCounter = 0
		for {
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
