package spectrocloud

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/spectrocloud/hapi/models"
	"github.com/spectrocloud/palette-sdk-go/client"
)

func resourceVirtualMachine() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVirtualMachineCreate,
		ReadContext:   resourceVirtualMachineRead,
		UpdateContext: resourceVirtualMachineUpdate,
		DeleteContext: resourceVirtualMachineDelete,

		Schema: map[string]*schema.Schema{
			"cluster_uid": {
				Type:     schema.TypeString,
				Required: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"labels": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"annotations": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"namespace": {
				Type:     schema.TypeString,
				Required: true,
			},
			"change_state": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "Running",
				ValidateFunc: validation.StringInSlice([]string{"stop", "restart", "pause", "clone", "resume", ""}, false),
			},
			"vm_state": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"cpu_cores": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  1,
			},
			"run_on_launch": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"memory": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "2G",
			},
			"image_url": {
				Type:          schema.TypeString,
				Optional:      true,
				Default:       "quay.io/kubevirt/alpine-container-disk-demo",
				ConflictsWith: []string{"volume"},
			},
			"network": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"cloud_init_user_data": {
				Type:          schema.TypeString,
				Optional:      true,
				Default:       "\n#cloud-config\nssh_pwauth: True\nchpasswd: { expire: False }\npassword: spectro\ndisable_root: false\n",
				ConflictsWith: []string{"volume"},
			},
			"devices": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"disk": {
							Type:     schema.TypeList,
							Required: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"name": {
										Type:     schema.TypeString,
										Required: true,
									},
									"bus": {
										Type:     schema.TypeString,
										Required: true,
									},
								},
							},
						},
						"interface": {
							Type:     schema.TypeList,
							Required: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"name": {
										Type:     schema.TypeString,
										Required: true,
									},
								},
							},
						},
					},
				},
			},
			"volume": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"container_disk": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"image_url": {
										Type:     schema.TypeString,
										Required: true,
									},
								},
							},
						},
						"cloud_init_no_cloud": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"user_data": {
										Type:     schema.TypeString,
										Required: true,
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func toVirtualMachineCreateRequest(d *schema.ResourceData) (*models.V1ClusterVirtualMachine, error) {
	vmBody := &models.V1ClusterVirtualMachine{
		APIVersion: "kubevirt.io/v1",
		Kind:       "VirtualMachine",
		Metadata: &models.V1VMObjectMeta{
			Name:      d.Get("name").(string),
			Namespace: d.Get("namespace").(string),
		},
		Spec: toSpecCreateRequest(d),
	}

	return vmBody, nil
}

func getDefaultDevices() ([]*models.V1VMDisk, []*models.V1VMInterface) {
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

func getCustomDevices(devices interface{}) ([]*models.V1VMDisk, []*models.V1VMInterface) {
	var vmDisks []*models.V1VMDisk
	var vmInterfaces []*models.V1VMInterface
	//var vmTempVar = new(string)

	for _, d := range devices.(*schema.Set).List() {
		device := d.(map[string]interface{})
		print(device)
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
			vmInterfaces = append(vmInterfaces, &models.V1VMInterface{
				Name:       &interName,
				Masquerade: make(map[string]interface{}),
			})
		}

	}
	return vmDisks, vmInterfaces
}

func getDefaultVolume(d *schema.ResourceData) []*models.V1VMVolume {
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

func getCustomVolume(volumes []interface{}) []*models.V1VMVolume {
	var vmVolumes []*models.V1VMVolume
	for _, vol := range volumes {
		v := vol.(map[string]interface{})
		cDisk := v["container_disk"].(*schema.Set).List() //[0].(map[string]interface{})

		cInit := v["cloud_init_no_cloud"].(*schema.Set).List() //[0].(map[string]interface{})
		if len(cDisk) > 0 {
			//var vmDiskName = new(string)
			vmDiskName := v["name"].(string)
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
			//var vmInitName = new(string)
			vmInitName := v["name"].(string)
			vmVolumes = append(vmVolumes, &models.V1VMVolume{
				Name: &vmInitName,
				CloudInitNoCloud: &models.V1VMCloudInitNoCloudSource{
					UserData: cInit[0].(map[string]interface{})["user_data"].(string),
				},
			})
		}

	}
	return vmVolumes
}

func getDefaultNetwork() []*models.V1VMNetwork {
	var vmNetworks []*models.V1VMNetwork
	var networkName = new(string)
	*networkName = "default" // d.Get("network").(map[string]interface{})["name"].(string)
	vmNetworks = append(vmNetworks, &models.V1VMNetwork{
		Name: networkName,
		Pod:  &models.V1VMPodNetwork{},
	})
	return vmNetworks
}

func getCustomNetwork(network []interface{}) []*models.V1VMNetwork {
	var vmNetworks []*models.V1VMNetwork
	var networkName = new(string)
	for _, n := range network {
		*networkName = n.(map[string]interface{})["name"].(string)
		vmNetworks = append(vmNetworks, &models.V1VMNetwork{
			Name: networkName,
			Pod:  &models.V1VMPodNetwork{},
		})
	}
	return vmNetworks
}

func toSpecCreateRequest(d *schema.ResourceData) *models.V1ClusterVirtualMachineSpec {

	var vmVolumes []*models.V1VMVolume
	var vmDisks []*models.V1VMDisk
	var vmInterfaces []*models.V1VMInterface
	var vmNetworks []*models.V1VMNetwork

	//Handling Network
	if network, ok := d.GetOk("network"); ok {
		vmNetworks = getCustomNetwork(network.([]interface{}))
	} else {
		vmNetworks = getDefaultNetwork()
	}

	// Handling Volume
	if volume, ok := d.GetOk("volume"); ok {
		vmVolumes = getCustomVolume(volume.([]interface{}))
	} else {
		vmVolumes = getDefaultVolume(d)
	}

	// Handling Disk
	if device, ok := d.GetOk("devices"); ok {
		vmDisks, vmInterfaces = getCustomDevices(device)
	} else {
		vmDisks, vmInterfaces = getDefaultDevices()
	}

	vmSpec := &models.V1ClusterVirtualMachineSpec{
		Running: d.Get("run_on_launch").(bool),
		Template: &models.V1VMVirtualMachineInstanceTemplateSpec{
			Spec: &models.V1VMVirtualMachineInstanceSpec{
				Domain: &models.V1VMDomainSpec{
					CPU: &models.V1VMCPU{
						Cores: int64(d.Get("cpu_cores").(int)),
					},
					Devices: &models.V1VMDevices{
						Disks:      vmDisks,
						Interfaces: vmInterfaces,
					},
					Resources: &models.V1VMResourceRequirements{
						Requests: map[string]interface{}{
							"memory": d.Get("memory").(string),
						},
					},
				},
				Networks: vmNetworks,
				Volumes:  vmVolumes,
			},
		},
	}
	return vmSpec
}

func resourceVirtualMachineCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.V1Client)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	clusterUid := d.Get("cluster_uid").(string)

	cluster, err := c.GetCluster(clusterUid)
	if err != nil {
		return diag.FromErr(err)
	}
	if err != nil && cluster == nil {
		return diag.FromErr(fmt.Errorf("cluster not found: %s", clusterUid))
	}

	virtualMachine, err := toVirtualMachineCreateRequest(d)
	if err != nil {
		return diag.FromErr(err)
	}

	vm, err := c.CreateVirtualMachine(cluster.Metadata.UID, virtualMachine)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(vm.Metadata.Name)
	if d.Get("run_on_lunch").(bool) {
		diags, _ = waitForVirtualMachineToRunning(ctx, d, d.Get("cluster_uid").(string), d.Get("name").(string), d.Get("namespace").(string), diags, c, vm.Status.PrintableStatus)
		if diags.HasError() {
			return diags
		}
	}
	resourceVirtualMachineRead(ctx, d, m)
	return diags
}

func resourceVirtualMachineRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.V1Client)
	var diags diag.Diagnostics
	// Read the virtual machine name and namespace from the resource data
	name := d.Get("name").(string)
	namespace := d.Get("namespace").(string)

	// Call the client's method to retrieve the virtual machine details
	vm, err := c.GetVirtualMachine(d.Get("cluster_uid").(string), name, namespace)
	if err != nil {
		return diag.FromErr(err)
	}

	// Update the resource data with the retrieved virtual machine details
	d.SetId(vm.Metadata.Name)
	d.Set("name", vm.Metadata.Name)
	d.Set("namespace", vm.Metadata.Namespace)
	d.Set("cpu_cores", vm.Spec.Template.Spec.Domain.CPU.Cores)
	d.Set("memory", vm.Spec.Template.Spec.Domain.Resources.Requests)
	if _, ok := d.GetOk("run_on_launch"); ok {
		d.Set("run_on_launch", vm.Spec.Running)
	}
	domain := vm.Spec.Template.Spec.Domain
	volume := vm.Spec.Template.Spec.Volumes
	if domain.CPU != nil {
		d.Set("cpu_cores", domain.CPU.Cores)
	}
	if domain.Resources != nil {
		if domain.Resources.Requests != nil {
			if memory := domain.Resources.Requests.(map[string]interface{})["memory"]; memory != nil && memory != "" {
				d.Set("memory", memory.(string))
			}
		}
	}
	if _, ok := d.GetOk("image_url"); ok {
		for _, v := range volume {
			if v.ContainerDisk != nil {
				d.Set("image_url", v.ContainerDisk.Image)
			}
		}
	}
	d.Set("vm_state", vm.Status.PrintableStatus)

	if _, ok := d.GetOk("network"); ok && vm.Spec.Template.Spec.Networks != nil {
		var net []interface{}
		for _, n := range vm.Spec.Template.Spec.Networks {
			net = append(net, map[string]interface{}{
				"name": n.Name,
			})
		}
		d.Set("network", net)
	}

	// set back devices

	//if _, ok := d.GetOk("network"); ok {
	//	d.Set("network", vm.Spec.Template.Spec.Networks)
	//}
	//if devices, ok := d.GetOk("devices"); ok {
	//	d.Set("devices", devices)
	//}
	//if volumes, ok := d.GetOk("volume"); ok {
	//	d.Set("volume", volumes)
	//}

	// Set the domain details
	/*domain := vm.Spec.SpecTemplate.Domain
	if domain != nil {
		if domain.Machine != nil {
			d.Set("machine_type", domain.Machine.Type)
		}
		if domain.CPU != nil {
			d.Set("cpu_cores", domain.CPU.Cores)
		}
		if domain.Resources != nil {
			if domain.Resources.Requests != nil {
				if memory, ok := domain.Resources.Requests["memory"]; ok {
					d.Set("memory", memory.String())
				}
			}
		}
		// Set the disks details
		disks := make([]map[string]interface{}, len(domain.Devices.Disks))
		for i, disk := range domain.Devices.Disks {
			diskMap := make(map[string]interface{})
			diskMap["name"] = disk.Name
			diskMap["bus"] = disk.Disk.Bus
			disks[i] = diskMap
		}
		d.Set("disks", disks)
		// Set the interfaces details
		interfaces := make([]map[string]interface{}, len(domain.Devices.Interfaces))
		for i, iface := range domain.Devices.Interfaces {
			if iface.Masquerade != nil {
				interfaces[i] = map[string]interface{}{
					"name":         iface.Name,
					"masquerade":   true,
					"model_type":   iface.Model.Type,
					"model_driver": iface.Model.Driver.Name,
				}
			} else if iface.Bridge != nil {
				interfaces[i] = map[string]interface{}{
					"name":      iface.Name,
					"bridge":    iface.Bridge.Name,
					"model":     iface.Model.Type,
					"mac":       iface.MACAddress,
					"vif_model": iface.Model.VIFModel,
				}
			}
		}
		d.Set("interfaces", interfaces)
	}*/

	return diags
}

func resourceVirtualMachineUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	/*c := m.(*client.V1Client)

	vm, err := c.VirtualMachineGet(ctx, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	if d.HasChange("spec") {
		updatedSpec := d.Get("spec").(map[string]interface{})
		vm.Spec = toVirtualMachineSpec(updatedSpec)
	}

	if _, err := client.VirtualMachineUpdate(ctx, vm); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(vm.ObjectMeta.Name)*/

	return resourceVirtualMachineRead(ctx, d, m)
}

func resourceVirtualMachineDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.V1Client)
	var diags diag.Diagnostics

	err := c.DeleteVirtualMachine(d.Get("cluster_uid").(string), d.Get("name").(string), d.Get("namespace").(string))
	if err != nil {
		return diag.FromErr(err)
	}
	return diags
}
