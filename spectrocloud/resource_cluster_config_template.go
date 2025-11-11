package spectrocloud

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/spectrocloud/palette-sdk-go/api/models"
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
			"profiles": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "List of cluster profile references.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"uid": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "UID of the cluster profile.",
						},
						"variables": {
							Type:        schema.TypeList,
							Optional:    true,
							Description: "List of profile variable values and assignment strategies.",
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
										Description: "Value of the variable to be applied to all clusters launched from this template. This value is used when assign_strategy is set to 'all'.",
									},
									"assign_strategy": {
										Type:         schema.TypeString,
										Optional:     true,
										Default:      "all",
										ValidateFunc: validation.StringInSlice([]string{"all", "cluster"}, false),
										Description:  "Assignment strategy for the variable. Allowed values are `all` or `cluster`. Default is `all`.",
									},
								},
							},
						},
					},
				},
			},
			"policies": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "List of policy references.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"uid": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "UID of the policy.",
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
			Profiles:  expandClusterTemplateProfiles(d.Get("profiles").([]interface{})),
			Policies:  expandClusterTemplatePolicies(d.Get("policies").([]interface{})),
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

		if err := d.Set("profiles", flattenClusterTemplateProfiles(template.Spec.Profiles)); err != nil {
			return diag.FromErr(err)
		}

		if err := d.Set("policies", flattenClusterTemplatePolicies(template.Spec.Policies)); err != nil {
			return diag.FromErr(err)
		}

		// Set attached clusters
		if err := d.Set("attached_cluster", flattenAttachedClusters(template.Spec.Clusters)); err != nil {
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

	// Handle profile updates (add/remove profiles or update variables)
	if d.HasChange("profiles") {
		oldProfiles, newProfiles := d.GetChange("profiles")

		// Check if profile list structure changed (UIDs added/removed/changed)
		if profileStructureChanged(oldProfiles.([]interface{}), newProfiles.([]interface{})) {
			// Use PUT endpoint to update entire profiles list
			profiles := newProfiles.([]interface{})
			profilesEntity := &models.V1ClusterTemplateProfilesUpdateEntity{
				Profiles: expandClusterTemplateProfiles(profiles),
			}

			err := c.UpdateClusterConfigTemplateProfiles(d.Id(), profilesEntity)
			if err != nil {
				return diag.FromErr(err)
			}
		} else {
			// Only variables changed within existing profiles - use PATCH endpoint
			profiles := newProfiles.([]interface{})
			variablesEntity := buildProfilesVariablesBatchEntity(profiles)

			err := c.UpdateClusterConfigTemplateProfilesVariables(d.Id(), variablesEntity)
			if err != nil {
				return diag.FromErr(err)
			}
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
	// The ID passed in is the UID
	d.SetId(d.Id())

	diags := resourceClusterConfigTemplateRead(ctx, d, m)
	if diags.HasError() {
		return nil, diags[0].Validate()
	}

	return []*schema.ResourceData{d}, nil
}

// Helper functions for expanding and flattening

// profileStructureChanged checks if the profile list structure changed (UIDs added/removed/changed)
// Returns true if profiles were added, removed, or UIDs changed
// Returns false if only variables within existing profiles changed
func profileStructureChanged(oldProfiles, newProfiles []interface{}) bool {
	// Different number of profiles = structure changed
	if len(oldProfiles) != len(newProfiles) {
		return true
	}

	// Build a set of old profile UIDs
	oldUIDs := make(map[string]bool)
	for _, p := range oldProfiles {
		profile := p.(map[string]interface{})
		oldUIDs[profile["uid"].(string)] = true
	}

	// Check if all new UIDs exist in old UIDs
	for _, p := range newProfiles {
		profile := p.(map[string]interface{})
		uid := profile["uid"].(string)
		if !oldUIDs[uid] {
			// New UID found = structure changed
			return true
		}
	}

	// Same UIDs in same order = only variables changed
	return false
}

// buildProfilesVariablesBatchEntity builds the request body for profile variables patch operation
func buildProfilesVariablesBatchEntity(profiles []interface{}) *models.V1ClusterTemplateProfilesVariablesBatchEntity {
	if len(profiles) == 0 {
		return &models.V1ClusterTemplateProfilesVariablesBatchEntity{
			Profiles: []*models.V1ClusterTemplateProfileVariablesGroup{},
		}
	}

	profileGroups := make([]*models.V1ClusterTemplateProfileVariablesGroup, 0)

	for _, profile := range profiles {
		p := profile.(map[string]interface{})
		profileUID := p["uid"].(string)

		// Check if this profile has variables
		variables, hasVariables := p["variables"].([]interface{})
		if !hasVariables || len(variables) == 0 {
			continue
		}

		// Build variable cluster mappings
		variableMappings := make([]*models.V1ClusterTemplateVariableClusterMapping, 0)
		for _, v := range variables {
			varMap := v.(map[string]interface{})
			varName := varMap["name"].(string)

			mapping := &models.V1ClusterTemplateVariableClusterMapping{
				Name:     &varName,
				Clusters: []*models.V1ClusterVariableValue{},
			}

			variableMappings = append(variableMappings, mapping)
		}

		profileGroup := &models.V1ClusterTemplateProfileVariablesGroup{
			UID:       &profileUID,
			Variables: variableMappings,
		}

		profileGroups = append(profileGroups, profileGroup)
	}

	return &models.V1ClusterTemplateProfilesVariablesBatchEntity{
		Profiles: profileGroups,
	}
}

func expandClusterTemplateProfiles(profiles []interface{}) []*models.V1ClusterTemplateProfile {
	if len(profiles) == 0 {
		return nil
	}

	result := make([]*models.V1ClusterTemplateProfile, len(profiles))
	for i, profile := range profiles {
		p := profile.(map[string]interface{})
		profileEntity := &models.V1ClusterTemplateProfile{
			UID: p["uid"].(string),
		}

		// Expand variables if present
		if variables, ok := p["variables"].([]interface{}); ok && len(variables) > 0 {
			profileEntity.Variables = expandClusterTemplateProfileVariables(variables)
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

		if value, ok := v["value"].(string); ok && value != "" {
			varEntity.Value = value
		}

		if assignStrategy, ok := v["assign_strategy"].(string); ok && assignStrategy != "" {
			varEntity.AssignStrategy = assignStrategy
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
			UID:  p["uid"].(string),
			Kind: p["kind"].(string),
		}
	}

	return result
}

func flattenClusterTemplateProfiles(profiles []*models.V1ClusterTemplateProfile) []interface{} {
	if profiles == nil {
		return []interface{}{}
	}

	result := make([]interface{}, len(profiles))
	for i, profile := range profiles {
		profileMap := map[string]interface{}{
			"uid": profile.UID,
		}

		// Flatten variables if present
		if len(profile.Variables) > 0 {
			profileMap["variables"] = flattenClusterTemplateProfileVariables(profile.Variables)
		}

		result[i] = profileMap
	}

	return result
}

func flattenClusterTemplateProfileVariables(variables []*models.V1ClusterTemplateVariable) []interface{} {
	if variables == nil {
		return []interface{}{}
	}

	result := make([]interface{}, len(variables))
	for i, variable := range variables {
		varMap := map[string]interface{}{
			"name": variable.Name,
		}

		if variable.Value != "" {
			varMap["value"] = variable.Value
		}

		if variable.AssignStrategy != "" {
			varMap["assign_strategy"] = variable.AssignStrategy
		}

		result[i] = varMap
	}

	return result
}

func flattenClusterTemplatePolicies(policies []*models.V1PolicyRef) []interface{} {
	if policies == nil {
		return []interface{}{}
	}

	result := make([]interface{}, len(policies))
	for i, policy := range policies {
		result[i] = map[string]interface{}{
			"uid":  policy.UID,
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
