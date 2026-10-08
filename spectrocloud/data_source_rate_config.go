package spectrocloud

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// publicCloudRateDataSchema returns the read-only schema for a public cloud rate config.
func publicCloudRateDataSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"compute_optimized": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "Rate proportions applied to general purpose, compute optimized and GPU optimized instance types.",
				Elem:        cloudInstanceRateDataSchema(),
			},
			"memory_optimized": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "Rate proportions applied to memory optimized and storage optimized instance types.",
				Elem:        cloudInstanceRateDataSchema(),
			},
		},
	}
}

func cloudInstanceRateDataSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"compute_rate_proportion": {
				Type:        schema.TypeFloat,
				Computed:    true,
				Description: "Percentage of the instance price attributed to compute.",
			},
			"memory_rate_proportion": {
				Type:        schema.TypeFloat,
				Computed:    true,
				Description: "Percentage of the instance price attributed to memory.",
			},
		},
	}
}

// privateCloudRateDataSchema returns the read-only schema for a private cloud rate config.
func privateCloudRateDataSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"cpu_unit_price_per_hour": {
				Type:        schema.TypeFloat,
				Computed:    true,
				Description: "Price in US dollars charged for one CPU core per hour.",
			},
			"gpu_unit_price_per_hour": {
				Type:        schema.TypeFloat,
				Computed:    true,
				Description: "Price in US dollars charged for one GPU per hour.",
			},
			"memory_unit_price_gib_per_hour": {
				Type:        schema.TypeFloat,
				Computed:    true,
				Description: "Price in US dollars charged for one GiB of memory per hour.",
			},
			"storage_unit_price_gib_per_hour": {
				Type:        schema.TypeFloat,
				Computed:    true,
				Description: "Price in US dollars charged for one GiB of storage per hour.",
			},
		},
	}
}

func publicCloudRateDataAttribute(cloudName string) *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeList,
		Computed:    true,
		Description: fmt.Sprintf("Rate proportions used to estimate the cost of %s instances.", cloudName),
		Elem:        publicCloudRateDataSchema(),
	}
}

func privateCloudRateDataAttribute(cloudName string) *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeList,
		Computed:    true,
		Description: fmt.Sprintf("Unit prices used to estimate the cost of %s resources.", cloudName),
		Elem:        privateCloudRateDataSchema(),
	}
}

func dataSourceRateConfig() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceRateConfigRead,
		Description: "Provides the tenant-level cloud rate config, the unit prices Palette uses to estimate cluster cloud cost and usage cost.",
		Schema: map[string]*schema.Schema{
			"aws":               publicCloudRateDataAttribute("AWS"),
			"azure":             publicCloudRateDataAttribute("Azure"),
			"gcp":               publicCloudRateDataAttribute("GCP"),
			"vsphere":           privateCloudRateDataAttribute("VMware vSphere"),
			"maas":              privateCloudRateDataAttribute("MAAS"),
			"edge":              privateCloudRateDataAttribute("Edge"),
			"edge_native":       privateCloudRateDataAttribute("Edge Native"),
			"generic":           privateCloudRateDataAttribute("generic cloud"),
			"apache_cloudstack": privateCloudRateDataAttribute("Apache CloudStack"),
			"custom": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "Unit prices applied to the custom cloud types registered in the tenant. One block per custom cloud.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"cloud_type": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Name of the custom cloud type the rates apply to.",
						},
						"rate_config": {
							Type:        schema.TypeList,
							Computed:    true,
							Description: "Unit prices used to estimate the cost of resources running on this custom cloud.",
							Elem:        privateCloudRateDataSchema(),
						},
					},
				},
			},
		},
	}
}

func dataSourceRateConfigRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := getV1ClientWithResourceContext(m, tenantString)
	var diags diag.Diagnostics
	tenantUID, err := c.GetTenantUID()
	if err != nil {
		return diag.FromErr(err)
	}
	rateConfig, err := c.GetRateConfig(tenantUID)
	if err != nil {
		return diag.FromErr(err)
	}
	if err := flattenRateConfig(rateConfig, d); err != nil {
		return diag.FromErr(err)
	}
	d.SetId(tenantUID)
	return diags
}
