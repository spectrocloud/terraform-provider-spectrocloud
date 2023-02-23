package spectrocloud

import (
	"context"
	"time"

	"github.com/spectrocloud/hapi/models"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/spectrocloud/hapi/client"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceTeam() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceTeamCreate,
		ReadContext:   resourceTeamRead,
		UpdateContext: resourceTeamUpdate,
		DeleteContext: resourceTeamDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		SchemaVersion: 2,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"users": {
				Type:     schema.TypeSet,
				Optional: true,
				Set:      schema.HashString,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"project_role_mapping": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Required: true,
						},
						"roles": {
							Type:     schema.TypeSet,
							Required: true,
							Set:      schema.HashString,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
					},
				},
			},
		},
	}
}

func resourceTeamCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.V1Client)
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
	d.SetId(uid)

	return diags
}

func resourceTeamRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.V1Client)
	var diags diag.Diagnostics

	team, err := c.GetTeam(d.Id())
	if err != nil {
		return diag.FromErr(err)
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

	projectRoles, err := c.GetTeamProjectRoleAssociation(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	if projectRoles != nil && len(projectRoles.Projects) > 0 {
		mappings := make([]interface{}, 0)
		for _, data := range projectRoles.Projects {
			dataMap := make(map[string]interface{})
			roles := make([]string, 0)
			for _, role := range data.Roles {
				roles = append(roles, role.UID)
			}
			dataMap["id"] = data.UID
			dataMap["roles"] = roles
			mappings = append(mappings, dataMap)
		}
		if err := d.Set("project_role_mapping", mappings); err != nil {
			return diag.FromErr(err)
		}
	}

	return diags
}

func resourceTeamUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.V1Client)
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

	return diags
}

func resourceTeamDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.V1Client)
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
	mappings := d.Get("project_role_mapping").([]interface{})
	for _, mapping := range mappings {
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
