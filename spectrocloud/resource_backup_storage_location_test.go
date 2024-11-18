package spectrocloud

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-sdk-go/api/models"
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

func TestToAwsAccountCredential(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]interface{}
		expected *models.V1AwsCloudAccount
	}{
		{
			name: "Secret Credentials",
			input: map[string]interface{}{
				"credential_type": "secret",
				"access_key":      "test-access-key",
				"secret_key":      "test-secret-key",
			},
			expected: &models.V1AwsCloudAccount{
				CredentialType: models.V1AwsCloudAccountCredentialTypeSecret,
				AccessKey:      "test-access-key",
				SecretKey:      "test-secret-key",
			},
		},
		{
			name: "STS Credentials",
			input: map[string]interface{}{
				"credential_type": "sts",
				"arn":             "test-arn",
				"external_id":     "test-external-id",
			},
			expected: &models.V1AwsCloudAccount{
				CredentialType: models.V1AwsCloudAccountCredentialTypeSts,
				Sts: &models.V1AwsStsCredentials{
					Arn:        "test-arn",
					ExternalID: "test-external-id",
				},
			},
		},
		{
			name: "Default to Secret Credentials",
			input: map[string]interface{}{
				"credential_type": "",
				"access_key":      "test-access-key",
				"secret_key":      "test-secret-key",
			},
			expected: &models.V1AwsCloudAccount{
				CredentialType: models.V1AwsCloudAccountCredentialTypeSecret,
				AccessKey:      "test-access-key",
				SecretKey:      "test-secret-key",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := toAwsAccountCredential(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestToAzureBackupStorageLocation(t *testing.T) {
	input := map[string]interface{}{
		"name":       "test-backup",
		"is_default": true,
		"azure_storage_config": []interface{}{
			map[string]interface{}{
				"container_name":      "test-container",
				"storage_name":        "test-storage",
				"stock_keeping_unit":  "Standard_LRS",
				"resource_group":      "test-resource-group",
				"azure_tenant_id":     "test-tenant-id",
				"azure_client_id":     "test-client-id",
				"azure_client_secret": "test-client-secret",
				"subscription_id":     "test-subscription-id",
			},
		},
	}

	resourceSchema := map[string]*schema.Schema{
		"name": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"is_default": {
			Type:     schema.TypeBool,
			Optional: true,
		},
		"azure_storage_config": {
			Type:     schema.TypeList,
			Optional: true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"container_name": {
						Type:     schema.TypeString,
						Optional: true,
					},
					"storage_name": {
						Type:     schema.TypeString,
						Optional: true,
					},
					"stock_keeping_unit": {
						Type:     schema.TypeString,
						Optional: true,
					},
					"resource_group": {
						Type:     schema.TypeString,
						Optional: true,
					},
					"azure_tenant_id": {
						Type:     schema.TypeString,
						Optional: true,
					},
					"azure_client_id": {
						Type:     schema.TypeString,
						Optional: true,
					},
					"azure_client_secret": {
						Type:     schema.TypeString,
						Optional: true,
					},
					"subscription_id": {
						Type:     schema.TypeString,
						Optional: true,
					},
				},
			},
		},
	}

	resourceData := schema.TestResourceDataRaw(t, resourceSchema, input)

	account, accountCredSpec := toAzureBackupStorageLocation(resourceData)

	assert.NotNil(t, account)
	assert.NotNil(t, accountCredSpec)

	assert.Equal(t, "test-backup", account.Metadata.Name)
	assert.Equal(t, true, account.Spec.IsDefault)
	assert.Equal(t, "azure", account.Spec.Type)
	assert.Equal(t, "test-container", *account.Spec.Config.ContainerName)
	assert.Equal(t, "test-storage", *account.Spec.Config.StorageName)
	assert.Equal(t, "Standard_LRS", account.Spec.Config.Sku)
	assert.Equal(t, "test-resource-group", *account.Spec.Config.ResourceGroup)
	assert.Equal(t, "test-tenant-id", account.Spec.Config.Credentials.TenantID)
	assert.Equal(t, "test-client-id", account.Spec.Config.Credentials.ClientID)
	assert.Equal(t, "test-client-secret", account.Spec.Config.Credentials.ClientSecret)
	assert.Equal(t, "test-subscription-id", account.Spec.Config.Credentials.SubscriptionID)

	assert.Equal(t, "AzurePublicCloud", *accountCredSpec.AzureEnvironment)
	assert.Equal(t, "test-client-id", *accountCredSpec.ClientID)
	assert.Equal(t, "test-client-secret", *accountCredSpec.ClientSecret)
	assert.Equal(t, "test-tenant-id", *accountCredSpec.TenantID)
}

func TestToGcpBackupStorageLocation(t *testing.T) {
	input := map[string]interface{}{
		"name":        "test-backup",
		"is_default":  true,
		"bucket_name": "test-bucket",
		"gcp_storage_config": []interface{}{
			map[string]interface{}{
				"project_id":           "test-project-id",
				"gcp_json_credentials": "test-json-credentials",
			},
		},
	}

	resourceSchema := map[string]*schema.Schema{
		"name": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"is_default": {
			Type:     schema.TypeBool,
			Optional: true,
		},
		"bucket_name": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"gcp_storage_config": {
			Type:     schema.TypeList,
			Optional: true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"project_id": {
						Type:     schema.TypeString,
						Optional: true,
					},
					"gcp_json_credentials": {
						Type:     schema.TypeString,
						Optional: true,
					},
				},
			},
		},
	}

	resourceData := schema.TestResourceDataRaw(t, resourceSchema, input)

	account, accountCredSpec := toGcpBackupStorageLocation(resourceData)

	assert.NotNil(t, account)
	assert.NotNil(t, accountCredSpec)

	assert.Equal(t, "test-backup", account.Metadata.Name)
	assert.Equal(t, true, account.Spec.IsDefault)
	assert.Equal(t, "gcp", account.Spec.Type)
	assert.Equal(t, "test-bucket", *account.Spec.Config.BucketName)
	assert.Equal(t, "test-project-id", account.Spec.Config.ProjectID)
	assert.Equal(t, "test-json-credentials", account.Spec.Config.Credentials.JSONCredentials)

	assert.Equal(t, "test-bucket", *accountCredSpec.BucketName)
	assert.Equal(t, "test-json-credentials", accountCredSpec.Credentials.JSONCredentials)
	assert.Equal(t, "test-project-id", accountCredSpec.ProjectID)
}

func TestToMinioBackupStorageLocation(t *testing.T) {
	input := map[string]interface{}{
		"name":        "test-minio",
		"is_default":  true,
		"bucket_name": "test-bucket",
		"region":      "test-region",
		"ca_cert":     "test-ca-cert",
		"s3": []interface{}{
			map[string]interface{}{
				"s3_force_path_style": true,
				"s3_url":              "http://test-s3-url",
				"credential_type":     "secret",
				"access_key":          "test-access-key",
				"secret_key":          "test-secret-key",
			},
		},
	}

	resourceSchema := map[string]*schema.Schema{
		"name": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"is_default": {
			Type:     schema.TypeBool,
			Optional: true,
		},
		"bucket_name": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"region": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"ca_cert": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"s3": {
			Type:     schema.TypeList,
			Optional: true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"s3_force_path_style": {
						Type:     schema.TypeBool,
						Optional: true,
					},
					"s3_url": {
						Type:     schema.TypeString,
						Optional: true,
					},
					"credential_type": {
						Type:     schema.TypeString,
						Optional: true,
					},
					"access_key": {
						Type:     schema.TypeString,
						Optional: true,
					},
					"secret_key": {
						Type:     schema.TypeString,
						Optional: true,
					},
				},
			},
		},
	}

	resourceData := schema.TestResourceDataRaw(t, resourceSchema, input)

	bslEntity, bslCredEntity := toMinioBackupStorageLocation(resourceData)

	assert.NotNil(t, bslEntity)
	assert.NotNil(t, bslCredEntity)

	assert.Equal(t, "test-minio", bslEntity.Metadata.Name)
	assert.Equal(t, true, bslEntity.Spec.IsDefault)
	assert.Equal(t, "test-bucket", *bslEntity.Spec.Config.BucketName)
	assert.Equal(t, "test-region", *bslEntity.Spec.Config.Region)
	assert.Equal(t, "test-ca-cert", bslEntity.Spec.Config.CaCert)
	assert.Equal(t, true, *bslEntity.Spec.Config.S3ForcePathStyle)
	assert.Equal(t, "http://test-s3-url", bslEntity.Spec.Config.S3URL)

	assert.Equal(t, "test-bucket", *bslCredEntity.Bucket)
	assert.Equal(t, "http://test-s3-url", bslCredEntity.Folder)
	assert.Equal(t, "test-region", *bslCredEntity.Region)

	credentials := bslEntity.Spec.Config.Credentials
	assert.NotNil(t, credentials)
	assert.Equal(t, "test-access-key", credentials.AccessKey)
	assert.Equal(t, "test-secret-key", credentials.SecretKey)
}

func TestToS3BackupStorageLocation(t *testing.T) {
	input := map[string]interface{}{
		"name":        "test-s3",
		"is_default":  true,
		"bucket_name": "test-bucket",
		"region":      "us-east-1",
		"ca_cert":     "test-ca-cert",
		"s3": []interface{}{
			map[string]interface{}{
				"s3_force_path_style": true,
				"s3_url":              "http://test-s3-url",
				"credential_type":     "secret",
				"access_key":          "test-access-key",
				"secret_key":          "test-secret-key",
			},
		},
	}

	resourceSchema := map[string]*schema.Schema{
		"name": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"is_default": {
			Type:     schema.TypeBool,
			Optional: true,
		},
		"bucket_name": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"region": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"ca_cert": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"s3": {
			Type:     schema.TypeList,
			Optional: true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"s3_force_path_style": {
						Type:     schema.TypeBool,
						Optional: true,
					},
					"s3_url": {
						Type:     schema.TypeString,
						Optional: true,
					},
					"credential_type": {
						Type:     schema.TypeString,
						Optional: true,
					},
					"access_key": {
						Type:     schema.TypeString,
						Optional: true,
					},
					"secret_key": {
						Type:     schema.TypeString,
						Optional: true,
					},
				},
			},
		},
	}

	resourceData := schema.TestResourceDataRaw(t, resourceSchema, input)

	bslEntity, bslCredEntity := toS3BackupStorageLocation(resourceData)

	assert.NotNil(t, bslEntity)
	assert.NotNil(t, bslCredEntity)

	assert.Equal(t, "test-s3", bslEntity.Metadata.Name)
	assert.Equal(t, true, bslEntity.Spec.IsDefault)
	assert.Equal(t, "test-bucket", *bslEntity.Spec.Config.BucketName)
	assert.Equal(t, "us-east-1", *bslEntity.Spec.Config.Region)
	assert.Equal(t, "test-ca-cert", bslEntity.Spec.Config.CaCert)
	assert.Equal(t, true, *bslEntity.Spec.Config.S3ForcePathStyle)
	assert.Equal(t, "http://test-s3-url", bslEntity.Spec.Config.S3URL)

	assert.Equal(t, "test-bucket", *bslCredEntity.Bucket)
	assert.Equal(t, "http://test-s3-url", bslCredEntity.Folder)
	assert.Equal(t, "us-east-1", *bslCredEntity.Region)

	credentials := bslEntity.Spec.Config.Credentials
	assert.NotNil(t, credentials)
	assert.Equal(t, "test-access-key", credentials.AccessKey)
	assert.Equal(t, "test-secret-key", credentials.SecretKey)
}
