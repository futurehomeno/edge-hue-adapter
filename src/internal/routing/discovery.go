package routing

import (
	"github.com/futurehomeno/cliffhanger/discovery"
)

// GetDiscoveryResource returns a service discovery configuration.
func GetDiscoveryResource() *discovery.Resource {
	return &discovery.Resource{
		ResourceName:           ResourceName,
		ResourceType:           discovery.ResourceTypeAd,
		ResourceFullName:       "Hue",
		Author:                 "support@futurehome.no",
		IsInstanceConfigurable: false,
		InstanceID:             "1",
		Version:                "1",
		AdapterInfo: discovery.AdapterInfo{
			Technology:            "hue",
			FwVersion:             "all",
			NetworkManagementType: "inclusion_exclusion",
		},
	}
}
