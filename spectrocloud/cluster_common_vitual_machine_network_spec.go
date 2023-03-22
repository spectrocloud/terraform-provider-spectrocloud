package spectrocloud

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/hapi/models"
)

func prepareDefaultNetworkSpec() []*models.V1VMNetwork {
	var vmNetworks []*models.V1VMNetwork
	var networkName = new(string)
	*networkName = "default"
	vmNetworks = append(vmNetworks, &models.V1VMNetwork{
		Name: networkName,
		Pod:  &models.V1VMPodNetwork{},
	})
	return vmNetworks
}

func prepareNetworkSpec(d *schema.ResourceData) []*models.V1VMNetwork {
	if networkSpecs, ok := d.GetOk("network_spec"); ok {
		var vmNetworks []*models.V1VMNetwork
		networkSpec := networkSpecs.(*schema.Set).List()[0].(map[string]interface{})["nic"]
		for _, nic := range networkSpec.([]interface{}) {
			var nicName *string
			if name, ok := nic.(map[string]interface{})["name"].(string); ok {
				nicName = &name
			}

			var pod *models.V1VMPodNetwork
			var multus *models.V1VMMultusNetwork
			if multusConfig, ok := nic.(map[string]interface{})["multus"]; ok && multusConfig != nil {
				// if multusConfig is not empty
				if len(multusConfig.([]interface{})) > 0 {
					multusMap := multusConfig.([]interface{})[0].(map[string]interface{})

					var multusName *string
					if name, ok := multusMap["network_name"].(string); ok {
						multusName = &name
					}

					var multusDefault bool
					if defaultVal, ok := multusMap["default"].(bool); ok {
						multusDefault = defaultVal
					}

					multus = &models.V1VMMultusNetwork{
						NetworkName: multusName,
						Default:     multusDefault,
					}
				} else {
					multus = nil
					pod = &models.V1VMPodNetwork{}
				}
			}

			var networkType *string
			if t, ok := nic.(map[string]interface{})["network_type"].(string); ok {
				networkType = &t
			}

			if networkType != nil && *networkType == "pod" {
				multus = nil
				pod = &models.V1VMPodNetwork{}
			}

			vmNetworks = append(vmNetworks, &models.V1VMNetwork{
				Multus: multus,
				Name:   nicName,
				Pod:    pod,
			})
		}
		return vmNetworks
	} else {
		return prepareDefaultNetworkSpec()
	}
}
