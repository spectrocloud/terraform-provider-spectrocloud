package spectrocloud

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-sdk-go/client"
)

func toVirtualMachineCreateRequest(d *schema.ResourceData) (*client.VirtualMachine, error) {
	virtualMachine := &client.VirtualMachine{
		APIVersion: "kubevirt.io/v1",
		APIGroup:   "kubevirt.io",
		Kind:       "VirtualMachine",
		Metadata: client.Metadata{
			Name:      d.Get("name").(string),
			Namespace: d.Get("namespace").(string),
		},
		Spec: toSpecCreateRequest(d),
	}
	return virtualMachine, nil
}

func toSpecCreateRequest(d *schema.ResourceData) client.Spec {
	template := client.SpecTemplate{
		Domain: client.Domain{
			CPU: client.CPU{
				Cores: d.Get("cpu_cores").(int),
			},
			Devices: client.Devices{
				Disks: []client.Disk{
					client.Disk{
						Name: d.Get("disk_name").(string),
						DiskType: client.DiskType{
							Bus: d.Get("disk_bus").(string),
						},
					},
				},
			},
			Machine: client.Machine{
				Type: "q35",
			},
			Resources: client.Resources{
				Requests: client.Requests{
					Memory: d.Get("memory").(string),
				},
			},
		},
		Networks: []client.Network{
			client.Network{
				Name: "default",
				Pod:  client.Pod{},
			},
		},
		Volumes: []client.Volume{
			client.Volume{
				Name: d.Get("volume_name").(string),
				ContainerDisk: client.ContainerDisk{
					Image: d.Get("image").(string),
				},
				CloudInitNoCloud: client.CloudInitNoCloud{
					UserData: d.Get("user_data").(string),
				},
			},
		},
	}

	return client.Spec{
		Status:       d.Get("status").(string),
		SpecTemplate: template,
	}
}

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

	err = c.CreateVirtualMachine(cluster.Metadata.UID, virtualMachine)
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
