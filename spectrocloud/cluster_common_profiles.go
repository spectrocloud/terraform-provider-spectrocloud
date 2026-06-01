package spectrocloud

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/spectrocloud/palette-sdk-go/client"

	"github.com/spectrocloud/terraform-provider-spectrocloud/types"
)

func normalizeInterfaceSliceFromListOrSet(v interface{}) []interface{} {
	switch t := v.(type) {
	case nil:
		return []interface{}{}
	case []interface{}:
		return t
	case *schema.Set:
		if t == nil {
			return []interface{}{}
		}
		return t.List()
	default:
		return []interface{}{}
	}
}

// validateProfileSource checks that only one of cluster_template or cluster_profile is specified
func validateProfileSource(d *schema.ResourceData) error {
	// cluster_template may not exist in all schemas (e.g., cluster_group)
	clusterTemplate := normalizeInterfaceSliceFromListOrSet(d.Get("cluster_template"))
	clusterProfile := normalizeInterfaceSliceFromListOrSet(d.Get("cluster_profile"))

	if len(clusterTemplate) > 0 && len(clusterProfile) > 0 {
		return errors.New("cannot specify both cluster_template and cluster_profile. Please use only one")
	}

	return nil
}

// extractProfilesFromTemplate extracts cluster_profile data from cluster_template schema
// and transforms it into the same structure as regular cluster_profile for processing
func extractProfilesFromTemplate(d *schema.ResourceData) ([]interface{}, error) {
	clusterTemplateRaw := d.Get("cluster_template")
	if clusterTemplateRaw == nil {
		return []interface{}{}, nil
	}
	clusterTemplate := clusterTemplateRaw.([]interface{})
	if len(clusterTemplate) == 0 {
		return []interface{}{}, nil
	}

	// cluster_template is a list with single item
	templateData := clusterTemplate[0].(map[string]interface{})

	// Extract cluster_profile set from template
	if clusterProfiles, ok := templateData["cluster_profile"]; ok && clusterProfiles != nil {
		profilesSet := clusterProfiles.(*schema.Set)
		rawProfiles := profilesSet.List()

		// Filter out empty/invalid profiles
		validProfiles := make([]interface{}, 0)
		for _, profile := range rawProfiles {
			if profile == nil {
				continue
			}

			p, ok := profile.(map[string]interface{})
			if !ok {
				continue
			}

			// Skip profiles without an ID
			profileID, hasID := p["id"]
			if !hasID || profileID == nil || profileID == "" {
				log.Printf("extractProfilesFromTemplate: skipping profile without ID")
				continue
			}

			validProfiles = append(validProfiles, profile)
		}

		log.Printf("extractProfilesFromTemplate: extracted %d valid profiles (filtered from %d total)", len(validProfiles), len(rawProfiles))
		return validProfiles, nil
	}

	return []interface{}{}, nil
}

// extractProfilesFromTemplateData extracts cluster_profile data from raw cluster_template data
// This is used during updates when we have the new template data from d.GetChange()
func extractProfilesFromTemplateData(clusterTemplateData []interface{}) ([]interface{}, error) {
	if len(clusterTemplateData) == 0 {
		log.Printf("extractProfilesFromTemplateData: empty cluster template data")
		return []interface{}{}, nil
	}

	// cluster_template is a list with single item
	templateData := clusterTemplateData[0].(map[string]interface{})

	// Extract cluster_profile set from template
	if clusterProfiles, ok := templateData["cluster_profile"]; ok && clusterProfiles != nil {
		profilesSet := clusterProfiles.(*schema.Set)
		rawProfiles := profilesSet.List()

		// Filter out empty/invalid profiles
		validProfiles := make([]interface{}, 0)
		for _, profile := range rawProfiles {
			if profile == nil {
				continue
			}

			p, ok := profile.(map[string]interface{})
			if !ok {
				continue
			}

			// Skip profiles without an ID
			profileID, hasID := p["id"]
			if !hasID || profileID == nil || profileID == "" {
				log.Printf("extractProfilesFromTemplateData: skipping profile without ID")
				continue
			}

			validProfiles = append(validProfiles, profile)
		}

		log.Printf("extractProfilesFromTemplateData: extracted %d valid profiles (filtered from %d total)", len(validProfiles), len(rawProfiles))
		return validProfiles, nil
	}

	log.Printf("extractProfilesFromTemplateData: no cluster_profile found in template data")
	return []interface{}{}, nil
}

// resolveProfileSource determines which source to use and returns the profile data
// Returns: (profiles, source, error) where source is "cluster_template" or "cluster_profile"
func resolveProfileSource(d *schema.ResourceData) ([]interface{}, string, error) {
	// First validate mutual exclusivity
	if err := validateProfileSource(d); err != nil {
		return nil, "", err
	}

	clusterTemplate := normalizeInterfaceSliceFromListOrSet(d.Get("cluster_template"))
	clusterProfile := normalizeInterfaceSliceFromListOrSet(d.Get("cluster_profile"))

	// Check cluster_template first
	if len(clusterTemplate) > 0 {
		profiles, err := extractProfilesFromTemplate(d)
		if err != nil {
			return nil, "", err
		}
		log.Printf("Using profiles from cluster_template")
		return profiles, "cluster_template", nil
	}

	// Fall back to cluster_profile
	if len(clusterProfile) > 0 {
		// Fall back to cluster_profile — filter out TypeSet zero-value artefacts (empty id)
		filtered := clusterProfile[:0]
		for _, p := range clusterProfile {
			if entry, ok := p.(map[string]interface{}); ok {
				if id, _ := entry["id"]; id != nil && id.(string) != "" {
					filtered = append(filtered, p)
				}
			}
		}
		if len(filtered) > 0 {
			log.Printf("Using profiles from cluster_profile")
			return filtered, "cluster_profile", nil
		}
	}

	return []interface{}{}, "", nil
}

func toProfiles(c *client.V1Client, d *schema.ResourceData, clusterContext string) ([]*models.V1SpectroClusterProfileEntity, error) {
	return toProfilesCommon(c, d, d.Id(), clusterContext)
}

func toAddonDeplProfiles(c *client.V1Client, d *schema.ResourceData) ([]*models.V1SpectroClusterProfileEntity, error) {
	clusterUid := ""
	clusterContext := ""
	// handling cluster attachment flow for cluster created outside terraform and attaching addon profile to it
	if uid, ok := d.GetOk("cluster_uid"); ok && uid != nil {
		clusterUid = uid.(string) //d.Get("cluster_uid").(string)
	}
	// handling cluster day 2 addon profile operation flow
	if clusterUid == "" {
		clusterUid = d.Id()
	}
	if ct, ok := d.GetOk("context"); ok && c != nil {
		clusterContext = ct.(string)
	}
	err := ValidateContext(clusterContext)
	if err != nil {
		return nil, err
	}
	return toProfilesCommon(c, d, clusterUid, clusterContext)
}

func toProfilesCommon(c *client.V1Client, d *schema.ResourceData, clusterUID, context string) ([]*models.V1SpectroClusterProfileEntity, error) {
	var cluster *models.V1SpectroCluster
	var err error
	if clusterUID != "" {
		cluster, err = c.GetClusterWithoutStatus(clusterUID)
		if err != nil || cluster == nil {
			return nil, fmt.Errorf("cluster %s cannot be retrieved in context %s", clusterUID, context)
		}
	}

	resp := make([]*models.V1SpectroClusterProfileEntity, 0)

	// Resolve profile source (cluster_template or cluster_profile)
	profiles, source, err := resolveProfileSource(d)
	if err != nil {
		return nil, err
	}

	if len(profiles) > 0 {
		for _, profile := range profiles {
			p := profile.(map[string]interface{})
			// Profile Variables handling
			pVars := make([]*models.V1SpectroClusterVariable, 0)
			if pv, ok := p["variables"]; ok && pv != nil {
				variables := p["variables"].(map[string]interface{})
				for key, value := range variables {
					pVars = append(pVars, &models.V1SpectroClusterVariable{
						Name:  StringPtr(key),
						Value: value.(string),
					})
				}
			}

			packValues := make([]*models.V1PackValuesEntity, 0)
			// Pack values only exist in cluster_profile, not in cluster_template
			if source == "cluster_profile" {
				if packs, ok := p["pack"]; ok && packs != nil {
					for _, pack := range p["pack"].([]interface{}) {
						p := toPack(cluster, pack)
						packValues = append(packValues, p)
					}
				}
			}

			resp = append(resp, &models.V1SpectroClusterProfileEntity{
				UID:        p["id"].(string),
				PackValues: packValues,
				Variables:  pVars,
			})
		}
	}

	return resp, nil
}

func toSpcApplySettings(d *schema.ResourceData) (*models.V1SpcApplySettings, error) {
	if d.Get("apply_setting") != nil {
		setting := d.Get("apply_setting").(string)
		if setting != "" {
			return &models.V1SpcApplySettings{
				ActionType: setting,
			}, nil
		}
	}

	return nil, nil
}

func toPack(cluster *models.V1SpectroCluster, pSrc interface{}) *models.V1PackValuesEntity {
	p := pSrc.(map[string]interface{})

	pack := &models.V1PackValuesEntity{
		Name: types.Ptr(p["name"].(string)),
	}

	setPackValues(pack, p)
	setPackTag(pack, p)
	setPackType(pack, p)
	setPackManifests(pack, p, cluster)

	return pack
}

func setPackValues(pack *models.V1PackValuesEntity, p map[string]interface{}) {
	if val, found := p["values"]; found && len(val.(string)) > 0 {
		pack.Values = val.(string)
	}
}

func setPackTag(pack *models.V1PackValuesEntity, p map[string]interface{}) {
	if val, found := p["tag"]; found && len(val.(string)) > 0 {
		pack.Tag = val.(string)
	}
}

func setPackType(pack *models.V1PackValuesEntity, p map[string]interface{}) {
	if val, found := p["type"]; found && len(val.(string)) > 0 {
		pack.Type = types.Ptr(models.V1PackType(val.(string)))
	}
}

func setPackManifests(pack *models.V1PackValuesEntity, p map[string]interface{}, cluster *models.V1SpectroCluster) {
	if val, found := p["manifest"]; found && len(val.([]interface{})) > 0 {
		manifestsData := val.([]interface{})
		manifests := make([]*models.V1ManifestRefUpdateEntity, len(manifestsData))
		for i := 0; i < len(manifestsData); i++ {
			data := manifestsData[i].(map[string]interface{})
			uid := ""
			if cluster != nil {
				packs := make([]*models.V1PackRef, 0)
				for _, profile := range cluster.Spec.ClusterProfileTemplates {
					packs = append(packs, profile.Packs...)
				}
				uid = getManifestUID(data["name"].(string), packs)
			}
			manifests[i] = &models.V1ManifestRefUpdateEntity{
				Name:    types.Ptr(data["name"].(string)),
				Content: data["content"].(string),
				UID:     uid,
			}
		}
		pack.Manifests = manifests
	}
}

// setReplaceWithProfileForExisting sets the ReplaceWithProfile field for each profile
// that already exists on the cluster. This is necessary when using PATCH to update
// profiles - without ReplaceWithProfile, PATCH would add duplicates instead of updating.
// It matches profiles by name: if a profile with the same name is already attached to
// the cluster, ReplaceWithProfile is set to that existing profile's UID.
func setReplaceWithProfileForExisting(c *client.V1Client, cluster *models.V1SpectroCluster, profiles []*models.V1SpectroClusterProfileEntity) error {
	if cluster == nil || len(profiles) == 0 {
		return nil
	}

	for _, profile := range profiles {
		if profile == nil || profile.UID == "" {
			continue
		}

		// Get the cluster profile to find its name
		clusterProfile, err := c.GetClusterProfile(profile.UID)
		if err != nil {
			return fmt.Errorf("failed to get cluster profile %s: %w", profile.UID, err)
		}
		if clusterProfile == nil || clusterProfile.Metadata == nil {
			continue
		}
		// Check if a profile with the same name is already attached to the cluster
		existingUID := findAttachedProfileByName(cluster, clusterProfile.Metadata.Name)
		// Infra swap: if no same-name match but this profile is infra/cluster type, replace the attached infra layer.
		if existingUID == "" && clusterProfile.Spec != nil && clusterProfile.Spec.Published != nil &&
			isInfraClusterProfileType(clusterProfile.Spec.Published.Type) {
			existingUID = findAttachedInfraProfile(cluster)
			if existingUID != "" {
				log.Printf("Profile %s (name: %s) is infra - will replace existing infra %s via PATCH",
					profile.UID, clusterProfile.Metadata.Name, existingUID)
			}
		}
		if existingUID != "" && existingUID != profile.UID {
			// Only set ReplaceWithProfile if the existing profile has a DIFFERENT UID
			log.Printf("Profile %s (name: %s) will replace existing attached profile %s",
				profile.UID, clusterProfile.Metadata.Name, existingUID)
			profile.ReplaceWithProfile = existingUID
		} else if existingUID == profile.UID {
			log.Printf("Profile %s (name: %s) is already attached with same UID, no replacement needed",
				profile.UID, clusterProfile.Metadata.Name)
		}
	}

	return nil
}

// isInfraClusterProfileType reports whether a profile template type is an infra layer (Palette treats
// both "infra" and "cluster" as infra; EKS profiles commonly use "cluster").
func isInfraClusterProfileType(profileType string) bool {
	return profileType == "infra" || profileType == "cluster"
}

// getAttachedProfileType returns the template type for a profile UID on the cluster, if attached.
func getAttachedProfileType(cluster *models.V1SpectroCluster, profileUID string) string {
	if cluster == nil || cluster.Spec == nil || profileUID == "" {
		return ""
	}
	for _, template := range cluster.Spec.ClusterProfileTemplates {
		if template != nil && template.UID == profileUID {
			return template.Type
		}
	}
	return ""
}

// findAttachedInfraProfile returns the UID of the first attached infra/cluster profile on the cluster.
func findAttachedInfraProfile(cluster *models.V1SpectroCluster) string {
	if cluster == nil || cluster.Spec == nil {
		return ""
	}
	for _, template := range cluster.Spec.ClusterProfileTemplates {
		if template != nil && isInfraClusterProfileType(template.Type) {
			return template.UID
		}
	}
	return ""
}

// // findAttachedProfileByType returns the UID of the first attached profile with the given type (e.g. "infra").
// // Used when replacing an infra profile with another of a different name.
// func findAttachedProfileByType(cluster *models.V1SpectroCluster, profileType string) string {
// 	if cluster == nil || cluster.Spec == nil || profileType == "" {
// 		return ""
// 	}
// 	for _, template := range cluster.Spec.ClusterProfileTemplates {
// 		if template != nil && template.Type == profileType {
// 			return template.UID
// 		}
// 	}
// 	return ""
// }

// findAttachedProfileByName finds a profile attached to the cluster by its name.
// Returns the UID of the attached profile if found, empty string otherwise.
func findAttachedProfileByName(cluster *models.V1SpectroCluster, profileName string) string {
	if cluster == nil || cluster.Spec == nil || profileName == "" {
		return ""
	}

	for _, template := range cluster.Spec.ClusterProfileTemplates {
		if template != nil && template.Name == profileName {
			return template.UID
		}
	}

	return ""
}

// isProfileAttachedToCluster reports whether profileUID is attached on the cluster document.
func isProfileAttachedToCluster(cluster *models.V1SpectroCluster, profileUID string) bool {
	if cluster == nil || cluster.Spec == nil || profileUID == "" {
		return false
	}
	for _, template := range cluster.Spec.ClusterProfileTemplates {
		if template != nil && template.UID == profileUID {
			return true
		}
	}
	return false
}

// getProfilesToDelete returns profile UIDs to remove from the cluster when they are no longer
// in Terraform config. Comparison is by UID: an old profile UID that is still attached on the
// cluster but absent from the new config is deleted via DeleteAddonDeployment.
//
// Previously deletion was skipped when another config profile shared the same profile name
// (treating it as a version upgrade via ReplaceWithProfile). That left duplicate attachments
// on the cluster when Palette attached a new profile version (new UID, same name) while the
// old UID was removed from config.
func getProfilesToDelete(c *client.V1Client, d *schema.ResourceData, cluster *models.V1SpectroCluster) []string {
	oldProfilesRaw, newProfilesRaw := d.GetChange("cluster_profile")
	oldProfiles := normalizeInterfaceSliceFromListOrSet(oldProfilesRaw)
	newProfiles := normalizeInterfaceSliceFromListOrSet(newProfilesRaw)

	newProfileUIDs := make(map[string]bool, len(newProfiles))
	for _, p := range newProfiles {
		if p == nil {
			continue
		}
		profile := p.(map[string]interface{})
		if id, ok := profile["id"].(string); ok && id != "" {
			newProfileUIDs[id] = true
		}
	}

	var profilesToDelete []string
	for _, p := range oldProfiles {
		if p == nil {
			continue
		}
		profile := p.(map[string]interface{})
		oldUID, ok := profile["id"].(string)
		if !ok || oldUID == "" || newProfileUIDs[oldUID] {
			continue
		}

		if !isProfileAttachedToCluster(cluster, oldUID) {
			log.Printf("Profile %s removed from config but not attached on cluster - skipping API delete", oldUID)
			continue
		}

		if isInfraClusterProfileType(getAttachedProfileType(cluster, oldUID)) {
			log.Printf("Profile %s is infra/cluster on cluster - skip delete, will be replaced via PATCH", oldUID)
			continue
		}

		if c != nil {
			clusterProfile, err := c.GetClusterProfile(oldUID)
			if err != nil {
				log.Printf("Warning: could not get profile %s for delete check: %v", oldUID, err)
				profilesToDelete = append(profilesToDelete, oldUID)
				continue
			}
			if clusterProfile != nil && clusterProfile.Spec != nil && clusterProfile.Spec.Published != nil &&
				isInfraClusterProfileType(clusterProfile.Spec.Published.Type) {
				log.Printf("Profile %s is infra/cluster profile - skip delete, will be replaced via PATCH", oldUID)
				continue
			}
		}

		log.Printf("Profile %s will be deleted (UID removed from cluster_profile but still attached on cluster)", oldUID)
		profilesToDelete = append(profilesToDelete, oldUID)
	}

	return profilesToDelete
}

func updateProfiles(c *client.V1Client, d *schema.ResourceData) error {
	log.Printf("Updating cluster_profile (not cluster_template)")

	// Capture old cluster_profile to restore on error (pre-apply snapshot, or API sync when flag is on).
	oldProfileRaw, _ := d.GetChange("cluster_profile")
	oldProfile := normalizeInterfaceSliceFromListOrSet(oldProfileRaw)
	rollbackProfiles := func() {
		rollbackClusterProfileOnUpdateError(c, d, oldProfile)
	}

	profiles, err := toAddonDeplProfiles(c, d)
	var variableEntity []*models.V1SpectroClusterVariableUpdateEntity
	if err != nil {
		rollbackProfiles()
		return err
	}
	settings, err := toSpcApplySettings(d)
	if err != nil {
		rollbackProfiles()
		return err
	}

	// Get the current cluster state to find existing profile UIDs for replacement
	cluster, err := c.GetCluster(d.Id())
	if err != nil {
		return fmt.Errorf("failed to get cluster for profile update: %w", err)
	}

	// Handle profile deletions: find profiles that were in old state but not in new state
	// These need to be explicitly deleted since PATCH doesn't remove profiles
	if d.HasChange("cluster_profile") {
		profilesToDelete := getProfilesToDelete(c, d, cluster)
		if len(profilesToDelete) > 0 {
			log.Printf("Deleting %d profiles that were removed from cluster_profile", len(profilesToDelete))
			deleteBody := &models.V1SpectroClusterProfilesDeleteEntity{
				ProfileUids: profilesToDelete,
			}
			if err := c.DeleteAddonDeployment(d.Id(), deleteBody); err != nil {
				rollbackProfiles()
				return fmt.Errorf("failed to delete removed profiles: %w", err)
			}
		}
	}

	// If there are no profiles to add/update, we're done
	if len(profiles) == 0 {
		return nil
	}

	// Set ReplaceWithProfile for profiles that already exist on the cluster
	// This ensures PATCH updates existing profiles instead of adding duplicates
	if err := setReplaceWithProfileForExisting(c, cluster, profiles); err != nil {
		rollbackProfiles()
		return fmt.Errorf("failed to resolve profile replacements: %w", err)
	}

	body := &models.V1SpectroClusterProfiles{
		Profiles:         profiles,
		SpcApplySettings: settings,
	}
	clusterContext := d.Get("context").(string)
	// Use PATCH instead of PUT to preserve add-on profiles attached via spectrocloud_addon_deployment
	if err := c.PatchClusterProfileValues(d.Id(), body); err != nil {
		rollbackProfiles()
		return err
	}

	if _, found := toTags(d)["skip_apply"]; found {
		return nil
	}

	ctx := context.Background()
	if err := waitForProfileDownload(ctx, c, clusterContext, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
		rollbackProfiles()
		return err
	}

	// Profile Variable Handling - only for cluster_profile
	var newProfiles []interface{}
	if d.HasChange("cluster_profile") {
		_, newProfilesRaw := d.GetChange("cluster_profile")
		newProfiles = normalizeInterfaceSliceFromListOrSet(newProfilesRaw)
	}

	for _, newProfile := range newProfiles {
		if newProfile == nil {
			continue
		}

		p := newProfile.(map[string]interface{})

		// Skip profiles without an ID
		profileID, hasID := p["id"]
		if !hasID || profileID == nil || profileID.(string) == "" {
			log.Printf("Skipping profile without ID during variable update")
			continue
		}

		pVars := make([]*models.V1SpectroClusterVariable, 0)
		if pv, ok := p["variables"]; ok && pv != nil {
			variables := p["variables"].(map[string]interface{})
			for key, value := range variables {
				if key != "" && value != nil {
					pVars = append(pVars, &models.V1SpectroClusterVariable{
						Name:  StringPtr(key),
						Value: value.(string),
					})
				}
			}
		}

		// Only add to variableEntity if there are variables to update
		if len(pVars) != 0 {
			log.Printf("Updating variables for profile: %s with %d variables", profileID.(string), len(pVars))
			variableEntity = append(variableEntity, &models.V1SpectroClusterVariableUpdateEntity{
				ProfileUID: StringPtr(p["id"].(string)),
				Variables:  pVars,
			})
		}
	}
	// Patching cluster profiles Variables
	if len(variableEntity) != 0 {
		err = c.UpdateClusterProfileVariableInCluster(d.Id(), variableEntity)
		if err != nil {
			rollbackProfiles()
			return err
		}
	}

	return nil
}

func flattenClusterProfileForImport(c *client.V1Client, d *schema.ResourceData) ([]interface{}, error) {
	cluster, err := c.GetCluster(d.Id())
	if err != nil {
		return nil, err
	}
	clusterProfiles, err := flattenClusterProfilesFromCluster(cluster)
	if err != nil {
		return nil, err
	}
	clusterProfiles, err = enrichClusterProfilesWithVariables(c, d, d.Id(), clusterProfiles)
	if err != nil {
		return nil, err
	}
	return enrichClusterProfilesWithPacks(c, d, cluster, clusterProfiles)
}

// flattenClusterProfilesFromCluster maps every ClusterProfileTemplates entry from the API into
// cluster_profile state. Used for import and for read refresh when addon deployment is disabled.
func flattenClusterProfilesFromCluster(cluster *models.V1SpectroCluster) ([]interface{}, error) {
	clusterProfiles := make([]interface{}, 0)
	if cluster == nil || cluster.Spec == nil {
		return clusterProfiles, nil
	}

	for _, profileTemplate := range cluster.Spec.ClusterProfileTemplates {
		if profileTemplate == nil || profileTemplate.UID == "" {
			continue
		}
		clusterProfiles = append(clusterProfiles, map[string]interface{}{
			"id":   profileTemplate.UID,
			"pack": []interface{}{},
		})
	}
	return clusterProfiles, nil
}

// enrichClusterProfilesWithPacks attaches pack values from ClusterProfileTemplates when the profile
// has pack blocks in Terraform config (same pattern as spectrocloud_addon_deployment read).
func enrichClusterProfilesWithPacks(c *client.V1Client, d *schema.ResourceData, cluster *models.V1SpectroCluster, clusterProfiles []interface{}) ([]interface{}, error) {
	if c == nil || cluster == nil || cluster.Spec == nil || len(clusterProfiles) == 0 {
		return clusterProfiles, nil
	}

	templateByUID := make(map[string]*models.V1ClusterProfileTemplate, len(cluster.Spec.ClusterProfileTemplates))
	for _, template := range cluster.Spec.ClusterProfileTemplates {
		if template != nil && template.UID != "" {
			templateByUID[template.UID] = template
		}
	}

	registryNameMap := buildPackRegistryNameMapFromClusterProfiles(d)
	registryUIDMap := buildPackRegistryUIDMapFromClusterProfiles(d)

	for i := range clusterProfiles {
		p := clusterProfiles[i].(map[string]interface{})
		uid, ok := p["id"].(string)
		if !ok || uid == "" || !clusterProfileHasPacksInConfig(d, uid) {
			continue
		}

		template, ok := templateByUID[uid]
		if !ok || len(template.Packs) == 0 {
			continue
		}

		packManifests, packDiags, done := getPacksContent(template.Packs, c, d)
		if done {
			if len(packDiags) > 0 {
				return clusterProfiles, fmt.Errorf("%s: %s", packDiags[0].Summary, packDiags[0].Detail)
			}
			return clusterProfiles, errors.New("failed to read pack manifest content")
		}

		packs, err := flattenPacksWithRegistryMaps(c, nil, template.Packs, packManifests, registryNameMap, registryUIDMap)
		if err != nil {
			return clusterProfiles, err
		}
		configPacks := getClusterProfilePacksFromConfig(d, uid)
		for j := range packs {
			pack, ok := packs[j].(map[string]interface{})
			if !ok {
				continue
			}
			name, _ := pack["name"].(string)
			alignPackStateWithConfig(pack, findConfigPackByName(configPacks, name))
			packs[j] = pack
		}
		p["pack"] = packs
	}

	return clusterProfiles, nil
}

// enrichClusterProfilesWithVariables attaches cluster profile variables from the Palette API so
// refreshed state matches what Terraform config expects (avoids spurious cluster_profile changes).
func enrichClusterProfilesWithVariables(c *client.V1Client, d *schema.ResourceData, clusterUID string, clusterProfiles []interface{}) ([]interface{}, error) {
	if c == nil || len(clusterProfiles) == 0 {
		return clusterProfiles, nil
	}

	clusterVars, err := c.GetClusterVariables(clusterUID)
	if err != nil {
		return clusterProfiles, err
	}

	profileVariablesMap := make(map[string]map[string]interface{})
	for _, cv := range clusterVars {
		if cv.ProfileUID == nil || cv.Variables == nil {
			continue
		}
		vars := profileVariablesMapFromAPI(d, *cv.ProfileUID, cv.Variables)
		if len(vars) > 0 {
			profileVariablesMap[*cv.ProfileUID] = vars
		}
	}

	for i := range clusterProfiles {
		p := clusterProfiles[i].(map[string]interface{})
		uid, ok := p["id"].(string)
		if !ok || uid == "" {
			continue
		}
		if vars, has := profileVariablesMap[uid]; has {
			p["variables"] = vars
		} else if clusterProfileHasVariablesInConfig(d, uid) {
			p["variables"] = map[string]interface{}{}
		} else {
			delete(p, "variables")
		}
	}

	return clusterProfiles, nil
}

func shouldSyncClusterProfilesFromAPI(d *schema.ResourceData) bool {
	if !addonDeploymentResourceDisabled() {
		return false
	}
	if raw := d.Get("cluster_template"); raw != nil {
		if list, ok := raw.([]interface{}); ok && len(list) > 0 {
			return false
		}
	}
	return true
}

// setClusterProfilesFromAPI builds cluster_profile state from the cluster API document and writes it
// to ResourceData (flatten, variable/pack enrichment, align with config). Used on read and on
// update rollback when disable_addon_deployment_resource is true.
func setClusterProfilesFromAPI(c *client.V1Client, d *schema.ResourceData, cluster *models.V1SpectroCluster) error {
	if cluster == nil {
		return fmt.Errorf("cluster is required to sync cluster_profile from API")
	}

	clusterProfiles, err := flattenClusterProfilesFromCluster(cluster)
	if err != nil {
		return err
	}
	clusterProfiles, err = enrichClusterProfilesWithVariables(c, d, d.Id(), clusterProfiles)
	if err != nil {
		return err
	}
	clusterProfiles, err = enrichClusterProfilesWithPacks(c, d, cluster, clusterProfiles)
	if err != nil {
		return err
	}
	alignClusterProfilesStateWithConfig(d, clusterProfiles)
	return d.Set("cluster_profile", clusterProfiles)
}

// rollbackClusterProfileOnUpdateError restores cluster_profile after a failed updateProfiles.
// When shouldSyncClusterProfilesFromAPI is true, re-fetches the cluster and syncs from API so
// ResourceData reflects Palette (including partial applies). Otherwise restores the pre-apply snapshot.
func rollbackClusterProfileOnUpdateError(c *client.V1Client, d *schema.ResourceData, oldProfile []interface{}) {
	if shouldSyncClusterProfilesFromAPI(d) && c != nil && d.Id() != "" {
		refreshed, err := c.GetCluster(d.Id())
		if err != nil {
			log.Printf("Warning: could not refresh cluster for profile rollback from API: %v; restoring pre-apply cluster_profile", err)
			_ = d.Set("cluster_profile", oldProfile)
			return
		}
		if err := setClusterProfilesFromAPI(c, d, refreshed); err != nil {
			log.Printf("Warning: could not sync cluster_profile from API on rollback: %v; restoring pre-apply cluster_profile", err)
			_ = d.Set("cluster_profile", oldProfile)
		}
		return
	}
	_ = d.Set("cluster_profile", oldProfile)
}

// syncClusterProfilesFromAPIWhenAddonDeploymentDisabled refreshes cluster_profile from the API during
// read when disable_addon_deployment_resource is true (addon profiles are owned by the cluster resource).
func syncClusterProfilesFromAPIWhenAddonDeploymentDisabled(c *client.V1Client, d *schema.ResourceData, cluster *models.V1SpectroCluster) diag.Diagnostics {
	if !shouldSyncClusterProfilesFromAPI(d) {
		return nil
	}
	if err := setClusterProfilesFromAPI(c, d, cluster); err != nil {
		return diag.FromErr(err)
	}
	return nil
}

// toClusterTemplateReference extracts cluster template reference from ResourceData
// Returns nil if cluster_template is not specified
func toClusterTemplateReference(d *schema.ResourceData) *models.V1ClusterTemplateRef {
	clusterTemplateRaw := d.Get("cluster_template")
	if clusterTemplateRaw == nil {
		return nil
	}
	clusterTemplate := clusterTemplateRaw.([]interface{})
	if len(clusterTemplate) == 0 {
		return nil
	}

	templateData := clusterTemplate[0].(map[string]interface{})
	templateID := templateData["id"].(string)

	return &models.V1ClusterTemplateRef{
		UID: templateID,
	}
}

// updateClusterTemplateVariables handles variable updates for cluster_template using the variables API
// This is a separate flow from updateProfiles and only patches variables without triggering full cluster update
func updateClusterTemplateVariables(c *client.V1Client, d *schema.ResourceData) error {
	log.Printf("Updating cluster_template variables using variables API")

	_, newTemplateData := d.GetChange("cluster_template")
	if len(newTemplateData.([]interface{})) == 0 {
		return nil
	}

	// Extract profiles with variables from the new template data
	profiles, err := extractProfilesFromTemplateData(newTemplateData.([]interface{}))
	if err != nil {
		return err
	}

	// Build variable update entities
	variableEntity := make([]*models.V1SpectroClusterVariableUpdateEntity, 0)
	for _, profile := range profiles {
		if profile == nil {
			continue
		}

		p := profile.(map[string]interface{})
		profileID, hasID := p["id"]
		if !hasID || profileID == nil || profileID.(string) == "" {
			continue
		}

		// Extract variables
		pVars := make([]*models.V1SpectroClusterVariable, 0)
		if pv, ok := p["variables"]; ok && pv != nil {
			variables := p["variables"].(map[string]interface{})
			for key, value := range variables {
				if key != "" && value != nil {
					pVars = append(pVars, &models.V1SpectroClusterVariable{
						Name:  StringPtr(key),
						Value: value.(string),
					})
				}
			}
		}

		// Only add if there are variables to update
		if len(pVars) > 0 {
			log.Printf("Updating variables for profile: %s with %d variables", profileID.(string), len(pVars))
			variableEntity = append(variableEntity, &models.V1SpectroClusterVariableUpdateEntity{
				ProfileUID: StringPtr(profileID.(string)),
				Variables:  pVars,
			})
		}
	}

	// Patch variables using the variables API (not full cluster update)
	if len(variableEntity) > 0 {
		log.Printf("Patching %d profile variables using variables API", len(variableEntity))
		err = c.UpdateClusterProfileVariableInCluster(d.Id(), variableEntity)
		if err != nil {
			// Rollback on error
			oldTemplate, _ := d.GetChange("cluster_template")
			_ = d.Set("cluster_template", oldTemplate)
			return err
		}

		// Refresh variables from API after update
		log.Printf("Refreshing cluster_template variables after update")
		if err := flattenClusterTemplateVariables(c, d, d.Id()); err != nil {
			log.Printf("Warning: Failed to refresh variables after update: %v", err)
			// Don't fail the update if refresh fails
		}
	} else {
		log.Printf("No variables to update for cluster_template")
	}

	return nil
}

// flattenClusterTemplateVariables reads variables from the cluster and updates only the variables
// in the cluster_template state, keeping the profile IDs from config
func flattenClusterTemplateVariables(c *client.V1Client, d *schema.ResourceData, clusterUID string) error {
	// Only process if cluster_template is used
	clusterTemplateRaw := d.Get("cluster_template")
	if clusterTemplateRaw == nil {
		return nil
	}
	clusterTemplate := clusterTemplateRaw.([]interface{})
	if len(clusterTemplate) == 0 {
		return nil
	}

	// Get variables from cluster using the variables API
	clusterVars, err := c.GetClusterVariables(clusterUID)
	if err != nil {
		log.Printf("Error fetching cluster variables: %v", err)
		// Don't fail read if variables API fails, just skip variable updates
		return nil
	}

	// Build a map of profileUID -> variables
	profileVariablesMap := make(map[string]map[string]string)
	for _, clusterVar := range clusterVars {
		if clusterVar.ProfileUID != nil && clusterVar.Variables != nil {
			vars := profileVariablesMapFromAPI(d, *clusterVar.ProfileUID, clusterVar.Variables)
			stringVars := make(map[string]string, len(vars))
			for k, v := range vars {
				stringVars[k] = v.(string)
			}
			if len(stringVars) > 0 {
				profileVariablesMap[*clusterVar.ProfileUID] = stringVars
			}
		}
	}

	// Get configured profile IDs from current state
	templateData := clusterTemplate[0].(map[string]interface{})
	templateID := templateData["id"].(string)
	configuredProfileIDs := make(map[string]bool)

	// Build updated profile set with variables from API
	updatedProfileSet := schema.NewSet(schema.HashResource(&schema.Resource{
		Schema: map[string]*schema.Schema{
			"id": {
				Type: schema.TypeString,
			},
			"variables": {
				Type: schema.TypeMap,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}), []interface{}{})

	// Update only the profiles that were in config with latest variables from API
	if clusterProfiles, ok := templateData["cluster_profile"]; ok && clusterProfiles != nil {
		profilesSet := clusterProfiles.(*schema.Set)
		for _, profile := range profilesSet.List() {
			p := profile.(map[string]interface{})
			profileID := p["id"].(string)
			configuredProfileIDs[profileID] = true

			// Get configured variable names from config
			configuredVarNames := make(map[string]bool)
			if configVars, hasVars := p["variables"]; hasVars && configVars != nil {
				configVarsMap := configVars.(map[string]interface{})
				for varName := range configVarsMap {
					configuredVarNames[varName] = true
				}
			}

			// Create updated profile with variables from API
			updatedProfile := make(map[string]interface{})
			updatedProfile["id"] = profileID

			// Get variables from API response - only include variables that are in config
			if apiVars, ok := profileVariablesMap[profileID]; ok && len(apiVars) > 0 {
				// Convert map[string]string to map[string]interface{} for Set compatibility
				// Only include variables that were in the original config
				variablesInterface := make(map[string]interface{})
				for k, v := range apiVars {
					if configuredVarNames[k] {
						variablesInterface[k] = v
					}
				}
				if len(variablesInterface) > 0 {
					updatedProfile["variables"] = variablesInterface
				}
			}

			updatedProfileSet.Add(updatedProfile)
		}
	}

	log.Printf("flattenClusterTemplateVariables: updated %d profiles with variables (filtered to match config)", len(configuredProfileIDs))

	// Update cluster_template in state with refreshed variables
	updatedTemplate := []interface{}{
		map[string]interface{}{
			"id":              templateID,
			"cluster_profile": updatedProfileSet,
		},
	}

	return d.Set("cluster_template", updatedTemplate)
}
