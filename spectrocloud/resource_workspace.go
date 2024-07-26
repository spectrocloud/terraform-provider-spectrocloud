package spectrocloud

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-api-go/models"
	"github.com/spectrocloud/palette-sdk-go/client"

	"github.com/spectrocloud/terraform-provider-spectrocloud/spectrocloud/schemas"
)

func resourceWorkspace() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceWorkspaceCreate,
		ReadContext:   resourceWorkspaceRead,
		UpdateContext: resourceWorkspaceUpdate,
		DeleteContext: resourceWorkspaceDelete,

		SchemaVersion: 2,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"tags": {
				Type:     schema.TypeSet,
				Optional: true,
				Set:      schema.HashString,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"clusters": {
				Type:     schema.TypeSet,
				Required: true,
				Set:      resourceClusterHash,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"uid": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"backup_policy":        schemas.BackupPolicySchema(),
			"cluster_rbac_binding": schemas.ClusterRbacBindingSchema(),
			"namespaces":           schemas.ClusterNamespacesSchema(),
		},
	}
}

func resourceWorkspaceCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.V1Client)

	var diags diag.Diagnostics

	workspace := toWorkspace(d)

	uid, err := c.CreateWorkspace(workspace)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(uid)

	resourceWorkspaceRead(ctx, d, m)

	return diags
}

func resourceWorkspaceRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.V1Client)

	var diags diag.Diagnostics

	uid := d.Id()

	workspace, err := c.GetWorkspace(uid)
	if err != nil {
		return diag.FromErr(err)
	} else if workspace == nil {
		d.SetId("")
		return diags
	}

	fp := flattenWorkspaceClusters(workspace)
	if err := d.Set("clusters", fp); err != nil {
		return diag.FromErr(err)
	}

	backup, err := c.GetWorkspaceBackup(uid)
	if err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("backup_policy", flattenWorkspaceBackupPolicy(backup)); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("cluster_rbac_binding", flattenClusterRBAC(workspace.Spec.ClusterRbacs)); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("namespaces", flattenWorkspaceClusterNamespaces(workspace.Spec.ClusterNamespaces)); err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func resourceWorkspaceUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	//c := m.(*client.V1Client)
	//
	var diags diag.Diagnostics
	//
	//workspace, err := c.GetWorkspace(d.Id())
	//if err != nil {
	//	return diag.FromErr(err)
	//}
	//
	//if d.HasChange("clusters") {
	//	// resource allocation should go first because clusters are inside.
	//	namespaces := toUpdateWorkspaceNamespaces(d)
	//	if err := c.UpdateWorkspaceResourceAllocation(d.Id(), namespaces); err != nil {
	//		return diag.FromErr(err)
	//	}
	//	diagnostics, done := updateWorkspaceRBACs(d, c, workspace)
	//	if done {
	//		return diagnostics
	//	}
	//} else {
	//	if d.HasChange("cluster_rbac_binding") {
	//		diagnostics, done := updateWorkspaceRBACs(d, c, workspace)
	//		if done {
	//			return diagnostics
	//		}
	//	}
	//	if d.HasChange("namespaces") {
	//		if err := c.UpdateWorkspaceResourceAllocation(d.Id(), toUpdateWorkspaceNamespaces(d)); err != nil {
	//			return diag.FromErr(err)
	//		}
	//	}
	//}
	//
	//if d.HasChange("backup_policy") {
	//	if len(d.Get("backup_policy").([]interface{})) == 0 {
	//		return diag.FromErr(errors.New("not implemented"))
	//	}
	//	if err := updateWorkspaceBackupPolicy(c, d); err != nil {
	//		return diag.FromErr(err)
	//	}
	//}
	//
	//resourceWorkspaceRead(ctx, d, m)
	//
	return diags
}

func updateWorkspaceRBACs(d *schema.ResourceData, c *client.V1Client, workspace *models.V1Workspace) (diag.Diagnostics, bool) {
	rbacs := toWorkspaceRBACs(d)
	for id, rbac := range rbacs {
		if err := c.UpdateWorkspaceRBACS(d.Id(), workspace.Spec.ClusterRbacs[id].Metadata.UID, rbac); err != nil {
			return diag.FromErr(err), true
		}
	}
	return nil, false
}

func resourceWorkspaceDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.V1Client)
	var diags diag.Diagnostics
	err := c.DeleteWorkspace(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func toWorkspace(d *schema.ResourceData) *models.V1WorkspaceEntity {
	annotations := make(map[string]string)
	if len(d.Get("description").(string)) > 0 {
		annotations["description"] = d.Get("description").(string)
	}

	workspace := &models.V1WorkspaceEntity{
		Metadata: &models.V1ObjectMeta{
			Name:        d.Get("name").(string),
			UID:         d.Id(),
			Labels:      toTags(d),
			Annotations: annotations,
		},
		Spec: &models.V1WorkspaceSpec{
			ClusterNamespaces: toWorkspaceNamespaces(d),
			ClusterRbacs:      toWorkspaceRBACs(d),
			ClusterRefs:       toClusterRefs(d),
			Policies:          toWorkspacePolicies(d),
			Quota:             toQuota(d),
		},
	}

	return workspace
}
