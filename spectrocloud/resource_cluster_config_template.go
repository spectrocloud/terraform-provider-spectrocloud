package spectrocloud

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/spectrocloud/palette-sdk-go/client/herr"
)

func resourceClusterConfigTemplate() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceClusterConfigTemplateCreate,
		ReadContext:   resourceClusterConfigTemplateRead,
		UpdateContext: resourceClusterConfigTemplateUpdate,
		DeleteContext: resourceClusterConfigTemplateDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceClusterConfigTemplateImport,
		},
		Description: "A resource for creating and managing cluster config templates.",

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		CustomizeDiff: keepStoredClusterTemplateVariables,

		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the cluster config template.",
			},
			"context": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "project",
				ValidateFunc: validation.StringInSlice([]string{"project", "tenant"}, false),
				Description: "The context of the cluster config template. Allowed values are `project` or `tenant`. " +
					"Default value is `project`. " + PROJECT_NAME_NUANCE,
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The description of the cluster config template.",
			},
			"tags": {
				Type:     schema.TypeSet,
				Optional: true,
				Set:      schema.HashString,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Description: "Assign tags to the cluster config template. Tags can be in the format `key:value` or just `key`.",
			},
			"cloud_type": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The cloud type for the cluster template. Examples: 'aws', 'azure', 'gcp', 'vsphere', etc.",
			},
			"cluster_profile": {
				Type:        schema.TypeSet,
				Optional:    true,
				Computed:    true,
				Description: "Set of cluster profile references.",
				Set:         resourceClusterConfigTemplateProfileHash,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "ID of the cluster profile.",
						},
						"variables": {
							Type:     schema.TypeSet,
							Optional: true,
							Computed: true,
							Description: "Set of profile variable values and assignment strategies stored on the template. " +
								"When omitted, the variables already stored on the template are kept. " +
								"Only the variables listed here are managed; other variables on the template (for example, ones added by Palette or the UI) are left untouched. " +
								"Values are applied to attached clusters on the next cluster profile version upgrade; " +
								"a variable whose definition is unchanged keeps the value already running on each cluster.",
							Set: resourceClusterConfigTemplateVariableHash,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"name": {
										Type:        schema.TypeString,
										Required:    true,
										Description: "Name of the variable.",
									},
									"value": {
										Type:        schema.TypeString,
										Optional:    true,
										Description: "Value of the variable. Applied to every attached cluster when `assign_strategy` is `all`, or to each cluster in `cluster_ids` when `assign_strategy` is `cluster`.",
									},
									"assign_strategy": {
										Type:         schema.TypeString,
										Optional:     true,
										Default:      "all",
										ValidateFunc: validation.StringInSlice([]string{"all", "cluster"}, false),
										Description:  "Assignment strategy for the variable. Allowed values are `all` or `cluster`. Default is `all`.",
									},
									"cluster_ids": {
										Type:        schema.TypeSet,
										Optional:    true,
										Set:         schema.HashString,
										Elem:        &schema.Schema{Type: schema.TypeString},
										Description: "UIDs of the specific clusters `value` applies to when `assign_strategy` is `cluster`. Ignored when `assign_strategy` is `all` (the value applies to every cluster attached to the template).",
									},
								},
							},
						},
					},
				},
			},
			"policy": {
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1, // Only one policy is supported for now
				Description: "List of policy references.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "ID of the policy.",
						},
						"kind": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Kind of the policy.",
						},
					},
				},
			},
			"attached_cluster": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "List of clusters attached to this template.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"cluster_uid": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "UID of the attached cluster.",
						},
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Name of the attached cluster.",
						},
					},
				},
			},
			"execution_state": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Current execution state of the cluster template. Possible values: `Pending`, `Applied`, `Failed`, `PartiallyApplied`.",
			},
			"upgrade_now": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.IsRFC3339Time,
				Description: "Timestamp to trigger an immediate upgrade for all clusters launched from this template. " +
					"NOTE: The upgrade executes immediately when this value changes - the timestamp does NOT schedule a future upgrade. " +
					"Set this to the current timestamp each time you want to trigger an upgrade. " +
					"This field can also be used for tracking when upgrades were triggered by the user. " +
					"Format: RFC3339 (e.g., '2024-01-15T10:30:00Z'). " +
					"Example: To trigger an upgrade now, set to current time like '2024-11-12T15:30:00Z'.",
			},
		},
	}
}

func resourceClusterConfigTemplateCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := getV1ClientWithResourceContext(m, d.Get("context").(string))

	metadata := &models.V1ObjectMetaInputEntity{
		Name:   d.Get("name").(string),
		Labels: toTags(d),
	}

	// Add description to annotations if provided
	if description, ok := d.GetOk("description"); ok {
		metadata.Annotations = map[string]string{
			"description": description.(string),
		}
	}

	template := &models.V1ClusterTemplateEntity{
		Metadata: metadata,
		Spec: &models.V1ClusterTemplateEntitySpec{
			CloudType: d.Get("cloud_type").(string),
			Profiles:  expandClusterTemplateProfiles(d.Get("cluster_profile").(*schema.Set).List()),
			Policies:  expandClusterTemplatePolicies(d.Get("policy").([]interface{})),
		},
	}

	uid, err := c.CreateClusterConfigTemplate(template)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(*uid.UID)
	return resourceClusterConfigTemplateRead(ctx, d, m)
}

func resourceClusterConfigTemplateRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := getV1ClientWithResourceContext(m, d.Get("context").(string))
	var diags diag.Diagnostics
	uid := d.Id()

	template, err := c.GetClusterConfigTemplate(uid)
	if err != nil {
		return handleReadError(d, err, diags)
	}

	if err := d.Set("name", template.Metadata.Name); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("tags", flattenTags(template.Metadata.Labels)); err != nil {
		return diag.FromErr(err)
	}

	// Get description from annotations if it exists
	if template.Metadata.Annotations != nil {
		if description, found := template.Metadata.Annotations["description"]; found {
			if err := d.Set("description", description); err != nil {
				return diag.FromErr(err)
			}
		}
	}

	if template.Spec != nil {
		if err := d.Set("cloud_type", template.Spec.CloudType); err != nil {
			return diag.FromErr(err)
		}

		if err := d.Set("attached_cluster", flattenAttachedClusters(template.Spec.Clusters)); err != nil {
			return diag.FromErr(err)
		}

		// Only the variables already managed in state are read back (all of
		// them on import); "cluster"-strategy values come from the
		// per-cluster assignments. See cluster_config_template_variables.go.
		prior := templateProfilesFromSet(d.Get("cluster_profile"))
		if err := d.Set("cluster_profile", flattenClusterTemplateProfilesForState(c, uid, template.Spec.Profiles, prior)); err != nil {
			return diag.FromErr(err)
		}

		if err := d.Set("policy", flattenClusterTemplatePolicies(template.Spec.Policies)); err != nil {
			return diag.FromErr(err)
		}
	}

	// Set execution state from status
	if template.Status != nil {
		if err := d.Set("execution_state", template.Status.State); err != nil {
			return diag.FromErr(err)
		}
	}

	return nil
}

func resourceClusterConfigTemplateUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := getV1ClientWithResourceContext(m, d.Get("context").(string))

	// Handle metadata updates (name, tags, description)
	if d.HasChanges("name", "tags", "description") {
		metadataEntity := &models.V1ObjectMetaInputEntity{
			Name:   d.Get("name").(string),
			Labels: toTags(d),
		}

		// Add description to annotations if provided
		if description, ok := d.GetOk("description"); ok {
			metadataEntity.Annotations = map[string]string{
				"description": description.(string),
			}
		}

		metadata := &models.V1ObjectMetaInputEntitySchema{
			Metadata: metadataEntity,
		}

		err := c.UpdateClusterConfigTemplate(d.Id(), metadata)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	// Handle policy updates
	if d.HasChange("policy") {
		policies := d.Get("policy").([]interface{})
		policiesEntity := &models.V1ClusterTemplatePoliciesUpdateEntity{
			Policies: expandClusterTemplatePolicies(policies),
		}

		err := c.UpdateClusterConfigTemplatePolicies(d.Id(), policiesEntity)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	// Structural changes go through the profiles PUT (merged with the
	// variables Terraform does not manage); each changed "cluster"-strategy
	// value goes through its own variables PATCH, as the UI does.
	if d.HasChange("cluster_profile") {
		if err := updateClusterConfigTemplateProfiles(c, d); err != nil {
			return diag.FromErr(err)
		}
	}

	// Handle upgrade trigger
	if d.HasChange("upgrade_now") {
		// Trigger immediate upgrade for all clusters launched from this template
		err := c.UpgradeClusterConfigTemplateClusters(d.Id())
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return resourceClusterConfigTemplateRead(ctx, d, m)
}

func resourceClusterConfigTemplateDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := getV1ClientWithResourceContext(m, d.Get("context").(string))

	err := c.DeleteClusterConfigTemplate(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")

	return nil
}

func resourceClusterConfigTemplateImport(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	// Import ID format: id_or_name:context (e.g. "test-tf-template:tenant" or "699d56ee625c44429c6c8b53:tenant")
	scope, templateID, err := ParseResourceID(d)
	if err != nil {
		return nil, err
	}
	c := getV1ClientWithResourceContext(m, scope)

	// Try by UID first
	template, err := c.GetClusterConfigTemplate(templateID)
	if err == nil && template != nil {
		d.SetId(template.Metadata.UID)
		if err := d.Set("context", scope); err != nil {
			return nil, err
		}
		diags := resourceClusterConfigTemplateRead(ctx, d, m)
		if diags.HasError() {
			return nil, diags[0].Validate()
		}
		return []*schema.ResourceData{d}, nil
	}
	if err != nil && !herr.IsNotFound(err) {
		return nil, fmt.Errorf("unable to retrieve cluster config template '%s': %w", templateID, err)
	}

	// Try by name
	templateSummary, nameErr := c.GetClusterConfigTemplateByName(templateID)
	if nameErr != nil {
		return nil, fmt.Errorf("unable to retrieve cluster config template by name or id '%s': %w", templateID, nameErr)
	}
	if templateSummary == nil || templateSummary.Metadata == nil {
		return nil, fmt.Errorf("cluster config template '%s' not found in context %s", templateID, scope)
	}
	templateUID := templateSummary.Metadata.UID
	if templateUID == "" {
		return nil, fmt.Errorf("cluster config template with name '%s' has no UID", templateID)
	}

	d.SetId(templateUID)
	if err := d.Set("context", scope); err != nil {
		return nil, err
	}
	diags := resourceClusterConfigTemplateRead(ctx, d, m)
	if diags.HasError() {
		return nil, diags[0].Validate()
	}
	return []*schema.ResourceData{d}, nil
}

// Hash functions for sets

func resourceClusterConfigTemplateProfileHash(v interface{}) int {
	var buf strings.Builder
	m := v.(map[string]interface{})

	if id, ok := m["id"].(string); ok {
		buf.WriteString(fmt.Sprintf("%s-", id))
	}
	if variablesSet, ok := m["variables"].(*schema.Set); ok && variablesSet.Len() > 0 {
		hashes := make([]int, 0, variablesSet.Len())
		for _, v := range variablesSet.List() {
			hashes = append(hashes, resourceClusterConfigTemplateVariableHash(v))
		}
		sort.Ints(hashes)
		for _, h := range hashes {
			buf.WriteString(fmt.Sprintf("%d-", h))
		}
	}

	return schema.HashString(buf.String())
}

func resourceClusterConfigTemplateVariableHash(v interface{}) int {
	var buf strings.Builder
	m := v.(map[string]interface{})

	if name, ok := m["name"].(string); ok {
		buf.WriteString(fmt.Sprintf("%s-", name))
	}
	if value, ok := m["value"].(string); ok {
		buf.WriteString(fmt.Sprintf("%s-", value))
	}
	if strategy, ok := m["assign_strategy"].(string); ok {
		buf.WriteString(fmt.Sprintf("%s-", strategy))
	}
	if clusterIDs, ok := m["cluster_ids"].(*schema.Set); ok && clusterIDs.Len() > 0 {
		ids := make([]string, 0, clusterIDs.Len())
		for _, id := range clusterIDs.List() {
			ids = append(ids, id.(string))
		}
		sort.Strings(ids)
		buf.WriteString(fmt.Sprintf("%s-", strings.Join(ids, ",")))
	}

	return schema.HashString(buf.String())
}

// keepStoredClusterTemplateVariables makes cluster_profile.variables behave as
// Optional+Computed. The profile set hash includes each profile's variables
// (so variable edits are diffed at all), which means a profile written without
// a variables block hashes differently from the stored one and would plan a
// replace that wipes the template's variables. For every configured profile
// with no variables, carry the stored variables of the same profile id
// forward into the plan.
func keepStoredClusterTemplateVariables(_ context.Context, diff *schema.ResourceDiff, _ interface{}) error {
	if diff.Id() == "" || !diff.HasChange("cluster_profile") {
		return nil
	}

	oldRaw, newRaw := diff.GetChange("cluster_profile")
	oldProfiles, ok := oldRaw.(*schema.Set)
	if !ok {
		return nil
	}
	newProfiles, ok := newRaw.(*schema.Set)
	if !ok {
		return nil
	}

	storedVariables := make(map[string]*schema.Set, oldProfiles.Len())
	for _, profile := range oldProfiles.List() {
		p := profile.(map[string]interface{})
		if variables, ok := p["variables"].(*schema.Set); ok && variables.Len() > 0 {
			storedVariables[p["id"].(string)] = variables
		}
	}

	merged := make([]interface{}, 0, newProfiles.Len())
	carried := false
	for _, profile := range newProfiles.List() {
		p := profile.(map[string]interface{})
		variables, _ := p["variables"].(*schema.Set)
		if stored, found := storedVariables[p["id"].(string)]; found && (variables == nil || variables.Len() == 0) {
			p = map[string]interface{}{
				"id":        p["id"],
				"variables": stored,
			}
			carried = true
		}
		merged = append(merged, p)
	}

	if !carried {
		return nil
	}
	return diff.SetNew("cluster_profile", merged)
}

// Helper functions for expanding and flattening

func expandClusterTemplateProfiles(profiles []interface{}) []*models.V1ClusterTemplateProfile {
	if len(profiles) == 0 {
		return nil
	}

	result := make([]*models.V1ClusterTemplateProfile, len(profiles))
	for i, profile := range profiles {
		p := profile.(map[string]interface{})
		profileEntity := &models.V1ClusterTemplateProfile{
			UID: p["id"].(string),
		}

		// Expand variables if present (now a set)
		if variablesSet, ok := p["variables"].(*schema.Set); ok && variablesSet.Len() > 0 {
			profileEntity.Variables = expandClusterTemplateProfileVariables(variablesSet.List())
		}

		result[i] = profileEntity
	}

	return result
}

func expandClusterTemplateProfileVariables(variables []interface{}) []*models.V1ClusterTemplateVariable {
	if len(variables) == 0 {
		return nil
	}

	result := make([]*models.V1ClusterTemplateVariable, len(variables))
	for i, variable := range variables {
		v := variable.(map[string]interface{})
		varEntity := &models.V1ClusterTemplateVariable{
			Name: v["name"].(string),
		}

		value, _ := v["value"].(string)
		varEntity.Value = value

		if assignStrategy, ok := v["assign_strategy"].(string); ok && assignStrategy != "" {
			varEntity.AssignStrategy = assignStrategy
		}

		// For "cluster" strategy the template resolves each cluster's value
		// from its stored clusters list, so cluster_ids is persisted there.
		if varEntity.AssignStrategy == "cluster" {
			if clusterIDs, ok := v["cluster_ids"].(*schema.Set); ok {
				ids := make([]string, 0, clusterIDs.Len())
				for _, id := range clusterIDs.List() {
					ids = append(ids, id.(string))
				}
				sort.Strings(ids)
				for _, id := range ids {
					varEntity.Clusters = append(varEntity.Clusters, &models.V1ClusterVariableValue{
						UID:   id,
						Value: value,
					})
				}
			}
		}

		result[i] = varEntity
	}

	return result
}

func expandClusterTemplatePolicies(policies []interface{}) []*models.V1PolicyRef {
	if len(policies) == 0 {
		return nil
	}

	result := make([]*models.V1PolicyRef, len(policies))
	for i, policy := range policies {
		p := policy.(map[string]interface{})
		result[i] = &models.V1PolicyRef{
			UID:  p["id"].(string),
			Kind: p["kind"].(string),
		}
	}

	return result
}

func flattenClusterTemplateProfiles(profiles []*models.V1ClusterTemplateProfile) *schema.Set {
	if profiles == nil {
		return schema.NewSet(resourceClusterConfigTemplateProfileHash, []interface{}{})
	}

	result := make([]interface{}, len(profiles))
	for i, profile := range profiles {
		profileMap := map[string]interface{}{
			"id": profile.UID,
		}

		// Flatten variables if present (now returns a set)
		if len(profile.Variables) > 0 {
			profileMap["variables"] = flattenClusterTemplateProfileVariables(profile.Variables)
		} else {
			profileMap["variables"] = schema.NewSet(resourceClusterConfigTemplateVariableHash, []interface{}{})
		}

		result[i] = profileMap
	}

	return schema.NewSet(resourceClusterConfigTemplateProfileHash, result)
}

func flattenClusterTemplateProfileVariables(variables []*models.V1ClusterTemplateVariable) *schema.Set {
	if variables == nil {
		return schema.NewSet(resourceClusterConfigTemplateVariableHash, []interface{}{})
	}

	result := make([]interface{}, 0, len(variables))
	for _, variable := range variables {
		if variable == nil {
			continue
		}

		// The backend treats an empty assign strategy as "cluster".
		assignStrategy := variable.AssignStrategy
		if assignStrategy == "" {
			assignStrategy = "cluster"
		}

		value := variable.Value
		clusterIDs := make([]interface{}, 0, len(variable.Clusters))
		for _, cluster := range variable.Clusters {
			if cluster == nil || cluster.UID == "" {
				continue
			}
			clusterIDs = append(clusterIDs, cluster.UID)
			if value == "" {
				value = cluster.Value
			}
		}

		result = append(result, map[string]interface{}{
			"name":            variable.Name,
			"value":           value,
			"assign_strategy": assignStrategy,
			"cluster_ids":     schema.NewSet(schema.HashString, clusterIDs),
		})
	}

	return schema.NewSet(resourceClusterConfigTemplateVariableHash, result)
}

func flattenClusterTemplatePolicies(policies []*models.V1PolicyRef) []interface{} {
	if policies == nil {
		return []interface{}{}
	}

	result := make([]interface{}, len(policies))
	for i, policy := range policies {
		result[i] = map[string]interface{}{
			"id":   policy.UID,
			"kind": policy.Kind,
		}
	}

	return result
}

func flattenAttachedClusters(clusters map[string]models.V1ClusterTemplateSpcRef) []interface{} {
	if len(clusters) == 0 {
		return []interface{}{}
	}

	result := make([]interface{}, 0, len(clusters))
	for _, cluster := range clusters {
		result = append(result, map[string]interface{}{
			"cluster_uid": cluster.ClusterUID,
			"name":        cluster.Name,
		})
	}

	return result
}
