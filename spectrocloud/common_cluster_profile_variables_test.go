package spectrocloud

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/stretchr/testify/assert"

	schemas "github.com/spectrocloud/terraform-provider-spectrocloud/spectrocloud/schemas"
)

func TestResolveProfileVariableValue(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name        string
		prior       string
		api         string
		isSensitive bool
		expect      string
	}{
		{"empty api keeps prior when not sensitive", "test123", "", false, "test123"},
		{"sensitive keeps prior and ignores masked api value", "test123", "********", true, "test123"},
		{"sensitive without prior yields empty", "", "********", true, ""},
		{"non-sensitive uses masked-looking api value as-is", "old", "********", false, "********"},
		{"cleartext api replaces prior when not sensitive", "old", "new-value", false, "new-value"},
		{"no prior uses cleartext api", "", "only-from-api", false, "only-from-api"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expect, resolveProfileVariableValue(tc.prior, tc.api, tc.isSensitive))
		})
	}
}

func TestProfileVariablesMapFromAPI_ImportExposesSensitiveVariableName(t *testing.T) {
	resourceSchema := map[string]*schema.Schema{
		"cluster_profile": schemas.ClusterProfileSchemaV2(),
	}
	r := schema.Resource{Schema: resourceSchema}
	d := r.TestResourceData()
	d.SetId("cluster-uid")

	apiVars := []*models.V1SpectroClusterVariableResponse{
		{
			Name:        StringPtr("extraVars"),
			Value:       "********",
			IsSensitive: true,
		},
	}

	vars := profileVariablesMapFromAPI(d, "profile-uid-1", apiVars)
	assert.Contains(t, vars, "extraVars")
	assert.Equal(t, "", vars["extraVars"])
}

func TestProfileVariablesMapFromAPI_PreservesMaskedSensitiveValue(t *testing.T) {
	resourceSchema := map[string]*schema.Schema{
		"cluster_profile": schemas.ClusterProfileSchemaV2(),
	}
	r := schema.Resource{Schema: resourceSchema}
	d := r.TestResourceData()
	_ = d.Set("cluster_profile", []interface{}{
		map[string]interface{}{
			"id": "profile-uid-1",
			"variables": map[string]interface{}{
				"extraVars": "test123",
			},
		},
	})
	d.SetId("cluster-uid")

	apiVars := []*models.V1SpectroClusterVariableResponse{
		{
			Name:        StringPtr("extraVars"),
			Value:       "********",
			IsSensitive: true,
		},
		{
			Name:  StringPtr("plainVar"),
			Value: "visible",
		},
	}

	vars := profileVariablesMapFromAPI(d, "profile-uid-1", apiVars)
	assert.Equal(t, "test123", vars["extraVars"])
	assert.Equal(t, "visible", vars["plainVar"])
}

func TestPriorClusterProfileVariable_FromClusterTemplate(t *testing.T) {
	resourceSchema := map[string]*schema.Schema{
		"cluster_template": {
			Type:     schema.TypeList,
			Optional: true,
			MaxItems: 1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"id":              {Type: schema.TypeString, Required: true},
					"cluster_profile": schemas.ClusterProfileSchemaV2(),
				},
			},
		},
	}
	r := schema.Resource{Schema: resourceSchema}
	d := r.TestResourceData()
	profileSet := schema.NewSet(schema.HashResource(schemas.ClusterProfileSchemaV2().Elem.(*schema.Resource)), []interface{}{
		map[string]interface{}{
			"id": "tpl-profile-uid",
			"variables": map[string]interface{}{
				"secret": "from-template-state",
			},
		},
	})
	_ = d.Set("cluster_template", []interface{}{
		map[string]interface{}{
			"id":              "template-uid",
			"cluster_profile": profileSet,
		},
	})

	assert.Equal(t, "from-template-state", priorClusterProfileVariable(d, "tpl-profile-uid", "secret"))
}
