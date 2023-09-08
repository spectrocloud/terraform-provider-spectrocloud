package spectrocloud

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/spectrocloud/hapi/models"
	"github.com/spectrocloud/palette-sdk-go/client"

	"github.com/spectrocloud/terraform-provider-spectrocloud/types"
)

func resourceCloudAccountAws() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceCloudAccountAwsCreate,
		ReadContext:   resourceCloudAccountAwsRead,
		UpdateContext: resourceCloudAccountAwsUpdate,
		DeleteContext: resourceCloudAccountAwsDelete,
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
				Description:  "The context of the AWS configuration. Can be `project` or `tenant`.",
			},
			"aws_access_key": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"aws_secret_key": {
				Type:      schema.TypeString,
				Optional:  true,
				Sensitive: true,
			},
			"type": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"secret", "sts"}, false),
				Default:      "secret",
				Description:  "The type of AWS credentials to use. Can be `secret` or `sts`. ",
			},
			"arn": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"external_id": {
				Type:      schema.TypeString,
				Optional:  true,
				Sensitive: true,
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
	c := m.(*client.V1Client)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	account, err := toAwsAccount(d)
	if err != nil {
		return diag.FromErr(err)
	}

	AccountContext := d.Get("context").(string)
	uid, err := c.CreateCloudAccountAws(account, AccountContext)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(uid)

	resourceCloudAccountAwsRead(ctx, d, m)

	return diags
}

func resourceCloudAccountAwsRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.V1Client)

	var diags diag.Diagnostics

	uid := d.Id()

	AccountContext := d.Get("context").(string)
	account, err := c.GetCloudAccountAws(uid, AccountContext)
	if err != nil {
		return diag.FromErr(err)
	} else if account == nil {
		// Deleted - Terraform will recreate it
		d.SetId("")
		return diags
	}

	if err := d.Set("name", account.Metadata.Name); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("type", string(account.Spec.CredentialType)); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("context", account.Metadata.Annotations["scope"]); err != nil {
		return diag.FromErr(err)
	}
	if account.Spec.CredentialType == models.V1AwsCloudAccountCredentialTypeSecret {
		if err := d.Set("aws_access_key", account.Spec.AccessKey); err != nil {
			return diag.FromErr(err)
		}
	} else {
		if err := d.Set("arn", account.Spec.Sts.Arn); err != nil {
			return diag.FromErr(err)
		}
	}
	if account.Spec.Partition != nil {
		if err := d.Set("partition", account.Spec.Partition); err != nil {
			return diag.FromErr(err)
		}
	}
	if account.Spec.PolicyARNs != nil {
		if err := d.Set("policy_arns", account.Spec.PolicyARNs); err != nil {
			return diag.FromErr(err)
		}
	}

	return diags
}

func resourceCloudAccountAwsUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.V1Client)

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
	c := m.(*client.V1Client)

	var diags diag.Diagnostics

	cloudAccountID := d.Id()

	AccountContext := d.Get("context").(string)
	err := c.DeleteCloudAccountAws(cloudAccountID, AccountContext)
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
			"scope": d.Get("context").(string),
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
