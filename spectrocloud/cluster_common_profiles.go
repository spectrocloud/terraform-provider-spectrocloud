package spectrocloud

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/spectrocloud/palette-sdk-go/client"

	"github.com/spectrocloud/terraform-provider-spectrocloud/types"
)

// validateProfileSource checks that only one of cluster_template or cluster_profile is specified
func validateProfileSource(d *schema.ResourceData) error {
	// cluster_template may not exist in all schemas (e.g., cluster_group)
	clusterTemplateRaw := d.Get("cluster_template")
	clusterProfileRaw := d.Get("cluster_profile")

	var clusterTemplate, clusterProfile []interface{}
	if clusterTemplateRaw != nil {
		clusterTemplate = clusterTemplateRaw.([]interface{})
	}
	if clusterProfileRaw != nil {
		clusterProfile = clusterProfileRaw.([]interface{})
	}

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

	clusterTemplateRaw := d.Get("cluster_template")
	clusterProfileRaw := d.Get("cluster_profile")

	var clusterTemplate, clusterProfile []interface{}
	if clusterTemplateRaw != nil {
		clusterTemplate = clusterTemplateRaw.([]interface{})
	}
	if clusterProfileRaw != nil {
		clusterProfile = clusterProfileRaw.([]interface{})
	}

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
		log.Printf("Using profiles from cluster_profile")
		return clusterProfile, "cluster_profile", nil
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

func updateProfiles(c *client.V1Client, d *schema.ResourceData) error {
	log.Printf("Updating cluster_profile (not cluster_template)")
	profiles, err := toAddonDeplProfiles(c, d)
	var variableEntity []*models.V1SpectroClusterVariableUpdateEntity
	if err != nil {
		// Restore old value on error
		oldProfile, _ := d.GetChange("cluster_profile")
		_ = d.Set("cluster_profile", oldProfile)
		return err
	}
	settings, err := toSpcApplySettings(d)
	if err != nil {
		return err
	}
	body := &models.V1SpectroClusterProfiles{
		Profiles:         profiles,
		SpcApplySettings: settings,
	}
	clusterContext := d.Get("context").(string)
	// Use PATCH instead of PUT to preserve add-on profiles attached via spectrocloud_addon_deployment
	if err := c.PatchClusterProfileValues(d.Id(), body); err != nil {
		return err
	}

	if _, found := toTags(d)["skip_apply"]; found {
		return nil
	}

	ctx := context.Background()
	if err := waitForProfileDownload(ctx, c, clusterContext, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
		return err
	}

	// Profile Variable Handling - only for cluster_profile
	var newProfiles []interface{}
	if d.HasChange("cluster_profile") {
		_, newProfilesRaw := d.GetChange("cluster_profile")
		newProfiles = newProfilesRaw.([]interface{})
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
			return err
		}
	}

	return nil
}

func flattenClusterProfileForImport(c *client.V1Client, d *schema.ResourceData) ([]interface{}, error) {
	//clusterContext := "project"
	//if v, ok := d.GetOk("context"); ok {
	//	clusterContext = v.(string)
	//}
	clusterProfiles := make([]interface{}, 0)
	cluster, err := c.GetCluster(d.Id())
	if err != nil {
		return clusterProfiles, err
	}
	for _, profileTemplate := range cluster.Spec.ClusterProfileTemplates {
		profile := make(map[string]interface{})
		profile["id"] = profileTemplate.UID
		clusterProfiles = append(clusterProfiles, profile)
	}
	return clusterProfiles, nil
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
			vars := make(map[string]string)
			for _, v := range clusterVar.Variables {
				if v.Name != nil && v.Value != "" {
					vars[*v.Name] = v.Value
				}
			}
			if len(vars) > 0 {
				profileVariablesMap[*clusterVar.ProfileUID] = vars
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
