package spectrocloud

import (
	"context"
	"github.com/spectrocloud/terraform-provider-spectrocloud/types"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/hapi/models"
	"github.com/spectrocloud/terraform-provider-spectrocloud/pkg/client"
)

func toProfiles(c *client.V1Client, d *schema.ResourceData) []*models.V1SpectroClusterProfileEntity {
	var cluster *models.V1SpectroCluster
	var err error
	if d.Id() != "" {
		cluster, err = c.GetClusterWithoutStatus(d.Id())
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
	} else {
		clusterProfileId := d.Get("cluster_profile_id")

		// consider removing backward compatibility for cluster_profile_id
		if clusterProfileId != nil {
			packValues := make([]*models.V1PackValuesEntity, 0)
			for _, pack := range d.Get("pack").([]interface{}) {
				p := toPack(cluster, pack)
				packValues = append(packValues, p)
			}
			resp = append(resp, &models.V1SpectroClusterProfileEntity{
				UID:        clusterProfileId.(string),
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
					for _, pack := range profile.Packs {
						packs = append(packs, pack)
					}
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
	if err := waitForProfileDownload(ctx, c, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
		return err
	}

	return nil
}
