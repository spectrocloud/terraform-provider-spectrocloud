package spectrocloud

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/hapi/models"
	"github.com/spectrocloud/terraform-provider-spectrocloud/pkg/client"
)

func resourceAddonDeployment() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceAddonDeploymentCreate,
		ReadContext:   resourceAddonDeploymentRead,
		UpdateContext: resourceAddonDeploymentUpdate,
		DeleteContext: resourceAddonDeploymentDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Minute),
			Update: schema.DefaultTimeout(60 * time.Minute),
			Delete: schema.DefaultTimeout(60 * time.Minute),
		},

		SchemaVersion: 2,
		Schema: map[string]*schema.Schema{
			"cluster_uid": {
				Type:     schema.TypeString,
				Required: true,
			},
			"cluster_profile": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Required: true,
						},
						"pack": {
							Type:     schema.TypeList,
							Optional: true,
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
									"name": {
										Type:     schema.TypeString,
										Required: true,
									},
									"tag": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"values": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"manifest": {
										Type:     schema.TypeList,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"name": {
													Type:     schema.TypeString,
													Required: true,
												},
												"content": {
													Type:     schema.TypeString,
													Required: true,
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
							},
						},
					},
				},
			},
			"apply_setting": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func resourceAddonDeploymentCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.V1Client)

	clusterUid := d.Get("cluster_uid").(string)

	cluster, err := c.GetCluster(clusterUid)
	if err != nil && cluster == nil {
		return diag.FromErr(errors.New(fmt.Sprintf("Cluster not found: %s", clusterUid)))
	}

	addonDeployment := toAddonDeployment(c, d)

	if d.Id() != "" || isProfileAttached(cluster, addonDeployment.Profiles[0].UID) {
		return diag.FromErr(errors.New(fmt.Sprintf("Profile is already attached: %s", addonDeployment.Profiles[0].UID)))
	}

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	err = c.CreateOrUpdateAddonDeployment(cluster.Metadata.UID, addonDeployment)
	if err != nil {
		return diag.FromErr(err)
	}

	// TODO: implement wait for.
	d.SetId(clusterUid + addonDeployment.Profiles[0].UID)
	/*diagnostics, isError := waitForAddonDeploymentCreation(ctx, d, cluster, diags, c)
	if isError {
		return diagnostics
	}*/

	resourceAddonDeploymentRead(ctx, d, m)

	return diags
}

func isProfileAttached(cluster *models.V1SpectroCluster, uid string) bool {
	for _, profile := range cluster.Spec.ClusterProfileTemplates {
		if profile.UID == uid {
			return true
		}
	}

	return false
}

//goland:noinspection GoUnhandledErrorResult
func resourceAddonDeploymentRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	/*c := m.(*client.V1Client)

	var diags diag.Diagnostics

	uid := d.Get("cluster_uid").(string)
	cluster, err := c.GetCluster(uid)
	if err != nil {
		return diag.FromErr(err)
	}

	diagnostics, done := readCommonFields(c, d, cluster)
	if done {
		return diagnostics
	}*/
	var diags diag.Diagnostics

	return diags
}

func resourceAddonDeploymentUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	if d.HasChanges("cluster_uid", "cluster_profile") {
		c := m.(*client.V1Client)

		clusterUid := d.Get("cluster_uid").(string)

		cluster, err := c.GetCluster(clusterUid)
		if err != nil && cluster == nil {
			return diag.FromErr(errors.New(fmt.Sprintf("Cluster not found: %s", clusterUid)))
		}

		addonDeployment := toAddonDeployment(c, d)

		err = c.CreateOrUpdateAddonDeployment(cluster.Metadata.UID, addonDeployment)
		if err != nil {
			return diag.FromErr(err)
		}

		// TODO: implement wait for.
		d.SetId(clusterUid + addonDeployment.Profiles[0].UID)
		/*diagnostics, isError := waitForAddonDeploymentCreation(ctx, d, cluster, diags, c)
		if isError {
			return diagnostics
		}*/

		resourceAddonDeploymentRead(ctx, d, m)

		return diags
	}

	return diags
}

func toAddonDeployment(c *client.V1Client, d *schema.ResourceData) *models.V1SpectroClusterProfiles {
	return &models.V1SpectroClusterProfiles{
		Profiles:         toProfiles(c, d),
		SpcApplySettings: toSpcApplySettings(d),
	}
}
