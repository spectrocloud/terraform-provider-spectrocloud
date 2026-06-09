package spectrocloud

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetCommonClusterProfile(t *testing.T) {
	t.Run("uid success", func(t *testing.T) {
		d := resourceClusterProfile().TestResourceData()
		d.SetId("cluster-profile-import-1:project")

		c, err := GetCommonClusterProfile(d, unitTestMockAPIClient)
		require.NoError(t, err)
		require.NotNil(t, c)
		assert.NotEmpty(t, d.Id())
		assert.Equal(t, "project", d.Get("context"))
		assert.Equal(t, "test-cluster-profile-1", d.Get("name"))
	})

	t.Run("version mismatch", func(t *testing.T) {
		d := resourceClusterProfile().TestResourceData()
		d.SetId("cluster-profile-import-1:project:9.9.9")

		_, err := GetCommonClusterProfile(d, unitTestMockAPIClient)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "version mismatch")
	})
}

func TestGetCommonApplicationProfile(t *testing.T) {
	t.Run("empty import id", func(t *testing.T) {
		d := resourceApplicationProfile().TestResourceData()
		d.SetId("")

		_, err := GetCommonApplicationProfile(d, unitTestMockAPIClient)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "import ID is required")
	})

	t.Run("invalid context", func(t *testing.T) {
		d := resourceApplicationProfile().TestResourceData()
		d.SetId("app-profile:invalid")

		_, err := GetCommonApplicationProfile(d, unitTestMockAPIClient)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid context")
	})

	t.Run("uid success", func(t *testing.T) {
		d := resourceApplicationProfile().TestResourceData()
		d.SetId("test-app-profile-id:project:1.0.0")

		_, err := GetCommonApplicationProfile(d, unitTestMockAPIClient)
		require.NoError(t, err)
		assert.Equal(t, "test-app-profile-id", d.Id())
		assert.Equal(t, "test-app-profile", d.Get("name"))
		assert.Equal(t, "project", d.Get("context"))
		assert.Equal(t, "1.0.0", d.Get("version"))
	})
}

func TestGetCommonApplication(t *testing.T) {
	t.Run("empty import id", func(t *testing.T) {
		d := resourceApplication().TestResourceData()
		d.SetId("")

		_, err := GetCommonApplication(d, unitTestMockAPIClient)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "required for import")
	})

	t.Run("uid success", func(t *testing.T) {
		d := resourceApplication().TestResourceData()
		d.SetId("test-app-id")

		_, err := GetCommonApplication(d, unitTestMockAPIClient)
		require.NoError(t, err)
		assert.Equal(t, "test-app-id", d.Id())
		assert.Equal(t, "test-app-deployment", d.Get("name"))
		assert.Equal(t, "test-app-profile-id", d.Get("application_profile_uid"))
	})
}

func TestGetCommonBackupStorageLocation(t *testing.T) {
	t.Run("name success", func(t *testing.T) {
		d := resourceBackupStorageLocation().TestResourceData()
		d.SetId("test-bsl-location:project")

		_, err := GetCommonBackupStorageLocation(d, unitTestMockAPIClient)
		require.NoError(t, err)
		assert.Equal(t, "test-bsl-location", d.Get("name"))
		assert.Equal(t, "test-bsl-location-id", d.Id())
		assert.Equal(t, "project", d.Get("context"))
		assert.Equal(t, "aws", d.Get("storage_provider"))
	})

	t.Run("invalid context", func(t *testing.T) {
		d := resourceBackupStorageLocation().TestResourceData()
		d.SetId("test-backup-location-id:system")

		_, err := GetCommonBackupStorageLocation(d, unitTestMockAPIClient)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid context")
	})
}

func TestResourceProjectImportValidation(t *testing.T) {
	d := resourceProject().TestResourceData()
	d.SetId("")
	_, err := resourceProjectImport(context.Background(), d, unitTestMockAPIClient)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "required")
}

func TestGetCommonClusterGroup(t *testing.T) {
	t.Run("invalid format", func(t *testing.T) {
		d := resourceClusterGroup().TestResourceData()
		d.SetId("only-name")

		_, err := GetCommonClusterGroup(d, unitTestMockAPIClient)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid import ID format")
	})

	t.Run("uid success", func(t *testing.T) {
		d := resourceClusterGroup().TestResourceData()
		d.SetId("test-cg-1:project")

		_, err := GetCommonClusterGroup(d, unitTestMockAPIClient)
		require.NoError(t, err)
		assert.Equal(t, "test-cg-1", d.Id())
		assert.Equal(t, "test-cg", d.Get("name"))
	})
}

func TestResourceAlertImportValidation(t *testing.T) {
	t.Run("invalid id format", func(t *testing.T) {
		d := resourceAlert().TestResourceData()
		d.SetId("missing-component")

		_, err := resourceAlertImport(context.Background(), d, unitTestMockAPIClient)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid import ID format")
	})

	t.Run("unsupported component", func(t *testing.T) {
		d := resourceAlert().TestResourceData()
		d.SetId("testprojectuid:NodeHealth")

		_, err := resourceAlertImport(context.Background(), d, unitTestMockAPIClient)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Only 'ClusterHealth' is supported")
	})
}
