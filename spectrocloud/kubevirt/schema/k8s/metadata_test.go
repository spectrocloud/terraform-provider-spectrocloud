package k8s

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

// metadata_test.go — Batch 8. Covers filterSystemAnnotations,
// NamespacedMetadataSchema, ExpandMetadataToObjectMeta, and the four
// Flatten* variants (Data/DataFromVM/Metadata/MetadataFromVM).

func TestFilterSystemAnnotations(t *testing.T) {
	assert.Nil(t, filterSystemAnnotations(nil))

	got := filterSystemAnnotations(map[string]string{
		"kubevirt.io/latest-observed-api-version": "v1",
		"kubevirt.io/some-other-system":           "v",
		"user.annotation/key":                     "value",
	})
	assert.Len(t, got, 1)
	assert.Equal(t, "value", got["user.annotation/key"])
}

func TestNamespacedMetadataSchema(t *testing.T) {
	// Basic.
	s := NamespacedMetadataSchema("virtualmachine", false)
	require.NotNil(t, s)
	assert.Equal(t, schema.TypeList, s.Type)
	assert.True(t, s.Required)
	res, ok := s.Elem.(*schema.Resource)
	require.True(t, ok)
	assert.Contains(t, res.Schema, "namespace")
	assert.NotContains(t, res.Schema, "generate_name")

	// With generatable name.
	s = NamespacedMetadataSchema("vm", true)
	res, ok = s.Elem.(*schema.Resource)
	require.True(t, ok)
	assert.Contains(t, res.Schema, "generate_name")
}

func TestExpandMetadataToObjectMeta(t *testing.T) {
	// Empty input → zero-value meta.
	got := ExpandMetadataToObjectMeta(nil)
	assert.Equal(t, metav1.ObjectMeta{}, got)

	// Populated (both map[string]interface{} and map[string]string branches).
	got = ExpandMetadataToObjectMeta([]interface{}{
		map[string]interface{}{
			"annotations":      map[string]interface{}{"a": "1"},
			"labels":           map[string]interface{}{"env": "prod"},
			"generate_name":    "vm-",
			"name":             "myvm",
			"namespace":        "default",
			"resource_version": "42",
		},
	})
	assert.Equal(t, "myvm", got.Name)
	assert.Equal(t, "default", got.Namespace)
	assert.Equal(t, "42", got.ResourceVersion)
	assert.Equal(t, "vm-", got.GenerateName)
	assert.Equal(t, "1", got.Annotations["a"])
	assert.Equal(t, "prod", got.Labels["env"])

	// map[string]string branch.
	got = ExpandMetadataToObjectMeta([]interface{}{
		map[string]interface{}{
			"annotations": map[string]string{"a": "1"},
			"labels":      map[string]string{"env": "prod"},
		},
	})
	assert.Equal(t, "1", got.Annotations["a"])
	assert.Equal(t, "prod", got.Labels["env"])
}

func TestFlattenMetadataDataVolume(t *testing.T) {
	got := FlattenMetadataDataVolume(metav1.ObjectMeta{
		Name:            "dv1",
		Namespace:       "ns1",
		GenerateName:    "dv-",
		ResourceVersion: "42",
		UID:             types.UID("abc-uid"),
		Generation:      3,
		Annotations: map[string]string{
			"kubevirt.io/system": "ignore",
			"user":               "keep",
		},
		Labels: map[string]string{"env": "prod"},
	})
	require.Len(t, got, 1)
	m := got[0].(map[string]interface{})
	assert.Equal(t, "dv1", m["name"])
	assert.Equal(t, "ns1", m["namespace"])
	assert.Equal(t, "dv-", m["generate_name"])
	assert.Equal(t, "42", m["resource_version"])
	// UID gets fmt.Sprintf'd for the k8s meta variant.
	assert.Equal(t, "abc-uid", m["uid"])
	assert.EqualValues(t, 3, m["generation"])
	// System annotations filtered. FlattenStringMap emits
	// map[string]interface{} for consumption by ResourceData.
	anns := m["annotations"].(map[string]interface{})
	assert.Contains(t, anns, "user")
	assert.NotContains(t, anns, "kubevirt.io/system")
}

func TestFlattenMetadataDataVolumeFromVM(t *testing.T) {
	// nil input still returns a slice with a map.
	got := FlattenMetadataDataVolumeFromVM(nil)
	require.Len(t, got, 1)

	got = FlattenMetadataDataVolumeFromVM(&models.V1VMObjectMeta{
		Name:            "dv1",
		Namespace:       "ns1",
		GenerateName:    "dv-",
		ResourceVersion: "42",
		UID:             "abc-uid",
		Generation:      3,
		Annotations:     map[string]string{"kubevirt.io/system": "x", "user": "keep"},
		Labels:          map[string]string{"env": "prod"},
	})
	require.Len(t, got, 1)
	m := got[0].(map[string]interface{})
	assert.Equal(t, "dv1", m["name"])
	assert.Equal(t, "ns1", m["namespace"])
	assert.Equal(t, "42", m["resource_version"])
	assert.Equal(t, "abc-uid", m["uid"])
	anns := m["annotations"].(map[string]interface{})
	assert.Contains(t, anns, "user")
	assert.NotContains(t, anns, "kubevirt.io/system")
}

// metadataTestResource is a permissive schema.Resource used by the two
// FlattenMetadata* tests below — its keys mirror what FlattenMetadata
// tries to Set on the ResourceData.
func metadataTestResource() *schema.Resource {
	return &schema.Resource{Schema: map[string]*schema.Schema{
		"annotations":      {Type: schema.TypeMap, Optional: true, Elem: &schema.Schema{Type: schema.TypeString}},
		"labels":           {Type: schema.TypeMap, Optional: true, Elem: &schema.Schema{Type: schema.TypeString}},
		"name":             {Type: schema.TypeString, Optional: true},
		"resource_version": {Type: schema.TypeString, Optional: true},
		"uid":              {Type: schema.TypeString, Optional: true},
		"generation":       {Type: schema.TypeInt, Optional: true},
		"namespace":        {Type: schema.TypeString, Optional: true},
	}}
}

func TestFlattenMetadata(t *testing.T) {
	// nil resourceData is a soft no-op.
	assert.NoError(t, FlattenMetadata(metav1.ObjectMeta{}, nil))

	d := metadataTestResource().TestResourceData()
	err := FlattenMetadata(metav1.ObjectMeta{
		Name:            "vm1",
		Namespace:       "ns",
		ResourceVersion: "1",
		UID:             types.UID("uid-1"),
		Generation:      2,
		Annotations:     map[string]string{"kubevirt.io/system": "drop", "keep": "yes"},
		Labels:          map[string]string{"env": "prod"},
	}, d)
	require.NoError(t, err)
	assert.Equal(t, "vm1", d.Get("name"))
	assert.Equal(t, "ns", d.Get("namespace"))
	assert.Equal(t, "1", d.Get("resource_version"))
	assert.Equal(t, "uid-1", d.Get("uid"))
	assert.Equal(t, 2, d.Get("generation"))
	anns := d.Get("annotations").(map[string]interface{})
	assert.Contains(t, anns, "keep")
	assert.NotContains(t, anns, "kubevirt.io/system")
}

func TestFlattenMetadataFromVM(t *testing.T) {
	// Both nil branches early-return without error.
	assert.NoError(t, FlattenMetadataFromVM(nil, nil))
	d := metadataTestResource().TestResourceData()
	assert.NoError(t, FlattenMetadataFromVM(nil, d))

	err := FlattenMetadataFromVM(&models.V1VMObjectMeta{
		Name:            "vm1",
		Namespace:       "ns",
		ResourceVersion: "1",
		UID:             "uid-1",
		Generation:      2,
		Annotations:     map[string]string{"kubevirt.io/system": "drop", "keep": "yes"},
		Labels:          map[string]string{"env": "prod"},
	}, d)
	require.NoError(t, err)
	assert.Equal(t, "vm1", d.Get("name"))
	assert.Equal(t, "uid-1", d.Get("uid"))
	anns := d.Get("annotations").(map[string]interface{})
	assert.Contains(t, anns, "keep")
}
