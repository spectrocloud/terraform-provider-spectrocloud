package spectrocloud

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

////
// Pure schema → SDK model builder helpers that were at 0%. Each takes
// simple input (map[string]interface{} for the cloud_config or a
// ResourceData) and returns a populated model struct — no SDK calls.

// TestToCloudConfigAws — trivial passthrough wrapper around
// toAwsClusterConfig (already covered).
func TestToCloudConfigAws(t *testing.T) {
	got := toCloudConfigAws(map[string]interface{}{
		"ssh_key_name":                "my-key",
		"region":                      "us-east-1",
		"vpc_id":                      "vpc-abc",
		"control_plane_lb":            "internal",
		"override_cluster_api_config": "",
	})
	require.NotNil(t, got)
	require.NotNil(t, got.ClusterConfig)
	assert.Equal(t, "my-key", got.ClusterConfig.SSHKeyName)
	require.NotNil(t, got.ClusterConfig.Region)
	assert.Equal(t, "us-east-1", *got.ClusterConfig.Region)
	assert.Equal(t, "vpc-abc", got.ClusterConfig.VpcID)
	assert.Equal(t, "internal", got.ClusterConfig.ControlPlaneLoadBalancer)
}
