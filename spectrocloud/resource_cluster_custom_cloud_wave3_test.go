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

// TestResourceClusterCustomCloudStateFuncs directly invokes the inline
// StateFunc closures on cloud_config.values and machine_pool.node_pool_config.
// StateFunc is only ever invoked by the SDK during a real Diff computation,
// not by TestResourceData().Set()/Get(), so the only way to exercise the
// closure bodies is to pull the *schema.Schema off the resource and call
// StateFunc directly.
func TestResourceClusterCustomCloudStateFuncs(t *testing.T) {
	res := resourceClusterCustomCloud()

	cloudConfigValues := res.Schema["cloud_config"].Elem.(*schema.Resource).Schema["values"]
	require.NotNil(t, cloudConfigValues.StateFunc)
	in := "kind: Cluster\nmetadata:\n    name: test"
	assert.Equal(t, NormalizeYamlContent(in), cloudConfigValues.StateFunc(in))
	assert.Panics(t, func() { cloudConfigValues.StateFunc(123) })

	nodePoolConfig := res.Schema["machine_pool"].Elem.(*schema.Resource).Schema["node_pool_config"]
	require.NotNil(t, nodePoolConfig.StateFunc)
	in2 := "kind: KubeadmControlPlane\nmetadata:\n    name: pool-1"
	assert.Equal(t, NormalizeYamlContent(in2), nodePoolConfig.StateFunc(in2))
	assert.Panics(t, func() { nodePoolConfig.StateFunc(456) })
}

func TestSetValueAtPath(t *testing.T) {
	tests := []struct {
		name       string
		data       interface{}
		pathParts  []string
		value      interface{}
		wantResult bool
		wantData   interface{}
	}{
		{
			name:       "empty path parts returns false",
			data:       map[string]interface{}{"a": "b"},
			pathParts:  []string{},
			value:      "x",
			wantResult: false,
			wantData:   map[string]interface{}{"a": "b"},
		},
		{
			name:       "map string, last part sets value",
			data:       map[string]interface{}{"a": "old"},
			pathParts:  []string{"a"},
			value:      "new",
			wantResult: true,
			wantData:   map[string]interface{}{"a": "new"},
		},
		{
			// The array created for "arr" starts empty (len=0,cap=0); the
			// extend-by-append inside the []interface{} case that follows
			// operates on a local copy of the slice header and is never
			// written back into the parent map, so all the array/nested
			// creation branches run but "arr" stays empty in the final
			// structure — a real quirk of this iterative (non-recursive)
			// implementation.
			name: "map string, not last, creates nested array then nested map then sets",
			data: map[string]interface{}{},
			pathParts: []string{
				"arr", "0", "x",
			},
			value:      "v",
			wantResult: true,
			wantData: map[string]interface{}{
				"arr": []interface{}{},
			},
		},
		{
			name:       "map string, not last, creates nested array then nested array then sets",
			data:       map[string]interface{}{},
			pathParts:  []string{"arr", "0", "1"},
			value:      "v2",
			wantResult: true,
			wantData: map[string]interface{}{
				"arr": []interface{}{},
			},
		},
		{
			name: "map string, not last, existing nested map navigates deeper",
			data: map[string]interface{}{
				"metadata": map[string]interface{}{"name": "old"},
			},
			pathParts:  []string{"metadata", "name"},
			value:      "new",
			wantResult: true,
			wantData: map[string]interface{}{
				"metadata": map[string]interface{}{"name": "new"},
			},
		},
		{
			name:       "array, part not an int returns false",
			data:       []interface{}{"a", "b"},
			pathParts:  []string{"foo"},
			value:      "x",
			wantResult: false,
			wantData:   []interface{}{"a", "b"},
		},
		{
			name:       "array, last part sets value in place",
			data:       []interface{}{"a", "b"},
			pathParts:  []string{"1"},
			value:      "z",
			wantResult: true,
			wantData:   []interface{}{"a", "z"},
		},
		{
			// Extending the slice via append reallocates its backing array
			// (starting from len=1,cap=1), which is local to this call and
			// never assigned back into *data, so the top-level slice is
			// left unchanged even though the extend/set branches all run.
			name:       "array, index beyond length extends with nils then sets",
			data:       []interface{}{"a"},
			pathParts:  []string{"3"},
			value:      "z",
			wantResult: true,
			wantData:   []interface{}{"a"},
		},
		{
			name: "map interface{} last part sets and converts",
			data: map[interface{}]interface{}{
				"a": "old",
			},
			pathParts:  []string{"a"},
			value:      "new",
			wantResult: true,
			wantData:   map[string]interface{}{"a": "new"},
		},
		{
			// map[interface{}]interface{} is only ever committed back into
			// *data on the isLast branch of THIS case; since "arr" is not
			// the last path part here, the conversion/creation branches all
			// run but *data itself is left as the original, untouched map.
			name:       "map interface{} not last, creates nested array/map then sets",
			data:       map[interface{}]interface{}{},
			pathParts:  []string{"arr", "0", "y"},
			value:      "z",
			wantResult: true,
			wantData:   map[interface{}]interface{}{},
		},
		{
			// The loop shares a single *data pointer across iterations, so
			// when the nested map[interface{}]interface{} is converted and
			// committed on its (last-part) iteration, it overwrites the
			// entire top-level *data rather than just the nested "a" key.
			name: "map interface{} not last, existing nested map navigates deeper",
			data: map[interface{}]interface{}{
				"a": map[interface{}]interface{}{"b": "old"},
			},
			pathParts:  []string{"a", "b"},
			value:      "new",
			wantResult: true,
			wantData: map[string]interface{}{
				"b": "new",
			},
		},
		{
			name:       "default case (string) returns false",
			data:       "hello",
			pathParts:  []string{"x"},
			value:      "v",
			wantResult: false,
			wantData:   "hello",
		},
		{
			name:       "default case (int) returns false",
			data:       42,
			pathParts:  []string{"x"},
			value:      "v",
			wantResult: false,
			wantData:   42,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			data := tt.data
			got := setValueAtPath(&data, tt.pathParts, tt.value)
			assert.Equal(t, tt.wantResult, got)
			assert.Equal(t, tt.wantData, data)
		})
	}
}

func TestFindAndUpdatePatternBranches(t *testing.T) {
	t.Run("empty pattern parts returns false", func(t *testing.T) {
		var data interface{} = map[string]interface{}{"a": "b"}
		assert.False(t, findAndUpdatePattern(&data, []string{}, "v", ""))
	})

	t.Run("map interface{} branch matches and recurses", func(t *testing.T) {
		var data interface{} = map[interface{}]interface{}{
			"size":  10,
			"other": map[interface{}]interface{}{"size": 20},
		}
		modified := findAndUpdatePattern(&data, []string{"size"}, 99, "")
		assert.True(t, modified)
		// A match at this level converts *data from
		// map[interface{}]interface{} to map[string]interface{}.
		m, ok := data.(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, 99, m["size"])
	})

	t.Run("array branch recurses into elements and mutates them", func(t *testing.T) {
		var data interface{} = []interface{}{
			map[string]interface{}{"size": 1},
			map[string]interface{}{"size": 2},
		}
		modified := findAndUpdatePattern(&data, []string{"size"}, 99, "")
		assert.True(t, modified)
		arr, ok := data.([]interface{})
		require.True(t, ok)
		assert.Equal(t, 99, arr[0].(map[string]interface{})["size"])
		assert.Equal(t, 99, arr[1].(map[string]interface{})["size"])
	})

	t.Run("no match returns false", func(t *testing.T) {
		var data interface{} = []interface{}{map[string]interface{}{"other": 1}}
		assert.False(t, findAndUpdatePattern(&data, []string{"missing"}, 99, ""))
	})
}

func TestFindAndUpdateWildcardPatternBranches(t *testing.T) {
	t.Run("empty pattern returns false", func(t *testing.T) {
		var data interface{} = map[string]interface{}{"a": "b"}
		assert.False(t, findAndUpdateWildcardPattern(&data, "", "v", ""))
	})

	t.Run("map interface{} branch matches on substring and recurses", func(t *testing.T) {
		var data interface{} = map[interface{}]interface{}{
			"node-group-max-size": 5,
			"nested":              map[interface{}]interface{}{"node-group-max-size": 6},
		}
		modified := findAndUpdateWildcardPattern(&data, "max-size", "patched", "")
		assert.True(t, modified)
		// A match at this level converts *data from
		// map[interface{}]interface{} to map[string]interface{}.
		m, ok := data.(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, "patched", m["node-group-max-size"])
	})

	t.Run("array branch recurses into elements and mutates them", func(t *testing.T) {
		var data interface{} = []interface{}{
			map[string]interface{}{"hostname": "old-host"},
		}
		modified := findAndUpdateWildcardPattern(&data, "host", "new-host", "")
		assert.True(t, modified)
		arr, ok := data.([]interface{})
		require.True(t, ok)
		assert.Equal(t, "new-host", arr[0].(map[string]interface{})["hostname"])
	})

	t.Run("no match returns false", func(t *testing.T) {
		var data interface{} = map[string]interface{}{"foo": "bar"}
		assert.False(t, findAndUpdateWildcardPattern(&data, "nomatch", "v", ""))
	})
}

func TestExtractMachinePoolNameFromYAMLBranches(t *testing.T) {
	t.Run("missing node_pool_config returns empty", func(t *testing.T) {
		assert.Equal(t, "", extractMachinePoolNameFromYAML(map[string]interface{}{}))
	})

	t.Run("empty node_pool_config returns empty", func(t *testing.T) {
		assert.Equal(t, "", extractMachinePoolNameFromYAML(map[string]interface{}{"node_pool_config": ""}))
	})

	t.Run("doc without spec.replicas or spec.template returns empty", func(t *testing.T) {
		yaml := "kind: Secret\nmetadata:\n  name: some-secret"
		assert.Equal(t, "", extractMachinePoolNameFromYAML(map[string]interface{}{"node_pool_config": yaml}))
	})

	t.Run("worker MachineDeployment with only spec.template is found", func(t *testing.T) {
		yaml := "kind: MachineDeployment\nmetadata:\n  name: pool-worker\nspec:\n  template:\n    spec: {}"
		assert.Equal(t, "pool-worker", extractMachinePoolNameFromYAML(map[string]interface{}{"node_pool_config": yaml}))
	})

	t.Run("leading document with wrong shape is skipped and later valid one is used", func(t *testing.T) {
		// The leading document is a YAML sequence (not a map), so decoding it
		// into `map[string]interface{}` fails with a non-EOF error; since it
		// is well-formed YAML, the decoder still advances past it to the
		// next "---"-delimited document.
		yaml := "- a\n- b\n---\nkind: KubeadmControlPlane\nmetadata:\n  name: pool-recovered\nspec:\n  replicas: 3"
		assert.Equal(t, "pool-recovered", extractMachinePoolNameFromYAML(map[string]interface{}{"node_pool_config": yaml}))
	})
}

func TestFlattenMachinePoolConfigsCustomCloudWithOverridesBranches(t *testing.T) {
	t.Run("empty machine pools returns empty slice", func(t *testing.T) {
		d := resourceClusterCustomCloud().TestResourceData()
		out := flattenMachinePoolConfigsCustomCloudWithOverrides(nil, d)
		assert.Empty(t, out)
	})

	apiYAMLMultiline := `apiVersion: v1
kind: KubeadmControlPlane
metadata:
  name: pool-1
spec:
  replicas: 3
  template:
    spec:
      foo: bar
      baz: qux`

	t.Run("no overrides in state uses normalized API values", func(t *testing.T) {
		d := resourceClusterCustomCloud().TestResourceData()
		require.NoError(t, d.Set("machine_pool", schema.NewSet(resourceMachinePoolCustomCloudHash, []interface{}{
			map[string]interface{}{
				"name":             "pool-1",
				"node_pool_config": "kind: KubeadmControlPlane\nmetadata:\n  name: pool-1",
			},
		})))
		pools := []*models.V1CustomMachinePoolConfig{
			{Name: "pool-1", Size: 3, IsControlPlane: boolPtr(true), UseControlPlaneAsWorker: true, Values: apiYAMLMultiline},
		}
		out := flattenMachinePoolConfigsCustomCloudWithOverrides(pools, d)
		require.Len(t, out, 1)
		mp := out[0].(map[string]interface{})
		assert.Equal(t, "pool-1", mp["name"])
		assert.Contains(t, mp["node_pool_config"], "pool-1")
	})

	t.Run("overrides with no drift preserves original config", func(t *testing.T) {
		currentConfig := "apiVersion: v1\nkind: Cluster\nmetadata:\n  name: ${region}\nspec:\n  template:\n    x: y"
		overrides := map[string]interface{}{"region": "us-west-2"}
		expectedApplied, err := applyYamlOverridesWithTemplates(currentConfig, overrides)
		require.NoError(t, err)

		d := resourceClusterCustomCloud().TestResourceData()
		require.NoError(t, d.Set("machine_pool", schema.NewSet(resourceMachinePoolCustomCloudHash, []interface{}{
			map[string]interface{}{
				"name":             "pool-1",
				"node_pool_config": currentConfig,
				"overrides":        overrides,
			},
		})))
		pools := []*models.V1CustomMachinePoolConfig{{Name: "pool-1", Values: expectedApplied}}
		out := flattenMachinePoolConfigsCustomCloudWithOverrides(pools, d)
		mp := out[0].(map[string]interface{})
		assert.Equal(t, NormalizeYamlContent(currentConfig), mp["node_pool_config"])
		assert.Equal(t, overrides, mp["overrides"])
	})

	t.Run("overrides with drift uses API config", func(t *testing.T) {
		currentConfig := "apiVersion: v1\nkind: Cluster\nmetadata:\n  name: ${region}\nspec:\n  template:\n    x: y"
		overrides := map[string]interface{}{"region": "us-west-2"}
		apiValues := "apiVersion: v1\nkind: Cluster\nmetadata:\n  name: totally-different\nspec:\n  template:\n    x: z"

		d := resourceClusterCustomCloud().TestResourceData()
		require.NoError(t, d.Set("machine_pool", schema.NewSet(resourceMachinePoolCustomCloudHash, []interface{}{
			map[string]interface{}{
				"name":             "pool-1",
				"node_pool_config": currentConfig,
				"overrides":        overrides,
			},
		})))
		pools := []*models.V1CustomMachinePoolConfig{{Name: "pool-1", Values: apiValues}}
		out := flattenMachinePoolConfigsCustomCloudWithOverrides(pools, d)
		mp := out[0].(map[string]interface{})
		assert.Equal(t, NormalizeYamlContent(apiValues), mp["node_pool_config"])
	})

	t.Run("taints present are flattened", func(t *testing.T) {
		d := resourceClusterCustomCloud().TestResourceData()
		require.NoError(t, d.Set("machine_pool", schema.NewSet(resourceMachinePoolCustomCloudHash, []interface{}{
			map[string]interface{}{"name": "pool-1", "node_pool_config": "kind: X\nmetadata:\n  name: pool-1"},
		})))
		pools := []*models.V1CustomMachinePoolConfig{
			{
				Name:   "pool-1",
				Values: "kind: X\nmetadata:\n  name: pool-1",
				Taints: []*models.V1Taint{{Key: "k", Value: "v", Effect: "NoSchedule"}},
			},
		}
		out := flattenMachinePoolConfigsCustomCloudWithOverrides(pools, d)
		mp := out[0].(map[string]interface{})
		assert.NotEmpty(t, mp["taints"])
	})
}

func TestFlattenCloudConfigsValuesCustomCloudWithOverridesBranches(t *testing.T) {
	t.Run("nil config returns empty slice", func(t *testing.T) {
		d := resourceClusterCustomCloud().TestResourceData()
		out := flattenCloudConfigsValuesCustomCloudWithOverrides(nil, d)
		assert.Empty(t, out)
	})

	t.Run("overrides with no drift preserves original config", func(t *testing.T) {
		currentValues := "apiVersion: v1\nkind: Cluster\nmetadata:\n  name: ${region}\nspec:\n  template:\n    x: y"
		overrides := map[string]interface{}{"region": "us-west-2"}
		expectedApplied, err := applyYamlOverridesWithTemplates(currentValues, overrides)
		require.NoError(t, err)

		d := resourceClusterCustomCloud().TestResourceData()
		require.NoError(t, d.Set("cloud_config", []interface{}{
			map[string]interface{}{"values": currentValues, "overrides": overrides},
		}))
		config := &models.V1CustomCloudConfig{
			Spec: &models.V1CustomCloudConfigSpec{
				ClusterConfig: &models.V1CustomClusterConfig{Values: StringPtr(expectedApplied)},
			},
		}
		out := flattenCloudConfigsValuesCustomCloudWithOverrides(config, d)
		require.Len(t, out, 1)
		m := out[0].(map[string]interface{})
		assert.Equal(t, NormalizeYamlContent(currentValues), m["values"])
		assert.Equal(t, overrides, m["overrides"])
	})

	t.Run("overrides with drift uses API values", func(t *testing.T) {
		currentValues := "apiVersion: v1\nkind: Cluster\nmetadata:\n  name: ${region}\nspec:\n  template:\n    x: y"
		overrides := map[string]interface{}{"region": "us-west-2"}
		apiValues := "apiVersion: v1\nkind: Cluster\nmetadata:\n  name: totally-different\nspec:\n  template:\n    x: z"

		d := resourceClusterCustomCloud().TestResourceData()
		require.NoError(t, d.Set("cloud_config", []interface{}{
			map[string]interface{}{"values": currentValues, "overrides": overrides},
		}))
		config := &models.V1CustomCloudConfig{
			Spec: &models.V1CustomCloudConfigSpec{
				ClusterConfig: &models.V1CustomClusterConfig{Values: StringPtr(apiValues)},
			},
		}
		out := flattenCloudConfigsValuesCustomCloudWithOverrides(config, d)
		m := out[0].(map[string]interface{})
		assert.Equal(t, NormalizeYamlContent(apiValues), m["values"])
	})

	t.Run("no overrides preserves existing overrides from state", func(t *testing.T) {
		d := resourceClusterCustomCloud().TestResourceData()
		require.NoError(t, d.Set("cloud_config", []interface{}{
			map[string]interface{}{"values": "kind: Cluster\nmetadata:\n  name: old"},
		}))
		config := &models.V1CustomCloudConfig{
			Spec: &models.V1CustomCloudConfigSpec{
				ClusterConfig: &models.V1CustomClusterConfig{Values: StringPtr("kind: Cluster\nmetadata:\n  name: new")},
			},
		}
		out := flattenCloudConfigsValuesCustomCloudWithOverrides(config, d)
		m := out[0].(map[string]interface{})
		assert.Contains(t, m["values"], "new")
	})
}

func TestFlattenCloudConfigCustomGetConfigError(t *testing.T) {
	d := prepareCustomCloudResourceData(t)
	c := mustUnitClient(t, true)

	diags, hasError := flattenCloudConfigCustom(customCloudConfigUID, d, c)
	assert.True(t, hasError)
	assert.NotEmpty(t, diags)
}

func TestResourceClusterCustomCloudUpdateGetCloudConfigError(t *testing.T) {
	d := prepareCustomCloudResourceData(t)
	d.SetId("test-cluster-id")

	diags := resourceClusterCustomCloudUpdate(context.Background(), d, unitTestMockAPINegativeClient)
	assert.True(t, diags.HasError())
}

// buildCustomCloudCloudConfigChangeResourceData builds a *schema.ResourceData
// with a genuine SDK-computed diff for cloud_config, so that d.HasChange
// reports true (schema.TestResourceData().Set() alone cannot do this, since
// HasChange compares the state and diff readers, not the "set" writer).
func buildCustomCloudCloudConfigChangeResourceData(t *testing.T, oldValues, newValues string) *schema.ResourceData {
	t.Helper()
	res := resourceClusterCustomCloud()

	base := map[string]interface{}{
		"name":             "test-custom-cluster",
		"cloud":            customCloudType,
		"cloud_account_id": customCloudAccountUID,
		"cloud_config_id":  customCloudConfigUID,
		"machine_pool": []interface{}{
			map[string]interface{}{
				"control_plane":           true,
				"control_plane_as_worker": true,
				"node_pool_config":        customCloudPoolYAML,
			},
		},
	}

	oldRaw := map[string]interface{}{}
	for k, v := range base {
		oldRaw[k] = v
	}
	oldRaw["cloud_config"] = []interface{}{map[string]interface{}{"values": oldValues}}
	oldRD := schema.TestResourceDataRaw(t, res.Schema, oldRaw)
	oldRD.SetId("test-custom-cluster-id")
	oldState := oldRD.State()
	require.NotNil(t, oldState)

	newRaw := map[string]interface{}{}
	for k, v := range base {
		newRaw[k] = v
	}
	newRaw["cloud_config"] = []interface{}{map[string]interface{}{"values": newValues}}
	newConfig := terraform.NewResourceConfigRaw(newRaw)

	diff, err := res.Diff(context.Background(), oldState, newConfig, nil)
	require.NoError(t, err)
	require.NotNil(t, diff)

	finalRD, err := schema.InternalMap(res.Schema).Data(oldState, diff)
	require.NoError(t, err)
	finalRD.SetId("test-custom-cluster-id")
	return finalRD
}

func TestResourceClusterCustomCloudUpdateCloudConfigRealDiff(t *testing.T) {
	d := buildCustomCloudCloudConfigChangeResourceData(t,
		"kind: Cluster\nmetadata:\n  name: old-name",
		"kind: Cluster\nmetadata:\n  name: new-name",
	)
	require.True(t, d.HasChange("cloud_config"))

	diags := resourceClusterCustomCloudUpdate(context.Background(), d, unitTestMockAPIClient)
	assert.False(t, diags.HasError())
}

// buildCustomCloudMachinePoolChangeResourceData mirrors
// buildMachinePoolChangeResourceData in resource_cluster_apache_cloudstack_test.go,
// adapted for custom_cloud's node_pool_config/YAML-name-derived machine pools.
func buildCustomCloudMachinePoolChangeResourceData(t *testing.T, oldPools, newPools []interface{}) *schema.ResourceData {
	t.Helper()
	res := resourceClusterCustomCloud()

	base := map[string]interface{}{
		"name":             "test-custom-cluster",
		"cloud":            customCloudType,
		"cloud_account_id": customCloudAccountUID,
		"cloud_config_id":  customCloudConfigUID,
		"cloud_config": []interface{}{
			map[string]interface{}{
				"values": "kind: Cluster\nmetadata:\n  name: test-custom-cluster",
			},
		},
	}

	oldRaw := map[string]interface{}{}
	for k, v := range base {
		oldRaw[k] = v
	}
	oldRaw["machine_pool"] = oldPools
	oldRD := schema.TestResourceDataRaw(t, res.Schema, oldRaw)
	oldRD.SetId("test-custom-cluster-id")
	oldState := oldRD.State()
	require.NotNil(t, oldState)

	newRaw := map[string]interface{}{}
	for k, v := range base {
		newRaw[k] = v
	}
	newRaw["machine_pool"] = newPools
	newConfig := terraform.NewResourceConfigRaw(newRaw)

	diff, err := res.Diff(context.Background(), oldState, newConfig, nil)
	require.NoError(t, err)
	require.NotNil(t, diff)

	finalRD, err := schema.InternalMap(res.Schema).Data(oldState, diff)
	require.NoError(t, err)
	finalRD.SetId("test-custom-cluster-id")
	return finalRD
}

func TestResourceClusterCustomCloudUpdateMachinePoolCreateUpdateDeleteRealDiff(t *testing.T) {
	pool1 := map[string]interface{}{
		"control_plane":           true,
		"control_plane_as_worker": true,
		"node_pool_config":        customCloudPoolYAML, // pool-1, unchanged
	}
	pool2 := map[string]interface{}{
		"control_plane":           false,
		"control_plane_as_worker": false,
		"node_pool_config":        customCloudPool2YAML, // pool-2, to be removed
	}
	pool3 := map[string]interface{}{
		"control_plane":           false,
		"control_plane_as_worker": false,
		"node_pool_config":        "kind: MachineDeployment\nmetadata:\n  name: pool-3\nspec:\n  replicas: 1",
	}

	d := buildCustomCloudMachinePoolChangeResourceData(t,
		[]interface{}{pool1, pool2},
		[]interface{}{pool1, pool3},
	)
	require.True(t, d.HasChange("machine_pool"))

	diags := resourceClusterCustomCloudUpdate(context.Background(), d, unitTestMockAPIClient)
	assert.False(t, diags.HasError())
}

func TestResourceClusterCustomCloudUpdateMachinePoolModifiedRealDiff(t *testing.T) {
	oldPool := map[string]interface{}{
		"control_plane":           true,
		"control_plane_as_worker": true,
		"node_pool_config":        customCloudPoolYAML,
	}
	newPool := map[string]interface{}{
		"control_plane":           true,
		"control_plane_as_worker": true,
		"node_pool_config":        "kind: KubeadmControlPlane\nmetadata:\n  name: pool-1\nspec:\n  replicas: 7",
	}

	d := buildCustomCloudMachinePoolChangeResourceData(t,
		[]interface{}{oldPool},
		[]interface{}{newPool},
	)
	require.True(t, d.HasChange("machine_pool"))

	diags := resourceClusterCustomCloudUpdate(context.Background(), d, unitTestMockAPIClient)
	assert.False(t, diags.HasError())
}
