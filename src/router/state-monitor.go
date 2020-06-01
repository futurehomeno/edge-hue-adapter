package router

import (
	"fmt"
	"github.com/amimof/huego"
	"github.com/futurehomeno/fimpgo"
	log "github.com/sirupsen/logrus"
	"strconv"
	"time"
)

type StateMonitor struct {
	mqt          *fimpgo.MqttTransport
	bridge       **huego.Bridge
	interval     time.Duration // pooling interval
	isRunnning   bool
	sensorStates map[int]huego.Sensor
	lightsStates map[int]huego.Light

	batteryLevels map[int]float64
	instanceId   string
	dimmerMaxValue int
}

func (st *StateMonitor) DimmerMaxValue() int {
	return st.dimmerMaxValue
}


func (st *StateMonitor) SetDimmerMaxValue(dimmerMaxValue int) {
	st.dimmerMaxValue = dimmerMaxValue
}

func NewStateMonitor(mqt *fimpgo.MqttTransport, bridge **huego.Bridge,instanceId string) *StateMonitor {
	st := &StateMonitor{mqt: mqt, bridge: bridge,interval:2}
	st.instanceId = instanceId
	st.sensorStates = map[int]huego.Sensor{}
	st.lightsStates = map[int]huego.Light{}
	st.batteryLevels = map[int]float64{}
	st.dimmerMaxValue = 255
	return st

}

func (st *StateMonitor) Start() {
	if !st.isRunnning {
		go st.monitor()
	}
	st.isRunnning = true
}
func (st *StateMonitor) Stop() {
	st.isRunnning = false
}

func (st *StateMonitor) TestConnection ()error {
	_, err := (*st.bridge).GetLights()
	return err
}

func (st *StateMonitor) SetPoolingInterval(interval time.Duration) {
	st.interval = interval
}

func (st *StateMonitor) monitor() {
	var lc int

	for {
		  time.Sleep(time.Second*st.interval)
		  lc++
		  sensors,err := (*st.bridge).GetSensors()

		  if err != nil {
			log.Debugf("Adapter can't get sensors from hue hub . Err ",err)
		  	time.Sleep(time.Second*30)
			continue
		  }
		  st.processSensors(sensors)
		  var lights []huego.Light
		  var lerr error
		  if lc > 3 {
		  	// check lights only every 3rd time
			lights, lerr = (*st.bridge).GetLights()
		  	lc = 0
			if lerr != nil {
				  log.Debugf("Adapter can't get lights from hue hub . Err ",lerr)
				  time.Sleep(time.Second*30)
				  continue
			}
			st.processLights(lights)
		  }

		  if len(sensors)==0 && len(lights)==0{
				time.Sleep(time.Second*30)
		  }

		  if !st.isRunnning {
		  	return
		  }
	}
}

func (st *StateMonitor) processSensors(sensors []huego.Sensor) {
	for i := range sensors {
		old , _ := st.sensorStates[sensors[i].ID]
		if st.isSensorUpdated(&old,&sensors[i]) {
			buttonEvent, butOk := sensors[i].State["buttonevent"]
			if butOk {
				buttonEventCode,ok := buttonEvent.(float64)
				if ok {
					log.Debugf("Sensor name = %s , type = %s , state = %+v ",sensors[i].Name,sensors[i].Type,sensors[i].State)
					st.sendButtonReport(buttonEventCode,sensors[i].ID)
				}
			}

			presenceEvent, butOk := sensors[i].State["presence"]
			if butOk {
				presenceState,ok := presenceEvent.(bool)
				if ok {
					st.sendPresenceReport(presenceState,sensors[i].ID)
				}
			}
			temperature, tempOk := sensors[i].State["temperature"]
			if tempOk {
				tempT,ok := temperature.(float64)
				if !ok {
					continue
				}
				tempT = tempT/100
				st.sendTemperatureReport(tempT,sensors[i].ID)
				//log.Debug("Temperature: ",tempT)
			}
			lightLvl, lightOk := sensors[i].State["lightlevel"]
			if lightOk {
				lightT,ok := lightLvl.(float64)
				if !ok {
					continue
				}
				st.sendLightLevelReport(lightT,sensors[i].ID)
				//log.Debug("Light level: ",lightT)
			}

			batt , battOk := sensors[i].Config["battery"]
			if battOk{
				newBattLvl,_ := batt.(float64)
				oldBattLvl , _ := st.batteryLevels[sensors[i].ID]

				if oldBattLvl != newBattLvl {
					st.sendBatteryReport(newBattLvl,sensors[i].ID)
					st.batteryLevels[sensors[i].ID] = newBattLvl
				}
				//log.Debug("Battery level: ",battT)
			}

			st.sensorStates[sensors[i].ID] = sensors[i]
		}
	}
}

func (st *StateMonitor) processLights(lights []huego.Light) {
	for i := range lights {
		oldState , _ := st.lightsStates[lights[i].ID]
		newState := lights[i]
		servAddr := fmt.Sprintf("l%d_0",lights[i].ID)
		var isStateChanged bool
		if oldState.State == nil {
			st.lightsStates[lights[i].ID] = lights[i]
			return
		}
		if oldState.State.On != newState.State.On {
			// binary report
			msg := fimpgo.NewBoolMessage("evt.binary.report", "out_lvl_switch", newState.State.On, nil, nil, nil)
			adr := &fimpgo.Address{MsgType: fimpgo.MsgTypeEvt, ResourceType: fimpgo.ResourceTypeDevice, ResourceName: "hue", ResourceAddress: st.instanceId, ServiceName: "out_lvl_switch", ServiceAddress: servAddr}
			st.mqt.Publish(adr,msg)
			isStateChanged = true
		}
		if oldState.State.Bri != newState.State.Bri {
			// level report
			isStateChanged = true
			var val int64
			if st.dimmerMaxValue == 100 {
				val = int64((100.0/255.0)*float64(newState.State.Bri))
			}else {
				val = int64(newState.State.Bri)
			}
			msg := fimpgo.NewIntMessage("evt.lvl.report", "out_lvl_switch", val, nil, nil, nil)
			adr := &fimpgo.Address{MsgType: fimpgo.MsgTypeEvt, ResourceType: fimpgo.ResourceTypeDevice, ResourceName: "hue", ResourceAddress: st.instanceId, ServiceName: "out_lvl_switch", ServiceAddress: servAddr}
			st.mqt.Publish(adr,msg)
			isStateChanged = true
		}
		if oldState.State.Reachable != newState.State.Reachable {
			// Device state
			isStateChanged = true
		}
		if isStateChanged {
			st.lightsStates[lights[i].ID] = lights[i]
		}
	}
}


func (st *StateMonitor) isSensorUpdated(old , new *huego.Sensor)bool {
	if old == nil {
		return true
	}
	oldLastUpdated,ok1 := old.State["lastupdated"]
	newLastUpdated,ok2 := new.State["lastupdated"]
	if !ok1 || !ok2 {
		return true
	}

	if oldLastUpdated == newLastUpdated {
		return false
	}
	return true
}


func (st *StateMonitor) sendButtonReport(code float64,deviceId int) {
	valS := strconv.FormatFloat(code,'f',-1,32)
	servAddr := fmt.Sprintf("s%d_0",deviceId)
	msg := fimpgo.NewStringMessage("evt.scene.report", "scene_ctrl", valS, nil, nil, nil)
	adr := &fimpgo.Address{MsgType: fimpgo.MsgTypeEvt, ResourceType: fimpgo.ResourceTypeDevice, ResourceName: "hue", ResourceAddress: st.instanceId, ServiceName: "scene_ctrl", ServiceAddress: servAddr}
	st.mqt.Publish(adr,msg)
}

func (st *StateMonitor) sendPresenceReport(state bool,deviceId int) {
	servAddr := fmt.Sprintf("s%d_0",deviceId)
	msg := fimpgo.NewBoolMessage("evt.presence.report", "sensor_presence", state, nil, nil, nil)
	adr := &fimpgo.Address{MsgType: fimpgo.MsgTypeEvt, ResourceType: fimpgo.ResourceTypeDevice, ResourceName: "hue", ResourceAddress: st.instanceId, ServiceName: "sensor_presence", ServiceAddress: servAddr}
	st.mqt.Publish(adr,msg)
}

func (st *StateMonitor) sendTemperatureReport(temp float64,deviceId int) {
	servAddr := fmt.Sprintf("s%d_0",deviceId)
	props := map[string]string{"unit":"C"}
	msg := fimpgo.NewFloatMessage("evt.sensor.report", "sensor_temp", temp, props, nil, nil)
	adr := &fimpgo.Address{MsgType: fimpgo.MsgTypeEvt, ResourceType: fimpgo.ResourceTypeDevice, ResourceName: "hue", ResourceAddress: st.instanceId, ServiceName: "sensor_temp", ServiceAddress: servAddr}
	st.mqt.Publish(adr,msg)
}

func (st *StateMonitor) sendLightLevelReport(lvl float64,deviceId int) {
	servAddr := fmt.Sprintf("s%d_0",deviceId)
	props := map[string]string{"unit":"Lux"}
	msg := fimpgo.NewFloatMessage("evt.sensor.report", "sensor_lumin",lvl, props, nil, nil)
	adr := &fimpgo.Address{MsgType: fimpgo.MsgTypeEvt, ResourceType: fimpgo.ResourceTypeDevice, ResourceName: "hue", ResourceAddress: st.instanceId, ServiceName: "sensor_lumin", ServiceAddress: servAddr}
	st.mqt.Publish(adr,msg)
}

func (st *StateMonitor) sendBatteryReport(lvl float64,deviceId int) {
	servAddr := fmt.Sprintf("s%d_0",deviceId)
	msg := fimpgo.NewIntMessage("evt.lvl.report", "battery",int64(lvl), nil, nil, nil)
	adr := &fimpgo.Address{MsgType: fimpgo.MsgTypeEvt, ResourceType: fimpgo.ResourceTypeDevice, ResourceName: "hue", ResourceAddress: st.instanceId, ServiceName: "battery", ServiceAddress: servAddr}
	st.mqt.Publish(adr,msg)
}