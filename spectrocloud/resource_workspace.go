package spectrocloud

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-sdk-go/api/models"
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
		Importer: &schema.ResourceImporter{
			StateContext: resourceWorkspaceImport,
		},
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
			"workspace_quota": {
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				Description: "Workspace quota default limits assigned to the namespace.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"cpu": {
							Type:        schema.TypeInt,
							Optional:    true,
							Description: "CPU that the entire workspace is allowed to consume. The default value is 0, which imposes no limit.",
						},
						"memory": {
							Type:        schema.TypeInt,
							Optional:    true,
							Description: "Memory in Mib that the entire workspace is allowed to consume. The default value is 0, which imposes no limit.",
						},
						"gpu": {
							Type:        schema.TypeInt,
							Optional:    true,
							Description: "GPU that the entire workspace is allowed to consume. The default value is 0, which imposes no limit.",
						},
					},
				},
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
						"cluster_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"backup_policy":        schemas.BackupPolicySchema(),
			"cluster_rbac_binding": schemas.ClusterRbacBindingSchema(),
			"namespaces":           schemas.WorkspaceNamespacesSchema(),
		},
	}
}

func resourceWorkspaceCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := getV1ClientWithResourceContext(m, "")

	var diags diag.Diagnostics

	workspace := toWorkspace(d, c)

	uid, err := c.CreateWorkspace(workspace)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(uid)

	resourceWorkspaceRead(ctx, d, m)

	return diags
}

func resourceWorkspaceRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := getV1ClientWithResourceContext(m, "")

	var diags diag.Diagnostics

	uid := d.Id()

	workspace, err := c.GetWorkspace(uid)
	if err != nil {
		return handleReadError(d, err, diags)
	} else if workspace == nil {
		d.SetId("")
		return diags
	}

	wsQuota := flattenWorkspaceQuota(workspace)
	if err := d.Set("workspace_quota", wsQuota); err != nil {
		return diag.FromErr(err)
	}
	fp := flattenWorkspaceClusters(workspace, c)
	if err := d.Set("clusters", fp); err != nil {
		return diag.FromErr(err)
	}

	backup, err := c.GetWorkspaceBackup(uid)
	if err != nil && !strings.Contains(err.Error(), "Backup not configured") {
		return diag.FromErr(err)
	}
	if backup != nil {
		if err := d.Set("backup_policy", flattenWorkspaceBackupPolicy(backup, d)); err != nil {
			return diag.FromErr(err)
		}
	}

	if err := d.Set("cluster_rbac_binding", flattenClusterRBAC(workspace.Spec.ClusterRbacs)); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("namespaces", flattenWorkspaceClusterNamespaces(workspace.Spec.ClusterNamespaces)); err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func flattenWorkspaceQuota(workspace *models.V1Workspace) []interface{} {
	wsq := make([]interface{}, 0)
	if workspace.Spec.Quota.ResourceAllocation != nil {
		quota := map[string]interface{}{
			"cpu":    workspace.Spec.Quota.ResourceAllocation.CPUCores,
			"memory": workspace.Spec.Quota.ResourceAllocation.MemoryMiB,
		}

		// Handle GPU configuration if present
		if workspace.Spec.Quota.ResourceAllocation.GpuConfig != nil {
			quota["gpu"] = int(workspace.Spec.Quota.ResourceAllocation.GpuConfig.Limit)
		} else {
			quota["gpu"] = 0
		}

		wsq = append(wsq, quota)
	}
	return wsq
}

func resourceWorkspaceUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := getV1ClientWithResourceContext(m, "")

	var diags diag.Diagnostics

	workspace, err := c.GetWorkspace(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	if d.HasChange("description") || d.HasChange("tags") {
		annotations := make(map[string]string)
		if len(d.Get("description").(string)) > 0 {
			annotations["description"] = d.Get("description").(string)
		}
		meta := &models.V1ObjectMeta{
			Name:        d.Get("name").(string),
			UID:         d.Id(),
			Labels:      toTags(d),
			Annotations: annotations,
		}
		if err := c.UpdateWorkspaceMetadata(workspace.Metadata.UID, meta); err != nil {
			return diag.FromErr(err)
		}
	}

	if d.HasChange("clusters") || d.HasChange("workspace_quota") {
		// resource allocation should go first because clusters are inside.
		namespaces := toUpdateWorkspaceNamespaces(d, c)
		if err := c.UpdateWorkspaceResourceAllocation(d.Id(), namespaces); err != nil {
			return diag.FromErr(err)
		}
		diagnostics, done := updateWorkspaceRBACs(d, c, workspace)
		if done {
			return diagnostics
		}
	} else {
		if d.HasChange("cluster_rbac_binding") {
			diagnostics, done := updateWorkspaceRBACs(d, c, workspace)
			if done {
				return diagnostics
			}
		}
		if d.HasChange("namespaces") {
			if err := c.UpdateWorkspaceResourceAllocation(d.Id(), toUpdateWorkspaceNamespaces(d, c)); err != nil {
				return diag.FromErr(err)
			}
		}
	}

	if d.HasChange("backup_policy") {
		oldBackup, newBackup := d.GetChange("backup_policy")
		if len(d.Get("backup_policy").([]interface{})) == 0 {
			if len(newBackup.([]interface{})) == 0 {
				return diag.FromErr(errors.New("backup configuration cannot be removed, but the schedule can be disabled"))
			}
		} else if len(newBackup.([]interface{})) > len(oldBackup.([]interface{})) {
			if err := createWorkspaceBackupPolicy(c, d); err != nil {
				return diag.FromErr(err)
			}
		} else {
			if err := updateWorkspaceBackupPolicy(c, d); err != nil {
				return diag.FromErr(err)
			}
		}

	}

	resourceWorkspaceRead(ctx, d, m)

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
	c := getV1ClientWithResourceContext(m, "")
	var diags diag.Diagnostics
	err := c.DeleteWorkspace(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func toWorkspace(d *schema.ResourceData, c *client.V1Client) *models.V1WorkspaceEntity {
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
			ClusterRefs:       toClusterRefs(d, c),
			Policies:          toWorkspacePolicies(d),
			Quota:             toQuota(d),
		},
	}

	return workspace
}

func resourceWorkspaceImport(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	c := getV1ClientWithResourceContext(m, "")

	// The import ID should be the workspace UID
	workspaceUID := d.Id()

	// Validate that the workspace exists and we can access it
	workspace, err := c.GetWorkspace(workspaceUID)
	if err != nil {
		return nil, fmt.Errorf("could not retrieve workspace for import: %s", err)
	}
	if workspace == nil {
		return nil, fmt.Errorf("workspace with ID %s not found", workspaceUID)
	}

	// Set the workspace name from the retrieved workspace
	if err := d.Set("name", workspace.Metadata.Name); err != nil {
		return nil, err
	}

	// Read all workspace data to populate the state
	diags := resourceWorkspaceRead(ctx, d, m)
	if diags.HasError() {
		return nil, fmt.Errorf("could not read workspace for import: %v", diags)
	}

	return []*schema.ResourceData{d}, nil
}
