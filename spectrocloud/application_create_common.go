package spectrocloud

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/gomi/pkg/ptr"
	"github.com/spectrocloud/hapi/models"
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
		ClusterGroupUID: ptr.StringPtr(config["cluster_group_uid"].(string)),
		ClusterLimits:   toAppDeploymentTargetClusterLimits(d),
		ClusterName:     ptr.StringPtr(config["cluster_name"].(string)),
	}
}

func toAppDeploymentTargetClusterLimits(d *schema.ResourceData) *models.V1AppDeploymentTargetClusterLimits {
	configList := d.Get("config")
	if configList.([]interface{})[0] != nil {
		config := configList.([]interface{})[0].(map[string]interface{})
		if config["limits"].([]interface{})[0] != nil {
			limits := config["limits"].([]interface{})[0].(map[string]interface{})
			if limits["cpu"] != nil && limits["memory"] != nil {
				return &models.V1AppDeploymentTargetClusterLimits{
					CPU:       int32(limits["cpu"].(int)),
					MemoryMiB: int32(limits["memory"].(int)),
				}
			}
		}
	}

	return &models.V1AppDeploymentTargetClusterLimits{}
}

func toV1AppDeploymentProfileEntity(d *schema.ResourceData) *models.V1AppDeploymentProfileEntity {
	return &models.V1AppDeploymentProfileEntity{
		AppProfileUID: ptr.StringPtr(d.Get("application_profile_uid").(string)),
	}
}

// Existing sandbox cluster
func toAppDeploymentNestedClusterEntity(d *schema.ResourceData) *models.V1AppDeploymentNestedClusterEntity {
	return &models.V1AppDeploymentNestedClusterEntity{
		Metadata: &models.V1ObjectMetaInputEntity{
			Name:   d.Get("name").(string),
			Labels: toTags(d),
		},
		Spec: toAppDeploymentNestedClusterSpec(d),
	}
}

func toAppDeploymentNestedClusterSpec(d *schema.ResourceData) *models.V1AppDeploymentNestedClusterSpec {
	return &models.V1AppDeploymentNestedClusterSpec{
		Config:  toAppDeploymentNestedClusterConfigEntity(d),
		Profile: toV1AppDeploymentProfileEntity(d),
	}
}

func toAppDeploymentNestedClusterConfigEntity(d *schema.ResourceData) *models.V1AppDeploymentNestedClusterConfigEntity {
	return &models.V1AppDeploymentNestedClusterConfigEntity{
		TargetSpec: toAppDeploymentNestedClusterTargetSpec(d),
	}
}

func toAppDeploymentNestedClusterTargetSpec(d *schema.ResourceData) *models.V1AppDeploymentNestedClusterTargetSpec {
	configList := d.Get("config")
	config := configList.([]interface{})[0].(map[string]interface{})

	return &models.V1AppDeploymentNestedClusterTargetSpec{
		ClusterUID: ptr.StringPtr(config["cluster_uid"].(string)),
	}
}
