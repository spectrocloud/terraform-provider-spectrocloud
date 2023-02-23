package spectrocloud

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/hapi/models"
	"github.com/spectrocloud/palette-sdk-go/client"
)

func flattenPacks(c *client.V1Client, diagPacks []*models.V1PackManifestEntity, packs []*models.V1PackRef, manifestContent map[string]map[string]string) ([]interface{}, error) {
	if packs == nil {
		return make([]interface{}, 0), nil
	}

	ps := make([]interface{}, len(packs))
	for i, pack := range packs {
		p := make(map[string]interface{})

		p["uid"] = pack.PackUID
		if isRegistryUID(diagPacks, *pack.Name) {
			p["registry_uid"] = c.GetPackRegistry(pack.PackUID, pack.Type)
		}
		p["name"] = *pack.Name
		p["tag"] = pack.Tag
		p["values"] = pack.Values
		p["type"] = pack.Type

		if _, ok := manifestContent[pack.PackUID]; ok {
			ma := make([]interface{}, len(pack.Manifests))
			for j, m := range pack.Manifests {
				mj := make(map[string]interface{})
				mj["name"] = m.Name
				mj["uid"] = m.UID
				mj["content"] = manifestContent[pack.PackUID][m.Name]

				ma[j] = mj
			}

			p["manifest"] = ma
		}
		ps[i] = p
	}

	return ps, nil
}

func isRegistryUID(diagPacks []*models.V1PackManifestEntity, name string) bool {
	for _, pack := range diagPacks {
		if *pack.Name == name && pack.RegistryUID != "" {
			return true
		}
	}
	return false
}

func getPacksContent(packs []*models.V1PackRef, c *client.V1Client, d *schema.ResourceData) (map[string]map[string]string, diag.Diagnostics, bool) {
	packManifests := make(map[string]map[string]string)
	for _, p := range packs {
		if len(p.Manifests) > 0 {
			content, err := c.GetClusterProfileManifestPack(d.Id(), *p.Name)
			if err != nil {
				return nil, diag.FromErr(err), true
			}

			if len(content) > 0 {
				c := make(map[string]string)
				for _, co := range content {
					c[co.Metadata.Name] = co.Spec.Published.Content
				}
				packManifests[p.PackUID] = c
			}
		}
	}
	return packManifests, nil, false
}
