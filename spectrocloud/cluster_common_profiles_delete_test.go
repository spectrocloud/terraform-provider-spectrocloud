package spectrocloud

import (
	"testing"

	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func TestIsInfraClusterProfileType(t *testing.T) {
	assert.True(t, isInfraClusterProfileType("infra"))
	assert.True(t, isInfraClusterProfileType("cluster"))
	assert.False(t, isInfraClusterProfileType("addon"))
	assert.False(t, isInfraClusterProfileType(""))
}

func TestFindAttachedInfraProfile(t *testing.T) {
	cluster := &models.V1SpectroCluster{
		Spec: &models.V1SpectroClusterSpec{
			ClusterProfileTemplates: []*models.V1ClusterProfileTemplate{
				{UID: "addon-1", Type: "addon", Name: "monitoring"},
				{UID: "old-infra", Type: "cluster", Name: "eks-infra-old"},
			},
		},
	}
	assert.Equal(t, "old-infra", findAttachedInfraProfile(cluster))
}

func TestGetProfilesToDeleteSkipsClusterTypeInfra(t *testing.T) {
	oldInfraUID := "6a0de2794ffb1a5d43004cb3"
	newInfraUID := "6a0de28515e92045e165656c"
	addonUID := "6a0de2000000000000000001"

	cluster := &models.V1SpectroCluster{
		Spec: &models.V1SpectroClusterSpec{
			ClusterProfileTemplates: []*models.V1ClusterProfileTemplate{
				{UID: oldInfraUID, Type: "cluster", Name: "eks-infra-old"},
				{UID: addonUID, Type: "addon", Name: "addon-profile"},
			},
		},
	}

	r := resourceClusterEks()
	d := r.TestResourceData()
	_ = d.Set("cluster_profile", []interface{}{
		map[string]interface{}{"id": newInfraUID},
		map[string]interface{}{"id": addonUID},
	})

	// Simulate state change: old infra + addon -> new infra + addon
	require.NoError(t, d.Set("cluster_profile", []interface{}{
		map[string]interface{}{"id": oldInfraUID, "variables": map[string]interface{}{}},
		map[string]interface{}{"id": addonUID},
	}))
	require.NoError(t, d.Set("cluster_profile", []interface{}{
		map[string]interface{}{"id": newInfraUID},
		map[string]interface{}{"id": addonUID},
	}))

	toDelete := getProfilesToDelete(nil, d, cluster)
	assert.NotContains(t, toDelete, oldInfraUID, "infra/cluster profile must not be deleted; use PATCH replace")
}

func TestGetProfilesToDeleteUIDNotSkippedForSameNameAddon(t *testing.T) {
	oldUID := "6a0c2a821d2aa718e3836f1a"
	newUID := "6a0c2a579caf38df9cc3e290"

	cluster := &models.V1SpectroCluster{
		Spec: &models.V1SpectroClusterSpec{
			ClusterProfileTemplates: []*models.V1ClusterProfileTemplate{
				{UID: oldUID, Name: "addon-profile", Type: "addon"},
				{UID: newUID, Name: "addon-profile", Type: "addon"},
				{UID: "6a06b7162bc7f49b5b6140f3", Name: "base", Type: "cluster"},
			},
		},
	}

	assert.False(t, isInfraClusterProfileType(getAttachedProfileType(cluster, oldUID)))
	assert.True(t, isProfileAttachedToCluster(cluster, oldUID))
}
