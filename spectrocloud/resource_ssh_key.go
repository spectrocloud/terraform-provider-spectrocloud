package spectrocloud

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/spectrocloud/palette-sdk-go/api/models"
)

func resourceSSHKey() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceSSHKeyCreate,
		ReadContext:   resourceSSHKeyRead,
		UpdateContext: resourceSSHKeyUpdate,
		DeleteContext: resourceSSHKeyDelete,
		Description:  "The SSH key resource allows you to manage SSH keys in Palette.",

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
				Description: "The name of the SSH key resource.",
			},
			"ssh_key": {
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   true,
				Description: "The SSH key value. This is the public key that will be used to access the cluster. Must be a valid RSA or DSA public key in PEM format.",
			},
			"context": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "project",
				ValidateFunc: validation.StringInSlice([]string{"project", "tenant"}, false),
				Description: "The context of the cluster profile. Allowed values are `project` or `tenant`. " +
					"Default value is `project`. " + PROJECT_NAME_NUANCE,
			},
		},
	}
}

func toSSHKey(d *schema.ResourceData) (*models.V1UserAssetSSH, error) {
	return &models.V1UserAssetSSH{
		Metadata: &models.V1ObjectMeta{
			Name: d.Get("name").(string),
		},
		Spec: &models.V1UserAssetSSHSpec{
			PublicKey: d.Get("ssh_key").(string),
		},
	}, nil
}

func flattenSSHKey(sshKey *models.V1UserAssetSSH, d *schema.ResourceData) error {
	err := d.Set("name", sshKey.Metadata.Name)
	if err != nil {
		return err
	}
	err = d.Set("ssh_key", sshKey.Spec.PublicKey)
	if err != nil {
		return err
	}
	return nil
}

func resourceSSHKeyCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	sshKeyContext := d.Get("context").(string)
	c := getV1ClientWithResourceContext(m, sshKeyContext)
	var diags diag.Diagnostics
	sshKey, err := toSSHKey(d)
	if err != nil {
		return diag.FromErr(err)
	}
	uid, err := c.CreateSSHKey(sshKey)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(uid)
	return diags
}

func resourceSSHKeyRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	sshKeyContext := d.Get("context").(string)
	c := getV1ClientWithResourceContext(m, sshKeyContext)
	var diags diag.Diagnostics

	sshKey, err := c.GetSSHKey(d.Id())
	if err != nil {
		return diag.FromErr(err)
	} else if sshKey == nil {
		// Deleted - Terraform will recreate it
		d.SetId("")
		return diags
	}
	err = flattenSSHKey(sshKey, d)
	if err != nil {
		return diag.FromErr(err)
	}
	return diags
}

func resourceSSHKeyUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	sshKeyContext := d.Get("context").(string)
	c := getV1ClientWithResourceContext(m, sshKeyContext)
	var diags diag.Diagnostics
	sshKey, err := toSSHKey(d)
	if err != nil {
		return diag.FromErr(err)
	}
	err = c.UpdateSSHKey(d.Id(), sshKey)
	if err != nil {
		return diag.FromErr(err)
	}
	return diags
}

func resourceSSHKeyDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	sshKeyContext := d.Get("context").(string)
	c := getV1ClientWithResourceContext(m, sshKeyContext)
	var diags diag.Diagnostics

	err := c.DeleteSSHKey(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	return diags
}
