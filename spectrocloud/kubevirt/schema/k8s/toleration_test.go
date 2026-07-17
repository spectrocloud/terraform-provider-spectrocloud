package k8s

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
)

// toleration_test.go — Batch 8. Covers 5 helpers.

func TestTolerationSchema(t *testing.T) {
	s := TolerationSchema()
	require.NotNil(t, s)
	assert.Equal(t, schema.TypeList, s.Type)

	res, ok := s.Elem.(*schema.Resource)
	require.True(t, ok)
	for _, k := range []string{"effect", "key", "operator", "toleration_seconds", "value"} {
		assert.Contains(t, res.Schema, k)
	}
}

func TestTolerationFieldsBuilder(t *testing.T) {
	f := tolerationFields()
	for _, k := range []string{"effect", "key", "operator", "toleration_seconds", "value"} {
		assert.Contains(t, f, k)
	}
}

func TestExpandTolerations(t *testing.T) {
	t.Run("empty input", func(t *testing.T) {
		got, err := ExpandTolerations(nil)
		require.NoError(t, err)
		assert.Empty(t, got)
	})

	t.Run("populated", func(t *testing.T) {
		got, err := ExpandTolerations([]interface{}{
			map[string]interface{}{
				"effect":             "NoSchedule",
				"key":                "dedicated",
				"operator":           "Equal",
				"toleration_seconds": "30",
				"value":              "gpu",
			},
		})
		require.NoError(t, err)
		require.Len(t, got, 1)
		assert.Equal(t, "NoSchedule", got[0].Effect)
		assert.Equal(t, "dedicated", got[0].Key)
		assert.Equal(t, "Equal", got[0].Operator)
		assert.EqualValues(t, 30, got[0].TolerationSeconds)
		assert.Equal(t, "gpu", got[0].Value)
	})

	t.Run("invalid toleration_seconds rejected", func(t *testing.T) {
		_, err := ExpandTolerations([]interface{}{
			map[string]interface{}{"toleration_seconds": "not-a-number"},
		})
		assert.Error(t, err)
	})

	t.Run("empty toleration_seconds is treated as unset", func(t *testing.T) {
		got, err := ExpandTolerations([]interface{}{
			map[string]interface{}{"key": "k", "toleration_seconds": ""},
		})
		require.NoError(t, err)
		require.Len(t, got, 1)
		assert.EqualValues(t, 0, got[0].TolerationSeconds)
	})
}

func TestFlattenTolerations(t *testing.T) {
	// Node-lifecycle keys should be stripped.
	secs := int64(30)
	got := FlattenTolerations([]v1.Toleration{
		{Key: "node.kubernetes.io/not-ready", Operator: v1.TolerationOpExists},
		{Effect: v1.TaintEffectNoSchedule, Key: "dedicated", Operator: v1.TolerationOpEqual, Value: "gpu", TolerationSeconds: &secs},
	})
	require.Len(t, got, 1, "system tolerations dropped")
	m := got[0].(map[string]interface{})
	assert.Equal(t, "NoSchedule", m["effect"])
	assert.Equal(t, "dedicated", m["key"])
	assert.Equal(t, "Equal", m["operator"])
	assert.Equal(t, "30", m["toleration_seconds"])
	assert.Equal(t, "gpu", m["value"])
}

func TestFlattenTolerationsFromVM(t *testing.T) {
	got := FlattenTolerationsFromVM([]*models.V1VMToleration{
		nil,                                   // skipped
		{Key: "node.kubernetes.io/not-ready"}, // system, skipped
		{Effect: "NoSchedule", Key: "dedicated", Operator: "Equal", Value: "gpu", TolerationSeconds: 30},
	})
	require.Len(t, got, 1)
	m := got[0].(map[string]interface{})
	assert.Equal(t, "NoSchedule", m["effect"])
	assert.Equal(t, "dedicated", m["key"])
	assert.Equal(t, "30", m["toleration_seconds"])
}
