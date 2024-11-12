package spectrocloud

import (
	"testing"

	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/stretchr/testify/assert"

	"github.com/spectrocloud/terraform-provider-spectrocloud/util/ptr"
)

func TestToPack_PacksMerging(t *testing.T) {
	cluster := &models.V1SpectroCluster{
		Spec: &models.V1SpectroClusterSpec{
			ClusterProfileTemplates: []*models.V1ClusterProfileTemplate{
				{
					Packs: []*models.V1PackRef{
						{
							Name: ptr.To("pack1"),
							Manifests: []*models.V1ObjectReference{
								{Name: "pack1", UID: "uid1"},
							},
						},
						{
							Name: ptr.To("pack2"),
							Manifests: []*models.V1ObjectReference{
								{Name: "pack2", UID: "uid2"},
							},
						},
					},
				},
				{
					Packs: []*models.V1PackRef{
						{
							Name: ptr.To("pack3"),
							Manifests: []*models.V1ObjectReference{
								{Name: "pack3", UID: "uid3"},
							},
						},
						{
							Name: ptr.To("pack4"),
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
		Name:   ptr.To("testPack"),
		Values: "someValues",
		Tag:    "v1",
		Type:   models.V1PackType("testType"),
		Manifests: []*models.V1ManifestRefUpdateEntity{
			{
				Name:    ptr.To("pack1"),
				Content: "content1",
				UID:     "uid1",
			},
			{
				Name:    ptr.To("pack2"),
				Content: "content2",
				UID:     "uid2",
			},
		},
	}

	result := toPack(cluster, pSrc)
	assert.Equal(t, expectedPack, result, "The packs should be equal")
}
