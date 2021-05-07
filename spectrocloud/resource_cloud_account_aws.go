package spectrocloud

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/hashicorp/go-cty/cty"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/hapi/models"
	"github.com/spectrocloud/terraform-provider-spectrocloud/pkg/client"
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
				ValidateFunc: validation.StringInSlice([]string{"secret", "sts"}, true),
				Default:      "secret",
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
		},
	}
}

func resourceCloudAccountAwsCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.V1alpha1Client)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	account := toAwsAccount(d)

	uid, err := c.CreateCloudAccountAws(account)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(uid)

	resourceCloudAccountAwsRead(ctx, d, m)

	return diags
}

func resourceCloudAccountAwsRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.V1alpha1Client)

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

	if err := d.Set("name", account.Metadata.Name); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("type", string(account.Spec.CredentialType)); err != nil {
		return diag.FromErr(err)
	}

	if account.Spec.CredentialType == models.V1alpha1AwsCloudAccountCredentialTypeSecret {
		if err := d.Set("aws_access_key", account.Spec.AccessKey); err != nil {
			return diag.FromErr(err)
		}
	} else {
		if err := d.Set("arn", account.Spec.AwsStsCredentials.Arn); err != nil {
			return diag.FromErr(err)
		}
	}

	return diags
}

//
func resourceCloudAccountAwsUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.V1alpha1Client)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	account := toAwsAccount(d)

	err := c.UpdateCloudAccountAws(account)
	if err != nil {
		return diag.FromErr(err)
	}

	resourceCloudAccountAwsRead(ctx, d, m)

	return diags
}

func resourceCloudAccountAwsDelete(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.V1alpha1Client)

	var diags diag.Diagnostics

	cloudAccountID := d.Id()

	err := c.DeleteCloudAccountAws(cloudAccountID)
	if err != nil {
		return diag.FromErr(err)
	}

	// d.SetId("") is automatically called assuming delete returns no errors

	return diags
}

func toAwsAccount(d *schema.ResourceData) *models.V1alpha1AwsAccount {
	account := &models.V1alpha1AwsAccount{
		Metadata: &models.V1ObjectMeta{
			Name: d.Get("name").(string),
			UID:  d.Id(),
		},
		Spec: &models.V1alpha1AwsCloudAccount{
			AccessKey: d.Get("aws_access_key").(string),
			SecretKey: d.Get("aws_secret_key").(string),
		},
	}
	if len(d.Get("type").(string)) == 0 || d.Get("type").(string) == "secret" {
		account.Spec.CredentialType = models.V1alpha1AwsCloudAccountCredentialTypeSecret
		account.Spec.AccessKey = d.Get("aws_access_key").(string)
		account.Spec.SecretKey = d.Get("aws_secret_key").(string)
	} else if d.Get("type").(string) == "sts" {
		account.Spec.CredentialType = models.V1alpha1AwsCloudAccountCredentialTypeSts
		account.Spec.AwsStsCredentials = &models.V1alpha1AwsStsCredentials{
			Arn:        d.Get("arn").(string),
			ExternalID: d.Get("external_id").(string),
		}
	}

	return account
}

func validateAwsCloudAccountType(data interface{}, path cty.Path) diag.Diagnostics {
	var diags diag.Diagnostics
	accType := data.(string)
	for _, accessType := range []string{"secret", "sts"} {
		if accessType == accType {
			return diags
		}
	}
	return diag.FromErr(fmt.Errorf("aws cloud account type '%s' is invalid. valid aws cloud account types are 'secret' and 'sts'", accType))
}
