package router

import (
	"fmt"
	"github.com/amimof/huego"
	"github.com/futurehomeno/fimpgo"
	"github.com/lucasb-eyer/go-colorful"
	log "github.com/sirupsen/logrus"
	"github.com/thingsplex/hue-ad/model"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const ServiceName = "hue"

type FromFimpRouter struct {
	inboundMsgCh       fimpgo.MessageCh
	mqt                *fimpgo.MqttTransport
	instanceId         string
	appLifecycle       *model.Lifecycle
	configs            *model.Configs
	bridge             **huego.Bridge
	netService         *model.NetworkService
	stateMonitor       *StateMonitor
	isInclusionRunning bool
}

func NewFromFimpRouter(mqt *fimpgo.MqttTransport, appLifecycle *model.Lifecycle, configs *model.Configs, bridge **huego.Bridge,monitor *StateMonitor) *FromFimpRouter {
	fc := FromFimpRouter{inboundMsgCh: make(fimpgo.MessageCh, 5), mqt: mqt, appLifecycle: appLifecycle, configs: configs, bridge: bridge,stateMonitor:monitor}
	fc.netService = model.NewNetworkService(mqt, bridge)
	fc.mqt.RegisterChannel("ch1", fc.inboundMsgCh)
	return &fc
}

func (fc *FromFimpRouter) Start() {
	fc.mqt.Subscribe("pt:j1/mt:cmd/rt:dev/rn:hue/ad:1/#")
	fc.mqt.Subscribe("pt:j1/mt:cmd/rt:ad/rn:hue/ad:1")
	fc.stateMonitor.SetDimmerMaxValue(fc.configs.DimmerMaxValue)
	fc.netService.SetDimmerMaxVal(fc.configs.DimmerMaxValue)
	go func(msgChan fimpgo.MessageCh) {
		for {
			select {
			case newMsg := <-msgChan:
				fc.routeFimpMessage(newMsg)
			}
		}

	}(fc.inboundMsgCh)
}

func (fc *FromFimpRouter) routeFimpMessage(newMsg *fimpgo.Message) {
	log.Debug("New fimp msg ", newMsg.Payload.Service)
	addr := strings.Replace(newMsg.Addr.ServiceAddress, "_0", "", 1)
	addr = strings.Replace(addr, "l", "", 1)
	addrNum, err := strconv.Atoi(addr)
	if newMsg.Payload.Service == "hue-ad" {
		newMsg.Payload.Service = "hue"
	}

	switch newMsg.Payload.Service {
	case "out_lvl_switch":
		if err != nil {
			return
		}
		var transitionTime  uint16
		val, _ := newMsg.Payload.GetBoolValue()
		light, _ := (*fc.bridge).GetLight(addrNum)
		duration,ok := newMsg.Payload.Properties["duration"]
		if ok {
			tt,err := strconv.Atoi(duration)
			if err == nil {
				transitionTime = uint16(tt)
			}
		}
		switch newMsg.Payload.Type {
		case "cmd.binary.set":
		   if val {
				state := huego.State{On: true,TransitionTime:transitionTime}

				_, err := (*fc.bridge).SetLightState(addrNum, state)
				if err != nil {
					return
				}
				light.State.On = true
			} else {
				state := huego.State{On: false,TransitionTime:transitionTime}
				_, err := (*fc.bridge).SetLightState(addrNum, state)
				if err != nil {
					return
				}
				light.State.On = false
			}
			msg := fimpgo.NewBoolMessage("evt.binary.report", "out_lvl_switch", val, nil, nil, newMsg.Payload)
			adr := &fimpgo.Address{MsgType: fimpgo.MsgTypeEvt, ResourceType: fimpgo.ResourceTypeDevice, ResourceName: ServiceName, ResourceAddress: fc.configs.InstanceAddress, ServiceName: "out_lvl_switch", ServiceAddress: newMsg.Addr.ServiceAddress}
			fc.mqt.Publish(adr,msg)
			//log.Debug("Status code = ",respH.StatusCode)
		case "cmd.lvl.set":
			val, _ := newMsg.Payload.GetIntValue()
			light, _ := (*fc.bridge).GetLight(addrNum)
			//light.Bri(uint8(val))
			if fc.configs.DimmerMaxValue == 100 {
				if val > 100 {
					val = 100
				}
				val = int64((255.0/100.0)*float64(val))
			}
			state := huego.State{On: true,Bri: uint8(val),TransitionTime:transitionTime}
			_, err := (*fc.bridge).SetLightState(addrNum, state)
			if err != nil {
				return
			}
			light.State.Bri = uint8(val)
			msg := fimpgo.NewIntMessage("evt.lvl.report", "out_lvl_switch", val, nil, nil, newMsg.Payload)
			adr := &fimpgo.Address{MsgType: fimpgo.MsgTypeEvt, ResourceType: fimpgo.ResourceTypeDevice, ResourceName: ServiceName, ResourceAddress: fc.configs.InstanceAddress, ServiceName: "out_lvl_switch", ServiceAddress: newMsg.Addr.ServiceAddress}
			fc.mqt.Publish(adr,msg)
		}

	case "color_ctrl":
		if err != nil {
			return
		}
		log.Debug("Sending color cmd")
		val, err := newMsg.Payload.GetIntMapValue()
		if err != nil {
			return
		}
		light, err := (*fc.bridge).GetLight(addrNum)
		if err != nil {
			log.Errorf("Can't find light with id = %d", addrNum)
			return
		}
		hue, hok := val["hue"]
		sat, sok := val["sat"]
		if hok && sok {
			light.Hue(uint16(hue))
			light.Sat(uint8(sat))
			return
		}

		red, rok := val["red"]
		green, gok := val["green"]
		blue, bok := val["blue"]
		if rok && gok && bok {
			c := colorful.Color{R: float64(red) / 255, G: float64(green) / 255, B: float64(blue) / 255}
			x, y, _ := c.Xyy()
			err := light.Xy([]float32{float32(x), float32(y)})
			if err != nil {
				log.Errorf("Errro setting color", err.Error())
				return
			}

		}
	case "scene_ctrl":
		val, err := newMsg.Payload.GetStringValue()
		if err != nil {
			return
		}
		light, err := (*fc.bridge).GetLight(addrNum)
		if err != nil {
			log.Errorf("Can't find light with id = %d", addrNum)
			return
		}
		if val == "colorloop" {
			light.Effect(val)
		}else if val == "none" {
			light.Effect(val)
			light.Alert(val)
		}else {
			light.Alert(val)
		}


	case ServiceName:
		adr := &fimpgo.Address{MsgType: fimpgo.MsgTypeEvt, ResourceType: fimpgo.ResourceTypeAdapter, ResourceName: ServiceName, ResourceAddress: "1"}
		log.Debug("New payload type ", newMsg.Payload.Type)
		switch newMsg.Payload.Type {
		case "cmd.system.sync":
			lights, err := (*fc.bridge).GetLights()
			if err != nil {
				return
			}
			for _, l := range lights {
				fc.netService.SendInclusionReport(fmt.Sprintf("l%d", l.ID))
			}

			sensors , err := (*fc.bridge).GetSensors()
			if err != nil {
				return
			}
			for _, l := range sensors {
				fc.netService.SendInclusionReport(fmt.Sprintf("s%d", l.ID))
			}
		case "cmd.config.get_extended_report":
			msg := fimpgo.NewMessage("evt.config.extended_report",ServiceName,fimpgo.VTypeObject,fc.configs,nil,nil,newMsg.Payload)
			if err := fc.mqt.RespondToRequest(newMsg.Payload, msg); err != nil {
				fc.mqt.Publish(adr, msg)
			}
		// Deprecated API
		case "cmd.system.get_connect_params":
			val,_ := fc.discoverBridge()
			msg := fimpgo.NewStrMapMessage("evt.system.connect_params_report", ServiceName, val, nil, nil, newMsg.Payload)
			if err := fc.mqt.RespondToRequest(newMsg.Payload, msg); err != nil {
				fc.mqt.Publish(adr, msg)
			}
		// Deprecated API
		case "cmd.config.set":
			configs , err :=newMsg.Payload.GetStrMapValue()
			if err != nil {
				return
			}
			dimmerRange, _ := configs["dimmer_range_mode"]
			if dimmerRange == "100" {
				fc.configs.DimmerRangeMode = dimmerRange
				fc.configs.DimmerMaxValue = 100
			}else if dimmerRange == "255" {
				fc.configs.DimmerRangeMode = dimmerRange
				fc.configs.DimmerMaxValue = 255
			}
			fc.configs.SaveToFile()
			fc.stateMonitor.SetDimmerMaxValue(fc.configs.DimmerMaxValue)
			fc.netService.SetDimmerMaxVal(fc.configs.DimmerMaxValue)
		case "cmd.config.get_report" :
			log.Info("Dimmer max value st = %d , ns = %d",fc.stateMonitor.DimmerMaxValue(),fc.netService.DimmerMaxVal())

		case "cmd.log.set_level":
			level , err :=newMsg.Payload.GetStringValue()
			if err != nil {
				return
			}
			logLevel, err := log.ParseLevel(level)
			if err == nil {
				log.SetLevel(logLevel)
				fc.configs.LogLevel = level
				fc.configs.SaveToFile()
			}

			log.Info("Log level updated to = ",logLevel)

		case "cmd.app.get_manifest":
			mode,err := newMsg.Payload.GetStringValue()
			if err != nil {
				log.Error("Incorrect request format ")
				return
			}
			manifest := model.NewManifest()
			err = manifest.LoadFromFile(filepath.Join(fc.configs.GetDefaultDir(),"app-manifest.json"))
			if err != nil {
				log.Error("Failed to load manifest file .Error :",err.Error())
				return
			}
			if mode == "manifest_state" {
				manifest.AppState = *fc.appLifecycle.GetAllStates()
				fc.configs.ConnectionState = string(fc.appLifecycle.ConnectionState())
				fc.configs.Errors = fc.appLifecycle.LastError()
				manifest.ConfigState = fc.configs

			}
			if errConf := manifest.GetAppConfig("errors");errConf !=nil {
				if fc.configs.Errors == "" {
					errConf.Hidden = true
				}else {
					errConf.Hidden = false
				}
			}

			connectButton := manifest.GetButton("connect")
			disconnectButton := manifest.GetButton("disconnect")
			if connectButton !=nil && disconnectButton != nil {
				if fc.appLifecycle.ConnectionState() == model.ConnStateConnected {
					connectButton.Hidden = true
					disconnectButton.Hidden = false
				}else {
					connectButton.Hidden = false
					disconnectButton.Hidden = true
				}
			}
			if syncButton := manifest.GetButton("sync");syncButton !=nil {
				if fc.appLifecycle.ConnectionState() == model.ConnStateConnected {
					syncButton.Hidden = false
				}else {
					syncButton.Hidden = true
				}
			}
			connBlock := manifest.GetUIBlock("connect")
			settingsBlock := manifest.GetUIBlock("settings")
			if connBlock != nil && settingsBlock !=nil{
				if fc.configs.BridgeId != "" || fc.configs.DiscoveredBridges != "" {
					connBlock.Hidden = false
					settingsBlock.Hidden = false
				}else {
					connBlock.Hidden = true
					settingsBlock.Hidden = true
				}
			}

			msg := fimpgo.NewMessage("evt.app.manifest_report",ServiceName,fimpgo.VTypeObject,manifest,nil,nil,newMsg.Payload)
			if err := fc.mqt.RespondToRequest(newMsg.Payload,msg); err != nil {
				// if response topic is not set , sending back to default application event topic
				fc.mqt.Publish(adr,msg)
			}

		case "cmd.config.extended_set":
			conf := model.Configs{}
			err := newMsg.Payload.GetObjectValue(&conf)
			if err != nil {
				log.Error("Incorrect extended set message ")
				return
			}
			username := conf.Username
			dimmerRange := conf.DimmerRangeMode
			if dimmerRange == "100" {
				fc.configs.DimmerRangeMode = dimmerRange
				fc.configs.DimmerMaxValue = 100
			}else if dimmerRange == "255" {
				fc.configs.DimmerRangeMode = dimmerRange
				fc.configs.DimmerMaxValue = 100
			}
			fc.stateMonitor.SetDimmerMaxValue(fc.configs.DimmerMaxValue)
			fc.netService.SetDimmerMaxVal(fc.configs.DimmerMaxValue)
			if username == "" {
				username = "thingsplex"
			}
			bridgeId := conf.BridgeId
			fc.configs.BridgeId = bridgeId
			fc.configs.Username = username
			fc.configs.Host = conf.Host
			fc.configs.SaveToFile()

			configReport := model.ConfigReport{
				OpStatus: "OK",
				AppState:  *fc.appLifecycle.GetAllStates(),
			}
			fc.appLifecycle.SetConfigState(model.ConfigStateConfigured)
			msg := fimpgo.NewMessage("evt.app.config_report",ServiceName,fimpgo.VTypeObject,configReport,nil,nil,newMsg.Payload)
			if err := fc.mqt.RespondToRequest(newMsg.Payload,msg); err != nil {
				fc.mqt.Publish(adr,msg)
			}

		case "cmd.bridge.connect":
			fc.appLifecycle.SetLastError("")
			if fc.appLifecycle.ConnectionState() == model.ConnStateConnected {
				err := fc.stateMonitor.TestConnection()
				if err == nil {

					val := model.ButtonActionResponse{
						Operation:       "cmd.bridge.connect",
						OperationStatus: "ok",
						Next:            "reload",
						ErrorCode:       "E1",
						ErrorText:       "Bridge already connected",
					}
					msg := fimpgo.NewMessage("evt.app.config_action_report",ServiceName,fimpgo.VTypeObject,val,nil,nil,newMsg.Payload)
					if err := fc.mqt.RespondToRequest(newMsg.Payload,msg); err != nil {
						fc.mqt.Publish(adr,msg)
					}
					log.Warn("Already connected. Connection request skipped")
					return
				}
			}
			fc.appLifecycle.SetConnectionState(model.ConnStateConnecting)
			status,errStr := fc.connectToBridge(fc.configs.BridgeId,"thingsplex",fc.configs.Host,"full")
			val := model.ButtonActionResponse{
				Operation:       "cmd.bridge.connect",
				OperationStatus: strings.ToUpper(status),
				Next:            "reload",
				ErrorCode:       "",
				ErrorText:       errStr,
			}
			if status == "ok" {
				fc.appLifecycle.SetConnectionState(model.ConnStateConnected)
				fc.appLifecycle.SetConfigState(model.ConfigStateConfigured)
				fc.appLifecycle.SetAppState(model.AppStateRunning,nil)
			}else {
				fc.appLifecycle.SetLastError(errStr)
			}
			msg := fimpgo.NewMessage("evt.app.config_action_report",ServiceName,fimpgo.VTypeObject,val,nil,nil,newMsg.Payload)
			if err := fc.mqt.RespondToRequest(newMsg.Payload,msg); err != nil {
				fc.mqt.Publish(adr,msg)
			}

		case "cmd.bridge.disconnect":
			val := model.ButtonActionResponse{
				Operation:       "cmd.bridge.disconnect",
				OperationStatus: "OK",
				Next:            "reload",
				ErrorCode:       "",
				ErrorText:       "",
			}
			fc.stateMonitor.Stop()
			fc.configs.LoadDefaults()
			fc.appLifecycle.SetConnectionState(model.ConnStateDisconnected)
			fc.appLifecycle.SetConfigState(model.ConfigStateNotConfigured)
			fc.appLifecycle.SetAppState(model.AppStateNotConfigured,nil)
			msg := fimpgo.NewMessage("evt.app.config_action_report",ServiceName,fimpgo.VTypeObject,val,nil,nil,newMsg.Payload)
			if err := fc.mqt.RespondToRequest(newMsg.Payload,msg); err != nil {
				fc.mqt.Publish(adr,msg)
			}
			//TODO : send exclusion reports

		case "cmd.bridge.discover":
			discoveryReport,err := fc.discoverBridge()
			status := "OK"
			errStr := ""
			if err !=nil {
				status = "ERROR"
				errStr = err.Error()
			}else {
				fc.configs.DiscoveredBridges,_ = discoveryReport["discovered"]
				fc.configs.Host,_ = discoveryReport["host"]
				fc.configs.BridgeId , _ = discoveryReport["bridge_id"]
			}
			val := model.ButtonActionResponse{
				Operation:       "cmd.bridge.connect",
				OperationStatus: status,
				Next:            "reload",
				ErrorCode:       "",
				ErrorText:       errStr,
			}
			msg := fimpgo.NewMessage("evt.app.config_action_report",ServiceName,fimpgo.VTypeObject,val,nil,nil,newMsg.Payload)
			if err := fc.mqt.RespondToRequest(newMsg.Payload,msg); err != nil {
				fc.mqt.Publish(adr,msg)
			}
		// Deprecated old API
		case "cmd.system.connect":
			if fc.appLifecycle.ConnectionState() == model.ConnStateConnected {
				err := fc.stateMonitor.TestConnection()
				if err == nil {
					val := map[string]string{"status": "ok", "error": "already connected"}
					msg := fimpgo.NewStrMapMessage("evt.system.connect_report", ServiceName, val, nil, nil, newMsg.Payload)
					if err := fc.mqt.RespondToRequest(newMsg.Payload, msg); err != nil {
						fc.mqt.Publish(adr, msg)
					}
					log.Warn("Already connected. Connection request skipped")
					return
				}
			}
			reqVal, err := newMsg.Payload.GetStrMapValue()
			var errStr string
			status := "ok"
			if err != nil {
				log.Error("Incorrect login message ")
				errStr = err.Error()
			}

			host, _ := reqVal["host"]
			username, _ := reqVal["username"]
			syncMode, _ := reqVal["sync_mode"]
			dimmerRange, _ := reqVal["dimmer_range_mode"]
			if dimmerRange == "100" {
				fc.configs.DimmerRangeMode = dimmerRange
				fc.configs.DimmerMaxValue = 100
			}else if dimmerRange == "255" {
				fc.configs.DimmerRangeMode = dimmerRange
				fc.configs.DimmerMaxValue = 100
			}
			fc.stateMonitor.SetDimmerMaxValue(fc.configs.DimmerMaxValue)
			fc.netService.SetDimmerMaxVal(fc.configs.DimmerMaxValue)
			if username == "" {
				username = "thingsplex"
			}
			bridgeId, ok := reqVal["bridge_id"]
			if !ok {
				log.Error("Incorrect bridge id")
			}

			status,errStr = fc.connectToBridge(bridgeId,username,host,syncMode)

			val := map[string]string{"status": status, "error": errStr}
			msg := fimpgo.NewStrMapMessage("evt.system.connect_report", ServiceName, val, nil, nil, newMsg.Payload)
			if err := fc.mqt.RespondToRequest(newMsg.Payload, msg); err != nil {
				fc.mqt.Publish(adr, msg)
			}

		case "cmd.network.get_all_nodes":
			fc.netService.SendListOfDevices()
		case "cmd.thing.get_inclusion_report":
			nodeId, _ := newMsg.Payload.GetStringValue()
			fc.netService.SendInclusionReport(nodeId)

		case "cmd.thing.inclusion":
			flag, _ := newMsg.Payload.GetBoolValue()
			if flag && fc.isInclusionRunning {
				log.Info("One inclusion is alredy running")
				return
			}
			if !flag {
				fc.isInclusionRunning = false
				return
			}

			res, err := (*fc.bridge).FindLights()
			if err != nil {
				log.Error("Failed to start lights inclusion ", err)
				return
			}
			log.Debug("inclusion start response", res.Success)
			fc.isInclusionRunning = true
			go func() {
				var c int
				defer func() {
					fc.isInclusionRunning = false
				}()
				for {
					lights, err := (*fc.bridge).GetNewLights()
					if err != nil {
						log.Debug("Nothing found err :", err)
						return
					}
					for _, l := range lights.Lights {
						log.Infof("Discovered light %s", l)
						fc.netService.SendInclusionReport("l" + l)
					}
					if len(lights.Lights) > 0 {
						log.Infof("Inclusion monitor just quit")
						return
					}
					time.Sleep(2 * time.Second)
					c++
					if c > 10 || !fc.isInclusionRunning {
						log.Infof("Inclusion monitor just quit")
						return
					}

				}
			}()
		case "cmd.thing.delete":
			// remove device from network
			val, err := newMsg.Payload.GetStrMapValue()
			if err != nil {
				log.Error("Wrong msg format")
				return
			}
			deviceId, ok := val["address"]
			deviceId = strings.Replace(deviceId, "l", "", 1)
			if ok {
				addr, err := strconv.Atoi(deviceId)
				if err != nil {
					return
				}
				err = (*fc.bridge).DeleteLight(addr)
				if err != nil {
					log.Error("Failed to delete resource ", err)
					return
				}
				exclReport := map[string]string{"address": deviceId}
				msg := fimpgo.NewMessage("evt.thing.exclusion_report", ServiceName, fimpgo.VTypeObject, exclReport, nil, nil, nil)
				adr := fimpgo.Address{MsgType: fimpgo.MsgTypeEvt, ResourceType: fimpgo.ResourceTypeAdapter, ResourceName: "hue", ResourceAddress: "1"}
				fc.mqt.Publish(&adr, msg)
				log.Info(deviceId)
			} else {
				log.Error("Incorrect address")

			}

		case "cmd.state.get_full_report":
			val := map[string]string{"app":string(fc.appLifecycle.AppState()),"connection":string(fc.appLifecycle.ConnectionState()),"last_err":fc.appLifecycle.LastError()}
			msg := fimpgo.NewStrMapMessage("evt.state.full_report",ServiceName,val,nil,nil,newMsg.Payload)
			if err := fc.mqt.RespondToRequest(newMsg.Payload,msg); err != nil {
				fc.mqt.Publish(adr,msg)
			}

		}
	}
}

func (fc *FromFimpRouter) connectToBridge(bridgeId,username,host,syncMode string)(status string , errStr string) {
	br , err := huego.DiscoverAll()
	var found bool
	for _,b := range br {
		if b.ID == bridgeId {
			*fc.bridge = &b
			found = true
		}
	}

	if found {
		var token string
		token, err = (*fc.bridge).CreateUser(username)
		if err != nil {
			log.Error("Can't create user bridge ", err)
			status = "error"
			errStr = err.Error()
		} else {
			log.Info("User added", token)
			fc.configs.Token = token
			fc.configs.BridgeId = bridgeId
			fc.appLifecycle.PublishEvent(model.EventConfigured, "from-fimp-router", nil)
			fc.configs.SaveToFile()
		}
		log.Debugf("%s,%s", host, token)

	}else {
		status = "error"
		errStr = "hue bridge can't be discovered"
	}

	if err != nil {
		status = "error"
	}else {
		if syncMode == "full" || syncMode == "lights" {
			initOk := false
			for i:=1 ; i<4 ; i++ {
				lights, err := (*fc.bridge).GetLights()
				if err != nil {
					errStr = err.Error()
					status = "error"
				}else {
					initOk = true
					for _, l := range lights {
						fc.netService.SendInclusionReport(fmt.Sprintf("l%d", l.ID))
					}
				}
				sensors, err := (*fc.bridge).GetSensors()
				if err != nil {
					errStr = err.Error()
					status = "error"
				}else {
					initOk = true
					for _, l := range sensors {
						fc.netService.SendInclusionReport(fmt.Sprintf("s%d", l.ID))
					}
				}

				if initOk {
					log.Info(" --- Hue bridge connected successfully -----")
					status = "ok"
					errStr = ""
					break
				}else {
					log.Info(" --- connection attempt failed . err :",errStr)
					time.Sleep(time.Second*time.Duration(i*2))
				}
			}


		}
	}
	if status == "ok" {
		fc.appLifecycle.SetConnectionState(model.ConnStateConnected)
	}else {
		fc.appLifecycle.SetConnectionState(model.ConnStateDisconnected)
		fc.appLifecycle.SetLastError(errStr)
	}
	return
}

func (fc *FromFimpRouter) discoverBridge() (map[string]string,error) {
	br,err := huego.DiscoverAll()
	if err != nil {
		return nil, err
	}
	var discoverdIds , discoveredIps,bridgeId string
	if len(br)==1 {
		discoverdIds = br[0].ID
		discoveredIps = br[0].Host
		bridgeId = br[0].ID
	}else {
		for _,b := range br {
			discoverdIds = discoverdIds+","+b.ID
			discoveredIps = discoveredIps+","+b.Host
		}
	}
	return map[string]string{"host": discoveredIps, "username": "thingsplex", "sync_mode": "lights","bridge_id":bridgeId,"discovered":discoverdIds,
		"instructions":"press hue link button first","dimmer_range_mode":fc.configs.DimmerRangeMode},nil
}