package spectrocloud

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/spectrocloud/palette-sdk-go/api/models"
)

func dataSourceSSHKey() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceSSHKeyRead,
		Description: "The SSH key data source allows you to retrieve information about SSH keys in Palette.",

		Schema: map[string]*schema.Schema{
			"id": {
				Type:          schema.TypeString,
				Computed:      true,
				Optional:      true,
				ConflictsWith: []string{"name"},
				Description:   "The Id of the SSH key resource.",
			},
			"name": {
				Type:        schema.TypeString,
				Computed:    true,
				Optional:    true,
				Description: "The name of the SSH key resource.",
			},
			"ssh_key": {
				Type:        schema.TypeString,
				Computed:    true,
				Sensitive:   true,
				Description: "The SSH key value. This is the public key that was uploaded to Palette.",
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

func dataSourceSSHKeyRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	sshKeyContext := d.Get("context").(string)
	c := getV1ClientWithResourceContext(m, sshKeyContext)
	var diags diag.Diagnostics
	id := d.Get("id").(string)
	name := d.Get("name").(string)
	var sshKey *models.V1UserAssetSSH
	var err error
	if id != "" {
		sshKey, err = c.GetSSHKey(d.Id())
		if err != nil {
			return handleReadError(d, err, diags)
		}
	} else if name != "" {
		sshKey, err = c.GetSSHKeyByName(name)
		if err != nil {
			return handleReadError(d, err, diags)
		}
	}
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(sshKey.Metadata.UID)
	err = d.Set("name", sshKey.Metadata.Name)
	if err != nil {
		return nil
	}
	err = d.Set("ssh_key", sshKey.Spec.PublicKey)
	if err != nil {
		return nil
	}
	return diags
}
