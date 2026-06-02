package spectrocloud

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/stretchr/testify/assert"
)

func TestGetSpectroComponentsUpgradeAdditional(t *testing.T) {
	tests := []struct {
		name     string
		cluster  *models.V1SpectroCluster
		expected string
	}{
		{
			name: "locked when annotation true",
			cluster: &models.V1SpectroCluster{
				Metadata: &models.V1ObjectMeta{
					Annotations: map[string]string{"spectroComponentsUpgradeForbidden": "true"},
				},
			},
			expected: "lock",
		},
		{
			name: "unlock when annotation false",
			cluster: &models.V1SpectroCluster{
				Metadata: &models.V1ObjectMeta{
					Annotations: map[string]string{"spectroComponentsUpgradeForbidden": "false"},
				},
			},
			expected: "unlock",
		},
		{
			name: "default unlock when annotation missing",
			cluster: &models.V1SpectroCluster{
				Metadata: &models.V1ObjectMeta{},
			},
			expected: "unlock",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, getSpectroComponentsUpgrade(tt.cluster))
		})
	}
}

func TestValidateOverrideScalingAdditionalCases(t *testing.T) {
	buildMachinePoolSet := func(pool map[string]interface{}) *schema.Set {
		return schema.NewSet(resourceMachinePoolApacheCloudStackHash, []interface{}{pool})
	}

	t.Run("override scaling strategy without block returns error", func(t *testing.T) {
		d := resourceClusterApacheCloudStack().TestResourceData()
		_ = d.Set("machine_pool", buildMachinePoolSet(map[string]interface{}{
			"name":            "mp-1",
			"update_strategy": "OverrideScaling",
		}))

		err := validateOverrideScaling(d, "machine_pool")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "override_scaling must be specified")
	})

	t.Run("override scaling with missing max_unavailable returns error", func(t *testing.T) {
		d := resourceClusterApacheCloudStack().TestResourceData()
		_ = d.Set("machine_pool", buildMachinePoolSet(map[string]interface{}{
			"name":            "mp-2",
			"update_strategy": "OverrideScaling",
			"override_scaling": []interface{}{
				map[string]interface{}{
					"max_surge":       "1",
					"max_unavailable": "",
				},
			},
		}))

		err := validateOverrideScaling(d, "machine_pool")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "max_unavailable is required")
	})

	t.Run("non override strategy succeeds", func(t *testing.T) {
		d := resourceClusterApacheCloudStack().TestResourceData()
		_ = d.Set("machine_pool", buildMachinePoolSet(map[string]interface{}{
			"name":            "mp-3",
			"update_strategy": "RollingUpdateScaleOut",
		}))

		assert.NoError(t, validateOverrideScaling(d, "machine_pool"))
	})
}
