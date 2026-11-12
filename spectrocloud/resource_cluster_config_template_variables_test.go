package spectrocloud

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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

}

// Variables are Optional+Computed: a cluster_profile whose variables block is
// omitted must keep whatever the template already stores, with no diff.
func TestResourceClusterConfigTemplate_OmittedVariablesShowNoDiff(t *testing.T) {
	oldRaw := map[string]interface{}{
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
	newRaw := map[string]interface{}{
		"name":       "test-template",
		"cloud_type": "aws",
		"cluster_profile": []interface{}{
			map[string]interface{}{"id": "profile-1"},
		},
	}

	d := buildClusterConfigTemplateUpdateResourceData(t, oldRaw, newRaw)
	assert.False(t, d.HasChange("cluster_profile"),
		"omitting variables must keep the stored template variables without a diff")
}

func TestExpandClusterTemplateProfileVariables_ClusterStrategySendsClusterIDs(t *testing.T) {
	got := expandClusterTemplateProfileVariables([]interface{}{
		variableWithStrategy("region", "us-west-2", "cluster", "cluster-2", "cluster-1"),
	})

	require.Len(t, got, 1)
	assert.Equal(t, "region", got[0].Name)
	assert.Equal(t, "us-west-2", got[0].Value)
	assert.Equal(t, "cluster", got[0].AssignStrategy)
	require.Len(t, got[0].Clusters, 2)
	assert.Equal(t, "cluster-1", got[0].Clusters[0].UID)
	assert.Equal(t, "us-west-2", got[0].Clusters[0].Value)
	assert.Equal(t, "cluster-2", got[0].Clusters[1].UID)
	assert.Equal(t, "us-west-2", got[0].Clusters[1].Value)
}

func TestExpandClusterTemplateProfileVariables_AllStrategySendsNoClusters(t *testing.T) {
	got := expandClusterTemplateProfileVariables([]interface{}{
		variableWithStrategy("region", "us-west-2", "all", "cluster-1"),
	})

	require.Len(t, got, 1)
	assert.Equal(t, "all", got[0].AssignStrategy)
	assert.Empty(t, got[0].Clusters)
}

func TestFlattenClusterTemplateProfileVariables_ReadsClusterIDsFromTemplate(t *testing.T) {
	got := flattenClusterTemplateProfileVariables([]*models.V1ClusterTemplateVariable{
		{
			Name:           "region",
			AssignStrategy: "cluster",
			Clusters: []*models.V1ClusterVariableValue{
				{UID: "cluster-1", Value: "us-east-1"},
				{UID: "cluster-2", Value: "us-east-1"},
			},
		},
	})

	require.Equal(t, 1, got.Len())
	variable := got.List()[0].(map[string]interface{})
	assert.Equal(t, "cluster", variable["assign_strategy"])
	assert.Equal(t, "us-east-1", variable["value"],
		"value falls back to the per-cluster value when the template stores none")
	assert.ElementsMatch(t, []interface{}{"cluster-1", "cluster-2"}, variable["cluster_ids"].(*schema.Set).List())
}

func TestFlattenClusterTemplateProfileVariables_EmptyStrategyIsCluster(t *testing.T) {
	got := flattenClusterTemplateProfileVariables([]*models.V1ClusterTemplateVariable{
		{Name: "region", Value: "us-west-2"},
	})

	require.Equal(t, 1, got.Len())
	variable := got.List()[0].(map[string]interface{})
	assert.Equal(t, "cluster", variable["assign_strategy"])
	assert.Equal(t, "us-west-2", variable["value"])
	assert.Equal(t, 0, variable["cluster_ids"].(*schema.Set).Len())
}

// A config -> API -> state round trip must hash identically, or every plan
// after apply shows drift.
func TestClusterTemplateVariables_RoundTripHashIsStable(t *testing.T) {
	for _, v := range []map[string]interface{}{
		variableWithStrategy("region", "us-west-2", "all"),
		variableWithStrategy("region", "us-west-2", "cluster", "cluster-1", "cluster-2"),
	} {
		roundTripped := flattenClusterTemplateProfileVariables(expandClusterTemplateProfileVariables([]interface{}{v}))
		require.Equal(t, 1, roundTripped.Len())
		assert.Equal(t, resourceClusterConfigTemplateVariableHash(v),
			resourceClusterConfigTemplateVariableHash(roundTripped.List()[0]),
			"strategy %q", v["assign_strategy"])
	}
}

// When the profile set does change, a profile with omitted variables must
// still carry its stored variables into the PUT, not clear them.
func TestResourceClusterConfigTemplate_OmittedVariablesCarriedOnProfileAdd(t *testing.T) {
	oldRaw := map[string]interface{}{
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
	newRaw := map[string]interface{}{
		"name":       "test-template",
		"cloud_type": "aws",
		"cluster_profile": []interface{}{
			map[string]interface{}{"id": "profile-1"},
			map[string]interface{}{"id": "profile-2"},
		},
	}

	d := buildClusterConfigTemplateUpdateResourceData(t, oldRaw, newRaw)
	require.True(t, d.HasChange("cluster_profile"))

	profiles := expandClusterTemplateProfiles(d.Get("cluster_profile").(*schema.Set).List())
	require.Len(t, profiles, 2)
	for _, p := range profiles {
		if p.UID == "profile-1" {
			require.Len(t, p.Variables, 1)
			assert.Equal(t, "region", p.Variables[0].Name)
			assert.Equal(t, "us-west-2", p.Variables[0].Value)
		} else {
			assert.Empty(t, p.Variables)
		}
	}
}
