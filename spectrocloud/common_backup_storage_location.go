package spectrocloud

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/spectrocloud/palette-sdk-go/client"
)

func schemaValidationForLocationProvider(ctx context.Context, d *schema.ResourceDiff, meta interface{}) error {
	provider := d.Get("storage_provider").(string)
	if (provider == StorageProviderAWS || provider == StorageProviderMinio) && (len(d.Get("s3").([]interface{})) == 0 || d.Get("bucket_name").(string) == "" || d.Get("region").(string) == "") {
		return fmt.Errorf("`s3, bucket_name & region` is required when location provider set to 'aws' or 'minio'")
	}
	if (provider == StorageProviderAWS || provider == StorageProviderMinio) && (len(d.Get("azure_storage_config").([]interface{})) != 0 || (len(d.Get("gcp_storage_config").([]interface{}))) != 0) {
		return fmt.Errorf("`gcp_storage_config or azure_storage_config` are not allowed when location provider set to 'aws' or 'minio'")
	}
	if (provider == StorageProviderGCP) && (len(d.Get("gcp_storage_config").([]interface{})) == 0 || d.Get("bucket_name").(string) == "") {
		return fmt.Errorf("`gcp_storage_config & bucket_name` is required when location provider set to 'gcp'")
	}
	if (provider == StorageProviderAzure) && len(d.Get("azure_storage_config").([]interface{})) == 0 {
		return fmt.Errorf("`azure_storage_config` is required when location provider set to 'azure'")
	}
	if provider == StorageProviderAzure && (len(d.Get("s3").([]interface{})) != 0 || d.Get("bucket_name").(string) != "" || d.Get("region").(string) != "" || d.Get("ca_cert").(string) != "") {
		return fmt.Errorf("`s3, bucket_name, region & ca_cert` are not allowed when location provider set to 'azure'")
	}
	if (provider == StorageProviderGCP) && (len(d.Get("azure_storage_config").([]interface{})) != 0 || len(d.Get("s3").([]interface{})) != 0 || d.Get("region").(string) != "" || d.Get("ca_cert").(string) != "") {
		return fmt.Errorf("`azure_storage_config, s3, region, ca_cert` are not allowed when location provider set to 'gcp'")
	}
	return nil
}

func S3BackupStorageLocationCreate(d *schema.ResourceData, c *client.V1Client) diag.Diagnostics {
	var diags diag.Diagnostics

	bsl, bslCred := toS3BackupStorageLocation(d)
	if err := c.ValidateS3BackupStorageLocation(bslCred); err != nil {
		return diag.FromErr(err)
	}

	uid, err := c.CreateS3BackupStorageLocation(bsl)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(uid)

	return diags
}

func MinioBackupStorageLocationCreate(d *schema.ResourceData, c *client.V1Client) diag.Diagnostics {
	var diags diag.Diagnostics

	bsl, _ := toMinioBackupStorageLocation(d)
	// No credential validation required for minio
	uid, err := c.CreateMinioBackupStorageLocation(bsl)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(uid)

	return diags
}

func GcpBackupStorageLocationCreate(d *schema.ResourceData, c *client.V1Client) diag.Diagnostics {
	var diags diag.Diagnostics

	bsl, bslCred := toGcpBackupStorageLocation(d)
	if err := c.ValidateGcpBackupStorageLocation(bslCred); err != nil {
		return diag.FromErr(err)
	}

	uid, err := c.CreateGcpBackupStorageLocation(bsl)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(uid)

	return diags
}

func AzureBackupStorageLocationCreate(d *schema.ResourceData, c *client.V1Client) diag.Diagnostics {
	var diags diag.Diagnostics

	bsl, bslCred := toAzureBackupStorageLocation(d)
	if err := c.ValidateAzureBackupStorageLocation(bslCred); err != nil {
		return diag.FromErr(err)
	}

	uid, err := c.CreateAzureBackupStorageLocation(bsl)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(uid)

	return diags
}

func S3BackupStorageLocationRead(d *schema.ResourceData, c *client.V1Client) diag.Diagnostics {
	var diags diag.Diagnostics

	bsl, err := c.GetBackupStorageLocation(d.Id())
	if err != nil {
		return handleReadError(d, err, diags)
	} else if bsl == nil {
		// Deleted - Terraform will recreate it
		d.SetId("")
		return diags
	}

	if err := d.Set("name", bsl.Metadata.Name); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("is_default", bsl.Spec.IsDefault); err != nil {
		return diag.FromErr(err)
	}

	if *bsl.Spec.Storage == models.V1LocationTypeS3 {
		s3Bsl, err := c.GetS3BackupStorageLocation(d.Id())
		if err != nil {
			return diag.FromErr(err)
		} else if s3Bsl == nil {
			// Deleted - Terraform will recreate it
			d.SetId("")
			return diags
		}
		if len(s3Bsl.Spec.Config.CaCert) > 0 {
			if err := d.Set("ca_cert", s3Bsl.Spec.Config.CaCert); err != nil {
				return diag.FromErr(err)
			}
		}
		if err := d.Set("region", *s3Bsl.Spec.Config.Region); err != nil {
			return diag.FromErr(err)
		}
		if err := d.Set("bucket_name", *s3Bsl.Spec.Config.BucketName); err != nil {
			return diag.FromErr(err)
		}

		s3 := make(map[string]interface{})
		if len(s3Bsl.Spec.Config.S3URL) > 0 {
			s3["s3_url"] = s3Bsl.Spec.Config.S3URL
		}

		if s3Bsl.Spec.Config.S3ForcePathStyle != nil {
			s3["s3_force_path_style"] = *s3Bsl.Spec.Config.S3ForcePathStyle
		}
		s3["credential_type"] = string(*s3Bsl.Spec.Config.Credentials.CredentialType)
		if *s3Bsl.Spec.Config.Credentials.CredentialType == models.V1AwsCloudAccountCredentialTypeSecret {
			s3["access_key"] = s3Bsl.Spec.Config.Credentials.AccessKey
			s3["secret_key"] = s3Bsl.Spec.Config.Credentials.SecretKey
		} else {
			s3["arn"] = s3Bsl.Spec.Config.Credentials.Sts.Arn
			if len(s3Bsl.Spec.Config.Credentials.Sts.ExternalID) > 0 {
				s3["external_id"] = s3Bsl.Spec.Config.Credentials.Sts.ExternalID
			}
		}
		s3Config := make([]interface{}, 0, 1)
		s3Config = append(s3Config, s3)
		if err := d.Set("s3", s3Config); err != nil {
			return diag.FromErr(err)
		}
	}

	return diags
}

func MinioBackupStorageLocationRead(d *schema.ResourceData, c *client.V1Client) diag.Diagnostics {
	var diags diag.Diagnostics

	bsl, err := c.GetBackupStorageLocation(d.Id())
	if err != nil {
		return handleReadError(d, err, diags)
	} else if bsl == nil {
		// Deleted - Terraform will recreate it
		d.SetId("")
		return diags
	}

	if err := d.Set("name", bsl.Metadata.Name); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("is_default", bsl.Spec.IsDefault); err != nil {
		return diag.FromErr(err)
	}

	if *bsl.Spec.Storage == models.V1LocationTypeMinio {
		s3Bsl, err := c.GetMinioBackupStorageLocation(d.Id())
		if err != nil {
			return diag.FromErr(err)
		} else if s3Bsl == nil {
			// Deleted - Terraform will recreate it
			d.SetId("")
			return diags
		}
		if len(s3Bsl.Spec.Config.CaCert) > 0 {
			if err := d.Set("ca_cert", s3Bsl.Spec.Config.CaCert); err != nil {
				return diag.FromErr(err)
			}
		}
		if err := d.Set("region", *s3Bsl.Spec.Config.Region); err != nil {
			return diag.FromErr(err)
		}
		if err := d.Set("bucket_name", *s3Bsl.Spec.Config.BucketName); err != nil {
			return diag.FromErr(err)
		}

		s3 := make(map[string]interface{})
		if len(s3Bsl.Spec.Config.S3URL) > 0 {
			s3["s3_url"] = s3Bsl.Spec.Config.S3URL
		}

		if s3Bsl.Spec.Config.S3ForcePathStyle != nil {
			s3["s3_force_path_style"] = *s3Bsl.Spec.Config.S3ForcePathStyle
		}
		// Minio only supports secret type credentials
		s3["credential_type"] = string(*s3Bsl.Spec.Config.Credentials.CredentialType)
		if *s3Bsl.Spec.Config.Credentials.CredentialType == models.V1AwsCloudAccountCredentialTypeSecret {
			s3["access_key"] = s3Bsl.Spec.Config.Credentials.AccessKey
			s3["secret_key"] = s3Bsl.Spec.Config.Credentials.SecretKey
		}
		s3Config := make([]interface{}, 0, 1)
		s3Config = append(s3Config, s3)
		if err := d.Set("s3", s3Config); err != nil {
			return diag.FromErr(err)
		}
	}

	return diags
}

func GcpBackupStorageLocationRead(d *schema.ResourceData, c *client.V1Client) diag.Diagnostics {
	var diags diag.Diagnostics

	bsl, err := c.GetBackupStorageLocation(d.Id())
	if err != nil {
		return handleReadError(d, err, diags)
	} else if bsl == nil {
		// Deleted - Terraform will recreate it
		d.SetId("")
		return diags
	}
	if err := d.Set("name", bsl.Metadata.Name); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("is_default", bsl.Spec.IsDefault); err != nil {
		return diag.FromErr(err)
	}
	if *bsl.Spec.Storage == models.V1LocationTypeGcp {
		gcpBsl, err := c.GetGCPBackupStorageLocation(d.Id())
		if err != nil {
			return diag.FromErr(err)
		} else if gcpBsl == nil {
			// Deleted - Terraform will recreate it
			d.SetId("")
			return diags
		}
		if err := d.Set("bucket_name", *gcpBsl.Spec.Config.BucketName); err != nil {
			return diag.FromErr(err)
		}
		gcpConfig := make([]interface{}, 0)
		if err := d.Set("bucket_name", *gcpBsl.Spec.Config.BucketName); err != nil {
			return diag.FromErr(err)
		}
		gcpConfig = append(gcpConfig, map[string]interface{}{
			"project_id":           gcpBsl.Spec.Config.ProjectID,
			"gcp_json_credentials": gcpBsl.Spec.Config.Credentials.JSONCredentials,
		})
		if err := d.Set("gcp_storage_config", gcpConfig); err != nil {
			return diag.FromErr(err)
		}
	}
	return diags
}

func AzureBackupStorageLocationRead(d *schema.ResourceData, c *client.V1Client) diag.Diagnostics {
	var diags diag.Diagnostics

	bsl, err := c.GetBackupStorageLocation(d.Id())
	if err != nil {
		return handleReadError(d, err, diags)
	} else if bsl == nil {
		// Deleted - Terraform will recreate it
		d.SetId("")
		return diags
	}
	if err := d.Set("name", bsl.Metadata.Name); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("is_default", bsl.Spec.IsDefault); err != nil {
		return diag.FromErr(err)
	}
	azureBsl, err := c.GetAzureBackupStorageLocation(d.Id())
	if err != nil {
		return diag.FromErr(err)
	} else if azureBsl == nil {
		// Deleted - Terraform will recreate it
		d.SetId("")
		return diags
	}
	azConfig := make([]interface{}, 0)
	azConfig = append(azConfig, map[string]interface{}{
		"container_name":      azureBsl.Spec.Config.ContainerName,
		"storage_name":        azureBsl.Spec.Config.StorageName,
		"stock_keeping_unit":  azureBsl.Spec.Config.Sku,
		"resource_group":      azureBsl.Spec.Config.ResourceGroup,
		"azure_tenant_id":     azureBsl.Spec.Config.Credentials.TenantID,
		"azure_client_id":     azureBsl.Spec.Config.Credentials.ClientID,
		"subscription_id":     azureBsl.Spec.Config.Credentials.SubscriptionID,
		"azure_client_secret": azureBsl.Spec.Config.Credentials.ClientSecret,
	})
	if err := d.Set("azure_storage_config", azConfig); err != nil {
		return diag.FromErr(err)
	}
	return diags
}

func S3BackupStorageLocationUpdate(d *schema.ResourceData, c *client.V1Client) diag.Diagnostics {
	var diags diag.Diagnostics
	bsl, bslCred := toS3BackupStorageLocation(d)
	if err := c.ValidateS3BackupStorageLocation(bslCred); err != nil {
		return diag.FromErr(err)
	}
	err := c.UpdateS3BackupStorageLocation(d.Id(), bsl)
	if err != nil {
		return diag.FromErr(err)
	}
	return diags
}

func MinioBackupStorageLocationUpdate(d *schema.ResourceData, c *client.V1Client) diag.Diagnostics {
	var diags diag.Diagnostics
	bsl, _ := toMinioBackupStorageLocation(d)
	// No credential validation required for minio
	err := c.UpdateMinioBackupStorageLocation(d.Id(), bsl)
	if err != nil {
		return diag.FromErr(err)
	}
	return diags
}

func GcpBackupStorageLocationUpdate(d *schema.ResourceData, c *client.V1Client) diag.Diagnostics {
	var diags diag.Diagnostics
	bsl, bslCred := toGcpBackupStorageLocation(d)
	if err := c.ValidateGcpBackupStorageLocation(bslCred); err != nil {
		return diag.FromErr(err)
	}
	err := c.UpdateGcpBackupStorageLocation(d.Id(), bsl)
	if err != nil {
		return diag.FromErr(err)
	}
	return diags
}

func AzureBackupStorageLocationUpdate(d *schema.ResourceData, c *client.V1Client) diag.Diagnostics {
	var diags diag.Diagnostics
	bsl, bslCred := toAzureBackupStorageLocation(d)
	if err := c.ValidateAzureBackupStorageLocation(bslCred); err != nil {
		return diag.FromErr(err)
	}
	err := c.UpdateAzureBackupStorageLocation(d.Id(), bsl)
	if err != nil {
		return diag.FromErr(err)
	}
	return diags
}

func toS3BackupStorageLocation(d *schema.ResourceData) (*models.V1UserAssetsLocationS3, *models.V1AwsS3BucketCredentials) {
	bucketName := d.Get("bucket_name").(string)
	region := d.Get("region").(string)
	s3config := d.Get("s3").([]interface{})[0].(map[string]interface{})
	s3ForcePathStyle := s3config["s3_force_path_style"].(bool)
	bslEntity := &models.V1UserAssetsLocationS3{
		Metadata: &models.V1ObjectMetaInputEntity{
			Name: d.Get("name").(string),
		},
		Spec: &models.V1UserAssetsLocationS3Spec{
			Config: &models.V1S3StorageConfig{
				BucketName:       &bucketName,
				CaCert:           d.Get("ca_cert").(string),
				Credentials:      toAwsAccountCredential(s3config),
				Region:           &region,
				S3ForcePathStyle: &s3ForcePathStyle,
				S3URL:            s3config["s3_url"].(string),
				UseRestic:        nil,
			},
			IsDefault: d.Get("is_default").(bool),
		},
	}
	bslCredEntity := &models.V1AwsS3BucketCredentials{
		Bucket:      bslEntity.Spec.Config.BucketName,
		Credentials: bslEntity.Spec.Config.Credentials,
		Folder:      bslEntity.Spec.Config.S3URL,
		Region:      bslEntity.Spec.Config.Region,
	}
	return bslEntity, bslCredEntity
}

func toMinioBackupStorageLocation(d *schema.ResourceData) (*models.V1UserAssetsLocationS3, *models.V1AwsS3BucketCredentials) {
	bucketName := d.Get("bucket_name").(string)
	region := d.Get("region").(string)
	s3config := d.Get("s3").([]interface{})[0].(map[string]interface{})
	s3ForcePathStyle := s3config["s3_force_path_style"].(bool)
	bslEntity := &models.V1UserAssetsLocationS3{
		Metadata: &models.V1ObjectMetaInputEntity{
			Name: d.Get("name").(string),
		},
		Spec: &models.V1UserAssetsLocationS3Spec{
			Config: &models.V1S3StorageConfig{
				BucketName:       &bucketName,
				CaCert:           d.Get("ca_cert").(string),
				Credentials:      toAwsAccountCredential(s3config),
				Region:           &region,
				S3ForcePathStyle: &s3ForcePathStyle,
				S3URL:            s3config["s3_url"].(string),
				UseRestic:        nil,
			},
			IsDefault: d.Get("is_default").(bool),
		},
	}
	bslCredEntity := &models.V1AwsS3BucketCredentials{
		Bucket:      bslEntity.Spec.Config.BucketName,
		Credentials: bslEntity.Spec.Config.Credentials,
		Folder:      bslEntity.Spec.Config.S3URL,
		Region:      bslEntity.Spec.Config.Region,
	}
	return bslEntity, bslCredEntity
}

func toGcpBackupStorageLocation(d *schema.ResourceData) (*models.V1UserAssetsLocationGcp, *models.V1GcpAccountNameValidateSpec) {
	var account *models.V1UserAssetsLocationGcp
	gcpCred := d.Get("gcp_storage_config").([]interface{})[0].(map[string]interface{})
	if len(gcpCred) > 0 {
		bslName := d.Get("name").(string)
		isDefault := d.Get("is_default").(bool)
		bucketName := d.Get("bucket_name").(string)
		projectId := gcpCred["project_id"].(string)
		jsonCred := gcpCred["gcp_json_credentials"].(string)
		account = &models.V1UserAssetsLocationGcp{
			Metadata: &models.V1ObjectMetaInputEntity{
				Annotations: nil,
				Labels:      nil,
				Name:        bslName,
			},
			Spec: &models.V1UserAssetsLocationGcpSpec{
				Config: &models.V1GcpStorageConfig{
					BucketName: &bucketName,
					Credentials: &models.V1GcpAccountEntitySpec{
						JSONCredentials: jsonCred,
					},
					ProjectID: projectId,
				},
				IsDefault: isDefault,
				Type:      StorageProviderGCP,
			},
		}
		accountCredSpec := &models.V1GcpAccountNameValidateSpec{
			BucketName: account.Spec.Config.BucketName,
			Credentials: &models.V1GcpAccountValidateSpec{
				JSONCredentials: account.Spec.Config.Credentials.JSONCredentials,
			},
			ProjectID: account.Spec.Config.ProjectID,
		}
		return account, accountCredSpec
	}

	return nil, nil
}

func toAzureBackupStorageLocation(d *schema.ResourceData) (*models.V1UserAssetsLocationAzure, *models.V1AzureCloudAccount) {
	var account *models.V1UserAssetsLocationAzure
	azureCred := d.Get("azure_storage_config").([]interface{})[0].(map[string]interface{})
	if len(azureCred) > 0 {
		bslName := d.Get("name").(string)
		isDefault := d.Get("is_default").(bool)
		containerName := azureCred["container_name"].(string)
		storageName := azureCred["storage_name"].(string)
		sku := azureCred["stock_keeping_unit"].(string)
		resourceGroup := azureCred["resource_group"].(string)
		azTenantId := azureCred["azure_tenant_id"].(string)
		azClientId := azureCred["azure_client_id"].(string)
		azClientSecret := azureCred["azure_client_secret"].(string)
		subId := azureCred["subscription_id"].(string)
		account = &models.V1UserAssetsLocationAzure{
			Metadata: &models.V1ObjectMetaInputEntity{
				Name: bslName,
			},
			Spec: &models.V1UserAssetsLocationAzureSpec{
				Config: &models.V1AzureStorageConfig{
					ContainerName: &containerName,
					Credentials: &models.V1AzureAccountEntitySpec{
						ClientCloud:    StringPtr("public"),
						ClientID:       azClientId,
						ClientSecret:   azClientSecret,
						SubscriptionID: subId,
						TenantID:       azTenantId,
					},
					ResourceGroup: &resourceGroup,
					Sku:           sku,
					StorageName:   &storageName,
				},
				IsDefault: isDefault,
				Type:      StorageProviderAzure,
			},
		}
		accountCredSpec := &models.V1AzureCloudAccount{
			AzureEnvironment: StringPtr("AzurePublicCloud"),
			ClientID:         &account.Spec.Config.Credentials.ClientID,
			ClientSecret:     &account.Spec.Config.Credentials.ClientSecret,
			Settings:         nil,
			TenantID:         &account.Spec.Config.Credentials.TenantID,
		}
		return account, accountCredSpec
	}
	return nil, nil
}

func toAwsAccountCredential(s3cred map[string]interface{}) *models.V1AwsCloudAccount {
	account := &models.V1AwsCloudAccount{}
	if len(s3cred["credential_type"].(string)) == 0 || s3cred["credential_type"].(string) == "secret" {
		account.CredentialType = models.V1AwsCloudAccountCredentialTypeSecret.Pointer()
		account.AccessKey = s3cred["access_key"].(string)
		account.SecretKey = s3cred["secret_key"].(string)
	} else if s3cred["credential_type"].(string) == "sts" {
		account.CredentialType = models.V1AwsCloudAccountCredentialTypeSts.Pointer()
		account.Sts = &models.V1AwsStsCredentials{
			Arn:        s3cred["arn"].(string),
			ExternalID: s3cred["external_id"].(string),
		}
	}
	return account
}
