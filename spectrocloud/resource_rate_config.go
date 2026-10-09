package spectrocloud

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/spectrocloud/palette-sdk-go/api/models"
)

// rateConfigID is the fixed identifier for the tenant rate config singleton resource.
const rateConfigID = "default-rate-config-id"

// Palette's built-in rate config, applied by the platform whenever a cloud is
// omitted from the update payload. Mirrored here so that destroying the
// resource restores the tenant to the platform defaults.
const (
	defaultCPUUnitPricePerHour        = 0.021811
	defaultGpuUnitPricePerHour        = 2.933908
	defaultMemoryUnitPriceGiBPerHour  = 0.002923
	defaultStorageUnitPriceGiBPerHour = 0.000028

	defaultComputeOptimizedComputeRate = 65
	defaultComputeOptimizedMemoryRate  = 35
	defaultMemoryOptimizedComputeRate  = 25
	defaultMemoryOptimizedMemoryRate   = 75
)

// publicCloudRateSchema returns the schema for a public cloud rate config, where
// the instance cost is split into compute and memory proportions.
func publicCloudRateSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"compute_optimized": {
				Type:        schema.TypeList,
				Required:    true,
				MaxItems:    1,
				Description: "Rate proportions applied to general purpose, compute optimized and GPU optimized instance types.",
				Elem:        cloudInstanceRateSchema(),
			},
			"memory_optimized": {
				Type:        schema.TypeList,
				Required:    true,
				MaxItems:    1,
				Description: "Rate proportions applied to memory optimized and storage optimized instance types.",
				Elem:        cloudInstanceRateSchema(),
			},
		},
	}
}

// cloudInstanceRateSchema returns the schema for a single public cloud instance
// category. The two proportions are percentages and must add up to 100.
func cloudInstanceRateSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"compute_rate_proportion": {
				Type:         schema.TypeFloat,
				Required:     true,
				ValidateFunc: validation.FloatBetween(0, 100),
				Description:  "Percentage of the instance price attributed to compute. Must add up to 100 together with `memory_rate_proportion`.",
			},
			"memory_rate_proportion": {
				Type:         schema.TypeFloat,
				Required:     true,
				ValidateFunc: validation.FloatBetween(0, 100),
				Description:  "Percentage of the instance price attributed to memory. Must add up to 100 together with `compute_rate_proportion`.",
			},
		},
	}
}

// privateCloudRateSchema returns the schema for a private cloud rate config,
// where each resource unit is priced directly.
func privateCloudRateSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"cpu_unit_price_per_hour": {
				Type:         schema.TypeFloat,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.FloatAtLeast(0),
				Description:  "Price in US dollars charged for one CPU core per hour.",
			},
			"gpu_unit_price_per_hour": {
				Type:         schema.TypeFloat,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.FloatAtLeast(0),
				Description:  "Price in US dollars charged for one GPU per hour.",
			},
			"memory_unit_price_gib_per_hour": {
				Type:         schema.TypeFloat,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.FloatAtLeast(0),
				Description:  "Price in US dollars charged for one GiB of memory per hour.",
			},
			"storage_unit_price_gib_per_hour": {
				Type:         schema.TypeFloat,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.FloatAtLeast(0),
				Description:  "Price in US dollars charged for one GiB of storage per hour.",
			},
		},
	}
}

func publicCloudRateAttribute(cloudName string) *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeList,
		Optional:    true,
		Computed:    true,
		MaxItems:    1,
		Description: fmt.Sprintf("Rate proportions used to estimate the cost of %s instances. When omitted, Palette applies its built-in proportions.", cloudName),
		Elem:        publicCloudRateSchema(),
	}
}

func privateCloudRateAttribute(cloudName string) *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeList,
		Optional:    true,
		Computed:    true,
		MaxItems:    1,
		Description: fmt.Sprintf("Unit prices used to estimate the cost of %s resources. When omitted, Palette applies its built-in rates.", cloudName),
		Elem:        privateCloudRateSchema(),
	}
}

func resourceRateConfig() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceRateConfigCreate,
		ReadContext:   resourceRateConfigRead,
		UpdateContext: resourceRateConfigUpdate,
		DeleteContext: resourceRateConfigDelete,
		Description:   "Resource for managing the tenant-level cloud rate config in Spectro Cloud. The rate config defines the unit prices Palette uses to estimate cluster cloud cost and usage cost.",
		Importer: &schema.ResourceImporter{
			StateContext: resourceRateConfigImport,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},
		SchemaVersion: 1,

		// The API model also carries `edge` and `edgeNative` rate configs.
		// Neither is exposed here: Palette surfaces neither in the UI, and
		// `edge` is dropped on update entirely.
		Schema: map[string]*schema.Schema{
			"aws":               publicCloudRateAttribute("AWS"),
			"azure":             publicCloudRateAttribute("Azure"),
			"gcp":               publicCloudRateAttribute("GCP"),
			"vsphere":           privateCloudRateAttribute("VMware vSphere"),
			"maas":              privateCloudRateAttribute("MAAS"),
			"generic":           privateCloudRateAttribute("generic cloud"),
			"apache_cloudstack": privateCloudRateAttribute("Apache CloudStack"),
			"custom": {
				Type:        schema.TypeList,
				Optional:    true,
				Computed:    true,
				Description: "Unit prices applied to custom cloud types registered in the tenant. Each block configures one custom cloud.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"cloud_type": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Name of the custom cloud type the rates apply to. The custom cloud must already be registered in the tenant.",
						},
						"rate_config": {
							Type:        schema.TypeList,
							Required:    true,
							MaxItems:    1,
							Description: "Unit prices used to estimate the cost of resources running on this custom cloud.",
							Elem:        privateCloudRateSchema(),
						},
					},
				},
			},
		},

		CustomizeDiff: validateRateProportions,
	}
}

// rateConfigReader is the subset of *schema.ResourceDiff and
// *schema.ResourceData that validateRateProportionsIn needs.
type rateConfigReader interface {
	Get(key string) interface{}
}

// validateRateProportions rejects public cloud rate proportions that do not add
// up to 100 percent, which Palette would otherwise accept and silently
// mis-apply to cost estimates.
func validateRateProportions(ctx context.Context, diff *schema.ResourceDiff, m interface{}) error {
	return validateRateProportionsIn(diff)
}

func validateRateProportionsIn(d rateConfigReader) error {
	for _, cloud := range []string{"aws", "azure", "gcp"} {
		cloudBlocks, ok := d.Get(cloud).([]interface{})
		if !ok || len(cloudBlocks) == 0 || cloudBlocks[0] == nil {
			continue
		}
		cloudConfig := cloudBlocks[0].(map[string]interface{})
		for _, category := range []string{"compute_optimized", "memory_optimized"} {
			categoryBlocks, ok := cloudConfig[category].([]interface{})
			if !ok || len(categoryBlocks) == 0 || categoryBlocks[0] == nil {
				continue
			}
			rates := categoryBlocks[0].(map[string]interface{})
			compute, _ := rates["compute_rate_proportion"].(float64)
			memory, _ := rates["memory_rate_proportion"].(float64)
			if compute+memory != 100 {
				return fmt.Errorf("%s.%s: compute_rate_proportion (%v) and memory_rate_proportion (%v) must add up to 100, got %v",
					cloud, category, compute, memory, compute+memory)
			}
		}
	}
	return nil
}

func toRateConfig(d *schema.ResourceData) *models.V1RateConfig {
	return &models.V1RateConfig{
		Aws:              toPublicCloudRateConfig(d.Get("aws")),
		Azure:            toPublicCloudRateConfig(d.Get("azure")),
		Gcp:              toPublicCloudRateConfig(d.Get("gcp")),
		Vsphere:          toPrivateCloudRateConfig(d.Get("vsphere")),
		Maas:             toPrivateCloudRateConfig(d.Get("maas")),
		Generic:          toPrivateCloudRateConfig(d.Get("generic")),
		ApacheCloudstack: toPrivateCloudRateConfig(d.Get("apache_cloudstack")),
		Custom:           toCustomCloudRateConfigs(d.Get("custom")),
	}
}

// toRateConfigDefault builds the payload that restores every cloud to the
// Palette defaults. Custom clouds are re-defaulted by the platform when the
// list is empty.
func toRateConfigDefault() *models.V1RateConfig {
	publicCloud := func() *models.V1PublicCloudRateConfig {
		return &models.V1PublicCloudRateConfig{
			ComputeOptimized: &models.V1CloudInstanceRateConfig{
				ComputeRateProportion: defaultComputeOptimizedComputeRate,
				MemoryRateProportion:  defaultComputeOptimizedMemoryRate,
			},
			MemoryOptimized: &models.V1CloudInstanceRateConfig{
				ComputeRateProportion: defaultMemoryOptimizedComputeRate,
				MemoryRateProportion:  defaultMemoryOptimizedMemoryRate,
			},
		}
	}
	privateCloud := func() *models.V1PrivateCloudRateConfig {
		return &models.V1PrivateCloudRateConfig{
			CPUUnitPricePerHour:        defaultCPUUnitPricePerHour,
			GpuUnitPricePerHour:        defaultGpuUnitPricePerHour,
			MemoryUnitPriceGiBPerHour:  defaultMemoryUnitPriceGiBPerHour,
			StorageUnitPriceGiBPerHour: defaultStorageUnitPriceGiBPerHour,
		}
	}
	return &models.V1RateConfig{
		Aws:              publicCloud(),
		Azure:            publicCloud(),
		Gcp:              publicCloud(),
		Vsphere:          privateCloud(),
		Maas:             privateCloud(),
		Generic:          privateCloud(),
		ApacheCloudstack: privateCloud(),
		Custom:           []*models.V1CustomCloudRateConfig{},
	}
}

func toPublicCloudRateConfig(raw interface{}) *models.V1PublicCloudRateConfig {
	config := firstBlock(raw)
	if config == nil {
		return nil
	}
	// Palette dereferences both categories without a nil check, so always send both.
	return &models.V1PublicCloudRateConfig{
		ComputeOptimized: toCloudInstanceRateConfig(config["compute_optimized"]),
		MemoryOptimized:  toCloudInstanceRateConfig(config["memory_optimized"]),
	}
}

func toCloudInstanceRateConfig(raw interface{}) *models.V1CloudInstanceRateConfig {
	config := firstBlock(raw)
	if config == nil {
		return &models.V1CloudInstanceRateConfig{}
	}
	compute, _ := config["compute_rate_proportion"].(float64)
	memory, _ := config["memory_rate_proportion"].(float64)
	return &models.V1CloudInstanceRateConfig{
		ComputeRateProportion: float32(compute),
		MemoryRateProportion:  float32(memory),
	}
}

func toPrivateCloudRateConfig(raw interface{}) *models.V1PrivateCloudRateConfig {
	config := firstBlock(raw)
	if config == nil {
		return nil
	}
	cpu, _ := config["cpu_unit_price_per_hour"].(float64)
	gpu, _ := config["gpu_unit_price_per_hour"].(float64)
	memory, _ := config["memory_unit_price_gib_per_hour"].(float64)
	storage, _ := config["storage_unit_price_gib_per_hour"].(float64)
	return &models.V1PrivateCloudRateConfig{
		CPUUnitPricePerHour:        cpu,
		GpuUnitPricePerHour:        gpu,
		MemoryUnitPriceGiBPerHour:  memory,
		StorageUnitPriceGiBPerHour: storage,
	}
}

func toCustomCloudRateConfigs(raw interface{}) []*models.V1CustomCloudRateConfig {
	// The API field has no omitempty, so a nil slice marshals as null.
	customConfigs := make([]*models.V1CustomCloudRateConfig, 0)
	blocks, ok := raw.([]interface{})
	if !ok {
		return customConfigs
	}
	for _, block := range blocks {
		config, ok := block.(map[string]interface{})
		if !ok {
			continue
		}
		rateConfig := toPrivateCloudRateConfig(config["rate_config"])
		if rateConfig == nil {
			continue
		}
		cloudType, _ := config["cloud_type"].(string)
		customConfigs = append(customConfigs, &models.V1CustomCloudRateConfig{
			CloudType:  cloudType,
			RateConfig: rateConfig,
		})
	}
	return customConfigs
}

// firstBlock unwraps a MaxItems 1 block into its attribute map, returning nil
// when the block is absent or empty.
func firstBlock(raw interface{}) map[string]interface{} {
	blocks, ok := raw.([]interface{})
	if !ok || len(blocks) == 0 || blocks[0] == nil {
		return nil
	}
	config, ok := blocks[0].(map[string]interface{})
	if !ok {
		return nil
	}
	return config
}

func flattenRateConfig(rateConfig *models.V1RateConfig, d *schema.ResourceData) error {
	if rateConfig == nil {
		return nil
	}
	publicClouds := map[string]*models.V1PublicCloudRateConfig{
		"aws":   rateConfig.Aws,
		"azure": rateConfig.Azure,
		"gcp":   rateConfig.Gcp,
	}
	for attribute, cloud := range publicClouds {
		if err := d.Set(attribute, flattenPublicCloudRateConfig(cloud)); err != nil {
			return err
		}
	}
	privateClouds := map[string]*models.V1PrivateCloudRateConfig{
		"vsphere":           rateConfig.Vsphere,
		"maas":              rateConfig.Maas,
		"generic":           rateConfig.Generic,
		"apache_cloudstack": rateConfig.ApacheCloudstack,
	}
	for attribute, cloud := range privateClouds {
		if err := d.Set(attribute, flattenPrivateCloudRateConfig(cloud)); err != nil {
			return err
		}
	}
	return d.Set("custom", flattenCustomCloudRateConfigs(rateConfig.Custom))
}

func flattenPublicCloudRateConfig(cloud *models.V1PublicCloudRateConfig) []interface{} {
	if cloud == nil {
		return []interface{}{}
	}
	return []interface{}{
		map[string]interface{}{
			"compute_optimized": flattenCloudInstanceRateConfig(cloud.ComputeOptimized),
			"memory_optimized":  flattenCloudInstanceRateConfig(cloud.MemoryOptimized),
		},
	}
}

func flattenCloudInstanceRateConfig(instance *models.V1CloudInstanceRateConfig) []interface{} {
	if instance == nil {
		return []interface{}{}
	}
	return []interface{}{
		map[string]interface{}{
			"compute_rate_proportion": float64(instance.ComputeRateProportion),
			"memory_rate_proportion":  float64(instance.MemoryRateProportion),
		},
	}
}

func flattenPrivateCloudRateConfig(cloud *models.V1PrivateCloudRateConfig) []interface{} {
	if cloud == nil {
		return []interface{}{}
	}
	return []interface{}{
		map[string]interface{}{
			"cpu_unit_price_per_hour":         cloud.CPUUnitPricePerHour,
			"gpu_unit_price_per_hour":         cloud.GpuUnitPricePerHour,
			"memory_unit_price_gib_per_hour":  cloud.MemoryUnitPriceGiBPerHour,
			"storage_unit_price_gib_per_hour": cloud.StorageUnitPriceGiBPerHour,
		},
	}
}

func flattenCustomCloudRateConfigs(customConfigs []*models.V1CustomCloudRateConfig) []interface{} {
	flattened := make([]interface{}, 0, len(customConfigs))
	for _, customConfig := range customConfigs {
		if customConfig == nil {
			continue
		}
		flattened = append(flattened, map[string]interface{}{
			"cloud_type":  customConfig.CloudType,
			"rate_config": flattenPrivateCloudRateConfig(customConfig.RateConfig),
		})
	}
	return flattened
}

func resourceRateConfigCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := getV1ClientWithResourceContext(m, tenantString)
	var diags diag.Diagnostics
	tenantUID, err := c.GetTenantUID()
	if err != nil {
		return diag.FromErr(err)
	}
	// The rate config always exists for a tenant, so creation is an update.
	if err := c.UpdateRateConfig(tenantUID, toRateConfig(d)); err != nil {
		return diag.FromErr(err)
	}
	d.SetId(rateConfigID)
	return diags
}

func resourceRateConfigRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := getV1ClientWithResourceContext(m, tenantString)
	var diags diag.Diagnostics
	tenantUID, err := c.GetTenantUID()
	if err != nil {
		return handleReadError(d, err, diags)
	}
	rateConfig, err := c.GetRateConfig(tenantUID)
	if err != nil {
		return handleReadError(d, err, diags)
	}
	// handling case for cross-plane for singleton resource
	if d.Id() != rateConfigID {
		d.SetId("")
		return diags
	}
	if err := flattenRateConfig(rateConfig, d); err != nil {
		return diag.FromErr(err)
	}
	return diags
}

func resourceRateConfigUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := getV1ClientWithResourceContext(m, tenantString)
	var diags diag.Diagnostics
	tenantUID, err := c.GetTenantUID()
	if err != nil {
		return diag.FromErr(err)
	}
	if err := c.UpdateRateConfig(tenantUID, toRateConfig(d)); err != nil {
		return diag.FromErr(err)
	}
	return diags
}

func resourceRateConfigDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := getV1ClientWithResourceContext(m, tenantString)
	var diags diag.Diagnostics
	tenantUID, err := c.GetTenantUID()
	if err != nil {
		return diag.FromErr(err)
	}
	// The rate config can't be deleted, instead we are restoring the Palette defaults.
	if err := c.UpdateRateConfig(tenantUID, toRateConfigDefault()); err != nil {
		return diag.FromErr(err)
	}
	d.SetId("")
	return diags
}

func resourceRateConfigImport(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	c := getV1ClientWithResourceContext(m, tenantString)

	// Resolve org name → tenant UID if a name was provided instead of a UID
	resolvedTenantUID, err := resolveUidorNameToContextID(m, c, d.Id())
	if err != nil {
		return nil, err
	}

	actualTenantId, err := c.GetTenantUID()
	if err != nil {
		return nil, err
	}
	if resolvedTenantUID != actualTenantId {
		return nil, fmt.Errorf("invalid import: tenant %q does not match your authorized tenant UID %q", d.Id(), actualTenantId)
	}

	// Set the canonical ID so resourceRateConfigRead passes its singleton check
	d.SetId(rateConfigID)

	diags := resourceRateConfigRead(ctx, d, m)
	if diags.HasError() {
		return nil, fmt.Errorf("could not read rate config for import: %v", diags)
	}
	return []*schema.ResourceData{d}, nil
}
