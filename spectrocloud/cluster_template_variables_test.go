package spectrocloud

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/spectrocloud/terraform-provider-spectrocloud/tests/mockApiServer/routes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testReconcile(state string, resolved, unresolved map[string][]string) *models.V1ClusterReconcile {
	profiles := map[string]*models.V1ClusterTemplateProfileVariables{}
	get := func(uid string) *models.V1ClusterTemplateProfileVariables {
		if profiles[uid] == nil {
			u := uid
			profiles[uid] = &models.V1ClusterTemplateProfileVariables{
				UID:       &u,
				Variables: &models.V1ClusterTemplateProfileVariablesVariables{},
			}
		}
		return profiles[uid]
	}
	for uid, names := range resolved {
		for _, n := range names {
			get(uid).Variables.ResolvedVariables = append(get(uid).Variables.ResolvedVariables, &models.V1SpectroClusterVariable{Name: StringPtr(n)})
		}
	}
	for uid, names := range unresolved {
		for _, n := range names {
			get(uid).Variables.UnResolvedVariables = append(get(uid).Variables.UnResolvedVariables, &models.V1SpectroClusterVariable{Name: StringPtr(n)})
		}
	}
	r := &models.V1ClusterReconcile{Spec: &models.V1ClusterReconcileSpec{}, Status: &models.V1ClusterReconcileStatus{State: state}}
	for _, p := range profiles {
		r.Spec.Profiles = append(r.Spec.Profiles, p)
	}
	return r
}

func testClusterVars(profiles map[string]map[string]string) map[string]map[string]*models.V1SpectroClusterVariableResponse {
	var apiVars []*models.V1SpectroClusterVariables
	for uid, vars := range profiles {
		cv := &models.V1SpectroClusterVariables{ProfileUID: StringPtr(uid)}
		for name, value := range vars {
			cv.Variables = append(cv.Variables, &models.V1SpectroClusterVariableResponse{Name: StringPtr(name), Value: value})
		}
		apiVars = append(apiVars, cv)
	}
	return clusterVariablesByProfile(apiVars)
}

func TestRouteClusterTemplateVariables(t *testing.T) {
	changed := profileVariableValues{
		"p1": {"resolvedOnCluster": "a", "unresolved": "b", "notListed": "c"},
		"p2": {"resolvedNewVersion": "d"},
	}
	reconcileFor := func(state string) *models.V1ClusterReconcile {
		return testReconcile(state,
			map[string][]string{"p1": {"resolvedOnCluster"}, "p2": {"resolvedNewVersion"}},
			map[string][]string{"p1": {"unresolved"}},
		)
	}
	// Cluster still runs only p1 (p2 is the new profile version, not applied yet).
	clusterVars := testClusterVars(map[string]map[string]string{
		"p1": {"resolvedOnCluster": "old", "unresolved": "", "notListed": "x"},
	})

	allToTemplate := profileVariableValues{
		"p1": {"resolvedOnCluster": "a", "unresolved": "b", "notListed": "c"},
		"p2": {"resolvedNewVersion": "d"},
	}
	split := struct{ template, cluster profileVariableValues }{
		template: profileVariableValues{
			"p1": {"unresolved": "b", "notListed": "c"},
			"p2": {"resolvedNewVersion": "d"},
		},
		cluster: profileVariableValues{"p1": {"resolvedOnCluster": "a"}},
	}

	cases := []struct {
		name         string
		reconcile    *models.V1ClusterReconcile
		wantTemplate profileVariableValues
		wantCluster  profileVariableValues
	}{
		{"applied splits resolved to cluster", reconcileFor("Applied"), split.template, split.cluster},
		{"pending splits resolved to cluster", reconcileFor("Pending"), split.template, split.cluster},
		{"created goes to template", reconcileFor("Created"), allToTemplate, profileVariableValues{}},
		{"in progress goes to template", reconcileFor("InProgress"), allToTemplate, profileVariableValues{}},
		{"failed goes to template", reconcileFor("Failed"), allToTemplate, profileVariableValues{}},
		{"reconcile GET failed goes to template", nil, allToTemplate, profileVariableValues{}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			toTemplate, toCluster := routeClusterTemplateVariables(changed, tc.reconcile, clusterVars)
			assert.Equal(t, tc.wantTemplate, toTemplate)
			assert.Equal(t, tc.wantCluster, toCluster)
		})
	}
}

func TestRouteClusterTemplateVariables_NoClusterVariablesGoesToTemplate(t *testing.T) {
	changed := profileVariableValues{"p1": {"region": "us-west-2"}}
	reconcile := testReconcile("Applied", map[string][]string{"p1": {"region"}}, nil)

	toTemplate, toCluster := routeClusterTemplateVariables(changed, reconcile, nil)
	assert.Equal(t, changed, toTemplate)
	assert.Empty(t, toCluster)
}

func TestChangedClusterTemplateVariables(t *testing.T) {
	oldVars := profileVariableValues{"p1": {"same": "1", "edited": "old", "removed": "x"}}
	newVars := profileVariableValues{
		"p1": {"same": "1", "edited": "new", "added": "a"},
		"p2": {"newProfile": "b"},
	}
	assert.Equal(t, profileVariableValues{
		"p1": {"edited": "new", "added": "a"},
		"p2": {"newProfile": "b"},
	}, changedClusterTemplateVariables(oldVars, newVars))
}

func TestClusterTemplateVariableReadBack(t *testing.T) {
	onCluster := &models.V1SpectroClusterVariableResponse{Name: StringPtr("region"), Value: "cluster-value"}
	sensitive := &models.V1SpectroClusterVariableResponse{Name: StringPtr("pwd"), Value: "****", IsSensitive: true}
	emptyOnCluster := &models.V1SpectroClusterVariableResponse{Name: StringPtr("region")}

	cases := []struct {
		name          string
		prior         string
		clusterVar    *models.V1SpectroClusterVariableResponse
		templateValue string
		templateHas   bool
		want          string
		wantOK        bool
	}{
		{"cluster document wins", "cfg", onCluster, "template-value", true, "cluster-value", true},
		{"not on cluster reads template", "cfg", nil, "template-value", true, "template-value", true},
		{"empty on cluster reads template", "cfg", emptyOnCluster, "template-value", true, "template-value", true},
		{"sensitive on cluster keeps prior", "secret", sensitive, "", false, "secret", true},
		{"nowhere is dropped", "cfg", nil, "", false, "", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := clusterTemplateVariableReadBack(tc.prior, tc.clusterVar, tc.templateValue, tc.templateHas)
			assert.Equal(t, tc.wantOK, ok)
			assert.Equal(t, tc.want, got)
		})
	}
}

// End to end against the mock API: reconcile is Applied, "region" is resolved
// and on the cluster (cluster PATCH), "newVar" is unresolved on a profile
// version the cluster does not run yet (template PATCH). Both calls succeed
// and read-back takes "region" from the cluster document.
func TestUpdateClusterTemplateVariables_SplitsByReconcile(t *testing.T) {
	c := castV1Client(t, unitTestMockAPIClient)

	oldTemplate := []interface{}{
		map[string]interface{}{
			"id": "template-uid-1",
			"cluster_profile": []interface{}{
				map[string]interface{}{"id": "cluster-profile-import-1"},
				map[string]interface{}{"id": "cluster-profile-v2"},
			},
		},
	}
	newTemplate := []interface{}{
		map[string]interface{}{
			"id": "template-uid-1",
			"cluster_profile": []interface{}{
				map[string]interface{}{
					"id":        "cluster-profile-import-1",
					"variables": map[string]interface{}{"region": "us-west-2"},
				},
				map[string]interface{}{
					"id":        "cluster-profile-v2",
					"variables": map[string]interface{}{"newVar": "v"},
				},
			},
		},
	}

	d := buildClusterTemplateChangeResourceData(t, routes.ClusterTemplateReconcileAppliedClusterUID, oldTemplate, newTemplate)
	require.True(t, d.HasChange("cluster_template"))
	require.NoError(t, updateClusterTemplateVariables(c, d))

	tmpl := d.Get("cluster_template").([]interface{})
	require.Len(t, tmpl, 1)
	profSet := tmpl[0].(map[string]interface{})["cluster_profile"].(*schema.Set)
	got := map[string]interface{}{}
	for _, p := range profSet.List() {
		pm := p.(map[string]interface{})
		got[pm["id"].(string)] = pm["variables"]
	}
	// Mock cluster variables document holds region=us-east-1 for cluster-profile-import-1.
	assert.Equal(t, map[string]interface{}{"region": "us-east-1"}, got["cluster-profile-import-1"])
}

func TestUpdateClusterTemplateVariables_NoChangeSkipsPatch(t *testing.T) {
	c := castV1Client(t, unitTestMockAPIClient)
	template := []interface{}{
		map[string]interface{}{
			"id": "template-uid-1",
			"cluster_profile": []interface{}{
				map[string]interface{}{
					"id":        "cluster-profile-import-1",
					"variables": map[string]interface{}{"region": "us-west-2"},
				},
			},
		},
	}
	// The error UID would fail any template PATCH; unchanged variables must not send one.
	d := buildClusterTemplateChangeResourceData(t, clusterTemplateVariablesPatchErrorUID, template, template)
	require.NoError(t, updateClusterTemplateVariables(c, d))
}
