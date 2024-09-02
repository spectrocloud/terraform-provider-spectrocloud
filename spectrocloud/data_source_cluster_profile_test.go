package spectrocloud

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
	"testing"
)

func prepareBaseDataSourceClusterProfileSchema() *schema.ResourceData {
	d := dataSourceClusterProfile().TestResourceData()
	return d
}

func TestReadClusterProfileFuncName(t *testing.T) {
	d := prepareBaseDataSourceClusterProfileSchema()
	var diags diag.Diagnostics

	var ctx context.Context
	_ = d.Set("context", "project")
	_ = d.Set("name", "test-cluster-profile-1")
	diags = dataSourceClusterProfileRead(ctx, d, unitTestMockAPIClient)
	assert.Equal(t, 0, len(diags))
}
func TestReadClusterProfileFuncId(t *testing.T) {
	d := prepareBaseDataSourceClusterProfileSchema()
	var diags diag.Diagnostics
	var ctx context.Context
	_ = d.Set("context", "project")
	_ = d.Set("id", "test-uid")
	_ = d.Set("name", "test-cluster-profile-1")
	diags = dataSourceClusterProfileRead(ctx, d, unitTestMockAPIClient)
	assert.Equal(t, 0, len(diags))
}

func TestReadClusterProfileFuncPacks(t *testing.T) {
	d := prepareBaseDataSourceClusterProfileSchema()
	var diags diag.Diagnostics
	var ctx context.Context
	_ = d.Set("context", "project")
	_ = d.Set("id", "test-uid")
	_ = d.Set("name", "test-cluster-profile-1")

	var packs []interface{}
	packs = append(packs, map[string]interface{}{
		"name":         "test-pack-1",
		"type":         "spectro",
		"tag":          "v1.0",
		"uid":          "test-uid",
		"registry_uid": "test-registry-uid",
		"values":       "test-values",
		"manifest":     []interface{}{},
	})
	manifest := map[string]string{
		"name":    "packmanifest",
		"content": "manifest-content",
	}
	packs = append(packs, map[string]interface{}{
		"name":         "test-pack-2",
		"type":         "spectro",
		"tag":          "v1.0",
		"uid":          "test-uid",
		"registry_uid": "test-registry-uid",
		"values":       "test-values",
		"manifest":     []interface{}{manifest},
	})
	_ = d.Set("pack", packs)
	diags = dataSourceClusterProfileRead(ctx, d, unitTestMockAPIClient)
	assert.Equal(t, 0, len(diags))
}

func TestReadClusterProfileFuncNameNegative(t *testing.T) {
	d := prepareBaseDataSourceClusterProfileSchema()
	var diags diag.Diagnostics

	var ctx context.Context
	_ = d.Set("context", "project")
	_ = d.Set("name", "test-cluster-profile-1")

	diags = dataSourceClusterProfileRead(ctx, d, unitTestMockAPINegativeClient)
	assertFirstDiagMessage(t, diags, "cluster profile not found")
}
