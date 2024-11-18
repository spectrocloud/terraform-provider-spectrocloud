package spectrocloud

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the backup storage location. This is a unique identifier for the backup location.",
			},
			"location_provider": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Default:      "aws",
				ValidateFunc: validation.StringInSlice([]string{"aws", "minio", "gcp", "azure"}, false),
				Description: "The location provider for backup storage location. Allowed values are `aws` or `minio` or `gcp` or `azure`. " +
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
							Description: "The stop-keeping unit.",
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
	locationProvider := d.Get("location_provider").(string)

	switch locationProvider {
	case "aws":
		return S3BackupStorageLocationCreate(d, c)
	case "minio":
		return MinioBackupStorageLocationCreate(d, c)
	case "gcp":
		return GcpBackupStorageLocationCreate(d, c)
	case "azure":
		return AzureBackupStorageLocationCreate(d, c)
	default:
		return S3BackupStorageLocationCreate(d, c)
	}
}

func resourceBackupStorageLocationRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	assetContext := d.Get("context").(string)
	c := getV1ClientWithResourceContext(m, assetContext)
	locationProvider := d.Get("location_provider").(string)

	switch locationProvider {
	case "aws":
		return S3BackupStorageLocationRead(d, c)
	case "minio":
		return MinioBackupStorageLocationRead(d, c)
	case "gcp":
		return GcpBackupStorageLocationRead(d, c)
	case "azure":
		return AzureBackupStorageLocationRead(d, c)
	default:
		return S3BackupStorageLocationRead(d, c)
	}
}

func resourceBackupStorageLocationUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	assetContext := d.Get("context").(string)
	c := getV1ClientWithResourceContext(m, assetContext)

	locationProvider := d.Get("location_provider").(string)

	switch locationProvider {
	case "aws":
		return S3BackupStorageLocationUpdate(d, c)
	case "minio":
		return MinioBackupStorageLocationUpdate(d, c)
	case "gcp":
		return GcpBackupStorageLocationUpdate(d, c)
	case "azure":
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
