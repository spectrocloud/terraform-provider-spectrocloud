package spectrocloud

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExpandOverrideHealthCheckConfiguration(t *testing.T) {
	t.Parallel()

	poolConfig := &models.V1MachinePoolConfigEntity{}
	expandOverrideHealthCheckConfiguration(map[string]interface{}{
		"override_health_check_configuration": "maxUnhealthy: 40%",
	}, poolConfig)
	assert.Equal(t, "maxUnhealthy: 40%", poolConfig.OverrideHealthCheckConfiguration)

	poolConfig = &models.V1MachinePoolConfigEntity{}
	expandOverrideHealthCheckConfiguration(map[string]interface{}{
		"override_health_check_configuration": "",
	}, poolConfig)
	assert.Empty(t, poolConfig.OverrideHealthCheckConfiguration)

	expandOverrideHealthCheckConfiguration(map[string]interface{}{}, nil)
}

func TestFlattenOverrideHealthCheckConfiguration(t *testing.T) {
	t.Parallel()

	oi := map[string]interface{}{}
	flattenOverrideHealthCheckConfiguration("maxUnhealthy: 40%", oi)
	assert.Equal(t, "maxUnhealthy: 40%", oi["override_health_check_configuration"])

	oi = map[string]interface{}{}
	flattenOverrideHealthCheckConfiguration("", oi)
	_, exists := oi["override_health_check_configuration"]
	assert.False(t, exists)
}

func TestResourceMachinePoolAwsHashOverrideHealthCheckConfiguration(t *testing.T) {
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
		"override_health_check_configuration": "maxUnhealthy: 40%",
	}

	assert.NotEqual(t, resourceMachinePoolAwsHash(base), resourceMachinePoolAwsHash(withOverride))
}

func TestAppendOverrideHealthCheckConfigurationCreateWarnings(t *testing.T) {
	t.Parallel()

	d := schema.TestResourceDataRaw(t, resourceClusterAws().Schema, map[string]interface{}{
		"name":             "test-aws",
		"context":          "project",
		"cloud_account_id": "account-uid",
		"cloud_config": []interface{}{
			map[string]interface{}{
				"region": "us-east-1",
				"vpc_id": "vpc-123",
			},
		},
		"cluster_profile": []interface{}{
			map[string]interface{}{
				"id": "profile-uid",
			},
		},
		"machine_pool": []interface{}{
			map[string]interface{}{
				"name":                                "worker",
				"count":                               1,
				"instance_type":                       "t3.medium",
				"control_plane":                       false,
				"override_health_check_configuration": "maxUnhealthy: 40%",
			},
		},
	})

	var diags diag.Diagnostics
	appendOverrideHealthCheckConfigurationCreateWarnings(d, &diags)
	require.Len(t, diags, 1)
	assert.Equal(t, diag.Warning, diags[0].Severity)
	assert.Contains(t, diags[0].Detail, `Machine pool "worker":`)
	assert.Contains(t, diags[0].Detail, overrideHealthCheckConfigurationRepaveWarning)
}

func TestAppendHealthCheckRepaveWarning(t *testing.T) {
	t.Parallel()

	var diags diag.Diagnostics
	appendHealthCheckRepaveWarning(&diags, "worker")
	require.Len(t, diags, 1)
	assert.Equal(t, diag.Warning, diags[0].Severity)
	assert.Equal(t, `Machine pool "worker": `+overrideHealthCheckConfigurationRepaveWarning, diags[0].Detail)
}
