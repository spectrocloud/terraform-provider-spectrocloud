package spectrocloud

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/spectrocloud/palette-sdk-go/client"
)

func getClusterMetadata(d *schema.ResourceData) *models.V1ObjectMeta {
	return &models.V1ObjectMeta{
		Name:        d.Get("name").(string),
		UID:         d.Id(),
		Labels:      toMergedTags(d),
		Annotations: map[string]string{"description": d.Get("description").(string)},
	}
}

func toClusterMetadataUpdate(d *schema.ResourceData) *models.V1ObjectMetaInputEntity {
	return &models.V1ObjectMetaInputEntity{
		Name:        d.Get("name").(string),
		Labels:      toMergedTags(d),
		Annotations: map[string]string{"description": d.Get("description").(string)},
	}
}

func updateClusterMetadata(c *client.V1Client, d *schema.ResourceData) error {
	clusterContext := d.Get("context").(string)
	err := ValidateContext(clusterContext)
	if err != nil {
		return err
	}
	return c.UpdateClusterMetadata(d.Id(), toUpdateClusterMetadata(d))
}

func toUpdateClusterMetadata(d *schema.ResourceData) *models.V1ObjectMetaInputEntitySchema {
	return &models.V1ObjectMetaInputEntitySchema{
		Metadata: toClusterMetadataUpdate(d),
	}
}

func updateClusterAdditionalMetadata(c *client.V1Client, d *schema.ResourceData) error {
	clusterContext := d.Get("context").(string)
	err := ValidateContext(clusterContext)
	if err != nil {
		return err
	}
	return c.UpdateAdditionalClusterMetadata(d.Id(), toUpdateClusterAdditionalMetadata(d))
}

func toUpdateClusterAdditionalMetadata(d *schema.ResourceData) *models.V1ClusterMetaAttributeEntity {
	return &models.V1ClusterMetaAttributeEntity{
		ClusterMetaAttribute: toClusterMetaAttribute(d),
	}
}
