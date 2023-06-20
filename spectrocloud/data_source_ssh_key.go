package spectrocloud

import (
	"context"
	"encoding/base64"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/spectrocloud/palette-sdk-go/client"
)

func dataSourceSSHKey() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceSSHKeyRead,

		Schema: map[string]*schema.Schema{
			"id": {
				Type:          schema.TypeString,
				Computed:      true,
				Optional:      true,
				ConflictsWith: []string{"name"},
			},
			"name": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
			},
			"ssh_key": {
				Type:      schema.TypeString,
				Computed:  true,
				Optional:  true,
				Sensitive: true,
				StateFunc: func(val any) string {
					return base64.StdEncoding.EncodeToString([]byte(val.(string)))
				},
			},
			"context": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "project",
				ValidateFunc: validation.StringInSlice([]string{"", "project", "tenant"}, false),
			},
		},
	}
}

func dataSourceSSHKeyRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.V1Client)
	var diags diag.Diagnostics
	if v, ok := d.GetOk("name"); ok {
		SSHKey, err := c.GetSSHKeyByName(v.(string), d.Get("context").(string))
		if err != nil {
			return diag.FromErr(err)
		}
		d.SetId(SSHKey.Metadata.UID)
		if err := d.Set("name", SSHKey.Metadata.Name); err != nil {
			return diag.FromErr(err)
		}
		if err := d.Set("ssh_key", base64.StdEncoding.EncodeToString([]byte(SSHKey.Spec.PublicKey))); err != nil {
			return diag.FromErr(err)
		}
	}
	return diags
}
