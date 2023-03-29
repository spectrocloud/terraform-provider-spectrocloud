package virtualmachine

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/spectrocloud/terraform-provider-spectrocloud/spectrocloud/kubevirt/schema/k8s"

	kubevirtapiv1 "kubevirt.io/api/core/v1"

	"github.com/spectrocloud/terraform-provider-spectrocloud/spectrocloud/kubevirt/utils/patch"
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
			Type:          schema.TypeString,
			Optional:      true,
			ForceNew:      true,
			ConflictsWith: []string{"spec"},
			Description:   "The name of the source virtual machine that a clone will be created of.",
		},
		"run_on_launch": {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     true,
			Description: "If set to `true`, the virtual machine will be started when the cluster is launched. Default value is `true`.",
		},
		"vm_state": {
			Type:        schema.TypeString,
			Optional:    true,
			Computed:    true,
			Description: "The state of the virtual machine.  The virtual machine can be in one of the following states: `running`, `stopped`, `paused`, `migrating`, `error`, `unknown`.",
		},
		"metadata": k8s.NamespacedMetadataSchema("VirtualMachine", false),
		"spec":     virtualMachineSpecSchema(),
		"status":   virtualMachineStatusSchema(),
	}
}

func ExpandVirtualMachine(virtualMachine []interface{}) (*kubevirtapiv1.VirtualMachine, error) {
	result := &kubevirtapiv1.VirtualMachine{}

	if len(virtualMachine) == 0 || virtualMachine[0] == nil {
		return result, nil
	}

	in := virtualMachine[0].(map[string]interface{})

	if v, ok := in["metadata"].([]interface{}); ok {
		result.ObjectMeta = k8s.ExpandMetadata(v)
	}
	if v, ok := in["spec"].([]interface{}); ok {
		spec, err := expandVirtualMachineSpec(v)
		if err != nil {
			return result, err
		}
		result.Spec = spec
	}
	if v, ok := in["status"].([]interface{}); ok {
		status, err := expandVirtualMachineStatus(v)
		if err != nil {
			return result, err
		}
		result.Status = status
	}

	return result, nil
}

func FlattenVirtualMachine(in kubevirtapiv1.VirtualMachine) []interface{} {
	att := make(map[string]interface{})

	att["metadata"] = k8s.FlattenMetadata(in.ObjectMeta)
	att["spec"] = flattenVirtualMachineSpec(in.Spec)
	att["status"] = flattenVirtualMachineStatus(in.Status)

	return []interface{}{att}
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

func AppendPatchOps(keyPrefix, pathPrefix string, resourceData *schema.ResourceData, ops []patch.PatchOperation) patch.PatchOperations {
	return k8s.AppendPatchOps(keyPrefix+"metadata.0.", pathPrefix+"/metadata/", resourceData, ops)
}
