package spectrocloud

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/hapi/models"
	"github.com/spectrocloud/terraform-provider-spectrocloud/pkg/client"
)

func updateClusterOsPatchConfig(c *client.V1Client, d *schema.ResourceData) error {
	machineConfig := toMachineManagementConfig(d)
	if machineConfig != nil {
		return c.UpdateClusterOsPatchConfig(d.Id(), toUpdateOsPatchEntityClusterRbac(machineConfig.OsPatchConfig))
	}
	return nil
}

func toUpdateOsPatchEntityClusterRbac(config *models.V1OsPatchConfig) *models.V1OsPatchEntity {
	return &models.V1OsPatchEntity{
		OsPatchConfig: config,
	}
}
