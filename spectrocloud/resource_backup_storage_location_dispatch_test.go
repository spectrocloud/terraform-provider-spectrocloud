package spectrocloud

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

////
// resourceBackupStorageLocationCreate and resourceBackupStorageLocationUpdate
// are pure dispatch functions on `storage_provider`. Batch 5 already
// covers the individual S3/Minio/GCP/Azure funcs; here we simply call
// the top-level dispatch entry-point for each provider so the switch
// branches are executed.

func TestResourceBackupStorageLocationCreate_Dispatch(t *testing.T) {
	cases := []struct {
		name     string
		provider string
		payload  map[string]interface{}
	}{
		{"AWS", StorageProviderAWS, map[string]interface{}{
			"credential_type": "secret",
			"access_key":      "AK",
			"secret_key":      "SK",
			"s3_url":          "https://s3.us-east-1.amazonaws.com",
		}},
		{"Minio", StorageProviderMinio, map[string]interface{}{
			"credential_type": "secret",
			"access_key":      "AK",
			"secret_key":      "SK",
		}},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			d := resourceBackupStorageLocation().TestResourceData()
			prepareBSLCommonFields(t, d, "batch19-"+tc.name)
			require.NoError(t, d.Set("storage_provider", tc.provider))
			require.NoError(t, d.Set("s3", []interface{}{tc.payload}))

			diags := resourceBackupStorageLocationCreate(context.Background(), d, unitTestMockAPIClient)
			assert.False(t, diags.HasError(), "diags: %+v", diags)
		})
	}
}

func TestResourceBackupStorageLocationUpdate_Dispatch(t *testing.T) {
	cases := []struct {
		name     string
		provider string
		payload  map[string]interface{}
	}{
		{"AWS", StorageProviderAWS, map[string]interface{}{
			"credential_type": "secret", "access_key": "AK", "secret_key": "SK",
			"s3_url": "https://s3.us-east-1.amazonaws.com",
		}},
		{"Minio", StorageProviderMinio, map[string]interface{}{
			"credential_type": "secret", "access_key": "AK", "secret_key": "SK",
		}},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			d := resourceBackupStorageLocation().TestResourceData()
			prepareBSLCommonFields(t, d, "batch19-"+tc.name)
			require.NoError(t, d.Set("storage_provider", tc.provider))
			require.NoError(t, d.Set("s3", []interface{}{tc.payload}))
			d.SetId("test-backup-location-id")

			diags := resourceBackupStorageLocationUpdate(context.Background(), d, unitTestMockAPIClient)
			assert.False(t, diags.HasError(), "diags: %+v", diags)
		})
	}
}

// TestResourceBackupStorageLocationCreate_Default — an unknown provider
// falls through the switch default to the S3 handler.
func TestResourceBackupStorageLocationCreate_Default(t *testing.T) {
	d := resourceBackupStorageLocation().TestResourceData()
	prepareBSLCommonFields(t, d, "batch19-unknown")
	require.NoError(t, d.Set("storage_provider", "unknown-provider"))
	require.NoError(t, d.Set("s3", []interface{}{
		map[string]interface{}{
			"credential_type": "secret",
			"access_key":      "AK",
			"secret_key":      "SK",
			"s3_url":          "https://s3.us-east-1.amazonaws.com",
		},
	}))

	_ = resourceBackupStorageLocationCreate(context.Background(), d, unitTestMockAPIClient)
}
