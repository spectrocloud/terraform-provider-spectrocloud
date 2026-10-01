package spectrocloud

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------
// buildProfilesVariablesBatchEntity — assign_strategy "all" vs "cluster".
//
// Regression for a real bug: the builder used to send Clusters: []
// unconditionally for every variable, so the configured value never
// reached Palette's PATCH /v1/clusterTemplates/{uid}/profiles/variables
// regardless of assign_strategy. "all" must now expand to every cluster in
// attached_cluster; "cluster" must expand to exactly the configured
// cluster_ids, both carrying the variable's value.
// ---------------------------------------------------------------------

func variableWithStrategy(name, value, assignStrategy string, clusterIDs ...string) map[string]interface{} {
	ids := schema.NewSet(schema.HashString, nil)
	for _, id := range clusterIDs {
		ids.Add(id)
	}
	return map[string]interface{}{
		"name":            name,
		"value":           value,
		"assign_strategy": assignStrategy,
		"cluster_ids":     ids,
	}
}

func TestBuildProfilesVariablesBatchEntity_AllStrategyUsesAttachedClusters(t *testing.T) {
	d := resourceClusterConfigTemplate().TestResourceData()
	require.NoError(t, d.Set("attached_cluster", []interface{}{
		map[string]interface{}{"cluster_uid": "cluster-1", "name": "c1"},
		map[string]interface{}{"cluster_uid": "cluster-2", "name": "c2"},
	}))

	varsSet := schema.NewSet(resourceClusterConfigTemplateVariableHash, []interface{}{
		variableWithStrategy("region", "us-west-2", "all"),
	})

	got := buildProfilesVariablesBatchEntity(d, []interface{}{
		map[string]interface{}{"id": "profile-1", "variables": varsSet},
	})

	require.Len(t, got.Profiles, 1)
	require.Len(t, got.Profiles[0].Variables, 1)
	clusters := got.Profiles[0].Variables[0].Clusters
	require.Len(t, clusters, 2)
	seen := map[string]string{}
	for _, c := range clusters {
		seen[c.UID] = c.Value
	}
	assert.Equal(t, "us-west-2", seen["cluster-1"])
	assert.Equal(t, "us-west-2", seen["cluster-2"])
}

func TestBuildProfilesVariablesBatchEntity_ClusterStrategyUsesClusterIDs(t *testing.T) {
	d := resourceClusterConfigTemplate().TestResourceData()
	// attached_cluster deliberately includes a cluster NOT in cluster_ids,
	// to confirm "cluster" strategy ignores it.
	require.NoError(t, d.Set("attached_cluster", []interface{}{
		map[string]interface{}{"cluster_uid": "cluster-1", "name": "c1"},
		map[string]interface{}{"cluster_uid": "cluster-2", "name": "c2"},
	}))

	varsSet := schema.NewSet(resourceClusterConfigTemplateVariableHash, []interface{}{
		variableWithStrategy("region", "us-west-2", "cluster", "cluster-2"),
	})

	got := buildProfilesVariablesBatchEntity(d, []interface{}{
		map[string]interface{}{"id": "profile-1", "variables": varsSet},
	})

	require.Len(t, got.Profiles, 1)
	require.Len(t, got.Profiles[0].Variables, 1)
	clusters := got.Profiles[0].Variables[0].Clusters
	require.Len(t, clusters, 1)
	assert.Equal(t, "cluster-2", clusters[0].UID)
	assert.Equal(t, "us-west-2", clusters[0].Value)
}

// TestRefreshProfileVariableValues_OverridesStaleDeclaredValue is the
// regression for a real customer-reported bug: after a successful Day-2
// variable update via buildProfilesVariablesBatchEntity, the next Read would
// write the OLD value straight back into state, making the change look like
// it never took effect (persistent drift). Root cause: GET
// /v1/clusterTemplates/{uid} (what flattenClusterTemplateProfiles reads) only
// ever returns the template's own declared/create-time value - the PATCH
// writes to a separate per-cluster assignment store this GET never reads
// from. refreshProfileVariableValues must override that stale value with the
// live per-cluster assignment from GetClusterTemplateProfileVariables - the
// mock fixture (getClusterTemplateProfileVariablesResponse) returns
// "region" assigned "us-east-1" for cluster "test-cluster-id".
func TestRefreshProfileVariableValues_OverridesStaleDeclaredValue(t *testing.T) {
	c := castV1Client(t, unitTestMockAPIClient)
	d := resourceClusterConfigTemplate().TestResourceData()
	require.NoError(t, d.Set("attached_cluster", []interface{}{
		map[string]interface{}{"cluster_uid": "test-cluster-id", "name": "c1"},
	}))

	// "stale" declared value, as if it just came back from GetClusterConfigTemplate
	staleProfiles := schema.NewSet(resourceClusterConfigTemplateProfileHash, []interface{}{
		map[string]interface{}{
			"id": "profile-1",
			"variables": schema.NewSet(resourceClusterConfigTemplateVariableHash, []interface{}{
				variableWithStrategy("region", "stale-declared-value", "all"),
			}),
		},
	})

	refreshed := refreshProfileVariableValues(c, d, "template-uid-1", staleProfiles)

	require.Equal(t, 1, refreshed.Len())
	profile := refreshed.List()[0].(map[string]interface{})
	variables := profile["variables"].(*schema.Set)
	require.Equal(t, 1, variables.Len())
	variable := variables.List()[0].(map[string]interface{})
	assert.Equal(t, "us-east-1", variable["value"],
		"must reflect the live per-cluster assignment, not the stale declared value")
}

// TestRefreshProfileVariableValues_PreservesClusterIDsFromPriorState is the
// regression for the reported "cluster_ids shows null/empty after apply,
// drift persists" bug: GET /v1/clusterTemplates/{uid} (what
// flattenClusterTemplateProfiles builds "fresh" from) has no concept of
// cluster_ids at all - it's pure Terraform-side targeting intent, never
// persisted anywhere the backend can be asked about independent of the
// PATCH that used it. refreshProfileVariableValues must restore
// assign_strategy/cluster_ids from prior state (via priorVariableTargeting),
// not leave them at whatever the backend-sourced "fresh" value carried
// (nothing, for cluster_ids - or a stale assign_strategy, for a PATCH-only
// update that never touched the template's own declared Spec).
func TestRefreshProfileVariableValues_PreservesClusterIDsFromPriorState(t *testing.T) {
	c := castV1Client(t, unitTestMockAPIClient)
	d := resourceClusterConfigTemplate().TestResourceData()
	require.NoError(t, d.Set("attached_cluster", []interface{}{
		map[string]interface{}{"cluster_uid": "test-cluster-id", "name": "c1"},
	}))

	// Prior state: what the user actually configured and successfully applied.
	require.NoError(t, d.Set("cluster_profile", []interface{}{
		map[string]interface{}{
			"id": "profile-1",
			"variables": []interface{}{
				variableWithStrategy("region", "us-east-1", "cluster", "test-cluster-id"),
			},
		},
	}))

	// "fresh" from GET /v1/clusterTemplates/{uid}: no cluster_ids concept at
	// all (base flatten never sets that key), and assign_strategy may be
	// stale/whatever the template's own declared Spec says.
	freshFromBackend := schema.NewSet(resourceClusterConfigTemplateProfileHash, []interface{}{
		map[string]interface{}{
			"id": "profile-1",
			"variables": schema.NewSet(resourceClusterConfigTemplateVariableHash, []interface{}{
				map[string]interface{}{
					"name":            "region",
					"value":           "stale-declared-value",
					"assign_strategy": "all",
				},
			}),
		},
	})

	refreshed := refreshProfileVariableValues(c, d, "template-uid-1", freshFromBackend)

	require.Equal(t, 1, refreshed.Len())
	profile := refreshed.List()[0].(map[string]interface{})
	variables := profile["variables"].(*schema.Set)
	require.Equal(t, 1, variables.Len())
	variable := variables.List()[0].(map[string]interface{})

	assert.Equal(t, "cluster", variable["assign_strategy"],
		"assign_strategy must come from prior state, not the stale backend declaration")
	clusterIDs, ok := variable["cluster_ids"].(*schema.Set)
	require.True(t, ok)
	assert.Equal(t, []interface{}{"test-cluster-id"}, clusterIDs.List(),
		"cluster_ids must be preserved from prior state, not wiped to empty")
	assert.Equal(t, "us-east-1", variable["value"])
}

// ---------------------------------------------------------------------
// validateClusterConfigTemplateVariableAssignment — CustomizeDiff guard.
// ---------------------------------------------------------------------

func baseClusterConfigTemplateRaw(variables map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		"name":       "test-template",
		"cloud_type": "aws",
		"cluster_profile": []interface{}{
			map[string]interface{}{
				"id":        "profile-1",
				"variables": []interface{}{variables},
			},
		},
	}
}

func diffClusterConfigTemplate(t *testing.T, raw map[string]interface{}) error {
	t.Helper()
	res := resourceClusterConfigTemplate()
	config := terraform.NewResourceConfigRaw(raw)
	_, err := res.Diff(context.Background(), nil, config, nil)
	return err
}

// buildClusterConfigTemplateUpdateResourceData builds a real InstanceState +
// config diff via Resource.Diff, the same real-diff pattern used throughout
// this suite (e.g. buildEksUpdateResourceData), so HasChange behaves the way
// Terraform's own apply pipeline would produce it - a bare Set()-then-Set()
// on schema.TestResourceData never fires HasChange for nested Set fields.
func buildClusterConfigTemplateUpdateResourceData(t *testing.T, oldRaw, newRaw map[string]interface{}) *schema.ResourceData {
	t.Helper()
	res := resourceClusterConfigTemplate()

	oldRD := schema.TestResourceDataRaw(t, res.Schema, oldRaw)
	oldRD.SetId("test-template-id")
	oldState := oldRD.State()
	require.NotNil(t, oldState)

	newConfig := terraform.NewResourceConfigRaw(newRaw)

	diff, err := res.Diff(context.Background(), oldState, newConfig, nil)
	require.NoError(t, err)

	finalRD, err := schema.InternalMap(res.Schema).Data(oldState, diff)
	require.NoError(t, err)
	finalRD.SetId("test-template-id")
	return finalRD
}

// TestResourceClusterConfigTemplateProfileHash_DetectsAddedVariable is a
// regression test for a real customer-reported bug: adding a new variable
// to an EXISTING cluster_profile (same id) in spectrocloud_cluster_config_template
// was not detected by terraform plan/apply at all - d.HasChange("cluster_profile")
// stayed false, so resourceClusterConfigTemplateUpdate's entire variables-
// update branch never ran.
//
// Root cause: resourceClusterConfigTemplateProfileHash hashed only the
// profile's "id", excluding its nested "variables" set entirely. Since the
// outer cluster_profile set matches old/new elements by hash, and the hash
// never changed (same id), Terraform's nested Set-within-Set diffing never
// recursed into the differing "variables" content. Fixed by folding each
// variable's own hash into the profile hash.
func TestResourceClusterConfigTemplateProfileHash_DetectsAddedVariable(t *testing.T) {
	oldRaw := map[string]interface{}{
		"name":       "test-template",
		"cloud_type": "aws",
		"cluster_profile": []interface{}{
			map[string]interface{}{
				"id":        "profile-1",
				"variables": []interface{}{},
			},
		},
	}
	newRaw := map[string]interface{}{
		"name":       "test-template",
		"cloud_type": "aws",
		"cluster_profile": []interface{}{
			map[string]interface{}{
				"id": "profile-1",
				"variables": []interface{}{
					map[string]interface{}{
						"name":            "region",
						"value":           "us-west-2",
						"assign_strategy": "all",
					},
				},
			},
		},
	}

	d := buildClusterConfigTemplateUpdateResourceData(t, oldRaw, newRaw)
	assert.True(t, d.HasChange("cluster_profile"),
		"adding a variable to an existing cluster_profile must be detected as a change")

	oldProfiles, newProfiles := d.GetChange("cluster_profile")
	assert.False(t, profileStructureChanged(oldProfiles, newProfiles),
		"same profile id on both sides must still route to the variables-only PATCH path, not the PUT path")
}

func TestValidateClusterConfigTemplateVariableAssignment_RejectsMissingClusterIDs(t *testing.T) {
	err := diffClusterConfigTemplate(t, baseClusterConfigTemplateRaw(map[string]interface{}{
		"name":            "region",
		"value":           "us-west-2",
		"assign_strategy": "cluster",
	}))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cluster_ids is required")
}

func TestValidateClusterConfigTemplateVariableAssignment_RejectsClusterIDsWithAllStrategy(t *testing.T) {
	err := diffClusterConfigTemplate(t, baseClusterConfigTemplateRaw(map[string]interface{}{
		"name":            "region",
		"value":           "us-west-2",
		"assign_strategy": "all",
		"cluster_ids":     []interface{}{"cluster-1"},
	}))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "must be empty when assign_strategy is")
}

func TestValidateClusterConfigTemplateVariableAssignment_AcceptsValidCombinations(t *testing.T) {
	err := diffClusterConfigTemplate(t, baseClusterConfigTemplateRaw(map[string]interface{}{
		"name":            "region",
		"value":           "us-west-2",
		"assign_strategy": "cluster",
		"cluster_ids":     []interface{}{"cluster-1"},
	}))
	require.NoError(t, err)

	err = diffClusterConfigTemplate(t, baseClusterConfigTemplateRaw(map[string]interface{}{
		"name":            "region",
		"value":           "us-west-2",
		"assign_strategy": "all",
	}))
	require.NoError(t, err)
}
