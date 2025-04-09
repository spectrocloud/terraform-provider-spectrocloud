package spectrocloud

import (
	"testing"

	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/stretchr/testify/assert"

	"github.com/spectrocloud/terraform-provider-spectrocloud/types"
)

func TestToPack_PacksMerging(t *testing.T) {
	cluster := &models.V1SpectroCluster{
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
	}

	pSrc := map[string]interface{}{
		"name":   "testPack",
		"values": "someValues",
		"tag":    "v1",
		"type":   "testType",
		"manifest": []interface{}{
			map[string]interface{}{
				"name":    "pack1",
				"content": "content1",
			},
			map[string]interface{}{
				"name":    "pack2",
				"content": "content2",
			},
		},
	}

	expectedPack := &models.V1PackValuesEntity{
		Name:   types.Ptr("testPack"),
		Values: "someValues",
		Tag:    "v1",
		Type:   models.V1PackTypeOci.Pointer(),
		Manifests: []*models.V1ManifestRefUpdateEntity{
			{
				Name:    types.Ptr("pack1"),
				Content: "content1",
				UID:     "uid1",
			},
			{
				Name:    types.Ptr("pack2"),
				Content: "content2",
				UID:     "uid2",
			},
		},
	}

	result := toPack(cluster, pSrc)
	assert.Equal(t, expectedPack, result, "The packs should be equal")
}
