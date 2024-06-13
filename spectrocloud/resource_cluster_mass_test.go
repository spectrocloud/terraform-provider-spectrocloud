package spectrocloud

import (
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-api-go/models"
	"github.com/spectrocloud/palette-sdk-go/client"

	"github.com/spectrocloud/terraform-provider-spectrocloud/types"
)

func TestFlattenMachinePoolConfigsMaas(t *testing.T) {
	t.Run("Nil Input", func(t *testing.T) {
		expected := make([]interface{}, 0)
		result := flattenMachinePoolConfigsMaas(nil, nil)

		if !reflect.DeepEqual(expected, result) {
			t.Errorf("Expected empty array for nil input, got: %v", result)
		}
	})

	t.Run("Valid Input", func(t *testing.T) {
		var mockMachinePools []*models.V1MaasMachinePoolConfig
		mp := &models.V1MaasMachinePoolConfig{
			AdditionalLabels: map[string]string{
				"TF": "test_label",
			},
			AdditionalTags: map[string]string{
				"TF": "test_tag",
			},
			Azs: []string{"zone1", "zone2"},
			InstanceType: &models.V1MaasInstanceType{
				MinCPU:     int32(2),
				MinMemInMB: int32(500),
			},
			IsControlPlane: false,
			Labels:         []string{"Masslabel1", "Masslabel2"},
			MachinePoolProperties: &models.V1MachinePoolProperties{
				ArchType: models.V1ArchType("amd64"),
			},
			MaxSize:            3,
			MinSize:            2,
			Name:               "mass_mp_worker",
			NodeRepaveInterval: 30,
			ResourcePool:       "maas_resource",
			Size:               2,
			Tags:               []string{"test"},
			Taints:             nil,
			UpdateStrategy: &models.V1UpdateStrategy{
				Type: "RollingUpdateScaleOut",
			},
			UseControlPlaneAsWorker: true,
		}
		mockMachinePools = append(mockMachinePools, mp)
		config := &models.V1MaasClusterConfig{
			Domain: types.Ptr("maas_resource_pool"),
		}

		expected := []interface{}{
			map[string]interface{}{
				"control_plane":   false,
				"name":            "mass_mp_worker",
				"count":           2,
				"update_strategy": "RollingUpdateScaleOut",
				"max":             3,
				"additional_labels": map[string]string{
					"TF": "test_label",
				},
				"node_repave_interval":    int32(30),
				"control_plane_as_worker": true,
				"min":                     2,
				"instance_type": []interface{}{
					map[string]interface{}{
						"min_memory_mb": 500,
						"min_cpu":       2,
					},
				},
				"azs":       []string{"zone1", "zone2"},
				"node_tags": []string{"test"},
				"placement": []interface{}{
					map[string]interface{}{
						"resource_pool": "maas_resource_pool",
					},
				},
			},
		}

		result := flattenMachinePoolConfigsMaas(mockMachinePools, config)

		if diff := cmp.Diff(expected, result); diff != "" {
			t.Errorf("Unexpected result (-want +got):\n%s", diff)
		}
	})
}

func TestToMachinePoolMaas(t *testing.T) {

	input := map[string]interface{}{
		"control_plane":   false,
		"name":            "mass_mp_worker",
		"count":           2,
		"update_strategy": "RollingUpdateScaleOut",
		"max":             3,
		"additional_labels": map[string]interface{}{
			"TF": "test_label",
		},
		"node_repave_interval":    30,
		"control_plane_as_worker": true,
		"min":                     2,
		"instance_type": []interface{}{
			map[string]interface{}{
				"min_memory_mb": 500,
				"min_cpu":       2,
			},
		},
		"placement": []interface{}{
			map[string]interface{}{
				"id":            "test_id",
				"resource_pool": "test_resource_pool",
			},
		},
		"azs":       schema.NewSet(schema.HashString, []interface{}{"zone1", "zone2"}),
		"node_tags": schema.NewSet(schema.HashString, []interface{}{"test"}),
	}
	rp := "test_resource_pool"
	size := int32(2)
	name := "mass_mp_worker"
	expectedMachinePool := &models.V1MaasMachinePoolConfigEntity{
		CloudConfig: &models.V1MaasMachinePoolCloudConfigEntity{
			Azs:          []string{"zone2", "zone1"},
			InstanceType: &models.V1MaasInstanceType{MinCPU: 2, MinMemInMB: 500},
			ResourcePool: &rp,
			Tags:         []string{"test"},
		},
		PoolConfig: &models.V1MachinePoolConfigEntity{
			AdditionalLabels:        map[string]string{"TF": "test_label"},
			Labels:                  []string{},
			MaxSize:                 3,
			MinSize:                 2,
			Name:                    &name,
			NodeRepaveInterval:      30,
			Size:                    &size,
			UpdateStrategy:          &models.V1UpdateStrategy{Type: "RollingUpdateScaleOut"},
			UseControlPlaneAsWorker: true,
		},
	}

	result, err := toMachinePoolMaas(input)

	if diff := cmp.Diff(expectedMachinePool, result); diff != "" {
		t.Errorf("Unexpected result (-want +got):\n%s", diff)
	}
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("Expected a non-nil result")
	}
}

func TestToMaasCluster(t *testing.T) {

	mockClient := &client.V1Client{}

	d := resourceClusterMaas().TestResourceData()
	d.Set("name", "test_maas_cluster")
	d.Set("context", "tenant")
	d.Set("tags", schema.NewSet(schema.HashString, []interface{}{"tf_tag"}))
	d.Set("cluster_meta_attribute", "zdsdfsdfafs34cada")
	d.Set("cluster_profile", []interface{}{
		map[string]interface{}{
			"id": "test_cluster+profile",
		},
	})
	d.Set("cloud_account_id", "test_account_uid")
	d.Set("os_patch_on_boot", true)
	d.Set("os_patch_schedule", "0 0 * * *")
	d.Set("cloud_config", []interface{}{
		map[string]interface{}{
			"domain": "testccdomain",
		},
	})
	mpools := []interface{}{
		map[string]interface{}{
			"control_plane":   true,
			"name":            "mass_mp_master",
			"count":           2,
			"update_strategy": "RollingUpdateScaleOut",
			"max":             3,
			"additional_labels": map[string]string{
				"TF": string("test_label"),
			},
			"control_plane_as_worker": true,
			"min":                     2,
			"instance_type": []interface{}{
				map[string]interface{}{
					"min_memory_mb": 500,
					"min_cpu":       2,
				},
			},
			"azs":       []string{"zone1", "zone2"},
			"node_tags": []string{"test"},
			"placement": []interface{}{
				map[string]interface{}{
					"id":            "id_placements",
					"resource_pool": "placement_rp",
				},
			},
		},
		map[string]interface{}{
			"control_plane":   false,
			"name":            "mass_mp_worker",
			"count":           2,
			"update_strategy": "RollingUpdateScaleOut",
			"max":             3,
			"additional_labels": map[string]string{
				"TF": string("test_label"),
			},
			"node_repave_interval":    30,
			"control_plane_as_worker": true,
			"min":                     2,
			"instance_type": []interface{}{
				map[string]interface{}{
					"min_memory_mb": 500,
					"min_cpu":       2,
				},
			},
			"azs":       []string{"zone1", "zone2"},
			"node_tags": []string{"test"},
			"placement": []interface{}{
				map[string]interface{}{
					"id":            "id_placements",
					"resource_pool": "placement_rp",
				},
			},
		},
	}
	d.Set("machine_pool", mpools)

	result, err := toMaasCluster(mockClient, d)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("Expected a non-nil result")
	}
	if d.Get("name") != result.Metadata.Name {
		t.Errorf("Expected %s, got %s", d.Get("name"), result.Metadata.Name)
	}
	if d.Get("cluster_meta_attribute") != result.Spec.ClusterConfig.ClusterMetaAttribute {
		t.Errorf("Expected %s, got %s", d.Get("cluster_meta_attribute"), result.Spec.ClusterConfig.ClusterMetaAttribute)
	}
	if d.Get("cloud_account_id") != *result.Spec.CloudAccountUID {
		t.Errorf("Expected %s, got %s", d.Get("cloud_account_id"), *result.Spec.CloudAccountUID)
	}
	if len(d.Get("machine_pool").(*schema.Set).List()) != len(result.Spec.Machinepoolconfig) {
		t.Errorf("Expected %d, got %d", len(d.Get("machine_pool").(*schema.Set).List()), len(result.Spec.Machinepoolconfig))
	}

}
