package addon_deployment_test

import (
	"testing"

	"github.com/spectrocloud/hapi/client"
	"github.com/spectrocloud/hapi/models"
	"github.com/stretchr/testify/assert"
)

func TestIsProfileAttachedByNamePositive(t *testing.T) {
	// Test where profile is attached
	cluster := &models.V1SpectroCluster{
		Spec: &models.V1SpectroClusterSpec{
			ClusterProfileTemplates: []*models.V1ClusterProfileTemplate{
				{
					UID:  "test-uid",
					Name: "test-name",
				},
			},
		},
	}
	newProfile := &models.V1ClusterProfile{
		Metadata: &models.V1ObjectMeta{
			Name: "test-name",
		},
	}
	isAttached, uid := client.IsProfileAttachedByName(cluster, newProfile)
	assert.True(t, isAttached)
	assert.Equal(t, "test-uid", uid)
}

func TestIsProfileAttachedByNameNegative(t *testing.T) {
	// Test where profile is not attached
	cluster := &models.V1SpectroCluster{
		Spec: &models.V1SpectroClusterSpec{
			ClusterProfileTemplates: []*models.V1ClusterProfileTemplate{
				{
					UID:  "test-uid",
					Name: "test-name",
				},
			},
		},
	}
	newProfile := &models.V1ClusterProfile{
		Metadata: &models.V1ObjectMeta{
			Name: "other-test-name",
		},
	}
	isAttached, uid := client.IsProfileAttachedByName(cluster, newProfile)
	assert.False(t, isAttached)
	assert.Equal(t, "", uid)
}
