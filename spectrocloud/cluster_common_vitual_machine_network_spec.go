package spectrocloud

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/hapi/models"
)

func prepareDefaultNetworkSpec() []*models.V1VMNetwork {
	var vmNetworks []*models.V1VMNetwork
	var networkName = new(string)
	*networkName = "default" // d.Get("network").(map[string]interface{})["name"].(string)
	vmNetworks = append(vmNetworks, &models.V1VMNetwork{
		Name: networkName,
		Pod:  &models.V1VMPodNetwork{},
	})
	return vmNetworks
}

func prepareNetworkSpec(d *schema.ResourceData) []*models.V1VMNetwork {
	if network, ok := d.GetOk("network_spec"); ok {
		var vmNetworks []*models.V1VMNetwork
		var networkName = new(string)
		networkSpec := network.(*schema.Set).List()[0].(map[string]interface{})["network"]
		for _, n := range networkSpec.([]interface{}) {
			*networkName = n.(map[string]interface{})["name"].(string)
			vmNetworks = append(vmNetworks, &models.V1VMNetwork{
				Name: networkName,
				Pod:  &models.V1VMPodNetwork{},
			})
		}
		return vmNetworks
	} else {
		return prepareDefaultNetworkSpec()
	}
}
