package spectrocloud

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-api-go/models"
)

func resourceBackupStorageLocation() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceBackupStorageLocationCreate,
		ReadContext:   resourceBackupStorageLocationRead,
		UpdateContext: resourceBackupStorageLocationUpdate,
		DeleteContext: resourceBackupStorageLocationDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		SchemaVersion: 2,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"is_default": {
				Type:     schema.TypeBool,
				Required: true,
			},
			"region": {
				Type:     schema.TypeString,
				Required: true,
			},
			"bucket_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"ca_cert": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"s3": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"s3_url": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"s3_force_path_style": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"credential_type": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice([]string{"secret", "sts"}, false),
						},
						"access_key": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"secret_key": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"arn": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"external_id": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
		},
	}
}

func resourceBackupStorageLocationCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := GetResourceLevelV1Client(m, "")
	var diags diag.Diagnostics

	bsl := toBackupStorageLocation(d)
	uid, err := c.CreateS3BackupStorageLocation(bsl)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(uid)

	return diags
}

func resourceBackupStorageLocationRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := GetResourceLevelV1Client(m, "")
	var diags diag.Diagnostics

	bsl, err := c.GetBackupStorageLocation(d.Id())
	if err != nil {
		return diag.FromErr(err)
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

	if bsl.Spec.Storage == "s3" {
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
		s3["credential_type"] = string(s3Bsl.Spec.Config.Credentials.CredentialType)
		if s3Bsl.Spec.Config.Credentials.CredentialType == models.V1AwsCloudAccountCredentialTypeSecret {
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

func resourceBackupStorageLocationUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := GetResourceLevelV1Client(m, "")
	var diags diag.Diagnostics

	bsl := toBackupStorageLocation(d)
	err := c.UpdateS3BackupStorageLocation(d.Id(), bsl)
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func resourceBackupStorageLocationDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := GetResourceLevelV1Client(m, "")
	var diags diag.Diagnostics
	err := c.DeleteS3BackupStorageLocation(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func toBackupStorageLocation(d *schema.ResourceData) *models.V1UserAssetsLocationS3 {
	bucketName := d.Get("bucket_name").(string)
	region := d.Get("region").(string)
	s3config := d.Get("s3").([]interface{})[0].(map[string]interface{})
	s3ForcePathStyle := s3config["s3_force_path_style"].(bool)
	return &models.V1UserAssetsLocationS3{
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
}

func toAwsAccountCredential(s3cred map[string]interface{}) *models.V1AwsCloudAccount {
	account := &models.V1AwsCloudAccount{}
	if len(s3cred["credential_type"].(string)) == 0 || s3cred["credential_type"].(string) == "secret" {
		account.CredentialType = models.V1AwsCloudAccountCredentialTypeSecret
		account.AccessKey = s3cred["access_key"].(string)
		account.SecretKey = s3cred["secret_key"].(string)
	} else if s3cred["credential_type"].(string) == "sts" {
		account.CredentialType = models.V1AwsCloudAccountCredentialTypeSts
		account.Sts = &models.V1AwsStsCredentials{
			Arn:        s3cred["arn"].(string),
			ExternalID: s3cred["external_id"].(string),
		}
	}
	return account
}
