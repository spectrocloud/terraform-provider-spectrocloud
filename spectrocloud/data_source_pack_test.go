package spectrocloud

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
	"testing"
)

func prepareBaseDataSourcePackResourceData() *schema.ResourceData {
	d := dataSourcePack().TestResourceData()
	d.SetId("test-pack-1")
	_ = d.Set("type", "manifest")
	return d
}

func TestDataSourcePacksReadManifest(t *testing.T) {
	d := prepareBaseDataSourcePackResourceData()
	diags := dataSourcePackRead(context.Background(), d, unitTestMockAPIClient)
	assert.Empty(t, diags)
}

func TestDataSourcePacksReadOci(t *testing.T) {
	d := prepareBaseDataSourcePackResourceData()
	_ = d.Set("type", "oci")
	_ = d.Set("registry_uid", "test-reg-uid")
	diags := dataSourcePackRead(context.Background(), d, unitTestMockAPIClient)
	assert.Empty(t, diags)
}

func TestDataSourcePacksReadHelm(t *testing.T) {
	d := prepareBaseDataSourcePackResourceData()
	_ = d.Set("type", "helm")
	_ = d.Set("name", "k8")
	_ = d.Set("registry_uid", "test-reg-uid")
	_ = d.Set("filters", "spec.cloudTypes=edge-nativeANDspec.layer=cniANDspec.displayName=CalicoANDspec.version>3.26.9ANDspec.registryUid=${data.spectrocloud_registry.palette_registry_oci.id}")
	diags := dataSourcePackRead(context.Background(), d, unitTestMockAPIClient)
	assert.Empty(t, diags)
}

func TestDataSourcePacksReadHelmMultiPacks(t *testing.T) {
	d := prepareBaseDataSourcePackResourceData()
	_ = d.Set("type", "helm")
	_ = d.Set("name", "k8")
	_ = d.Set("registry_uid", "test-reg-uid")
	_ = d.Set("filters", "spec.cloudTypes=edge-nativeANDspec.layer=cniANDspec.displayName=CalicoANDspec.version>3.26.9ANDspec.registryUid=${data.spectrocloud_registry.palette_registry_oci.id}")
	diags := dataSourcePackRead(context.Background(), d, unitTestMockAPINegativeClient)
	assertFirstDiagMessage(t, diags, "Multiple packs returned")

}
