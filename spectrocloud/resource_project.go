package spectrocloud

import (
	"context"
	"time"

	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/spectrocloud/palette-sdk-go/client/herr"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceProject() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceProjectCreate,
		ReadContext:   resourceProjectRead,
		UpdateContext: resourceProjectUpdate,
		DeleteContext: resourceProjectDelete,
		Description:   "Create and manage projects in Palette.",

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
				Description: "The name of the project.",
			},
			"tags": {
				Type:     schema.TypeSet,
				Optional: true,
				Set:      schema.HashString,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Description: "Assign tags to the project.",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The description of the project.",
			},
		},
	}
}

func resourceProjectCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	c := getV1ClientWithResourceContext(m, "")
	var diags diag.Diagnostics

	uid, err := c.CreateProject(toProject(d))
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(uid)

	return diags
}

func resourceProjectRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	c := getV1ClientWithResourceContext(m, "")
	var diags diag.Diagnostics

	project, err := c.GetProject(d.Id())
	if err != nil {
		if herr.IsNotFound(err) {
			d.SetId("")
			return diags
		}
		return diag.FromErr(err)
	} else if project == nil {
		// Deleted - Terraform will recreate it
		d.SetId("")
		return diags
	}

	if err := d.Set("name", project.Metadata.Name); err != nil {
		return diag.FromErr(err)
	}

	if v, found := project.Metadata.Annotations["description"]; found {
		if err := d.Set("description", v); err != nil {
			return diag.FromErr(err)
		}
	}

	if err := d.Set("tags", flattenTags(project.Metadata.Labels)); err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func resourceProjectUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	c := getV1ClientWithResourceContext(m, "")
	var diags diag.Diagnostics

	err := c.UpdateProject(d.Id(), toProject(d))
	if err != nil {
		return diag.FromErr(err)
	}
	return diags
}

func resourceProjectDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	c := getV1ClientWithResourceContext(m, "")
	var diags diag.Diagnostics

	err := c.DeleteProject(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func toProject(d *schema.ResourceData) *models.V1ProjectEntity {
	annotations := make(map[string]string)
	if len(d.Get("description").(string)) > 0 {
		annotations["description"] = d.Get("description").(string)
	}
	return &models.V1ProjectEntity{
		Metadata: &models.V1ObjectMeta{
			Name:        d.Get("name").(string),
			UID:         d.Id(),
			Labels:      toTags(d),
			Annotations: annotations,
		},
	}
}
