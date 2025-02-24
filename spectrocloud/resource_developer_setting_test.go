package spectrocloud

import (
	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestToDeveloperSetting(t *testing.T) {
	d := resourceDeveloperSetting().TestResourceData()

	// Set custom values
	d.Set("virtual_clusters_limit", int32(10))
	d.Set("cpu", int32(4))
	d.Set("memory", int32(16))
	d.Set("storage", int32(50))

	devCredit, sysClusterGroupPref := toDeveloperSetting(d)

	assert.NotNil(t, devCredit)
	assert.NotNil(t, sysClusterGroupPref)
	assert.Equal(t, int32(10), devCredit.VirtualClustersLimit)
	assert.Equal(t, int32(4), devCredit.CPU)
	assert.Equal(t, int32(16), devCredit.MemoryGiB)
	assert.Equal(t, int32(50), devCredit.StorageGiB)
	assert.False(t, sysClusterGroupPref.HideSystemClusterGroups)
}

func TestToDeveloperSettingDefault(t *testing.T) {
	d := resourceDeveloperSetting().TestResourceData()

	devCredit, sysClusterGroupPref := toDeveloperSettingDefault(d)

	assert.NotNil(t, devCredit)
	assert.NotNil(t, sysClusterGroupPref)
	assert.Equal(t, int32(12), devCredit.CPU)
	assert.Equal(t, int32(16), devCredit.MemoryGiB)
	assert.Equal(t, int32(20), devCredit.StorageGiB)
	assert.Equal(t, int32(2), devCredit.VirtualClustersLimit)
	assert.False(t, sysClusterGroupPref.HideSystemClusterGroups)
}

func TestFlattenDeveloperSetting(t *testing.T) {
	d := resourceDeveloperSetting().TestResourceData()

	devSetting := &models.V1DeveloperCredit{
		CPU:                  8,
		MemoryGiB:            32,
		StorageGiB:           100,
		VirtualClustersLimit: 5,
	}
	sysClusterGroupPref := &models.V1TenantEnableClusterGroup{
		HideSystemClusterGroups: true,
	}

	err := flattenDeveloperSetting(devSetting, sysClusterGroupPref, d)
	assert.NoError(t, err)

	// Verify values set in schema
	assert.Equal(t, 8, d.Get("cpu"))
	assert.Equal(t, 32, d.Get("memory"))
	assert.Equal(t, 100, d.Get("storage"))
	assert.Equal(t, 5, d.Get("virtual_clusters_limit"))
	assert.True(t, d.Get("hide_system_cluster_group").(bool))
}
