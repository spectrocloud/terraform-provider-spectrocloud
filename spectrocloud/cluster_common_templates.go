package spectrocloud

import (
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-sdk-go/api/models"
)

// validateClusterProfileAndTemplate ensures only one of cluster_profile or cluster_template is specified
func validateClusterProfileAndTemplate(d *schema.ResourceData) error {
	hasProfile := false
	hasTemplate := false

	if v, ok := d.GetOk("cluster_profile"); ok {
		profiles := v.([]interface{})
		if len(profiles) > 0 {
			hasProfile = true
		}
	}

	if v, ok := d.GetOk("cluster_template"); ok && v.(string) != "" {
		hasTemplate = true
	}

	if hasProfile && hasTemplate {
		return fmt.Errorf("cluster_profile and cluster_template are mutually exclusive. Only one can be specified")
	}

	if !hasProfile && !hasTemplate {
		return fmt.Errorf("either cluster_profile or cluster_template must be specified")
	}

	return nil
}

// validateClusterTemplateUpdate ensures cluster_template is not changed after cluster creation
func validateClusterTemplateUpdate(d *schema.ResourceData) error {
	if d.HasChange("cluster_template") {
		old, new := d.GetChange("cluster_template")
		oldTemplate := old.(string)
		newTemplate := new.(string)

		// Allow setting from empty to a value (during creation)
		if oldTemplate == "" && newTemplate != "" {
			return nil
		}

		// Allow unsetting (from value to empty)
		if oldTemplate != "" && newTemplate == "" {
			return nil
		}

		// Disallow changing from one template to another
		if oldTemplate != "" && newTemplate != "" && oldTemplate != newTemplate {
			return fmt.Errorf("cluster_template does not support day 2 operations. Changing cluster_template after cluster creation is not allowed. Old: %s, New: %s", oldTemplate, newTemplate)
		}
	}
	return nil
}

// toClusterTemplate converts the schema cluster_template field to API model
func toClusterTemplate(d *schema.ResourceData) *models.V1ClusterTemplateRef {
	if v, ok := d.GetOk("cluster_template"); ok && v.(string) != "" {
		templateUID := v.(string)
		log.Printf("Setting cluster template UID: %s", templateUID)
		return &models.V1ClusterTemplateRef{
			UID: templateUID,
		}
	}
	return nil
}

// flattenClusterTemplate extracts the cluster template UID from the cluster spec
func flattenClusterTemplate(clusterTemplate *models.V1SpectroClusterTemplateRef) string {
	if clusterTemplate != nil && clusterTemplate.UID != "" {
		return clusterTemplate.UID
	}
	return ""
}
