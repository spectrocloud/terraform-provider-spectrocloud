package convert

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/spectrocloud/terraform-provider-spectrocloud/spectrocloud/kubevirt/schema/datavolume"
)

// FromHapiVolume writes fields from a Palette V1VMAddVolumeEntity into Terraform resource data
// without converting through CDI (cdiv1.DataVolume).
func FromHapiVolume(hapiVolume *models.V1VMAddVolumeEntity, d *schema.ResourceData) error {
	return datavolume.ToResourceDataFromVMAddVolumeEntity(hapiVolume, d)
}
