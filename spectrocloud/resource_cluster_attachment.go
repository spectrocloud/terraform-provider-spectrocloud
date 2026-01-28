package spectrocloud

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/spectrocloud/palette-sdk-go/client"

	"github.com/spectrocloud/terraform-provider-spectrocloud/spectrocloud/schemas"
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
				Type:        schema.TypeString,
				Required:    true,
				Description: "The UID of the cluster to attach the addon profile to.",
			},
			"context": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"project", "tenant"}, false),
				Default:      "project",
				Description: "Specifies cluster context where addon profile is attached. " +
					"Allowed values are `project` or `tenant`. Defaults to `project`. " + PROJECT_NAME_NUANCE,
			},
			"cluster_profile": schemas.ClusterProfileSchema(),
			"apply_setting": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "DownloadAndInstall",
				ValidateFunc: validation.StringInSlice([]string{"DownloadAndInstall", "DownloadAndInstallLater"}, false),
				Description: "The setting to apply the cluster profile. `DownloadAndInstall` will download and install packs in one action. " +
					"`DownloadAndInstallLater` will only download artifact and postpone install for later. " +
					"Default value is `DownloadAndInstall`.",
			},
		},
	}
}

// validateSingleClusterProfile validates that exactly one cluster_profile is specified
func validateSingleClusterProfile(d *schema.ResourceData) error {
	profiles := d.Get("cluster_profile").([]interface{})
	if len(profiles) == 0 {
		return errors.New("exactly one cluster_profile is required, but none was specified")
	}
	if len(profiles) > 1 {
		return fmt.Errorf("exactly one cluster_profile is allowed per addon deployment, but %d were specified. Use separate spectrocloud_addon_deployment resources for each profile", len(profiles))
	}
	return nil
}

func resourceAddonDeploymentCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Validate exactly one cluster_profile is specified
	if err := validateSingleClusterProfile(d); err != nil {
		return diag.FromErr(err)
	}

	resourceContext := d.Get("context").(string)
	c := getV1ClientWithResourceContext(m, resourceContext)

	var diags diag.Diagnostics

	clusterUid := d.Get("cluster_uid").(string)

	cluster, err := c.GetCluster(clusterUid)
	if err != nil && cluster == nil {
		return diag.FromErr(fmt.Errorf("cluster not found: %s", clusterUid))
	}

	addonDeployment, err := toAddonDeployment(c, d)
	if err != nil {
		return diag.FromErr(err)
	}

	diagnostics, isError := waitForClusterCreation(ctx, d, clusterUid, diags, c, false)
	if isError {
		return diagnostics
	}
	// Clear the ID to skip resource tainted
	d.SetId("")

	if isProfileAttached(cluster, addonDeployment.Profiles[0].UID) {
		return updateAddonDeployment(ctx, d, m, c, cluster, clusterUid, diags)
	}

	err = c.CreateAddonDeployment(cluster, addonDeployment)
	if err != nil {
		// Clear the ID to skip resource tainted
		d.SetId("")
		return diag.FromErr(err)
	}

	clusterProfile, err := c.GetClusterProfile(addonDeployment.Profiles[0].UID)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(getAddonDeploymentId(clusterUid, clusterProfile))

	diagnostics, isError = waitForAddonDeploymentCreation(ctx, d, *cluster, addonDeployment.Profiles[0].UID, diags, c)
	if isError {
		return diagnostics
	}

	return resourceAddonDeploymentRead(ctx, d, m)
}

func getAddonDeploymentId(clusterUid string, clusterProfile *models.V1ClusterProfile) string {
	return clusterUid + "_" + clusterProfile.Metadata.UID
}

func getClusterUID(addonDeploymentId string) string {
	return strings.Split(addonDeploymentId, "_")[0]
}

func getClusterProfileUID(addonDeploymentId string) (string, error) {
	sp := strings.Split(addonDeploymentId, "_")
	if len(sp) < 2 {
		return "", errors.New("invalid addon deployment ID format")
	}
	return sp[1], nil
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
	resourceContext := d.Get("context").(string)
	c := getV1ClientWithResourceContext(m, resourceContext)

	var diags diag.Diagnostics

	clusterUid := d.Get("cluster_uid").(string)
	if !strings.Contains(d.Id(), clusterUid) {
		d.SetId("")
		return diags
	}

	cluster, err := c.GetCluster(clusterUid)
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
	// Validate exactly one cluster_profile is specified
	if err := validateSingleClusterProfile(d); err != nil {
		return diag.FromErr(err)
	}

	var diags diag.Diagnostics

	if d.HasChanges("cluster_uid", "cluster_profile") {
		resourceContext := d.Get("context").(string)
		c := getV1ClientWithResourceContext(m, resourceContext)

		clusterUid := d.Get("cluster_uid").(string)

		cluster, err := c.GetCluster(clusterUid)
		if err != nil && cluster == nil {
			return diag.FromErr(fmt.Errorf("cluster not found: %s", clusterUid))
		}

		return updateAddonDeployment(ctx, d, m, c, cluster, clusterUid, diags)
	}

	return diags
}

func updateAddonDeployment(ctx context.Context, d *schema.ResourceData, m interface{}, c *client.V1Client, cluster *models.V1SpectroCluster, clusterUid string, diags diag.Diagnostics) diag.Diagnostics {
	addonDeployment, err := toAddonDeployment(c, d)
	if err != nil {
		return diag.FromErr(err)
	}
	if len(addonDeployment.Profiles) == 0 {
		return diag.FromErr(errors.New("cannot convert addon deployment: zero profiles found"))
	}

	newProfile, err := c.GetClusterProfile(addonDeployment.Profiles[0].UID)
	if err != nil {
		return diag.FromErr(err)
	}
	err = c.UpdateAddonDeployment(cluster, addonDeployment, newProfile)
	if err != nil {
		return diag.FromErr(err)
	}

	clusterProfile, err := c.GetClusterProfile(addonDeployment.Profiles[0].UID)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(getAddonDeploymentId(clusterUid, clusterProfile))
	diagnostics, isError := waitForAddonDeploymentUpdate(ctx, d, *cluster, addonDeployment.Profiles[0].UID, diags, c)
	if isError {
		return diagnostics
	}

	return resourceAddonDeploymentRead(ctx, d, m)
}

func toAddonDeployment(c *client.V1Client, d *schema.ResourceData) (*models.V1SpectroClusterProfiles, error) {
	profiles, err := toAddonDeplProfiles(c, d)
	if err != nil {
		return nil, err
	}
	settings, err := toSpcApplySettings(d)
	if err != nil {
		return nil, err
	}
	return &models.V1SpectroClusterProfiles{
		Profiles:         profiles,
		SpcApplySettings: settings,
	}, nil
}
