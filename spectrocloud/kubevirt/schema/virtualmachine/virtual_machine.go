package virtualmachine

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/spectrocloud/terraform-provider-spectrocloud/spectrocloud/convert"
	"github.com/spectrocloud/terraform-provider-spectrocloud/spectrocloud/kubevirt/schema/k8s"
	"github.com/spectrocloud/terraform-provider-spectrocloud/types"

	// "github.com/spectrocloud/terraform-provider-spectrocloud/spectrocloud/kubevirt/schema/virtualmachine"
	"github.com/spectrocloud/terraform-provider-spectrocloud/spectrocloud/kubevirt/schema/virtualmachineinstance"
	"github.com/spectrocloud/terraform-provider-spectrocloud/spectrocloud/kubevirt/utils"
	kubevirtapiv1 "kubevirt.io/api/core/v1"
)

func VirtualMachineFields() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		// Flatten metadata data attributes
		"name": {
			Type:         schema.TypeString,
			Description:  "Name of the virtual machine, must be unique. Cannot be updated.",
			Required:     true,
			ForceNew:     true,
			ValidateFunc: utils.ValidateName,
		},
		"generate_name": {
			Type:         schema.TypeString,
			Description:  "Prefix, used by the server, to generate a unique name ONLY IF the `name` field has not been provided. This value will also be combined with a unique suffix. Read more: https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#idempotency",
			Optional:     true,
			ValidateFunc: utils.ValidateGenerateName,
		},
		"namespace": {
			Type:        schema.TypeString,
			Description: "Namespace defines the space within, Name must be unique.",
			Optional:    true,
			ForceNew:    true,
			Default:     "default",
		},
		"labels": {
			Type:         schema.TypeMap,
			Description:  "Map of string keys and values that can be used to organize and categorize (scope and select). May match selectors of replication controllers and services.",
			Optional:     true,
			Elem:         &schema.Schema{Type: schema.TypeString},
			ValidateFunc: utils.ValidateLabels,
		},
		"annotations": {
			Type:         schema.TypeMap,
			Description:  "An unstructured key value map stored with the VM that may be used to store arbitrary metadata.",
			Optional:     true,
			Elem:         &schema.Schema{Type: schema.TypeString},
			ValidateFunc: utils.ValidateAnnotations,
			Computed:     true,
		},
		"generation": {
			Type:        schema.TypeInt,
			Description: "A sequence number representing a specific generation of the desired state.",
			Computed:    true,
		},
		"resource_version": {
			Type:        schema.TypeString,
			Description: "An opaque value that represents the internal version of this VM that can be used by clients to determine when VM has changed.",
			Computed:    true,
		},
		"self_link": {
			Type:        schema.TypeString,
			Description: "A URL representing this VM.",
			Computed:    true,
		},

		"uid": {
			Type:        schema.TypeString,
			Description: "The unique in time and space value for this VM.",
			Computed:    true,
		},

		"cluster_uid": {
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
			Description: "The cluster UID to which the virtual machine belongs to.",
		},
		"cluster_context": {
			Type:         schema.TypeString,
			Optional:     true,
			Default:      "project",
			Description:  "Context of the cluster. Allowed values are `project`, `tenant`. Default value is `project`.",
			ValidateFunc: validation.StringInSlice([]string{"", "project", "tenant"}, false),
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
			Description: "If set to `true`, the virtual machine will be started when the cluster is launched. Default value is `true`.",
		},
		"vm_action": {
			Type:         schema.TypeString,
			Optional:     true,
			ValidateFunc: validation.StringInSlice([]string{"", "start", "stop", "restart", "pause", "resume", "migrate"}, false),
			Description:  "The action to be performed on the virtual machine. Valid values are: `start`, `stop`, `restart`, `pause`, `resume`, `migrate`. Default value is `start`.",
		},
		"data_volume_templates": dataVolumeTemplatesSchema(),
		"run_strategy": {
			Type:         schema.TypeString,
			Description:  "Running state indicates the requested running state of the VirtualMachineInstance, mutually exclusive with Running.",
			ExactlyOneOf: []string{"run_strategy", "run_on_launch"},
			Optional:     true,
			ValidateFunc: validation.StringInSlice([]string{
				"",
				"Always",
				"Halted",
				"Manual",
				"RerunOnFailure",
			}, false),
		},
		"disk": {
			Type:        schema.TypeList,
			Description: "Disks describes disks, cdroms, floppy and luns which are connected to the vmi.",
			Optional:    true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"name": {
						Type:        schema.TypeString,
						Description: "Name is the device name",
						Required:    true,
					},
					"disk_device": {
						Type:        schema.TypeList,
						Description: "DiskDevice specifies as which device the disk should be added to the guest.",
						Required:    true,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"disk": {
									Type:        schema.TypeList,
									Description: "Attach a volume as a disk to the vmi.",
									Optional:    true,
									Elem: &schema.Resource{
										Schema: map[string]*schema.Schema{
											"bus": {
												Type:        schema.TypeString,
												Description: "Bus indicates the type of disk device to emulate.",
												Required:    true,
											},
											"read_only": {
												Type:        schema.TypeBool,
												Description: "ReadOnly. Defaults to false.",
												Optional:    true,
											},
											"pci_address": {
												Type:        schema.TypeString,
												Description: "If specified, the virtual disk will be placed on the guests pci address with the specified PCI address. For example: 0000:81:01.10",
												Optional:    true,
											},
										},
									},
								},
							},
						},
					},
					"serial": {
						Type:        schema.TypeString,
						Description: "Serial provides the ability to specify a serial number for the disk device.",
						Optional:    true,
					},
					"boot_order": {
						Type:        schema.TypeInt,
						Description: "BootOrder is an integer value > 0, used to determine ordering of boot devices. Lower values take precedence.",
						Optional:    true,
					},
				},
			},
		},
		"interface": {
			Type:        schema.TypeList,
			Description: "Interfaces describe network interfaces which are added to the vmi.",
			Optional:    true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"name": {
						Type:        schema.TypeString,
						Description: "Logical name of the interface as well as a reference to the associated networks.",
						Required:    true,
					},
					"interface_binding_method": {
						Type: schema.TypeString,
						ValidateFunc: validation.StringInSlice([]string{
							"InterfaceBridge",
							"InterfaceSlirp",
							"InterfaceMasquerade",
							"InterfaceSRIOV",
						}, false),
						Description: "Represents the Interface model, One of: e1000, e1000e, ne2k_pci, pcnet, rtl8139, virtio. Defaults to virtio.",
						Required:    true,
					},
					"model": {
						Type:     schema.TypeString,
						Optional: true,
						ValidateFunc: validation.StringInSlice([]string{
							"",
							"e1000",
							"e1000e",
							"ne2k_pci",
							"pcnet",
							"rtl8139",
							"virtio",
						}, false),
						Description: "Represents the method which will be used to connect the interface to the guest.",
					},
				},
			},
		},
		"resources": {
			Type:        schema.TypeList,
			Description: "Resources describes the Compute Resources required by this vmi.",
			MaxItems:    1,
			Required:    true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"requests": {
						Type:        schema.TypeMap,
						Description: "Requests is a description of the initial vmi resources.",
						Optional:    true,
					},
					"limits": {
						Type:        schema.TypeMap,
						Description: "Requests is the maximum amount of compute resources allowed. Valid resource keys are \"memory\" and \"cpu\"",
						Optional:    true,
					},
					"over_commit_guest_overhead": {
						Type:        schema.TypeBool,
						Description: "Don't ask the scheduler to take the guest-management overhead into account. Instead put the overhead only into the container's memory limit. This can lead to crashes if all memory is in use on a node. Defaults to false.",
						Optional:    true,
					},
				},
			},
		},
		"cpu": {
			Type:        schema.TypeList,
			Description: "CPU allows to specifying the CPU topology. Valid resource keys are \"cores\" , \"sockets\" and \"threads\"",
			MaxItems:    1,
			Optional:    true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"cores": {
						Type:        schema.TypeInt,
						Description: "Cores is the number of cores inside the vmi. Must be a value greater or equal 1",
						Optional:    true,
					},
					"sockets": {
						Type:        schema.TypeInt,
						Description: "Sockets is the number of sockets inside the vmi. Must be a value greater or equal 1.",
						Optional:    true,
					},
					"threads": {
						Type:        schema.TypeInt,
						Description: "Threads is the number of threads inside the vmi. Must be a value greater or equal 1.",
						Optional:    true,
					},
				},
			},
		},
		"memory": {
			Type:        schema.TypeList,
			Description: "Memory allows specifying the vmi memory features.",
			MaxItems:    1,
			Optional:    true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"guest": {
						Type:        schema.TypeString,
						Description: "Guest is the amount of memory allocated to the vmi. This value must be less than or equal to the limit if specified.",
						Optional:    true,
					},
					"hugepages": {
						Type: schema.TypeString,
						// PageSize specifies the hugepage size, for x86_64 architecture valid values are 1Gi and 2Mi.
						Description: "Hugepages attribute specifies the hugepage size, for x86_64 architecture valid values are 1Gi and 2Mi.",
						Optional:    true,
					},
				},
			},
		},
		"firmware": {
			Type:        schema.TypeList,
			Description: "Firmware configuration for the virtual machine.",
			MaxItems:    1,
			Optional:    true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"uuid": {
						Type:        schema.TypeString,
						Description: "UUID reported by the vmi bios. Defaults to a random generated uid.",
						Optional:    true,
						Computed:    true,
					},
					"serial": {
						Type:        schema.TypeString,
						Optional:    true,
						Computed:    true,
						Description: "The system-serial-number in SMBIOS.",
					},
					"bootloader": {
						Type:        schema.TypeList,
						Description: "Settings to control the bootloader that is used.",
						MaxItems:    1,
						Optional:    true,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"bios": {
									Type:        schema.TypeList,
									Description: "If set (default), BIOS will be used.",
									MaxItems:    1,
									Optional:    true,
									Elem: &schema.Resource{
										Schema: map[string]*schema.Schema{
											"use_serial": {
												Type:        schema.TypeBool,
												Description: "If set, the BIOS output will be transmitted over serial.",
												Optional:    true,
											},
										},
									},
								},
								"efi": {
									Type:        schema.TypeList,
									Description: "If set, EFI will be used instead of BIOS.",
									MaxItems:    1,
									Optional:    true,
									Elem: &schema.Resource{
										Schema: map[string]*schema.Schema{
											"secure_boot": {
												Type:        schema.TypeBool,
												Description: "If set, SecureBoot will be enabled and the OVMF roms will be swapped for SecureBoot-enabled ones. Requires SMM to be enabled. Defaults to true.",
												Optional:    true,
											},
											"persistent": {
												Type:        schema.TypeBool,
												Description: "If set to true, Persistent will persist the EFI NVRAM across reboots. Defaults to false.",
												Optional:    true,
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		"features": {
			Type:        schema.TypeList,
			Description: "Features allows to configure various virtualization features.",
			MaxItems:    1,
			Optional:    true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"acpi": {
						Type:        schema.TypeList,
						Description: "ACPI enables/disables ACPI inside the guest. Defaults to enabled.",
						MaxItems:    1,
						Optional:    true,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"enabled": {
									Type:        schema.TypeBool,
									Description: "Enabled determines if the feature should be enabled or disabled on the guest. Defaults to true.",
									Optional:    true,
									Default:     true,
								},
							},
						},
					},
					"apic": {
						Type:        schema.TypeList,
						Description: "APIC enables/disables APIC inside the guest. Defaults to enabled.",
						MaxItems:    1,
						Optional:    true,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"enabled": {
									Type:        schema.TypeBool,
									Description: "Enabled determines if the feature should be enabled or disabled on the guest. Defaults to true.",
									Optional:    true,
									Default:     true,
								},
							},
						},
					},
					"smm": {
						Type:        schema.TypeList,
						Description: "SMM enables/disables System Management Mode. Required for Secure Boot with EFI.",
						MaxItems:    1,
						Optional:    true,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"enabled": {
									Type:        schema.TypeBool,
									Description: "Enabled determines if the feature should be enabled or disabled on the guest.",
									Optional:    true,
								},
							},
						},
					},
				},
			},
		},
		"network": virtualmachineinstance.NetworksSchema(),
		"volume":  virtualmachineinstance.VolumesSchema(),
		"priority_class_name": {
			Type:        schema.TypeString,
			Description: "If specified, indicates the pod's priority. If not specified, the pod priority will be default or zero if there is no default.",
			Optional:    true,
		},
		"node_selector": {
			Type:        schema.TypeMap,
			Description: "NodeSelector is a selector which must be true for the vmi to fit on a node. Selector which must match a node's labels for the vmi to be scheduled on that node.",
			Optional:    true,
			Elem:        &schema.Schema{Type: schema.TypeString},
		},
		"affinity": k8s.AffinitySchema(),
		"scheduler_name": {
			Type:        schema.TypeString,
			Description: "If specified, the VMI will be dispatched by specified scheduler. If not specified, the VMI will be dispatched by default scheduler.",
			Optional:    true,
		},
		"tolerations": k8s.TolerationSchema(),
		"eviction_strategy": {
			Type:        schema.TypeString,
			Description: "EvictionStrategy can be set to \"LiveMigrate\" if the VirtualMachineInstance should be migrated instead of shut-off in case of a node drain.",
			Optional:    true,
			ValidateFunc: validation.StringInSlice([]string{
				"LiveMigrate",
			}, false),
		},
		"termination_grace_period_seconds": {
			Type:        schema.TypeInt,
			Description: "Grace period observed after signalling a VirtualMachineInstance to stop after which the VirtualMachineInstance is force terminated.",
			Optional:    true,
		},
		"liveness_probe":  virtualmachineinstance.ProbeSchema(),
		"readiness_probe": virtualmachineinstance.ProbeSchema(),
		"hostname": {
			Type:        schema.TypeString,
			Description: "Specifies the hostname of the vmi.",
			Optional:    true,
		},
		"subdomain": {
			Type:        schema.TypeString,
			Description: "If specified, the fully qualified vmi hostname will be \"<hostname>.<subdomain>.<pod namespace>.svc.<cluster domain>\".",
			Optional:    true,
		},
		"dns_policy": {
			Type:        schema.TypeString,
			Description: "DNSPolicy defines how a pod's DNS will be configured.",
			Optional:    true,
			ValidateFunc: validation.StringInSlice([]string{
				"ClusterFirstWithHostNet",
				"ClusterFirst",
				"Default",
				"None",
			}, false),
		},
		"pod_dns_config": k8s.PodDnsConfigSchema(),

		"status": virtualMachineStatusSchema(),
	}
}

// ResourceDataToHapiVmMeta builds Spectro HAPI VM object metadata from Terraform resource data only (HAPI model, no KubeVirt/Kubernetes types).
func ResourceDataToHapiVmMeta(d *schema.ResourceData) *models.V1VMObjectMeta {
	out := &models.V1VMObjectMeta{}
	if v, ok := d.GetOk("name"); ok {
		out.Name = v.(string)
	}
	if v, ok := d.GetOk("namespace"); ok {
		out.Namespace = v.(string)
	} else {
		out.Namespace = "default"
	}
	if v, ok := d.GetOk("labels"); ok && len(v.(map[string]interface{})) > 0 {
		out.Labels = utils.ExpandStringMap(v.(map[string]interface{}))
	}
	if v, ok := d.GetOk("annotations"); ok && len(v.(map[string]interface{})) > 0 {
		out.Annotations = utils.ExpandStringMap(v.(map[string]interface{}))
	}
	if v, ok := d.GetOk("generate_name"); ok {
		out.GenerateName = v.(string)
	}
	if v, ok := d.GetOk("resource_version"); ok {
		out.ResourceVersion = v.(string)
	}
	if v, ok := d.GetOk("uid"); ok {
		out.UID = v.(string)
	}
	if v, ok := d.GetOk("generation"); ok {
		out.Generation = int64(v.(int))
	}
	return out
}

// ResourceDataToHapiVm builds a Spectro HAPI VM directly from Terraform resource data using only HAPI model types (no KubeVirt VM in between).
func ResourceDataToHapiVm(d *schema.ResourceData) (*models.V1ClusterVirtualMachine, error) {
	hapiSpec, err := ExpandVirtualMachineSpecToHapi(d)
	if err != nil {
		return nil, err
	}
	var statusList []interface{}
	if v := d.Get("status"); v != nil {
		if sl, ok := v.([]interface{}); ok {
			statusList = sl
		}
	}
	hapiStatus, err := expandVirtualMachineStatusToHapi(statusList)
	if err != nil {
		return nil, err
	}
	hapiVM := &models.V1ClusterVirtualMachine{
		Metadata: ResourceDataToHapiVmMeta(d),
		Spec:     hapiSpec,
		Status:   hapiStatus,
	}
	if v, ok := d.GetOk("run_on_launch"); ok {
		if !v.(bool) {
			hapiVM.Spec.RunStrategy = "Manual"
		} else {
			hapiVM.Spec.Running = v.(bool)
		}
	}
	return hapiVM, nil
}

// ExpandVirtualMachineSpecToHapi returns HAPI VM spec from Terraform resource data (HAPI model only; KubeVirt used only internally for conversion).
func ExpandVirtualMachineSpecToHapi(d *schema.ResourceData) (*models.V1ClusterVirtualMachineSpec, error) {
	spec, err := ExpandVirtualMachineSpec(d)
	if err != nil {
		return nil, err
	}
	return convert.ToHapiVmSpecM(spec)
}

// expandVirtualMachineStatusToHapi builds HAPI VM status from Terraform status list using only HAPI models (no KubeVirt types).
func expandVirtualMachineStatusToHapi(virtualMachineStatus []interface{}) (*models.V1ClusterVirtualMachineStatus, error) {
	out := &models.V1ClusterVirtualMachineStatus{}
	if len(virtualMachineStatus) == 0 || virtualMachineStatus[0] == nil {
		return out, nil
	}
	in := virtualMachineStatus[0].(map[string]interface{})
	if v, ok := in["created"].(bool); ok {
		out.Created = v
	}
	if v, ok := in["ready"].(bool); ok {
		out.Ready = v
	}
	if v, ok := in["conditions"].([]interface{}); ok {
		out.Conditions = expandConditionsToHapi(v)
	}
	if v, ok := in["state_change_requests"].([]interface{}); ok {
		out.StateChangeRequests = expandStateChangeRequestsToHapi(v)
	}
	return out, nil
}

func expandConditionsToHapi(conditions []interface{}) []*models.V1VMVirtualMachineCondition {
	if len(conditions) == 0 {
		return nil
	}
	result := make([]*models.V1VMVirtualMachineCondition, 0, len(conditions))
	for _, c := range conditions {
		if c == nil {
			continue
		}
		m := c.(map[string]interface{})
		cond := &models.V1VMVirtualMachineCondition{}
		if v, ok := m["type"].(string); ok {
			cond.Type = types.Ptr(v)
		}
		if v, ok := m["status"].(string); ok {
			cond.Status = types.Ptr(v)
		}
		if v, ok := m["reason"].(string); ok {
			cond.Reason = v
		}
		if v, ok := m["message"].(string); ok {
			cond.Message = v
		}
		result = append(result, cond)
	}
	return result
}

func expandStateChangeRequestsToHapi(requests []interface{}) []*models.V1VMVirtualMachineStateChangeRequest {
	if len(requests) == 0 {
		return nil
	}
	result := make([]*models.V1VMVirtualMachineStateChangeRequest, 0, len(requests))
	for _, r := range requests {
		if r == nil {
			continue
		}
		m := r.(map[string]interface{})
		req := &models.V1VMVirtualMachineStateChangeRequest{}
		if v, ok := m["action"].(string); ok {
			req.Action = v
		}
		if v, ok := m["data"].(map[string]interface{}); ok && len(v) > 0 {
			req.Data = utils.ExpandStringMap(v)
		}
		if v, ok := m["uid"].(string); ok {
			req.UID = types.Ptr(v)
		}
		result = append(result, req)
	}
	return result
}

func FromResourceData(resourceData *schema.ResourceData) (*kubevirtapiv1.VirtualMachine, error) {
	result := &kubevirtapiv1.VirtualMachine{}

	result.ObjectMeta = k8s.ConvertToBasicMetadata(resourceData)
	spec, err := ExpandVirtualMachineSpec(resourceData)
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
	if err := k8s.FlattenMetadata(vm.ObjectMeta, resourceData); err != nil {
		return err
	}
	if err := FlattenVMMToSpectroSchema(vm.Spec, resourceData); err != nil {
		return err
	}
	if err := resourceData.Set("status", flattenVirtualMachineStatus(vm.Status)); err != nil {
		return err
	}

	return nil
}
