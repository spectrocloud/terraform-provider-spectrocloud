package spectrocloud

import (
	"fmt"
	"log"
	"sort"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/spectrocloud/palette-sdk-go/client"
)

// Variable handling for spectrocloud_cluster_config_template.
//
// The template document can hold variables Terraform does not manage: Hubble
// adds every profile variable for a cluster when it is created from or
// attached to the template, and the UI assigns values per cluster. To keep
// that from showing as drift or being wiped:
//   - Read keeps only the variable names already in state for each profile
//     (everything on import, or when state holds none for that profile).
//   - "cluster"-strategy values are read from the per-cluster assignments
//     (assignedValue), because the template document does not carry them.
//   - Structural changes go through the profiles PUT, merged with the
//     unmanaged variables already on the template.
//   - Each changed "cluster"-strategy value goes through its own variables
//     PATCH, the same call the UI makes.

const (
	clusterTemplateAssignStrategyAll     = "all"
	clusterTemplateAssignStrategyCluster = "cluster"
)

// templateVariable is one cluster_profile.variables element in plain form.
type templateVariable struct {
	Name           string
	Value          string
	AssignStrategy string
	ClusterIDs     []string // sorted
}

// templateProfiles maps profile id -> variable name -> variable.
type templateProfiles map[string]map[string]templateVariable

// clusterTemplateVariablePatch is one per-variable PATCH request.
type clusterTemplateVariablePatch struct {
	ProfileUID string
	Name       string
	Value      string
	ClusterIDs []string
}

func templateVariableFromMap(v map[string]interface{}) templateVariable {
	tv := templateVariable{}
	tv.Name, _ = v["name"].(string)
	tv.Value, _ = v["value"].(string)
	tv.AssignStrategy, _ = v["assign_strategy"].(string)
	if tv.AssignStrategy == "" {
		tv.AssignStrategy = clusterTemplateAssignStrategyAll
	}
	if ids, ok := v["cluster_ids"].(*schema.Set); ok {
		for _, id := range ids.List() {
			tv.ClusterIDs = append(tv.ClusterIDs, id.(string))
		}
	}
	sort.Strings(tv.ClusterIDs)
	return tv
}

// templateProfilesFromSet converts a cluster_profile set into templateProfiles.
// Every profile id is present, even when it has no variables.
func templateProfilesFromSet(raw interface{}) templateProfiles {
	out := templateProfiles{}
	set, ok := raw.(*schema.Set)
	if !ok || set == nil {
		return out
	}
	for _, p := range set.List() {
		pm, ok := p.(map[string]interface{})
		if !ok {
			continue
		}
		id, _ := pm["id"].(string)
		vars := map[string]templateVariable{}
		if vs, ok := pm["variables"].(*schema.Set); ok {
			for _, v := range vs.List() {
				if vm, ok := v.(map[string]interface{}); ok {
					tv := templateVariableFromMap(vm)
					vars[tv.Name] = tv
				}
			}
		}
		out[id] = vars
	}
	return out
}

func sameStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// planClusterTemplateProfileUpdate decides how a cluster_profile change is
// written. A PUT is needed for any structural change: profiles added or
// removed, variables added or removed, or a change of assign_strategy,
// cluster_ids, or an "all"-strategy value. Every "cluster"-strategy variable
// that is new or whose value, cluster_ids or strategy changed, in a profile
// that already existed, gets its own PATCH. Variables of a newly added profile
// are carried by the PUT alone.
func planClusterTemplateProfileUpdate(oldProfiles, newProfiles templateProfiles) (needPut bool, patches []clusterTemplateVariablePatch) {
	if len(oldProfiles) != len(newProfiles) {
		needPut = true
	}

	profileIDs := make([]string, 0, len(newProfiles))
	for id := range newProfiles {
		profileIDs = append(profileIDs, id)
	}
	sort.Strings(profileIDs)

	for _, profileID := range profileIDs {
		newVars := newProfiles[profileID]
		oldVars, profileExisted := oldProfiles[profileID]
		if !profileExisted {
			needPut = true
			continue
		}
		if len(oldVars) != len(newVars) {
			needPut = true
		}

		for _, name := range sortedKeys(newVars) {
			nv := newVars[name]
			ov, varExisted := oldVars[name]
			if !varExisted ||
				ov.AssignStrategy != nv.AssignStrategy ||
				!sameStrings(ov.ClusterIDs, nv.ClusterIDs) ||
				(nv.AssignStrategy == clusterTemplateAssignStrategyAll && ov.Value != nv.Value) {
				needPut = true
			}

			if nv.AssignStrategy != clusterTemplateAssignStrategyCluster || len(nv.ClusterIDs) == 0 {
				continue
			}
			if varExisted && ov.AssignStrategy == nv.AssignStrategy &&
				ov.Value == nv.Value && sameStrings(ov.ClusterIDs, nv.ClusterIDs) {
				continue
			}
			patches = append(patches, clusterTemplateVariablePatch{
				ProfileUID: profileID,
				Name:       name,
				Value:      nv.Value,
				ClusterIDs: nv.ClusterIDs,
			})
		}
	}

	// A profile present before but absent now is also a structural change.
	for id := range oldProfiles {
		if _, ok := newProfiles[id]; !ok {
			needPut = true
		}
	}
	return needPut, patches
}

// mergeUnmanagedTemplateVariables appends to each profile in the PUT body the
// variables the template already holds that Terraform does not manage: not
// in the new configuration and not in the previous state. Without this the
// PUT, which replaces spec.profiles, would drop variables and per-cluster
// assignments set by Hubble or the UI.
func mergeUnmanagedTemplateVariables(body []*models.V1ClusterTemplateProfile, server []*models.V1ClusterTemplateProfile, oldProfiles templateProfiles) {
	serverByUID := make(map[string]*models.V1ClusterTemplateProfile, len(server))
	for _, sp := range server {
		if sp != nil {
			serverByUID[sp.UID] = sp
		}
	}
	for _, bp := range body {
		if bp == nil {
			continue
		}
		sp, ok := serverByUID[bp.UID]
		if !ok {
			continue
		}
		managed := map[string]bool{}
		for _, v := range bp.Variables {
			if v != nil {
				managed[v.Name] = true
			}
		}
		for name := range oldProfiles[bp.UID] {
			managed[name] = true
		}
		for _, sv := range sp.Variables {
			if sv == nil || managed[sv.Name] {
				continue
			}
			bp.Variables = append(bp.Variables, sv)
		}
	}
}

func toClusterTemplateVariablePatchEntity(p clusterTemplateVariablePatch) *models.V1ClusterTemplateProfilesVariablesBatchEntity {
	clusters := make([]*models.V1ClusterVariableValue, 0, len(p.ClusterIDs))
	for _, id := range p.ClusterIDs {
		clusters = append(clusters, &models.V1ClusterVariableValue{UID: id, Value: p.Value})
	}
	profileUID := p.ProfileUID
	name := p.Name
	return &models.V1ClusterTemplateProfilesVariablesBatchEntity{
		Profiles: []*models.V1ClusterTemplateProfileVariablesGroup{
			{
				UID: &profileUID,
				Variables: []*models.V1ClusterTemplateVariableClusterMapping{
					{Name: &name, Clusters: clusters},
				},
			},
		},
	}
}

// updateClusterConfigTemplateProfiles writes a cluster_profile change: one
// PUT for structural changes (merged with unmanaged server variables), then
// one PATCH per changed "cluster"-strategy variable.
func updateClusterConfigTemplateProfiles(c *client.V1Client, d *schema.ResourceData) error {
	oldRaw, newRaw := d.GetChange("cluster_profile")
	oldProfiles := templateProfilesFromSet(oldRaw)
	newProfiles := templateProfilesFromSet(newRaw)

	needPut, patches := planClusterTemplateProfileUpdate(oldProfiles, newProfiles)

	if needPut {
		body := expandClusterTemplateProfiles(newRaw.(*schema.Set).List())
		template, err := c.GetClusterConfigTemplate(d.Id())
		if err != nil {
			return fmt.Errorf("reading cluster config template %s before updating profiles: %w", d.Id(), err)
		}
		if template != nil && template.Spec != nil {
			mergeUnmanagedTemplateVariables(body, template.Spec.Profiles, oldProfiles)
		}
		if err := c.UpdateClusterConfigTemplateProfiles(d.Id(), &models.V1ClusterTemplateProfilesUpdateEntity{Profiles: body}); err != nil {
			return err
		}
	}

	for _, p := range patches {
		if err := c.UpdateClusterConfigTemplateProfilesVariables(d.Id(), toClusterTemplateVariablePatchEntity(p)); err != nil {
			return fmt.Errorf("updating variable %q of profile %s: %w", p.Name, p.ProfileUID, err)
		}
	}
	return nil
}

// clusterStrategyVariableReadValue picks the value to store for a
// "cluster"-strategy variable from its per-cluster assignments. Clusters in
// clusterIDs that have no assignment entry are ignored. If any assigned value
// differs from prior, that value is returned so the drift shows in the plan
// (a cluster still waiting for a value counts as ""). With no assignment
// information it falls back to prior, then to the template document's value.
func clusterStrategyVariableReadValue(prior, docValue string, clusterIDs []string, assigned map[string]string) string {
	seen := false
	for _, id := range clusterIDs {
		v, ok := assigned[id]
		if !ok {
			continue
		}
		seen = true
		if v != prior {
			return v
		}
	}
	if seen || prior != "" {
		return prior
	}
	return docValue
}

// clusterTemplateAssignedValues returns variable name -> cluster uid ->
// assignedValue for one profile of the template.
func clusterTemplateAssignedValues(c *client.V1Client, templateUID, profileUID string) (map[string]map[string]string, error) {
	resp, err := c.GetClusterTemplateProfileVariables(templateUID, profileUID)
	if err != nil {
		return nil, err
	}
	out := map[string]map[string]string{}
	if resp == nil {
		return out, nil
	}
	for _, v := range resp.Variables {
		if v == nil || v.Variable == nil || v.Variable.Name == nil {
			continue
		}
		byCluster := map[string]string{}
		for _, cl := range v.Clusters {
			if cl == nil || cl.UID == nil {
				continue
			}
			byCluster[*cl.UID] = cl.AssignedValue
		}
		out[*v.Variable.Name] = byCluster
	}
	return out, nil
}

// flattenClusterTemplateProfilesForState builds the cluster_profile set for
// the resource state. For each profile it keeps only the variable names in
// prior (all of them when prior holds none for that profile) and reads
// "cluster"-strategy values from the per-cluster assignments.
func flattenClusterTemplateProfilesForState(c *client.V1Client, templateUID string, profiles []*models.V1ClusterTemplateProfile, prior templateProfiles) *schema.Set {
	result := make([]interface{}, 0, len(profiles))
	for _, profile := range profiles {
		if profile == nil {
			continue
		}
		priorVars := prior[profile.UID]

		kept := make([]*models.V1ClusterTemplateVariable, 0, len(profile.Variables))
		for _, v := range profile.Variables {
			if v == nil {
				continue
			}
			if len(priorVars) > 0 {
				if _, managed := priorVars[v.Name]; !managed {
					continue
				}
			}
			kept = append(kept, v)
		}

		flat := flattenClusterTemplateProfileVariables(kept)

		var assigned map[string]map[string]string
		fetched := false
		vars := make([]interface{}, 0, flat.Len())
		for _, raw := range flat.List() {
			vm := raw.(map[string]interface{})
			tv := templateVariableFromMap(vm)
			if tv.AssignStrategy == clusterTemplateAssignStrategyCluster && len(tv.ClusterIDs) > 0 {
				if !fetched {
					fetched = true
					var err error
					assigned, err = clusterTemplateAssignedValues(c, templateUID, profile.UID)
					if err != nil {
						log.Printf("[WARN] cluster config template %s: reading variable assignments of profile %s: %v", templateUID, profile.UID, err)
					}
				}
				vm["value"] = clusterStrategyVariableReadValue(priorVars[tv.Name].Value, tv.Value, tv.ClusterIDs, assigned[tv.Name])
			}
			vars = append(vars, vm)
		}

		result = append(result, map[string]interface{}{
			"id":        profile.UID,
			"variables": schema.NewSet(resourceClusterConfigTemplateVariableHash, vars),
		})
	}
	return schema.NewSet(resourceClusterConfigTemplateProfileHash, result)
}
