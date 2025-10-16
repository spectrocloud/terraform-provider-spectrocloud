package virtualmachineinstance

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubevirtapiv1 "kubevirt.io/api/core/v1"
)

func ExpandVirtualMachineInstanceTemplateSpec(d *schema.ResourceData) (*kubevirtapiv1.VirtualMachineInstanceTemplateSpec, error) {
	result := &kubevirtapiv1.VirtualMachineInstanceTemplateSpec{}

	// we have removed metadata for template hence set empty metadata object (TBD)***
	result.ObjectMeta = metav1.ObjectMeta{}

	if spec, err := expandVirtualMachineInstanceSpec(d); err == nil {
		result.Spec = spec
	} else {
		return result, err
	}

	return result, nil
}

func FlattenVirtualMachineInstanceTemplateSpec(in kubevirtapiv1.VirtualMachineInstanceTemplateSpec, resourceData *schema.ResourceData) []interface{} {
	att := make(map[string]interface{})

	// Since we removed metadata support in VM instance Spec metadata is not set.
	att["spec"] = flattenVirtualMachineInstanceSpec(in.Spec, resourceData)

	return []interface{}{att}
}
