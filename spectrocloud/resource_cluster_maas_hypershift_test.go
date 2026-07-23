package spectrocloud

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/spectrocloud/palette-sdk-go/api/models"
)

// (test file header)
//
// hyper_shift_config is a MaaS-specific block (HyperShift / OpenShift
// control-plane hosting) not present on any other cloud resource in this
// provider. validateHyperShiftMaasConfig, toHyperShiftConfigMaas and
// flattenHyperShiftConfigMaas were all under-covered — this file exercises
// every branch of the three functions.

func maasResourceDataWithHyperShift(t *testing.T, hyperShift []interface{}) *schema.ResourceData {
	t.Helper()
	d := resourceClusterMaas().TestResourceData()
	if hyperShift != nil {
		require.NoError(t, d.Set("hyper_shift_config", hyperShift))
	}
	return d
}

func TestValidateHyperShiftMaasConfig(t *testing.T) {
	tests := []struct {
		name      string
		hs        []interface{}
		expectErr bool
	}{
		{
			name:      "not set at all",
			hs:        nil,
			expectErr: false,
		},
		{
			name: "hypershift type without host_cluster_uid is fine",
			hs: []interface{}{map[string]interface{}{
				"cluster_deployment_type": string(models.V1ClusterDeploymentTypeHypershift),
				"host_cluster_uid":        "",
			}},
			expectErr: false,
		},
		{
			name: "openshift type without host_cluster_uid errors",
			hs: []interface{}{map[string]interface{}{
				"cluster_deployment_type": string(models.V1ClusterDeploymentTypeOpenshift),
				"host_cluster_uid":        "",
			}},
			expectErr: true,
		},
		{
			name: "openshift type with host_cluster_uid is fine",
			hs: []interface{}{map[string]interface{}{
				"cluster_deployment_type": string(models.V1ClusterDeploymentTypeOpenshift),
				"host_cluster_uid":        "host-cluster-uid-1",
			}},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := maasResourceDataWithHyperShift(t, tt.hs)
			err := validateHyperShiftMaasConfig(d)
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestToHyperShiftConfigMaas(t *testing.T) {
	t.Run("not set returns nil", func(t *testing.T) {
		d := maasResourceDataWithHyperShift(t, nil)
		got := toHyperShiftConfigMaas(d)
		assert.Nil(t, got)
	})

	t.Run("hypershift type without host_cluster_uid", func(t *testing.T) {
		d := maasResourceDataWithHyperShift(t, []interface{}{map[string]interface{}{
			"cluster_deployment_type": string(models.V1ClusterDeploymentTypeHypershift),
			"host_cluster_uid":        "",
		}})
		got := toHyperShiftConfigMaas(d)
		require.NotNil(t, got)
		assert.Equal(t, models.V1ClusterDeploymentTypeHypershift, got.ClusterDeploymentType)
		assert.Empty(t, got.HostClusterUID)
	})

	t.Run("openshift type with host_cluster_uid", func(t *testing.T) {
		d := maasResourceDataWithHyperShift(t, []interface{}{map[string]interface{}{
			"cluster_deployment_type": string(models.V1ClusterDeploymentTypeOpenshift),
			"host_cluster_uid":        "host-cluster-uid-1",
		}})
		got := toHyperShiftConfigMaas(d)
		require.NotNil(t, got)
		assert.Equal(t, models.V1ClusterDeploymentTypeOpenshift, got.ClusterDeploymentType)
		assert.Equal(t, "host-cluster-uid-1", got.HostClusterUID)
	})
}

func TestFlattenHyperShiftConfigMaas(t *testing.T) {
	t.Run("nil Spec clears the field", func(t *testing.T) {
		d := maasResourceDataWithHyperShift(t, []interface{}{map[string]interface{}{
			"cluster_deployment_type": string(models.V1ClusterDeploymentTypeHypershift),
			"host_cluster_uid":        "",
		}})
		cluster := &models.V1SpectroCluster{}
		err := flattenHyperShiftConfigMaas(cluster, d)
		require.NoError(t, err)
		hs := d.Get("hyper_shift_config").([]interface{})
		assert.Len(t, hs, 0)
	})

	t.Run("nil ClusterConfig clears the field", func(t *testing.T) {
		d := maasResourceDataWithHyperShift(t, nil)
		cluster := &models.V1SpectroCluster{
			Spec: &models.V1SpectroClusterSpec{},
		}
		err := flattenHyperShiftConfigMaas(cluster, d)
		require.NoError(t, err)
		hs := d.Get("hyper_shift_config").([]interface{})
		assert.Len(t, hs, 0)
	})

	t.Run("nil HyperShiftConfig clears the field", func(t *testing.T) {
		d := maasResourceDataWithHyperShift(t, nil)
		cluster := &models.V1SpectroCluster{
			Spec: &models.V1SpectroClusterSpec{
				ClusterConfig: &models.V1ClusterConfig{},
			},
		}
		err := flattenHyperShiftConfigMaas(cluster, d)
		require.NoError(t, err)
		hs := d.Get("hyper_shift_config").([]interface{})
		assert.Len(t, hs, 0)
	})

	t.Run("populated HyperShiftConfig is flattened", func(t *testing.T) {
		d := maasResourceDataWithHyperShift(t, nil)
		cluster := &models.V1SpectroCluster{
			Spec: &models.V1SpectroClusterSpec{
				ClusterConfig: &models.V1ClusterConfig{
					HyperShiftConfig: &models.V1HyperShiftConfig{
						ClusterDeploymentType: models.V1ClusterDeploymentTypeOpenshift,
						HostClusterUID:        "host-cluster-uid-1",
					},
				},
			},
		}
		err := flattenHyperShiftConfigMaas(cluster, d)
		require.NoError(t, err)
		hs := d.Get("hyper_shift_config").([]interface{})
		require.Len(t, hs, 1)
		m := hs[0].(map[string]interface{})
		assert.Equal(t, string(models.V1ClusterDeploymentTypeOpenshift), m["cluster_deployment_type"])
		assert.Equal(t, "host-cluster-uid-1", m["host_cluster_uid"])
	})
}
