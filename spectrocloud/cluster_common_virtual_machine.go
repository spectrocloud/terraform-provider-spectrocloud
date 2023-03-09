package spectrocloud

import (
	"context"
	"fmt"
	"github.com/spectrocloud/hapi/models"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-sdk-go/client"
)

// fix as needed with real statuses
var resourceVirtualMachineCreatePendingStates = []string{
	"Stopped",
	"Starting",
	"Creating",
	"Created",
	"Running",
	// Restart|Stop
	"Stopping",
	// Pause
	"Pausing",
	// Migration
	"Migrating",
}

func waitForVirtualMachineToTargetState(ctx context.Context, d *schema.ResourceData, clusterUid string, vmName string, namespace string, diags diag.Diagnostics, c *client.V1Client, state string, targetState string) (diag.Diagnostics, bool) {
	vm, err := c.GetVirtualMachine(clusterUid, vmName, namespace)
	if err != nil {
		return diags, true
	}

	if _, found := vm.Metadata.Labels["skip_vms"]; found {
		return diags, true
	}

	stateConf := &resource.StateChangeConf{
		Pending:    resourceVirtualMachineCreatePendingStates,
		Target:     []string{targetState},
		Refresh:    resourceVirtualMachineStateRefreshFunc(c, clusterUid, vmName, namespace),
		Timeout:    d.Timeout(state) - 1*time.Minute,
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second,
	}

	// Wait, catching any errors
	_, err = stateConf.WaitForStateContext(ctx)
	if err != nil {
		return diag.FromErr(err), true
	}
	return nil, false
}

func resourceVirtualMachineStateRefreshFunc(c *client.V1Client, clusterUid string, vmName string, vmNamespace string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		//cluster, err := c.GetCluster(clusterUid)
		//if err != nil {
		//	return nil, "", err
		//} else if cluster == nil {
		//	return nil, "Deleted", nil
		//}
		vm, err := c.GetVirtualMachine(clusterUid, vmName, vmNamespace)
		if err != nil {
			return nil, "", err
		} else if vm == nil {
			return nil, "Deleted", nil
		}

		return vm, vm.Status.PrintableStatus, nil
	}
}

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
		//var vmTempVar = new(string)

		for _, d := range device.(*schema.Set).List() {
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
	} else {
		return prepareDefaultDevices()
	}
}

func prepareDefaultVolume(d *schema.ResourceData) []*models.V1VMVolume {
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

func prepareVolume(d *schema.ResourceData) []*models.V1VMVolume {
	if volumes, ok := d.GetOk("volume"); ok {
		var vmVolumes []*models.V1VMVolume
		for _, vol := range volumes.([]interface{}) {
			v := vol.(map[string]interface{})
			cDisk := v["container_disk"].(*schema.Set).List()
			cInit := v["cloud_init_no_cloud"].(*schema.Set).List()
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
	} else {
		return prepareDefaultVolume(d)
	}
}

func prepareDefaultNetwork() []*models.V1VMNetwork {
	var vmNetworks []*models.V1VMNetwork
	var networkName = new(string)
	*networkName = "default" // d.Get("network").(map[string]interface{})["name"].(string)
	vmNetworks = append(vmNetworks, &models.V1VMNetwork{
		Name: networkName,
		Pod:  &models.V1VMPodNetwork{},
	})
	return vmNetworks
}

func prepareNetwork(d *schema.ResourceData) []*models.V1VMNetwork {
	if network, ok := d.GetOk("network"); ok {
		var vmNetworks []*models.V1VMNetwork
		var networkName = new(string)
		for _, n := range network.([]interface{}) {
			*networkName = n.(map[string]interface{})["name"].(string)
			vmNetworks = append(vmNetworks, &models.V1VMNetwork{
				Name: networkName,
				Pod:  &models.V1VMPodNetwork{},
			})
		}
		return vmNetworks
	} else {
		return prepareDefaultNetwork()
	}
}

func toVMLabels(d *schema.ResourceData) map[string]string {
	labels := make(map[string]string)
	if _, ok := d.GetOk("labels"); ok {
		for _, t := range d.Get("labels").(*schema.Set).List() {
			tag := t.(string)
			if strings.Contains(tag, "=") {
				labels[strings.Split(tag, "=")[0]] = strings.Split(tag, "=")[1]
			} else {
				labels[tag] = "spectro__tag"
			}
		}
		return labels
	} else {
		return nil
	}
}

func toVMAnnotations(d *schema.ResourceData) map[string]string {
	annotation := make(map[string]string)
	if _, ok := d.GetOk("annotations"); ok {
		for k, a := range d.Get("annotations").(map[string]interface{}) {
			annotation[k] = a.(string)
		}
		return annotation
	} else {
		return nil
	}
}

func flattenVMAnnotations(annotation map[string]string, d *schema.ResourceData) map[string]interface{} {

	if len(annotation) > 0 {
		annot := map[string]interface{}{}
		oldAnn := d.Get("annotations").(map[string]interface{})
		for k, v := range annotation {
			if oldAnn[k] != nil {
				annot[k] = v
			}
		}
		return annot
	} else {
		return nil
	}
}

func flattenVMLabels(labels map[string]string) []interface{} {
	tags := make([]interface{}, 0)
	if len(labels) > 0 {
		for k, v := range labels {
			if v == "spectro__tag" {
				tags = append(tags, k)
			} else {
				tags = append(tags, fmt.Sprintf("%s=%s", k, v))
			}
		}
		return tags
	} else {
		return nil
	}
}

func flattenVMNetwork(network []*models.V1VMNetwork) []interface{} {
	var net []interface{}
	for _, n := range network {
		net = append(net, map[string]interface{}{
			"name": n.Name,
		})
	}
	return net
}

func flattenVMVolumes(volumes []*models.V1VMVolume) []interface{} {
	vol := make([]interface{}, 0)
	for _, v := range volumes {
		if v.ContainerDisk != nil {
			vol = append(vol, map[string]interface{}{
				"name": v.Name,
				"container_disk": []interface{}{map[string]interface{}{
					"image_url": v.ContainerDisk.Image,
				}},
			})
		}
		if v.CloudInitNoCloud != nil {
			vol = append(vol, map[string]interface{}{
				"name": v.Name,
				"cloud_init_no_cloud": []interface{}{map[string]interface{}{
					"user_data": v.CloudInitNoCloud.UserData,
				}},
			})
		}
	}
	return vol
}

func flattenVMDevices(d *schema.ResourceData, vmDevices *models.V1VMDevices) []interface{} {
	var devices []interface{}
	if _, ok := d.GetOk("devices"); ok && vmDevices.Disks != nil {
		var disks []interface{}
		for _, disk := range vmDevices.Disks {
			if disk != nil {
				disks = append(disks, map[string]interface{}{
					"name": disk.Name,
					"bus":  disk.Disk.Bus,
				})
			}
		}
		devices = append(devices, disks)
	}

	//set back interface
	if _, ok := d.GetOk("devices"); ok && vmDevices.Interfaces != nil {
		var interfaces []interface{}
		for _, inter := range vmDevices.Disks {
			if inter != nil {
				interfaces = append(interfaces, map[string]interface{}{
					"name": inter.Name,
				})
			}
		}
		devices = append(devices, interfaces)
	}
	return devices
}

func toVirtualMachineCreateRequest(d *schema.ResourceData) (*models.V1ClusterVirtualMachine, error) {
	vmBody := &models.V1ClusterVirtualMachine{
		APIVersion: "kubevirt.io/v1",
		Kind:       "VirtualMachine",
		Metadata: &models.V1VMObjectMeta{
			Name:        d.Get("name").(string),
			Namespace:   d.Get("namespace").(string),
			Labels:      toVMLabels(d),
			Annotations: toVMAnnotations(d),
		},
		Spec: toSpecCreateRequest(d),
	}
	return vmBody, nil
}

func toSpecCreateRequest(d *schema.ResourceData) *models.V1ClusterVirtualMachineSpec {

	var vmVolumes []*models.V1VMVolume
	var vmDisks []*models.V1VMDisk
	var vmInterfaces []*models.V1VMInterface
	var vmNetworks []*models.V1VMNetwork

	//Handling Network
	vmNetworks = prepareNetwork(d)
	// Handling Volume
	vmVolumes = prepareVolume(d)
	// Handling Disk
	vmDisks, vmInterfaces = prepareDevices(d)

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

func toVirtualMachineUpdateRequest(d *schema.ResourceData, vm *models.V1ClusterVirtualMachine) (bool, *models.V1ClusterVirtualMachine, error) {
	requireUpdate := false
	if d.HasChange("name") {
		vm.Metadata.Name = d.Get("name").(string)
		requireUpdate = true
	}
	if d.HasChange("namespace") {
		vm.Metadata.Namespace = d.Get("namespace").(string)
		requireUpdate = true
	}
	if d.HasChange("cpu_cores") {
		vm.Spec.Template.Spec.Domain.CPU.Cores = int64(d.Get("cpu_cores").(int))
		requireUpdate = true
	}
	if d.HasChange("memory") {
		vm.Spec.Template.Spec.Domain.Resources.Requests = map[string]interface{}{
			"memory": d.Get("memory").(string),
		}
		requireUpdate = true
	}
	if _, ok := d.GetOk("image_url"); ok && d.HasChange("image_url") {
		vm.Metadata.Namespace = d.Get("namespace").(string)
		requireUpdate = true
	}

	// There is issue in Ally side, team asked as to explicitly make deletion-time to nil before put operation, after fix will remove.
	vm.Spec.Template.Metadata.DeletionTimestamp = nil
	vm.Metadata.DeletionTimestamp = nil
	return requireUpdate, vm, nil
}
