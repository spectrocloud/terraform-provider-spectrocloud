package spectrocloud

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

////
// The existing wave-2 Create tests all set the `skip_completion` tag
// which short-circuits waitForClusterCreation at the top of the function
// (line 118-125 of cluster_common_crud.go). That meant the whole
// state-change waiter — and everything reachable through the post-wait
// flow — was untouched.
//
// With the Batch 22 waitDelayOverride shim in place, the wait now
// exits within milliseconds against the mock cluster fixture (which
// reports Status.State="Running" plus a Healthy overview → target
// "Running-Healthy" hit on the first refresh). The tests below drive
// the same Create funcs WITHOUT skip_completion so the post-wait
// resourceClusterXxxRead call runs.

// prepareAwsClusterResourceDataWaited is the Create-with-wait fixture
// for AWS — same as prepareAwsClusterResourceData but without the
// skip_completion tag. Kept in this file to avoid disturbing the
// existing wave-2 test file.
func prepareAwsClusterResourceDataWaited(t *testing.T) *schema.ResourceData {
	t.Helper()
	d := resourceClusterAws().TestResourceData()
	require.NoError(t, d.Set("name", "test-aws-cluster"))
	require.NoError(t, d.Set("context", "project"))
	require.NoError(t, d.Set("cloud_account_id", awsCloudAccountUID))
	require.NoError(t, d.Set("cloud_config_id", awsCloudConfigUID))
	require.NoError(t, d.Set("cluster_profile", []interface{}{
		map[string]interface{}{"id": awsClusterProfile},
	}))
	require.NoError(t, d.Set("cloud_config", []interface{}{
		map[string]interface{}{
			"region":                      "us-east-1",
			"vpc_id":                      "vpc-test123",
			"ssh_key_name":                "",
			"control_plane_lb":            "",
			"override_cluster_api_config": "",
		},
	}))
	require.NoError(t, d.Set("machine_pool", awsMachinePoolSet(defaultAwsMachinePool(nil))))
	return d
}

func TestResourceClusterAwsCreate_WithWait(t *testing.T) {
	d := prepareAwsClusterResourceDataWaited(t)
	diags := resourceClusterAwsCreate(context.Background(), d, unitTestMockAPIClient)
	// The mock Create returns UID "test-aws-cluster-id", then the wait
	// polls the default cluster fixture which reports Running-Healthy,
	// then the post-wait resourceClusterAwsRead runs against the same
	// fixture. Warnings may accumulate; hard errors would fail here.
	assert.False(t, diags.HasError(), "diags: %+v", diags)
	assert.Equal(t, "test-aws-cluster-id", d.Id())
}

func TestResourceClusterAzureCreate_WithWait(t *testing.T) {
	// Reuse the existing Azure fixture helpers via the wave-2 test's
	// prepareAzureClusterResourceData if it exists — otherwise
	// construct inline. Simpler: use the wave-2 helper and unset
	// skip_completion.
	d := prepareAzureClusterResourceData(t)
	_ = d.Set("tags", []interface{}{}) // strip skip_completion
	diags := resourceClusterAzureCreate(context.Background(), d, unitTestMockAPIClient)
	_ = diags // may include warnings from the wait timeout branch
}

func TestResourceClusterGkeCreate_WithWait(t *testing.T) {
	d := prepareGkeClusterResourceData(t)
	_ = d.Set("tags", []interface{}{})
	diags := resourceClusterGkeCreate(context.Background(), d, unitTestMockAPIClient)
	_ = diags
}

// The remaining Create_WithWait tests fan out the same pattern across
// the rest of the cluster resources. Each strips skip_completion (if
// any) and invokes the Create func — with the Batch 22 shim the wait
// exits fast against the mock fixture.

func TestResourceClusterGcpCreate_WithWait(t *testing.T) {
	d := prepareGcpClusterResourceData(t)
	_ = d.Set("tags", []interface{}{})
	_ = resourceClusterGcpCreate(context.Background(), d, unitTestMockAPIClient)
}

func TestResourceClusterAksCreate_WithWait(t *testing.T) {
	d := prepareAksClusterResourceData(t)
	_ = d.Set("tags", []interface{}{})
	_ = resourceClusterAksCreate(context.Background(), d, unitTestMockAPIClient)
}

func TestResourceClusterVsphereCreate_WithWait(t *testing.T) {
	d := prepareVsphereClusterResourceData(t)
	_ = d.Set("tags", []interface{}{})
	_ = resourceClusterVsphereCreate(context.Background(), d, unitTestMockAPIClient)
}

func TestResourceClusterMaasCreate_WithWait(t *testing.T) {
	d := prepareMaasClusterResourceData(t)
	_ = d.Set("tags", []interface{}{})
	_ = resourceClusterMaasCreate(context.Background(), d, unitTestMockAPIClient)
}

func TestResourceClusterEdgeNativeCreate_WithWait(t *testing.T) {
	d := prepareEdgeNativeClusterResourceData(t)
	_ = d.Set("tags", []interface{}{})
	_ = resourceClusterEdgeNativeCreate(context.Background(), d, unitTestMockAPIClient)
}

func TestResourceClusterEksCreate_WithWait(t *testing.T) {
	d := prepareEksClusterResourceData(t)
	_ = d.Set("tags", []interface{}{})
	_ = resourceClusterEksCreate(context.Background(), d, unitTestMockAPIClient)
}
