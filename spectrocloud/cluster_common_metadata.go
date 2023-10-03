package spectrocloud

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/hapi/models"
	"github.com/spectrocloud/palette-sdk-go/client"
)

func updateClusterMetadata(c *client.V1Client, d *schema.ResourceData) error {
	return c.UpdateClusterMetadata(d.Id(), toUpdateClusterMetadata(d))
}

func toUpdateClusterMetadata(d *schema.ResourceData) *models.V1ObjectMetaInputEntitySchema {
	return &models.V1ObjectMetaInputEntitySchema{
		Metadata: toClusterMeta(d),
	}
}

func updateClusterAdditionalMetadata(c *client.V1Client, d *schema.ResourceData) error {
	return c.UpdateAdditionalClusterMetadata(d.Id(), toUpdateClusterAdditionalMetadata(d))
}

func toUpdateClusterAdditionalMetadata(d *schema.ResourceData) *models.V1ClusterMetaAttributeEntity {
	return &models.V1ClusterMetaAttributeEntity{
		ClusterMetaAttribute: toClusterMetaAttribute(d),
	}
}
