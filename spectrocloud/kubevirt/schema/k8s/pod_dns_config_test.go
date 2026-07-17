package k8s

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
)

// pod_dns_config_test.go — Batch 7.
// Round-trip tests for the pod DNS config schema (native k8s + Palette
// VM model), covering 10 previously-unreached funcs.

// samplePodDNSConfigResourceData is a fully-populated schema value that
// exercises every branch of the expand helpers.
func samplePodDNSConfigResourceData() []interface{} {
	return []interface{}{
		map[string]interface{}{
			"nameservers": []interface{}{"1.1.1.1", "8.8.8.8"},
			"searches":    []interface{}{"internal.example.com"},
			"option": []interface{}{
				map[string]interface{}{"name": "ndots", "value": "2"},
				map[string]interface{}{"name": "edns0", "value": ""},
			},
		},
	}
}

func TestPodDnsConfigSchema(t *testing.T) {
	s := PodDnsConfigSchema()
	require.NotNil(t, s)
	assert.Equal(t, schema.TypeList, s.Type)
	assert.Equal(t, 1, s.MaxItems)

	res, ok := s.Elem.(*schema.Resource)
	require.True(t, ok)
	assert.Contains(t, res.Schema, "nameservers")
	assert.Contains(t, res.Schema, "searches")
	assert.Contains(t, res.Schema, "option")
}

// TestExpandPodDNSConfig covers both a populated input (all three
// branches) and the empty-input contract.
func TestExpandPodDNSConfig(t *testing.T) {
	t.Run("populated", func(t *testing.T) {
		got, err := ExpandPodDNSConfig(samplePodDNSConfigResourceData())
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, []string{"1.1.1.1", "8.8.8.8"}, got.Nameservers)
		assert.Equal(t, []string{"internal.example.com"}, got.Searches)
		require.Len(t, got.Options, 2)
		assert.Equal(t, "ndots", got.Options[0].Name)
		require.NotNil(t, got.Options[0].Value)
		assert.Equal(t, "2", *got.Options[0].Value)
	})

	t.Run("empty input yields empty struct", func(t *testing.T) {
		got, err := ExpandPodDNSConfig(nil)
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Empty(t, got.Nameservers)
		assert.Empty(t, got.Searches)
		assert.Empty(t, got.Options)
	})
}

func TestExpandPodDNSConfigToVM(t *testing.T) {
	t.Run("populated", func(t *testing.T) {
		got, err := ExpandPodDNSConfigToVM(samplePodDNSConfigResourceData())
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, []string{"1.1.1.1", "8.8.8.8"}, got.Nameservers)
		assert.Equal(t, []string{"internal.example.com"}, got.Searches)
		require.Len(t, got.Options, 2)
		assert.Equal(t, "ndots", got.Options[0].Name)
		assert.Equal(t, "2", got.Options[0].Value)
	})

	t.Run("empty input", func(t *testing.T) {
		got, err := ExpandPodDNSConfigToVM(nil)
		require.NoError(t, err)
		require.NotNil(t, got)
	})
}

// TestFlattenPodDNSConfig — flatten writes []string / []interface{} into
// the map. Real usage goes through schema.ResourceData which marshals
// []string → []interface{} before re-expand; in a raw round-trip the
// type assertion in ExpandPodDNSConfig fails silently, so we assert on
// the flattened map's shape directly.
func TestFlattenPodDNSConfig(t *testing.T) {
	t.Run("populated flatten", func(t *testing.T) {
		src := &v1.PodDNSConfig{
			Nameservers: []string{"1.1.1.1"},
			Searches:    []string{"example.com"},
			Options: []v1.PodDNSConfigOption{
				{Name: "ndots", Value: pStr("2")},
				{Name: "attempts", Value: pStr("3")},
			},
		}
		flat := FlattenPodDNSConfig(src)
		require.Len(t, flat, 1)

		m := flat[0].(map[string]interface{})
		assert.Equal(t, []string{"1.1.1.1"}, m["nameservers"])
		assert.Equal(t, []string{"example.com"}, m["searches"])
		opts := m["option"].([]interface{})
		require.Len(t, opts, 2)
		assert.Equal(t, "ndots", opts[0].(map[string]interface{})["name"])
		assert.Equal(t, "2", opts[0].(map[string]interface{})["value"])
	})

	t.Run("empty input yields empty slice", func(t *testing.T) {
		got := FlattenPodDNSConfig(&v1.PodDNSConfig{})
		assert.Empty(t, got)
	})

	t.Run("option with nil value is dropped from map", func(t *testing.T) {
		// flattenPodDNSConfigOptions only writes "value" when non-nil.
		src := &v1.PodDNSConfig{
			Options: []v1.PodDNSConfigOption{{Name: "edns0"}},
		}
		flat := FlattenPodDNSConfig(src)
		require.Len(t, flat, 1)
		opt := flat[0].(map[string]interface{})["option"].([]interface{})[0].(map[string]interface{})
		assert.Equal(t, "edns0", opt["name"])
		_, hasValue := opt["value"]
		assert.False(t, hasValue, "nil value should not be stored")
	})
}

// TestFlattenPodDNSConfigFromVM covers the Palette-model variant.
func TestFlattenPodDNSConfigFromVM(t *testing.T) {
	t.Run("nil returns empty slice", func(t *testing.T) {
		got := FlattenPodDNSConfigFromVM(nil)
		assert.Empty(t, got)
	})

	t.Run("populated flatten", func(t *testing.T) {
		// Same rationale as TestFlattenPodDNSConfig — assert on the
		// map shape rather than round-trip through Expand, because the
		// raw []string never reaches back through ResourceData in a
		// unit-test harness.
		src := &models.V1VMPodDNSConfig{
			Nameservers: []string{"1.1.1.1"},
			Searches:    []string{"example.com"},
			Options: []*models.V1VMPodDNSConfigOption{
				{Name: "ndots", Value: "2"},
			},
		}
		flat := FlattenPodDNSConfigFromVM(src)
		require.Len(t, flat, 1)

		m := flat[0].(map[string]interface{})
		assert.Equal(t, []string{"1.1.1.1"}, m["nameservers"])
		assert.Equal(t, []string{"example.com"}, m["searches"])
		opts := m["option"].([]interface{})
		require.Len(t, opts, 1)
		assert.Equal(t, "ndots", opts[0].(map[string]interface{})["name"])
		assert.Equal(t, "2", opts[0].(map[string]interface{})["value"])
	})

	t.Run("nil option entries are skipped", func(t *testing.T) {
		// flattenPodDNSConfigOptionsFromVM guards against nil pointers
		// in the input slice — pin that behavior.
		src := &models.V1VMPodDNSConfig{
			Options: []*models.V1VMPodDNSConfigOption{
				nil,
				{Name: "ndots", Value: "2"},
			},
		}
		flat := FlattenPodDNSConfigFromVM(src)
		require.Len(t, flat, 1)
		opts := flat[0].(map[string]interface{})["option"].([]interface{})
		assert.Len(t, opts, 1, "nil entries dropped")
	})
}

// pStr is a local helper for the k8s v1 model's string pointers.
func pStr(s string) *string { return &s }
