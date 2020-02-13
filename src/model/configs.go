package model

import (
	"encoding/json"
	"io/ioutil"
	"time"
)

type Configs struct {
	path                 string
	InstanceAddress      string        `json:"instance_address"`
	MqttServerURI        string        `json:"mqtt_server_uri"`
	MqttUsername         string        `json:"mqtt_server_username"`
	MqttPassword         string        `json:"mqtt_server_password"`
	MqttClientIdPrefix   string        `json:"mqtt_client_id_prefix"`
	Token                string        `json:"token"`
	BridgeId             string        `json:"bridge_id"`
	StatePoolingInterval time.Duration `json:"state_pooling_interval"`
	DimmerRangeMode      string        `json:"dimmer_range_mode"`
	LogFile              string        `json:"log_file"`
	LogLevel             string        `json:"log_level"`
	LogFormat            string        `json:"log_format"`
	ConfiguredAt         string        `json:"configured_at"`
	ConfiguredBy         string        `json:"configured_by"`
	DimmerMaxValue       int           `json:"-"`

}

func NewConfigs(path string) *Configs {
	return &Configs{path:path}
}

func (cf * Configs) LoadFromFile() error {
	configFileBody, err := ioutil.ReadFile(cf.path)
	if err != nil {
		cf.InitDefault()
		return cf.SaveToFile()
	}
	err = json.Unmarshal(configFileBody, cf)
	if err != nil {
		cf.InitDefault()
		return cf.SaveToFile()
	}
	if cf.StatePoolingInterval == 0 {
		cf.StatePoolingInterval = 1
	}
	if cf.DimmerRangeMode == "" {
		cf.DimmerRangeMode = "255"
		cf.DimmerMaxValue = 255
	}
	return nil
}

func (cf *Configs) SaveToFile() error {
	cf.ConfiguredBy = "auto"
	cf.ConfiguredAt = time.Now().Format(time.RFC3339)
	bpayload, err := json.Marshal(cf)
	err = ioutil.WriteFile(cf.path, bpayload, 0664)
	if err != nil {
		return err
	}
	return err
}

func (cf *Configs) InitDefault() {
	cf.InstanceAddress = "1"
	cf.MqttServerURI = "tcp://localhost:1883"
	cf.MqttClientIdPrefix = "hue-ad"
	cf.LogFile = "/var/log/thingsplex/hue-ad/hue-ad.log"
	cf.LogLevel = "info"
	cf.LogFormat = "text"
	cf.StatePoolingInterval = 1
	cf.DimmerRangeMode = "255"
	cf.DimmerMaxValue = 255
}

func (cf *Configs) IsConfigured()bool {
	if cf.Token != "" {
		return true
	}else {
		return false
	}

}