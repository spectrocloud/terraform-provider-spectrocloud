package spectrocloud

import (
	"context"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/spectrocloud/hapi/models"
	"github.com/spectrocloud/palette-sdk-go/client"

	"github.com/spectrocloud/terraform-provider-spectrocloud/types"
)

func toProfiles(c *client.V1Client, d *schema.ResourceData) []*models.V1SpectroClusterProfileEntity {
	var cluster *models.V1SpectroCluster
	var err error
	clusterContext := d.Get("context").(string)

	if d.Id() != "" {
		cluster, err = c.GetClusterWithoutStatus(clusterContext, d.Id())
		if err != nil {
			return nil
		}
	}
	return toProfilesCommon(d, cluster)
}

func toProfilesCommon(d *schema.ResourceData, cluster *models.V1SpectroCluster) []*models.V1SpectroClusterProfileEntity {
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

	return resp
}

func toSpcApplySettings(d *schema.ResourceData) *models.V1SpcApplySettings {
	if d.Get("apply_setting") != nil {
		setting := d.Get("apply_setting").(string)
		if setting != "" {
			return &models.V1SpcApplySettings{
				ActionType: setting,
			}
		}
	}

	return nil
}

func toPack(cluster *models.V1SpectroCluster, pSrc interface{}) *models.V1PackValuesEntity {
	p := pSrc.(map[string]interface{})

	pack := &models.V1PackValuesEntity{
		Name: types.Ptr(p["name"].(string)),
	}

	if val, found := p["values"]; found && len(val.(string)) > 0 {
		pack.Values = val.(string)
	}
	if val, found := p["tag"]; found && len(val.(string)) > 0 {
		pack.Tag = val.(string)
	}
	if val, found := p["type"]; found && len(val.(string)) > 0 {
		pack.Type = models.V1PackType(val.(string))
	}
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

	return pack
}

func updateProfiles(c *client.V1Client, d *schema.ResourceData) error {
	log.Printf("Updating profiles")
	body := &models.V1SpectroClusterProfiles{
		Profiles:         toProfiles(c, d),
		SpcApplySettings: toSpcApplySettings(d),
	}
	if err := c.UpdateClusterProfileValues(d.Id(), body); err != nil {
		return err
	}

	if _, found := toTags(d)["skip_apply"]; found {
		return nil
	}

	ctx := context.Background()
	clusterContext := d.Get("context").(string)
	if err := waitForProfileDownload(ctx, c, clusterContext, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
		return err
	}

	return nil
}
