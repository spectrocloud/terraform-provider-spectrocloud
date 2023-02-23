package spectrocloud

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/hapi/client"
	"github.com/spectrocloud/hapi/models"
)

func updateClusterMetadata(c *client.V1Client, d *schema.ResourceData) error {
	return c.UpdateClusterMetadata(d.Id(), toUpdateClusterMetadata(d))
}

func toUpdateClusterMetadata(d *schema.ResourceData) *models.V1ObjectMetaInputEntitySchema {
	return &models.V1ObjectMetaInputEntitySchema{
		Metadata: toClusterMeta(d),
	}
}
