package datavolume

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	cdiv1 "kubevirt.io/containerized-data-importer-api/pkg/apis/core/v1beta1"

	"github.com/spectrocloud/terraform-provider-spectrocloud/spectrocloud/kubevirt/utils"
)

func dataVolumeStatusFields() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"phase": {
			Type:        schema.TypeString,
			Description: "DataVolumePhase is the current phase of the DataVolume.",
			Optional:    true,
			Computed:    true,
			ValidateFunc: validation.StringInSlice([]string{
				"",
				"Pending",
				"PVCBound",
				"ImportScheduled",
				"ImportInProgress",
				"CloneScheduled",
				"CloneInProgress",
				"SnapshotForSmartCloneInProgress",
				"SmartClonePVCInProgress",
				"UploadScheduled",
				"UploadReady",
				"Succeeded",
				"Failed",
				"Unknown",
			}, false),
		},
		"progress": {
			Type:             schema.TypeString,
			Description:      "DataVolumePhase is the current phase of the DataVolume.",
			Optional:         true,
			Computed:         true,
			ValidateDiagFunc: utils.StringIsIntInRange(0, 100),
		},
	}
}

func dataVolumeStatusSchema() *schema.Schema {
	fields := dataVolumeStatusFields()

	return &schema.Schema{
		Type:        schema.TypeList,
		Description: "DataVolumeStatus provides the parameters to store the phase of the Data Volume",
		Optional:    true,
		MaxItems:    1,
		Computed:    true,
		Elem: &schema.Resource{
			Schema: fields,
		},
	}

}

func expandDataVolumeStatus(dataVolumeStatus []interface{}) cdiv1.DataVolumeStatus {
	result := cdiv1.DataVolumeStatus{}

	if len(dataVolumeStatus) == 0 || dataVolumeStatus[0] == nil {
		return result
	}

	in := dataVolumeStatus[0].(map[string]interface{})

	if v, ok := in["phase"].(string); ok {
		result.Phase = cdiv1.DataVolumePhase(v)
	}
	if v, ok := in["progress"].(string); ok {
		result.Progress = cdiv1.DataVolumeProgress(v)
	}

	return result
}

func flattenDataVolumeStatus(in cdiv1.DataVolumeStatus) []interface{} {
	att := map[string]interface{}{
		"phase":    string(in.Phase),
		"progress": string(in.Progress),
	}
	return []interface{}{att}
}
