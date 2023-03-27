package spectrocloud

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/hapi/models"
)

func prepareDefaultDevices() ([]*models.V1VMDisk, []*models.V1VMInterface) {
	var containerDisk = new(string)
	*containerDisk = "containerdisk"
	var cloudinitdisk = new(string)
	*cloudinitdisk = "cloudinitdisk"
	var vmDisks []*models.V1VMDisk
	vmDisks = append(vmDisks, &models.V1VMDisk{
		Name: containerDisk,
		Disk: &models.V1VMDiskTarget{
			Bus: "virtio",
		},
	})
	vmDisks = append(vmDisks, &models.V1VMDisk{
		Name: cloudinitdisk,
		Disk: &models.V1VMDiskTarget{
			Bus: "virtio",
		},
	})
	var vmInterfaces []*models.V1VMInterface
	var def = new(string)
	*def = "default"
	vmInterfaces = append(vmInterfaces, &models.V1VMInterface{
		Name:       def,
		Masquerade: make(map[string]interface{}),
	})

	return vmDisks, vmInterfaces
}

func prepareDevices(d *schema.ResourceData) ([]*models.V1VMDisk, []*models.V1VMInterface) {
	if device, ok := d.GetOk("devices"); ok {
		var vmDisks []*models.V1VMDisk
		var vmInterfaces []*models.V1VMInterface

		for _, d := range device.(*schema.Set).List() {
			device := d.(map[string]interface{})

			// For Disk
			for _, disk := range device["disk"].([]interface{}) {
				diskName := disk.(map[string]interface{})["name"].(string)
				vmDisks = append(vmDisks, &models.V1VMDisk{
					Name: &diskName,
					Disk: &models.V1VMDiskTarget{
						Bus: disk.(map[string]interface{})["bus"].(string),
					},
				})
			}

			// For Interface
			for _, inter := range device["interface"].([]interface{}) {
				interName := inter.(map[string]interface{})["name"].(string)

				var interfaceModel string
				if model, ok := inter.(map[string]interface{})["model"].(string); ok {
					interfaceModel = model
				} else {
					interfaceModel = "virtio"
				}

				interfaceType := inter.(map[string]interface{})["type"].(string)
				var vmInterface *models.V1VMInterface
				switch interfaceType {
				case "masquerade":
					vmInterface = &models.V1VMInterface{
						Name:       &interName,
						Model:      interfaceModel,
						Masquerade: make(map[string]interface{}),
					}
				case "bridge":
					vmInterface = &models.V1VMInterface{
						Name:   &interName,
						Model:  interfaceModel,
						Bridge: make(map[string]interface{}),
					}
				case "macvtap":
					vmInterface = &models.V1VMInterface{
						Name:    &interName,
						Model:   interfaceModel,
						Macvtap: make(map[string]interface{}),
					}
				}

				vmInterfaces = append(vmInterfaces, vmInterface)
			}
		}
		return vmDisks, vmInterfaces
	} else {
		return prepareDefaultDevices()
	}
}
