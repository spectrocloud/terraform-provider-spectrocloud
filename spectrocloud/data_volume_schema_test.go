package spectrocloud

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/stretchr/testify/assert"

	"github.com/spectrocloud/terraform-provider-spectrocloud/spectrocloud/kubevirt/utils"
	"github.com/spectrocloud/terraform-provider-spectrocloud/util/ptr"
)

func prepareDataVolumeTestData() *schema.ResourceData {
	rd := resourceKubevirtDataVolume().TestResourceData()

	rd.Set("cluster_uid", "cluster-123")
	rd.Set("cluster_context", "tenant")
	rd.Set("vm_name", "vm-test")
	rd.Set("vm_namespace", "default")
	rd.Set("metadata", []interface{}{
		map[string]interface{}{
			"name":      "vol-test",
			"namespace": "default",
			"labels": map[string]interface{}{
				"key1": "value1",
			},
			"annotations": map[string]interface{}{
				"key1": "value1",
			},
		},
	})
	rd.Set("add_volume_options", []interface{}{
		map[string]interface{}{
			"name": "vol-test",
			"disk": []interface{}{
				map[string]interface{}{
					"name": "vol-test",
					"bus":  "scsi",
				},
			},
			"volume_source": []interface{}{
				map[string]interface{}{
					"data_volume": []interface{}{
						map[string]interface{}{
							"name":         "vol-test",
							"hotpluggable": true,
						},
					},
				},
			},
		},
	})
	rd.Set("spec", []interface{}{
		map[string]interface{}{
			"source": []interface{}{
				map[string]interface{}{
					"http": []interface{}{
						map[string]interface{}{
							"url": "https://cloud.centos.org/centos/7/images/CentOS-7-x86_64-GenericCloud.qcow2",
						},
					},
				},
			},
			"pvc": []interface{}{
				map[string]interface{}{
					"access_modes": []interface{}{
						"ReadWriteOnce",
					},
					"resources": []interface{}{
						map[string]interface{}{
							"requests": map[string]interface{}{
								"storage": "10Gi",
							},
						},
					},
					"storage_class_name": "local-storage",
				},
			},
		},
	})
	buildUID := utils.BuildIdDV(rd.Get("cluster_context").(string), rd.Get("cluster_uid").(string), rd.Get("vm_namespace").(string), rd.Get("vm_name").(string), &models.V1VMObjectMeta{
		Name:      "vol-test",
		Namespace: "default",
	})

	rd.SetId(buildUID)

	return rd
}

func TestExpandAddVolumeOptions(t *testing.T) {
	assert := assert.New(t)

	addVolumeOptions := []interface{}{
		map[string]interface{}{
			"name": "test-volume",
			"disk": []interface{}{
				map[string]interface{}{
					"name": "test-disk",
					"bus":  "scsi",
				},
			},
			"volume_source": []interface{}{
				map[string]interface{}{
					"data_volume": []interface{}{
						map[string]interface{}{
							"name":         "test-data-volume",
							"hotpluggable": true,
						},
					},
				},
			},
		},
	}

	expected := &models.V1VMAddVolumeOptions{
		Name: ptr.To("test-volume"),
		Disk: &models.V1VMDisk{
			Name: ptr.To("test-disk"),
			Disk: &models.V1VMDiskTarget{
				Bus: "scsi",
			},
		},
		VolumeSource: &models.V1VMHotplugVolumeSource{
			DataVolume: &models.V1VMCoreDataVolumeSource{
				Name:         ptr.To("test-data-volume"),
				Hotpluggable: true,
			},
		},
	}

	result := ExpandAddVolumeOptions(addVolumeOptions)

	assert.Equal(expected, result, "ExpandAddVolumeOptions returned unexpected result")
}
