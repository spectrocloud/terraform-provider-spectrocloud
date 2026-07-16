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
			{LatestVersion: "1.0.0"},
			{LatestVersion: "1.2.0"},
			{LatestVersion: "1.1.0"},
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

// ---------------------------------------------------------------------------
// Batch 9 — cover the remaining 0% pure helpers in data_source_pack.go.
// ---------------------------------------------------------------------------

func TestConvertToV1PackType(t *testing.T) {
	set := schema.NewSet(schema.HashString, []interface{}{"spectro", "helm", "oci"})
	got := convertToV1PackType(set)
	assert.Len(t, got, 3)
	assert.ElementsMatch(t, []models.V1PackType{"spectro", "helm", "oci"}, got)

	assert.Empty(t, convertToV1PackType(schema.NewSet(schema.HashString, nil)))
}

func TestConvertToV1PackLayer(t *testing.T) {
	set := schema.NewSet(schema.HashString, []interface{}{"kernel", "os", "k8s"})
	got := convertToV1PackLayer(set)
	assert.Len(t, got, 3)
	assert.ElementsMatch(t, []models.V1PackLayer{"kernel", "os", "k8s"}, got)
}

func TestConvertToAddOnType(t *testing.T) {
	// Non-empty input passes through as strings.
	got := convertToAddOnType([]interface{}{"logging", "monitoring"},
		schema.NewSet(schema.HashString, []interface{}{"addon"}))
	assert.Equal(t, []string{"logging", "monitoring"}, got)

	// Empty input + a pack layer containing "addon" → defaults to AllowedAddonType.
	got = convertToAddOnType(nil,
		schema.NewSet(schema.HashString, []interface{}{"addon"}))
	assert.Equal(t, AllowedAddonType, got)

	// Empty input + non-addon pack layer → stays empty.
	got = convertToAddOnType(nil,
		schema.NewSet(schema.HashString, []interface{}{"k8s"}))
	assert.Empty(t, got)
}

func TestConvertToStringSlice(t *testing.T) {
	assert.Equal(t, []string{"a", "b", "c"}, convertToStringSlice([]interface{}{"a", "b", "c"}))
	assert.Empty(t, convertToStringSlice(nil))
	// Non-string entries produce empty strings in that slot.
	assert.Equal(t, []string{"a", "", "c"}, convertToStringSlice([]interface{}{"a", 42, "c"}))
}

func TestGetLatestPackTag(t *testing.T) {
	t.Run("highest wins", func(t *testing.T) {
		got, err := GetLatestPackTag([]*models.V1PackTags{
			nil,
			{Version: "1.0.0"},
			{Version: "2.5.1"},
			{Version: "1.9.9"},
		})
		assert.NoError(t, err)
		assert.NotNil(t, got)
		if got != nil {
			assert.Equal(t, "2.5.1", got.Version)
		}
	})

	t.Run("invalid semver rejected", func(t *testing.T) {
		_, err := GetLatestPackTag([]*models.V1PackTags{{Version: "not-a-version"}})
		assert.Error(t, err)
	})

	t.Run("empty and all-nil input rejected", func(t *testing.T) {
		_, err := GetLatestPackTag(nil)
		assert.Error(t, err)
		_, err = GetLatestPackTag([]*models.V1PackTags{nil, {Version: ""}})
		assert.Error(t, err)
	})
}

// TestSetLatestPackVersionToFilters exercises the function body via the
// negative mock — SearchPacks won't return exactly one hit, so the
// function returns "". Both the empty-registryUID and populated branches
// are hit.
func TestSetLatestPackVersionToFilters(t *testing.T) {
	c := getV1ClientWithResourceContext(unitTestMockAPINegativeClient, "project")
	assert.Equal(t, "", setLatestPackVersionToFilters("mypack", "", c))
	assert.Equal(t, "", setLatestPackVersionToFilters("mypack", "reg-uid-1", c))
}
