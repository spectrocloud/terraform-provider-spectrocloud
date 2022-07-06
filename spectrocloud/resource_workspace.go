package spectrocloud

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/hapi/models"
	"github.com/spectrocloud/terraform-provider-spectrocloud/pkg/client"
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
						"name": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"backup_policy": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"prefix": {
							Type:     schema.TypeString,
							Required: true,
						},
						"backup_location_id": {
							Type:     schema.TypeString,
							Required: true,
						},
						"schedule": {
							Type:     schema.TypeString,
							Required: true,
						},
						"expiry_in_hour": {
							Type:     schema.TypeInt,
							Required: true,
						},
						"include_disks": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  true,
						},
						"include_workspace_resources": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  true,
						},
						"namespaces": {
							Type:     schema.TypeSet,
							Optional: true,
							Set:      schema.HashString,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"cluster_uids": {
							Type:     schema.TypeSet,
							Optional: true,
							Set:      schema.HashString,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"include_all_clusters": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  true,
						},
						"include_cluster_resources": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  true,
						},
					},
				},
			},
			"cluster_rbac_binding": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:     schema.TypeString,
							Required: true,
						},
						"namespace": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"role": {
							Type:     schema.TypeMap,
							Optional: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"subjects": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"type": {
										Type:     schema.TypeString,
										Required: true,
									},
									"name": {
										Type:     schema.TypeString,
										Required: true,
									},
									"namespace": {
										Type:     schema.TypeString,
										Optional: true,
									},
								},
							},
						},
					},
				},
			},
			"namespaces": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"resource_allocation": {
							Type:     schema.TypeMap,
							Required: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"images_blacklist": {
							Type:     schema.TypeList,
							Optional: true,
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
	c := m.(*client.V1Client)

	var diags diag.Diagnostics

	workspace, err := c.GetWorkspace(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	if d.HasChange("clusters") {
		// resource allocation should go first because clusters are inside.
		namespaces := toUpdateWorkspaceNamespaces(d)
		if err := c.UpdateWorkspaceResourceAllocation(d.Id(), namespaces); err != nil {
			return diag.FromErr(err)
		}
		rbacs := toUpdateWorkspaceRBACs(d)
		if err := c.UpdateWorkspaceRBACS(d.Id(), workspace.Spec.ClusterRbacs[0].Metadata.UID, rbacs); err != nil {
			return diag.FromErr(err)
		}

	} else {
		rbacs := toUpdateWorkspaceRBACs(d)
		if d.HasChange("cluster_rbac_binding") {
			if err := c.UpdateWorkspaceRBACS(d.Id(), workspace.Spec.ClusterRbacs[0].Metadata.UID, rbacs); err != nil {
				return diag.FromErr(err)
			}
		}
		if d.HasChange("namespaces") {
			if err := c.UpdateWorkspaceResourceAllocation(d.Id(), toUpdateWorkspaceNamespaces(d)); err != nil {
				return diag.FromErr(err)
			}
		}
	}

	if d.HasChange("backup_policy") {
		if len(d.Get("backup_policy").([]interface{})) == 0 {
			if err := c.WorkspaceBackupDelete(); err != nil {
				return diag.FromErr(err)
			}
		}
		if err := updateWorkspaceBackupPolicy(c, d); err != nil {
			return diag.FromErr(err)
		}
	}

	resourceWorkspaceRead(ctx, d, m)

	return diags
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
