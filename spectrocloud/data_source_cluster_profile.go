package spectrocloud

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/spectrocloud/palette-sdk-go/client"
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
				Description:  "The unique ID of the cluster profile. Either `id` or `name` must be provided, but not both.",
			},
			"name": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ExactlyOneOf: []string{"id", "name"},
				Description:  "The name of the cluster profile. Either `id` or `name` must be provided, but not both.",
			},
			"version": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The version of the cluster profile.",
			},
			"context": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "project",
				ValidateFunc: validation.StringInSlice([]string{"", "project", "tenant", "system"}, false),
				Description: "Cluster profile context. Allowed values are `project` or `tenant`. " +
					"Defaults to `project`." + PROJECT_NAME_NUANCE,
			},
			"pack": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "spectro",
							Description: "The type of pack. Defaults to `spectro`.",
						},
						"registry_uid": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "The UID of the registry associated with the pack.",
						},
						"uid": {
							Type:        schema.TypeString,
							Computed:    true,
							Optional:    true,
							Description: "The unique identifier for the pack.",
						},
						"name": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The name of the pack.",
						},
						"manifest": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"uid": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "The unique ID of the manifest.",
									},
									"name": {
										Type:        schema.TypeString,
										Required:    true,
										Description: "The name of the manifest.",
									},
									"content": {
										Type:     schema.TypeString,
										Required: true,
										DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
											return strings.TrimSpace(old) == strings.TrimSpace(new)
										},
										Description: "The content of the manifest.",
									},
								},
							},
						},
						"tag": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "The tag associated with the pack.",
						},
						"values": {
							Type:     schema.TypeString,
							Optional: true,
							DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
								// UI strips the trailing newline on save
								return strings.TrimSpace(old) == strings.TrimSpace(new)
							},
							Description: "The YAML values associated with the pack.",
						},
					},
				},
			},
		},
	}
}

func dataSourceClusterProfileRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	ProjectContext := "project"
	if Pcontext, ok_context := d.GetOk("context"); ok_context {
		ProjectContext = Pcontext.(string)
	}
	c := getV1ClientWithResourceContext(m, ProjectContext)

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

	profile, err := getProfile(profiles, d, version, ProjectContext, c)
	if err != nil {
		return diag.FromErr(err)
	}

	if profile == nil || profile.Metadata == nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to find cluster profile",
			Detail:   fmt.Sprintf("Unable to find the specified cluster profile name: %s, version: %s", d.Get("name").(string), version),
		})
		return diags
	}

	d.SetId(profile.Metadata.UID)
	if err := d.Set("name", profile.Metadata.Name); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("version", profile.Spec.Version); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("context", profile.Metadata.Annotations["scope"]); err != nil {
		return diag.FromErr(err)
	}

	if profile.Spec.Published != nil && len(profile.Spec.Published.Packs) > 0 {
		packManifests, d2, done2 := getPacksContent(profile.Spec.Published.Packs, c, d)
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

func getProfile(profiles []*models.V1ClusterProfileMetadata, d *schema.ResourceData, version, ProfileContext string, c *client.V1Client) (*models.V1ClusterProfile, error) {

	for _, p := range profiles {
		if v, ok := d.GetOk("id"); ok && v.(string) == p.Metadata.UID {
			fullProfile, err := c.GetClusterProfile(p.Metadata.UID)
			if err != nil {
				return nil, err
			}
			return fullProfile, nil
		} else if v, ok := d.GetOk("name"); ok && v.(string) == p.Metadata.Name {
			if p.Spec.Version == version || (p.Spec.Version == "" && version == "1.0.0") {
				fullProfile, err := c.GetClusterProfile(p.Metadata.UID)
				if err != nil {
					return nil, err
				}
				if ProfileContext == fullProfile.Metadata.Annotations["scope"] {
					return fullProfile, nil
				}
			}
		}
	}
	return nil, fmt.Errorf("cluster profile not found: name: %s, version: %s, context: %s", d.Get("name").(string), version, ProfileContext)
}
