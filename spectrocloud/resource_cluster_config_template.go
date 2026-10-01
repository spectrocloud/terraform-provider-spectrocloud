package spectrocloud

import (
	"context"
	"fmt"
	"log"
	"sort"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/spectrocloud/palette-sdk-go/client"
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

		CustomizeDiff: validateClusterConfigTemplateVariableAssignment,

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
							Type:        schema.TypeSet,
							Optional:    true,
							Description: "Set of profile variable values and assignment strategies.",
							Set:         resourceClusterConfigTemplateVariableHash,
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
									"cluster_ids": {
										Type:        schema.TypeSet,
										Optional:    true,
										Set:         schema.HashString,
										Elem:        &schema.Schema{Type: schema.TypeString},
										Description: "UIDs of the specific clusters `value` applies to. Required when `assign_strategy` is `cluster`; must be empty when `assign_strategy` is `all` (the value applies to every cluster attached to the template).",
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

		// Set attached clusters first - refreshProfileVariableValues (below)
		// needs attached_cluster already populated to resolve "all" strategy
		// variables' target clusters.
		if err := d.Set("attached_cluster", flattenAttachedClusters(template.Spec.Clusters)); err != nil {
			return diag.FromErr(err)
		}

		profiles := refreshProfileVariableValues(c, d, uid, flattenClusterTemplateProfiles(template.Spec.Profiles))
		if err := d.Set("cluster_profile", profiles); err != nil {
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

	// Handle profile updates (add/remove profiles or update variables)
	if d.HasChange("cluster_profile") {
		oldProfiles, newProfiles := d.GetChange("cluster_profile")

		// Check if profile set structure changed (IDs added/removed/changed)
		if profileStructureChanged(oldProfiles, newProfiles) {
			// Use PUT endpoint to update entire profiles list
			profiles := newProfiles.(*schema.Set).List()
			profilesEntity := &models.V1ClusterTemplateProfilesUpdateEntity{
				Profiles: expandClusterTemplateProfiles(profiles),
			}

			err := c.UpdateClusterConfigTemplateProfiles(d.Id(), profilesEntity)
			if err != nil {
				return diag.FromErr(err)
			}
		} else {
			// Only variables changed within existing profiles - use PATCH endpoint
			profiles := newProfiles.(*schema.Set).List()
			variablesEntity := buildProfilesVariablesBatchEntity(d, profiles)

			err := c.UpdateClusterConfigTemplateProfilesVariables(d.Id(), variablesEntity)
			if err != nil {
				return diag.FromErr(err)
			}
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

// validateClusterConfigTemplateVariableAssignment is a CustomizeDiff guard:
// assign_strategy "cluster" requires at least one cluster_ids entry to have
// any effect (buildProfilesVariablesBatchEntity only assigns the variable's
// value to clusters listed there); assign_strategy "all" must not set
// cluster_ids, since that value is sent to every attached cluster already
// and a populated cluster_ids there would be silently ignored.
func validateClusterConfigTemplateVariableAssignment(_ context.Context, diff *schema.ResourceDiff, _ interface{}) error {
	profilesRaw, ok := diff.Get("cluster_profile").(*schema.Set)
	if !ok {
		return nil
	}

	for _, profile := range profilesRaw.List() {
		p, ok := profile.(map[string]interface{})
		if !ok {
			continue
		}
		profileID, _ := p["id"].(string)

		variablesSet, ok := p["variables"].(*schema.Set)
		if !ok {
			continue
		}

		for _, v := range variablesSet.List() {
			varMap, ok := v.(map[string]interface{})
			if !ok {
				continue
			}
			varName, _ := varMap["name"].(string)
			assignStrategy, _ := varMap["assign_strategy"].(string)
			clusterIDs, _ := varMap["cluster_ids"].(*schema.Set)
			hasClusterIDs := clusterIDs != nil && clusterIDs.Len() > 0

			switch assignStrategy {
			case "cluster":
				if !hasClusterIDs {
					return fmt.Errorf("cluster_profile %q variable %q: cluster_ids is required when assign_strategy is \"cluster\"", profileID, varName)
				}
			case "all", "":
				if hasClusterIDs {
					return fmt.Errorf("cluster_profile %q variable %q: cluster_ids must be empty when assign_strategy is \"all\" (value is applied to every attached cluster)", profileID, varName)
				}
			}
		}
	}

	return nil
}

// Helper functions for expanding and flattening

// profileStructureChanged checks if the profile set structure changed (IDs added/removed/changed)
// Returns true if profiles were added, removed, or IDs changed
// Returns false if only variables within existing profiles changed
func profileStructureChanged(oldProfilesSet, newProfilesSet interface{}) bool {
	oldProfiles := oldProfilesSet.(*schema.Set).List()
	newProfiles := newProfilesSet.(*schema.Set).List()

	// Different number of profiles = structure changed
	if len(oldProfiles) != len(newProfiles) {
		return true
	}

	// Build a set of old profile IDs
	oldIDs := make(map[string]bool)
	for _, p := range oldProfiles {
		profile := p.(map[string]interface{})
		oldIDs[profile["id"].(string)] = true
	}

	// Check if all new IDs exist in old IDs
	for _, p := range newProfiles {
		profile := p.(map[string]interface{})
		id := profile["id"].(string)
		if !oldIDs[id] {
			// New ID found = structure changed
			return true
		}
	}

	// Same IDs = only variables changed
	return false
}

// variableTargeting holds assign_strategy/cluster_ids - the two fields that
// only ever exist as Terraform-side config intent. The backend has no
// queryable record of them independent of whichever write path last ran:
// GET /v1/clusterTemplates/{uid} only reflects whatever was declared at
// create time or a full-profile PUT, and a variables-only PATCH never
// updates that declaration. So these must come from prior state, not from
// the backend, or every Read after a PATCH-only update would silently
// revert the user's targeting back to whatever (if anything) it was
// before - typically empty.
type variableTargeting struct {
	assignStrategy string
	clusterIDs     *schema.Set
}

// priorVariableTargeting reads cluster_profile from the current ResourceData
// (prior state - must be called before d.Set("cluster_profile", ...)) and
// indexes each variable's targeting by (profile id, variable name).
func priorVariableTargeting(d *schema.ResourceData) map[[2]string]variableTargeting {
	meta := make(map[[2]string]variableTargeting)
	prior, ok := d.Get("cluster_profile").(*schema.Set)
	if !ok || prior == nil {
		return meta
	}
	for _, p := range prior.List() {
		pm, ok := p.(map[string]interface{})
		if !ok {
			continue
		}
		profileID, _ := pm["id"].(string)
		variablesSet, ok := pm["variables"].(*schema.Set)
		if !ok {
			continue
		}
		for _, v := range variablesSet.List() {
			vm, ok := v.(map[string]interface{})
			if !ok {
				continue
			}
			varName, _ := vm["name"].(string)
			assignStrategy, _ := vm["assign_strategy"].(string)
			clusterIDs, _ := vm["cluster_ids"].(*schema.Set)
			meta[[2]string{profileID, varName}] = variableTargeting{assignStrategy, clusterIDs}
		}
	}
	return meta
}

// refreshProfileVariableValues replaces each variable's "value" with the
// live per-cluster assigned value from GetClusterTemplateProfileVariables,
// and restores assign_strategy/cluster_ids from prior state (see
// variableTargeting). GET /v1/clusterTemplates/{uid} (which
// flattenClusterTemplateProfiles reads) only ever returns the template's own
// declared/create-time value - the Day-2 variables PATCH
// (buildProfilesVariablesBatchEntity) writes to a completely separate
// per-cluster assignment store that this GET never reads from. Without this
// refresh, Read would write the stale declared value (and wipe cluster_ids)
// straight back into state right after a successful variable update, making
// the change look like it never took effect (persistent drift on the next
// plan).
func refreshProfileVariableValues(c *client.V1Client, d *schema.ResourceData, templateID string, profiles *schema.Set) *schema.Set {
	priorMeta := priorVariableTargeting(d)
	updated := schema.NewSet(resourceClusterConfigTemplateProfileHash, []interface{}{})

	for _, profile := range profiles.List() {
		p := profile.(map[string]interface{})
		profileID, _ := p["id"].(string)

		variablesSet, ok := p["variables"].(*schema.Set)
		if !ok || variablesSet.Len() == 0 {
			updated.Add(p)
			continue
		}

		resp, err := c.GetClusterTemplateProfileVariables(templateID, profileID)
		if err != nil {
			log.Printf("Error fetching cluster_config_template profile variables for profile %s: %v", profileID, err)
			updated.Add(p)
			continue
		}

		// varName -> clusterUID -> assigned value
		assignments := make(map[string]map[string]string)
		for _, v := range resp.Variables {
			if v == nil || v.Variable == nil || v.Variable.Name == nil {
				continue
			}
			byCluster := make(map[string]string)
			for _, assignment := range v.Clusters {
				if assignment != nil && assignment.UID != nil && assignment.AssignedValue != "" {
					byCluster[*assignment.UID] = assignment.AssignedValue
				}
			}
			assignments[*v.Variable.Name] = byCluster
		}

		updatedVars := schema.NewSet(resourceClusterConfigTemplateVariableHash, []interface{}{})
		for _, v := range variablesSet.List() {
			varMap := v.(map[string]interface{})
			varName, _ := varMap["name"].(string)
			assignStrategy, _ := varMap["assign_strategy"].(string)
			clusterIDs, _ := varMap["cluster_ids"].(*schema.Set)

			// Prior state is authoritative for targeting (see variableTargeting);
			// the backend-declared assign_strategy is only a bootstrap fallback
			// for a variable Terraform has never seen before (e.g. right after
			// Create, or a full-profile PUT that added it).
			if meta, ok := priorMeta[[2]string{profileID, varName}]; ok {
				assignStrategy = meta.assignStrategy
				clusterIDs = meta.clusterIDs
			}
			if clusterIDs == nil {
				clusterIDs = schema.NewSet(schema.HashString, nil)
			}

			var targetUIDs []string
			if assignStrategy == "cluster" {
				for _, id := range clusterIDs.List() {
					targetUIDs = append(targetUIDs, id.(string))
				}
			} else {
				targetUIDs = allAttachedClusterUIDs(d)
			}

			newValue, _ := varMap["value"].(string)
			if byCluster, ok := assignments[varName]; ok {
				for _, uid := range targetUIDs {
					if val, ok := byCluster[uid]; ok {
						newValue = val
						break
					}
				}
			}

			updatedVars.Add(map[string]interface{}{
				"name":            varName,
				"value":           newValue,
				"assign_strategy": assignStrategy,
				"cluster_ids":     clusterIDs,
			})
		}

		updated.Add(map[string]interface{}{
			"id":        profileID,
			"variables": updatedVars,
		})
	}

	return updated
}

// allAttachedClusterUIDs reads attached_cluster (Computed, set on Read) and
// returns every attached cluster's UID - the target set for a variable whose
// assign_strategy is "all".
func allAttachedClusterUIDs(d *schema.ResourceData) []string {
	attached, ok := d.Get("attached_cluster").([]interface{})
	if !ok {
		return nil
	}
	uids := make([]string, 0, len(attached))
	for _, a := range attached {
		am, ok := a.(map[string]interface{})
		if !ok {
			continue
		}
		if uid, ok := am["cluster_uid"].(string); ok && uid != "" {
			uids = append(uids, uid)
		}
	}
	return uids
}

// buildProfilesVariablesBatchEntity builds the request body for profile variables patch operation
func buildProfilesVariablesBatchEntity(d *schema.ResourceData, profiles []interface{}) *models.V1ClusterTemplateProfilesVariablesBatchEntity {
	if len(profiles) == 0 {
		return &models.V1ClusterTemplateProfilesVariablesBatchEntity{
			Profiles: []*models.V1ClusterTemplateProfileVariablesGroup{},
		}
	}

	profileGroups := make([]*models.V1ClusterTemplateProfileVariablesGroup, 0)

	for _, profile := range profiles {
		p := profile.(map[string]interface{})
		profileUID := p["id"].(string)

		// Check if this profile has variables (now a set)
		variablesSet, hasVariables := p["variables"].(*schema.Set)
		if !hasVariables || variablesSet.Len() == 0 {
			continue
		}

		variables := variablesSet.List()

		// Build variable cluster mappings
		variableMappings := make([]*models.V1ClusterTemplateVariableClusterMapping, 0)
		for _, v := range variables {
			varMap := v.(map[string]interface{})
			varName := varMap["name"].(string)
			value, _ := varMap["value"].(string)
			assignStrategy, _ := varMap["assign_strategy"].(string)

			var targetUIDs []string
			if assignStrategy == "cluster" {
				if clusterIDs, ok := varMap["cluster_ids"].(*schema.Set); ok {
					for _, id := range clusterIDs.List() {
						targetUIDs = append(targetUIDs, id.(string))
					}
				}
			} else {
				targetUIDs = allAttachedClusterUIDs(d)
			}

			clusters := make([]*models.V1ClusterVariableValue, 0, len(targetUIDs))
			for _, uid := range targetUIDs {
				clusters = append(clusters, &models.V1ClusterVariableValue{
					UID:   uid,
					Value: value,
				})
			}

			mapping := &models.V1ClusterTemplateVariableClusterMapping{
				Name:     &varName,
				Clusters: clusters,
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
