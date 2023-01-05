package client

import (
	"fmt"
	"github.com/spectrocloud/hapi/models"
	clusterC "github.com/spectrocloud/hapi/spectrocluster/client/v1"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestUpdateAddonDeploymentIsNotAttached(t *testing.T) {
	// Create a mock V1Client
	h := &V1Client{
		ClustersPatchProfilesFn: func(params *clusterC.V1SpectroClustersPatchProfilesParams) error {
			// Check that the correct params are passed to ClustersPatchProfiles
			assert.Equal(t, "test-cluster", params.UID)
			assert.Equal(t, "test-profile", params.Body.Profiles[0].UID)
			assert.Equal(t, "test-profile-to-replace", params.Body.Profiles[0].ReplaceWithProfile)
			assert.True(t, *params.ResolveNotification)
			return nil
		},
	}

	// Create mock cluster
	cluster := &models.V1SpectroCluster{
		Metadata: &models.V1ObjectMeta{
			UID: "test-cluster",
		},
		Spec: &models.V1SpectroClusterSpec{
			ClusterProfileTemplates: []*models.V1ClusterProfileTemplate{
				{
					UID:  "test-profile-uid",
					Name: "test-profile-name",
				},
			},
		},
	}

	// Create mock body
	body := &models.V1SpectroClusterProfiles{
		Profiles: []*models.V1SpectroClusterProfileEntity{
			{UID: "test-profile"},
		},
	}

	// Create mock newProfile
	newProfile := &models.V1ClusterProfile{
		Metadata: &models.V1ObjectMeta{
			UID: "new-test-profile-uid",
		},
	}

	// Call UpdateAddonDeployment
	err := h.UpdateAddonDeployment(cluster, body, newProfile)

	// Assert there was no error
	assert.NoError(t, err)
}

func TestUpdateAddonDeploymentIsAttached(t *testing.T) {
	// Create a mock V1Client
	h := &V1Client{
		ClustersPatchProfilesFn: func(params *clusterC.V1SpectroClustersPatchProfilesParams) error {
			// Check that the correct params are passed to ClustersPatchProfiles
			assert.Equal(t, "test-cluster", params.UID)
			assert.Equal(t, "test-profile", params.Body.Profiles[0].UID)
			assert.Equal(t, "test-profile-uid", params.Body.Profiles[0].ReplaceWithProfile)
			assert.True(t, *params.ResolveNotification)
			return nil
		},
	}

	// Create mock cluster
	cluster := &models.V1SpectroCluster{
		Metadata: &models.V1ObjectMeta{
			UID: "test-cluster",
		},
		Spec: &models.V1SpectroClusterSpec{
			ClusterProfileTemplates: []*models.V1ClusterProfileTemplate{
				{
					UID:  "test-profile-uid",
					Name: "test-profile-name",
				},
			},
		},
	}

	// Create mock body
	body := &models.V1SpectroClusterProfiles{
		Profiles: []*models.V1SpectroClusterProfileEntity{
			{UID: "test-profile"},
		},
	}

	// Create mock newProfile
	newProfile := &models.V1ClusterProfile{
		Metadata: &models.V1ObjectMeta{
			Name: "test-profile-name",
		},
	}

	// Call UpdateAddonDeployment
	err := h.UpdateAddonDeployment(cluster, body, newProfile)

	// Assert there was no error
	assert.NoError(t, err)
}

func TestPatchWithRetry(t *testing.T) {
	// Create a mock V1Client
	var patchCalled int

	// Create a mock for client.V1SpectroClustersPatchProfiles(params)
	h := &V1Client{
		retryAttempts: 3,
		ClustersPatchProfilesFn: func(params *clusterC.V1SpectroClustersPatchProfilesParams) error {
			patchCalled++
			if patchCalled < 3 {
				return fmt.Errorf("test error")
			}
			return nil
		},
	}

	// Create mock params
	params := &clusterC.V1SpectroClustersPatchProfilesParams{
		UID: "test-cluster",
		Body: &models.V1SpectroClusterProfiles{
			Profiles: []*models.V1SpectroClusterProfileEntity{
				{UID: "test-profile"},
			},
		},
	}

	// Call patchWithRetry
	err := patchWithRetry(h, params)

	// Assert patch was called 3 times and there was no error
	assert.Equal(t, 3, patchCalled)
	assert.NoError(t, err)
}

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
	isAttached, uid := isProfileAttachedByName(cluster, newProfile)
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
	isAttached, uid := isProfileAttachedByName(cluster, newProfile)
	assert.False(t, isAttached)
	assert.Equal(t, "", uid)
}
