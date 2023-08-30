package spectrocloud

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/hapi/models"
	"github.com/spectrocloud/palette-sdk-go/client"

	"github.com/spectrocloud/terraform-provider-spectrocloud/spectrocloud/convert"
	"github.com/spectrocloud/terraform-provider-spectrocloud/spectrocloud/kubevirt/schema/datavolume"
	"github.com/spectrocloud/terraform-provider-spectrocloud/spectrocloud/kubevirt/utils"
	"github.com/spectrocloud/terraform-provider-spectrocloud/types"
)

func resourceKubevirtDataVolume() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceKubevirtDataVolumeCreate,
		ReadContext:   resourceKubevirtDataVolumeRead,
		UpdateContext: resourceKubevirtDataVolumeUpdate,
		DeleteContext: resourceKubevirtDataVolumeDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(40 * time.Minute),
			Delete: schema.DefaultTimeout(5 * time.Minute),
		},
		Schema: datavolume.DataVolumeFields(),
	}
}

func resourceKubevirtDataVolumeCreate(ctx context.Context, resourceData *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.V1Client)

	var diags diag.Diagnostics
	dv, err := datavolume.FromResourceData(resourceData)
	if err != nil {
		return diag.FromErr(err)
	}

	// Extract "add_volume_options" from the Terraform schema
	addVolumeOptionsData := resourceData.Get("add_volume_options").([]interface{})
	AddVolumeOptions := ExpandAddVolumeOptions(addVolumeOptionsData)

	hapiVolume, err := convert.ToHapiVolume(dv, AddVolumeOptions)
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[INFO] Creating new data volume: %#v", dv)
	// Warning or errors can be collected in a slice type
	clusterUid := resourceData.Get("cluster_uid").(string)
	ClusterContext := resourceData.Get("cluster_context").(string)
	_, err = c.GetCluster(ClusterContext, clusterUid)
	if err != nil {
		return diag.FromErr(err)
	}

	if resourceData.Get("vm_name") == nil {
		return diag.FromErr(errors.New("vm_name is required"))
	}
	vmName := resourceData.Get("vm_name").(string)

	if resourceData.Get("vm_namespace") == nil {
		return diag.FromErr(errors.New("vm_namespace is required"))
	}
	vmNamespace := resourceData.Get("vm_namespace").(string)

	if _, err := c.CreateDataVolume(ClusterContext, clusterUid, vmName, hapiVolume); err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[INFO] Submitted new data volume: %#v", dv)
	if err := datavolume.ToResourceData(*dv, resourceData); err != nil {
		return diag.FromErr(err)
	}
	resourceData.SetId(utils.BuildIdDV(ClusterContext, clusterUid, vmNamespace, vmName, hapiVolume.DataVolumeTemplate.Metadata))

	return diags
}

func resourceKubevirtDataVolumeRead(ctx context.Context, resourceData *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cli := (meta).(*client.V1Client)

	scope, clusterUid, namespace, vm_name, _, err := utils.IdPartsDV(resourceData.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[INFO] Reading virtual machine %s", vm_name)

	hapiVM, err := cli.GetVirtualMachine(scope, clusterUid, namespace, vm_name)
	if err != nil {
		log.Printf("[DEBUG] Received error: %#v", err)
		return diag.FromErr(err)
	}
	if hapiVM == nil {
		return diag.FromErr(fmt.Errorf("virtual machine not found %s, %s, %s to read data volume", clusterUid, namespace, vm_name))
	}

	metadataSlice := resourceData.Get("metadata").([]interface{})
	rd_metadata := metadataSlice[0].(map[string]interface{})
	rd_metadataName := rd_metadata["name"].(string)
	rd_metadataNamespace := rd_metadata["namespace"].(string)
	// Read data volume templates from vm.Spec.DataVolumeTemplates and filter by name
	for _, dv := range hapiVM.Spec.DataVolumeTemplates {
		name := dv.Metadata.Name
		namespace := dv.Metadata.Namespace

		if name == rd_metadataName && namespace == rd_metadataNamespace {
			kvVolume, err := convert.FromHapiVolume(&models.V1VMAddVolumeEntity{
				DataVolumeTemplate: dv,
			})
			if err != nil {
				return diag.FromErr(err)
			}
			err = datavolume.ToResourceData(*kvVolume, resourceData)
			if err != nil {
				return diag.FromErr(err)
			}
			break
		}
	}
	return diag.Diagnostics{}
}

func resourceKubevirtDataVolumeUpdate(ctx context.Context, resourceData *schema.ResourceData, m interface{}) diag.Diagnostics {
	// implement update as delete followed by create
	if err := resourceKubevirtDataVolumeDelete(ctx, resourceData, m); err != nil {
		return err
	}

	return resourceKubevirtDataVolumeCreate(ctx, resourceData, m)

}

func resourceKubevirtDataVolumeDelete(ctx context.Context, resourceData *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.V1Client)

	var diags diag.Diagnostics
	scope, clusterUid, namespace, vm_name, vol_name, err := utils.IdPartsDV(resourceData.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = c.GetCluster(scope, clusterUid)
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[INFO] Deleting data volume: %#v", vm_name)
	if err := c.DeleteDataVolume(scope, clusterUid, namespace, vm_name, &models.V1VMRemoveVolumeEntity{
		Persist: true,
		RemoveVolumeOptions: &models.V1VMRemoveVolumeOptions{
			Name: types.Ptr(vol_name),
		},
	}); err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[INFO] data volume %s deleted", vm_name)

	resourceData.SetId("")
	return diags
}

func FlattenAddVolumeOptions(addVolumeOptions *models.V1VMAddVolumeOptions) []interface{} {
	if addVolumeOptions == nil {
		return []interface{}{}
	}

	result := map[string]interface{}{
		"name": addVolumeOptions.Name,
	}

	if addVolumeOptions.Disk != nil && addVolumeOptions.Disk.Disk != nil {
		result["disk"] = []interface{}{
			map[string]interface{}{
				"name": addVolumeOptions.Disk.Name,
				"bus":  addVolumeOptions.Disk.Disk.Bus,
			},
		}
	}

	if addVolumeOptions.VolumeSource != nil && addVolumeOptions.VolumeSource.DataVolume != nil {
		result["volume_source"] = []interface{}{
			map[string]interface{}{
				"data_volume": []interface{}{
					map[string]interface{}{
						"name":         addVolumeOptions.VolumeSource.DataVolume.Name,
						"hotpluggable": addVolumeOptions.VolumeSource.DataVolume.Hotpluggable,
					},
				},
			},
		}
	}

	return []interface{}{result}
}

func ExpandAddVolumeOptions(addVolumeOptions []interface{}) *models.V1VMAddVolumeOptions {
	if len(addVolumeOptions) == 0 || addVolumeOptions[0] == nil {
		return nil
	}

	m := addVolumeOptions[0].(map[string]interface{})

	result := &models.V1VMAddVolumeOptions{
		Name: types.Ptr(m["name"].(string)),
	}

	if diskList, ok := m["disk"].([]interface{}); ok && len(diskList) > 0 {
		if diskMap, ok := diskList[0].(map[string]interface{}); ok {
			result.Disk = &models.V1VMDisk{
				Name: types.Ptr(diskMap["name"].(string)),
				Disk: &models.V1VMDiskTarget{
					Bus: diskMap["bus"].(string),
				},
			}
		}
	}

	if volumeSourceList, ok := m["volume_source"].([]interface{}); ok && len(volumeSourceList) > 0 {
		if volumeSourceMap, ok := volumeSourceList[0].(map[string]interface{}); ok {
			if dataVolumeList, ok := volumeSourceMap["data_volume"].([]interface{}); ok && len(dataVolumeList) > 0 {
				if dataVolumeMap, ok := dataVolumeList[0].(map[string]interface{}); ok {
					result.VolumeSource = &models.V1VMHotplugVolumeSource{
						DataVolume: &models.V1VMCoreDataVolumeSource{
							Name:         types.Ptr(dataVolumeMap["name"].(string)),
							Hotpluggable: dataVolumeMap["hotpluggable"].(bool),
						},
					}
				}
			}
		}
	}

	return result
}
