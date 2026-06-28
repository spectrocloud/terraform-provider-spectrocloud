package spectrocloud

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/spectrocloud/terraform-provider-spectrocloud/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const sampleOverrideHealthCheckYAML = `maxUnhealthy: 40%
nodeStartupTimeout: 10m`

func TestMachinePoolHashOverrideHealthCheckMulticloud(t *testing.T) {
	t.Parallel()

	base := map[string]interface{}{
		"name":          "worker",
		"count":         1,
		"instance_type": "t3.medium",
		"capacity_type": "on-demand",
		"max_price":     "",
		"azs":           schema.NewSet(schema.HashString, []interface{}{"us-east-1a"}),
		"az_subnets":    map[string]interface{}{},
	}
	withOverride := map[string]interface{}{
		"name":                                "worker",
		"count":                               1,
		"instance_type":                       "t3.medium",
		"capacity_type":                       "on-demand",
		"max_price":                           "",
		"azs":                                 schema.NewSet(schema.HashString, []interface{}{"us-east-1a"}),
		"az_subnets":                          map[string]interface{}{},
		"override_health_check_configuration": sampleOverrideHealthCheckYAML,
	}

	tests := []struct {
		name     string
		hashFunc func(interface{}) int
		base     map[string]interface{}
		override map[string]interface{}
	}{
		{"Azure", resourceMachinePoolAzureHash, baseAzureHashInput(), withAzureHashInput()},
		{"GCP", resourceMachinePoolGcpHash, baseGcpHashInput(), withGcpHashInput()},
		{"vSphere", resourceMachinePoolVsphereHash, baseVsphereHashInput(), withVsphereHashInput()},
		{"MAAS", resourceMachinePoolMaasHash, baseMaasHashInput(), withMaasHashInput()},
		{"EdgeNative", resourceMachinePoolEdgeNativeHash, baseEdgeNativeHashInput(), withEdgeNativeHashInput()},
		{"AWS", resourceMachinePoolAwsHash, base, withOverride},
		{"CloudStack", resourceMachinePoolApacheCloudStackHash, baseCloudStackHashInput(), withCloudStackHashInput()},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.NotEqual(t, tt.hashFunc(tt.base), tt.hashFunc(tt.override))
		})
	}
}

func TestFlattenOverrideHealthCheckConfigurationMulticloud(t *testing.T) {
	t.Parallel()

	azureResult := flattenMachinePoolConfigsAzure([]*models.V1AzureMachinePoolConfig{{
		Name:                             "worker",
		IsControlPlane:                   types.Ptr(false),
		Size:                             1,
		OverrideHealthCheckConfiguration: sampleOverrideHealthCheckYAML,
	}})
	require.Len(t, azureResult, 1)
	assert.Equal(t, sampleOverrideHealthCheckYAML, azureResult[0].(map[string]interface{})["override_health_check_configuration"])

	gcpResult := flattenMachinePoolConfigsGcp([]*models.V1GcpMachinePoolConfig{{
		Name:                             "worker",
		IsControlPlane:                   types.Ptr(false),
		Size:                             1,
		InstanceType:                     types.Ptr("n1-standard-4"),
		OverrideHealthCheckConfiguration: sampleOverrideHealthCheckYAML,
	}})
	require.Len(t, gcpResult, 1)
	assert.Equal(t, sampleOverrideHealthCheckYAML, gcpResult[0].(map[string]interface{})["override_health_check_configuration"])

	cloudStackResult := flattenMachinePoolConfigsApacheCloudStack([]*models.V1CloudStackMachinePoolConfig{{
		V1MachinePoolBaseConfig: models.V1MachinePoolBaseConfig{
			Name:                             "worker",
			IsControlPlane:                   types.Ptr(false),
			Size:                             1,
			OverrideHealthCheckConfiguration: sampleOverrideHealthCheckYAML,
		},
	}})
	require.Len(t, cloudStackResult, 1)
	assert.Equal(t, sampleOverrideHealthCheckYAML, cloudStackResult[0].(map[string]interface{})["override_health_check_configuration"])
}

func baseAzureHashInput() map[string]interface{} {
	return map[string]interface{}{
		"name":          "worker",
		"count":         1,
		"instance_type": "Standard_D2_v3",
	}
}

func withAzureHashInput() map[string]interface{} {
	m := baseAzureHashInput()
	m["override_health_check_configuration"] = sampleOverrideHealthCheckYAML
	return m
}

func baseGcpHashInput() map[string]interface{} {
	return map[string]interface{}{
		"name":          "worker",
		"count":         1,
		"instance_type": "n1-standard-4",
		"disk_size_gb":  65,
		"azs":           schema.NewSet(schema.HashString, []interface{}{"us-central1-a"}),
	}
}

func withGcpHashInput() map[string]interface{} {
	m := baseGcpHashInput()
	m["override_health_check_configuration"] = sampleOverrideHealthCheckYAML
	return m
}

func baseVsphereHashInput() map[string]interface{} {
	return map[string]interface{}{
		"name":  "worker",
		"count": 1,
		"instance_type": []interface{}{
			map[string]interface{}{"cpu": 2, "disk_size_gb": 50, "memory_mb": 4096},
		},
	}
}

func withVsphereHashInput() map[string]interface{} {
	m := baseVsphereHashInput()
	m["override_health_check_configuration"] = sampleOverrideHealthCheckYAML
	return m
}

func baseMaasHashInput() map[string]interface{} {
	return map[string]interface{}{
		"name":  "worker",
		"count": 1,
		"instance_type": []interface{}{
			map[string]interface{}{
				"min_cpu":       2,
				"min_memory_mb": 4096,
			},
		},
		"azs": schema.NewSet(schema.HashString, []interface{}{"zone-a"}),
		"placement": []interface{}{
			map[string]interface{}{"resource_pool": "default"},
		},
	}
}

func withMaasHashInput() map[string]interface{} {
	m := baseMaasHashInput()
	m["override_health_check_configuration"] = sampleOverrideHealthCheckYAML
	return m
}

func baseEdgeNativeHashInput() map[string]interface{} {
	return map[string]interface{}{
		"name":          "worker",
		"count":         1,
		"instance_type": "edge-native",
		"hosts":         schema.NewSet(schema.HashString, []interface{}{"host-1"}),
	}
}

func withEdgeNativeHashInput() map[string]interface{} {
	m := baseEdgeNativeHashInput()
	m["override_health_check_configuration"] = sampleOverrideHealthCheckYAML
	return m
}

func baseCloudStackHashInput() map[string]interface{} {
	return map[string]interface{}{
		"name":     "worker",
		"count":    1,
		"offering": "medium",
	}
}

func withCloudStackHashInput() map[string]interface{} {
	m := baseCloudStackHashInput()
	m["override_health_check_configuration"] = sampleOverrideHealthCheckYAML
	return m
}

func TestToMachinePoolAzureOverrideHealthCheckConfiguration(t *testing.T) {
	t.Parallel()

	machinePool := map[string]interface{}{
		"name":                    "worker",
		"count":                   2,
		"instance_type":           "Standard_D2_v3",
		"os_type":                 "Linux",
		"control_plane":           false,
		"control_plane_as_worker": false,
		"is_system_node_pool":     false,
		"azs":                     schema.NewSet(schema.HashString, []interface{}{"eastus"}),
		"disk": []interface{}{
			map[string]interface{}{"size_gb": 65, "type": "Standard_LRS"},
		},
		"override_health_check_configuration": sampleOverrideHealthCheckYAML,
	}

	result, err := toMachinePoolAzure(machinePool)
	require.NoError(t, err)
	assert.Equal(t, sampleOverrideHealthCheckYAML, result.PoolConfig.OverrideHealthCheckConfiguration)
}

func TestToMachinePoolCloudStackOverrideHealthCheckConfiguration(t *testing.T) {
	t.Parallel()

	machinePool := map[string]interface{}{
		"name":                                "worker",
		"count":                               2,
		"offering":                            "medium",
		"control_plane":                       false,
		"control_plane_as_worker":             false,
		"override_health_check_configuration": sampleOverrideHealthCheckYAML,
	}

	result, err := toMachinePoolCloudStack(machinePool)
	require.NoError(t, err)
	assert.Equal(t, sampleOverrideHealthCheckYAML, result.PoolConfig.OverrideHealthCheckConfiguration)
}
