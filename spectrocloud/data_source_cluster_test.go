package spectrocloud

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
)

// Shared schema for cluster datasource tests (match data_source_cluster schema)
var testDataSourceClusterSchema = map[string]*schema.Schema{
	"name":              {Type: schema.TypeString, Required: true},
	"context":           {Type: schema.TypeString, Optional: true, Default: "project"},
	"virtual":           {Type: schema.TypeBool, Optional: true},
	"kube_config":       {Type: schema.TypeString, Computed: true},
	"admin_kube_config": {Type: schema.TypeString, Computed: true},
	"state":             {Type: schema.TypeString, Computed: true},
	"health":            {Type: schema.TypeString, Computed: true},
	"cluster_timezone":  {Type: schema.TypeString, Computed: true},
}

func TestDataSourceClusterRead(t *testing.T) {
	tests := []struct {
		name          string
		resourceData  *schema.ResourceData
		mockClient    interface{}
		expectedError bool
	}{
		{
			name: "Successful read",
			resourceData: schema.TestResourceDataRaw(t, testDataSourceClusterSchema, map[string]interface{}{
				"name":    "test-cluster",
				"context": "project",
				"virtual": false,
			}),
			mockClient:    unitTestMockAPIClient,
			expectedError: false,
		},
		{
			name: "Cluster not found",
			resourceData: schema.TestResourceDataRaw(t, testDataSourceClusterSchema, map[string]interface{}{
				"name":    "missing-cluster",
				"context": "project",
				"virtual": false,
			}),
			mockClient:    unitTestMockAPINegativeClient,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := tt.resourceData
			diags := dataSourceClusterRead(context.Background(), d, tt.mockClient)

			if tt.expectedError {
				assert.NotEmpty(t, diags)
				return
			}
			assert.Empty(t, diags)
			assert.Equal(t, "test-cluster-id", d.Id())
			assert.Equal(t, "UTC", d.Get("cluster_timezone"))
		})
	}
}
