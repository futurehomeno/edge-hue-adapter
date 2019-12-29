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
	instanceId   string
}

func NewStateMonitor(mqt *fimpgo.MqttTransport, bridge **huego.Bridge,instanceId string) *StateMonitor {
	st := &StateMonitor{mqt: mqt, bridge: bridge,interval:2}
	st.instanceId = instanceId
	st.sensorStates = map[int]huego.Sensor{}
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

func (st *StateMonitor) SetPoolingInterval(interval time.Duration) {
	st.interval = interval
}

func (st *StateMonitor) monitor() {
	for {
		  time.Sleep(time.Second*st.interval)
		  sensors,err := (*st.bridge).GetSensors()
		  if len(sensors)==0 {
		  	time.Sleep(time.Second*60)
		  }
		  if err != nil {
			time.Sleep(time.Second*30)
		  	continue
		  }
		  for i := range sensors {
			  old , _ := st.sensorStates[sensors[i].ID]
			  if st.isUpdated(&old,&sensors[i]) {
			  	  buttonEvent, butOk := sensors[i].State["buttonevent"]
			  	  if butOk {
			  	  	buttonEventCode,ok := buttonEvent.(float64)
			  	  	if ok {
						log.Debugf("Sensor name = %s , type = %s , state = %+v , config = %+v",sensors[i].Name,sensors[i].Type,sensors[i].State,sensors[i].Config)
						st.sendButtonReport(buttonEventCode,sensors[i].ID)
					}
				  }
				  presenceEvent, butOk := sensors[i].State["presence"]
				  if butOk {
					  presenceState,ok := presenceEvent.(bool)
					  if ok {
						  //log.Debugf("Sensor name = %s , type = %s , state = %+v , config = %+v",sensors[i].Name,sensors[i].Type,sensors[i].State,sensors[i].Config)
						  st.sendPresenceReport(presenceState,sensors[i].ID)
					  }
				  }
				  //log.Debugf("%+v",sensors[i])
				  st.sensorStates[sensors[i].ID] = sensors[i]
			  }
		  }
		  if !st.isRunnning {
		  	return
		  }
	}
}

func (st *StateMonitor) isUpdated(old , new *huego.Sensor)bool {
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


