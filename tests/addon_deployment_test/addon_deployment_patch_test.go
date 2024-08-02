package addon_deployment_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/spectrocloud/terraform-provider-spectrocloud/tests/mock"
)

func TestPatchWithRetry(t *testing.T) {
	// Create a cluster client mock
	// h := client.V1Client{}
	mock := &mock.ClusterClientMock{
		PatchSPCProfilesErr: errors.New("test error"),
	}

	// Call patchWithRetry
	//err := h.PatchWithRetry(mock, params)

	// Assert patch was called 3 times and there was no error
	assert.Equal(t, 3, mock.PatchSPCProfilesCount)
	//assert.NoError(t, err)
}
