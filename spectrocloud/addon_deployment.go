package spectrocloud

import (
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/spectrocloud/palette-sdk-go/client"

	"github.com/spectrocloud/terraform-provider-spectrocloud/spectrocloud/schemas"
	"github.com/spectrocloud/terraform-provider-spectrocloud/types"
)

func readAddonDeployment(c *client.V1Client, d *schema.ResourceData, cluster *models.V1SpectroCluster) (diag.Diagnostics, bool) {

	var diags diag.Diagnostics

	// Get cluster UID from resource data
	clusterUid := d.Get("cluster_uid").(string)
	if clusterUid == "" {
		d.SetId("")
		return diags, false
	}

	if cluster == nil || cluster.Spec == nil || cluster.Spec.ClusterProfileTemplates == nil {
		d.SetId("")
		return diags, false
	}

	// During ReadContext, we need to read from STATE to detect all profiles
	// The issue: resource ID format is clusterUid_profileUid (only one profile)
	// But we might have multiple profiles. We need to read ALL profiles from state.
	// CRITICAL FIX: Also read ALL add-on profiles from cluster to handle incomplete state.
	// If state only has 1 profile but cluster has 2 add-on profiles, we'll read both.
	tfProfileIDs := make(map[string]bool)

	// First, read from state (if resource exists)
	if d.Id() != "" {
		stateProfilesRaw := d.Get("cluster_profile")
		if stateProfilesRaw != nil {
			if stateProfilesList, ok := stateProfilesRaw.([]interface{}); ok && len(stateProfilesList) > 0 {
				// State has profiles - add them all
				for _, profileRaw := range stateProfilesList {
					profile := profileRaw.(map[string]interface{})
					if id, ok := profile["id"].(string); ok && id != "" {
						tfProfileIDs[id] = true
					}
				}
			}
		}

		// FIX: Only verify profiles from state exist on cluster, don't add new ones
		// This ensures we detect deletions (profile removed from config but still in state)
		// while still handling cases where a profile in state was deleted externally
		profilesToRemove := make([]string, 0)
		for profileID := range tfProfileIDs {
			// Verify this profile exists on cluster
			foundOnCluster := false
			for _, templateProfile := range cluster.Spec.ClusterProfileTemplates {
				if templateProfile != nil && templateProfile.UID == profileID {
					// Verify it's an add-on profile
					profileDef, err := c.GetClusterProfile(templateProfile.UID)
					if err == nil && profileDef != nil && profileDef.Spec != nil && profileDef.Spec.Published != nil {
						if string(profileDef.Spec.Published.Type) == string(models.V1ProfileTypeAddDashOn) {
							foundOnCluster = true
							break
						}
					}
				}
			}
			// If profile in state doesn't exist on cluster, remove it (was deleted externally)
			if !foundOnCluster {
				profilesToRemove = append(profilesToRemove, profileID)
			}
		}
		// Remove profiles that don't exist on cluster
		for _, profileID := range profilesToRemove {
			delete(tfProfileIDs, profileID)
		}
	} else {
		// New resource - read from config
		configProfilesRaw, ok := d.GetOk("cluster_profile")
		if !ok || configProfilesRaw == nil {
			d.SetId("")
			return diags, false
		}
		if configProfilesList, ok := configProfilesRaw.([]interface{}); ok {
			for _, profileRaw := range configProfilesList {
				profile := profileRaw.(map[string]interface{})
				if id, ok := profile["id"].(string); ok && id != "" {
					tfProfileIDs[id] = true
				}
			}
		}
	}

	// If no profiles found, mark as deleted
	if len(tfProfileIDs) == 0 {
		d.SetId("")
		return diags, false
	}

	// Check if state already has packs BEFORE we start processing profiles
	// This is critical: if state doesn't have packs and config doesn't have packs,
	// we should NOT set packs in state to avoid Terraform showing a diff
	diagPacks, _, _ := GetAddonDeploymentDiagPacks(d, nil)
	stateHasPacks := len(diagPacks) > 0

	// Read all profiles from cluster that match Terraform config (current + old)
	cluster_profiles := make([]interface{}, 0)

	// Iterate through all profile IDs (from both current config and old state)
	for profileID := range tfProfileIDs {
		// Get the profile definition to get name and version
		clusterProfile, err := c.GetClusterProfile(profileID)
		if err != nil || clusterProfile == nil {
			log.Printf("Warning: Could not get profile %s: %v", profileID, err)
			continue
		}

		if clusterProfile.Metadata == nil || clusterProfile.Spec == nil || clusterProfile.Spec.Published == nil {
			continue
		}

		// Find matching profile on cluster
		var clusterTemplateProfile *models.V1ClusterProfileTemplate
		for _, templateProfile := range cluster.Spec.ClusterProfileTemplates {
			if templateProfile != nil {
				if templateProfile.Name == clusterProfile.Metadata.Name {
					if templateProfile.ProfileVersion == clusterProfile.Spec.Published.ProfileVersion {
						clusterTemplateProfile = templateProfile
						break
					}
				}
			}
		}

		// If profile not found on cluster, skip it (it will be added on next apply, or was already deleted)
		if clusterTemplateProfile == nil {
			log.Printf("Profile %s (name: %s) not found on cluster, skipping", profileID, clusterProfile.Metadata.Name)
			continue
		}

		// Use pack values from cluster profile definition (has full values)
		profileTemplate := &models.V1ClusterProfileTemplate{
			UID:   clusterTemplateProfile.UID,
			Packs: clusterProfile.Spec.Published.Packs, // Use packs from profile definition which have full values
		}

		// Use the refactored function to flatten this profile
		cluster_profile, diagnostics, done := flattenAddonDeployment(c, d, profileTemplate, stateHasPacks)
		if done {
			return diagnostics, false
		}
		if cluster_profile != nil {
			cluster_profiles = append(cluster_profiles, cluster_profile)
		}
	}

	// If no profiles found on cluster, mark as deleted
	if len(cluster_profiles) == 0 {
		d.SetId("")
		return diags, false
	}

	// Set all profiles in state
	if err := d.Set("cluster_profile", cluster_profiles); err != nil {
		return diag.FromErr(err), false
	}

	return diags, true
}

// flattenAddonDeployment flattens a single profile and returns it as a map
// Returns (profileMap, diagnostics, done)
func flattenAddonDeployment(c *client.V1Client, d *schema.ResourceData, profile *models.V1ClusterProfileTemplate, stateHasPacks bool) (map[string]interface{}, diag.Diagnostics, bool) {
	var diags diag.Diagnostics

	packManifests, d2, done2 := getPacksContent(profile.Packs, c, d)
	if done2 {
		return nil, d2, true
	}

	// Get diagPacks from state to check if state already has packs
	diagPacks, diagnostics, done := GetAddonDeploymentDiagPacks(d, nil)
	if done {
		return nil, diagnostics, true
	}

	// If diagPacks is empty (config didn't specify packs), create diagPacks from profile definition
	// This ensures registry maps are built correctly even when config doesn't have pack block
	if len(diagPacks) == 0 && len(profile.Packs) > 0 {
		// Create diagPacks from profile definition packs for registry mapping
		for _, pack := range profile.Packs {
			if pack.Name != nil {
				packType := models.V1PackType(pack.Type)
				diagPack := &models.V1PackManifestEntity{
					Name: pack.Name,
					Tag:  pack.Tag,
					Type: &packType,
				}
				diagPacks = append(diagPacks, diagPack)
			}
		}
	}

	// Build registry maps to track which packs use registry_name or registry_uid
	registryNameMap := buildPackRegistryNameMap(d)
	registryUIDMap := buildPackRegistryUIDMap(d)
	packs, err := flattenPacksWithRegistryMaps(c, diagPacks, profile.Packs, packManifests, registryNameMap, registryUIDMap)
	if err != nil {
		return nil, diag.FromErr(err), true
	}

	cluster_profile := make(map[string]interface{})
	// Only set packs in state if state already had packs
	// If state didn't have packs and config doesn't have packs, set empty array to avoid diff
	if stateHasPacks {
		cluster_profile["pack"] = packs
	} else {
		// Config doesn't have packs, so don't set them in state (set empty to match config)
		// This prevents Terraform from showing a diff when config and state both don't have packs
		cluster_profile["pack"] = make([]interface{}, 0)
	}
	cluster_profile["id"] = profile.UID

	return cluster_profile, diags, false
}

func GetAddonDeploymentDiagPacks(d *schema.ResourceData, err error) ([]*models.V1PackManifestEntity, diag.Diagnostics, bool) {
	diagPacks := make([]*models.V1PackManifestEntity, 0)
	profilesRaw := d.Get("cluster_profile")
	if profilesRaw == nil {
		return diagPacks, nil, false
	}

	profiles, ok := profilesRaw.([]interface{})
	if !ok || len(profiles) == 0 {
		return diagPacks, nil, false
	}

	for _, profile := range profiles {
		p, ok := profile.(map[string]interface{})
		if !ok {
			continue
		}

		packRaw, ok := p["pack"]
		if !ok || packRaw == nil {
			continue
		}

		packs, ok := packRaw.([]interface{})
		if !ok || len(packs) == 0 {
			continue
		}

		for _, pack := range packs {
			if p, e := toAddonDeploymentPackCreate(pack); e != nil {
				return nil, diag.FromErr(err), true
			} else {
				diagPacks = append(diagPacks, p)
			}
		}
	}

	return diagPacks, nil, false
}

func toAddonDeploymentPackCreate(pSrc interface{}) (*models.V1PackManifestEntity, error) {
	p := pSrc.(map[string]interface{})

	pName := p["name"].(string)
	pTag := p["tag"].(string)
	pRegistryUID := ""
	if p["registry_uid"] != nil {
		pRegistryUID = p["registry_uid"].(string)
	}
	pType := models.V1PackType(p["type"].(string))

	// Validate pack fields (validates both registry_uid and registry_name)
	if err := schemas.ValidatePackUIDOrResolutionFields(p); err != nil {
		return nil, err
	}

	pack := &models.V1PackManifestEntity{
		Name:        types.Ptr(pName),
		Tag:         pTag,
		RegistryUID: pRegistryUID,
		Type:        &pType,
		// UI strips a single newline, so we should do the same
		Values: strings.TrimSpace(p["values"].(string)),
	}

	manifests := make([]*models.V1ManifestInputEntity, 0)
	if len(p["manifest"].([]interface{})) > 0 {
		for _, manifest := range p["manifest"].([]interface{}) {
			m := manifest.(map[string]interface{})
			manifests = append(manifests, &models.V1ManifestInputEntity{
				Content: strings.TrimSpace(m["content"].(string)),
				Name:    m["name"].(string),
			})
		}
	}
	pack.Manifests = manifests

	return pack, nil
}
