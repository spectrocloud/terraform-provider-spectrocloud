package addon_deployment

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/spectrocloud/hapi/models"
	"github.com/spectrocloud/palette-sdk-go/client"
	"github.com/spectrocloud/terraform-provider-spectrocloud/tests/mock"
)

func TestUpdateAddonDeploymentIsNotAttached(t *testing.T) {
	h := client.V1Client{}
	mock := &mock.ClusterClientMock{
		PatchClusterProfileErr: nil,
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
	err := h.UpdateAddonDeployment(mock, cluster, body, newProfile)

	// Assert there was no error
	assert.NoError(t, err)
}

func TestUpdateAddonDeploymentIsAttached(t *testing.T) {
	h := client.V1Client{}
	mock := &mock.ClusterClientMock{
		PatchClusterProfileErr: nil,
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
	err := h.UpdateAddonDeployment(mock, cluster, body, newProfile)

	// Assert there was no error
	assert.NoError(t, err)
}
