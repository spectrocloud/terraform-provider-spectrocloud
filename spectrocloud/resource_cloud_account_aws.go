package spectrocloud

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/spectrocloud/terraform-provider-spectrocloud/types"
)

func resourceCloudAccountAws() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceCloudAccountAwsCreate,
		ReadContext:   resourceCloudAccountAwsRead,
		UpdateContext: resourceCloudAccountAwsUpdate,
		DeleteContext: resourceCloudAccountAwsDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceAccountAwsImport,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"context": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "project",
				ValidateFunc: validation.StringInSlice([]string{"", "project", "tenant"}, false),
				Description: "The context of the AWS configuration. Allowed values are `project` or `tenant`. " +
					"Default value is `project`. " + PROJECT_NAME_NUANCE,
			},
			"private_cloud_gateway_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "ID of the private cloud gateway. This is the ID of the private cloud gateway that is used to connect to the private cluster endpoint.",
			},
			"aws_access_key": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The AWS access key used to authenticate. **Deprecated:** Use `aws_secured_access_key` instead for enhanced security. **Note:** This field is mutually exclusive with `aws_secured_access_key`.",
			},
			"aws_secured_access_key": {
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				Description: "The AWS access key used to authenticate. This is a secure alternative to `aws_access_key` with sensitive attribute enabled. **Note:** This field is mutually exclusive with `aws_access_key`.",
			},
			"aws_secret_key": {
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				Description: "The AWS secret key used in conjunction with the access key for authentication.",
			},
			"type": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"secret", "sts", "pod-identity"}, false),
				Default:      "secret",
				Description:  "The type of AWS credentials to use. Can be `secret`, `sts`, or `pod-identity`. ",
			},
			"arn": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The Amazon Resource Name (ARN) associated with the AWS resource. This is used for identifying resources in AWS. Used for STS credential type.",
			},
			"external_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				Description: "An optional external ID that can be used for cross-account access in AWS. Used for STS credential type.",
			},
			"role_arn": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The IAM Role ARN for AWS EKS Pod Identity authentication. Required when type is `pod-identity`.",
			},
			"permission_boundary_arn": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Optional Permission Boundary ARN to limit the maximum permissions for roles created by Hubble. Used with `pod-identity` credential type.",
			},
			"partition": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "aws",
				ValidateFunc: validation.StringInSlice([]string{"aws", "aws-us-gov"}, false),
				Description: `The AWS partition in which the cloud account is located. 
Can be 'aws' for standard AWS regions or 'aws-us-gov' for AWS GovCloud (US) regions.
Default is 'aws'.`,
			},
			"policy_arns": {
				Type:        schema.TypeSet,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "A set of ARNs for the IAM policies that should be associated with the cloud account.",
			},
		},
	}
}

func resourceCloudAccountAwsCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	resourceContext := d.Get("context").(string)
	c := getV1ClientWithResourceContext(m, resourceContext)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	account, err := toAwsAccount(d)
	if err != nil {
		return diag.FromErr(err)
	}

	uid, err := c.CreateCloudAccountAws(account)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(uid)

	resourceCloudAccountAwsRead(ctx, d, m)

	return diags
}

func resourceCloudAccountAwsRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	resourceContext := d.Get("context").(string)
	c := getV1ClientWithResourceContext(m, resourceContext)

	var diags diag.Diagnostics

	uid := d.Id()
	account, err := c.GetCloudAccountAws(uid)
	if err != nil {
		return handleReadError(d, err, diags)
	} else if account == nil {
		// Deleted - Terraform will recreate it
		d.SetId("")
		return diags
	}

	diagnostics, done := flattenCloudAccountAws(d, account)
	if done {
		return diagnostics
	}

	return diags
}

func resourceCloudAccountAwsUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	resourceContext := d.Get("context").(string)
	c := getV1ClientWithResourceContext(m, resourceContext)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	account, err := toAwsAccount(d)
	if err != nil {
		return diag.FromErr(err)
	}

	err = c.UpdateCloudAccountAws(account)
	if err != nil {
		return diag.FromErr(err)
	}

	resourceCloudAccountAwsRead(ctx, d, m)

	return diags
}

func resourceCloudAccountAwsDelete(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	resourceContext := d.Get("context").(string)
	c := getV1ClientWithResourceContext(m, resourceContext)

	var diags diag.Diagnostics

	cloudAccountID := d.Id()

	err := c.DeleteCloudAccountAws(cloudAccountID)
	if err != nil {
		return diag.FromErr(err)
	}

	// d.SetId("") is automatically called assuming delete returns no errors

	return diags
}

func toAwsAccount(d *schema.ResourceData) (*models.V1AwsAccount, error) {
	// Validate that only one access key field is set
	securedAccessKey := d.Get("aws_secured_access_key").(string)
	legacyAccessKey := d.Get("aws_access_key").(string)

	// if securedAccessKey != "" && legacyAccessKey != "" {
	// 	return nil, fmt.Errorf("conflicting configuration arguments: only one of 'aws_access_key' or 'aws_secured_access_key' can be set")
	// }

	// Determine which access key field to use (prefer secured, fallback to legacy)
	accessKey := securedAccessKey
	if accessKey == "" {
		accessKey = legacyAccessKey
	}

	account := &models.V1AwsAccount{
		Metadata: &models.V1ObjectMeta{
			Name: d.Get("name").(string),
			UID:  d.Id(),
		},
		Spec: &models.V1AwsCloudAccount{
			AccessKey: accessKey,
			SecretKey: d.Get("aws_secret_key").(string),
		},
	}
	if d.Get("context") != nil {
		ctxAnnotation := map[string]string{
			"scope":     d.Get("context").(string),
			OverlordUID: d.Get("private_cloud_gateway_id").(string),
		}
		account.Metadata.Annotations = ctxAnnotation
	}
	if len(d.Get("type").(string)) == 0 || d.Get("type").(string) == "secret" {
		account.Spec.CredentialType = models.V1AwsCloudAccountCredentialTypeSecret.Pointer()
		account.Spec.AccessKey = accessKey
		account.Spec.SecretKey = d.Get("aws_secret_key").(string)
	} else if d.Get("type").(string) == "sts" {
		account.Spec.CredentialType = models.V1AwsCloudAccountCredentialTypeSts.Pointer()
		account.Spec.Sts = &models.V1AwsStsCredentials{
			Arn:        d.Get("arn").(string),
			ExternalID: d.Get("external_id").(string),
		}
	} else if d.Get("type").(string) == "pod-identity" {
		account.Spec.CredentialType = models.V1AwsCloudAccountCredentialTypePodDashIdentity.Pointer()
		account.Spec.PodIdentity = &models.V1AwsPodIdentityCredentials{
			RoleArn:               d.Get("role_arn").(string),
			PermissionBoundaryArn: d.Get("permission_boundary_arn").(string),
		}
	}

	// add partition to account
	if d.Get("partition") != nil {
		account.Spec.Partition = types.Ptr(d.Get("partition").(string))
	}

	// add policy arns to account
	if d.Get("policy_arns") != nil && len(d.Get("policy_arns").(*schema.Set).List()) > 0 {
		policyArns := d.Get("policy_arns").(*schema.Set).List()
		policies := make([]string, 0)
		for _, v := range policyArns {
			policies = append(policies, v.(string))
		}
		account.Spec.PolicyARNs = policies
	}

	return account, nil
}

func flattenCloudAccountAws(d *schema.ResourceData, account *models.V1AwsAccount) (diag.Diagnostics, bool) {
	if err := d.Set("name", account.Metadata.Name); err != nil {
		return diag.FromErr(err), true
	}
	if err := d.Set("context", account.Metadata.Annotations["scope"]); err != nil {
		return diag.FromErr(err), true
	}
	if err := d.Set("private_cloud_gateway_id", account.Metadata.Annotations[OverlordUID]); err != nil {
		return diag.FromErr(err), true
	}
	if err := d.Set("type", account.Spec.CredentialType); err != nil {
		return diag.FromErr(err), true
	}
	switch *account.Spec.CredentialType {
	case models.V1AwsCloudAccountCredentialTypeSecret:
		// Set the access key to the appropriate field based on which one is currently in use
		// Prefer aws_secured_access_key if it was set, otherwise use aws_access_key for backward compatibility
		if d.Get("aws_secured_access_key").(string) != "" {
			if err := d.Set("aws_secured_access_key", account.Spec.AccessKey); err != nil {
				return diag.FromErr(err), true
			}
			// Clear the conflicting field to avoid conflicts
			if err := d.Set("aws_access_key", ""); err != nil {
				return diag.FromErr(err), true
			}
		} else {
			if err := d.Set("aws_access_key", account.Spec.AccessKey); err != nil {
				return diag.FromErr(err), true
			}
			// Clear the conflicting field to avoid conflicts
			if err := d.Set("aws_secured_access_key", ""); err != nil {
				return diag.FromErr(err), true
			}
		}
	case models.V1AwsCloudAccountCredentialTypeSts:
		if err := d.Set("arn", account.Spec.Sts.Arn); err != nil {
			return diag.FromErr(err), true
		}
	case models.V1AwsCloudAccountCredentialTypePodDashIdentity:
		if err := d.Set("role_arn", account.Spec.PodIdentity.RoleArn); err != nil {
			return diag.FromErr(err), true
		}
		if err := d.Set("permission_boundary_arn", account.Spec.PodIdentity.PermissionBoundaryArn); err != nil {
			return diag.FromErr(err), true
		}
	}
	if account.Spec.Partition != nil {
		if err := d.Set("partition", account.Spec.Partition); err != nil {
			return diag.FromErr(err), true
		}
	}
	if account.Spec.PolicyARNs != nil {
		if err := d.Set("policy_arns", account.Spec.PolicyARNs); err != nil {
			return diag.FromErr(err), true
		}
	}

	return nil, false
}
