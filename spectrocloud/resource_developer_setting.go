package spectrocloud

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/spectrocloud/terraform-provider-spectrocloud/spectrocloud/constants"
)

func resourceDeveloperSetting() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceDeveloperSettingCreate,
		ReadContext:   resourceDeveloperSettingRead,
		UpdateContext: resourceDeveloperSettingUpdate,
		DeleteContext: resourceDeveloperSettingDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceDeveloperSettingImport,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},
		SchemaVersion: 2,

		Schema: map[string]*schema.Schema{
			"virtual_clusters_limit": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      2,
				ValidateFunc: validation.IntBetween(1, 1000),
				Description:  "Specifies the number of virtual clusters to be created.",
			},
			"cpu": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      12,
				ValidateFunc: validation.IntBetween(4, 1000),
				Description:  "Defines the number of CPU cores allocated to the cluster.",
			},
			"memory": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      16,
				ValidateFunc: validation.IntBetween(4, 1000),
				Description:  "Specifies the amount of memory (in GiB) allocated to the cluster.",
			},
			"storage": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      20,
				ValidateFunc: validation.IntBetween(2, 100000),
				Description:  "Defines the storage capacity (in GiB) allocated to the cluster.",
			},
			"hide_system_cluster_group": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "If set to `true`, hides the system cluster.",
			},
		},
	}
}

func toDeveloperSetting(d *schema.ResourceData) (*models.V1DeveloperCredit, *models.V1TenantEnableClusterGroup) {
	cpuInt := d.Get("cpu").(int)
	memoryInt := d.Get("memory").(int)
	storageInt := d.Get("storage").(int)
	virtualClustersLimitInt := d.Get("virtual_clusters_limit").(int)

	// Check bounds for int32 conversion
	if cpuInt > constants.Int32MaxValue || memoryInt > constants.Int32MaxValue || storageInt > constants.Int32MaxValue || virtualClustersLimitInt > constants.Int32MaxValue {
		// Return default values if any value is out of range
		return &models.V1DeveloperCredit{
				CPU:                  12,
				MemoryGiB:            16,
				StorageGiB:           20,
				VirtualClustersLimit: 2,
			}, &models.V1TenantEnableClusterGroup{
				HideSystemClusterGroups: false,
			}
	}

	devCredit := &models.V1DeveloperCredit{
		CPU:                  SafeInt32(cpuInt),
		MemoryGiB:            SafeInt32(memoryInt),
		StorageGiB:           SafeInt32(storageInt),
		VirtualClustersLimit: SafeInt32(virtualClustersLimitInt),
	}
	sysClusterGroupPref := &models.V1TenantEnableClusterGroup{
		HideSystemClusterGroups: d.Get("hide_system_cluster_group").(bool),
	}
	return devCredit, sysClusterGroupPref
}

func toDeveloperSettingDefault(d *schema.ResourceData) (*models.V1DeveloperCredit, *models.V1TenantEnableClusterGroup) {
	return &models.V1DeveloperCredit{
			CPU:                  12,
			MemoryGiB:            16,
			StorageGiB:           20,
			VirtualClustersLimit: 2,
		}, &models.V1TenantEnableClusterGroup{
			HideSystemClusterGroups: false,
		}
}

func flattenDeveloperSetting(devSetting *models.V1DeveloperCredit, sysClusterGroupPref *models.V1TenantEnableClusterGroup, d *schema.ResourceData) error {
	if err := d.Set("virtual_clusters_limit", devSetting.VirtualClustersLimit); err != nil {
		return err
	}
	if err := d.Set("cpu", devSetting.CPU); err != nil {
		return err
	}
	if err := d.Set("memory", devSetting.MemoryGiB); err != nil {
		return err
	}
	if err := d.Set("storage", devSetting.StorageGiB); err != nil {
		return err
	}
	if err := d.Set("hide_system_cluster_group", sysClusterGroupPref.HideSystemClusterGroups); err != nil {
		return err
	}
	return nil
}

func resourceDeveloperSettingCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := getV1ClientWithResourceContext(m, tenantString)
	var diags diag.Diagnostics
	devSettings, sysClusterGroupPref := toDeveloperSetting(d)
	tenantUID, err := c.GetTenantUID()
	if err != nil {
		return diag.FromErr(err)
	}
	// For developer setting we don't have support for creation it's always an update
	err = c.UpdateDeveloperSetting(tenantUID, devSettings)
	if err != nil {
		return diag.FromErr(err)
	}
	err = c.UpdateSystemClusterGroupPreference(tenantUID, sysClusterGroupPref)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId("default-dev-setting-id")
	return diags
}

func resourceDeveloperSettingRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := getV1ClientWithResourceContext(m, tenantString)
	var diags diag.Diagnostics
	tenantUID, err := c.GetTenantUID()
	if err != nil {
		return handleReadError(d, err, diags)
	}
	respDevSettings, err := c.GetDeveloperSetting(tenantUID)
	if err != nil {
		return handleReadError(d, err, diags)
	}
	respSysClusterGroupPref, err := c.GetSystemClusterGroupPreference(tenantUID)
	if err != nil {
		return handleReadError(d, err, diags)
	}
	// handling case for cross-plane for singleton resource
	if d.Id() != "default-dev-setting-id" {
		d.SetId("")
		return diags
	}
	err = flattenDeveloperSetting(respDevSettings, respSysClusterGroupPref, d)
	if err != nil {
		return diag.FromErr(err)
	}
	return diags
}

func resourceDeveloperSettingUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := getV1ClientWithResourceContext(m, tenantString)
	var diags diag.Diagnostics
	devSettings, sysClusterGroupPref := toDeveloperSetting(d)
	tenantUID, err := c.GetTenantUID()
	if err != nil {
		return diag.FromErr(err)
	}
	// For developer setting we don't have support for creation it's always an update
	err = c.UpdateDeveloperSetting(tenantUID, devSettings)
	if err != nil {
		return diag.FromErr(err)
	}
	err = c.UpdateSystemClusterGroupPreference(tenantUID, sysClusterGroupPref)
	if err != nil {
		return diag.FromErr(err)
	}
	return diags
}

func resourceDeveloperSettingDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := getV1ClientWithResourceContext(m, tenantString)
	var diags diag.Diagnostics
	// We can't delete the base developer setting, instead we are setting it to default
	devSettings, sysClusterGroupPref := toDeveloperSettingDefault(d)
	tenantUID, err := c.GetTenantUID()
	if err != nil {
		return diag.FromErr(err)
	}
	// For developer setting we don't have support for creation it's always an update
	err = c.UpdateDeveloperSetting(tenantUID, devSettings)
	if err != nil {
		return diag.FromErr(err)
	}
	err = c.UpdateSystemClusterGroupPreference(tenantUID, sysClusterGroupPref)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId("")
	return diags
}

func resourceDeveloperSettingImport(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	c := getV1ClientWithResourceContext(m, "tenant")
	var diags diag.Diagnostics
	givenTenantId := d.Id()
	actualTenantId, err := c.GetTenantUID()
	if err != nil {
		return nil, err
	}
	if givenTenantId != actualTenantId {
		return nil, fmt.Errorf("tenant id is not valid with current user: %v", diags)
	}
	diags = resourceDeveloperSettingRead(ctx, d, m)
	if diags.HasError() {
		return nil, fmt.Errorf("could not read developer settings for import: %v", diags)
	}
	return []*schema.ResourceData{d}, nil
}
