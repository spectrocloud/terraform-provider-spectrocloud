package spectrocloud

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/spectrocloud/terraform-provider-spectrocloud/tests/mockApiServer/routes"
)

// (test file header)
//
// resource_cluster_maas_wave2_test.go's TestResourceClusterMaasReadWithMock
// only exercises the happy path. This file rounds out
// resourceClusterMaasRead's error/early-return branches and
// flattenCloudConfigMaas's GetCloudConfigMaas error branch, mirroring
// resource_cluster_gcp_wave2_test.go's Read-error coverage.

// TestResourceClusterMaasReadClusterDeleted exercises the "cluster == nil"
// branch: GetCluster returns (nil, nil) once the fixture's State is
// "Deleted", so Read must clear the resource's ID instead of erroring.
func TestResourceClusterMaasReadClusterDeleted(t *testing.T) {
	d := prepareMaasClusterResourceData(t)
	d.SetId("cluster-uid-deleted-state")

	diags := resourceClusterMaasRead(context.Background(), d, unitTestMockAPIClient)
	assert.False(t, diags.HasError(), "diags: %+v", diags)
	assert.Equal(t, "", d.Id())
}

// TestResourceClusterMaasReadCloudTypeMismatch exercises the
// ValidateCloudType error branch: the default cluster fixture reports
// CloudType="aws", which resourceClusterMaasRead must reject.
func TestResourceClusterMaasReadCloudTypeMismatch(t *testing.T) {
	d := prepareMaasClusterResourceData(t)
	d.SetId("test-cluster-id")

	diags := resourceClusterMaasRead(context.Background(), d, unitTestMockAPIClient)
	assert.True(t, diags.HasError())
}

// TestFlattenCloudConfigMaas_GetCloudConfigError exercises
// flattenCloudConfigMaas's GetCloudConfigMaas error branch directly.
func TestFlattenCloudConfigMaas_GetCloudConfigError(t *testing.T) {
	d := resourceClusterMaas().TestResourceData()
	c := mustUnitClient(t, false)

	diags := flattenCloudConfigMaas(routes.MaasCloudConfigErrorUID, d, c)
	assert.True(t, diags.HasError())
}

// TestFlattenCloudConfigMaasWithMock covers flattenCloudConfigMaas's happy
// path directly, mirroring TestFlattenCloudConfigGcpWithMock.
func TestFlattenCloudConfigMaasWithMock(t *testing.T) {
	d := prepareMaasClusterResourceData(t)
	c := mustUnitClient(t, false)

	diags := flattenCloudConfigMaas(maasCloudConfigUID, d, c)
	assert.False(t, diags.HasError(), "diags: %+v", diags)
	assert.Equal(t, maasCloudAccountUID, d.Get("cloud_account_id"))

	cfg := d.Get("cloud_config").([]interface{})
	require.Len(t, cfg, 1)
	assert.Equal(t, "test-domain", cfg[0].(map[string]interface{})["domain"])
}
