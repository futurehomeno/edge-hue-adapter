package model

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/futurehomeno/edge-hue-adapter/utils"
	log "github.com/sirupsen/logrus"
)

type Configs struct {
	path                  string
	InstanceAddress       string        `json:"instance_address"`
	MqttServerURI         string        `json:"mqtt_server_uri"`
	MqttUsername          string        `json:"mqtt_server_username"`
	MqttPassword          string        `json:"mqtt_server_password"`
	MqttClientIdPrefix    string        `json:"mqtt_client_id_prefix"`
	Token                 string        `json:"token"`
	BridgeId              string        `json:"bridge_id"`
	StatePoolingInterval  time.Duration `json:"state_pooling_interval"`
	DimmerRangeMode       string        `json:"dimmer_range_mode"`
	LogFile               string        `json:"log_file"`
	LogLevel              string        `json:"log_level"`
	LogFormat             string        `json:"log_format"`
	ConfiguredAt          string        `json:"configured_at"`
	ConfiguredBy          string        `json:"configured_by"`
	Username              string        `json:"username"`
	Host                  string        `json:"host"`
	DimmerMaxValue        int           `json:"-"`
	WorkDir               string        `json:"-"`
	DiscoveredBridges     string        `json:"discovered_bridges"`
	DiscoveredBridgesTest string        `json:"discovered_bridges_test"`
	DiscoveredBridgesList []string      `json:"discovered_bridges_list"`
	// This is temp solution for better UI interaction
	ConnectionState string `json:"connection_state"`
	Errors          string `json:"errors"`
}

func NewConfigs(workDir string) *Configs {
	conf := &Configs{WorkDir: workDir}
	conf.path = filepath.Join(workDir, "data", "config.json")
	if !utils.FileExists(conf.path) {
		log.Info("Config file doesn't exist.Loading default config")
		defaultConfigFile := filepath.Join(workDir, "defaults", "config.json")
		err := utils.CopyFile(defaultConfigFile, conf.path)
		if err != nil {
			fmt.Print(err)
			panic("Can't copy config file.")
		}
	}
	return conf
}

func (cf *Configs) LoadFromFile() error {
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
	if cf.DimmerRangeMode == "" && cf.DimmerRangeMode == "255" {
		cf.DimmerRangeMode = "255"
		cf.DimmerMaxValue = 255
	} else {
		cf.DimmerMaxValue = 100
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

func (cf *Configs) IsConfigured() bool {
	if cf.Token != "" {
		return true
	} else {
		return false
	}
}

func (cf *Configs) GetDataDir() string {
	return filepath.Join(cf.WorkDir, "data")
}

func (cf *Configs) GetDefaultDir() string {
	return filepath.Join(cf.WorkDir, "defaults")
}

func (cf *Configs) LoadDefaults() error {
	configFile := filepath.Join(cf.WorkDir, "data", "config.json")
	os.Remove(configFile)
	log.Info("Config file doesn't exist.Loading default config")
	defaultConfigFile := filepath.Join(cf.WorkDir, "defaults", "config.json")
	return utils.CopyFile(defaultConfigFile, configFile)
}

type ConfigReport struct {
	OpStatus string    `json:"op_status"`
	OpError  string    `json:"op_error"`
	AppState AppStates `json:"app_state"`
}
