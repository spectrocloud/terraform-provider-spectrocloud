package spectrocloud

import (
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/spectrocloud/palette-sdk-go/client"

	"github.com/spectrocloud/terraform-provider-spectrocloud/spectrocloud/schemas"
	"github.com/spectrocloud/terraform-provider-spectrocloud/types"
)

// readAddonDeployment reads a single addon profile from the cluster
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
	profiles := normalizeInterfaceSliceFromListOrSet(d.Get("cluster_profile"))
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
	profiles := normalizeInterfaceSliceFromListOrSet(d.Get("cluster_profile"))
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
