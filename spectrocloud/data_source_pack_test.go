package spectrocloud

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-sdk-go/api/models"
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

func TestGetLatestVersion(t *testing.T) {
	t.Run("valid versions", func(t *testing.T) {
		versions := []*models.V1RegistryPackMetadata{
			{LatestVersion: "v1.0.0"},
			{LatestVersion: "v1.2.0"},
			{LatestVersion: "v1.1.0"},
		}
		latest, err := getLatestVersion(versions)

		assert.NoError(t, err, "Expected no error")
		assert.Equal(t, "1.2.0", latest, "The latest version should be returned")
	})

	t.Run("empty versions list", func(t *testing.T) {
		versions := []*models.V1RegistryPackMetadata{}
		latest, err := getLatestVersion(versions)

		assert.Error(t, err, "Expected an error for empty versions list")
		assert.Equal(t, "", latest, "No version should be returned")
		assert.Equal(t, "no versions provided", err.Error(), "Expected specific error message")
	})

	t.Run("invalid version string", func(t *testing.T) {
		versions := []*models.V1RegistryPackMetadata{
			{LatestVersion: "1.0.0"},
			{LatestVersion: "invalid-version"},
			{LatestVersion: "1.1.0"},
		}
		latest, err := getLatestVersion(versions)

		assert.Error(t, err, "Expected an error for invalid version string")
		assert.Equal(t, "", latest, "No version should be returned for invalid input")
		assert.Contains(t, err.Error(), "invalid version", "Error message should indicate invalid version")
	})

	t.Run("single version", func(t *testing.T) {
		versions := []*models.V1RegistryPackMetadata{
			{LatestVersion: "2.0.0"},
		}
		latest, err := getLatestVersion(versions)

		assert.NoError(t, err, "Expected no error")
		assert.Equal(t, "2.0.0", latest, "The single version should be returned")
	})

	t.Run("pre-release versions", func(t *testing.T) {
		versions := []*models.V1RegistryPackMetadata{
			{LatestVersion: "1.0.0-alpha"},
			{LatestVersion: "1.0.0-beta"},
			{LatestVersion: "1.0.0"},
		}
		latest, err := getLatestVersion(versions)

		assert.NoError(t, err, "Expected no error")
		assert.Equal(t, "1.0.0", latest, "The stable version should be returned as the latest")
	})
}
