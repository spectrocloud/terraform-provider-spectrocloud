package spectrocloud

import (
	"github.com/spectrocloud/palette-api-go/models"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetAddonDeploymentIdANDReverse(t *testing.T) {
	clusterId := "5eea74ed19"
	clusterProfileId := "0d445deb3ca"
	addonDeploymentId := clusterId + "_" + clusterProfileId

	testAddonDeploymentId := getAddonDeploymentId(clusterId, &models.V1ClusterProfile{Metadata: &models.V1ObjectMeta{UID: clusterProfileId}})
	if testAddonDeploymentId != addonDeploymentId {
		t.Errorf("got %s, wanted %s", testAddonDeploymentId, addonDeploymentId)
	}

	testClusterId := getClusterUID(testAddonDeploymentId)
	if testClusterId != clusterId {
		t.Errorf("got %s, wanted %s", testClusterId, clusterId)
	}

	testClusterProfileId, _ := getClusterProfileUID(testAddonDeploymentId)
	if testClusterProfileId != clusterProfileId {
		t.Errorf("got %s, wanted %s", testClusterProfileId, clusterProfileId)
	}
}

func TestIsProfileAttached(t *testing.T) {
	tests := []struct {
		name     string
		cluster  *models.V1SpectroCluster
		uid      string
		expected bool
	}{
		{
			name: "Profile Attached",
			cluster: &models.V1SpectroCluster{
				Spec: &models.V1SpectroClusterSpec{
					ClusterProfileTemplates: []*models.V1ClusterProfileTemplate{
						{UID: "profile-123"},
						{UID: "profile-456"},
					},
				},
			},
			uid:      "profile-123",
			expected: true,
		},
		{
			name: "Profile Not Attached",
			cluster: &models.V1SpectroCluster{
				Spec: &models.V1SpectroClusterSpec{
					ClusterProfileTemplates: []*models.V1ClusterProfileTemplate{
						{UID: "profile-123"},
						{UID: "profile-456"},
					},
				},
			},
			uid:      "profile-789",
			expected: false,
		},
		{
			name: "Empty Profile List",
			cluster: &models.V1SpectroCluster{
				Spec: &models.V1SpectroClusterSpec{
					ClusterProfileTemplates: []*models.V1ClusterProfileTemplate{},
				},
			},
			uid:      "profile-123",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := isProfileAttached(tt.cluster, tt.uid)
			assert.Equal(t, tt.expected, output)
		})
	}
}
