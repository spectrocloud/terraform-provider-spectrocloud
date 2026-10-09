package spectrocloud

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/spectrocloud/terraform-provider-spectrocloud/tests/mockApiServer/routes"
)

func prepareHardenedImagesResourceData() *schema.ResourceData {
	routes.ResetHardenedImagesState()
	d := resourceHardenedImages().TestResourceData()
	_ = d.Set("excluded_cluster_uids", []interface{}{
		routes.MockHardenedImagesClusterUID(1),
		routes.MockHardenedImagesClusterUID(2),
	})
	return d
}

// ---------------------------------------------------------------------------
// Flatten
// ---------------------------------------------------------------------------

func TestFlattenHardenedImagesClusters(t *testing.T) {
	items := []*models.V1ImagePullSecretTenantPropagationClusterStatus{
		{
			Cluster: &models.V1ObjectReference{UID: "cluster-1", Name: "prod"},
			Project: &models.V1ObjectReference{UID: "project-1", Name: "Default"},
			Exclude: true,
			State:   models.V1ImagePullSecretPropagationStateFailed,
			Reason:  "ConnectivityIssue",
			Message: "agent unreachable",
		},
		// A tenant-scoped cluster has no project, and a nil entry must be skipped outright.
		{
			Cluster: &models.V1ObjectReference{UID: "cluster-2", Name: "staging"},
			State:   models.V1ImagePullSecretPropagationStateCompleted,
		},
		nil,
	}

	got := flattenHardenedImagesClusters(items)
	require.Len(t, got, 2)

	first := got[0].(map[string]interface{})
	assert.Equal(t, "cluster-1", first["uid"])
	assert.Equal(t, "prod", first["name"])
	assert.Equal(t, "project-1", first["project_uid"])
	assert.Equal(t, "Default", first["project_name"])
	assert.Equal(t, true, first["exclude"])
	assert.Equal(t, "Failed", first["state"])
	assert.Equal(t, "ConnectivityIssue", first["reason"])
	assert.Equal(t, "agent unreachable", first["message"])

	second := got[1].(map[string]interface{})
	assert.Equal(t, "cluster-2", second["uid"])
	assert.NotContains(t, second, "project_uid")
	assert.Equal(t, false, second["exclude"])
}

func TestExcludedClusterUIDs(t *testing.T) {
	items := []*models.V1ImagePullSecretTenantPropagationClusterStatus{
		{Cluster: &models.V1ObjectReference{UID: "cluster-1"}, Exclude: true},
		{Cluster: &models.V1ObjectReference{UID: "cluster-2"}, Exclude: false},
		// Excluded but missing the cluster reference — nothing usable to record.
		{Exclude: true},
		nil,
	}

	assert.Equal(t, []string{"cluster-1"}, excludedClusterUIDs(items))
}

func TestFlattenHardenedImages(t *testing.T) {
	d := resourceHardenedImages().TestResourceData()
	status := &models.V1ImagePullSecretTenantPropagationStatus{
		State:    models.V1ImagePullSecretPropagationStateInProgress,
		Clusters: &models.V1ImagePullSecretTenantPropagationClusterCounts{Failed: 3},
		Items: []*models.V1ImagePullSecretTenantPropagationClusterStatus{
			{Cluster: &models.V1ObjectReference{UID: "cluster-1"}, Exclude: true},
		},
	}

	require.NoError(t, flattenHardenedImages(status, d))
	assert.Equal(t, "InProgress", d.Get("state"))
	assert.Equal(t, 3, d.Get("failed_clusters_count"))
	assert.Len(t, d.Get("clusters").([]interface{}), 1)
}

// A status payload with no cluster counts must flatten to a zero failure count rather than panic.
func TestFlattenHardenedImagesWithoutCounts(t *testing.T) {
	d := resourceHardenedImages().TestResourceData()
	status := &models.V1ImagePullSecretTenantPropagationStatus{
		State: models.V1ImagePullSecretPropagationStateCompleted,
	}

	require.NoError(t, flattenHardenedImages(status, d))
	assert.Equal(t, "Completed", d.Get("state"))
	assert.Equal(t, 0, d.Get("failed_clusters_count"))
	assert.Empty(t, d.Get("clusters").([]interface{}))
}

// ---------------------------------------------------------------------------
// Paging
// ---------------------------------------------------------------------------

// The mock tenant holds more clusters than one page can carry, so a correct Read has to follow
// the offset until the API runs out of rows.
func TestGetHardenedImagesStatusPagesThroughEveryCluster(t *testing.T) {
	routes.ResetHardenedImagesState()
	c := getV1ClientWithResourceContext(unitTestMockAPIClient, tenantString)

	status, err := getHardenedImagesStatus(c, "", "")

	require.NoError(t, err)
	require.Len(t, status.Items, routes.MockHardenedImagesClusterCount)
	assert.Equal(t, routes.MockHardenedImagesClusterUID(1), status.Items[0].Cluster.UID)
	assert.Equal(t,
		routes.MockHardenedImagesClusterUID(routes.MockHardenedImagesClusterCount),
		status.Items[routes.MockHardenedImagesClusterCount-1].Cluster.UID)
}

// ---------------------------------------------------------------------------
// CRUD
// ---------------------------------------------------------------------------

func TestResourceHardenedImagesCRUD(t *testing.T) {
	testResourceCRUD(t, prepareHardenedImagesResourceData, unitTestMockAPIClient,
		resourceHardenedImagesCreate, resourceHardenedImagesRead,
		resourceHardenedImagesUpdate, resourceHardenedImagesDelete)
}

// Create must push the configured exclusions, and the following Read must observe them.
func TestResourceHardenedImagesCreateAppliesExclusions(t *testing.T) {
	d := prepareHardenedImagesResourceData()

	diags := resourceHardenedImagesCreate(context.Background(), d, unitTestMockAPIClient)

	require.Empty(t, diags)
	assert.Equal(t, hardenedImagesID, d.Id())
	excluded := d.Get("excluded_cluster_uids").(*schema.Set)
	assert.Equal(t, 2, excluded.Len())
	assert.True(t, excluded.Contains(routes.MockHardenedImagesClusterUID(1)))
	assert.True(t, excluded.Contains(routes.MockHardenedImagesClusterUID(2)))
	assert.Equal(t, "InProgress", d.Get("state"))
	assert.Len(t, d.Get("clusters").([]interface{}), routes.MockHardenedImagesClusterCount)
}

// Dropping a UID from the set must re-enroll that cluster on the next apply.
func TestResourceHardenedImagesUpdateReenrollsDroppedCluster(t *testing.T) {
	d := prepareHardenedImagesResourceData()
	require.Empty(t, resourceHardenedImagesCreate(context.Background(), d, unitTestMockAPIClient))

	require.NoError(t, d.Set("excluded_cluster_uids", []interface{}{routes.MockHardenedImagesClusterUID(2)}))
	diags := resourceHardenedImagesUpdate(context.Background(), d, unitTestMockAPIClient)

	require.Empty(t, diags)
	excluded := d.Get("excluded_cluster_uids").(*schema.Set)
	assert.Equal(t, 1, excluded.Len())
	assert.False(t, excluded.Contains(routes.MockHardenedImagesClusterUID(1)))
	assert.True(t, excluded.Contains(routes.MockHardenedImagesClusterUID(2)))
}

// Destroying the resource drops it from state and deliberately leaves the tenant untouched.
func TestResourceHardenedImagesDeleteLeavesExclusionsInPlace(t *testing.T) {
	d := prepareHardenedImagesResourceData()
	require.Empty(t, resourceHardenedImagesCreate(context.Background(), d, unitTestMockAPIClient))

	diags := resourceHardenedImagesDelete(context.Background(), d, unitTestMockAPIClient)

	require.Empty(t, diags)
	assert.Empty(t, d.Id())

	// The clusters excluded before the destroy are still excluded afterwards.
	c := getV1ClientWithResourceContext(unitTestMockAPIClient, tenantString)
	status, err := getHardenedImagesStatus(c, "", "")
	require.NoError(t, err)
	assert.ElementsMatch(t,
		[]string{routes.MockHardenedImagesClusterUID(1), routes.MockHardenedImagesClusterUID(2)},
		excludedClusterUIDs(status.Items))
}

// A state entry that does not carry the singleton ID is dropped rather than refreshed.
func TestResourceHardenedImagesReadWithoutID(t *testing.T) {
	routes.ResetHardenedImagesState()
	d := resourceHardenedImages().TestResourceData()
	d.SetId("some-other-id")

	diags := resourceHardenedImagesRead(context.Background(), d, unitTestMockAPIClient)

	assert.Empty(t, diags)
	assert.Empty(t, d.Id())
}

func TestResourceHardenedImagesCRUDNegative(t *testing.T) {
	// Delete is intentionally excluded: it makes no API call, so it has no error path.
	for _, op := range []string{"Create", "Read", "Update"} {
		t.Run(op+" surfaces API error", func(t *testing.T) {
			testResourceCRUDNegative(t, op, prepareHardenedImagesResourceData,
				unitTestMockAPINegativeClient,
				resourceHardenedImagesCreate, resourceHardenedImagesRead,
				resourceHardenedImagesUpdate, resourceHardenedImagesDelete,
				op != "Create", "Invalid hardened images status request")
		})
	}
}

// ---------------------------------------------------------------------------
// Import
// ---------------------------------------------------------------------------

func TestResourceHardenedImagesImport(t *testing.T) {
	t.Run("uid matches tenant", func(t *testing.T) {
		routes.ResetHardenedImagesState()
		d := resourceHardenedImages().TestResourceData()
		d.SetId("test-tenant-uid") // matches getMockUserInfoPayload().tenantUid
		got, err := resourceHardenedImagesImport(context.Background(), d, unitTestMockAPIClient)
		require.NoError(t, err)
		require.Len(t, got, 1)
		assert.Equal(t, hardenedImagesID, got[0].Id())
		assert.Len(t, got[0].Get("clusters").([]interface{}), routes.MockHardenedImagesClusterCount)
	})

	t.Run("uid mismatch errors", func(t *testing.T) {
		d := resourceHardenedImages().TestResourceData()
		d.SetId("some-other-tenant-uid")
		_, err := resourceHardenedImagesImport(context.Background(), d, unitTestMockAPIClient)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "does not match")
	})
}
