package spectrocloud

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func dataSourceRegistry() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceRegistryRead,

		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"type": {
				Type:         schema.TypeString,
				Default:      "",
				Optional:     true,
				Description:  "The type of the registry. Possible values are 'oci', 'helm', or 'spectro'. If not provided, the registry type will be inferred from the registry name.",
				ValidateFunc: validation.StringInSlice([]string{"", "oci", "helm", "spectro"}, false),
			},
		},
	}
}

func dataSourceRegistryRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := getV1ClientWithResourceContext(m, "")
	var diags diag.Diagnostics

	name, ok := d.GetOk("name")
	if !ok {
		return diags
	}

	registryType := d.Get("type").(string)

	var uid, registryName string
	var err error

	switch registryType {
	case "oci":
		registry, e := c.GetOciRegistryByName(name.(string))
		if e != nil {
			return diag.FromErr(e)
		}
		uid = registry.Metadata.UID
		registryName = registry.Metadata.Name
	case "helm":
		registry, e := c.GetHelmRegistryByName(name.(string))
		if e != nil {
			return diag.FromErr(e)
		}
		uid = registry.Metadata.UID
		registryName = registry.Metadata.Name
	default: // "" or "spectro"
		registry, e := c.GetPackRegistryCommonByName(name.(string))
		if e != nil {
			return handleReadError(d, e, diags)
		}
		uid = registry.UID
		registryName = registry.Name
	}

	d.SetId(uid)
	if err = d.Set("name", registryName); err != nil {
		return diag.FromErr(err)
	}

	return diags
}
