package spectrocloud

import (
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/spectrocloud/palette-sdk-go/client"

	"log"

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

	// CRITICAL FIX: Read profiles from CURRENT CONFIG first (desired state)
	// Then fallback to state if config is empty (for backward compatibility)
	// This ensures that if a profile is removed from config, it won't be read back from state
	tfProfileIDs := make(map[string]bool)

	// First, read from CURRENT CONFIG (desired state)
	configProfilesRaw, ok := d.GetOk("cluster_profile")
	if ok && configProfilesRaw != nil {
		if configProfilesList, ok := configProfilesRaw.([]interface{}); ok {
			for _, profileRaw := range configProfilesList {
				profile := profileRaw.(map[string]interface{})
				if id, ok := profile["id"].(string); ok && id != "" {
					tfProfileIDs[id] = true
					log.Printf("[DEBUG] readAddonDeployment: Found profile %s in config", id)
				}
			}
		}
	}

	// If no profiles in config, fallback to state (for backward compatibility or initial read)
	if len(tfProfileIDs) == 0 && d.Id() != "" {
		stateProfilesRaw := d.Get("cluster_profile")
		if stateProfilesRaw != nil {
			if stateProfilesList, ok := stateProfilesRaw.([]interface{}); ok && len(stateProfilesList) > 0 {
				// State has profiles - add them all (for initial read or when config is empty)
				for _, profileRaw := range stateProfilesList {
					profile := profileRaw.(map[string]interface{})
					if id, ok := profile["id"].(string); ok && id != "" {
						tfProfileIDs[id] = true
						log.Printf("[DEBUG] readAddonDeployment: Found profile %s in state (fallback)", id)
					}
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
	diagPacks, _, _ := GetAddonDeploymentDiagPacks(d, nil)
	stateHasPacks := len(diagPacks) > 0

	// Read all profiles from cluster that match Terraform config
	cluster_profiles := make([]interface{}, 0)

	// Iterate through all profile IDs from config (or state if config is empty)
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

		// Verify it's an add-on profile
		if string(clusterProfile.Spec.Published.Type) != string(models.V1ProfileTypeAddDashOn) {
			log.Printf("Warning: Profile %s is not an add-on profile, skipping", profileID)
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

		// If profile not found on cluster, skip it (it was deleted or not yet attached)
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
		cluster_profile, diagnostics, done := flattenAddonDeployment(c, d, profileTemplate, profileID, stateHasPacks)
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
func flattenAddonDeployment(c *client.V1Client, d *schema.ResourceData, profile *models.V1ClusterProfileTemplate, profileId string, stateHasPacks bool) (map[string]interface{}, diag.Diagnostics, bool) {
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

	// CRITICAL FIX: If diagPacks is empty (config didn't specify packs), create diagPacks from profile definition
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
	cluster_profile["id"] = profileId

	// CRITICAL FIX: Only set packs in state if state already had packs
	// If state didn't have packs and config doesn't have packs, set empty array to avoid diff
	// This prevents Terraform from showing a diff when config and state both don't have packs
	if stateHasPacks {
		// State had packs - set the flattened packs from cluster
		log.Printf("[DEBUG] flattenAddonDeployment: State had packs, setting %d packs for profile %s", len(packs), profileId)
		cluster_profile["pack"] = packs
	} else {
		// Config doesn't have packs, so don't set them in state (set empty to match config)
		// This prevents Terraform from showing a diff when config and state both don't have packs
		log.Printf("[DEBUG] flattenAddonDeployment: State had no packs, setting empty array for profile %s", profileId)
		cluster_profile["pack"] = make([]interface{}, 0)
	}

	// Flatten profile variables
	clusterUID := d.Get("cluster_uid").(string)
	if clusterUID != "" {
		clusterVars, err := c.GetClusterVariables(clusterUID)
		if err == nil {
			// Find variables for this specific profile (use profileId, not profile.UID)
			for _, clusterVar := range clusterVars {
				if clusterVar.ProfileUID != nil && *clusterVar.ProfileUID == profileId && clusterVar.Variables != nil {
					vars := make(map[string]interface{})
					for _, v := range clusterVar.Variables {
						if v.Name != nil && v.Value != "" {
							vars[*v.Name] = v.Value
						}
					}
					if len(vars) > 0 {
						cluster_profile["variables"] = vars
					}
					break
				}
			}
		}
	}

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
