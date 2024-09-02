package spectrocloud

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
	"testing"
)

func prepareProfileImportTestdata() *schema.ResourceData {
	d := resourceClusterProfileImportFeature().TestResourceData()
	_ = d.Set("import_file", "./resource_cluster_profile_import_feature.go")
	_ = d.Set("context", "project")
	return d
}

func TestResourceClusterProfileImportFeatureCreate(t *testing.T) {
	d := prepareProfileImportTestdata()
	var ctx context.Context
	diags := resourceClusterProfileImportFeatureCreate(ctx, d, unitTestMockAPIClient)
	assert.Empty(t, diags)
	assert.Equal(t, "cluster-profile-import-1", d.Id())
}

func TestResourceClusterProfileImportFeatureRead(t *testing.T) {
	d := prepareProfileImportTestdata()
	var ctx context.Context
	d.SetId("cluster-profile-import-1")
	diags := resourceClusterProfileImportFeatureRead(ctx, d, unitTestMockAPIClient)
	assert.NotEmpty(t, diags)

}

func TestResourceClusterProfileImportFeatureUpdate(t *testing.T) {
	d := prepareProfileImportTestdata()
	var ctx context.Context
	d.SetId("cluster-profile-import-1")
	diags := resourceClusterProfileImportFeatureUpdate(ctx, d, unitTestMockAPIClient)
	assert.Empty(t, diags)

}

func TestResourceClusterProfileImportFeatureDelete(t *testing.T) {
	d := prepareProfileImportTestdata()
	var ctx context.Context
	d.SetId("cluster-profile-import-1")
	diags := resourceClusterProfileImportFeatureDelete(ctx, d, unitTestMockAPIClient)
	assert.Empty(t, diags)

}
