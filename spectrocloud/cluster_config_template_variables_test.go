package spectrocloud

import (
	"encoding/json"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/spectrocloud/terraform-provider-spectrocloud/tests/mockApiServer/routes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func tplVar(name, value, strategy string, clusterIDs ...string) templateVariable {
	return templateVariable{Name: name, Value: value, AssignStrategy: strategy, ClusterIDs: clusterIDs}
}

func tplProfiles(profiles map[string][]templateVariable) templateProfiles {
	out := templateProfiles{}
	for id, vars := range profiles {
		out[id] = map[string]templateVariable{}
		for _, v := range vars {
			out[id][v.Name] = v
		}
	}
	return out
}

func TestPlanClusterTemplateProfileUpdate(t *testing.T) {
	base := tplProfiles(map[string][]templateVariable{
		"p1": {
			tplVar("type", "gp3", "cluster", "c1"),
			tplVar("policy", "Create", "cluster", "c1"),
			tplVar("region", "us-east-1", "all"),
		},
	})

	cases := []struct {
		name        string
		newProfiles templateProfiles
		wantPut     bool
		wantPatches []clusterTemplateVariablePatch
	}{
		{
			name:        "no change",
			newProfiles: base,
		},
		{
			name: "cluster value change patches only",
			newProfiles: tplProfiles(map[string][]templateVariable{"p1": {
				tplVar("type", "gp4", "cluster", "c1"),
				tplVar("policy", "Create", "cluster", "c1"),
				tplVar("region", "us-east-1", "all"),
			}}),
			wantPatches: []clusterTemplateVariablePatch{{ProfileUID: "p1", Name: "type", Value: "gp4", ClusterIDs: []string{"c1"}}},
		},
		{
			name: "two value changes are two separate patches",
			newProfiles: tplProfiles(map[string][]templateVariable{"p1": {
				tplVar("type", "gp4", "cluster", "c1"),
				tplVar("policy", "Delete", "cluster", "c1"),
				tplVar("region", "us-east-1", "all"),
			}}),
			wantPatches: []clusterTemplateVariablePatch{
				{ProfileUID: "p1", Name: "policy", Value: "Delete", ClusterIDs: []string{"c1"}},
				{ProfileUID: "p1", Name: "type", Value: "gp4", ClusterIDs: []string{"c1"}},
			},
		},
		{
			name: "all strategy value change needs put",
			newProfiles: tplProfiles(map[string][]templateVariable{"p1": {
				tplVar("type", "gp3", "cluster", "c1"),
				tplVar("policy", "Create", "cluster", "c1"),
				tplVar("region", "us-west-2", "all"),
			}}),
			wantPut: true,
		},
		{
			name: "cluster_ids change needs put and patch",
			newProfiles: tplProfiles(map[string][]templateVariable{"p1": {
				tplVar("type", "gp3", "cluster", "c1", "c2"),
				tplVar("policy", "Create", "cluster", "c1"),
				tplVar("region", "us-east-1", "all"),
			}}),
			wantPut:     true,
			wantPatches: []clusterTemplateVariablePatch{{ProfileUID: "p1", Name: "type", Value: "gp3", ClusterIDs: []string{"c1", "c2"}}},
		},
		{
			name: "added cluster variable needs put and patch",
			newProfiles: tplProfiles(map[string][]templateVariable{"p1": {
				tplVar("type", "gp3", "cluster", "c1"),
				tplVar("policy", "Create", "cluster", "c1"),
				tplVar("region", "us-east-1", "all"),
				tplVar("csi", "ext4", "cluster", "c1"),
			}}),
			wantPut:     true,
			wantPatches: []clusterTemplateVariablePatch{{ProfileUID: "p1", Name: "csi", Value: "ext4", ClusterIDs: []string{"c1"}}},
		},
		{
			name: "removed variable needs put",
			newProfiles: tplProfiles(map[string][]templateVariable{"p1": {
				tplVar("type", "gp3", "cluster", "c1"),
				tplVar("region", "us-east-1", "all"),
			}}),
			wantPut: true,
		},
		{
			name: "new profile is carried by put alone",
			newProfiles: tplProfiles(map[string][]templateVariable{"p2": {
				tplVar("type", "gp3", "cluster", "c1"),
			}}),
			wantPut: true,
		},
		{
			name: "cluster variable without cluster_ids is not patched",
			newProfiles: tplProfiles(map[string][]templateVariable{"p1": {
				tplVar("type", "gp9", "cluster"),
				tplVar("policy", "Create", "cluster", "c1"),
				tplVar("region", "us-east-1", "all"),
			}}),
			wantPut: true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			put, patches := planClusterTemplateProfileUpdate(base, tc.newProfiles)
			assert.Equal(t, tc.wantPut, put)
			assert.Equal(t, tc.wantPatches, patches)
		})
	}
}

func TestMergeUnmanagedTemplateVariables(t *testing.T) {
	body := []*models.V1ClusterTemplateProfile{
		{UID: "p1", Variables: []*models.V1ClusterTemplateVariable{{Name: "type", Value: "gp4", AssignStrategy: "cluster"}}},
		{UID: "p-new"},
	}
	unmanaged := &models.V1ClusterTemplateVariable{Name: "csi", AssignStrategy: "cluster", Clusters: []*models.V1ClusterVariableValue{{UID: "c2"}}}
	server := []*models.V1ClusterTemplateProfile{
		{UID: "p1", Variables: []*models.V1ClusterTemplateVariable{
			{Name: "type", AssignStrategy: "cluster"},
			{Name: "removed", AssignStrategy: "cluster"},
			unmanaged,
		}},
	}
	// "removed" was managed in the previous state and dropped from config, so
	// it must stay removed; "csi" was never managed, so it is preserved.
	old := tplProfiles(map[string][]templateVariable{"p1": {
		tplVar("type", "gp3", "cluster", "c1"),
		tplVar("removed", "x", "cluster", "c1"),
	}})

	mergeUnmanagedTemplateVariables(body, server, old)

	require.Len(t, body[0].Variables, 2)
	assert.Equal(t, "type", body[0].Variables[0].Name)
	assert.Equal(t, "gp4", body[0].Variables[0].Value)
	assert.Same(t, unmanaged, body[0].Variables[1])
	assert.Empty(t, body[1].Variables)
}

func TestClusterStrategyVariableReadValue(t *testing.T) {
	cases := []struct {
		name     string
		prior    string
		docValue string
		ids      []string
		assigned map[string]string
		want     string
	}{
		{"all clusters match prior", "gp3", "", []string{"c1", "c2"}, map[string]string{"c1": "gp3", "c2": "gp3"}, "gp3"},
		{"one cluster differs shows drift", "gp3", "", []string{"c1", "c2"}, map[string]string{"c1": "gp3", "c2": "gp2"}, "gp2"},
		{"pending cluster shows empty", "gp4", "", []string{"c1"}, map[string]string{"c1": ""}, ""},
		{"unlisted clusters are ignored", "gp3", "", []string{"c1"}, map[string]string{"c1": "gp3", "c9": "other"}, "gp3"},
		{"no assignment info keeps prior", "gp3", "doc", []string{"c1"}, nil, "gp3"},
		{"import without prior reads assignment", "", "", []string{"c1"}, map[string]string{"c1": "ext4"}, "ext4"},
		{"import without any info reads document", "", "doc", []string{"c1"}, nil, "doc"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, clusterStrategyVariableReadValue(tc.prior, tc.docValue, tc.ids, tc.assigned))
		})
	}
}

func recorderTemplate(t *testing.T) []*models.V1ClusterTemplateProfile {
	t.Helper()
	c := castV1Client(t, unitTestMockAPIClient)
	tmpl, err := c.GetClusterConfigTemplate(routes.ClusterConfigTemplateRecorderUID)
	require.NoError(t, err)
	return tmpl.Spec.Profiles
}

func stateVariables(t *testing.T, set *schema.Set) map[string]string {
	t.Helper()
	require.Equal(t, 1, set.Len())
	out := map[string]string{}
	vars := set.List()[0].(map[string]interface{})["variables"].(*schema.Set)
	for _, v := range vars.List() {
		vm := v.(map[string]interface{})
		out[vm["name"].(string)] = vm["value"].(string)
	}
	return out
}

// Read keeps only the variables in state and reads "cluster"-strategy values
// from the per-cluster assignments: "type" is assigned gp3 (matches), while
// "policy" is still pending on its cluster, so it reads back empty and the
// plan shows the value to set. "csi" is not managed and is dropped.
func TestFlattenClusterTemplateProfilesForState_KeepsOnlyManagedVariables(t *testing.T) {
	c := castV1Client(t, unitTestMockAPIClient)
	prior := tplProfiles(map[string][]templateVariable{routes.ClusterConfigTemplateRecorderProfileUID: {
		tplVar("type", "gp3", "cluster", "cid-1"),
		tplVar("policy", "Create", "cluster", "cid-1"),
	}})

	set := flattenClusterTemplateProfilesForState(c, routes.ClusterConfigTemplateRecorderUID, recorderTemplate(t), prior)

	assert.Equal(t, map[string]string{"type": "gp3", "policy": ""}, stateVariables(t, set))
}

// On import (nothing in state) every variable on the template is read.
func TestFlattenClusterTemplateProfilesForState_ImportReadsAll(t *testing.T) {
	c := castV1Client(t, unitTestMockAPIClient)

	set := flattenClusterTemplateProfilesForState(c, routes.ClusterConfigTemplateRecorderUID, recorderTemplate(t), templateProfiles{})

	assert.Equal(t, map[string]string{"type": "gp3", "policy": "", "csi": "ext4"}, stateVariables(t, set))
}

// rawTemplateVariable is a cluster_profile.variables element in raw config form.
func rawTemplateVariable(name, value, strategy string, clusterIDs ...string) map[string]interface{} {
	ids := make([]interface{}, 0, len(clusterIDs))
	for _, id := range clusterIDs {
		ids = append(ids, id)
	}
	return map[string]interface{}{"name": name, "value": value, "assign_strategy": strategy, "cluster_ids": ids}
}

func recorderProfileRaw(vars ...map[string]interface{}) map[string]interface{} {
	vs := make([]interface{}, 0, len(vars))
	for _, v := range vars {
		vs = append(vs, v)
	}
	return map[string]interface{}{
		"name":       "recorder",
		"cloud_type": "eks",
		"cluster_profile": []interface{}{
			map[string]interface{}{"id": routes.ClusterConfigTemplateRecorderProfileUID, "variables": vs},
		},
	}
}

func buildRecorderUpdateResourceData(t *testing.T, oldRaw, newRaw map[string]interface{}) *schema.ResourceData {
	t.Helper()
	d := buildClusterConfigTemplateUpdateResourceData(t, oldRaw, newRaw)
	d.SetId(routes.ClusterConfigTemplateRecorderUID)
	require.True(t, d.HasChange("cluster_profile"))
	return d
}

// Changing only "cluster"-strategy values sends one PATCH per variable, the
// same call the UI makes, and no profiles PUT.
func TestUpdateClusterConfigTemplateProfiles_PatchPerVariable(t *testing.T) {
	c := castV1Client(t, unitTestMockAPIClient)
	routes.ResetClusterConfigTemplateCalls()

	d := buildRecorderUpdateResourceData(t,
		recorderProfileRaw(
			rawTemplateVariable("type", "gp3", "cluster", "cid-1"),
			rawTemplateVariable("policy", "Create", "cluster", "cid-1"),
		),
		recorderProfileRaw(
			rawTemplateVariable("type", "gp4", "cluster", "cid-1"),
			rawTemplateVariable("policy", "Delete", "cluster", "cid-1"),
		),
	)
	require.NoError(t, updateClusterConfigTemplateProfiles(c, d))

	calls := routes.ClusterConfigTemplateCalls()
	require.Len(t, calls, 2)
	got := map[string]string{}
	for _, call := range calls {
		assert.Equal(t, "PATCH", call.Method)
		var body models.V1ClusterTemplateProfilesVariablesBatchEntity
		require.NoError(t, json.Unmarshal(call.Body, &body))
		require.Len(t, body.Profiles, 1)
		assert.Equal(t, routes.ClusterConfigTemplateRecorderProfileUID, *body.Profiles[0].UID)
		require.Len(t, body.Profiles[0].Variables, 1)
		v := body.Profiles[0].Variables[0]
		require.Len(t, v.Clusters, 1)
		assert.Equal(t, "cid-1", v.Clusters[0].UID)
		got[*v.Name] = v.Clusters[0].Value
	}
	assert.Equal(t, map[string]string{"type": "gp4", "policy": "Delete"}, got)
}

// Adding a variable needs the profiles PUT; the PUT keeps the template's
// unmanaged "csi" variable, then the new variable's value is PATCHed.
func TestUpdateClusterConfigTemplateProfiles_PutKeepsUnmanagedVariables(t *testing.T) {
	c := castV1Client(t, unitTestMockAPIClient)
	routes.ResetClusterConfigTemplateCalls()

	d := buildRecorderUpdateResourceData(t,
		recorderProfileRaw(
			rawTemplateVariable("type", "gp3", "cluster", "cid-1"),
		),
		recorderProfileRaw(
			rawTemplateVariable("type", "gp3", "cluster", "cid-1"),
			rawTemplateVariable("policy", "Create", "cluster", "cid-1"),
		),
	)
	require.NoError(t, updateClusterConfigTemplateProfiles(c, d))

	calls := routes.ClusterConfigTemplateCalls()
	require.Len(t, calls, 2)
	assert.Equal(t, "PUT", calls[0].Method)
	assert.Equal(t, "PATCH", calls[1].Method)

	var put models.V1ClusterTemplateProfilesUpdateEntity
	require.NoError(t, json.Unmarshal(calls[0].Body, &put))
	require.Len(t, put.Profiles, 1)
	names := []string{}
	for _, v := range put.Profiles[0].Variables {
		names = append(names, v.Name)
	}
	assert.ElementsMatch(t, []string{"type", "policy", "csi"}, names)

	var patch models.V1ClusterTemplateProfilesVariablesBatchEntity
	require.NoError(t, json.Unmarshal(calls[1].Body, &patch))
	assert.Equal(t, "policy", *patch.Profiles[0].Variables[0].Name)
}

// No variable change sends nothing.
func TestUpdateClusterConfigTemplateProfiles_NoVariableChangeSendsNothing(t *testing.T) {
	c := castV1Client(t, unitTestMockAPIClient)
	routes.ResetClusterConfigTemplateCalls()

	raw := recorderProfileRaw(rawTemplateVariable("type", "gp3", "cluster", "cid-1"))
	d := buildClusterConfigTemplateUpdateResourceData(t, raw, raw)
	d.SetId(routes.ClusterConfigTemplateRecorderUID)
	require.NoError(t, updateClusterConfigTemplateProfiles(c, d))

	assert.Empty(t, routes.ClusterConfigTemplateCalls())
}
