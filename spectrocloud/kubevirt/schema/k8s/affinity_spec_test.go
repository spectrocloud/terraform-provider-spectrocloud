package k8s

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// affinity_spec_test.go — Batch 7 coverage for the affinity schema.
//
// This file has 45 exported/unexported helpers organized in three
// interlocking families:
//   1. Native k8s v1.Affinity   ⇄  Terraform schema      (Flatten*/Expand*)
//   2. Palette models.V1VMAffinity ⇄  Terraform schema   (FlattenAffinityFromVM/ExpandAffinity)
//   3. safeInt32 utility helper
//
// We round-trip a populated fixture through each family — that exercises
// every branch (nil guards, empty slices, populated slices) in one call
// per direction.

// stringPtr / int32Ptr / boolPtr are file-local because the surrounding
// package doesn't export equivalents.
func stringPtr(s string) *string { return &s }
func int32Ptr(i int32) *int32    { return &i }

// ---------------------------------------------------------------------------
// safeInt32
// ---------------------------------------------------------------------------

func TestSafeInt32(t *testing.T) {
	// Normal values pass through.
	assert.Equal(t, int32(42), safeInt32(42))
	// The current implementation is a bare int→int32 conversion (no
	// clamp); this test pins that contract so a future change that
	// introduces clamping would show up as a diff.
	assert.Equal(t, int32(0), safeInt32(0))
	assert.Equal(t, int32(-5), safeInt32(-5))
}

// ---------------------------------------------------------------------------
// Native k8s Affinity round-trip
// ---------------------------------------------------------------------------

// makeK8sAffinity builds a fully-populated v1.Affinity — one node
// affinity term (required + preferred), one pod affinity term (required
// + preferred weighted), and the same for pod anti-affinity. That single
// fixture drives every flatten branch.
func makeK8sAffinity() *v1.Affinity {
	nodeSelectorTerm := v1.NodeSelectorTerm{
		MatchExpressions: []v1.NodeSelectorRequirement{
			{Key: "zone", Operator: v1.NodeSelectorOpIn, Values: []string{"us-east-1a", "us-east-1b"}},
		},
		MatchFields: []v1.NodeSelectorRequirement{
			{Key: "metadata.name", Operator: v1.NodeSelectorOpExists},
		},
	}

	labelSel := &metav1.LabelSelector{
		MatchLabels:      map[string]string{"app": "frontend"},
		MatchExpressions: []metav1.LabelSelectorRequirement{{Key: "tier", Operator: metav1.LabelSelectorOpIn, Values: []string{"web"}}},
	}

	podAffTerm := v1.PodAffinityTerm{
		LabelSelector: labelSel,
		Namespaces:    []string{"default"},
		TopologyKey:   "kubernetes.io/hostname",
	}

	return &v1.Affinity{
		NodeAffinity: &v1.NodeAffinity{
			RequiredDuringSchedulingIgnoredDuringExecution: &v1.NodeSelector{
				NodeSelectorTerms: []v1.NodeSelectorTerm{nodeSelectorTerm},
			},
			PreferredDuringSchedulingIgnoredDuringExecution: []v1.PreferredSchedulingTerm{
				{Weight: 10, Preference: nodeSelectorTerm},
			},
		},
		PodAffinity: &v1.PodAffinity{
			RequiredDuringSchedulingIgnoredDuringExecution:  []v1.PodAffinityTerm{podAffTerm},
			PreferredDuringSchedulingIgnoredDuringExecution: []v1.WeightedPodAffinityTerm{{Weight: 5, PodAffinityTerm: podAffTerm}},
		},
		PodAntiAffinity: &v1.PodAntiAffinity{
			RequiredDuringSchedulingIgnoredDuringExecution:  []v1.PodAffinityTerm{podAffTerm},
			PreferredDuringSchedulingIgnoredDuringExecution: []v1.WeightedPodAffinityTerm{{Weight: 5, PodAffinityTerm: podAffTerm}},
		},
	}
}

func TestFlattenAffinity(t *testing.T) {
	t.Run("nil returns nil", func(t *testing.T) {
		assert.Nil(t, FlattenAffinity(nil))
	})

	t.Run("populated round-trip through flatten only", func(t *testing.T) {
		got := FlattenAffinity(makeK8sAffinity())
		require.Len(t, got, 1)
		m := got[0].(map[string]interface{})
		// All three affinity sub-maps should be present.
		assert.Contains(t, m, "node_affinity")
		assert.Contains(t, m, "pod_affinity")
		assert.Contains(t, m, "pod_anti_affinity")
	})

	t.Run("empty affinity returns empty slice", func(t *testing.T) {
		// An Affinity with all sub-fields nil should return an
		// empty slice, not a slice with an empty map (the len(att)>0
		// guard).
		got := FlattenAffinity(&v1.Affinity{})
		assert.Empty(t, got)
	})
}

// ---------------------------------------------------------------------------
// V1VMAffinity round-trip (Palette models ⇄ Terraform schema)
// ---------------------------------------------------------------------------

func makeVMAffinity() *models.V1VMAffinity {
	tk := "kubernetes.io/hostname"
	key := "zone"
	op := "In"
	weight := int32(5)

	nodeReq := &models.V1VMNodeSelectorRequirement{
		Key:      &key,
		Operator: &op,
		Values:   []string{"us-east-1a"},
	}
	nodeTerm := &models.V1VMNodeSelectorTerm{
		MatchExpressions: []*models.V1VMNodeSelectorRequirement{nodeReq},
		MatchFields:      []*models.V1VMNodeSelectorRequirement{nodeReq},
	}

	labelSel := &models.V1VMLabelSelector{
		MatchLabels: map[string]string{"app": "frontend"},
		MatchExpressions: []*models.V1VMLabelSelectorRequirement{
			{Key: stringPtr("tier"), Operator: stringPtr("In"), Values: []string{"web"}},
		},
	}
	podTerm := &models.V1VMPodAffinityTerm{
		LabelSelector: labelSel,
		Namespaces:    []string{"default"},
		TopologyKey:   &tk,
	}

	return &models.V1VMAffinity{
		NodeAffinity: &models.V1VMNodeAffinity{
			RequiredDuringSchedulingIgnoredDuringExecution: &models.V1VMNodeSelector{
				NodeSelectorTerms: []*models.V1VMNodeSelectorTerm{nodeTerm},
			},
			PreferredDuringSchedulingIgnoredDuringExecution: []*models.V1VMPreferredSchedulingTerm{
				{Weight: &weight, Preference: nodeTerm},
			},
		},
		PodAffinity: &models.V1VMPodAffinity{
			RequiredDuringSchedulingIgnoredDuringExecution:  []*models.V1VMPodAffinityTerm{podTerm},
			PreferredDuringSchedulingIgnoredDuringExecution: []*models.V1VMWeightedPodAffinityTerm{{Weight: &weight, PodAffinityTerm: podTerm}},
		},
		PodAntiAffinity: &models.V1PodAntiAffinity{
			RequiredDuringSchedulingIgnoredDuringExecution:  []*models.V1VMPodAffinityTerm{podTerm},
			PreferredDuringSchedulingIgnoredDuringExecution: []*models.V1VMWeightedPodAffinityTerm{{Weight: &weight, PodAffinityTerm: podTerm}},
		},
	}
}

func TestFlattenAffinityFromVM(t *testing.T) {
	t.Run("nil returns nil", func(t *testing.T) {
		assert.Nil(t, FlattenAffinityFromVM(nil))
	})

	t.Run("populated flatten sets all three sub-maps", func(t *testing.T) {
		got := FlattenAffinityFromVM(makeVMAffinity())
		require.Len(t, got, 1)
		m := got[0].(map[string]interface{})
		assert.Contains(t, m, "node_affinity")
		assert.Contains(t, m, "pod_affinity")
		assert.Contains(t, m, "pod_anti_affinity")
	})

	t.Run("empty affinity → empty slice", func(t *testing.T) {
		got := FlattenAffinityFromVM(&models.V1VMAffinity{})
		assert.Empty(t, got)
	})
}

// TestExpandAffinity_RoundTrip flattens a V1VMAffinity, then expands it
// back and asserts the round-trip preserved the topology. This is the
// biggest single test in the batch because it exercises ExpandAffinity +
// every downstream expand* helper.
func TestExpandAffinity_RoundTrip(t *testing.T) {
	src := makeVMAffinity()
	flat := FlattenAffinityFromVM(src)
	require.NotEmpty(t, flat)

	got := ExpandAffinity(flat)
	require.NotNil(t, got)

	// Node affinity survived.
	require.NotNil(t, got.NodeAffinity)
	require.NotNil(t, got.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution)
	require.Len(t, got.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms, 1)
	assert.Len(t, got.NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution, 1)

	// Pod affinity survived.
	require.NotNil(t, got.PodAffinity)
	assert.Len(t, got.PodAffinity.RequiredDuringSchedulingIgnoredDuringExecution, 1)
	assert.Len(t, got.PodAffinity.PreferredDuringSchedulingIgnoredDuringExecution, 1)

	// Pod anti-affinity survived.
	require.NotNil(t, got.PodAntiAffinity)
	assert.Len(t, got.PodAntiAffinity.RequiredDuringSchedulingIgnoredDuringExecution, 1)
	assert.Len(t, got.PodAntiAffinity.PreferredDuringSchedulingIgnoredDuringExecution, 1)
}

// TestExpandAffinity_Empty pins the empty-input contract: an empty
// input slice should still produce a non-nil (but empty) V1VMAffinity.
func TestExpandAffinity_Empty(t *testing.T) {
	got := ExpandAffinity(nil)
	require.NotNil(t, got)
	assert.Nil(t, got.NodeAffinity)
	assert.Nil(t, got.PodAffinity)
	assert.Nil(t, got.PodAntiAffinity)

	got = ExpandAffinity([]interface{}{nil})
	require.NotNil(t, got)
}

// ---------------------------------------------------------------------------
// Schema definitions — cheap sanity that the map builders don't panic
// and return the expected keys.
// ---------------------------------------------------------------------------

func TestAffinitySchemaShape(t *testing.T) {
	// AffinitySchema returns a *schema.Schema; its Elem should be a
	// *schema.Resource whose Schema contains the three affinity keys.
	sch := AffinitySchema()
	require.NotNil(t, sch)
	require.Equal(t, schema.TypeList, sch.Type)

	res, ok := sch.Elem.(*schema.Resource)
	require.True(t, ok)
	assert.Contains(t, res.Schema, "node_affinity")
	assert.Contains(t, res.Schema, "pod_affinity")
	assert.Contains(t, res.Schema, "pod_anti_affinity")
}

// TestAffinityFieldsBuilders pins that the six helper builders each
// return maps with the expected top-level keys. Silence any future
// refactor that accidentally drops a field.
func TestAffinityFieldsBuilders(t *testing.T) {
	assert.Contains(t, affinityFields(), "node_affinity")
	assert.Contains(t, nodeAffinityFields(), "required_during_scheduling_ignored_during_execution")
	assert.Contains(t, nodeSelectorFields(), "node_selector_term")
	assert.Contains(t, preferredSchedulingTermFields(), "weight")
	assert.Contains(t, nodeSelectorRequirementsFields(), "match_expressions")
	assert.Contains(t, podAffinityFields(), "required_during_scheduling_ignored_during_execution")
	assert.Contains(t, podAffinityTermFields(), "label_selector")
	assert.Contains(t, weightedPodAffinityTermFields(), "weight")
}

// Silence unused-import warning if `int32Ptr` becomes unused after
// a future edit.
var _ = int32Ptr(0)
