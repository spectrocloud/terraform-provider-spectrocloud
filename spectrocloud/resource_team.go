package spectrocloud

import (
	"bytes"
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/spectrocloud/palette-sdk-go/api/models"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/spectrocloud/palette-sdk-go/client"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceTeam() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceTeamCreate,
		ReadContext:   resourceTeamRead,
		UpdateContext: resourceTeamUpdate,
		DeleteContext: resourceTeamDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceTeamImport,
		},

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
				ForceNew:    true,
				Description: "Name of the team. ",
			},
			"users": {
				Type:     schema.TypeSet,
				Optional: true,
				Set:      schema.HashString,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Description: "List of user ids to be associated with the team. ",
			},
			"project_role_mapping": {
				Type:     schema.TypeSet,
				Set:      resourceProjectRoleMappingHash,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Project id to be associated with the team.",
						},
						"roles": {
							Type:     schema.TypeSet,
							Required: true,
							Set:      schema.HashString,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
							Description: "List of project roles to be associated with the team. ",
						},
					},
				},
				Description: "List of project roles to be associated with the team. ",
			},
			"tenant_role_mapping": {
				Type:     schema.TypeSet,
				Optional: true,
				Set:      schema.HashString,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Description: "List of tenant role ids to be associated with the team. ",
			},
			"workspace_role_mapping": {
				Type:     schema.TypeSet,
				Set:      resourceWorkspaceRoleMappingHash,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Project id to be associated with the team.",
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
										Description: "Workspace id to be associated with the team.",
									},
									"roles": {
										Type:     schema.TypeSet,
										Set:      schema.HashString,
										Required: true,
										Elem: &schema.Schema{
											Type: schema.TypeString,
										},
										Description: "List of workspace roles to be associated with the team.",
									},
								},
							},
							Description: "List of workspace roles to be associated with the team. ",
						},
					},
				},
				Description: "List of workspace roles to be associated with the team. ",
			},
		},
	}
}

func resourceTeamCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	c := getV1ClientWithResourceContext(m, "tenant")
	var diags diag.Diagnostics

	uid, err := c.CreateTeam(toTeam(d))
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(uid)

	//associate roles with team
	err = c.AssociateTeamProjectRole(uid, toTeamProjectRoleMapping(d))
	if err != nil {
		return diag.FromErr(err)
	}

	//associate tenant roles with team
	err = c.AssociateTeamTenantRole(uid, toTeamTenantRoleMapping(d))
	if err != nil {
		return diag.FromErr(err)
	}

	//associate workspace roles with team
	err = c.AssociateTeamWorkspaceRole(uid, toTeamWorkspaceRoleMapping(d))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(uid)

	return diags
}

func resourceTeamRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	c := getV1ClientWithResourceContext(m, "tenant")
	var diags diag.Diagnostics

	team, err := c.GetTeam(d.Id())
	if err != nil {
		return handleReadError(d, err, diags)
	} else if team == nil {
		// Deleted - Terraform will recreate it
		d.SetId("")
		return diags
	}

	if err := d.Set("name", team.Metadata.Name); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("users", team.Spec.Users); err != nil {
		return diag.FromErr(err)
	}

	err = setProjectRoles(c, d)
	if err != nil {
		return diag.FromErr(err)
	}

	err = setTenantRoles(c, d)
	if err != nil {
		return diag.FromErr(err)
	}

	err = setWorkspaceRoles(c, d)
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func setProjectRoles(c *client.V1Client, d *schema.ResourceData) error {
	projectRoles, err := c.GetTeamProjectRoleAssociation(d.Id())
	if err != nil {
		return err
	}

	if projectRoles != nil && len(projectRoles.Projects) > 0 {
		mappings := make([]interface{}, 0)
		for _, project := range projectRoles.Projects {
			dataMap := make(map[string]interface{})
			roles := make([]string, 0)
			for _, role := range project.Roles {
				roles = append(roles, role.UID)
			}
			dataMap["id"] = project.UID
			dataMap["roles"] = roles
			mappings = append(mappings, dataMap)
		}
		if err := d.Set("project_role_mapping", mappings); err != nil {
			return err
		}
	}
	return nil
}

func setTenantRoles(c *client.V1Client, d *schema.ResourceData) error {
	tenantRoles, err := c.GetTeamTenantRoleAssociation(d.Id())
	if err != nil {
		return err
	}
	tenantRolesIDs := make([]string, 0)
	for _, role := range tenantRoles.Roles {
		tenantRolesIDs = append(tenantRolesIDs, role.UID)
	}
	if err := d.Set("tenant_role_mapping", tenantRolesIDs); err != nil {
		return err
	}
	return nil
}

func setWorkspaceRoles(c *client.V1Client, d *schema.ResourceData) error {
	workspaceRoles, err := c.GetTeamWorkspaceRoleAssociation(d.Id())
	if err != nil {
		return err
	}

	if workspaceRoles != nil && len(workspaceRoles.Projects) > 0 {
		projects := make([]interface{}, 0)
		for _, project := range workspaceRoles.Projects {
			projectMap := make(map[string]interface{})
			workspaces := make([]interface{}, 0)
			for _, workspace := range project.Workspaces {
				workspaceMap := make(map[string]interface{})
				roles := make([]string, 0)
				for _, role := range workspace.Roles {
					roles = append(roles, role.UID)
				}
				if len(roles) > 0 {
					workspaceMap["id"] = workspace.UID
					workspaceMap["roles"] = roles
					workspaces = append(workspaces, workspaceMap)
				}
			}
			if len(workspaces) > 0 {
				projectMap["id"] = project.UID
				projectMap["workspace"] = workspaces
				projects = append(projects, projectMap)
			}

		}
		if err := d.Set("workspace_role_mapping", projects); err != nil {
			return err
		}
	}
	return nil
}

func resourceTeamUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	c := getV1ClientWithResourceContext(m, "tenant")
	var diags diag.Diagnostics

	err := c.UpdateTeam(d.Id(), toTeam(d))
	if err != nil {
		return diag.FromErr(err)
	}

	//associate roles with team
	err = c.AssociateTeamProjectRole(d.Id(), toTeamProjectRoleMapping(d))
	if err != nil {
		return diag.FromErr(err)
	}

	//associate tenant roles with team
	err = c.AssociateTeamTenantRole(d.Id(), toTeamTenantRoleMapping(d))
	if err != nil {
		return diag.FromErr(err)
	}

	//associate workspace roles with team
	err = c.AssociateTeamWorkspaceRole(d.Id(), toTeamWorkspaceRoleMapping(d))
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func resourceTeamDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := getV1ClientWithResourceContext(m, "tenant")
	var diags diag.Diagnostics

	err := c.DeleteTeam(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func toTeam(d *schema.ResourceData) *models.V1Team {
	userIDs := make([]string, 0)

	if d.Get("users") != nil {
		for _, userID := range d.Get("users").(*schema.Set).List() {
			userIDs = append(userIDs, userID.(string))
		}
	}

	return &models.V1Team{
		Metadata: &models.V1ObjectMeta{
			Name: d.Get("name").(string),
			UID:  d.Id(),
		},
		Spec: &models.V1TeamSpec{
			Users: userIDs,
		},
	}
}

func toTeamProjectRoleMapping(d *schema.ResourceData) *models.V1ProjectRolesPatch {
	projects := make([]*models.V1ProjectRolesPatchProjectsItems0, 0)
	projectRoleMappings := d.Get("project_role_mapping").(*schema.Set).List()
	for _, mapping := range projectRoleMappings {
		data := mapping.(map[string]interface{})

		roles := make([]string, 0)
		if data["roles"] != nil {
			for _, role := range data["roles"].(*schema.Set).List() {
				roles = append(roles, role.(string))
			}
		}

		projects = append(projects, &models.V1ProjectRolesPatchProjectsItems0{
			ProjectUID: data["id"].(string),
			Roles:      roles,
		})
	}

	return &models.V1ProjectRolesPatch{
		Projects: projects,
	}
}

func toTeamTenantRoleMapping(d *schema.ResourceData) *models.V1TeamTenantRolesUpdate {
	roles := make([]string, 0)
	if d.Get("tenant_role_mapping") != nil {
		for _, role := range d.Get("tenant_role_mapping").(*schema.Set).List() {
			roles = append(roles, role.(string))
		}
	}

	return &models.V1TeamTenantRolesUpdate{
		Roles: roles,
	}
}

func toTeamWorkspaceRoleMapping(d *schema.ResourceData) *models.V1WorkspacesRolesPatch {
	oldWS, newWS := d.GetChange("workspace_role_mapping")

	oldList := oldWS.(*schema.Set).List()
	newList := newWS.(*schema.Set).List()

	// Map keyed by "projectID::workspaceID"
	workspaceMap := make(map[string]*models.V1WorkspaceRolesPatch)
	seenNew := make(map[string]struct{})

	// Step 1: Process OLD config (initialize map)
	for _, mapping := range oldList {
		mappingData := mapping.(map[string]interface{})
		projectID := mappingData["id"].(string)

		for _, workspace := range mappingData["workspace"].(*schema.Set).List() {
			workspaceData := workspace.(map[string]interface{})
			workspaceID := workspaceData["id"].(string)
			key := fmt.Sprintf("%s::%s", projectID, workspaceID)

			workspaceMap[key] = &models.V1WorkspaceRolesPatch{
				UID:   workspaceID,
				Roles: []string{}, // Assume removed unless new overrides
			}
		}
	}

	// Step 2: Process NEW config (override old or populate fresh)
	for _, mapping := range newList {
		mappingData := mapping.(map[string]interface{})
		projectID := mappingData["id"].(string)

		for _, workspace := range mappingData["workspace"].(*schema.Set).List() {
			workspaceData := workspace.(map[string]interface{})
			workspaceID := workspaceData["id"].(string)
			key := fmt.Sprintf("%s::%s", projectID, workspaceID)
			seenNew[key] = struct{}{}

			roles := make([]string, 0)
			if v, ok := workspaceData["roles"]; ok && v != nil {
				for _, role := range v.(*schema.Set).List() {
					roles = append(roles, role.(string))
				}
			}

			workspaceMap[key] = &models.V1WorkspaceRolesPatch{
				UID:   workspaceID,
				Roles: roles,
			}
		}
	}

	// Step 3: Finalize list
	workspaces := make([]*models.V1WorkspaceRolesPatch, 0, len(workspaceMap))
	for _, w := range workspaceMap {
		workspaces = append(workspaces, w)
	}

	return &models.V1WorkspacesRolesPatch{
		Workspaces: workspaces,
	}
}

func resourceProjectRoleMappingHash(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})

	// Sort the roles to ensure order does not affect the hash
	roles := make([]string, len(m["roles"].(*schema.Set).List()))
	for i, role := range m["roles"].(*schema.Set).List() {
		roles[i] = role.(string)
	}
	sort.Strings(roles)

	buf.WriteString(fmt.Sprintf("%s-", m["id"].(string)))

	for _, role := range roles {
		buf.WriteString(fmt.Sprintf("%s-", role))
	}

	return int(hash(buf.String()))
}

func resourceWorkspaceRoleMappingHash(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})

	// Hash project id
	if v, ok := m["id"].(string); ok {
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
			hashes[i] = resourceProjectRoleMappingHash(workspace)
		}
		sort.Ints(hashes)

		for _, h := range hashes {
			buf.WriteString(fmt.Sprintf("%d-", h))
		}
	}

	return int(hash(buf.String()))
}
