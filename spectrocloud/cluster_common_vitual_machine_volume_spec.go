package spectrocloud

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/gomi/pkg/ptr"
	"github.com/spectrocloud/hapi/models"
)

func prepareDefaultVolumeSpec(d *schema.ResourceData) []*models.V1VMVolume {
	//VM Volume
	var vmVolumes []*models.V1VMVolume
	var vmImage = new(string)
	*vmImage = d.Get("image_url").(string)
	var containerDisk = new(string)
	*containerDisk = "containerdisk"
	vmVolumes = append(vmVolumes, &models.V1VMVolume{
		Name: containerDisk,
		ContainerDisk: &models.V1VMContainerDiskSource{
			Image: vmImage,
		},
	})
	var cloudinitdisk = new(string)
	*cloudinitdisk = "cloudinitdisk"
	vmVolumes = append(vmVolumes, &models.V1VMVolume{
		Name: cloudinitdisk,
		CloudInitNoCloud: &models.V1VMCloudInitNoCloudSource{
			//UserDataBase64: "SGkuXG4=",
			UserData: d.Get("cloud_init_user_data").(string),
		},
	})
	return vmVolumes
}

func prepareVolumeSpec(d *schema.ResourceData) []*models.V1VMVolume {
	if volumesSpec, ok := d.GetOk("volume_spec"); ok {
		var vmVolumes []*models.V1VMVolume
		volumes := volumesSpec.(*schema.Set).List()[0].(map[string]interface{})["volume"]
		for _, vol := range volumes.([]interface{}) {
			v := vol.(map[string]interface{})
			cDisk := v["container_disk"].(*schema.Set).List()
			cInit := v["cloud_init_no_cloud"].(*schema.Set).List()
			dataVolumeDisk := v["data_volume"].(*schema.Set).List()

			vmDiskName := v["name"].(string)

			if len(dataVolumeDisk) > 0 {
				vmVolumes = append(vmVolumes, &models.V1VMVolume{
					Name: &vmDiskName,
					DataVolume: &models.V1VMCoreDataVolumeSource{
						Name: ptr.StringPtr("disk-0-vol"),
					},
				})
			}
			if len(cDisk) > 0 {
				var vmImg = new(string)
				*vmImg = cDisk[0].(map[string]interface{})["image_url"].(string)
				vmVolumes = append(vmVolumes, &models.V1VMVolume{
					Name: &vmDiskName,
					ContainerDisk: &models.V1VMContainerDiskSource{
						Image: vmImg,
					},
				})
			}
			if len(cInit) > 0 {
				vmVolumes = append(vmVolumes, &models.V1VMVolume{
					Name: &vmDiskName,
					CloudInitNoCloud: &models.V1VMCloudInitNoCloudSource{
						UserData: cInit[0].(map[string]interface{})["user_data"].(string),
					},
				})
			}
		}
		return vmVolumes
	} else {
		return prepareDefaultVolumeSpec(d)
	}
}
