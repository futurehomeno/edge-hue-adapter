package router

import (
	"fmt"
	"github.com/amimof/huego"
	"github.com/futurehomeno/fimpgo"
	"github.com/lucasb-eyer/go-colorful"
	log "github.com/sirupsen/logrus"
	"github.com/thingsplex/hue-ad/model"
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
			adr := &fimpgo.Address{MsgType: fimpgo.MsgTypeEvt, ResourceType: fimpgo.ResourceTypeDevice, ResourceName: "hue", ResourceAddress: fc.configs.InstanceAddress, ServiceName: "out_lvl_switch", ServiceAddress: newMsg.Addr.ServiceAddress}
			fc.mqt.Publish(adr,msg)
			//log.Debug("Status code = ",respH.StatusCode)
		case "cmd.lvl.set":
			val, _ := newMsg.Payload.GetIntValue()
			light, _ := (*fc.bridge).GetLight(addrNum)
			//light.Bri(uint8(val))
			state := huego.State{On: true,Bri: uint8(val),TransitionTime:transitionTime}
			_, err := (*fc.bridge).SetLightState(addrNum, state)
			if err != nil {
				return
			}
			light.State.Bri = uint8(val)
			msg := fimpgo.NewIntMessage("evt.lvl.report", "out_lvl_switch", val, nil, nil, newMsg.Payload)
			adr := &fimpgo.Address{MsgType: fimpgo.MsgTypeEvt, ResourceType: fimpgo.ResourceTypeDevice, ResourceName: "hue", ResourceAddress: fc.configs.InstanceAddress, ServiceName: "out_lvl_switch", ServiceAddress: newMsg.Addr.ServiceAddress}
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
		case "cmd.auth.login":
			reqVal, err := newMsg.Payload.GetStrMapValue()
			status := "ok"
			if err != nil {
				log.Error("Incorrect login message ")
				return
			}
			username, _ := reqVal["username"]
			password, _ := reqVal["password"]
			if username != "" && password != "" {

			}
			fc.configs.SaveToFile()
			if err != nil {
				status = "error"
			}
			msg := fimpgo.NewStringMessage("evt.system.login_report", ServiceName, status, nil, nil, newMsg.Payload)
			if err := fc.mqt.RespondToRequest(newMsg.Payload, msg); err != nil {
				fc.mqt.Publish(adr, msg)
			}
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

		case "cmd.system.get_connect_params":
			br,_ := huego.DiscoverAll()
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
			val := map[string]string{"host": discoveredIps, "username": "thingsplex", "sync_mode": "lights","bridge_id":bridgeId,"discovered":discoverdIds,"instructions":"press hue link button first"}
			msg := fimpgo.NewStrMapMessage("evt.system.connect_params_report", ServiceName, val, nil, nil, newMsg.Payload)
			if err := fc.mqt.RespondToRequest(newMsg.Payload, msg); err != nil {
				fc.mqt.Publish(adr, msg)
			}
		case "cmd.config.set":
			fallthrough
		case "cmd.system.connect":
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
			if username == "" {
				username = "thingsplex"
			}
			bridgeId, ok := reqVal["bridge_id"]
			if !ok {
				log.Error("Incorrect bridge id")
			}

			br , err := huego.DiscoverAll()
			var found bool
			for _,b := range br {
				if b.ID == bridgeId {
					*fc.bridge = &b
					found = true
				}
			}

			if found {
				token, err := (*fc.bridge).CreateUser(username)
				if err != nil {
					log.Error("Can't create user bridge ", err)
					errStr = err.Error()
				} else {
					log.Info("User added", token)
					fc.configs.Token = token
					fc.configs.BridgeId = bridgeId
					fc.appLifecycle.PublishEvent(model.EventConfigured, "from-fimp-router", nil)
				}
				log.Debugf("%s,%s", host, token)

				fc.configs.SaveToFile()
			}else {
				status = "error"
				errStr = "hue bridge can't be discovered"
			}

			if err != nil {
				status = "error"
			}
			val := map[string]string{"status": status, "error": errStr}
			msg := fimpgo.NewStrMapMessage("evt.system.connect_report", ServiceName, val, nil, nil, newMsg.Payload)
			if err := fc.mqt.RespondToRequest(newMsg.Payload, msg); err != nil {
				fc.mqt.Publish(adr, msg)
			}

			if syncMode == "full" || syncMode == "lights" {
				lights, err := (*fc.bridge).GetLights()
				if err != nil {
					return
				}
				for _, l := range lights {
					fc.netService.SendInclusionReport(fmt.Sprintf("l%d", l.ID))
				}

				sensors, err := (*fc.bridge).GetSensors()
				if err != nil {
					return
				}
				for _, l := range sensors {
					fc.netService.SendInclusionReport(fmt.Sprintf("s%d", l.ID))
				}

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
				msg := fimpgo.NewMessage("evt.thing.exclusion_report", "hue", fimpgo.VTypeObject, exclReport, nil, nil, nil)
				adr := fimpgo.Address{MsgType: fimpgo.MsgTypeEvt, ResourceType: fimpgo.ResourceTypeAdapter, ResourceName: "hue", ResourceAddress: "1"}
				fc.mqt.Publish(&adr, msg)
				log.Info(deviceId)
			} else {
				log.Error("Incorrect address")

			}

		}
		//

	}

}
