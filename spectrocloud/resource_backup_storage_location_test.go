package spectrocloud

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
	"testing"
)

func prepareResourceBackupStorageLocation() *schema.ResourceData {
	d := resourceBackupStorageLocation().TestResourceData()
	d.SetId("test-backup-location-id")
	_ = d.Set("name", "test-backup-location")
	_ = d.Set("is_default", false)
	_ = d.Set("region", "test-east")
	_ = d.Set("bucket_name", "test-bucket")
	_ = d.Set("ca_cert", "test-cert")
	s3 := make([]interface{}, 0)
	s3 = append(s3, map[string]interface{}{
		"s3_url":              "s3://test/test",
		"s3_force_path_style": false,
		"credential_type":     "secret",
		"access_key":          "test-access-key",
		"secret_key":          "test-secret-key",
		"arn":                 "test-arn",
		"external_id":         "test-external-id",
	})
	_ = d.Set("s3", s3)

	return d
}

func TestResourceBackupStorageLocationCreate(t *testing.T) {
	ctx := context.Background()
	d := prepareResourceBackupStorageLocation()
	diags := resourceBackupStorageLocationCreate(ctx, d, unitTestMockAPIClient)
	assert.Len(t, diags, 0)
	assert.Equal(t, "test-backup-location-id", d.Id())
}

func TestResourceBackupStorageLocationCreateSTS(t *testing.T) {
	ctx := context.Background()
	d := prepareResourceBackupStorageLocation()
	s3 := make([]interface{}, 0)
	s3 = append(s3, map[string]interface{}{
		"s3_url":              "s3://test/test",
		"s3_force_path_style": false,
		"credential_type":     "sts",
		"access_key":          "test-access-key",
		"secret_key":          "test-secret-key",
		"arn":                 "test-arn",
		"external_id":         "test-external-id",
	})
	_ = d.Set("s3", s3)
	diags := resourceBackupStorageLocationCreate(ctx, d, unitTestMockAPIClient)
	assert.Len(t, diags, 0)
	assert.Equal(t, "test-backup-location-id", d.Id())
}

func TestResourceBackupStorageLocationRead(t *testing.T) {
	ctx := context.Background()
	d := prepareResourceBackupStorageLocation()
	d.SetId("test-bsl-location-id")
	diags := resourceBackupStorageLocationRead(ctx, d, unitTestMockAPIClient)
	assert.Len(t, diags, 0)
	assert.Equal(t, "test-bsl-location-id", d.Id())
}

func TestResourceBackupStorageLocationUpdate(t *testing.T) {
	ctx := context.Background()
	d := prepareResourceBackupStorageLocation()
	diags := resourceBackupStorageLocationUpdate(ctx, d, unitTestMockAPIClient)
	assert.Len(t, diags, 0)
	assert.Equal(t, "test-backup-location-id", d.Id())
}

func TestResourceBackupStorageLocationDelete(t *testing.T) {
	ctx := context.Background()
	d := prepareResourceBackupStorageLocation()
	diags := resourceBackupStorageLocationDelete(ctx, d, unitTestMockAPIClient)
	assert.Len(t, diags, 0)
	assert.Equal(t, "test-backup-location-id", d.Id())
}
