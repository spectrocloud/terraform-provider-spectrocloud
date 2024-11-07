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
			"teams": {
				Type:        schema.TypeList,
				Optional:    true,
				ForceNew:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "The team id's assigned to the user.",
			},
			"project_role_mapping": {
				Type:          schema.TypeSet,
				Set:           resourceProjectRoleMappingHash,
				Optional:      true,
				ConflictsWith: []string{"teams"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Project id to be associated with the user.",
						},
						"roles": {
							Type:     schema.TypeSet,
							Required: true,
							Set:      schema.HashString,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
							Description: "List of project role ids to be associated with the user. ",
						},
					},
				},
				Description: "List of project roles to be associated with the user. ",
			},
			"tenant_role_mapping": {
				Type:          schema.TypeSet,
				Optional:      true,
				ConflictsWith: []string{"teams"},
				Set:           schema.HashString,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Description: "List of tenant role ids to be associated with the user. ",
			},
			"workspace_role_mapping": {
				Type:          schema.TypeSet,
				ConflictsWith: []string{"teams"},
				Set:           resourceWorkspaceRoleMappingHash,
				Optional:      true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Project id to be associated with the user.",
						},
						"workspace": {
							Type:     schema.TypeSet,
							Set:      resourceProjectRoleMappingHash,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"id": {
										Type:        schema.TypeString,
										Required:    true,
										Description: "Workspace id to be associated with the user.",
									},
									"roles": {
										Type:     schema.TypeSet,
										Set:      schema.HashString,
										Optional: true,
										Elem: &schema.Schema{
											Type: schema.TypeString,
										},
										Description: "List of workspace role ids to be associated with the user.",
									},
								},
							},
							Description: "List of workspace roles to be associated with the user. ",
						},
					},
				},
				Description: "List of workspace roles to be associated with the user. ",
			},
			"resource_role_mapping": {
				Type:          schema.TypeSet,
				ConflictsWith: []string{"teams"},
				Set:           resourceWorkspaceRoleMappingHash,
				Optional:      true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Project id to be associated with the user.",
						},
						"filters": {
							Type:     schema.TypeSet,
							Set:      schema.HashString,
							Optional: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
							Description: "List of filter ids.",
						},
						"roles": {
							Type:     schema.TypeSet,
							Set:      schema.HashString,
							Optional: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
							Description: "List of resource role ids to be associated with the user.",
						},
					},
				},
			},
		},
	}
}

func resourceUserCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	c := getV1ClientWithResourceContext(m, "tenant")
	var diags diag.Diagnostics

	uid, err := c.CreateUser(toUser(d))
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(uid)

	return diags
}

func toProjectRoleMapping(d *schema.ResourceData) *models.V1ProjectRolesEntity {
	if projectRoles, ok := d.GetOk("project_role_mapping"); ok && projectRoles != nil {
		var role *models.V1ProjectRolesEntity
		projects := make([]*models.V1UIDRoleSummary, 0)
		for _, r := range projectRoles.([]interface{}) {
			rids := make([]*models.V1UIDSummary, 0)
			for _, id := range r.(map[string]interface{})["roles"].(*schema.Set).List() {
				rids = append(rids, &models.V1UIDSummary{
					UID: id.(string),
				})
			}

			projects = append(projects, &models.V1UIDRoleSummary{
				Roles: rids,
				UID:   r.(map[string]interface{})["id"].(string),
			})
		}
		role.Projects = projects
		return role
	}

	return nil
}

func resourceUserRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	c := getV1ClientWithResourceContext(m, "tenant")
	var diags diag.Diagnostics

	email := d.Get("email").(string)
	user, err := c.GetUserSummaryByEmail(email)
	if err != nil {
		return diag.FromErr(err)
	} else if user == nil {
		// Deleted - Terraform will recreate it
		d.SetId("")
		return diags
	}
	err = flattenUser(user, d)
	if err != nil {
		return diag.FromErr(err)
	}
	return diags
}

func resourceUserUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	//c := getV1ClientWithResourceContext(m, "tenant")
	var diags diag.Diagnostics

	if d.HasChanges("project_role_mapping") {

	}
	if d.HasChanges("tenant_role_mapping") {

	}
	if d.HasChanges("workspace_role_mapping") {

	}
	if d.HasChanges("resource_role_mapping") {

	}

	return diags
}

func resourceUserDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	c := getV1ClientWithResourceContext(m, "tenant")
	var diags diag.Diagnostics

	err := c.DeleteUser(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func flattenUser(user *models.V1UserSummary, d *schema.ResourceData) error {
	if user != nil {
		if err := d.Set("first_name", user.Spec.FirstName); err != nil {
			return err
		}
		if err := d.Set("last_name", user.Spec.LastName); err != nil {
			return err
		}
		if err := d.Set("email", user.Spec.EmailID); err != nil {
			return err
		}
		if user.Spec.Roles != nil {
			var roleIds []string
			for _, role := range user.Spec.Roles {
				roleIds = append(roleIds, role.UID)
			}
			if err := d.Set("role_ids", roleIds); err != nil {
				return err
			}
		}
		if user.Spec.Teams != nil {
			var teamIds []string
			for _, team := range user.Spec.Teams {
				teamIds = append(teamIds, team.UID)
			}
			if err := d.Set("teams", teamIds); err != nil {
				return err
			}
		}
	}

	return nil
}

func toUser(d *schema.ResourceData) *models.V1UserEntity {
	return &models.V1UserEntity{
		Metadata: &models.V1ObjectMeta{},
		Spec: &models.V1UserSpecEntity{
			EmailID:   d.Get("email").(string),
			FirstName: d.Get("first_name").(string),
			LastName:  d.Get("last_name").(string),
			Teams:     d.Get("teams").([]string),
		},
	}
}
