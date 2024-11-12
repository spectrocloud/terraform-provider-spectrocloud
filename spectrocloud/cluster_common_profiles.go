package spectrocloud

import (
	"context"
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/spectrocloud/palette-sdk-go/client"

	"github.com/spectrocloud/terraform-provider-spectrocloud/util/ptr"
)

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
	profiles := d.Get("cluster_profile").([]interface{})
	if len(profiles) > 0 {
		for _, profile := range profiles {
			p := profile.(map[string]interface{})

			packValues := make([]*models.V1PackValuesEntity, 0)
			for _, pack := range p["pack"].([]interface{}) {
				p := toPack(cluster, pack)
				packValues = append(packValues, p)
			}
			resp = append(resp, &models.V1SpectroClusterProfileEntity{
				UID:        p["id"].(string),
				PackValues: packValues,
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
		Name: ptr.To(p["name"].(string)),
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
		pack.Type = models.V1PackType(val.(string))
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
				Name:    ptr.To(data["name"].(string)),
				Content: data["content"].(string),
				UID:     uid,
			}
		}
		pack.Manifests = manifests
	}
}

func updateProfiles(c *client.V1Client, d *schema.ResourceData) error {
	log.Printf("Updating profiles")
	profiles, err := toAddonDeplProfiles(c, d)
	if err != nil {
		oldProfile, _ := d.GetChange("cluster_profile")
		// When profile modification failed we are setting back the un applied version back to tf state
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
	if err := c.UpdateClusterProfileValues(d.Id(), body); err != nil {
		return err
	}

	if _, found := toTags(d)["skip_apply"]; found {
		return nil
	}

	ctx := context.Background()
	if err := waitForProfileDownload(ctx, c, clusterContext, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
		return err
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
