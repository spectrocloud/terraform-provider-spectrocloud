package addon_deployment_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	clusterC "github.com/spectrocloud/hapi/spectrocluster/client/v1"
	"github.com/spectrocloud/palette-api-go/models"
	"github.com/spectrocloud/palette-sdk-go/client"
	"github.com/spectrocloud/terraform-provider-spectrocloud/tests/mock"
)

func TestPatchWithRetry(t *testing.T) {
	// Create a cluster client mock
	h := client.V1Client{
		RetryAttempts: 3,
	}
	mock := &mock.ClusterClientMock{
		PatchSPCProfilesErr: errors.New("test error"),
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
	err := h.PatchWithRetry(mock, params)

	// Assert patch was called 3 times and there was no error
	assert.Equal(t, 3, mock.PatchSPCProfilesCount)
	assert.NoError(t, err)
}
