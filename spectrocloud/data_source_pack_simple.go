package spectrocloud

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/spectrocloud/terraform-provider-spectrocloud/pkg/client"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourcePackSimple() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourcePackReadSimple,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the pack.",
			},
			"version": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The version of the pack.",
			},
			"context": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "project",
				ValidateFunc: validation.StringInSlice([]string{"system", "project", "tenant"}, false),
			},
			"registry_uid": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The unique identifier of the registry the pack belongs to.",
			},
			"type": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"helm", "manifest", "container", "operator-instance"}, false),
				Description:  "The type of Pack. Allowed values are `helm`, `manifest`, `container` or `operator-instance`.",
			},
			"values": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "This is a stringified YAML object containing the pack configuration details. ",
			},
		},
	}
}

func dataSourcePackReadSimple(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.V1Client)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	packContext := d.Get("context").(string)
	packName := ""
	registryUID := ""
	if v, ok := d.GetOk("type"); ok {
		if v.(string) == "manifest" {
			return diags
		} else {
			if regUID, ok := d.GetOk("registry_uid"); ok {
				registryUID = regUID.(string)
				registry, err := c.GetHelmRegistry(regUID.(string))
				if err != nil {
					return diag.FromErr(err)
				}
				if registry.Spec.IsPrivate {
					return diags
				}
			} else {
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "No registry uid provided.",
					Detail:   fmt.Sprintf("Registry uid is required for pack name:%s, type:%s ", d.Get("name").(string), d.Get("type").(string)),
				})
				return diags
			}
		}
	}

	if v, ok := d.GetOk("name"); ok {

		/*
			Cluster profile now supports packs duplication, but pack name has to be unique and will be double dashed
			and first part would be any random name to make overall pack name unique and 2nd part is actual pack name.
			Thus, splitting pack name with '--' to get the correct pack name to find pack uuid
		*/
		if strings.Contains(v.(string), "--") {
			v = strings.Split(v.(string), "--")[1]
		}
		packName = v.(string)
	}

	pack, err := c.GetPacksByNameAndRegistry(packName, registryUID, packContext)
	if err != nil {
		return diag.FromErr(err)
	}

	version := "1.0.0"
	if v, ok := d.GetOk("version"); ok {
		version = v.(string)
	}
	for _, tag := range pack.Tags {
		if tag.Version == version {
			d.SetId(tag.PackUID)
			err = d.Set("name", pack.Name)
			if err != nil {
				return diag.FromErr(err)
			}
			err = d.Set("version", tag.Version)
			if err != nil {
				return diag.FromErr(err)
			}
			err = d.Set("registry_uid", pack.RegistryUID)
			if err != nil {
				return diag.FromErr(err)
			}
			for _, value := range pack.PackValues {
				if value.PackUID == tag.PackUID {
					err = d.Set("values", value.Values)
					if err != nil {
						return diag.FromErr(err)
					}
					return diags
				}
			}
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "No values for pack found.",
				Detail:   fmt.Sprintf("Values not found for pack name:%s, version:%s ", d.Get("name").(string), d.Get("version").(string)),
			})
			return diags
		}
	}

	diags = append(diags, diag.Diagnostic{
		Severity: diag.Error,
		Summary:  "No version for pack found.",
		Detail:   fmt.Sprintf("Version not found for pack name:%s, version:%s ", d.Get("name").(string), d.Get("version").(string)),
	})
	return diags
}
