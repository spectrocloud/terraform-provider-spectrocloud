package spectrocloud

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/spectrocloud/palette-sdk-go/api/models"

	"github.com/spectrocloud/terraform-provider-spectrocloud/util/ptr"
)

// New Sandbox cluster.
func toAppDeploymentClusterGroupEntity(d *schema.ResourceData) *models.V1AppDeploymentClusterGroupEntity {
	return &models.V1AppDeploymentClusterGroupEntity{
		Metadata: &models.V1ObjectMetaInputEntity{
			Name:   d.Get("name").(string),
			Labels: toTags(d),
		},
		Spec: toAppDeploymentClusterGroupSpec(d),
	}
}

func toAppDeploymentClusterGroupSpec(d *schema.ResourceData) *models.V1AppDeploymentClusterGroupSpec {
	return &models.V1AppDeploymentClusterGroupSpec{
		Config:  toV1AppDeploymentClusterGroupConfigEntity(d),
		Profile: toV1AppDeploymentProfileEntity(d),
	}
}

func toV1AppDeploymentClusterGroupConfigEntity(d *schema.ResourceData) *models.V1AppDeploymentClusterGroupConfigEntity {
	return &models.V1AppDeploymentClusterGroupConfigEntity{
		TargetSpec: toAppDeploymentClusterGroupTargetSpec(d),
	}
}

func toAppDeploymentClusterGroupTargetSpec(d *schema.ResourceData) *models.V1AppDeploymentClusterGroupTargetSpec {
	configList := d.Get("config")
	config := configList.([]interface{})[0].(map[string]interface{})

	return &models.V1AppDeploymentClusterGroupTargetSpec{
		ClusterGroupUID: ptr.To(config["cluster_group_uid"].(string)),
		ClusterLimits:   toAppDeploymentTargetClusterLimits(d),
		ClusterName:     ptr.To(config["cluster_name"].(string)),
	}
}

func toAppDeploymentTargetClusterLimits(d *schema.ResourceData) *models.V1AppDeploymentTargetClusterLimits {
	configList := d.Get("config")
	if configList.([]interface{})[0] != nil {
		config := configList.([]interface{})[0].(map[string]interface{})
		for i := range config["limits"].([]interface{}) {
			if config["limits"].([]interface{})[i] != nil {
				limits := config["limits"].([]interface{})[i].(map[string]interface{})
				if limits["cpu"] != nil && limits["memory"] != nil {
					return &models.V1AppDeploymentTargetClusterLimits{
						CPU:        int32(limits["cpu"].(int)),
						MemoryMiB:  int32(limits["memory"].(int)),
						StorageGiB: int32(limits["storage"].(int)),
					}
				}
			}
		}
	}

	return &models.V1AppDeploymentTargetClusterLimits{}
}

func toV1AppDeploymentProfileEntity(d *schema.ResourceData) *models.V1AppDeploymentProfileEntity {
	return &models.V1AppDeploymentProfileEntity{
		AppProfileUID: ptr.To(d.Get("application_profile_uid").(string)),
	}
}

// Existing sandbox cluster
func toAppDeploymentVirtualClusterEntity(d *schema.ResourceData) *models.V1AppDeploymentVirtualClusterEntity {
	return &models.V1AppDeploymentVirtualClusterEntity{
		Metadata: &models.V1ObjectMetaInputEntity{
			Name:   d.Get("name").(string),
			Labels: toTags(d),
		},
		Spec: toAppDeploymentVirtualClusterSpec(d),
	}
}

func toAppDeploymentVirtualClusterSpec(d *schema.ResourceData) *models.V1AppDeploymentVirtualClusterSpec {
	return &models.V1AppDeploymentVirtualClusterSpec{
		Config:  toAppDeploymentVirtualClusterConfigEntity(d),
		Profile: toV1AppDeploymentProfileEntity(d),
	}
}

func toAppDeploymentVirtualClusterConfigEntity(d *schema.ResourceData) *models.V1AppDeploymentVirtualClusterConfigEntity {
	return &models.V1AppDeploymentVirtualClusterConfigEntity{
		TargetSpec: toAppDeploymentVirtualClusterTargetSpec(d),
	}
}

func toAppDeploymentVirtualClusterTargetSpec(d *schema.ResourceData) *models.V1AppDeploymentVirtualClusterTargetSpec {
	configList := d.Get("config")
	config := configList.([]interface{})[0].(map[string]interface{})

	return &models.V1AppDeploymentVirtualClusterTargetSpec{
		ClusterUID: ptr.To(config["cluster_uid"].(string)),
	}
}
