package spectrocloud

import (
	"testing"

	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/stretchr/testify/assert"
)

func TestIsProfileAttachedToCluster(t *testing.T) {
	cluster := &models.V1SpectroCluster{
		Spec: &models.V1SpectroClusterSpec{
			ClusterProfileTemplates: []*models.V1ClusterProfileTemplate{
				{UID: "profile-old", Name: "monitoring"},
				{UID: "profile-new", Name: "monitoring"},
				{UID: "profile-base", Name: "base"},
			},
		},
	}

	assert.True(t, isProfileAttachedToCluster(cluster, "profile-old"))
	assert.True(t, isProfileAttachedToCluster(cluster, "profile-new"))
	assert.False(t, isProfileAttachedToCluster(cluster, "profile-stale-state-only"))
}

func TestGetProfilesToDeleteUIDNotSkippedForSameName(t *testing.T) {
	// Documents the AKS/UI corner case: old UID still on cluster, new UID in config, same profile name.
	oldUID := "6a0c2a821d2aa718e3836f1a"
	newUID := "6a0c2a579caf38df9cc3e290"

	newProfileUIDs := map[string]bool{
		"6a06b7162bc7f49b5b6140f3": true,
		newUID:                     true,
	}

	cluster := &models.V1SpectroCluster{
		Spec: &models.V1SpectroClusterSpec{
			ClusterProfileTemplates: []*models.V1ClusterProfileTemplate{
				{UID: oldUID, Name: "addon-profile"},
				{UID: newUID, Name: "addon-profile"},
				{UID: "6a06b7162bc7f49b5b6140f3", Name: "base"},
			},
		},
	}

	assert.False(t, newProfileUIDs[oldUID])
	assert.True(t, isProfileAttachedToCluster(cluster, oldUID))
	// Name-based logic treated this as a version upgrade and skipped API delete; UID-based delete includes oldUID.
}
