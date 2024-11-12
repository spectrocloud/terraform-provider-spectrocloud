package spectrocloud

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/spectrocloud/palette-sdk-go/api/models"

	"github.com/spectrocloud/terraform-provider-spectrocloud/util/ptr"
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
				Description: "The AWS access key used to authenticate.",
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
				ValidateFunc: validation.StringInSlice([]string{"secret", "sts"}, false),
				Default:      "secret",
				Description:  "The type of AWS credentials to use. Can be `secret` or `sts`. ",
			},
			"arn": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The Amazon Resource Name (ARN) associated with the AWS resource. This is used for identifying resources in AWS.",
			},
			"external_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				Description: "An optional external ID that can be used for cross-account access in AWS.",
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
		return diag.FromErr(err)
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
	account := &models.V1AwsAccount{
		Metadata: &models.V1ObjectMeta{
			Name: d.Get("name").(string),
			UID:  d.Id(),
		},
		Spec: &models.V1AwsCloudAccount{
			AccessKey: d.Get("aws_access_key").(string),
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
		account.Spec.CredentialType = models.V1AwsCloudAccountCredentialTypeSecret
		account.Spec.AccessKey = d.Get("aws_access_key").(string)
		account.Spec.SecretKey = d.Get("aws_secret_key").(string)
	} else if d.Get("type").(string) == "sts" {
		account.Spec.CredentialType = models.V1AwsCloudAccountCredentialTypeSts
		account.Spec.Sts = &models.V1AwsStsCredentials{
			Arn:        d.Get("arn").(string),
			ExternalID: d.Get("external_id").(string),
		}
	}

	// add partition to account
	if d.Get("partition") != nil {
		account.Spec.Partition = ptr.To(d.Get("partition").(string))
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
	if account.Spec.CredentialType == models.V1AwsCloudAccountCredentialTypeSecret {
		if err := d.Set("aws_access_key", account.Spec.AccessKey); err != nil {
			return diag.FromErr(err), true
		}
	} else {
		if err := d.Set("arn", account.Spec.Sts.Arn); err != nil {
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
