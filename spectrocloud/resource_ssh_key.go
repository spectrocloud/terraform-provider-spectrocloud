package spectrocloud

import (
	"context"
	"encoding/base64"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/spectrocloud/hapi/models"
	"github.com/spectrocloud/palette-sdk-go/client"
	"time"
)

func resourceSSHKey() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceSSHKeyCreate,
		ReadContext:   resourceSSHKeyRead,
		UpdateContext: resourceSSHKeyUpdate,
		DeleteContext: resourceSSHKeyDelete,
		Description:   "A resource for creating and managing ssh keys.",
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
				Description: "Name of the ssh key. This name is used to identify the ssh key in the Spectro Cloud UI.",
			},
			"ssh_key": {
				Type:      schema.TypeString,
				Required:  true,
				Sensitive: true,
				StateFunc: func(val any) string {
					return base64.StdEncoding.EncodeToString([]byte(val.(string)))
				},
				Description: "The SSH key to be used for the cluster. This is the public key that will be used to access the cluster.",
			},
			"context": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Default:      "project",
				ValidateFunc: validation.StringInSlice([]string{"", "project", "tenant"}, false),
				Description:  "Context of the ssh key. This can be either project or tenant. If not specified, the default value is `project`.",
			},
		},
	}
}

func resourceSSHKeyCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.V1Client)
	var diags diag.Diagnostics
	SSHKeyContext := d.Get("context").(string)
	uid, err := c.CreateSSHKey(toSSHKey(d), SSHKeyContext)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(uid)

	return diags
}

func resourceSSHKeyRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.V1Client)
	var diags diag.Diagnostics
	SSHKeyContext := d.Get("context").(string)
	SSHKey, err := c.GetSSHKeyByUID(d.Id(), SSHKeyContext)
	if err != nil {
		return diag.FromErr(err)
	} else if SSHKey == nil {
		// Deleted - Terraform will recreate it
		d.SetId("")
		return diags
	}

	if err := d.Set("name", SSHKey.Metadata.Name); err != nil {
		return diag.FromErr(err)
	}
	// Setting back public ssh key into sate file leads to security break hence commenting it out
	//if err := d.Set("ssh_key", base64.StdEncoding.EncodeToString([]byte(SSHKey.Spec.PublicKey))); err != nil {
	//	return diag.FromErr(err)
	//}

	return diags

}

func resourceSSHKeyUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.V1Client)
	var diags diag.Diagnostics
	SSHKeyContext := d.Get("context").(string)
	err := c.UpdateSSHKey(d.Id(), toSSHKey(d), SSHKeyContext)
	if err != nil {
		return diag.FromErr(err)
	}
	return diags
}

func resourceSSHKeyDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.V1Client)
	var diags diag.Diagnostics
	SSHKeyContext := d.Get("context").(string)
	err := c.DeleteSSHKey(d.Id(), SSHKeyContext)
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func toSSHKey(d *schema.ResourceData) *models.V1UserAssetSSH {
	return &models.V1UserAssetSSH{
		Metadata: &models.V1ObjectMeta{
			Name: d.Get("name").(string),
		},
		Spec: &models.V1UserAssetSSHSpec{
			PublicKey: d.Get("ssh_key").(string),
		},
	}
}
