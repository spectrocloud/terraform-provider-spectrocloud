package spectrocloud

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
)

func prepareBaseClusterConfigTemplateDataSourceTestData() *schema.ResourceData {
	d := dataSourceClusterConfigTemplate().TestResourceData()
	_ = d.Set("name", "test-cluster-config-template")
	_ = d.Set("context", "project")
	return d
}

func TestDataSourceClusterConfigTemplateRead(t *testing.T) {
	d := prepareBaseClusterConfigTemplateDataSourceTestData()
	var ctx context.Context
	diags := dataSourceClusterConfigTemplateRead(ctx, d, unitTestMockAPIClient)
	assert.Empty(t, diags)

	// Verify all fields are populated
	assert.Equal(t, "test-cluster-config-template-id", d.Id())
	assert.Equal(t, "test-cluster-config-template", d.Get("name").(string))
	assert.Equal(t, "aws", d.Get("cloud_type").(string))
	assert.Equal(t, "Test cluster config template", d.Get("description").(string))

	// Verify attached_cluster field
	attachedClusters := d.Get("attached_cluster").([]interface{})
	assert.NotNil(t, attachedClusters)
	assert.GreaterOrEqual(t, len(attachedClusters), 0)

	// Verify execution_state field
	executionState := d.Get("execution_state").(string)
	assert.NotEmpty(t, executionState)
	assert.Contains(t, []string{"Pending", "Applied", "Failed", "PartiallyApplied"}, executionState)
}
