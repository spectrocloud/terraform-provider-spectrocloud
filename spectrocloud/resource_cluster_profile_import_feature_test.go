package spectrocloud

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
)

func prepareProfileImportTestdata() *schema.ResourceData {
	d := resourceClusterProfileImportFeature().TestResourceData()
	_ = d.Set("import_file", "./resource_cluster_profile_import_feature.go")
	_ = d.Set("context", "project")
	return d
}

func TestResourceClusterProfileImportFeatureCRUD(t *testing.T) {
	testResourceCRUD(t, prepareProfileImportTestdata, unitTestMockAPIClient,
		resourceClusterProfileImportFeatureCreate, resourceClusterProfileImportFeatureRead, resourceClusterProfileImportFeatureUpdate, resourceClusterProfileImportFeatureDelete)
}

func TestResourceClusterProfileImportFeatureReadNegative(t *testing.T) {
	d := prepareProfileImportTestdata()
	d.SetId("cluster-profile-import-1")
	diags := resourceClusterProfileImportFeatureRead(context.Background(), d, unitTestMockAPINegativeClient)
	assert.NotEmpty(t, diags)
}
