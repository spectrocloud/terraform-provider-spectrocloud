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

// readAddonDeploymentMultiple reads multiple addon profiles from the cluster
func readAddonDeploymentMultiple(c *client.V1Client, d *schema.ResourceData, cluster *models.V1SpectroCluster) (diag.Diagnostics, bool) {
	var diags diag.Diagnostics

	// Get the profile UIDs from resource ID
	profileUIDs, err := getClusterProfileUIDs(d.Id())
	if err != nil {
		// Fallback to legacy single profile handling
		return readAddonDeployment(c, d, cluster)
	}

	// Build a map of profile UIDs we're tracking
	trackedProfiles := make(map[string]bool)
	for _, uid := range profileUIDs {
		trackedProfiles[uid] = true
	}

	// Get existing profiles from config to preserve pack configurations
	existingProfiles := d.Get("cluster_profile").([]interface{})
	existingProfilesMap := make(map[string]map[string]interface{})
	for _, p := range existingProfiles {
		if p == nil {
			continue
		}
		profile := p.(map[string]interface{})
		if id, ok := profile["id"].(string); ok && id != "" {
			existingProfilesMap[id] = profile
		}
	}

	// Flatten all tracked profiles from cluster state
	clusterProfiles := make([]interface{}, 0)
	foundProfiles := make([]string, 0)

	for _, template := range cluster.Spec.ClusterProfileTemplates {
		if template == nil {
			continue
		}

		// Check if this profile is one we're tracking
		if !trackedProfiles[template.UID] {
			// Check by fetching the profile and matching by name (for version upgrades)
			found := false
			for trackedUID := range trackedProfiles {
				trackedProfile, err := c.GetClusterProfile(trackedUID)
				if err != nil {
					continue
				}
				if trackedProfile != nil && trackedProfile.Metadata != nil &&
					template.Name == trackedProfile.Metadata.Name {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}

		// Flatten this profile
		profileData, err := flattenAddonDeploymentProfile(c, d, template, existingProfilesMap)
		if err != nil {
			log.Printf("Warning: failed to flatten profile %s: %v", template.UID, err)
			continue
		}

		clusterProfiles = append(clusterProfiles, profileData)
		foundProfiles = append(foundProfiles, template.UID)
	}

	// If no profiles found, mark resource as deleted
	if len(clusterProfiles) == 0 {
		d.SetId("")
		return diags, false
	}

	// Update cluster_profile in state
	if err := d.Set("cluster_profile", clusterProfiles); err != nil {
		return diag.FromErr(err), false
	}

	// Update resource ID with current profile UIDs
	clusterUid := d.Get("cluster_uid").(string)
	d.SetId(buildAddonDeploymentId(clusterUid, foundProfiles))

	return diags, true
}

// flattenAddonDeploymentProfile flattens a single profile template
func flattenAddonDeploymentProfile(c *client.V1Client, d *schema.ResourceData, profile *models.V1ClusterProfileTemplate, existingProfilesMap map[string]map[string]interface{}) (map[string]interface{}, error) {
	clusterProfile := make(map[string]interface{})
	clusterProfile["id"] = profile.UID

	// Check if user has defined packs in their config for this profile
	hasPacksInConfig := false
	if existingConfig, exists := existingProfilesMap[profile.UID]; exists {
		if packs, ok := existingConfig["pack"]; ok && packs != nil {
			packsList := packs.([]interface{})
			if len(packsList) > 0 {
				hasPacksInConfig = true
			}
		}
	}

	// Only flatten packs if user has defined them in config
	if hasPacksInConfig {
		packManifests, diagResult, done := getPacksContent(profile.Packs, c, d)
		if done {
			log.Printf("Warning: error getting pack content for profile %s: %v", profile.UID, diagResult)
		} else {
			diagPacks := make([]*models.V1PackManifestEntity, 0)
			if existingConfig, exists := existingProfilesMap[profile.UID]; exists {
				if packsRaw, ok := existingConfig["pack"]; ok && packsRaw != nil {
					for _, pack := range packsRaw.([]interface{}) {
						if p, e := toAddonDeploymentPackCreate(pack); e == nil {
							diagPacks = append(diagPacks, p)
						}
					}
				}
			}

			// Build registry maps for this profile
			registryNameMap := buildPackRegistryNameMapForProfile(existingProfilesMap[profile.UID])
			registryUIDMap := buildPackRegistryUIDMapForProfile(existingProfilesMap[profile.UID])

			packs, err := flattenPacksWithRegistryMaps(c, diagPacks, profile.Packs, packManifests, registryNameMap, registryUIDMap)
			if err == nil {
				clusterProfile["pack"] = packs
			}
		}
	}

	// Flatten profile variables
	clusterUID := d.Get("cluster_uid").(string)
	if clusterUID != "" {
		clusterVars, err := c.GetClusterVariables(clusterUID)
		if err == nil {
			for _, clusterVar := range clusterVars {
				if clusterVar.ProfileUID != nil && *clusterVar.ProfileUID == profile.UID && clusterVar.Variables != nil {
					vars := make(map[string]interface{})
					for _, v := range clusterVar.Variables {
						if v.Name != nil && v.Value != "" {
							vars[*v.Name] = v.Value
						}
					}
					if len(vars) > 0 {
						clusterProfile["variables"] = vars
					}
					break
				}
			}
		}
	}

	return clusterProfile, nil
}

// buildPackRegistryNameMapForProfile builds a registry name map for a single profile
// Returns a map indicating which packs have registry_name set
func buildPackRegistryNameMapForProfile(profileConfig map[string]interface{}) map[string]bool {
	registryNameMap := make(map[string]bool)
	if profileConfig == nil {
		return registryNameMap
	}

	if packsRaw, ok := profileConfig["pack"]; ok && packsRaw != nil {
		for _, pack := range packsRaw.([]interface{}) {
			p := pack.(map[string]interface{})
			packName := p["name"].(string)
			if regName, ok := p["registry_name"]; ok && regName != nil && regName.(string) != "" {
				registryNameMap[packName] = true
			}
		}
	}
	return registryNameMap
}

// buildPackRegistryUIDMapForProfile builds a registry UID map for a single profile
// Returns a map indicating which packs have registry_uid set
func buildPackRegistryUIDMapForProfile(profileConfig map[string]interface{}) map[string]bool {
	registryUIDMap := make(map[string]bool)
	if profileConfig == nil {
		return registryUIDMap
	}

	if packsRaw, ok := profileConfig["pack"]; ok && packsRaw != nil {
		for _, pack := range packsRaw.([]interface{}) {
			p := pack.(map[string]interface{})
			packName := p["name"].(string)
			if regUID, ok := p["registry_uid"]; ok && regUID != nil && regUID.(string) != "" {
				registryUIDMap[packName] = true
			}
		}
	}
	return registryUIDMap
}

// readAddonDeployment is the legacy function for reading a single addon profile (kept for backward compatibility)
func readAddonDeployment(c *client.V1Client, d *schema.ResourceData, cluster *models.V1SpectroCluster) (diag.Diagnostics, bool) {
	var diags diag.Diagnostics

	profileId, err := getClusterProfileUID(d.Id())
	if err != nil {
		return nil, false
	}

	clusterProfile, err := c.GetClusterProfile(profileId)
	if err != nil {
		return nil, false
	}

	for _, profile := range cluster.Spec.ClusterProfileTemplates {
		if profile != nil && clusterProfile != nil {
			if profile.Name == clusterProfile.Metadata.Name {
				if profile.ProfileVersion == clusterProfile.Spec.Published.ProfileVersion {
					diagnostics, done := flattenAddonDeployment(c, d, profile)
					if done {
						return diagnostics, true
					}
					return diags, true
				}
			}
		}
	}

	d.SetId("") // deleted.
	return diags, false
}

func flattenAddonDeployment(c *client.V1Client, d *schema.ResourceData, profile *models.V1ClusterProfileTemplate) (diag.Diagnostics, bool) {
	var diags diag.Diagnostics

	// Check if user has defined packs in their config
	hasPacksInConfig := false
	profiles := d.Get("cluster_profile").([]interface{})
	if len(profiles) > 0 {
		for _, p := range profiles {
			profileMap := p.(map[string]interface{})
			if packs, ok := profileMap["pack"]; ok && packs != nil {
				packsList := packs.([]interface{})
				if len(packsList) > 0 {
					hasPacksInConfig = true
					break
				}
			}
		}
	}

	cluster_profiles := make([]interface{}, 0)
	cluster_profile := make(map[string]interface{})
	cluster_profile["id"] = profile.UID

	// Only flatten packs if user has defined them in config
	if hasPacksInConfig {
		packManifests, d2, done2 := getPacksContent(profile.Packs, c, d)
		if done2 {
			return d2, false
		}

		diagPacks, diagnostics, done := GetAddonDeploymentDiagPacks(d, nil)
		if done {
			return diagnostics, false
		}

		// Build registry maps to track which packs use registry_name or registry_uid
		registryNameMap := buildPackRegistryNameMap(d)
		registryUIDMap := buildPackRegistryUIDMap(d)
		packs, err := flattenPacksWithRegistryMaps(c, diagPacks, profile.Packs, packManifests, registryNameMap, registryUIDMap)
		if err != nil {
			return diag.FromErr(err), false
		}
		cluster_profile["pack"] = packs
	}

	// Flatten profile variables
	clusterUID := d.Get("cluster_uid").(string)
	if clusterUID != "" {
		clusterVars, err := c.GetClusterVariables(clusterUID)
		if err == nil {
			// Find variables for this specific profile
			for _, clusterVar := range clusterVars {
				if clusterVar.ProfileUID != nil && *clusterVar.ProfileUID == profile.UID && clusterVar.Variables != nil {
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

	cluster_profiles = append(cluster_profiles, cluster_profile)

	if err := d.Set("cluster_profile", cluster_profiles); err != nil {
		return diag.FromErr(err), false
	}

	return diags, true
}

func GetAddonDeploymentDiagPacks(d *schema.ResourceData, err error) ([]*models.V1PackManifestEntity, diag.Diagnostics, bool) {
	diagPacks := make([]*models.V1PackManifestEntity, 0)
	profiles := d.Get("cluster_profile").([]interface{})
	if len(profiles) > 0 {
		for _, profile := range profiles {
			p := profile.(map[string]interface{})
			for _, pack := range p["pack"].([]interface{}) {
				if p, e := toAddonDeploymentPackCreate(pack); e != nil {
					return nil, diag.FromErr(err), true
				} else {
					diagPacks = append(diagPacks, p)
				}
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
