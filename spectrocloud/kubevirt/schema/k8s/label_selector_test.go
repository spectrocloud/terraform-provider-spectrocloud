package k8s

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// label_selector_test.go — Batch 8. Covers expandLabelSelector plus the
// round-trip through flattenLabelSelector for completeness (the flatten
// side is already reached transitively by other tests but pinning it
// here keeps the file self-contained).

func TestLabelSelectorFieldsBuilder(t *testing.T) {
	// Both branches: updatable=true and updatable=false influence ForceNew.
	fUpd := labelSelectorFields(true)
	assert.Contains(t, fUpd, "match_expressions")
	assert.Contains(t, fUpd, "match_labels")
	assert.False(t, fUpd["match_labels"].ForceNew)

	fFixed := labelSelectorFields(false)
	assert.True(t, fFixed["match_labels"].ForceNew)
}

func TestExpandLabelSelector(t *testing.T) {
	t.Run("empty input", func(t *testing.T) {
		got := expandLabelSelector(nil)
		require.NotNil(t, got)
		assert.Nil(t, got.MatchLabels)
		assert.Nil(t, got.MatchExpressions)
	})

	t.Run("match_labels only", func(t *testing.T) {
		got := expandLabelSelector([]interface{}{
			map[string]interface{}{
				"match_labels": map[string]interface{}{"env": "prod", "tier": "web"},
			},
		})
		require.NotNil(t, got)
		assert.Equal(t, "prod", got.MatchLabels["env"])
		assert.Equal(t, "web", got.MatchLabels["tier"])
	})

	t.Run("match_expressions", func(t *testing.T) {
		got := expandLabelSelector([]interface{}{
			map[string]interface{}{
				"match_expressions": []interface{}{
					map[string]interface{}{
						"key":      "tier",
						"operator": "In",
						"values":   schema.NewSet(schema.HashString, []interface{}{"web", "api"}),
					},
				},
			},
		})
		require.NotNil(t, got)
		require.Len(t, got.MatchExpressions, 1)
		assert.Equal(t, "tier", got.MatchExpressions[0].Key)
		assert.Equal(t, metav1.LabelSelectorOpIn, got.MatchExpressions[0].Operator)
		assert.ElementsMatch(t, []string{"web", "api"}, got.MatchExpressions[0].Values)
	})
}

func TestFlattenLabelSelector(t *testing.T) {
	got := flattenLabelSelector(&metav1.LabelSelector{
		MatchLabels: map[string]string{"env": "prod"},
		MatchExpressions: []metav1.LabelSelectorRequirement{
			{Key: "tier", Operator: metav1.LabelSelectorOpIn, Values: []string{"web"}},
		},
	})
	require.Len(t, got, 1)
	m := got[0].(map[string]interface{})
	assert.Contains(t, m, "match_labels")
	assert.Contains(t, m, "match_expressions")
}
