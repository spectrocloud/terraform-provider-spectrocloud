package spectrocloud

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
	"testing"
)

func prepareBaseUserResourceData() *schema.ResourceData {
	d := dataSourceUser().TestResourceData()
	err := d.Set("email", "test@spectrocloud.com")
	if err != nil {
		return nil
	}
	return d
}

func TestDataSourceUserRead(t *testing.T) {
	// Initialize ResourceData with a test email
	resourceData := prepareBaseUserResourceData()

	// Call the dataSourceUserRead function
	diags := dataSourceUserRead(context.Background(), resourceData, unitTestMockAPIClient)

	// Assertions
	assert.Equal(t, "12345", resourceData.Id())
	assert.NoError(t, resourceData.Set("email", "test@spectrocloud.com"))
	assert.Empty(t, diags)
}

func TestDataSourceUserNegativeRead(t *testing.T) {
	// Initialize ResourceData with a test email
	resourceData := prepareBaseUserResourceData()

	// Call the dataSourceUserRead function
	diags := dataSourceUserRead(context.Background(), resourceData, unitTestMockAPINegativeClient)

	// Assertions

	if assert.NotEmpty(t, diags, "Expected diags to contain at least one element") {
		assert.Contains(t, diags[0].Summary, "User not found", "The first diagnostic message does not contain the expected error message")
	}
}
