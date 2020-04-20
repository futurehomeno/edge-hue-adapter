package model

import (
	"encoding/json"
	"github.com/futurehomeno/fimpgo/fimptype"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
)

type Manifest struct {
	Configs     []AppConfig  `json:"configs"`
	UIBlocks    []AppUBLock  `json:"ui_blocks"`
	UIButtons   []UIButton   `json:"ui_buttons"`
	Auth        AppAuth      `json:"auth"`
	InitFlow    []string     `json:"init_flow"`
	Services    []AppService `json:"services"`
	AppState    AppStates    `json:"app_state"`
	ConfigState interface{}  `json:"config_state"`
}

type AppConfig struct {
	ID          string            `json:"id"`
	Label       MultilingualLabel `json:"label"`
	ValT        string            `json:"val_t"`
	UI          AppConfigUI       `json:"ui"`
	Val         Value             `json:"val"`
	IsRequired  bool              `json:"is_required"`
	ConfigPoint string            `json:"config_point"`
}

type MultilingualLabel map[string]string

type AppAuth struct {
	Type         string `json:"type"`
	CodeGrantLoginPageUrl string `json:"code_grant_login_page_url"`
	RedirectURL  string `json:"redirect_url"`
	ClientID     string `json:"client_id"`
	Secret       string `json:"secret"`
	PartnerID    string `json:"partner_id"`
	AuthEndpoint string `json:"auth_endpoint"`
}

type AppService struct {
	Name       string               `json:"name"`
	Alias      string               `json:"alias"`
	Address    string               `json:"address"`
	Interfaces []fimptype.Interface `json:"interfaces"`
}

type Value struct {
	Default interface{} `json:"default"`
}

type AppConfigUI struct {
	Type   string      `json:"type"`
	Select interface{} `json:"select"`
}

type UIButton struct {
	ID    string            `json:"id"`
	Label MultilingualLabel `json:"label"`
	Req   struct {
		Serv  string `json:"serv"`
		IntfT string `json:"intf_t"`
		Val   string `json:"val"`
	} `json:"req"`
	ReloadConfig bool `json:"reload_config"`
}

type ButtonActionResponse struct {
	Operation       string `json:"op"`
	OperationStatus string `json:"op_status"`
	Next            string `json:"next"`
	ErrorCode       string `json:"error_code"`
	ErrorText       string `json:"error_text"`
}

type AppUBLock struct {
	Header  MultilingualLabel `json:"header"`
	Text    MultilingualLabel `json:"text"`
	Configs []string          `json:"configs"`
	Buttons []string          `json:"buttons"`
	Footer  MultilingualLabel `json:"footer"`
}

func NewManifest() *Manifest {
	return &Manifest{}
}

func (m *Manifest) LoadFromFile(filePath string) error {
	log.Debug("<manifest> Loading flow from file : ", filePath)
	file, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Error("<manifest> Can't open manifest file.")
		return err
	}
	err = json.Unmarshal(file, m)
	if err != nil {
		log.Error("<FlMan> Can't unmarshal manifest file.")
		return err
	}
	return nil
}

func (m *Manifest) SaveToFile(filePath string) error {
	flowMetaByte, err := json.Marshal(m)
	if err != nil {
		log.Error("<manifest> Can't marshal imported file ")
		return err
	}
	log.Debugf("<manifest> Saving manifest to file %s :", filePath)
	err = ioutil.WriteFile(filePath, flowMetaByte, 0644)
	if err != nil {
		log.Error("<manifest>Can't save flow to file . Error : ", err)
		return err
	}
	return nil
}

//func (m *Manifest) ConfigureServices() {
//	adInterfaces := []fimptype.Interface{{
//		Type:      "in",
//		MsgType:   "cmd.network.get_all_nodes",
//		ValueType: "null",
//		Version:   "1",
//	}, {
//		Type:      "in",
//		MsgType:   "cmd.thing.get_inclusion_report",
//		ValueType: "string",
//		Version:   "1",
//	}, {
//		Type:      "in",
//		MsgType:   "cmd.thing.inclusion",
//		ValueType: "string",
//		Version:   "1",
//	}, {
//		Type:      "in",
//		MsgType:   "cmd.thing.delete",
//		ValueType: "string",
//		Version:   "1",
//	}, {
//		Type:      "in",
//		MsgType:   "cmd.auth.login",
//		ValueType: "str_map", // username/password
//		Version:   "1",
//	}, {
//		Type:      "out",
//		MsgType:   "evt.auth.login_report",
//		ValueType: "string", // success , failed
//		Version:   "1",
//	}, {
//		Type:      "out",
//		MsgType:   "evt.thing.inclusion_report",
//		ValueType: "object",
//		Version:   "1",
//	}, {
//		Type:      "out",
//		MsgType:   "evt.thing.exclusion_report",
//		ValueType: "object",
//		Version:   "1",
//	}, {
//		Type:      "out",
//		MsgType:   "evt.network.all_nodes_report",
//		ValueType: "object",
//		Version:   "1",
//	}, {
//		Type:      "in",
//		MsgType:   "cmd.log.set_level",
//		ValueType: "string",
//		Version:   "1",
//	}}
//
//	adService := AppService{
//		Name:             "hue",
//		Alias:            "Network managment",
//		Address:          "/rt:ad/rn:hue/ad:1",
//		Interfaces: adInterfaces,
//	}
//	m.Services = []AppService{adService}
//
//}