package spectrocloud

import (
	"context"
	"fmt"
	"github.com/spectrocloud/hapi/apiutil/transport"
	"github.com/spectrocloud/hapi/models"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-sdk-go/client"
)

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
	//Deleting VM
	"Terminating",
	"Deleted",
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
			if err.(*transport.TransportError).HttpCode == 500 && strings.Contains(err.(*transport.TransportError).Payload.Message, fmt.Sprintf("Failed to get virtual machine '%s'", vmName)) {
				emptyVM := &models.V1ClusterVirtualMachine{}
				return emptyVM, "Deleted", nil
			} else {
				return nil, "", err
			}
		}
		return vm, vm.Status.PrintableStatus, nil
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

func flattenVMNetwork(netModel []*models.V1VMNetwork) []interface{} {
	result := make([]interface{}, 0)
	netSpec := make(map[string]interface{})
	var networks []interface{}
	for _, n := range netModel {
		networks = append(networks, map[string]interface{}{
			"name": n.Name,
		})
	}
	netSpec["network"] = networks
	result = append(result, netSpec)
	return result
}

func flattenVMVolumes(volumeModel []*models.V1VMVolume) []interface{} {
	result := make([]interface{}, 0)
	volumeSpec := make(map[string]interface{})
	var volume []interface{}
	for _, v := range volumeModel {
		if v.ContainerDisk != nil {
			volume = append(volume, map[string]interface{}{
				"name": v.Name,
				"container_disk": []interface{}{map[string]interface{}{
					"image_url": v.ContainerDisk.Image,
				}},
			})
		}
		if v.CloudInitNoCloud != nil {
			volume = append(volume, map[string]interface{}{
				"name": v.Name,
				"cloud_init_no_cloud": []interface{}{map[string]interface{}{
					"user_data": v.CloudInitNoCloud.UserData,
				}},
			})
		}
	}
	volumeSpec["volume"] = volume
	result = append(result, volumeSpec)
	return result
}

func flattenVMDevices(d *schema.ResourceData, vmDevices *models.V1VMDevices) []interface{} {
	result := make([]interface{}, 0)
	devices := make(map[string]interface{})
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
		devices["disk"] = disks
	}

	//set back interface
	if _, ok := d.GetOk("devices"); ok && vmDevices.Interfaces != nil {
		var interfaces []interface{}
		for _, inter := range vmDevices.Interfaces {
			if inter != nil {
				interfaces = append(interfaces, map[string]interface{}{
					"name": inter.Name,
				})
			}
		}
		devices["interface"] = interfaces
	}
	result = append(result, devices)
	return result
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

	//Handling Network
	var vmNetworks []*models.V1VMNetwork
	vmNetworks = prepareNetworkSpec(d)

	// Handling Volume
	var vmVolumes []*models.V1VMVolume
	vmVolumes = prepareVolumeSpec(d)

	// Handling Disk
	var vmDisks []*models.V1VMDisk
	var vmInterfaces []*models.V1VMInterface
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
	if !d.Get("run_on_launch").(bool) {
		vmSpec.RunStrategy = "Manual"
	}
	return vmSpec
}

func toVirtualMachineUpdateRequest(d *schema.ResourceData, vm *models.V1ClusterVirtualMachine) (bool, bool, *models.V1ClusterVirtualMachine, error) {
	requireUpdate := false
	needRestart := false
	if d.HasChange("cpu_cores") {
		vm.Spec.Template.Spec.Domain.CPU.Cores = int64(d.Get("cpu_cores").(int))
		requireUpdate = true
		needRestart = true
	}
	if d.HasChange("memory") {
		vm.Spec.Template.Spec.Domain.Resources.Requests = map[string]interface{}{
			"memory": d.Get("memory").(string),
		}
		requireUpdate = true
		needRestart = true
	}
	//if _, ok := d.GetOk("image_url"); ok && d.HasChange("image_url") {
	//	vm.Metadata.Namespace = d.Get("namespace").(string)
	//	requireUpdate = true
	//	needRestart = true
	//}
	if _, ok := d.GetOk("labels"); ok && d.HasChange("labels") {
		vm.Metadata.Labels = toVMLabels(d)
		requireUpdate = true
	}
	if _, ok := d.GetOk("annotations"); ok && d.HasChange("annotations") {
		vm.Metadata.Annotations = toVMUpdateAnnotations(vm.Metadata.Annotations, d)
		requireUpdate = true
	}
	if _, ok := d.GetOk("volume_spec"); ok && d.HasChange("volume_spec") {
		vm.Spec.Template.Spec.Volumes = prepareVolumeSpec(d)
		requireUpdate = true
		needRestart = true
	}
	if _, ok := d.GetOk("network_spec"); ok && d.HasChange("network_spec") {
		vm.Spec.Template.Spec.Networks = prepareNetworkSpec(d)
		requireUpdate = true
		needRestart = true
	}
	if _, ok := d.GetOk("devices"); ok && d.HasChange("devices") {
		vm.Spec.Template.Spec.Domain.Devices.Disks, vm.Spec.Template.Spec.Domain.Devices.Interfaces = prepareDevices(d)
		requireUpdate = true
		needRestart = true
	}
	if run, ok := d.GetOk("run_on_launch"); ok && d.HasChange("run_on_launch") {
		vm.Spec.Running = run.(bool)
		if run.(bool) {
			vm.Spec.RunStrategy = ""
		} else {
			vm.Spec.RunStrategy = "Manual"
		}

	}

	// There is issue in Ally side, team asked as to explicitly make deletion-time to nil before put operation, after fix will remove.
	vm.Spec.Template.Metadata.DeletionTimestamp = nil
	vm.Metadata.DeletionTimestamp = nil
	return requireUpdate, needRestart, vm, nil
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

func toVMUpdateAnnotations(existingAnnotation map[string]string, d *schema.ResourceData) map[string]string {
	if _, ok := d.GetOk("annotations"); ok {
		for k, a := range d.Get("annotations").(map[string]interface{}) {
			existingAnnotation[k] = a.(string)
		}
		return existingAnnotation
	} else {
		return nil
	}
}
