package model

import (
	"fmt"
	"github.com/amimof/huego"
	"github.com/futurehomeno/fimpgo"
	"github.com/futurehomeno/fimpgo/fimptype"
	log "github.com/sirupsen/logrus"
	"strconv"
	"strings"
)

type ListReportRecord struct {
	Address     string `json:"address"`
	Alias       string `json:"alias"`
	PowerSource string `json:"power_source"`
	Hash        string `json:"hash"`
}

type OpResponse struct {

}

type NetworkService struct {
	mqt          *fimpgo.MqttTransport
	bridge       **huego.Bridge
	dimmerMaxVal int
}

func (ns *NetworkService) DimmerMaxVal() int {
	return ns.dimmerMaxVal
}

func (ns *NetworkService) SetDimmerMaxVal(dimmerMaxVal int) {
	ns.dimmerMaxVal = dimmerMaxVal
}

func NewNetworkService(mqt *fimpgo.MqttTransport, bridge  **huego.Bridge) *NetworkService {
	return &NetworkService{mqt: mqt, bridge:bridge, dimmerMaxVal:255}
}

func (ns *NetworkService) OpenNetwork(open bool) error {

	return nil
}

func (ns *NetworkService) DeleteThing(deviceId string) error {

	exclReport := map[string]string{"address":deviceId}
	msg := fimpgo.NewMessage("evt.thing.exclusion_report", "conbee", fimpgo.VTypeObject, exclReport, nil, nil, nil)
	adr := fimpgo.Address{MsgType: fimpgo.MsgTypeEvt, ResourceType: fimpgo.ResourceTypeAdapter, ResourceName: "conbee", ResourceAddress: "1"}
	ns.mqt.Publish(&adr, msg)

	return nil

}

func (ns *NetworkService) SendInclusionReport(nodeId string) error {
	var deviceId int
	var deviceType string
	var err error
	if strings.Contains(nodeId,"l"){
		nodeId = strings.Replace(nodeId,"l","",1)
		deviceId , err= strconv.Atoi(nodeId)
		if err != nil {
			return err
		}
		deviceType = "lights"
	}else {
		nodeId = strings.Replace(nodeId,"s","",1)
		deviceId , err= strconv.Atoi(nodeId)
		if err != nil {
			return err
		}
		deviceType = "sensors"
	}


	var productId,name, manufacturer, powerSource, swVersion, serialNr string
	var deviceAddr string
	services := []fimptype.Service{}

	outLvlSwitchInterfaces := []fimptype.Interface{{
		Type:      "in",
		MsgType:   "cmd.binary.set",
		ValueType: "bool",
		Version:   "1",
	}, {
		Type:      "in",
		MsgType:   "cmd.lvl.set",
		ValueType: "int",
		Version:   "1",
	}, {
		Type:      "in",
		MsgType:   "cmd.lvl.start",
		ValueType: "string",
		Version:   "1",
	}, {
		Type:      "in",
		MsgType:   "cmd.lvl.stop",
		ValueType: "null",
		Version:   "1",
	}, {
		Type:      "out",
		MsgType:   "evt.lvl.report",
		ValueType: "int",
		Version:   "1",
	}, {
		Type:      "out",
		MsgType:   "evt.binary.report",
		ValueType: "bool",
		Version:   "1",
	}}

	batteryInterfaces := []fimptype.Interface{{
		Type:      "in",
		MsgType:   "cmd.lvl.get_report",
		ValueType: "null",
		Version:   "1",
	}, {
		Type:      "out",
		MsgType:   "evt.lvl.report",
		ValueType: "int",
		Version:   "1",
	}, {
		Type:      "out",
		MsgType:   "evt.alarm.report",
		ValueType: "str_map",
		Version:   "1",
	}}
	//
	//
	presenceSensorInterfaces := []fimptype.Interface{{
		Type:      "in",
		MsgType:   "cmd.presence.get_report",
		ValueType: "null",
		Version:   "1",
	}, {
		Type:      "out",
		MsgType:   "evt.presence.report",
		ValueType: "bool",
		Version:   "1",
	}}
	//
	//contactSensorInterfaces := []fimptype.Interface{{
	//	Type:      "in",
	//	MsgType:   "cmd.open.get_report",
	//	ValueType: "null",
	//	Version:   "1",
	//}, {
	//	Type:      "out",
	//	MsgType:   "evt.open.report",
	//	ValueType: "bool",
	//	Version:   "1",
	//}}
	//
	sensorInterfaces := []fimptype.Interface{{
		Type:      "in",
		MsgType:   "cmd.sensor.get_report",
		ValueType: "string",
		Version:   "1",
	}, {
		Type:      "out",
		MsgType:   "evt.sensor.report",
		ValueType: "float",
		Version:   "1",
	}}
	sceneInterfaces := []fimptype.Interface{{
		Type:      "in",
		MsgType:   "cmd.scene.set",
		ValueType: "string",
		Version:   "1",
	},{
		Type:      "in",
		MsgType:   "cmd.scene.get_report",
		ValueType: "null",
		Version:   "1",
	}, {
		Type:      "out",
		MsgType:   "evt.scene.report",
		ValueType: "string",
		Version:   "1",
	}}

	colorInterfaces := []fimptype.Interface{{
		Type:      "in",
		MsgType:   "cmd.color.set",
		ValueType: "int_map",
		Version:   "1",
	},{
		Type:      "in",
		MsgType:   "cmd.color.get_report",
		ValueType: "null",
		Version:   "1",
	}, {
		Type:      "out",
		MsgType:   "evt.color.report",
		ValueType: "int_map",
		Version:   "1",
	}}



	outLvlSwitchService := fimptype.Service{
		Name:    "out_lvl_switch",
		Alias:   "Light control",
		Address: "/rt:dev/rn:hue/ad:1/sv:out_lvl_switch/ad:",
		Enabled: true,
		Groups:  []string{"ch_0"},
		Props: map[string]interface{}{
			"max_lvl": ns.dimmerMaxVal,
			"min_lvl": 0,
		},
		Tags:             nil,
		PropSetReference: "",
		Interfaces:       outLvlSwitchInterfaces,
	}

	batteryService := fimptype.Service{
		Name:    "battery",
		Alias:   "battery",
		Address: "/rt:dev/rn:hue/ad:1/sv:battery/ad:",
		Enabled: true,
		Groups:  []string{"ch_0"},
		Props: map[string]interface{}{},
		Tags:             nil,
		PropSetReference: "",
		Interfaces:       batteryInterfaces,
	}

	tempSensorService := fimptype.Service{
		Name:    "sensor_temp",
		Alias:   "Temperature sensor",
		Address: "/rt:dev/rn:hue/ad:1/sv:sensor_temp/ad:",
		Enabled: true,
		Groups:  []string{"ch_0"},
		Props: map[string]interface{}{
			"sup_units": []string{"C"},
		},
		Tags:             nil,
		PropSetReference: "",
		Interfaces:       sensorInterfaces,
	}

	luminSensorService := fimptype.Service{
		Name:    "sensor_lumin",
		Alias:   "Light level sensor",
		Address: "/rt:dev/rn:hue/ad:1/sv:sensor_lumin/ad:",
		Enabled: true,
		Groups:  []string{"ch_0"},
		Props: map[string]interface{}{
			"sup_units": []string{"Lux"},
		},
		Tags:             nil,
		PropSetReference: "",
		Interfaces:       sensorInterfaces,
	}

	sceneService := fimptype.Service{
		Name:    "scene_ctrl",
		Alias:   "Alert scene",
		Address: "/rt:dev/rn:hue/ad:1/sv:scene_ctrl/ad:",
		Enabled: true,
		Groups:  []string{"ch_0"},
		Props: map[string]interface{}{"sup_scenes":[]string{"none","select","lselect","colorloop"}},
		Tags:             nil,
		PropSetReference: "",
		Interfaces:       sceneInterfaces,
	}

	colorService := fimptype.Service{
		Name:    "color_ctrl",
		Alias:   "Color control",
		Address: "/rt:dev/rn:hue/ad:1/sv:color_ctrl/ad:",
		Enabled: true,
		Groups:  []string{"ch_0"},
		Props: map[string]interface{}{"sup_components":[]string{"hue","sat","red","green","blue"}},
		Tags:             nil,
		PropSetReference: "",
		Interfaces:       colorInterfaces,
	}

	//
	//contactService := fimptype.Service{
	//	Name:    "sensor_contact",
	//	Alias:   "Door/window contact",
	//	Address: "/rt:dev/rn:hue/ad:1/sv:sensor_contact/ad:",
	//	Enabled: true,
	//	Groups:  []string{"ch_0"},
	//	Props: map[string]interface{}{},
	//	Tags:             nil,
	//	PropSetReference: "",
	//	Interfaces:       contactSensorInterfaces,
	//}

	presenceService := fimptype.Service{
		Name:    "sensor_presence",
		Alias:   "Motion sensor",
		Address: "/rt:dev/rn:hue/ad:1/sv:sensor_presence/ad:",
		Enabled: true,
		Groups:  []string{"ch_0"},
		Props: map[string]interface{}{},
		Tags:             nil,
		PropSetReference: "",
		Interfaces:       presenceSensorInterfaces,
	}

	if deviceType == "lights" {
		    l,_ := (*ns.bridge).GetLight(deviceId)
			productId = l.ProductID
			manufacturer = l.ManufacturerName
			swVersion = l.SwVersion
			serialNr = l.UniqueID
			name = l.Name
			serviceAddres := fmt.Sprintf("l%d_0",deviceId)
			outLvlSwitchService.Address = outLvlSwitchService.Address + serviceAddres
			sceneService.Address = sceneService.Address+serviceAddres
			colorService.Address = colorService.Address+serviceAddres
			services = append(services,outLvlSwitchService,sceneService,colorService)
		    deviceAddr = fmt.Sprintf("l%d",deviceId)
			powerSource = "ac"
		}else if deviceType == "sensors" {
			l,_ := (*ns.bridge).GetSensor(deviceId)
			productId = l.ModelID
			manufacturer = l.ManufacturerName
			swVersion = l.SwVersion
			serialNr = l.UniqueID
			name = l.Name
			serviceAddress := fmt.Sprintf("s%d_0",deviceId)
			batteryService.Address = batteryService.Address+ serviceAddress
		    services = append(services,batteryService)
			if l.Type == "ZLLSwitch" {
				sceneService.Props["sup_scenes"] = []string{}
				sceneService.Address = sceneService.Address+ serviceAddress
				services = append(services,sceneService)
			}
			if l.Type == "ZLLPresence" {
				presenceService.Address = presenceService.Address+ serviceAddress
				services = append(services,presenceService)
			}
			if l.Type == "ZLLTemperature" {
				tempSensorService.Address = tempSensorService.Address+ serviceAddress
				services = append(services,tempSensorService)
			}
			if l.Type == "ZLLLightLevel" {
				luminSensorService.Address = luminSensorService.Address+ serviceAddress
				services = append(services,luminSensorService)
			}
			deviceAddr = fmt.Sprintf("s%d",deviceId)
			powerSource = "battery"
		}



	//}
	//if deviceType == "sensors" {
	//	sensorDeviceDescriptor := conbee.Sensor{}
	//	_ , err := ns.conbeeClient.SendConbeeRequest("GET", "sensors/"+deviceId, nil, &sensorDeviceDescriptor)
	//	if err != nil {
	//		log.Error("Can't get device descriptor . Err :", err)
	//		return err
	//	}
	//	productId = sensorDeviceDescriptor.Modelid
	//	manufacturer = sensorDeviceDescriptor.Manufacturername
	//	swVersion = sensorDeviceDescriptor.Swversion
	//	serialNr = sensorDeviceDescriptor.Uniqueid
	//
	//	serviceAddres := "s"+deviceId+"_0"
	//	batteryService.Address = batteryService.Address + serviceAddres
	//	services = append(services,batteryService)
	//
	//	switch sensorDeviceDescriptor.Type {
	//	case "ZHATemperature":
	//		tempSensorService.Address = tempSensorService.Address + serviceAddres
	//		services = append(services,tempSensorService)
	//	case "ZHAHumidity":
	//		humidSensorService.Address = humidSensorService.Address + serviceAddres
	//		services = append(services,humidSensorService)
	//	case "ZHASwitch":
	//		sceneService.Address = sceneService.Address + serviceAddres
	//		services = append(services,sceneService)
	//	case "ZHAOpenClose":
	//		contactService.Address = contactService.Address + serviceAddres
	//		services = append(services,contactService)
	//	case "ZHAPresence":
	//		presenceService.Address = presenceService.Address + serviceAddres
	//		services = append(services,presenceService)
	//
	//	}
	//	powerSource = "battery"
	//	deviceId = "s"+deviceId
	//}

	inclReport := fimptype.ThingInclusionReport{
		IntegrationId:     "",
		Address:           deviceAddr,
		Type:              "",
		ProductHash:       manufacturer + "_" + productId,
		Alias:             productId,
		CommTechnology:    "hue",
		ProductId:         productId,
		ProductName:       name,
		ManufacturerId:    manufacturer,
		DeviceId:          serialNr,
		HwVersion:         "1",
		SwVersion:         swVersion,
		PowerSource:       powerSource,
		WakeUpInterval:    "-1",
		Security:          "",
		Tags:              nil,
		Groups:            []string{"ch_0"},
		PropSets:          nil,
		TechSpecificProps: nil,
		Services:          services,
	}

	msg := fimpgo.NewMessage("evt.thing.inclusion_report", "hue", fimpgo.VTypeObject, inclReport, nil, nil, nil)
	adr := fimpgo.Address{MsgType: fimpgo.MsgTypeEvt, ResourceType: fimpgo.ResourceTypeAdapter, ResourceName: "hue", ResourceAddress: "1"}
	ns.mqt.Publish(&adr, msg)
	return nil

}

func (ns *NetworkService) SendListOfDevices() error {

	report := []ListReportRecord{}

	lights,_ := (*ns.bridge).GetLights()
	for _,l := range lights {
		rec := ListReportRecord{Address:fmt.Sprintf("l%d",l.ID) ,Alias:l.ManufacturerName+" "+l.ModelID+" "+l.Name}
		report = append(report,rec)
	}
	sensors,_ := (*ns.bridge).GetSensors()
	for _,s := range sensors {
		if s.Type == "ZLLPresence" || s.Type == "ZLLTemperature" || s.Type == "ZLLLightLevel" || s.Type == "ZLLSwitch" {
			rec := ListReportRecord{Address:fmt.Sprintf("s%d",s.ID),Alias:" "+s.ManufacturerName+" "+s.ModelID+" "+s.Name}
			report = append(report,rec)
		}else {
			log.Debugf("Unsupported type %s , id = %d",s.Type,s.ID)
		}

	}
	//groups,_ := (*ns.bridge).GetGroups()
	//for _,s := range groups {
	//	rec := ListReportRecord{Address:fmt.Sprintf("g%d",s.ID),Alias:s.Name}
	//	report = append(report,rec)
	//}
	//scenes,_ := (*ns.bridge).GetScenes()
	//for _,sc := range scenes {
	//	log.Debugf("Scene %s , %s",sc.Name,sc.ID)
	//
	//}

	rules,_ := (*ns.bridge).GetRules()
	for _,sc := range rules {
		log.Debugf("Rule %s , %s",sc.Name)
	}

	msg := fimpgo.NewMessage("evt.network.all_nodes_report", "hue", fimpgo.VTypeObject, report, nil, nil, nil)
	adr := fimpgo.Address{MsgType: fimpgo.MsgTypeEvt, ResourceType: fimpgo.ResourceTypeAdapter, ResourceName: "hue", ResourceAddress: "1"}
	ns.mqt.Publish(&adr, msg)

	return nil
}
