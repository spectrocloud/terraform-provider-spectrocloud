package spectrocloud

import (
	"github.com/spectrocloud/hapi/models"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetEdgeHostDeviceUID(t *testing.T) {
	edgeDevice := &models.V1EdgeHostDevice{
		Metadata: &models.V1ObjectMeta{
			UID: "uid",
		},
	}

	assert.Equal(t, "uid", getEdgeHostDeviceUID(edgeDevice))
}
