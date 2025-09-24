package spectrocloud

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
)

func TestDataSourceClusterRead(t *testing.T) {
	tests := []struct {
		name          string
		resourceData  *schema.ResourceData
		expectedError bool
		expectedDiags diag.Diagnostics
	}{
		{
			name: "Successful read",
			resourceData: schema.TestResourceDataRaw(t, map[string]*schema.Schema{
				"name":              {Type: schema.TypeString, Required: true},
				"context":           {Type: schema.TypeString, Required: true},
				"virtual":           {Type: schema.TypeBool, Optional: true},
				"kube_config":       {Type: schema.TypeString, Computed: true},
				"admin_kube_config": {Type: schema.TypeString, Computed: true},
				"state":             {Type: schema.TypeString, Computed: true},
				"health":            {Type: schema.TypeString, Computed: true},
			}, map[string]interface{}{
				"name":    "test-cluster",
				"context": "some-context",
				"virtual": false,
			}),
			expectedError: false,
			expectedDiags: diag.Diagnostics{},
		},
		{
			name: "Cluster not found",
			resourceData: schema.TestResourceDataRaw(t, map[string]*schema.Schema{
				"name":              {Type: schema.TypeString, Required: true},
				"context":           {Type: schema.TypeString, Required: true},
				"virtual":           {Type: schema.TypeBool, Optional: true},
				"kube_config":       {Type: schema.TypeString, Computed: true},
				"admin_kube_config": {Type: schema.TypeString, Computed: true},
				"state":             {Type: schema.TypeString, Computed: true},
				"health":            {Type: schema.TypeString, Computed: true},
			}, map[string]interface{}{
				"name":              "test-cluster",
				"context":           "some-context",
				"virtual":           false,
				"kube_config":       "",
				"admin_kube_config": "",
			}),
			expectedError: false,
			expectedDiags: diag.Diagnostics{
				diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "Cluster not found",
					Detail:   "The cluster 'test-cluster' was not found.",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			d := tt.resourceData
			var ctx context.Context
			diags := dataSourceClusterRead(ctx, d, unitTestMockAPIClient)

			if tt.expectedError {
				assert.NotEmpty(t, diags)
			} else {
				assert.Empty(t, diags)
			}
			if d.Id() != "" {
				assert.Equal(t, "test-cluster-id", d.Id())
			}
		})
	}
}
