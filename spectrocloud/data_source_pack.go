package spectrocloud

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/spectrocloud/terraform-provider-spectrocloud/pkg/client"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourcePack() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourcePackRead,

		Schema: map[string]*schema.Schema{
			"filters": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"id", "cloud", "name", "version", "registry_uid"},
			},
			"id": {
				Type:          schema.TypeString,
				Computed:      true,
				Optional:      true,
				ConflictsWith: []string{"filters", "cloud", "name", "version", "registry_uid"},
			},
			"name": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
			},
			"cloud": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Set:      schema.HashString,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"version": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
			},
			"registry_uid": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
			},
			"type": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
			},
			"values": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourcePackRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.V1Client)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	if v, ok := d.GetOk("type"); ok {
		if v.(string) == "manifest" {
			return diags
		}
		if v.(string) == "helm" {
			if regUID, ok := d.GetOk("registry_uid"); ok {
				_, err := c.GetHelmRegistry(regUID.(string))
				if err != nil {
					return diag.FromErr(err)
				}
				// we don't have provision to get all helm chart/packs from helm registry, hence skipping validation
				return diags
			}
		}
		if v.(string) == "oci" {
			if regUID, ok := d.GetOk("registry_uid"); ok {
				_, err := c.GetOciBasicRegistry(regUID.(string))
				if err != nil {
					return diag.FromErr(err)
				}
				// we don't have provision to get all helm chart/packs from oci basic type registry, hence skipping validation
				return diags
			}
		}
		if v.(string) == "ecr" {
			if regUID, ok := d.GetOk("registry_uid"); ok {
				_, err := c.GetOciRegistry(regUID.(string))
				if err != nil {
					return diag.FromErr(err)
				}
				// we don't have provision to get all helm chart/packs from ECR type registry, hence skipping validation
				return diags
			}
		}
	}
	// Validation is only for spectro packs
	filters := make([]string, 0)
	registryUID := ""
	if v, ok := d.GetOk("filters"); ok {
		filters = append(filters, v.(string))
	} else if v, ok := d.GetOk("id"); ok {
		filters = append(filters, fmt.Sprintf("metadata.uid=%s", v.(string)))
	} else {
		if v, ok := d.GetOk("name"); ok {

			/*
				Cluster profile now supports packs duplication, but pack name has to be unique and will be double dashed
				and first part would be any random name to make overall pack name unique and 2nd part is actual pack name.
				Thus, splitting pack name with '--' to get the correct pack name to find pack uuid
			*/
			if strings.Contains(v.(string), "--") {
				v = strings.Split(v.(string), "--")[1]
			}
			filters = append(filters, fmt.Sprintf("spec.name=%s", v.(string)))
		}
		if v, ok := d.GetOk("version"); ok {
			filters = append(filters, fmt.Sprintf("spec.version=%s", v.(string)))
		}
		if v, ok := d.GetOk("registry_uid"); ok {
			registryUID = v.(string)
		}
		if v, ok := d.GetOk("cloud"); ok {
			clouds := expandStringList(v.(*schema.Set).List())
			if !stringContains(clouds, "all") {
				clouds = append(clouds, "all")
			}
			filters = append(filters, fmt.Sprintf("spec.cloudTypes_in_%s", strings.Join(clouds, ",")))
		}
	}

	packs, err := c.GetPacks(filters, registryUID)
	if err != nil {
		return diag.FromErr(err)
	}

	packName := "unknown"
	if v, ok := d.GetOk("name"); ok {
		packName = v.(string)
	}

	if len(packs) == 0 {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("%s: no matching packs", packName),
			Detail:   "No packs matching criteria found",
		})
		return diags
	} else if len(packs) > 1 {
		packs_map := make(map[string]string, 0)
		for _, pack := range packs {
			packs_map[pack.Spec.RegistryUID] = pack.Spec.Name
		}
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("%s: Multiple packs returned", packName),
			Detail:   fmt.Sprintf("Found %d matching packs. Restrict packs criteria to a single match. %s", len(packs), packs_map),
		})
		return diags
	}

	pack := packs[0]

	clouds := make([]string, 0)
	for _, cloudType := range pack.Spec.CloudTypes {
		clouds = append(clouds, string(cloudType))
	}

	d.SetId(pack.Metadata.UID)
	err = d.Set("name", pack.Spec.Name)
	if err != nil {
		return diag.FromErr(err)
	}
	err = d.Set("cloud", clouds)
	if err != nil {
		return diag.FromErr(err)
	}
	err = d.Set("version", pack.Spec.Version)
	if err != nil {
		return diag.FromErr(err)
	}
	err = d.Set("registry_uid", pack.Spec.RegistryUID)
	if err != nil {
		return diag.FromErr(err)
	}
	err = d.Set("values", pack.Spec.Values)
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}
