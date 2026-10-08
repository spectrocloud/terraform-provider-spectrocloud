package spectrocloud

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/spectrocloud/terraform-provider-spectrocloud/tests/mockApiServer/routes"
)

func TestDataSourceHardenedImagesRead(t *testing.T) {
	routes.ResetHardenedImagesState()
	d := dataSourceHardenedImages().TestResourceData()

	diags := dataSourceHardenedImagesRead(context.Background(), d, unitTestMockAPIClient)

	assert.Empty(t, diags)
	assert.Equal(t, "test-tenant-uid", d.Id())
	assert.Equal(t, "InProgress", d.Get("state"))
	// Every third synthetic cluster is reported as failed.
	assert.Equal(t, routes.MockHardenedImagesClusterCount/3, d.Get("failed_clusters_count"))
	assert.Len(t, d.Get("clusters").([]interface{}), routes.MockHardenedImagesClusterCount)
	assert.Equal(t, 0, d.Get("excluded_cluster_uids").(*schema.Set).Len())

	first := d.Get("clusters").([]interface{})[0].(map[string]interface{})
	assert.Equal(t, routes.MockHardenedImagesClusterUID(1), first["uid"])
	assert.Equal(t, "Default", first["project_name"])
}

// The data source must report the clusters that have been opted out.
func TestDataSourceHardenedImagesReadReportsExclusions(t *testing.T) {
	routes.ResetHardenedImagesState()
	c := getV1ClientWithResourceContext(unitTestMockAPIClient, tenantString)
	require.NoError(t, c.UpdateImagePullSecretClusterExclusion(
		[]string{routes.MockHardenedImagesClusterUID(4)}, true))

	d := dataSourceHardenedImages().TestResourceData()
	require.Empty(t, dataSourceHardenedImagesRead(context.Background(), d, unitTestMockAPIClient))

	excluded := d.Get("excluded_cluster_uids").(*schema.Set)
	assert.Equal(t, 1, excluded.Len())
	assert.True(t, excluded.Contains(routes.MockHardenedImagesClusterUID(4)))

	routes.ResetHardenedImagesState()
}

func TestDataSourceHardenedImagesReadWithFilters(t *testing.T) {
	routes.ResetHardenedImagesState()
	d := dataSourceHardenedImages().TestResourceData()
	require.NoError(t, d.Set("cluster_name", "hardened-cluster-1"))
	require.NoError(t, d.Set("project_uid", "test-project-uid"))

	diags := dataSourceHardenedImagesRead(context.Background(), d, unitTestMockAPIClient)

	assert.Empty(t, diags)
	assert.Equal(t, "test-tenant-uid", d.Id())
}

func TestDataSourceHardenedImagesReadNegative(t *testing.T) {
	d := dataSourceHardenedImages().TestResourceData()

	diags := dataSourceHardenedImagesRead(context.Background(), d, unitTestMockAPINegativeClient)

	require.NotEmpty(t, diags)
	assert.Contains(t, diags[0].Summary, "Invalid hardened images status request")
}
