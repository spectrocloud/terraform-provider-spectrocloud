package spectrocloud

import (
	"context"
	"errors"
	"fmt"
	"log"
	"reflect"
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
				Type:     schema.TypeString,
				Required: true,
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

	// Warning or errors can be collected in a slice type
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
		//return diag.FromErr(errors.New(fmt.Sprintf("Cluster: %s: Profile is already attached: %s", cluster.Metadata.UID, addonDeployment.Profiles[0].UID)))
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

	resourceAddonDeploymentRead(ctx, d, m)

	return diags
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
		return "", errors.New("")
	}
	return strings.Split(addonDeploymentId, "_")[1], nil
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
	// Warning or errors can be collected in a slice type
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
	log.Printf("[DEBUG] updateAddonDeployment: Starting update for cluster %s", clusterUid)

	// Get old and new cluster_profile values
	oldProfilesRaw, newProfilesRaw := d.GetChange("cluster_profile")

	var oldProfiles, newProfiles []interface{}
	if oldProfilesRaw != nil {
		oldProfiles = oldProfilesRaw.([]interface{})
	}
	if newProfilesRaw != nil {
		newProfiles = newProfilesRaw.([]interface{})
	}

	log.Printf("[DEBUG] updateAddonDeployment: Old profiles count: %d, New profiles count: %d", len(oldProfiles), len(newProfiles))

	// Build maps for comparison
	// Key: profile name (for matching), Value: profile map
	oldProfilesMap := make(map[string]map[string]interface{})
	oldProfileUIDs := make(map[string]string) // name -> UID mapping for old profiles
	oldProfileUIDSet := make(map[string]bool) // UID -> bool for direct UID matching

	for _, p := range oldProfiles {
		profileMap := p.(map[string]interface{})
		if profileID, ok := profileMap["id"].(string); ok && profileID != "" {
			// Add to UID set for direct matching (more reliable)
			oldProfileUIDSet[profileID] = true

			// Get profile name from cluster or profile definition (for name-based matching)
			profileDef, err := c.GetClusterProfile(profileID)
			if err == nil && profileDef != nil && profileDef.Metadata != nil {
				profileName := profileDef.Metadata.Name
				oldProfilesMap[profileName] = profileMap
				oldProfileUIDs[profileName] = profileID
				log.Printf("[DEBUG] updateAddonDeployment: Old profile - Name: %s, UID: %s", profileName, profileID)
			} else {
				// Even if GetClusterProfile fails, we still track the UID for deletion
				log.Printf("[DEBUG] updateAddonDeployment: Old profile - UID: %s (could not get name, will use UID for deletion)", profileID)
			}
		}
	}

	newProfilesMap := make(map[string]map[string]interface{})
	newProfileUIDs := make(map[string]string) // name -> UID mapping for new profiles
	newProfileUIDSet := make(map[string]bool) // UID -> bool for direct UID matching

	for _, p := range newProfiles {
		profileMap := p.(map[string]interface{})
		if profileID, ok := profileMap["id"].(string); ok && profileID != "" {
			// Add to UID set for direct matching
			newProfileUIDSet[profileID] = true

			// Get profile name from profile definition (for name-based matching)
			profileDef, err := c.GetClusterProfile(profileID)
			if err == nil && profileDef != nil && profileDef.Metadata != nil {
				profileName := profileDef.Metadata.Name
				newProfilesMap[profileName] = profileMap
				newProfileUIDs[profileName] = profileID
				log.Printf("[DEBUG] updateAddonDeployment: New profile - Name: %s, UID: %s", profileName, profileID)
			} else {
				log.Printf("[WARN] updateAddonDeployment: Could not get profile definition for UID: %s", profileID)
			}
		}
	}

	// CRITICAL FIX: Detect profiles to delete by UID directly (more reliable than name matching)
	// A profile should be deleted if it's in old but not in new (by UID)
	profilesToDelete := make([]string, 0)

	for oldUID := range oldProfileUIDSet {
		if !newProfileUIDSet[oldUID] {
			// Profile is in old state but not in new config - should be deleted
			profilesToDelete = append(profilesToDelete, oldUID)
			log.Printf("[DEBUG] updateAddonDeployment: Profile UID %s will be deleted (not in new config)", oldUID)
		}
	}

	// Check for replace scenario: exactly one delete and one add
	if len(profilesToDelete) == 1 && len(newProfilesMap) == 1 && len(oldProfilesMap) == 1 {
		// This is a replace scenario - use ReplaceWithProfile instead of delete + create
		oldUIDToReplace := profilesToDelete[0]
		var newProfileName string
		var newProfileMap map[string]interface{}
		for name, profileMap := range newProfilesMap {
			newProfileName = name
			newProfileMap = profileMap
			break
		}

		log.Printf("[DEBUG] updateAddonDeployment: REPLACE SCENARIO detected - Replacing profile UID %s with profile %s (UID: %s)",
			oldUIDToReplace, newProfileName, newProfileUIDs[newProfileName])

		// Create update body with ReplaceWithProfile
		updateBody, err := toAddonDeploymentForProfile(c, d, newProfileMap)
		if err != nil {
			return diag.FromErr(fmt.Errorf("failed to convert profile %s for replace: %w", newProfileName, err))
		}

		// Set ReplaceWithProfile to old UID
		if len(updateBody.Profiles) > 0 {
			updateBody.Profiles[0].ReplaceWithProfile = oldUIDToReplace
		}

		newProfile, err := c.GetClusterProfile(newProfileUIDs[newProfileName])
		if err != nil {
			return diag.FromErr(fmt.Errorf("failed to get profile %s: %w", newProfileName, err))
		}

		// Refresh cluster to get latest state
		cluster, err = c.GetCluster(clusterUid)
		if err != nil {
			return diag.FromErr(fmt.Errorf("failed to refresh cluster: %w", err))
		}

		err = c.UpdateAddonDeployment(cluster, updateBody, newProfile)
		if err != nil {
			return diag.FromErr(fmt.Errorf("failed to replace profile: %w", err))
		}

		// Wait for replace to complete
		diagnostics, isError := waitForAddonDeploymentUpdate(ctx, d, *cluster, newProfileUIDs[newProfileName], diags, c)
		if isError {
			return diagnostics
		}

		// Update resource ID to reflect the new profile
		if len(newProfiles) > 0 {
			firstProfile := newProfiles[0].(map[string]interface{})
			if profileID, ok := firstProfile["id"].(string); ok && profileID != "" {
				clusterProfile, err := c.GetClusterProfile(profileID)
				if err == nil && clusterProfile != nil {
					d.SetId(getAddonDeploymentId(clusterUid, clusterProfile))
				}
			}
		}

		// Refresh state
		resourceAddonDeploymentRead(ctx, d, m)

		return diags
	}

	// 1. Handle PROFILE DELETIONS (in old but not in new)
	// Only delete if not in replace scenario
	if len(profilesToDelete) > 0 {
		log.Printf("[DEBUG] updateAddonDeployment: Deleting %d profiles: %v", len(profilesToDelete), profilesToDelete)
		err := c.DeleteAddonDeployment(clusterUid, &models.V1SpectroClusterProfilesDeleteEntity{
			ProfileUids: profilesToDelete,
		})
		if err != nil {
			return diag.FromErr(fmt.Errorf("failed to delete profiles: %w", err))
		}

		// CRITICAL FIX: Wait for deletion to complete before proceeding
		// Wait for each deleted profile to be fully removed
		for _, deletedUID := range profilesToDelete {
			log.Printf("[DEBUG] updateAddonDeployment: Waiting for profile %s deletion to complete", deletedUID)
			// Refresh cluster to get latest state
			cluster, err = c.GetCluster(clusterUid)
			if err != nil {
				return diag.FromErr(fmt.Errorf("failed to refresh cluster after deletion: %w", err))
			}

			// Wait until profile is no longer attached
			maxRetries := 30
			retryCount := 0
			for retryCount < maxRetries {
				if !isProfileAttached(cluster, deletedUID) {
					log.Printf("[DEBUG] updateAddonDeployment: Profile %s successfully deleted", deletedUID)
					break
				}
				retryCount++
				log.Printf("[DEBUG] updateAddonDeployment: Profile %s still attached, waiting... (retry %d/%d)", deletedUID, retryCount, maxRetries)
				time.Sleep(5 * time.Second)
				cluster, err = c.GetCluster(clusterUid)
				if err != nil {
					return diag.FromErr(fmt.Errorf("failed to refresh cluster during deletion wait: %w", err))
				}
			}
			if retryCount >= maxRetries {
				log.Printf("[WARN] updateAddonDeployment: Profile %s deletion wait timeout, proceeding anyway", deletedUID)
			}
		}
	}

	// 2. Handle PROFILE ADDITIONS and UPDATES (in new)
	for profileName, newProfileMap := range newProfilesMap {
		newProfileUID := newProfileUIDs[profileName]

		if oldProfileMap, exists := oldProfilesMap[profileName]; exists {
			// Profile exists in both old and new - check if it's an update
			oldProfileUID := oldProfileUIDs[profileName]

			if oldProfileUID != newProfileUID {
				// Profile ID changed but name is same - VERSION UPDATE (ReplaceWithProfile)
				log.Printf("[DEBUG] updateAddonDeployment: Profile %s version update - Old UID: %s, New UID: %s", profileName, oldProfileUID, newProfileUID)

				// Create update body with ReplaceWithProfile
				updateBody, err := toAddonDeploymentForProfile(c, d, newProfileMap)
				if err != nil {
					return diag.FromErr(fmt.Errorf("failed to convert profile %s for update: %w", profileName, err))
				}

				// Set ReplaceWithProfile to old UID
				if len(updateBody.Profiles) > 0 {
					updateBody.Profiles[0].ReplaceWithProfile = oldProfileUID
				}

				newProfile, err := c.GetClusterProfile(newProfileUID)
				if err != nil {
					return diag.FromErr(fmt.Errorf("failed to get profile %s: %w", profileName, err))
				}

				// Refresh cluster to get latest state
				cluster, err = c.GetCluster(clusterUid)
				if err != nil {
					return diag.FromErr(fmt.Errorf("failed to refresh cluster: %w", err))
				}

				err = c.UpdateAddonDeployment(cluster, updateBody, newProfile)
				if err != nil {
					return diag.FromErr(fmt.Errorf("failed to update profile %s: %w", profileName, err))
				}

				// Wait for update to complete
				diagnostics, isError := waitForAddonDeploymentUpdate(ctx, d, *cluster, newProfileUID, diags, c)
				if isError {
					return diagnostics
				}
			} else {
				// Same profile ID - check if packs or variables changed
				if hasProfileChanges(oldProfileMap, newProfileMap) {
					log.Printf("[DEBUG] updateAddonDeployment: Profile %s (UID: %s) has pack/variable changes", profileName, newProfileUID)

					// Create update body
					updateBody, err := toAddonDeploymentForProfile(c, d, newProfileMap)
					if err != nil {
						return diag.FromErr(fmt.Errorf("failed to convert profile %s for update: %w", profileName, err))
					}

					newProfile, err := c.GetClusterProfile(newProfileUID)
					if err != nil {
						return diag.FromErr(fmt.Errorf("failed to get profile %s: %w", profileName, err))
					}

					// Refresh cluster to get latest state
					cluster, err = c.GetCluster(clusterUid)
					if err != nil {
						return diag.FromErr(fmt.Errorf("failed to refresh cluster: %w", err))
					}

					err = c.UpdateAddonDeployment(cluster, updateBody, newProfile)
					if err != nil {
						return diag.FromErr(fmt.Errorf("failed to update profile %s: %w", profileName, err))
					}

					// Wait for update to complete
					diagnostics, isError := waitForAddonDeploymentUpdate(ctx, d, *cluster, newProfileUID, diags, c)
					if isError {
						return diagnostics
					}
				} else {
					log.Printf("[DEBUG] updateAddonDeployment: Profile %s (UID: %s) unchanged, skipping", profileName, newProfileUID)
				}
			}
		} else {
			// Profile is NEW - ADD it
			log.Printf("[DEBUG] updateAddonDeployment: Adding new profile %s (UID: %s)", profileName, newProfileUID)

			// Create addon deployment body for this profile
			addonBody, err := toAddonDeploymentForProfile(c, d, newProfileMap)
			if err != nil {
				return diag.FromErr(fmt.Errorf("failed to convert profile %s for creation: %w", profileName, err))
			}

			// Refresh cluster to get latest state
			cluster, err = c.GetCluster(clusterUid)
			if err != nil {
				return diag.FromErr(fmt.Errorf("failed to refresh cluster: %w", err))
			}

			// Check if profile is already attached (might have been added by another operation)
			if !isProfileAttached(cluster, newProfileUID) {
				err = c.CreateAddonDeployment(cluster, addonBody)
				if err != nil {
					return diag.FromErr(fmt.Errorf("failed to create addon deployment for profile %s: %w", profileName, err))
				}

				// Wait for creation to complete
				diagnostics, isError := waitForAddonDeploymentCreation(ctx, d, *cluster, newProfileUID, diags, c)
				if isError {
					return diagnostics
				}
			} else {
				log.Printf("[DEBUG] updateAddonDeployment: Profile %s (UID: %s) already attached, skipping creation", profileName, newProfileUID)
			}
		}
	}

	// Update resource ID to reflect the first profile (for backward compatibility)
	// Note: This assumes single profile per resource. For multi-profile support,
	// you might want to change the ID format or use a different approach
	if len(newProfiles) > 0 {
		firstProfile := newProfiles[0].(map[string]interface{})
		if profileID, ok := firstProfile["id"].(string); ok && profileID != "" {
			clusterProfile, err := c.GetClusterProfile(profileID)
			if err == nil && clusterProfile != nil {
				d.SetId(getAddonDeploymentId(clusterUid, clusterProfile))
			}
		}
	}

	// Refresh state
	resourceAddonDeploymentRead(ctx, d, m)

	return diags
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

// Helper function to create addon deployment body for a single profile
func toAddonDeploymentForProfile(c *client.V1Client, d *schema.ResourceData, profileMap map[string]interface{}) (*models.V1SpectroClusterProfiles, error) {
	profileID := profileMap["id"].(string)
	profileEntity := &models.V1SpectroClusterProfileEntity{
		UID: profileID,
	}

	// Handle variables
	if pv, ok := profileMap["variables"]; ok && pv != nil {
		variables := pv.(map[string]interface{})
		pVars := make([]*models.V1SpectroClusterVariable, 0)
		for key, value := range variables {
			if key != "" && value != nil {
				pVars = append(pVars, &models.V1SpectroClusterVariable{
					Name:  StringPtr(key),
					Value: value.(string),
				})
			}
		}
		profileEntity.Variables = pVars
	}

	// Handle pack values
	packValues := make([]*models.V1PackValuesEntity, 0)
	if packs, ok := profileMap["pack"]; ok && packs != nil {
		cluster, err := c.GetClusterWithoutStatus(d.Get("cluster_uid").(string))
		if err != nil {
			return nil, fmt.Errorf("failed to get cluster for pack conversion: %w", err)
		}
		for _, pack := range packs.([]interface{}) {
			p := toPack(cluster, pack)
			packValues = append(packValues, p)
		}
	}
	profileEntity.PackValues = packValues

	settings, err := toSpcApplySettings(d)
	if err != nil {
		return nil, err
	}

	return &models.V1SpectroClusterProfiles{
		Profiles:         []*models.V1SpectroClusterProfileEntity{profileEntity},
		SpcApplySettings: settings,
	}, nil
}

// Helper function to check if profile has changes (packs or variables)
func hasProfileChanges(oldProfile, newProfile map[string]interface{}) bool {
	// Check if packs changed
	oldPacks := oldProfile["pack"]
	newPacks := newProfile["pack"]
	if !reflect.DeepEqual(oldPacks, newPacks) {
		return true
	}

	// Check if variables changed
	oldVars := oldProfile["variables"]
	newVars := newProfile["variables"]
	return !reflect.DeepEqual(oldVars, newVars)
}
