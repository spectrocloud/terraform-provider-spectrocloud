package addon_deployment_test

import (
	"fmt"
	"testing"

	"github.com/spectrocloud/hapi/models"
	clusterC "github.com/spectrocloud/hapi/spectrocluster/client/v1"
	"github.com/spectrocloud/palette-sdk-go/client"
	"github.com/stretchr/testify/assert"
)

func TestPatchWithRetry(t *testing.T) {
	// Create a mock V1Client
	var patchCalled int

	// Create a mock for client.V1SpectroClustersPatchProfiles(params)
	h := &client.V1Client{
		RetryAttempts: 3,
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
	err := client.PatchWithRetry(h, params)

	// Assert patch was called 3 times and there was no error
	assert.Equal(t, 3, patchCalled)
	assert.NoError(t, err)
}
