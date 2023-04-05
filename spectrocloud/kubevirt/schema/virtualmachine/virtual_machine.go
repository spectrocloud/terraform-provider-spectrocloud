package virtualmachine

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/spectrocloud/terraform-provider-spectrocloud/spectrocloud/kubevirt/schema/k8s"

	kubevirtapiv1 "kubevirt.io/api/core/v1"
)

func VirtualMachineFields() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"cluster_uid": {
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
			Description: "The cluster UID to which the virtual machine belongs to.",
		},
		"base_vm_name": {
			Type:        schema.TypeString,
			Optional:    true,
			ForceNew:    true,
			Description: "The name of the source virtual machine that a clone will be created of.",
		},
		"run_on_launch": {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     true,
			Description: "If set to `true`, the virtual machine will be started when the cluster is launched. Default value is `true`.",
		},
		"vm_action": {
			Type:         schema.TypeString,
			Optional:     true,
			ValidateFunc: validation.StringInSlice([]string{"", "start", "stop", "restart", "pause", "resume", "migrate"}, false),
			Description:  "The action to be performed on the virtual machine. Valid values are: `start`, `stop`, `restart`, `pause`, `resume`, `migrate`. Default value is `start`.",
		},
		"metadata": k8s.NamespacedMetadataSchema("VirtualMachine", false),
		"spec":     virtualMachineSpecSchema(),
		"status":   virtualMachineStatusSchema(),
	}
}

func FromResourceData(resourceData *schema.ResourceData) (*kubevirtapiv1.VirtualMachine, error) {
	result := &kubevirtapiv1.VirtualMachine{}

	result.ObjectMeta = k8s.ExpandMetadata(resourceData.Get("metadata").([]interface{}))
	spec, err := expandVirtualMachineSpec(resourceData.Get("spec").([]interface{}))
	if err != nil {
		return result, err
	}
	result.Spec = spec
	status, err := expandVirtualMachineStatus(resourceData.Get("status").([]interface{}))
	if err != nil {
		return result, err
	}
	result.Status = status

	return result, nil
}

func ToResourceData(vm kubevirtapiv1.VirtualMachine, resourceData *schema.ResourceData) error {
	if err := resourceData.Set("metadata", k8s.FlattenMetadata(vm.ObjectMeta)); err != nil {
		return err
	}
	if err := resourceData.Set("spec", flattenVirtualMachineSpec(vm.Spec)); err != nil {
		return err
	}
	if err := resourceData.Set("status", flattenVirtualMachineStatus(vm.Status)); err != nil {
		return err
	}

	return nil
}
