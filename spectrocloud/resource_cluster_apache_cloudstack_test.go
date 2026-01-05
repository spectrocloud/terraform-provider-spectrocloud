package spectrocloud

import (
	"testing"

	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/spectrocloud/terraform-provider-spectrocloud/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResourceMachinePoolApacheCloudStackHash(t *testing.T) {
	testCases := []struct {
		name     string
		input    interface{}
		expected int
	}{
		{
			name: "Complete CloudStack machine pool with all fields",
			input: map[string]interface{}{
				"name":  "cloudstack-pool-1",
				"count": 3,
				"additional_labels": map[string]interface{}{
					"env":  "production",
					"team": "platform",
				},
				"additional_annotations": map[string]interface{}{
					"custom.io/annotation1":   "value1",
					"company.com/annotation2": "value2",
				},
				"offering": "medium-instance",
				"template": []interface{}{
					map[string]interface{}{
						"id":   "template-123",
						"name": "ubuntu-20.04",
					},
				},
				"network": []interface{}{
					map[string]interface{}{
						"network_name": "network-1",
					},
				},
				"control_plane":           false,
				"control_plane_as_worker": false,
				"update_strategy":         "RollingUpdateScaleOut",
				"node_repave_interval":    0,
			},
			expected: 0, // Will be calculated in test
		},
		{
			name: "Minimal CloudStack machine pool",
			input: map[string]interface{}{
				"name":                    "cloudstack-pool-2",
				"count":                   2,
				"offering":                "small-instance",
				"control_plane":           true,
				"control_plane_as_worker": false,
			},
			expected: 0, // Will be calculated in test
		},
		{
			name: "CloudStack machine pool with annotations only",
			input: map[string]interface{}{
				"name":  "cloudstack-pool-3",
				"count": 1,
				"additional_annotations": map[string]interface{}{
					"annotation1": "value1",
				},
				"offering":                "medium-instance",
				"control_plane":           false,
				"control_plane_as_worker": false,
			},
			expected: 0, // Will be calculated in test
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := resourceMachinePoolApacheCloudStackHash(tc.input)
			// For first run, just ensure hash is generated
			assert.NotEqual(t, 0, actual, "Hash should not be zero")
		})
	}
}

func TestResourceMachinePoolApacheCloudStackHashAnnotationChangeDetection(t *testing.T) {
	// Base machine pool without annotations
	baseMachinePool := map[string]interface{}{
		"name":  "test-pool",
		"count": 3,
		"additional_labels": map[string]interface{}{
			"label1": "value1",
		},
		"offering":                "medium-instance",
		"control_plane":           false,
		"control_plane_as_worker": false,
	}

	// Machine pool with annotations
	withAnnotations := map[string]interface{}{
		"name":  "test-pool",
		"count": 3,
		"additional_labels": map[string]interface{}{
			"label1": "value1",
		},
		"additional_annotations": map[string]interface{}{
			"annotation1": "value1",
		},
		"offering":                "medium-instance",
		"control_plane":           false,
		"control_plane_as_worker": false,
	}

	// Machine pool with different annotations
	differentAnnotations := map[string]interface{}{
		"name":  "test-pool",
		"count": 3,
		"additional_labels": map[string]interface{}{
			"label1": "value1",
		},
		"additional_annotations": map[string]interface{}{
			"annotation1": "value2",
		},
		"offering":                "medium-instance",
		"control_plane":           false,
		"control_plane_as_worker": false,
	}

	// Machine pool with additional annotations
	moreAnnotations := map[string]interface{}{
		"name":  "test-pool",
		"count": 3,
		"additional_labels": map[string]interface{}{
			"label1": "value1",
		},
		"additional_annotations": map[string]interface{}{
			"annotation1": "value1",
			"annotation2": "value2",
		},
		"offering":                "medium-instance",
		"control_plane":           false,
		"control_plane_as_worker": false,
	}

	baseHash := resourceMachinePoolApacheCloudStackHash(baseMachinePool)
	withAnnotationsHash := resourceMachinePoolApacheCloudStackHash(withAnnotations)
	differentAnnotationsHash := resourceMachinePoolApacheCloudStackHash(differentAnnotations)
	moreAnnotationsHash := resourceMachinePoolApacheCloudStackHash(moreAnnotations)

	// Hash should be different when annotations are added
	assert.NotEqual(t, baseHash, withAnnotationsHash, "Adding annotations should change hash")

	// Hash should be different when annotation values change
	assert.NotEqual(t, withAnnotationsHash, differentAnnotationsHash, "Changing annotation values should change hash")

	// Hash should be different when adding more annotations
	assert.NotEqual(t, withAnnotationsHash, moreAnnotationsHash, "Adding more annotations should change hash")

	// Hash should be consistent for same input
	sameHash := resourceMachinePoolApacheCloudStackHash(withAnnotations)
	assert.Equal(t, withAnnotationsHash, sameHash, "Same input should produce same hash")
}

func TestResourceMachinePoolApacheCloudStackHashAllFields(t *testing.T) {
	testCases := []struct {
		name        string
		baseInput   map[string]interface{}
		modifyField func(map[string]interface{})
		description string
	}{
		{
			name: "Offering change affects hash",
			baseInput: map[string]interface{}{
				"name":                    "pool-1",
				"count":                   2,
				"offering":                "small-instance",
				"control_plane":           false,
				"control_plane_as_worker": false,
			},
			modifyField: func(m map[string]interface{}) {
				m["offering"] = "medium-instance"
			},
			description: "Changing offering should change hash",
		},
		{
			name: "Template change affects hash",
			baseInput: map[string]interface{}{
				"name":     "pool-1",
				"count":    2,
				"offering": "small-instance",
				"template": []interface{}{
					map[string]interface{}{
						"id":   "template-123",
						"name": "ubuntu-20.04",
					},
				},
				"control_plane":           false,
				"control_plane_as_worker": false,
			},
			modifyField: func(m map[string]interface{}) {
				m["template"] = []interface{}{
					map[string]interface{}{
						"id":   "template-456",
						"name": "ubuntu-22.04",
					},
				}
			},
			description: "Changing template should change hash",
		},
		{
			name: "Network change affects hash",
			baseInput: map[string]interface{}{
				"name":     "pool-1",
				"count":    2,
				"offering": "small-instance",
				"network": []interface{}{
					map[string]interface{}{
						"network_name": "network-1",
					},
				},
				"control_plane":           false,
				"control_plane_as_worker": false,
			},
			modifyField: func(m map[string]interface{}) {
				m["network"] = []interface{}{
					map[string]interface{}{
						"network_name": "network-2",
					},
				}
			},
			description: "Changing network should change hash",
		},
		{
			name: "Annotation change affects hash",
			baseInput: map[string]interface{}{
				"name":     "pool-1",
				"count":    2,
				"offering": "small-instance",
				"additional_annotations": map[string]interface{}{
					"annotation1": "value1",
				},
				"control_plane":           false,
				"control_plane_as_worker": false,
			},
			modifyField: func(m map[string]interface{}) {
				m["additional_annotations"] = map[string]interface{}{
					"annotation1": "value2",
				}
			},
			description: "Changing annotations should change hash",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Get hash of base input
			baseHash := resourceMachinePoolApacheCloudStackHash(tc.baseInput)

			// Create modified copy
			modified := copyMap(tc.baseInput)
			tc.modifyField(modified)

			// Get hash of modified input
			modifiedHash := resourceMachinePoolApacheCloudStackHash(modified)

			// Hashes should be different
			if baseHash == modifiedHash {
				t.Errorf("%s: Base hash %d equals modified hash %d, but they should differ.\nBase: %+v\nModified: %+v",
					tc.description, baseHash, modifiedHash, tc.baseInput, modified)
			}
		})
	}
}

func TestResourceMachinePoolApacheCloudStackHashKubernetesStyleAnnotations(t *testing.T) {
	machinePool := map[string]interface{}{
		"name":  "test-pool",
		"count": 2,
		"additional_annotations": map[string]interface{}{
			"custom.io/annotation":        "value1",
			"company.com/another-annot":   "value2",
			"kubernetes.io/some-metadata": "metadata-value",
		},
		"offering":                "medium-instance",
		"control_plane":           false,
		"control_plane_as_worker": false,
	}

	// Should generate a valid hash with Kubernetes-style annotations
	hash := resourceMachinePoolApacheCloudStackHash(machinePool)
	assert.NotEqual(t, 0, hash, "Hash should be generated for Kubernetes-style annotations")

	// Hash should be consistent
	sameHash := resourceMachinePoolApacheCloudStackHash(machinePool)
	assert.Equal(t, hash, sameHash, "Same input should produce same hash")
}

func TestResourceMachinePoolApacheCloudStackHashOverrideKubeadmConfiguration(t *testing.T) {
	// Base machine pool without override_kubeadm_configuration
	baseMachinePool := map[string]interface{}{
		"name":                    "test-pool",
		"count":                   3,
		"offering":                "medium-instance",
		"control_plane":           false,
		"control_plane_as_worker": false,
	}

	// Machine pool with override_kubeadm_configuration
	withOverride := map[string]interface{}{
		"name":                           "test-pool",
		"count":                          3,
		"offering":                       "medium-instance",
		"control_plane":                  false,
		"control_plane_as_worker":        false,
		"override_kubeadm_configuration": "kubeletExtraArgs:\n  node-labels: custom=value",
	}

	// Machine pool with different override_kubeadm_configuration
	differentOverride := map[string]interface{}{
		"name":                           "test-pool",
		"count":                          3,
		"offering":                       "medium-instance",
		"control_plane":                  false,
		"control_plane_as_worker":        false,
		"override_kubeadm_configuration": "preKubeadmCommands:\n  - echo 'test'",
	}

	baseHash := resourceMachinePoolApacheCloudStackHash(baseMachinePool)
	withOverrideHash := resourceMachinePoolApacheCloudStackHash(withOverride)
	differentOverrideHash := resourceMachinePoolApacheCloudStackHash(differentOverride)

	// Hash should be different when override_kubeadm_configuration is added
	assert.NotEqual(t, baseHash, withOverrideHash, "Adding override_kubeadm_configuration should change hash")

	// Hash should be different when override_kubeadm_configuration values change
	assert.NotEqual(t, withOverrideHash, differentOverrideHash, "Changing override_kubeadm_configuration should change hash")

	// Hash should be consistent for same input
	sameHash := resourceMachinePoolApacheCloudStackHash(withOverride)
	assert.Equal(t, withOverrideHash, sameHash, "Same input should produce same hash")
}

func TestResourceMachinePoolApacheCloudStackHashOverrideKubeadmEmptyString(t *testing.T) {
	// Machine pool with empty override_kubeadm_configuration should be same as no override
	poolWithEmptyOverride := map[string]interface{}{
		"name":                           "test-pool",
		"count":                          3,
		"offering":                       "medium-instance",
		"control_plane":                  false,
		"control_plane_as_worker":        false,
		"override_kubeadm_configuration": "",
	}

	poolWithoutOverride := map[string]interface{}{
		"name":                    "test-pool",
		"count":                   3,
		"offering":                "medium-instance",
		"control_plane":           false,
		"control_plane_as_worker": false,
	}

	emptyOverrideHash := resourceMachinePoolApacheCloudStackHash(poolWithEmptyOverride)
	withoutOverrideHash := resourceMachinePoolApacheCloudStackHash(poolWithoutOverride)

	// Empty string should be treated same as no override
	assert.Equal(t, emptyOverrideHash, withoutOverrideHash, "Empty override_kubeadm_configuration should have same hash as no override")
}

func TestToMachinePoolCloudStackOverrideKubeadmConfiguration(t *testing.T) {
	tests := []struct {
		name                           string
		input                          map[string]interface{}
		expectOverrideKubeadmConfigSet bool
		expectedValue                  string
	}{
		{
			name: "Worker pool with override_kubeadm_configuration",
			input: map[string]interface{}{
				"name":                           "worker-pool",
				"count":                          3,
				"offering":                       "medium",
				"control_plane":                  false,
				"control_plane_as_worker":        false,
				"node_repave_interval":           0,
				"override_kubeadm_configuration": "kubeletExtraArgs:\n  node-labels: \"custom=value\"",
			},
			expectOverrideKubeadmConfigSet: true,
			expectedValue:                  "kubeletExtraArgs:\n  node-labels: \"custom=value\"",
		},
		{
			name: "Worker pool without override_kubeadm_configuration",
			input: map[string]interface{}{
				"name":                    "worker-pool",
				"count":                   3,
				"offering":                "medium",
				"control_plane":           false,
				"control_plane_as_worker": false,
				"node_repave_interval":    0,
			},
			expectOverrideKubeadmConfigSet: false,
			expectedValue:                  "",
		},
		{
			name: "Worker pool with empty override_kubeadm_configuration",
			input: map[string]interface{}{
				"name":                           "worker-pool",
				"count":                          3,
				"offering":                       "medium",
				"control_plane":                  false,
				"control_plane_as_worker":        false,
				"node_repave_interval":           0,
				"override_kubeadm_configuration": "",
			},
			expectOverrideKubeadmConfigSet: false,
			expectedValue:                  "",
		},
		{
			name: "Control plane pool with override_kubeadm_configuration should be ignored",
			input: map[string]interface{}{
				"name":                           "control-plane-pool",
				"count":                          3,
				"offering":                       "large",
				"control_plane":                  true,
				"control_plane_as_worker":        false,
				"node_repave_interval":           0,
				"override_kubeadm_configuration": "kubeletExtraArgs:\n  node-labels: \"custom=value\"",
			},
			expectOverrideKubeadmConfigSet: false,
			expectedValue:                  "",
		},
		{
			name: "Worker pool with complex YAML override",
			input: map[string]interface{}{
				"name":                    "worker-pool",
				"count":                   3,
				"offering":                "medium",
				"control_plane":           false,
				"control_plane_as_worker": false,
				"node_repave_interval":    0,
				"override_kubeadm_configuration": `kubeletExtraArgs:
  node-labels: "custom=value,env=prod"
  max-pods: "110"
preKubeadmCommands:
  - echo 'Setting up node'
  - sysctl -w net.ipv4.ip_forward=1
postKubeadmCommands:
  - echo 'Node setup complete'`,
			},
			expectOverrideKubeadmConfigSet: true,
			expectedValue: `kubeletExtraArgs:
  node-labels: "custom=value,env=prod"
  max-pods: "110"
preKubeadmCommands:
  - echo 'Setting up node'
  - sysctl -w net.ipv4.ip_forward=1
postKubeadmCommands:
  - echo 'Node setup complete'`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := toMachinePoolCloudStack(tt.input)
			require.NoError(t, err, "toMachinePoolCloudStack should not return error")
			require.NotNil(t, result, "Result should not be nil")
			require.NotNil(t, result.PoolConfig, "PoolConfig should not be nil")

			if tt.expectOverrideKubeadmConfigSet {
				assert.Equal(t, tt.expectedValue, result.PoolConfig.OverrideKubeadmConfiguration,
					"OverrideKubeadmConfiguration should match expected value")
			} else {
				assert.Empty(t, result.PoolConfig.OverrideKubeadmConfiguration,
					"OverrideKubeadmConfiguration should be empty")
			}
		})
	}
}

func TestFlattenMachinePoolConfigsApacheCloudStackOverrideKubeadmConfiguration(t *testing.T) {
	tests := []struct {
		name                           string
		input                          []*models.V1CloudStackMachinePoolConfig
		expectedPoolName               string
		expectOverrideKubeadmConfigSet bool
		expectedValue                  string
	}{
		{
			name: "Worker pool with override_kubeadm_configuration",
			input: []*models.V1CloudStackMachinePoolConfig{
				{
					V1MachinePoolBaseConfig: models.V1MachinePoolBaseConfig{
						AdditionalLabels:             map[string]string{},
						AdditionalAnnotations:        map[string]string{},
						IsControlPlane:               types.Ptr(false),
						Name:                         "worker-pool",
						Size:                         3,
						OverrideKubeadmConfiguration: "kubeletExtraArgs:\n  node-labels: \"custom=value\"",
						UseControlPlaneAsWorker:      false,
						UpdateStrategy:               &models.V1UpdateStrategy{Type: "RollingUpdateScaleOut"},
					},
					V1CloudStackMachineConfig: models.V1CloudStackMachineConfig{
						Offering: &models.V1CloudStackResource{Name: "medium"},
					},
				},
			},
			expectedPoolName:               "worker-pool",
			expectOverrideKubeadmConfigSet: true,
			expectedValue:                  "kubeletExtraArgs:\n  node-labels: \"custom=value\"",
		},
		{
			name: "Worker pool without override_kubeadm_configuration",
			input: []*models.V1CloudStackMachinePoolConfig{
				{
					V1MachinePoolBaseConfig: models.V1MachinePoolBaseConfig{
						AdditionalLabels:             map[string]string{},
						AdditionalAnnotations:        map[string]string{},
						IsControlPlane:               types.Ptr(false),
						Name:                         "worker-pool",
						Size:                         3,
						OverrideKubeadmConfiguration: "",
						UseControlPlaneAsWorker:      false,
						UpdateStrategy:               &models.V1UpdateStrategy{Type: "RollingUpdateScaleOut"},
					},
					V1CloudStackMachineConfig: models.V1CloudStackMachineConfig{
						Offering: &models.V1CloudStackResource{Name: "medium"},
					},
				},
			},
			expectedPoolName:               "worker-pool",
			expectOverrideKubeadmConfigSet: false,
			expectedValue:                  "",
		},
		{
			name: "Control plane pool with override_kubeadm_configuration should not be set",
			input: []*models.V1CloudStackMachinePoolConfig{
				{
					V1MachinePoolBaseConfig: models.V1MachinePoolBaseConfig{
						AdditionalLabels:             map[string]string{},
						AdditionalAnnotations:        map[string]string{},
						IsControlPlane:               types.Ptr(true),
						Name:                         "control-plane-pool",
						Size:                         3,
						OverrideKubeadmConfiguration: "kubeletExtraArgs:\n  node-labels: \"custom=value\"",
						UseControlPlaneAsWorker:      false,
						UpdateStrategy:               &models.V1UpdateStrategy{Type: "RollingUpdateScaleOut"},
					},
					V1CloudStackMachineConfig: models.V1CloudStackMachineConfig{
						Offering: &models.V1CloudStackResource{Name: "large"},
					},
				},
			},
			expectedPoolName:               "control-plane-pool",
			expectOverrideKubeadmConfigSet: false,
			expectedValue:                  "",
		},
		{
			name: "Worker pool with complex YAML override",
			input: []*models.V1CloudStackMachinePoolConfig{
				{
					V1MachinePoolBaseConfig: models.V1MachinePoolBaseConfig{
						AdditionalLabels:      map[string]string{},
						AdditionalAnnotations: map[string]string{},
						IsControlPlane:        types.Ptr(false),
						Name:                  "worker-pool",
						Size:                  3,
						OverrideKubeadmConfiguration: `kubeletExtraArgs:
  node-labels: "env=prod"
  max-pods: "110"
preKubeadmCommands:
  - echo 'Setup complete'`,
						UseControlPlaneAsWorker: false,
						UpdateStrategy:          &models.V1UpdateStrategy{Type: "RollingUpdateScaleOut"},
					},
					V1CloudStackMachineConfig: models.V1CloudStackMachineConfig{
						Offering: &models.V1CloudStackResource{Name: "medium"},
					},
				},
			},
			expectedPoolName:               "worker-pool",
			expectOverrideKubeadmConfigSet: true,
			expectedValue: `kubeletExtraArgs:
  node-labels: "env=prod"
  max-pods: "110"
preKubeadmCommands:
  - echo 'Setup complete'`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := flattenMachinePoolConfigsApacheCloudStack(tt.input)
			require.NotNil(t, result, "Result should not be nil")
			require.Len(t, result, 1, "Should have exactly one machine pool")

			pool := result[0].(map[string]interface{})
			assert.Equal(t, tt.expectedPoolName, pool["name"], "Pool name should match")

			if tt.expectOverrideKubeadmConfigSet {
				value, exists := pool["override_kubeadm_configuration"]
				assert.True(t, exists, "override_kubeadm_configuration should exist in flattened data")
				assert.Equal(t, tt.expectedValue, value, "override_kubeadm_configuration value should match")
			} else {
				_, exists := pool["override_kubeadm_configuration"]
				assert.False(t, exists, "override_kubeadm_configuration should not exist in flattened data")
			}
		})
	}
}

func TestOverrideKubeadmConfigurationRoundTrip(t *testing.T) {
	// Test that data survives a round trip from Terraform -> API model -> Terraform
	originalInput := map[string]interface{}{
		"name":                    "worker-pool",
		"count":                   3,
		"offering":                "medium",
		"control_plane":           false,
		"control_plane_as_worker": false,
		"node_repave_interval":    0,
		"override_kubeadm_configuration": `kubeletExtraArgs:
  node-labels: "env=production,tier=frontend"
  max-pods: "110"
preKubeadmCommands:
  - echo 'Starting node setup'
  - sysctl -w net.ipv4.ip_forward=1
postKubeadmCommands:
  - echo 'Node setup complete'
  - systemctl restart kubelet`,
	}

	// Convert to API model (CREATE)
	apiModel, err := toMachinePoolCloudStack(originalInput)
	require.NoError(t, err)
	require.NotNil(t, apiModel)
	require.NotNil(t, apiModel.PoolConfig)

	// Simulate API response (READ)
	apiResponse := &models.V1CloudStackMachinePoolConfig{
		V1MachinePoolBaseConfig: models.V1MachinePoolBaseConfig{
			AdditionalLabels:             map[string]string{},
			AdditionalAnnotations:        map[string]string{},
			IsControlPlane:               types.Ptr(false),
			Name:                         "worker-pool",
			Size:                         3,
			OverrideKubeadmConfiguration: apiModel.PoolConfig.OverrideKubeadmConfiguration,
			UseControlPlaneAsWorker:      false,
			UpdateStrategy:               &models.V1UpdateStrategy{Type: "RollingUpdateScaleOut"},
		},
		V1CloudStackMachineConfig: models.V1CloudStackMachineConfig{
			Offering: &models.V1CloudStackResource{Name: "medium"},
		},
	}

	// Flatten back to Terraform state (READ)
	flattened := flattenMachinePoolConfigsApacheCloudStack([]*models.V1CloudStackMachinePoolConfig{apiResponse})
	require.Len(t, flattened, 1)

	pool := flattened[0].(map[string]interface{})
	flattenedValue, exists := pool["override_kubeadm_configuration"]
	require.True(t, exists, "override_kubeadm_configuration should exist after round trip")

	// Verify the value matches the original
	assert.Equal(t, originalInput["override_kubeadm_configuration"], flattenedValue,
		"override_kubeadm_configuration should match original after round trip")
}
