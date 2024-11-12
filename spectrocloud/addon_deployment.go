package spectrocloud

import (
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/spectrocloud/palette-sdk-go/client"

	"github.com/spectrocloud/terraform-provider-spectrocloud/util/ptr"
)

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

	packManifests, d2, done2 := getPacksContent(profile.Packs, c, d)
	if done2 {
		return d2, false
	}

	diagPacks, diagnostics, done := GetAddonDeploymentDiagPacks(d, nil)
	if done {
		return diagnostics, false
	}
	packs, err := flattenPacks(c, diagPacks, profile.Packs, packManifests)
	if err != nil {
		return diag.FromErr(err), false
	}

	cluster_profiles := make([]interface{}, 0)
	cluster_profile := make(map[string]interface{})
	cluster_profile["pack"] = packs
	cluster_profile["id"] = profile.UID
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

	pack := &models.V1PackManifestEntity{
		Name:        ptr.To(pName),
		Tag:         pTag,
		RegistryUID: pRegistryUID,
		Type:        pType,
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
