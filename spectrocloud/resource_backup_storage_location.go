package spectrocloud

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	StorageProviderAWS   = "aws"
	StorageProviderMinio = "minio"
	StorageProviderGCP   = "gcp"
	StorageProviderAzure = "azure"
)

func resourceBackupStorageLocation() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceBackupStorageLocationCreate,
		ReadContext:   resourceBackupStorageLocationRead,
		UpdateContext: resourceBackupStorageLocationUpdate,
		DeleteContext: resourceBackupStorageLocationDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceBackupStorageLocationImport,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		SchemaVersion: 2,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the backup storage location. This is a unique identifier for the backup location.",
			},
			"storage_provider": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Default:  StorageProviderAWS,
				ValidateFunc: validation.StringInSlice([]string{
					StorageProviderAWS,
					StorageProviderMinio,
					StorageProviderGCP,
					StorageProviderAzure,
				}, false),
				Description: "The storage location provider for backup storage. Allowed values are `aws` or `minio` or `gcp` or `azure`. " +
					"Default value is `aws`.",
			},
			"context": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "project",
				ValidateFunc: validation.StringInSlice([]string{"project", "tenant"}, false),
				Description: "The context of the backup storage location. Allowed values are `project` or `tenant`. " +
					"Default value is `project`. " + PROJECT_NAME_NUANCE,
			},
			"is_default": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Specifies if this backup storage location should be used as the default location for storing backups.",
			},
			"region": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The region where the backup storage is located, typically corresponding to the region of the cloud provider. This is relevant for S3 or S3-compatible(minio) storage services.",
			},
			"bucket_name": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The name of the storage bucket where backups are stored. This is relevant for S3 or S3-compatible(minio) or gcp storage services.",
			},
			"ca_cert": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "An optional CA certificate used for SSL connections to ensure secure communication with the storage provider. This is relevant for S3 or S3-compatible(minio) storage services.",
			},
			"s3": {
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				Description: "S3-specific settings for configuring the backup storage location.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"s3_url": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "The S3 URL endpoint.",
						},
						"s3_force_path_style": {
							Type:        schema.TypeBool,
							Optional:    true,
							Description: "A boolean flag indicating whether to enforce the path-style URL for accessing S3.",
						},
						"credential_type": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice([]string{"secret", "sts"}, false),
							Description:  "The type of credentials used to access the S3 storage. Supported values are 'secret' for static credentials and 'sts' for temporary, token-based credentials.",
						},
						"access_key": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "The access key for S3 authentication, required if 'credential_type' is set to 'secret'.",
						},
						"secret_key": {
							Type:        schema.TypeString,
							Optional:    true,
							Sensitive:   true,
							Description: "The secret key for S3 authentication, required if 'credential_type' is set to 'secret'.",
						},
						"arn": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "The Amazon Resource Name (ARN) of the IAM role to assume for accessing S3 when using 'sts' credentials.",
						},
						"external_id": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "An external ID used for cross-account access to the S3 storage when using 'sts' credentials.",
						},
					},
				},
			},
			"gcp_storage_config": {
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				Description: "GCP storage settings for configuring the backup storage location.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"project_id": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The GCP project ID.",
						},
						"gcp_json_credentials": {
							Type:        schema.TypeString,
							Required:    true,
							Sensitive:   true,
							Description: "The GCP credentials in JSON format. These credentials are required to authenticate and manage.",
						},
					},
				},
			},
			"azure_storage_config": {
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				Description: "Azure storage settings for configuring the backup storage location.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"container_name": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The container name.",
						},
						"storage_name": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The storage name.",
						},
						"stock_keeping_unit": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The stop-keeping unit. eg: `Standard_LRS`",
						},
						"resource_group": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The resource group name.",
						},
						"azure_tenant_id": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Unique tenant Id from Azure console.",
						},
						"azure_client_id": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Unique client Id from Azure console.",
						},
						"subscription_id": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Unique subscription Id from Azure console.",
						},
						"azure_client_secret": {
							Type:        schema.TypeString,
							Required:    true,
							Sensitive:   true,
							Description: "Azure secret for authentication.",
						},
					},
				},
			},
		},
		CustomizeDiff: schemaValidationForLocationProvider,
	}
}

func resourceBackupStorageLocationCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	assetContext := d.Get("context").(string)
	c := getV1ClientWithResourceContext(m, assetContext)
	storageProvider := d.Get("storage_provider").(string)

	switch storageProvider {
	case StorageProviderAWS:
		return S3BackupStorageLocationCreate(d, c)
	case StorageProviderMinio:
		return MinioBackupStorageLocationCreate(d, c)
	case StorageProviderGCP:
		return GcpBackupStorageLocationCreate(d, c)
	case StorageProviderAzure:
		return AzureBackupStorageLocationCreate(d, c)
	default:
		return S3BackupStorageLocationCreate(d, c)
	}
}

func resourceBackupStorageLocationRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	assetContext := d.Get("context").(string)
	c := getV1ClientWithResourceContext(m, assetContext)
	storageProvider := d.Get("storage_provider").(string)

	switch storageProvider {
	case StorageProviderAWS:
		return S3BackupStorageLocationRead(d, c)
	case StorageProviderMinio:
		return MinioBackupStorageLocationRead(d, c)
	case StorageProviderGCP:
		return GcpBackupStorageLocationRead(d, c)
	case StorageProviderAzure:
		return AzureBackupStorageLocationRead(d, c)
	default:
		return S3BackupStorageLocationRead(d, c)
	}
}

func resourceBackupStorageLocationUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	assetContext := d.Get("context").(string)
	c := getV1ClientWithResourceContext(m, assetContext)

	storageProvider := d.Get("storage_provider").(string)

	switch storageProvider {
	case StorageProviderAWS:
		return S3BackupStorageLocationUpdate(d, c)
	case StorageProviderMinio:
		return MinioBackupStorageLocationUpdate(d, c)
	case StorageProviderGCP:
		return GcpBackupStorageLocationUpdate(d, c)
	case StorageProviderAzure:
		return AzureBackupStorageLocationUpdate(d, c)
	default:
		return S3BackupStorageLocationUpdate(d, c)
	}

}

func resourceBackupStorageLocationDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	assetContext := d.Get("context").(string)
	c := getV1ClientWithResourceContext(m, assetContext)
	var diags diag.Diagnostics
	err := c.DeleteS3BackupStorageLocation(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}
