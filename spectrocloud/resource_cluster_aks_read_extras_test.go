package spectrocloud

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

// (test file header)
//
// resource_cluster_aks_wave2_test.go's TestResourceClusterAksReadWithMock
// only exercises the happy path. This file rounds out
// resourceClusterAksRead's error/early-return branches, mirroring
// resource_cluster_gcp_wave2_test.go's Read-error coverage.

// TestResourceClusterAksReadClusterDeleted exercises the "cluster == nil"
// branch: GetCluster returns (nil, nil) once the fixture's State is
// "Deleted", so Read must clear the resource's ID instead of erroring.
func TestResourceClusterAksReadClusterDeleted(t *testing.T) {
	d := prepareAksClusterResourceData(t)
	d.SetId("cluster-uid-deleted-state")

	diags := resourceClusterAksRead(context.Background(), d, unitTestMockAPIClient)
	assert.False(t, diags.HasError(), "diags: %+v", diags)
	assert.Equal(t, "", d.Id())
}

// TestResourceClusterAksReadCloudTypeMismatch exercises the
// ValidateCloudType error branch: the default cluster fixture reports
// CloudType="aws", which resourceClusterAksRead must reject.
func TestResourceClusterAksReadCloudTypeMismatch(t *testing.T) {
	d := prepareAksClusterResourceData(t)
	d.SetId("test-cluster-id")

	diags := resourceClusterAksRead(context.Background(), d, unitTestMockAPIClient)
	assert.True(t, diags.HasError())
}

// TestResourceClusterAksRead_GetCloudConfigError exercises
// resourceClusterAksRead's GetCloudConfigAks error branch: the cluster
// fixture's CloudConfigRef points at AksCloudConfigGetErrorUID, so the
// cloud-config fetch after the cluster read fails.
func TestResourceClusterAksRead_GetCloudConfigError(t *testing.T) {
	d := prepareAksClusterResourceData(t)
	d.SetId("test-aks-cluster-cloudconfig-error-id")

	diags := resourceClusterAksRead(context.Background(), d, unitTestMockAPIClient)
	assert.True(t, diags.HasError())
}
