package spectrocloud

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/spectrocloud/hapi/apiutil/transport"
	"github.com/spectrocloud/hapi/models"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-sdk-go/client"

	"github.com/spectrocloud/terraform-provider-spectrocloud/types"
)

var resourceVirtualMachineCreatePendingStates = []string{
	"Stopped",
	"Starting",
	"Creating",
	"Provisioning",
	"Created",
	"WaitingForVolumeBinding",
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
	vm, err := c.GetVirtualMachine(clusterUid, namespace, vmName)
	if err != nil {
		return diags, true
	}

	if _, found := vm.Metadata.Labels["skip_vms"]; found {
		return diags, true
	}

	stateConf := &retry.StateChangeConf{
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

func resourceVirtualMachineStateRefreshFunc(c *client.V1Client, clusterUid string, vmName string, vmNamespace string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		//cluster, err := c.GetCluster(clusterUid)
		//if err != nil {
		//	return nil, "", err
		//} else if cluster == nil {
		//	return nil, "Deleted", nil
		//}
		vm, err := c.GetVirtualMachine(clusterUid, vmNamespace, vmName)
		if err != nil {
			if err.(*transport.TransportError).HttpCode == 500 && strings.Contains(err.(*transport.TransportError).Payload.Message, fmt.Sprintf("Failed to get virtual machine '%s'", vmName)) {
				emptyVM := &models.V1ClusterVirtualMachine{}
				return emptyVM, "Deleted", nil
			} else {
				return nil, "", err
			}
		}
		if vm == nil {
			emptyVM := &models.V1ClusterVirtualMachine{}
			return emptyVM, "", nil
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
	var nics []interface{}
	for _, nic := range netModel {
		nicSpec := make(map[string]interface{})
		nicSpec["name"] = nic.Name
		if nic.Pod != nil {
			nicSpec["network_type"] = "pod"
		} else {
			if nic.Multus != nil {
				multusSpec := make(map[string]interface{})
				multusSpec["network_name"] = *nic.Multus.NetworkName
				multusSpec["default"] = nic.Multus.Default
				nicSpec["multus"] = []interface{}{multusSpec}
			}
		}
		nics = append(nics, nicSpec)
	}
	netSpec["nic"] = nics
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
		if v.DataVolume != nil {
			volume = append(volume, map[string]interface{}{
				"name": v.Name,
				"data_volume": []interface{}{map[string]interface{}{
					"storage": "3Gi",
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

	if _, ok := d.GetOk("devices"); ok && vmDevices.Interfaces != nil {
		var interfaces []interface{}
		for _, iface := range vmDevices.Interfaces {
			if iface != nil {
				var ifaceType string
				switch {
				case iface.Bridge != nil:
					ifaceType = "bridge"
				case iface.Masquerade != nil:
					ifaceType = "masquerade"
				case iface.Macvtap != nil:
					ifaceType = "macvtap"
				}
				interfaces = append(interfaces, map[string]interface{}{
					"name":  iface.Name,
					"type":  ifaceType,
					"model": iface.Model,
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
	vmNetworks := prepareNetworkSpec(d)

	// Handling Volume
	vmVolumes := prepareVolumeSpec(d)

	// Handling Disk
	var vmDisks []*models.V1VMDisk
	var vmInterfaces []*models.V1VMInterface
	vmDisks, vmInterfaces = prepareDevices(d)

	vmSpec := &models.V1ClusterVirtualMachineSpec{
		DataVolumeTemplates: toDataVolumeTemplates(d),
		Running:             d.Get("run_on_launch").(bool),
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

func toDataVolumeTemplates(d *schema.ResourceData) []*models.V1VMDataVolumeTemplateSpec {
	volumeSpec := d.Get("volume_spec").(*schema.Set)
	var dataVolumeTemplates []*models.V1VMDataVolumeTemplateSpec

	if volumeSpec != nil {
		volumeSpecList := volumeSpec.List()
		for _, volume := range volumeSpecList {
			volumeMap := volume.(map[string]interface{})
			if volumeData, ok := volumeMap["volume"]; ok {
				volumeDataList := volumeData.([]interface{})
				for _, dataVolume := range volumeDataList {
					dataVolumeMap := dataVolume.(map[string]interface{})
					if _, ok := dataVolumeMap["data_volume"]; ok && len(dataVolumeMap["data_volume"].(*schema.Set).List()) > 0 {
						dataVolumeTemplates = append(dataVolumeTemplates, toDataVolumeTemplateSpecCreateRequest(dataVolumeMap["data_volume"], dataVolumeMap["name"].(string), d.Get("name").(string)))
					}
				}
			}
		}
	}
	return dataVolumeTemplates
}

func toDataVolumeTemplateSpecCreateRequest(dataVolumeSet interface{}, name string, vmname string) *models.V1VMDataVolumeTemplateSpec {
	dataVolumeList := dataVolumeSet.(*schema.Set).List()

	for _, dataVolume := range dataVolumeList {
		volume := dataVolume.(map[string]interface{})
		storage := volume["storage"].(string)

		dataVolumeTemplate := &models.V1VMDataVolumeTemplateSpec{
			Metadata: &models.V1VMObjectMeta{
				OwnerReferences: []*models.V1VMOwnerReference{
					{
						APIVersion: types.Ptr("kubevirt.io/v1"),
						Kind:       types.Ptr("VirtualMachine"),
						Name:       types.Ptr(vmname),
						UID:        types.Ptr(""),
					},
				},
				Name: "disk-0-vol",
			},
			Spec: &models.V1VMDataVolumeSpec{
				Storage: toV1VMStorageSpec(storage),
				Pvc:     toV1VMPersistentVolumeClaimSpec(storage),
				Source: &models.V1VMDataVolumeSource{
					Blank: make(map[string]interface{}),
				},
			},
		}

		return dataVolumeTemplate
	}

	return &models.V1VMDataVolumeTemplateSpec{}
}

func toV1VMPersistentVolumeClaimSpec(storage string) *models.V1VMPersistentVolumeClaimSpec {
	return &models.V1VMPersistentVolumeClaimSpec{
		Resources: &models.V1VMCoreResourceRequirements{
			Requests: map[string]models.V1VMQuantity{
				"storage": models.V1VMQuantity(storage),
			},
		},
		StorageClassName: "sumit-storage-class",
		AccessModes:      []string{"ReadWriteOnce"},
	}
}

func toV1VMStorageSpec(storage string) *models.V1VMStorageSpec {
	return &models.V1VMStorageSpec{
		Resources: &models.V1VMCoreResourceRequirements{
			Requests: map[string]models.V1VMQuantity{
				"storage": models.V1VMQuantity(storage),
			},
		},
		StorageClassName: "spectro-storage-class",
		VolumeMode:       "Block",
		AccessModes:      []string{"ReadWriteOnce"},
	}
}

func toVirtualMachineUpdateRequest(d *schema.ResourceData, vm *models.V1ClusterVirtualMachine) (bool, bool, *models.V1ClusterVirtualMachine, error) {
	requireUpdate := false
	needRestart := false
	if d.HasChange("cpu") {
		vm.Spec.Template.Spec.Domain.CPU.Cores = int64(d.Get("cpu").(int))
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
	if _, ok := d.GetOk("volume"); ok && d.HasChange("volume") {
		vm.Spec.Template.Spec.Volumes = prepareVolumeSpec(d)
		requireUpdate = true
		needRestart = true
	}
	if _, ok := d.GetOk("network"); ok && d.HasChange("network") {
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
