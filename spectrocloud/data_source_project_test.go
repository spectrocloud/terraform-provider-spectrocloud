package spectrocloud

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDataSourceProjectRead(t *testing.T) {
	d := schema.TestResourceDataRaw(t, map[string]*schema.Schema{
		"name": {
			Type:     schema.TypeString,
			Required: true,
		},
	}, map[string]interface{}{
		"name": "Default",
	})

	diags := dataSourceProjectRead(context.Background(), d, unitTestMockAPIClient)

	assert.Empty(t, diags)
	assert.Equal(t, "Default", d.Get("name"))
}

func TestDataSourceProjectReadById(t *testing.T) {
	d := dataSourceProject().TestResourceData()
	_ = d.Set("id", "testprojectuid")

	diags := dataSourceProjectRead(context.Background(), d, unitTestMockAPIClient)

	assert.Empty(t, diags)
	assert.Equal(t, "testprojectuid", d.Id())
	assert.Equal(t, "Default", d.Get("name"))
}

func TestDataSourceProjectReadByIdNegative(t *testing.T) {
	d := dataSourceProject().TestResourceData()
	_ = d.Set("id", "testprojectuid")

	diags := dataSourceProjectRead(context.Background(), d, unitTestMockAPINegativeClient)

	assert.NotEmpty(t, diags)
}
