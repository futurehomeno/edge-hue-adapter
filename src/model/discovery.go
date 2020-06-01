package model

import (
	"github.com/futurehomeno/fimpgo/discovery"
)

func GetDiscoveryResource() discovery.Resource {

	//adService := fimptype.Service{
	//	Name:             "hue",
	//	Alias:            "Network managment",
	//	Address:          "/rt:ad/rn:hue/ad:1",
	//	Enabled:          true,
	//	Groups:           []string{"ch_0"},
	//	Tags:             nil,
	//	PropSetReference: "",
	//}
	return discovery.Resource{
		ResourceName:           "hue",
		ResourceType:           discovery.ResourceTypeAd,
		PackageName:            "hue-ad",
		Author:                 "aleksandrs.livincovs@gmail.com",
		IsInstanceConfigurable: false,
		InstanceId:             "1",
		Version:                "1",
		AdapterInfo: discovery.AdapterInfo{
			Technology:            "hue",
			FwVersion:             "all",
			NetworkManagementType: "inclusion_exclusion",
		},
	}

}
