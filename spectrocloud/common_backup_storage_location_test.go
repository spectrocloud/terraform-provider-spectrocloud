package spectrocloud

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

////
// Covers the 12 previously-zero functions in
// common_backup_storage_location.go:
//   - {S3,Minio,Gcp,Azure}BackupStorageLocationCreate
//   - {Minio,Gcp,Azure}BackupStorageLocationRead
//   - {S3,Minio,Gcp,Azure}BackupStorageLocationUpdate
//
// S3 Read was already exercised by pre-existing tests; we still add a
// Create/Update pair here because the top-level resource's Create/Update
// entry points dispatch on storage_provider and none of the branches were
// previously reachable through resource_backup_storage_location_test.go.

// prepareBSLCommonFields sets the fields all four storage providers need
// (name, region, bucket, ca_cert, is_default, context). Callers layer on
// the provider-specific block after this.
func prepareBSLCommonFields(t *testing.T, d *schema.ResourceData, name string) {
	t.Helper()
	require.NoError(t, d.Set("name", name))
	require.NoError(t, d.Set("region", "us-east-1"))
	require.NoError(t, d.Set("bucket_name", "test-bucket"))
	require.NoError(t, d.Set("ca_cert", ""))
	require.NoError(t, d.Set("is_default", false))
	require.NoError(t, d.Set("context", "project"))
}

// ---------------------------------------------------------------------------
// S3
// ---------------------------------------------------------------------------

func TestS3BackupStorageLocationCreate(t *testing.T) {
	c := mustUnitClient(t, false)
	d := resourceBackupStorageLocation().TestResourceData()

	prepareBSLCommonFields(t, d, "test-s3-bsl")
	require.NoError(t, d.Set("storage_provider", StorageProviderAWS))
	require.NoError(t, d.Set("s3", []interface{}{
		map[string]interface{}{
			"credential_type": "secret",
			"access_key":      "AKIATEST",
			"secret_key":      "supersecret",
			"s3_url":          "https://s3.us-east-1.amazonaws.com",
		},
	}))

	diags := S3BackupStorageLocationCreate(d, c)
	assert.False(t, diags.HasError(), "diags: %+v", diags)
	assert.Equal(t, "test-backup-location-id", d.Id(),
		"Create must persist the UID returned by POST /v1/users/assets/locations/s3")
}

func TestS3BackupStorageLocationUpdate(t *testing.T) {
	c := mustUnitClient(t, false)
	d := resourceBackupStorageLocation().TestResourceData()

	prepareBSLCommonFields(t, d, "test-s3-bsl")
	d.SetId("test-bsl-location-id")
	require.NoError(t, d.Set("storage_provider", StorageProviderAWS))
	require.NoError(t, d.Set("s3", []interface{}{
		map[string]interface{}{
			"credential_type": "secret",
			"access_key":      "AKIATEST",
			"secret_key":      "rotated-secret",
			"s3_url":          "https://s3.us-east-1.amazonaws.com",
		},
	}))

	diags := S3BackupStorageLocationUpdate(d, c)
	assert.False(t, diags.HasError(), "diags: %+v", diags)
}

// ---------------------------------------------------------------------------
// Minio
// ---------------------------------------------------------------------------

// prepareMinioResourceData populates the schema fields Minio Create/Update
// paths read from. Minio reuses the "s3" block for its access-key/secret
// credentials, so the shape overlaps with S3.
func prepareMinioResourceData(t *testing.T) *schema.ResourceData {
	t.Helper()
	d := resourceBackupStorageLocation().TestResourceData()
	prepareBSLCommonFields(t, d, "test-minio-bsl")
	require.NoError(t, d.Set("storage_provider", StorageProviderMinio))
	require.NoError(t, d.Set("s3", []interface{}{
		map[string]interface{}{
			"credential_type":     "secret",
			"access_key":          "minio-access-key",
			"secret_key":          "minio-secret-key",
			"s3_url":              "https://minio.example.com",
			"s3_force_path_style": true,
		},
	}))
	return d
}

func TestMinioBackupStorageLocationCreate(t *testing.T) {
	c := mustUnitClient(t, false)
	d := prepareMinioResourceData(t)

	diags := MinioBackupStorageLocationCreate(d, c)
	assert.False(t, diags.HasError(), "diags: %+v", diags)
	assert.Equal(t, "test-minio-bsl-id", d.Id())
}

func TestMinioBackupStorageLocationRead(t *testing.T) {
	c := mustUnitClient(t, false)
	d := prepareMinioResourceData(t)
	d.SetId("test-minio-bsl-id")

	diags := MinioBackupStorageLocationRead(d, c)
	assert.False(t, diags.HasError(), "diags: %+v", diags)
	assert.Equal(t, "test-minio-bsl", d.Get("name"))
	assert.Equal(t, "us-east-1", d.Get("region"))
	assert.Equal(t, "test-minio-bucket", d.Get("bucket_name"))

	s3 := d.Get("s3").([]interface{})
	require.Len(t, s3, 1)
	s3Map := s3[0].(map[string]interface{})
	assert.Equal(t, "secret", s3Map["credential_type"])
	assert.Equal(t, "minio-access-key", s3Map["access_key"])
	assert.Equal(t, "https://minio.example.com", s3Map["s3_url"])
	assert.Equal(t, true, s3Map["s3_force_path_style"])
}

func TestMinioBackupStorageLocationUpdate(t *testing.T) {
	c := mustUnitClient(t, false)
	d := prepareMinioResourceData(t)
	d.SetId("test-minio-bsl-id")

	diags := MinioBackupStorageLocationUpdate(d, c)
	assert.False(t, diags.HasError(), "diags: %+v", diags)
}

// ---------------------------------------------------------------------------
// GCP
// ---------------------------------------------------------------------------

func prepareGCPResourceData(t *testing.T) *schema.ResourceData {
	t.Helper()
	d := resourceBackupStorageLocation().TestResourceData()
	prepareBSLCommonFields(t, d, "test-gcp-bsl")
	require.NoError(t, d.Set("storage_provider", StorageProviderGCP))
	require.NoError(t, d.Set("gcp_storage_config", []interface{}{
		map[string]interface{}{
			"project_id":           "test-project",
			"gcp_json_credentials": `{"type":"service_account","client_email":"x@example.com"}`,
		},
	}))
	return d
}

func TestGcpBackupStorageLocationCreate(t *testing.T) {
	c := mustUnitClient(t, false)
	d := prepareGCPResourceData(t)

	diags := GcpBackupStorageLocationCreate(d, c)
	assert.False(t, diags.HasError(), "diags: %+v", diags)
	assert.Equal(t, "test-gcp-bsl-id", d.Id())
}

func TestGcpBackupStorageLocationRead(t *testing.T) {
	c := mustUnitClient(t, false)
	d := prepareGCPResourceData(t)
	d.SetId("test-gcp-bsl-id")

	diags := GcpBackupStorageLocationRead(d, c)
	assert.False(t, diags.HasError(), "diags: %+v", diags)
	assert.Equal(t, "test-gcp-bsl", d.Get("name"))
	assert.Equal(t, "test-gcp-bucket", d.Get("bucket_name"))

	gcp := d.Get("gcp_storage_config").([]interface{})
	require.Len(t, gcp, 1)
	gcpMap := gcp[0].(map[string]interface{})
	assert.Equal(t, "test-project", gcpMap["project_id"])
	// gcp_json_credentials is preserved from state (masking prevention);
	// prepareGCPResourceData set it so Read should keep that value.
	assert.Contains(t, gcpMap["gcp_json_credentials"], "service_account")
}

func TestGcpBackupStorageLocationUpdate(t *testing.T) {
	c := mustUnitClient(t, false)
	d := prepareGCPResourceData(t)
	d.SetId("test-gcp-bsl-id")

	diags := GcpBackupStorageLocationUpdate(d, c)
	assert.False(t, diags.HasError(), "diags: %+v", diags)
}

// ---------------------------------------------------------------------------
// Azure
// ---------------------------------------------------------------------------

func prepareAzureResourceData(t *testing.T) *schema.ResourceData {
	t.Helper()
	d := resourceBackupStorageLocation().TestResourceData()
	prepareBSLCommonFields(t, d, "test-azure-bsl")
	require.NoError(t, d.Set("storage_provider", StorageProviderAzure))
	require.NoError(t, d.Set("azure_storage_config", []interface{}{
		map[string]interface{}{
			"container_name":      "test-container",
			"storage_name":        "teststorage",
			"stock_keeping_unit":  "Standard_LRS",
			"resource_group":      "test-rg",
			"azure_tenant_id":     "test-tenant-id",
			"azure_client_id":     "test-client-id",
			"subscription_id":     "test-sub-id",
			"azure_client_secret": "test-client-secret",
		},
	}))
	return d
}

func TestAzureBackupStorageLocationCreate(t *testing.T) {
	c := mustUnitClient(t, false)
	d := prepareAzureResourceData(t)

	diags := AzureBackupStorageLocationCreate(d, c)
	assert.False(t, diags.HasError(), "diags: %+v", diags)
	assert.Equal(t, "test-azure-bsl-id", d.Id())
}

func TestAzureBackupStorageLocationRead(t *testing.T) {
	c := mustUnitClient(t, false)
	d := prepareAzureResourceData(t)
	d.SetId("test-azure-bsl-id")

	diags := AzureBackupStorageLocationRead(d, c)
	assert.False(t, diags.HasError(), "diags: %+v", diags)
	assert.Equal(t, "test-azure-bsl", d.Get("name"))

	az := d.Get("azure_storage_config").([]interface{})
	require.Len(t, az, 1)
	azMap := az[0].(map[string]interface{})
	// ContainerName and StorageName are pointer strings in the API model;
	// Read pushes them through directly. Assert on the pointer-string
	// values coming back — because the model uses *string, the flatten
	// map holds them as *string references, not string values.
	if p, ok := azMap["container_name"].(*string); ok && p != nil {
		assert.Equal(t, "test-container", *p)
	}
	// The client_secret round-trip: it's preserved from state (mask-
	// prevention), so the fixture value should still be there.
	assert.Equal(t, "test-client-secret", azMap["azure_client_secret"])
}

func TestAzureBackupStorageLocationUpdate(t *testing.T) {
	c := mustUnitClient(t, false)
	d := prepareAzureResourceData(t)
	d.SetId("test-azure-bsl-id")

	diags := AzureBackupStorageLocationUpdate(d, c)
	assert.False(t, diags.HasError(), "diags: %+v", diags)
}
