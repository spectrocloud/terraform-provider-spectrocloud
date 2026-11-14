package spectrocloud

import (
	"log"
	"sort"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/spectrocloud/palette-sdk-go/client"
)

// Reconcile states of a cluster attached to a cluster template in which
// variable updates can be split between the template and the cluster.
const (
	clusterTemplateReconcileStateApplied = "Applied"
	clusterTemplateReconcileStatePending = "Pending"
)

// profileVariableValues maps profile UID -> variable name -> value.
type profileVariableValues map[string]map[string]string

func (p profileVariableValues) add(profileUID, name, value string) {
	if p[profileUID] == nil {
		p[profileUID] = map[string]string{}
	}
	p[profileUID][name] = value
}

// clusterTemplateVariablesFromData flattens cluster_template.cluster_profile
// variables into profile UID -> name -> value.
func clusterTemplateVariablesFromData(templateData []interface{}) (profileVariableValues, error) {
	out := profileVariableValues{}
	profiles, err := extractProfilesFromTemplateData(templateData)
	if err != nil {
		return nil, err
	}
	for _, profile := range profiles {
		p, ok := profile.(map[string]interface{})
		if !ok {
			continue
		}
		profileUID, _ := p["id"].(string)
		if profileUID == "" {
			continue
		}
		variables, ok := p["variables"].(map[string]interface{})
		if !ok {
			continue
		}
		for name, value := range variables {
			if name == "" || value == nil {
				continue
			}
			out.add(profileUID, name, value.(string))
		}
	}
	return out, nil
}

// changedClusterTemplateVariables returns the variables in newVars that were
// added or whose value differs from oldVars. Variables removed from config
// are not sent (there is no API to unset them).
func changedClusterTemplateVariables(oldVars, newVars profileVariableValues) profileVariableValues {
	changed := profileVariableValues{}
	for profileUID, vars := range newVars {
		for name, value := range vars {
			if oldValue, ok := oldVars[profileUID][name]; ok && oldValue == value {
				continue
			}
			changed.add(profileUID, name, value)
		}
	}
	return changed
}

// clusterTemplateReconcileVariableStates maps profile UID -> variable name ->
// true when the variable is resolved, false when it is unresolved, as
// reported by the cluster's template reconcile document.
func clusterTemplateReconcileVariableStates(reconcile *models.V1ClusterReconcile) map[string]map[string]bool {
	states := map[string]map[string]bool{}
	if reconcile == nil || reconcile.Spec == nil {
		return states
	}
	for _, profile := range reconcile.Spec.Profiles {
		if profile == nil || profile.UID == nil || profile.Variables == nil {
			continue
		}
		profileStates := map[string]bool{}
		for _, v := range profile.Variables.ResolvedVariables {
			if v != nil && v.Name != nil {
				profileStates[*v.Name] = true
			}
		}
		for _, v := range profile.Variables.UnResolvedVariables {
			if v != nil && v.Name != nil {
				profileStates[*v.Name] = false
			}
		}
		states[*profile.UID] = profileStates
	}
	return states
}

// clusterVariablesByProfile indexes the cluster's own variables document
// (GET /v1/spectroclusters/{uid}/variables) by profile UID and variable name.
func clusterVariablesByProfile(clusterVars []*models.V1SpectroClusterVariables) map[string]map[string]*models.V1SpectroClusterVariableResponse {
	out := map[string]map[string]*models.V1SpectroClusterVariableResponse{}
	for _, cv := range clusterVars {
		if cv == nil || cv.ProfileUID == nil {
			continue
		}
		for _, v := range cv.Variables {
			if v == nil || v.Name == nil {
				continue
			}
			if out[*cv.ProfileUID] == nil {
				out[*cv.ProfileUID] = map[string]*models.V1SpectroClusterVariableResponse{}
			}
			out[*cv.ProfileUID][*v.Name] = v
		}
	}
	return out
}

// routeClusterTemplateVariables decides, per changed variable, whether the
// update goes to the cluster template document (applied at the template's
// scheduled reconcile) or to the cluster's own document (applied now, like
// the cluster_profile flow).
//
// Only when the reconcile state is Applied or Pending is a variable sent to
// the cluster, and only if it is resolved in the reconcile document and the
// cluster already runs that profile with that variable. Everything else -
// unresolved variables, a profile version not yet on the cluster, variables
// the reconcile document does not list, any other reconcile state, or a
// failed reconcile GET (reconcile == nil) - goes to the template document.
func routeClusterTemplateVariables(changed profileVariableValues, reconcile *models.V1ClusterReconcile,
	clusterVars map[string]map[string]*models.V1SpectroClusterVariableResponse) (toTemplate, toCluster profileVariableValues) {
	toTemplate = profileVariableValues{}
	toCluster = profileVariableValues{}

	splittable := false
	if reconcile != nil && reconcile.Status != nil {
		switch reconcile.Status.State {
		case clusterTemplateReconcileStateApplied, clusterTemplateReconcileStatePending:
			splittable = true
		}
	}
	states := clusterTemplateReconcileVariableStates(reconcile)

	for profileUID, vars := range changed {
		for name, value := range vars {
			resolved, listed := states[profileUID][name]
			_, onCluster := clusterVars[profileUID][name]
			if splittable && listed && resolved && onCluster {
				toCluster.add(profileUID, name, value)
			} else {
				toTemplate.add(profileUID, name, value)
			}
		}
	}
	return toTemplate, toCluster
}

func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func toSpectroClusterVariables(vars map[string]string) []*models.V1SpectroClusterVariable {
	out := make([]*models.V1SpectroClusterVariable, 0, len(vars))
	for _, name := range sortedKeys(vars) {
		out = append(out, &models.V1SpectroClusterVariable{
			Name:  StringPtr(name),
			Value: vars[name],
		})
	}
	return out
}

func toClusterTemplateVariablesUpdateEntity(vars profileVariableValues) *models.V1ClusterTemplateVariablesUpdateEntity {
	profiles := make([]*models.V1ClusterTemplateProfileVariable, 0, len(vars))
	for _, profileUID := range sortedKeys(vars) {
		profiles = append(profiles, &models.V1ClusterTemplateProfileVariable{
			UID:       profileUID,
			Variables: toSpectroClusterVariables(vars[profileUID]),
		})
	}
	return &models.V1ClusterTemplateVariablesUpdateEntity{Profiles: profiles}
}

func toClusterVariablesUpdateEntity(vars profileVariableValues) []*models.V1SpectroClusterVariableUpdateEntity {
	out := make([]*models.V1SpectroClusterVariableUpdateEntity, 0, len(vars))
	for _, profileUID := range sortedKeys(vars) {
		out = append(out, &models.V1SpectroClusterVariableUpdateEntity{
			ProfileUID: StringPtr(profileUID),
			Variables:  toSpectroClusterVariables(vars[profileUID]),
		})
	}
	return out
}

func clusterTemplateIDFromData(templateData []interface{}) string {
	if len(templateData) == 0 {
		return ""
	}
	m, ok := templateData[0].(map[string]interface{})
	if !ok {
		return ""
	}
	id, _ := m["id"].(string)
	return id
}

// updateClusterTemplateVariables handles day-2 variable changes in
// cluster_template for a cluster attached to a template. It reads the
// cluster's template reconcile document once (no polling) and splits the
// changed variables: unresolved ones (and anything not yet on the cluster) go
// to the cluster template document, resolved ones go to the cluster's own
// variables, like the cluster_profile flow. See routeClusterTemplateVariables.
func updateClusterTemplateVariables(c *client.V1Client, d *schema.ResourceData) error {
	oldTemplateData, newTemplateData := d.GetChange("cluster_template")
	newTemplate, _ := newTemplateData.([]interface{})
	if len(newTemplate) == 0 {
		return nil
	}
	oldTemplate, _ := oldTemplateData.([]interface{})

	newVars, err := clusterTemplateVariablesFromData(newTemplate)
	if err != nil {
		return err
	}
	oldVars, err := clusterTemplateVariablesFromData(oldTemplate)
	if err != nil {
		return err
	}

	changed := changedClusterTemplateVariables(oldVars, newVars)
	if len(changed) == 0 {
		log.Printf("No variables to update for cluster_template")
		return nil
	}

	templateID := clusterTemplateIDFromData(newTemplate)
	var reconcile *models.V1ClusterReconcile
	var clusterVars map[string]map[string]*models.V1SpectroClusterVariableResponse
	if templateID != "" {
		reconcile, err = c.GetClusterTemplateClusterReconcile(templateID, d.Id())
		if err != nil {
			// Fall back to the template document for every variable.
			log.Printf("Unable to read cluster template reconcile for cluster %s, sending all variables to the template: %v", d.Id(), err)
			reconcile = nil
		}
	}
	if reconcile != nil {
		apiVars, err := c.GetClusterVariables(d.Id())
		if err != nil {
			log.Printf("Unable to read cluster variables for cluster %s, sending all variables to the template: %v", d.Id(), err)
		} else {
			clusterVars = clusterVariablesByProfile(apiVars)
		}
	}

	toTemplate, toCluster := routeClusterTemplateVariables(changed, reconcile, clusterVars)

	rollback := func(err error) error {
		_ = d.Set("cluster_template", oldTemplateData)
		return err
	}

	if len(toTemplate) > 0 {
		log.Printf("Patching %d profile(s) of variables on the cluster template document for cluster %s", len(toTemplate), d.Id())
		if err := c.UpdateClusterTemplateVariablesForCluster(d.Id(), toClusterTemplateVariablesUpdateEntity(toTemplate)); err != nil {
			return rollback(err)
		}
	}
	if len(toCluster) > 0 {
		log.Printf("Patching %d profile(s) of variables on the cluster document for cluster %s", len(toCluster), d.Id())
		if err := c.UpdateClusterProfileVariableInCluster(d.Id(), toClusterVariablesUpdateEntity(toCluster)); err != nil {
			return rollback(err)
		}
	}

	if err := flattenClusterTemplateVariables(c, d, d.Id()); err != nil {
		log.Printf("Warning: Failed to refresh variables after update: %v", err)
	}
	return nil
}

// clusterTemplateVariableReadBack picks the state value of one configured
// cluster_template variable: the cluster's own document when it holds the
// variable for that profile, otherwise this cluster's assignment on the
// cluster template document (the variable is not applied to the cluster yet).
func clusterTemplateVariableReadBack(prior string, clusterVar *models.V1SpectroClusterVariableResponse, templateValue string, templateHasValue bool) (string, bool) {
	if clusterVar != nil && clusterVar.Value != "" {
		// Sensitive values come back masked; keep the prior state/config value.
		if v := resolveProfileVariableValue(prior, clusterVar.Value, clusterVar.IsSensitive); v != "" {
			return v, true
		}
	}
	if templateHasValue {
		return templateValue, true
	}
	return "", false
}

// flattenClusterTemplateVariables refreshes the variables configured under
// cluster_template, keeping the profile IDs and variable names from config.
// See clusterTemplateVariableReadBack for which document each value is read from.
func flattenClusterTemplateVariables(c *client.V1Client, d *schema.ResourceData, clusterUID string) error {
	clusterTemplate, _ := d.Get("cluster_template").([]interface{})
	if len(clusterTemplate) == 0 {
		return nil
	}

	templateData := clusterTemplate[0].(map[string]interface{})
	templateID := templateData["id"].(string)

	updatedProfileSet := schema.NewSet(schema.HashResource(&schema.Resource{
		Schema: map[string]*schema.Schema{
			"id": {
				Type: schema.TypeString,
			},
			"variables": {
				Type: schema.TypeMap,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}), []interface{}{})

	profilesSet, _ := templateData["cluster_profile"].(*schema.Set)
	if profilesSet != nil && profilesSet.Len() > 0 {
		var clusterVars map[string]map[string]*models.V1SpectroClusterVariableResponse
		if apiVars, err := c.GetClusterVariables(clusterUID); err != nil {
			log.Printf("Error fetching cluster variables for cluster %s: %v", clusterUID, err)
		} else {
			clusterVars = clusterVariablesByProfile(apiVars)
		}

		for _, profile := range profilesSet.List() {
			p := profile.(map[string]interface{})
			profileID := p["id"].(string)
			configVars, _ := p["variables"].(map[string]interface{})

			updatedProfile := map[string]interface{}{"id": profileID}

			// This cluster's assignments on the cluster template document.
			templateValues := map[string]string{}
			if len(configVars) > 0 {
				if resp, err := c.GetClusterTemplateProfileVariables(templateID, profileID); err != nil {
					log.Printf("Error fetching cluster_template profile variables for profile %s: %v", profileID, err)
				} else if resp != nil {
					for _, v := range resp.Variables {
						if v == nil || v.Variable == nil || v.Variable.Name == nil {
							continue
						}
						for _, assignment := range v.Clusters {
							if assignment != nil && assignment.UID != nil && *assignment.UID == clusterUID && assignment.AssignedValue != "" {
								templateValues[*v.Variable.Name] = assignment.AssignedValue
							}
						}
					}
				}
			}

			variables := map[string]interface{}{}
			for name, configValue := range configVars {
				prior, _ := configValue.(string)
				templateValue, templateHasValue := templateValues[name]
				if value, ok := clusterTemplateVariableReadBack(prior, clusterVars[profileID][name], templateValue, templateHasValue); ok {
					variables[name] = value
				}
			}
			if len(variables) > 0 {
				updatedProfile["variables"] = variables
			}

			updatedProfileSet.Add(updatedProfile)
		}
	}

	return d.Set("cluster_template", []interface{}{
		map[string]interface{}{
			"id":              templateID,
			"cluster_profile": updatedProfileSet,
		},
	})
}
