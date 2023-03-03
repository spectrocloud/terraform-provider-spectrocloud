package spectrocloud

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/hapi/models"
	"github.com/spectrocloud/palette-sdk-go/client"
)

//func toVirtualMachineCreateRequest(d *schema.ResourceData) (*models.V1SpectroClusterVMCreateEntity, error) {
//	virtualMachine := &client.VirtualMachine{
//		APIVersion: "kubevirt.io/v1",
//		APIGroup:   "kubevirt.io",
//		Kind:       "VirtualMachine",
//		Metadata: client.Metadata{
//			Name:      d.Get("name").(string),
//			Namespace: d.Get("namespace").(string),
//		},
//		Spec: toSpecCreateRequest(d),
//	}
//	return virtualMachine, nil
//}

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

func toSpecCreateRequest(d *schema.ResourceData) *models.V1ClusterVirtualMachineSpec {
	// Network
	var vmNetworks []*models.V1VMNetwork
	var networkName = new(string)
	*networkName = d.Get("Network_name").(string)
	vmNetworks = append(vmNetworks, &models.V1VMNetwork{
		Name: networkName,
	})

	//VM Volume
	var vmVolumes []*models.V1VMVolume
	var vmImage = new(string)
	*vmImage = d.Get("image").(string)
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
			UserDataBase64: "SGkuXG4=",
		},
	})

	// Default Disk Configuration
	var vmdisks []*models.V1VMDisk
	vmdisks = append(vmdisks, &models.V1VMDisk{
		Name: containerDisk,
		Disk: &models.V1VMDiskTarget{
			Bus: "virtio",
		},
	})
	vmdisks = append(vmdisks, &models.V1VMDisk{
		Name: cloudinitdisk,
		Disk: &models.V1VMDiskTarget{
			Bus: "virtio",
		},
	})

	// Interface
	var vmInterfaces []*models.V1VMInterface
	var def = new(string)
	*def = "default"
	vmInterfaces = append(vmInterfaces, &models.V1VMInterface{
		Name:       def,
		Masquerade: nil,
	})

	// memory
	type vmRequest struct {
		memory string
		cpu    int64
	}
	vmspec := &models.V1ClusterVirtualMachineSpec{
		RunStrategy: d.Get("state").(string),
		Running:     d.Get("run_on_launch").(bool),
		Template: &models.V1VMVirtualMachineInstanceTemplateSpec{
			Spec: &models.V1VMVirtualMachineInstanceSpec{
				DNSPolicy: "Default",
				Domain: &models.V1VMDomainSpec{
					Chassis: nil,
					Clock:   nil,
					CPU: &models.V1VMCPU{
						Cores: d.Get("CPU").(int64),
					},
					Devices: &models.V1VMDevices{
						Disks:      vmdisks,
						Interfaces: vmInterfaces,
					},
					//Memory: &models.V1VMMemory{
					//	Guest: models.V1VMQuantity(d.Get("memory").(string)),
					//},
					Resources: &models.V1VMResourceRequirements{
						Requests: vmRequest{
							memory: d.Get("memory").(string),
							cpu:    d.Get("CPU").(int64),
						},
					},
				},
				Networks: vmNetworks,
				Volumes:  vmVolumes,
			},
		},
	}
	return vmspec
}

//func toSpecCreateRequest(d *schema.ResourceData) client.Spec {
//	template := client.SpecTemplate{
//		Domain: client.Domain{
//			CPU: client.CPU{
//				Cores: d.Get("cpu_cores").(int),
//			},
//			Devices: client.Devices{
//				Disks: []client.Disk{
//					client.Disk{
//						Name: d.Get("disk_name").(string),
//						DiskType: client.DiskType{
//							Bus: d.Get("disk_bus").(string),
//						},
//					},
//				},
//			},
//			Machine: client.Machine{
//				Type: "q35",
//			},
//			Resources: client.Resources{
//				Requests: client.Requests{
//					Memory: d.Get("memory").(string),
//				},
//			},
//		},
//		Networks: []client.Network{
//			client.Network{
//				Name: "default",
//				Pod:  client.Pod{},
//			},
//		},
//		Volumes: []client.Volume{
//			client.Volume{
//				Name: d.Get("volume_name").(string),
//				ContainerDisk: client.ContainerDisk{
//					Image: d.Get("image").(string),
//				},
//				CloudInitNoCloud: client.CloudInitNoCloud{
//					UserData: d.Get("user_data").(string),
//				},
//			},
//		},
//	}
//
//	return client.Spec{
//		Status:       d.Get("status").(string),
//		SpecTemplate: template,
//	}
//}

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
			"namespace": {
				Type:     schema.TypeString,
				Required: true,
			},
			"state": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "Running",
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
			"image": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "quay.io/kubevirt/alpine-container-disk-demo",
			},
			"cloudinit_user_data": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "\n#cloud-config\nssh_pwauth: True\nchpasswd: { expire: False }\npassword: spectro\ndisable_root: false\n            ",
			},
		},
	}
}

func resourceVirtualMachineCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.V1Client)

	clusterUid := d.Get("cluster_uid").(string)

	cluster, err := c.GetCluster(clusterUid)
	if err != nil && cluster == nil {
		return diag.FromErr(fmt.Errorf("cluster not found: %s", clusterUid))
	}

	virtualMachine, err := toVirtualMachineCreateRequest(d)
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = c.CreateVirtualMachine(cluster.Metadata.UID, virtualMachine)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceVirtualMachineRead(ctx, d, m)
}

func getVMId(clusterUid string, namespace string) string {
	return clusterUid + "_" + namespace
}

func resourceVirtualMachineRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.V1Client)

	// Read the virtual machine name and namespace from the resource data
	name := d.Get("name").(string)
	namespace := d.Get("namespace").(string)

	// Call the client's method to retrieve the virtual machine details
	vm, err := c.GetVirtualMachine(getVMId(name, namespace))
	if err != nil {
		return diag.FromErr(err)
	}

	// Update the resource data with the retrieved virtual machine details
	d.SetId(fmt.Sprintf("%s/%s", vm.Metadata.Namespace, vm.Metadata.Name))
	d.Set("name", vm.Metadata.Name)
	d.Set("namespace", vm.Metadata.Namespace)
	d.Set("status", vm.Spec.Status)
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

	return nil
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
