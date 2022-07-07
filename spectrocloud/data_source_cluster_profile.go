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
			"version": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
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

	version := "1.0.0" //default
	if ver, ok_version := d.GetOk("version"); ok_version {
		version = ver.(string)
	}

	profile, err := c.GetClusterProfile(getProfileUID(profiles, d, version))
	if err != nil {
		return diag.FromErr(err)
	}

	if profile == nil || profile.Metadata == nil {
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
		packManifests, d2, done2 := getPacksContent(profile, c, d)
		if done2 {
			return d2
		}

		diagPacks, diagnostics, done := GetDiagPacks(d, err)
		if done {
			return diagnostics
		}
		packs, err := flattenPacks(c, diagPacks, profile.Spec.Published.Packs, packManifests)
		if err != nil {
			return diag.FromErr(err)
		}
		if err := d.Set("pack", packs); err != nil {
			return diag.FromErr(err)
		}
	}

	return diags
}

func GetDiagPacks(d *schema.ResourceData, err error) ([]*models.V1PackManifestEntity, diag.Diagnostics, bool) {
	diagPacks := make([]*models.V1PackManifestEntity, 0)
	for _, pack := range d.Get("pack").([]interface{}) {
		if p, e := toClusterProfilePackCreate(pack); e != nil {
			return nil, diag.FromErr(err), true
		} else {
			diagPacks = append(diagPacks, p)
		}
	}
	return diagPacks, nil, false
}

func getProfileUID(profiles []*models.V1ClusterProfile, d *schema.ResourceData, version string) string {
	for _, p := range profiles {
		if v, ok := d.GetOk("id"); ok && v.(string) == p.Metadata.UID {
			return p.Metadata.UID
		} else if v, ok := d.GetOk("name"); ok && v.(string) == p.Metadata.Name {
			if p.Spec.Version == version || (p.Spec.Version == "" && version == "1.0.0") {
				return p.Metadata.UID
			}
		}
	}
	return ""
}
