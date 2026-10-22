package spectrocloud

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/spectrocloud/terraform-provider-spectrocloud/tests/mockApiServer/routes"
)

// TestResourceClusterAwsRead_AdminKubeConfigForbidden covers PLT-2355: a
// caller without the cluster.adminKubeconfigDownload permission gets a 403
// from GetClusterAdminKubeConfig on every plan/apply. Read must tolerate
// that specific failure (leaving admin_kube_config unset) instead of
// aborting the whole read, since the permission is unrelated to the
// caller's ability to manage the cluster itself.
func TestResourceClusterAwsRead_AdminKubeConfigForbidden(t *testing.T) {
	d := prepareAwsClusterResourceData(t)
	d.SetId(routes.AdminKubeConfigForbiddenUID)

	diags := resourceClusterAwsRead(context.Background(), d, unitTestMockAPIClient)
	assert.False(t, diags.HasError(), "diags: %+v", diags)
	assert.Empty(t, d.Get("admin_kube_config"))
}

// TestWarnDeprecatedOverrideKubeadmConfigurationForManagedCluster covers
// PLT-2475: override_kubeadm_configuration is not supported by Palette for
// managed clusters (EKS, GKE, AKS - the UI already hides it there, PCP-6073).
// It's deprecated rather than hard-removed, so setting it must surface a
// warning instead of silently doing nothing or blocking the plan.
func TestWarnDeprecatedOverrideKubeadmConfigurationForManagedCluster(t *testing.T) {
	t.Run("EKS warns when set", func(t *testing.T) {
		d := prepareEksClusterResourceData(t)
		pool := defaultEksMachinePool(map[string]interface{}{
			"override_kubeadm_configuration": "kubeletExtraArgs:\n  max-pods: \"110\"",
		})
		require.NoError(t, d.Set("machine_pool", eksMachinePoolSet(pool)))

		diags := warnDeprecatedOverrideKubeadmConfigurationForManagedCluster(d, "EKS")
		require.Len(t, diags, 1)
		assert.Equal(t, diag.Warning, diags[0].Severity)
		assert.Contains(t, diags[0].Summary, "deprecated")
	})

	t.Run("EKS no warning when unset", func(t *testing.T) {
		d := prepareEksClusterResourceData(t)
		diags := warnDeprecatedOverrideKubeadmConfigurationForManagedCluster(d, "EKS")
		assert.Empty(t, diags)
	})

	t.Run("GKE warns when set", func(t *testing.T) {
		d := prepareGkeClusterResourceData(t)
		pool := defaultGkeMachinePool(map[string]interface{}{
			"override_kubeadm_configuration": "kubeletExtraArgs:\n  max-pods: \"110\"",
		})
		require.NoError(t, d.Set("machine_pool", gkeMachinePoolSet(pool)))

		diags := warnDeprecatedOverrideKubeadmConfigurationForManagedCluster(d, "GKE")
		require.Len(t, diags, 1)
		assert.Equal(t, diag.Warning, diags[0].Severity)
	})

	t.Run("GKE no warning when unset", func(t *testing.T) {
		d := prepareGkeClusterResourceData(t)
		diags := warnDeprecatedOverrideKubeadmConfigurationForManagedCluster(d, "GKE")
		assert.Empty(t, diags)
	})

	t.Run("AKS warns when set", func(t *testing.T) {
		d := prepareAksClusterResourceData(t)
		pool := defaultAksMachinePool(map[string]interface{}{
			"override_kubeadm_configuration": "kubeletExtraArgs:\n  max-pods: \"110\"",
		})
		require.NoError(t, d.Set("machine_pool", aksMachinePoolSet(pool)))

		diags := warnDeprecatedOverrideKubeadmConfigurationForManagedCluster(d, "AKS")
		require.Len(t, diags, 1)
		assert.Equal(t, diag.Warning, diags[0].Severity)
	})

	t.Run("AKS no warning when unset", func(t *testing.T) {
		d := prepareAksClusterResourceData(t)
		diags := warnDeprecatedOverrideKubeadmConfigurationForManagedCluster(d, "AKS")
		assert.Empty(t, diags)
	})
}
