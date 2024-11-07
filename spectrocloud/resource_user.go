package spectrocloud

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"regexp"
	"time"

	"github.com/spectrocloud/palette-sdk-go/api/models"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceUser() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceUserCreate,
		ReadContext:   resourceUserRead,
		UpdateContext: resourceUserUpdate,
		DeleteContext: resourceUserDelete,
		Description:   "Create and manage projects in Palette.",

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		SchemaVersion: 2,
		Schema: map[string]*schema.Schema{
			"first_name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The first name of the user.",
			},
			"last_name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The last name of the user.",
			},
			"email": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringMatch(
					regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`),
					"must be a valid email address",
				),
				Description: "The email of the user.",
			},
			"role_ids": {
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "The roles id's assigned to the user.",
			},
			"team_ids": {
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "The roles id's assigned to the user.",
			},
		},
	}
}

func resourceUserCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	c := getV1ClientWithResourceContext(m, "")
	var diags diag.Diagnostics

	uid, err := c.CreateUser(toUser(d))
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(uid)

	return diags
}

func resourceUserRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	//c := getV1ClientWithResourceContext(m, "")
	var diags diag.Diagnostics

	//project, err := c.GetProject(d.Id())
	//if err != nil {
	//	return diag.FromErr(err)
	//} else if project == nil {
	//	// Deleted - Terraform will recreate it
	//	d.SetId("")
	//	return diags
	//}
	//
	//if err := d.Set("name", project.Metadata.Name); err != nil {
	//	return diag.FromErr(err)
	//}
	//
	//if v, found := project.Metadata.Annotations["description"]; found {
	//	if err := d.Set("description", v); err != nil {
	//		return diag.FromErr(err)
	//	}
	//}
	//
	//if err := d.Set("tags", flattenTags(project.Metadata.Labels)); err != nil {
	//	return diag.FromErr(err)
	//}

	return diags
}

func resourceUserUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	//c := getV1ClientWithResourceContext(m, "")
	var diags diag.Diagnostics

	//err := c.UpdateProject(d.Id(), toProject(d))
	//if err != nil {
	//	return diag.FromErr(err)
	//}
	return diags
}

func resourceUserDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	//c := getV1ClientWithResourceContext(m, "")
	var diags diag.Diagnostics

	//err := c.DeleteProject(d.Id())
	//if err != nil {
	//	return diag.FromErr(err)
	//}

	return diags
}

func toUser(d *schema.ResourceData) *models.V1UserEntity {
	return &models.V1UserEntity{
		Metadata: &models.V1ObjectMeta{},
		Spec: &models.V1UserSpecEntity{
			EmailID:   d.Get("email").(string),
			FirstName: d.Get("first_name").(string),
			LastName:  d.Get("last_name").(string),
			Roles:     nil,
			Teams:     nil,
		},
	}
}
