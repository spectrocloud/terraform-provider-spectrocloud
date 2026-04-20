package spectrocloud

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/spectrocloud/palette-sdk-go/client"
	"github.com/spectrocloud/palette-sdk-go/client/herr"

	"github.com/spectrocloud/terraform-provider-spectrocloud/spectrocloud/schemas"
)

func resourceWorkspace() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceWorkspaceCreate,
		ReadContext:   resourceWorkspaceRead,
		UpdateContext: resourceWorkspaceUpdate,
		DeleteContext: resourceWorkspaceDelete,

		SchemaVersion: 3,
		StateUpgraders: []schema.StateUpgrader{
			{
				Type:    resourceWorkspaceResourceV2().CoreConfigSchema().ImpliedType(),
				Upgrade: resourceWorkspaceStateUpgradeV2,
				Version: 2,
			},
		},
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

	workspace, err := toWorkspace(d, c)
	if err != nil {
		return diag.FromErr(err)
	}

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
		namespaces, err := toUpdateWorkspaceNamespaces(d, c)
		if err != nil {
			return diag.FromErr(err)
		}
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
			namespaces, err := toUpdateWorkspaceNamespaces(d, c)
			if err != nil {
				return diag.FromErr(err)
			}
			if err := c.UpdateWorkspaceResourceAllocation(d.Id(), namespaces); err != nil {
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

func toWorkspace(d *schema.ResourceData, c *client.V1Client) (*models.V1WorkspaceEntity, error) {
	annotations := make(map[string]string)
	if len(d.Get("description").(string)) > 0 {
		annotations["description"] = d.Get("description").(string)
	}

	quota, err := toQuota(d)
	if err != nil {
		return nil, err
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
			Quota:             quota,
		},
	}

	return workspace, nil
}

// isWorkspaceNotFound returns true if the error indicates the workspace was not found,
// so the importer can fall back to name lookup (same pattern as project/user import).
func isWorkspaceNotFound(err error) bool {
	if err == nil {
		return false
	}
	if herr.IsNotFound(err) {
		return true
	}
	s := err.Error()
	return strings.Contains(s, "not found") || strings.Contains(s, "NotFound") || strings.Contains(s, "WorkspaceNotFound")
}

func resourceWorkspaceImport(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	c := getV1ClientWithResourceContext(m, "")

	importID := d.Id()
	if importID == "" {
		return nil, fmt.Errorf("workspace import ID or name is required")
	}

	// Try to get by UID first (same pattern as resource_project_import)
	workspace, err := c.GetWorkspace(importID)
	if err != nil {
		if !isWorkspaceNotFound(err) {
			return nil, fmt.Errorf("unable to retrieve workspace '%s': %w", importID, err)
		}
		// Not found by UID — try by name
		wsByName, nameErr := c.GetWorkspaceByName(importID)
		if nameErr != nil {
			return nil, fmt.Errorf("unable to retrieve workspace by name or id '%s': %w", importID, nameErr)
		}
		if wsByName == nil || wsByName.Metadata == nil || wsByName.Metadata.UID == "" {
			return nil, fmt.Errorf("workspace '%s' not found", importID)
		}
		d.SetId(wsByName.Metadata.UID)
	} else if workspace != nil {
		d.SetId(importID)
	} else {
		// UID lookup returned nil workspace (e.g. import by name) — try by name
		wsByName, nameErr := c.GetWorkspaceByName(importID)
		if nameErr != nil {
			return nil, fmt.Errorf("unable to retrieve workspace by name or id '%s': %w", importID, nameErr)
		}
		if wsByName == nil || wsByName.Metadata == nil || wsByName.Metadata.UID == "" {
			return nil, fmt.Errorf("workspace '%s' not found", importID)
		}
		d.SetId(wsByName.Metadata.UID)
	}

	// Read all workspace data to populate the state
	diags := resourceWorkspaceRead(ctx, d, m)
	if diags.HasError() {
		return nil, fmt.Errorf("could not read workspace for import: %v", diags)
	}

	return []*schema.ResourceData{d}, nil
}

// resourceWorkspaceResourceV2 returns the schema for version 2 of the resource
// This represents the old schema where "namespaces" was TypeList
func resourceWorkspaceResourceV2() *schema.Resource {
	return &schema.Resource{
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
			"namespaces": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "The namespaces for the cluster.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Name of the namespace. This is the name of the Kubernetes namespace in the cluster.",
						},
						"resource_allocation": {
							Type:     schema.TypeMap,
							Required: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
							Description: "Resource allocation for the namespace. This is a map containing the resource type and the resource value. Only the following field names are supported for resource configuration: `cpu_cores`, `memory_MiB`, `gpu`, and `gpu_provider`. Any other field names will not be honored by the system. For example, `{cpu_cores: '2', memory_MiB: '2048', gpu: '1', gpu_provider: 'nvidia'}`",
						},
						"cluster_resource_allocations": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"uid": {
										Type:     schema.TypeString,
										Required: true,
									},
									"resource_allocation": {
										Type:     schema.TypeMap,
										Required: true,
										Elem: &schema.Schema{
											Type: schema.TypeString,
										},
										Description: "Resource allocation for the cluster. This is a map containing the resource type and the resource value. Only the following field names are supported for resource configuration: `cpu_cores`, `memory_MiB`, `gpu`. Any other field names will not be honored by the system. For example, `{cpu_cores: '2', memory_MiB: '2048', gpu: '1'}`. Note: gpu_provider is not supported here; use the default resource_allocation for GPU provider configuration.",
									},
								},
							},
						},
						"images_blacklist": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
							Description: "List of images to disallow for the namespace. For example, `['nginx:latest', 'redis:latest']`",
						},
					},
				},
			},
		},
	}
}

// resourceWorkspaceStateUpgradeV2 migrates state from version 2 to version 3
func resourceWorkspaceStateUpgradeV2(ctx context.Context, rawState map[string]interface{}, meta interface{}) (map[string]interface{}, error) {
	log.Printf("[DEBUG] Upgrading workspace state from version 2 to 3")

	// Convert namespaces from TypeList to TypeSet
	// Note: We keep the data as a list in rawState and let Terraform's schema processing
	// convert it to TypeSet during normal resource loading. This avoids JSON serialization
	// issues with schema.Set objects that contain hash functions.
	if namespacesRaw, exists := rawState["namespaces"]; exists {
		if namespacesList, ok := namespacesRaw.([]interface{}); ok {
			log.Printf("[DEBUG] Keeping namespaces as list during state upgrade with %d items", len(namespacesList))

			// Keep the namespaces data as-is (as a list)
			// Terraform will convert it to TypeSet when loading the resource using the schema
			rawState["namespaces"] = namespacesList

			log.Printf("[DEBUG] Successfully prepared namespaces for TypeSet conversion")
		} else {
			log.Printf("[DEBUG] namespaces is not a list, skipping conversion")
		}
	} else {
		log.Printf("[DEBUG] No namespaces found in state, skipping conversion")
	}

	return rawState, nil
}
