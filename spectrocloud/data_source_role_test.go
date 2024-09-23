package spectrocloud

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
	"testing"
)

func prepareBaseRoleResourceData() *schema.ResourceData {
	d := dataSourceRole().TestResourceData()
	err := d.Set("name", "test-role")
	if err != nil {
		return nil
	}
	return d
}

func TestDataSourceRoleRead(t *testing.T) {
	d := prepareBaseRoleResourceData()
	diags := dataSourceRoleRead(context.Background(), d, unitTestMockAPIClient)
	assert.Empty(t, diags)
}

func TestDataSourceRoleErrorRead(t *testing.T) {
	d := prepareBaseRoleResourceData()
	_ = d.Set("name", "test-role-invalid")
	diags := dataSourceRoleRead(context.Background(), d, unitTestMockAPIClient)
	assertFirstDiagMessage(t, diags, "role 'test-role-invalid' not found")
}

func TestDataSourceRoleNegativeRead(t *testing.T) {
	d := prepareBaseRoleResourceData()
	diags := dataSourceRoleRead(context.Background(), d, unitTestMockAPINegativeClient)
	assertFirstDiagMessage(t, diags, "No roles are found")
}
