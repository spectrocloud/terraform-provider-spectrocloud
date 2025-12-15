package spectrocloud

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sort"
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
				Description: "The UID of the cluster to attach the addon profile(s) to.",
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

func resourceAddonDeploymentCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
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

	if len(addonDeployment.Profiles) == 0 {
		return diag.FromErr(errors.New("at least one cluster_profile is required"))
	}

	diagnostics, isError := waitForClusterCreation(ctx, d, clusterUid, diags, c, false)
	if isError {
		return diagnostics
	}
	// Clear the ID to skip resource tainted
	d.SetId("")

	// Collect profile UIDs for the resource ID
	profileUIDs := make([]string, 0, len(addonDeployment.Profiles))

	// Process each profile - check if already attached, if so update, otherwise create
	for _, profile := range addonDeployment.Profiles {
		if isProfileAttached(cluster, profile.UID) {
			// Profile already attached, update it
			log.Printf("Profile %s already attached to cluster %s, updating", profile.UID, clusterUid)
			singleProfileBody := &models.V1SpectroClusterProfiles{
				Profiles:         []*models.V1SpectroClusterProfileEntity{profile},
				SpcApplySettings: addonDeployment.SpcApplySettings,
			}
			clusterProfile, err := c.GetClusterProfile(profile.UID)
			if err != nil {
				d.SetId("")
				return diag.FromErr(err)
			}
			err = c.UpdateAddonDeployment(cluster, singleProfileBody, clusterProfile)
			if err != nil {
				d.SetId("")
				return diag.FromErr(err)
			}
		} else {
			// Create new profile attachment
			log.Printf("Attaching profile %s to cluster %s", profile.UID, clusterUid)
			singleProfileBody := &models.V1SpectroClusterProfiles{
				Profiles:         []*models.V1SpectroClusterProfileEntity{profile},
				SpcApplySettings: addonDeployment.SpcApplySettings,
			}
			err = c.CreateAddonDeployment(cluster, singleProfileBody)
			if err != nil {
				d.SetId("")
				return diag.FromErr(err)
			}
		}
		profileUIDs = append(profileUIDs, profile.UID)
	}

	// Set the resource ID with cluster UID and all profile UIDs
	d.SetId(buildAddonDeploymentId(clusterUid, profileUIDs))

	// Wait for all profiles to be deployed
	for _, profile := range addonDeployment.Profiles {
		diagnostics, isError = waitForAddonDeploymentCreation(ctx, d, *cluster, profile.UID, diags, c)
		if isError {
			return diagnostics
		}
	}

	return resourceAddonDeploymentRead(ctx, d, m)
}

// buildAddonDeploymentId creates a resource ID from cluster UID and profile UIDs
// Format: {clusterUID}_{profileUID1}_{profileUID2}...
func buildAddonDeploymentId(clusterUid string, profileUIDs []string) string {
	// Sort profile UIDs for consistent ID generation
	sortedUIDs := make([]string, len(profileUIDs))
	copy(sortedUIDs, profileUIDs)
	sort.Strings(sortedUIDs)

	parts := []string{clusterUid}
	parts = append(parts, sortedUIDs...)
	return strings.Join(parts, "_")
}

// getAddonDeploymentId creates a resource ID from cluster UID and a single profile (legacy compatibility)
// Format: {clusterUID}_{profileUID}
func getAddonDeploymentId(clusterUid string, clusterProfile *models.V1ClusterProfile) string {
	return clusterUid + "_" + clusterProfile.Metadata.UID
}

// parseAddonDeploymentId extracts cluster UID and profile UIDs from resource ID
func parseAddonDeploymentId(id string) (clusterUID string, profileUIDs []string, err error) {
	parts := strings.Split(id, "_")
	if len(parts) < 2 {
		return "", nil, fmt.Errorf("invalid addon deployment ID format: %s", id)
	}
	return parts[0], parts[1:], nil
}

// getClusterUID extracts cluster UID from resource ID (legacy support)
func getClusterUID(addonDeploymentId string) string {
	return strings.Split(addonDeploymentId, "_")[0]
}

// getClusterProfileUID extracts the first profile UID from resource ID (legacy support)
func getClusterProfileUID(addonDeploymentId string) (string, error) {
	sp := strings.Split(addonDeploymentId, "_")
	if len(sp) < 2 {
		return "", errors.New("invalid addon deployment ID format")
	}
	return sp[1], nil
}

// getClusterProfileUIDs extracts all profile UIDs from resource ID
func getClusterProfileUIDs(addonDeploymentId string) ([]string, error) {
	_, profileUIDs, err := parseAddonDeploymentId(addonDeploymentId)
	return profileUIDs, err
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

	diagnostics, done := readAddonDeploymentMultiple(c, d, cluster)
	if done {
		return diagnostics
	}

	return diags
}

func resourceAddonDeploymentUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	if d.HasChanges("cluster_uid", "cluster_profile") {
		resourceContext := d.Get("context").(string)
		c := getV1ClientWithResourceContext(m, resourceContext)

		clusterUid := d.Get("cluster_uid").(string)

		return updateAddonDeploymentMultiple(ctx, d, m, c, clusterUid, diags)
	}

	return diags
}

func updateAddonDeploymentMultiple(ctx context.Context, d *schema.ResourceData, m interface{}, c *client.V1Client, clusterUid string, diags diag.Diagnostics) diag.Diagnostics {
	// Get old and new profile configurations
	oldProfilesRaw, newProfilesRaw := d.GetChange("cluster_profile")
	oldProfiles := oldProfilesRaw.([]interface{})
	newProfiles := newProfilesRaw.([]interface{})

	// Build maps for comparison
	oldProfileMap := make(map[string]map[string]interface{})
	for _, p := range oldProfiles {
		if p == nil {
			continue
		}
		profile := p.(map[string]interface{})
		if id, ok := profile["id"].(string); ok && id != "" {
			oldProfileMap[id] = profile
		}
	}

	newProfileMap := make(map[string]map[string]interface{})
	for _, p := range newProfiles {
		if p == nil {
			continue
		}
		profile := p.(map[string]interface{})
		if id, ok := profile["id"].(string); ok && id != "" {
			newProfileMap[id] = profile
		}
	}

	// Find profiles to delete (in old but not in new)
	profilesToDelete := make([]string, 0)
	for oldID := range oldProfileMap {
		if _, exists := newProfileMap[oldID]; !exists {
			profilesToDelete = append(profilesToDelete, oldID)
		}
	}

	// Delete removed profiles
	if len(profilesToDelete) > 0 {
		log.Printf("Deleting %d profiles from cluster %s: %v", len(profilesToDelete), clusterUid, profilesToDelete)
		deleteBody := &models.V1SpectroClusterProfilesDeleteEntity{
			ProfileUids: profilesToDelete,
		}
		if err := c.DeleteAddonDeployment(clusterUid, deleteBody); err != nil {
			return diag.FromErr(fmt.Errorf("failed to delete profiles: %w", err))
		}
	}

	// Get the addon deployment for new/updated profiles
	addonDeployment, err := toAddonDeployment(c, d)
	if err != nil {
		return diag.FromErr(err)
	}

	// Get cluster state for profile operations
	cluster, err := c.GetCluster(clusterUid)
	if err != nil {
		return diag.FromErr(err)
	}
	if cluster == nil {
		return diag.FromErr(fmt.Errorf("cluster not found: %s", clusterUid))
	}

	// Process each profile - add new or update existing
	newProfileUIDs := make([]string, 0, len(addonDeployment.Profiles))
	for _, profile := range addonDeployment.Profiles {
		newProfileUIDs = append(newProfileUIDs, profile.UID)

		// Get the cluster profile details
		clusterProfile, err := c.GetClusterProfile(profile.UID)
		if err != nil {
			return diag.FromErr(err)
		}

		if isProfileAttached(cluster, profile.UID) {
			// Profile exists, update it
			log.Printf("Updating profile %s on cluster %s", profile.UID, clusterUid)
			singleProfileBody := &models.V1SpectroClusterProfiles{
				Profiles:         []*models.V1SpectroClusterProfileEntity{profile},
				SpcApplySettings: addonDeployment.SpcApplySettings,
			}
			err = c.UpdateAddonDeployment(cluster, singleProfileBody, clusterProfile)
			if err != nil {
				return diag.FromErr(err)
			}
		} else {
			// New profile, create it
			log.Printf("Adding profile %s to cluster %s", profile.UID, clusterUid)
			singleProfileBody := &models.V1SpectroClusterProfiles{
				Profiles:         []*models.V1SpectroClusterProfileEntity{profile},
				SpcApplySettings: addonDeployment.SpcApplySettings,
			}
			err = c.CreateAddonDeployment(cluster, singleProfileBody)
			if err != nil {
				return diag.FromErr(err)
			}
		}
	}

	// Update resource ID with new profile UIDs
	d.SetId(buildAddonDeploymentId(clusterUid, newProfileUIDs))

	// Wait for all profiles to be deployed
	for _, profile := range addonDeployment.Profiles {
		diagnostics, isError := waitForAddonDeploymentUpdate(ctx, d, *cluster, profile.UID, diags, c)
		if isError {
			return diagnostics
		}
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
