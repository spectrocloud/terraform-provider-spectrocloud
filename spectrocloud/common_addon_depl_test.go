package spectrocloud

import (
	"errors"
	"testing"

	"github.com/spectrocloud/palette-api-go/models"
	"github.com/spectrocloud/palette-sdk-go/client"
	"github.com/stretchr/testify/assert"

	"github.com/spectrocloud/terraform-provider-spectrocloud/types"
)

func TestToAddonDeployment(t *testing.T) {
	assert := assert.New(t)

	// Create a mock ResourceData object
	d := prepareAddonDeploymentTestData("depl-test-id")

	m := &client.V1Client{
		GetClusterWithoutStatusFn: func(uid string) (*models.V1SpectroCluster, error) {
			if uid != "cluster-123" {
				return nil, errors.New("unexpected cluster_uid")
			}
			return &models.V1SpectroCluster{
				Metadata: nil,
				Spec: &models.V1SpectroClusterSpec{
					ClusterProfileTemplates: []*models.V1ClusterProfileTemplate{
						{
							Packs: []*models.V1PackRef{
								{
									Name: types.Ptr("pack1"),
									Manifests: []*models.V1ObjectReference{
										{Name: "pack1", UID: "uid1"},
									},
								},
								{
									Name: types.Ptr("pack2"),
									Manifests: []*models.V1ObjectReference{
										{Name: "pack2", UID: "uid2"},
									},
								},
							},
						},
						{
							Packs: []*models.V1PackRef{
								{
									Name: types.Ptr("pack3"),
									Manifests: []*models.V1ObjectReference{
										{Name: "pack3", UID: "uid3"},
									},
								},
								{
									Name: types.Ptr("pack4"),
									Manifests: []*models.V1ObjectReference{
										{Name: "pack4", UID: "uid4"},
									},
								},
							},
						},
					},
				},
				Status: &models.V1SpectroClusterStatus{
					State: "Deleted",
				},
			}, nil
		},
	}

	addonDeployment, err := toAddonDeployment(m, d)
	assert.Nil(err)

	// Verifying apply setting
	assert.Equal(d.Get("apply_setting"), addonDeployment.SpcApplySettings.ActionType)

	// Verifying cluster profile
	profiles := d.Get("cluster_profile").([]interface{})
	for i, profile := range profiles {
		p := profile.(map[string]interface{})
		assert.Equal(p["id"].(string), addonDeployment.Profiles[i].UID)

		// Verifying pack values
		packValues := p["pack"].([]interface{})
		for j, pack := range packValues {
			packMap := pack.(map[string]interface{})
			assert.Equal(packMap["name"], *addonDeployment.Profiles[i].PackValues[j].Name)
			assert.Equal(packMap["tag"], addonDeployment.Profiles[i].PackValues[j].Tag)
			assert.Equal(packMap["type"], string(addonDeployment.Profiles[i].PackValues[j].Type))
			assert.Equal(packMap["values"], addonDeployment.Profiles[i].PackValues[j].Values)
		}

	}
}
