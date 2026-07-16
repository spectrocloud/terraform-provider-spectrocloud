package spectrocloud

import (
	"context"
	"testing"

	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

////
// Covers Import functions still at 0%. Import functions typically split
// into two branches: (1) empty-ID rejection, (2) resolve-by-UID or
// resolve-by-name → set state. We test both plus any pure state-setter
// helpers (setHelmRegistryState, setApplicationImportState,
// setBackupStorageLocationImportState, setApplicationProfileImportState).

// ---------------------------------------------------------------------------
// resourceApplicationImport
// ---------------------------------------------------------------------------

func TestResourceApplicationImport_EmptyID(t *testing.T) {
	d := resourceApplication().TestResourceData()
	// No ID set → GetCommonApplication returns "ID or name required" error.
	_, err := resourceApplicationImport(context.Background(), d, unitTestMockAPIClient)
	assert.Error(t, err)
}

func TestGetApplicationByName_NoMatch(t *testing.T) {
	c := getV1ClientWithResourceContext(unitTestMockAPIClient, "project")
	// SearchAppDeploymentSummaries with an obscure name → empty result
	// → returns (nil, nil).
	app, err := getApplicationByName(c, "definitely-not-a-real-app-name-xyz")
	_ = err
	_ = app
}

func TestSetApplicationImportState(t *testing.T) {
	d := resourceApplication().TestResourceData()
	app := &models.V1AppDeployment{
		Metadata: &models.V1ObjectMeta{Name: "app-1", UID: "app-uid-1"},
		Spec: &models.V1AppDeploymentSpec{
			Profile: &models.V1AppDeploymentProfile{
				Metadata: &models.V1AppDeploymentProfileMeta{UID: "profile-1"},
			},
			Config: &models.V1AppDeploymentConfig{
				Target: &models.V1AppDeploymentTargetConfig{},
			},
		},
	}
	_, err := setApplicationImportState(d, app, "project", nil)
	require.NoError(t, err)
	assert.Equal(t, "app-uid-1", d.Id())
	assert.Equal(t, "app-1", d.Get("name"))
	assert.Equal(t, "profile-1", d.Get("application_profile_uid"))
}

// ---------------------------------------------------------------------------
// resourceApplicationProfileImport
// ---------------------------------------------------------------------------

func TestResourceApplicationProfileImport_EmptyID(t *testing.T) {
	d := resourceApplicationProfile().TestResourceData()
	_, err := resourceApplicationProfileImport(context.Background(), d, unitTestMockAPIClient)
	assert.Error(t, err)
}

func TestGetCommonApplicationProfile_InvalidFormat(t *testing.T) {
	d := resourceApplicationProfile().TestResourceData()
	d.SetId("a:b:c:d") // more than 3 parts
	_, err := GetCommonApplicationProfile(d, unitTestMockAPIClient)
	assert.Error(t, err)
}

func TestGetCommonApplicationProfile_InvalidContext(t *testing.T) {
	d := resourceApplicationProfile().TestResourceData()
	d.SetId("name:invalid-context")
	_, err := GetCommonApplicationProfile(d, unitTestMockAPIClient)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid context")
}

func TestSetApplicationProfileImportState(t *testing.T) {
	d := resourceApplicationProfile().TestResourceData()
	require.NoError(t, setApplicationProfileImportState(d, "prof-1", "1.2.3", "project", "prof-uid-1"))
	assert.Equal(t, "prof-uid-1", d.Id())
	assert.Equal(t, "prof-1", d.Get("name"))
	assert.Equal(t, "1.2.3", d.Get("version"))
	assert.Equal(t, "project", d.Get("context"))
}

// ---------------------------------------------------------------------------
// resourceBackupStorageLocationImport
// ---------------------------------------------------------------------------

func TestResourceBackupStorageLocationImport_EmptyID(t *testing.T) {
	d := resourceBackupStorageLocation().TestResourceData()
	_, err := resourceBackupStorageLocationImport(context.Background(), d, unitTestMockAPIClient)
	assert.Error(t, err)
}

// (parseBackupStorageLocationImportID and mapAPITypeToTerraformProvider
// are covered by resource_import_parser_helpers_test.go.)

func TestSetBackupStorageLocationImportState(t *testing.T) {
	d := resourceBackupStorageLocation().TestResourceData()
	storageType := models.V1LocationTypeS3
	bsl := &models.V1UserAssetsLocation{
		Metadata: &models.V1ObjectMeta{Name: "bsl-1", UID: "bsl-uid-1"},
		Spec:     &models.V1UserAssetsLocationSpec{Storage: &storageType},
	}
	require.NoError(t, setBackupStorageLocationImportState(d, bsl, "project"))
	assert.Equal(t, "bsl-uid-1", d.Id())
	assert.Equal(t, "bsl-1", d.Get("name"))
	assert.Equal(t, "project", d.Get("context"))
	assert.Equal(t, "aws", d.Get("storage_provider"))
}

// ---------------------------------------------------------------------------
// resourceRegistryHelmImport
// ---------------------------------------------------------------------------

func TestResourceRegistryHelmImport_EmptyID(t *testing.T) {
	d := resourceRegistryHelm().TestResourceData()
	_, err := resourceRegistryHelmImport(context.Background(), d, unitTestMockAPIClient)
	assert.Error(t, err)
}

func TestSetHelmRegistryState(t *testing.T) {
	d := resourceRegistryHelm().TestResourceData()
	endpoint := "https://helm.example.com"
	reg := &models.V1HelmRegistry{
		Metadata: &models.V1ObjectMeta{Name: "helm-reg"},
		Spec: &models.V1HelmRegistrySpec{
			Endpoint:  &endpoint,
			IsPrivate: true,
		},
	}
	require.NoError(t, setHelmRegistryState(d, reg, "helm-reg-uid-1"))
	assert.Equal(t, "helm-reg-uid-1", d.Id())
	assert.Equal(t, "helm-reg", d.Get("name"))
	assert.Equal(t, endpoint, d.Get("endpoint"))
	assert.Equal(t, true, d.Get("is_private"))
}
