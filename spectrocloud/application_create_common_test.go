package spectrocloud

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestToAppDeploymentClusterGroupEntity(t *testing.T) {
	d := schema.TestResourceDataRaw(t, appDeploymentSchema(), map[string]interface{}{
		"name": "test-cluster-group",
		"config": []interface{}{
			map[string]interface{}{
				"cluster_group_uid": "cg-uid",
				"cluster_name":      "cluster-name",
				"limits": []interface{}{
					map[string]interface{}{
						"cpu":     4,
						"memory":  2048,
						"storage": 100,
					},
				},
			},
		},
		"application_profile_uid": "app-profile-uid",
		"labels": map[string]interface{}{
			"env": "test",
		},
	})

	entity := toAppDeploymentClusterGroupEntity(d)

	assert.NotNil(t, entity)
	assert.Equal(t, "test-cluster-group", entity.Metadata.Name)
	assert.Equal(t, "cg-uid", *entity.Spec.Config.TargetSpec.ClusterGroupUID)
	assert.Equal(t, int32(4), entity.Spec.Config.TargetSpec.ClusterLimits.CPU)
	assert.Equal(t, int32(2048), entity.Spec.Config.TargetSpec.ClusterLimits.MemoryMiB)
	assert.Equal(t, int32(100), entity.Spec.Config.TargetSpec.ClusterLimits.StorageGiB)
	assert.Equal(t, "cluster-name", *entity.Spec.Config.TargetSpec.ClusterName)
	assert.Equal(t, "app-profile-uid", *entity.Spec.Profile.AppProfileUID)
}

func TestToAppDeploymentVirtualClusterEntity(t *testing.T) {
	d := schema.TestResourceDataRaw(t, appDeploymentSchema(), map[string]interface{}{
		"name": "test-virtual-cluster",
		"config": []interface{}{
			map[string]interface{}{
				"cluster_uid": "vc-uid",
			},
		},
		"application_profile_uid": "app-profile-uid",
		"labels": map[string]interface{}{
			"env": "prod",
		},
	})

	entity := toAppDeploymentVirtualClusterEntity(d)

	assert.NotNil(t, entity)
	assert.Equal(t, "test-virtual-cluster", entity.Metadata.Name)
	assert.Equal(t, "vc-uid", *entity.Spec.Config.TargetSpec.ClusterUID)
	assert.Equal(t, "app-profile-uid", *entity.Spec.Profile.AppProfileUID)
}

// Helper function to return a schema.ResourceData schema for testing
func appDeploymentSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"name": {
			Type:     schema.TypeString,
			Required: true,
		},
		"config": {
			Type:     schema.TypeList,
			Required: true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"cluster_group_uid": {
						Type:     schema.TypeString,
						Optional: true,
					},
					"cluster_name": {
						Type:     schema.TypeString,
						Optional: true,
					},
					"cluster_uid": {
						Type:     schema.TypeString,
						Optional: true,
					},
					"limits": {
						Type:     schema.TypeList,
						Optional: true,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"cpu": {
									Type:     schema.TypeInt,
									Optional: true,
								},
								"memory": {
									Type:     schema.TypeInt,
									Optional: true,
								},
								"storage": {
									Type:     schema.TypeInt,
									Optional: true,
								},
							},
						},
					},
				},
			},
		},
		"application_profile_uid": {
			Type:     schema.TypeString,
			Required: true,
		},
		"labels": {
			Type:     schema.TypeMap,
			Optional: true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
	}
}
