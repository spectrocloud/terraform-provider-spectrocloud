package spectrocloud

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
)

func prepareBaseClusterConfigPolicyTestData() *schema.ResourceData {
	d := resourceClusterConfigPolicy().TestResourceData()
	_ = d.Set("name", "test-cluster-config-policy")
	_ = d.Set("context", "project")
	tags := schema.NewSet(schema.HashString, []interface{}{
		"env:production",
		"team:devops",
	})
	_ = d.Set("tags", tags)

	// Create schedules set
	schedulesSet := schema.NewSet(resourceClusterConfigPolicyScheduleHash, []interface{}{
		map[string]interface{}{
			"name":         "weekly-maintenance",
			"start_cron":   "0 2 * * SUN",
			"duration_hrs": 4,
		},
	})
	_ = d.Set("schedules", schedulesSet)

	d.SetId("test-cluster-config-policy-id")
	return d
}

func TestResourceClusterConfigPolicyCreate(t *testing.T) {
	d := prepareBaseClusterConfigPolicyTestData()
	var ctx context.Context
	diags := resourceClusterConfigPolicyCreate(ctx, d, unitTestMockAPIClient)
	assert.Empty(t, diags)
	assert.Equal(t, "test-cluster-config-policy-id", d.Id())
}

func TestResourceClusterConfigPolicyRead(t *testing.T) {
	d := prepareBaseClusterConfigPolicyTestData()
	var ctx context.Context
	diags := resourceClusterConfigPolicyRead(ctx, d, unitTestMockAPIClient)
	assert.Empty(t, diags)
	assert.Equal(t, "test-cluster-config-policy-id", d.Id())
}

func TestResourceClusterConfigPolicyUpdate(t *testing.T) {
	d := prepareBaseClusterConfigPolicyTestData()
	var ctx context.Context
	diags := resourceClusterConfigPolicyUpdate(ctx, d, unitTestMockAPIClient)
	assert.Empty(t, diags)
	assert.Equal(t, "test-cluster-config-policy-id", d.Id())
}

func TestResourceClusterConfigPolicyDelete(t *testing.T) {
	d := prepareBaseClusterConfigPolicyTestData()
	var ctx context.Context
	diags := resourceClusterConfigPolicyDelete(ctx, d, unitTestMockAPIClient)
	assert.Empty(t, diags)
}

func TestExpandClusterConfigPolicySchedules(t *testing.T) {
	schedules := []interface{}{
		map[string]interface{}{
			"name":         "daily-maintenance",
			"start_cron":   "0 2 * * *",
			"duration_hrs": 2,
		},
		map[string]interface{}{
			"name":         "weekly-maintenance",
			"start_cron":   "0 3 * * SUN",
			"duration_hrs": 6,
		},
	}

	result := expandClusterConfigPolicySchedules(schedules)
	assert.NotNil(t, result)
	assert.Len(t, result, 2)
	assert.Equal(t, "daily-maintenance", *result[0].Name)
	assert.Equal(t, "0 2 * * *", *result[0].StartCron)
	assert.Equal(t, int64(2), *result[0].DurationHrs)
	assert.Equal(t, "weekly-maintenance", *result[1].Name)
	assert.Equal(t, "0 3 * * SUN", *result[1].StartCron)
	assert.Equal(t, int64(6), *result[1].DurationHrs)
}
