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
									"uid": {
										Type:     schema.TypeString,
										Computed: true,
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
														return strings.TrimSpace(old) == strings.TrimSpace(new)
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
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	clusterUid := d.Get("cluster_uid").(string)

	cluster, err := c.GetCluster(clusterUid)
	if err != nil && cluster == nil {
		return diag.FromErr(errors.New(fmt.Sprintf("Cluster not found: %s", clusterUid)))
	}

	addonDeployment := toAddonDeployment(c, d)

	diagnostics, isError := waitForClusterCreation(ctx, d, clusterUid, diags, c, false)
	if isError {
		return diagnostics
	}

	if isProfileAttached(cluster, addonDeployment.Profiles[0].UID) {
		return diag.FromErr(errors.New(fmt.Sprintf("Cluster: %s: Profile is already attached: %s", cluster.Metadata.UID, addonDeployment.Profiles[0].UID)))
	}

	err = c.CreateAddonDeployment(cluster.Metadata.UID, addonDeployment)
	if err != nil {
		return diag.FromErr(err)
	}

	clusterProfile, err := c.GetClusterProfile(addonDeployment.Profiles[0].UID)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(getAddonDeploymentId(clusterUid, clusterProfile))

	diagnostics, isError = waitForAddonDeploymentCreation(ctx, d, cluster.Metadata.UID, addonDeployment.Profiles[0].UID, diags, c)
	if isError {
		return diagnostics
	}

	resourceAddonDeploymentRead(ctx, d, m)

	return diags
}

func getAddonDeploymentId(clusterUid string, clusterProfile *models.V1ClusterProfile) string {
	return clusterUid + "_" + clusterProfile.Metadata.UID
}

func getClusterUID(addonDeploymentId string) string {
	return strings.Split(addonDeploymentId, "_")[0]
}

func getClusterProfileUID(addonDeploymentId string) string {
	return strings.Split(addonDeploymentId, "_")[1]
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
	c := m.(*client.V1Client)

	var diags diag.Diagnostics

	uid := d.Get("cluster_uid").(string)
	cluster, err := c.GetCluster(uid)
	if err != nil {
		return diag.FromErr(err)
	}

	diagnostics, done := readAddonDeployment(c, d, cluster)
	if done {
		return diagnostics
	}

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

		newProfile, err := c.GetClusterProfile(addonDeployment.Profiles[0].UID)
		err = c.UpdateAddonDeployment(cluster, addonDeployment, newProfile)
		if err != nil {
			return diag.FromErr(err)
		}

		clusterProfile, err := c.GetClusterProfile(addonDeployment.Profiles[0].UID)
		if err != nil {
			return diag.FromErr(err)
		}
		d.SetId(getAddonDeploymentId(clusterUid, clusterProfile))
		diagnostics, isError := waitForAddonDeploymentUpdate(ctx, d, cluster.Metadata.UID, addonDeployment.Profiles[0].UID, diags, c)
		if isError {
			return diagnostics
		}

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
