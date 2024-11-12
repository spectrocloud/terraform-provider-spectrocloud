package spectrocloud

import (
	"bytes"
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"regexp"
	"sort"
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
			"project_role": {
				Type:          schema.TypeSet,
				Set:           resourceUserProjectRoleMappingHash,
				Optional:      true,
				ConflictsWith: []string{"teams"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"project_id": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Project id to be associated with the user.",
						},
						"role_ids": {
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
			"tenant_role": {
				Type:          schema.TypeSet,
				Optional:      true,
				ConflictsWith: []string{"teams"},
				Set:           schema.HashString,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Description: "List of tenant role ids to be associated with the user. ",
			},
			"workspace_role": {
				Type:          schema.TypeSet,
				ConflictsWith: []string{"teams"},
				Set:           resourceUserWorkspaceRoleMappingHash,
				Optional:      true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"project_id": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Project id to be associated with the user.",
						},
						"workspace": {
							Type:     schema.TypeSet,
							Set:      resourceUserWorkspaceRoleMappingHashInternal,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"id": {
										Type:        schema.TypeString,
										Required:    true,
										Description: "Workspace id to be associated with the user.",
									},
									"role_ids": {
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
			"resource_role": {
				Type:          schema.TypeSet,
				ConflictsWith: []string{"teams"},
				Set:           resourceUserResourceRoleMappingHash,
				Optional:      true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"project_ids": {
							Type:     schema.TypeSet,
							Set:      schema.HashString,
							Required: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
							Description: "Project id's to be associated with the user.",
						},
						"filter_ids": {
							Type:     schema.TypeSet,
							Set:      schema.HashString,
							Required: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
							Description: "List of filter ids.",
						},
						"role_ids": {
							Type:     schema.TypeSet,
							Set:      schema.HashString,
							Required: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
							Description: "List of resource role ids to be associated with the user.",
						},
						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func resourceUserResourceRoleMappingHash(i interface{}) int {
	var buf bytes.Buffer
	m := i.(map[string]interface{})

	// Sort the roles to ensure order does not affect the hash
	pids := make([]string, len(m["project_ids"].(*schema.Set).List()))
	for i, pid := range m["project_ids"].(*schema.Set).List() {
		pids[i] = pid.(string)
	}
	sort.Strings(pids)

	fids := make([]string, len(m["filter_ids"].(*schema.Set).List()))
	for i, fid := range m["filter_ids"].(*schema.Set).List() {
		fids[i] = fid.(string)
	}
	sort.Strings(fids)

	rids := make([]string, len(m["role_ids"].(*schema.Set).List()))
	for i, rid := range m["role_ids"].(*schema.Set).List() {
		rids[i] = rid.(string)
	}
	sort.Strings(rids)

	//buf.WriteString(fmt.Sprintf("%s-", m["project_id"].(string)))

	for _, id := range pids {
		buf.WriteString(fmt.Sprintf("%s-", id))
	}
	for _, id := range fids {
		buf.WriteString(fmt.Sprintf("%s-", id))
	}
	for _, id := range rids {
		buf.WriteString(fmt.Sprintf("%s-", id))
	}

	return int(hash(buf.String()))
}

func resourceUserWorkspaceRoleMappingHash(i interface{}) int {
	var buf bytes.Buffer
	m := i.(map[string]interface{})

	// Hash project id
	if v, ok := m["project_id"].(string); ok {
		h := schema.HashString(v)
		buf.WriteString(fmt.Sprintf("%d-", h))
	}

	// Hash workspaces
	if v, ok := m["workspace"].(*schema.Set); ok {
		// Sort workspace hashes to ensure consistent ordering
		workspaces := v.List()
		hashes := make([]int, len(workspaces))
		for i, workspaceInterface := range workspaces {
			workspace := workspaceInterface.(map[string]interface{})
			hashes[i] = resourceUserWorkspaceRoleMappingHashInternal(workspace)
		}
		sort.Ints(hashes)

		for _, h := range hashes {
			buf.WriteString(fmt.Sprintf("%d-", h))
		}
	}

	return int(hash(buf.String()))
}

func resourceUserWorkspaceRoleMappingHashInternal(workspace interface{}) int {
	var buf bytes.Buffer
	m := workspace.(map[string]interface{})
	// Sort the roles to ensure order does not affect the hash
	roles := make([]string, len(m["role_ids"].(*schema.Set).List()))
	for i, role := range m["role_ids"].(*schema.Set).List() {
		roles[i] = role.(string)
	}
	sort.Strings(roles)

	buf.WriteString(fmt.Sprintf("%s-", m["id"].(string)))

	for _, role := range roles {
		buf.WriteString(fmt.Sprintf("%s-", role))
	}

	return int(hash(buf.String()))
}

func resourceUserProjectRoleMappingHash(i interface{}) int {
	var buf bytes.Buffer
	m := i.(map[string]interface{})

	// Sort the roles to ensure order does not affect the hash
	roles := make([]string, len(m["role_ids"].(*schema.Set).List()))
	for i, role := range m["role_ids"].(*schema.Set).List() {
		roles[i] = role.(string)
	}
	sort.Strings(roles)

	buf.WriteString(fmt.Sprintf("%s-", m["project_id"].(string)))

	for _, role := range roles {
		buf.WriteString(fmt.Sprintf("%s-", role))
	}

	return int(hash(buf.String()))
}

func resourceUserCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	c := getV1ClientWithResourceContext(m, "tenant")
	var diags diag.Diagnostics
	user := toUser(d)
	uid, err := c.CreateUser(user)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(uid)
	//creating roles
	if pRoles, ok := d.GetOk("project_role"); ok && pRoles != nil {
		projectRole := toUserProjectRoleMapping(d)
		err := c.AssociateUserProjectRole(uid, projectRole)
		if err != nil {
			_ = c.DeleteUser(uid)
			return diag.FromErr(err)
		}
	}

	if rRoles, ok := d.GetOk("tenant_role"); ok && rRoles != nil {
		tenantRole := toUserTenantRoleMapping(d)
		err := c.AssociateUserTenantRole(uid, tenantRole)
		if err != nil {
			_ = c.DeleteUser(uid)
			return diag.FromErr(err)
		}
	}

	if wRoles, ok := d.GetOk("workspace_role"); ok && wRoles != nil {
		workspaceRole := toUserWorkspaceRoleMapping(d)
		err := c.AssociateUserWorkspaceRole(uid, workspaceRole)
		if err != nil {
			_ = c.DeleteUser(uid)
			return diag.FromErr(err)
		}
	}

	if rRoles, ok := d.GetOk("resource_role"); ok && rRoles != nil {
		resourceRoles := toUserResourceRoleMapping(d)
		for _, role := range resourceRoles {
			err := c.CreateUserResourceRole(uid, role)
			if err != nil {
				_ = c.DeleteUser(uid)
				return diag.FromErr(err)
			}
		}
	}

	return diags
}

func toUserResourceRoleMapping(d *schema.ResourceData) []*models.V1ResourceRolesUpdateEntity {
	if resourceRoles, ok := d.GetOk("resource_role"); ok && resourceRoles != nil {
		resourceRoleEntities := make([]*models.V1ResourceRolesUpdateEntity, 0)
		for _, re := range d.Get("resource_role").(*schema.Set).List() {
			resourceEntity := &models.V1ResourceRolesUpdateEntity{
				FilterRefs:  setToStringArray(re.(map[string]interface{})["filter_ids"]),
				ProjectUids: setToStringArray(re.(map[string]interface{})["project_ids"]),
				Roles:       setToStringArray(re.(map[string]interface{})["role_ids"]),
			}
			resourceRoleEntities = append(resourceRoleEntities, resourceEntity)
		}
		return resourceRoleEntities
	}
	return nil
}

func toUserProjectRoleMapping(d *schema.ResourceData) *models.V1ProjectRolesPatch {
	if projectRoles, ok := d.GetOk("project_role"); ok && projectRoles != nil {
		//var role *models.V1ProjectRolesPatch
		var projects []*models.V1ProjectRolesPatchProjectsItems0
		for _, r := range projectRoles.(*schema.Set).List() {
			projects = append(projects, &models.V1ProjectRolesPatchProjectsItems0{
				ProjectUID: r.(map[string]interface{})["project_id"].(string),
				Roles:      setToStringArray(r.(map[string]interface{})["role_ids"]),
			})
		}
		return &models.V1ProjectRolesPatch{
			Projects: projects,
		}
	}

	return nil
}

func toUserTenantRoleMapping(d *schema.ResourceData) *models.V1UserRoleUIDs {
	roles := make([]string, 0)
	if d.Get("tenant_role") != nil {
		for _, role := range d.Get("tenant_role").(*schema.Set).List() {
			roles = append(roles, role.(string))
		}
	}

	return &models.V1UserRoleUIDs{
		Roles: roles,
	}
}

func toUserWorkspaceRoleMapping(d *schema.ResourceData) *models.V1WorkspacesRolesPatch {
	workspaces := make([]*models.V1WorkspaceRolesPatch, 0)
	workspaceRoleMappings := d.Get("workspace_role").(*schema.Set).List()

	for _, mapping := range workspaceRoleMappings {
		data := mapping.(map[string]interface{})

		for _, workspace := range data["workspace"].(*schema.Set).List() {
			workspaceData := workspace.(map[string]interface{})
			roles := make([]string, 0)
			if workspaceData["role_ids"] != nil {
				for _, role := range workspaceData["role_ids"].(*schema.Set).List() {
					roles = append(roles, role.(string))
				}
			}

			workspaces = append(workspaces, &models.V1WorkspaceRolesPatch{
				UID:   workspaceData["id"].(string),
				Roles: roles,
			})
		}

	}

	return &models.V1WorkspacesRolesPatch{
		Workspaces: workspaces,
	}
}

func setToStringArray(ids interface{}) []string {
	idList := make([]string, 0)
	for _, id := range ids.(*schema.Set).List() {
		idList = append(idList, id.(string))
	}
	return idList
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

	c := getV1ClientWithResourceContext(m, "tenant")
	uid := d.Id()
	var diags diag.Diagnostics

	if d.HasChanges("project_role") {
		projectRole := toUserProjectRoleMapping(d)
		err := c.AssociateUserProjectRole(uid, projectRole)
		if err != nil {
			return diag.FromErr(err)
		}
	}
	if d.HasChanges("tenant_role") {
		tenantRole := toUserTenantRoleMapping(d)
		err := c.AssociateUserTenantRole(uid, tenantRole)
		if err != nil {
			return diag.FromErr(err)
		}
	}
	if d.HasChanges("workspace_role") {
		workspaceRole := toUserWorkspaceRoleMapping(d)
		err := c.AssociateUserWorkspaceRole(uid, workspaceRole)
		if err != nil {
			return diag.FromErr(err)
		}
	}
	if d.HasChanges("resource_role") {
		resourceRoles := toUserResourceRoleMapping(d)
		_ = deleteUserResourceRoles(m, uid)
		for _, role := range resourceRoles {
			err := c.CreateUserResourceRole(uid, role)
			if err != nil {
				return diag.FromErr(err)
			}
		}
	}

	return diags
}

func deleteUserResourceRoles(m interface{}, userUID string) error {
	c := getV1ClientWithResourceContext(m, "tenant")
	_, resourceRoles := c.GetUserResourceRoles(userUID)
	for _, re := range resourceRoles {
		_ = c.DeleteUserResourceRoles(userUID, re.UID)
	}
	return nil
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
	fName := d.Get("first_name").(string)
	lName := d.Get("last_name").(string)
	user := &models.V1UserEntity{
		Metadata: &models.V1ObjectMeta{
			Name: fName + " " + lName,
		},
		Spec: &models.V1UserSpecEntity{
			EmailID:   d.Get("email").(string),
			FirstName: fName,
			LastName:  lName,
		},
	}
	if teams, ok := d.GetOk("teams"); ok && teams != nil {
		user.Spec.Teams = teams.([]string)
	}
	return user
}
