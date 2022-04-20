package spectrocloud

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/spectrocloud/hapi/models"
	"github.com/spectrocloud/terraform-provider-spectrocloud/pkg/client"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceClusterProfile() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceClusterProfileRead,

		Schema: map[string]*schema.Schema{
			"id": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ExactlyOneOf: []string{"id", "name"},
			},
			"name": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ExactlyOneOf: []string{"id", "name"},
			},
			"pack": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  "spectro",
						},
						"registry_uid": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"uid": {
							Type:     schema.TypeString,
							Computed: true,
							Optional: true,
						},
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"manifest": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"uid": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"name": {
										Type:     schema.TypeString,
										Required: true,
									},
									"content": {
										Type:     schema.TypeString,
										Required: true,
										DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
											if strings.TrimSpace(old) == strings.TrimSpace(new) {
												return true
											}
											return false
										},
									},
								},
							},
						},
						"tag": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"values": {
							Type:     schema.TypeString,
							Optional: true,
							DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
								// UI strips the trailing newline on save
								if strings.TrimSpace(old) == strings.TrimSpace(new) {
									return true
								}
								return false
							},
						},
					},
				},
			},
		},
	}
}

func dataSourceClusterProfileRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.V1Client)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics
	profiles, err := c.GetClusterProfiles()
	if err != nil {
		return diag.FromErr(err)
	}

	var profile *models.V1ClusterProfile
	for _, p := range profiles {

		if v, ok := d.GetOk("id"); ok && v.(string) == p.Metadata.UID {
			profile = p
			break
		} else if v, ok := d.GetOk("name"); ok && v.(string) == p.Metadata.Name {
			profile = p
			break
		}
	}

	if profile == nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to find cluster profile",
			Detail:   "Unable to find the specified cluster profile",
		})
		return diags
	}

	d.SetId(profile.Metadata.UID)
	d.Set("name", profile.Metadata.Name)
	if profile.Spec.Published != nil && len(profile.Spec.Published.Packs) > 0 {
		packManifests := make(map[string][]string)
		for _, p := range profile.Spec.Published.Packs {
			if len(p.Manifests) > 0 {
				content, err := c.GetClusterProfileManifestPack(d.Id(), *p.Name)
				if err != nil {
					return diag.FromErr(err)
				}

				if len(content) > 0 {
					c := make([]string, len(content))
					for i, co := range content {
						c[i] = co.Spec.Published.Content
					}
					packManifests[p.PackUID] = c
				}
			}
		}

		packs, err := flattenPacks(c, profile.Spec.Published.Packs, packManifests)
		if err != nil {
			return diag.FromErr(err)
		}
		if err := d.Set("pack", packs); err != nil {
			return diag.FromErr(err)
		}
	}

	return diags
}
