package k8s

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
)

// local_object_reference_test.go — Batch 8. Covers 4 helpers.

func TestLocalObjectReferenceFieldsBuilder(t *testing.T) {
	f := localObjectReferenceFields()
	assert.Contains(t, f, "name")
}

func TestLocalObjectReferenceSchema(t *testing.T) {
	s := LocalObjectReferenceSchema("a reference")
	require.NotNil(t, s)
	assert.Equal(t, schema.TypeList, s.Type)
	assert.Equal(t, 1, s.MaxItems)
	assert.Equal(t, "a reference", s.Description)
	res, ok := s.Elem.(*schema.Resource)
	require.True(t, ok)
	assert.Contains(t, res.Schema, "name")
}

func TestExpandLocalObjectReferences(t *testing.T) {
	assert.Nil(t, ExpandLocalObjectReferences(nil))

	got := ExpandLocalObjectReferences([]interface{}{
		map[string]interface{}{"name": "my-secret"},
	})
	require.NotNil(t, got)
	assert.Equal(t, "my-secret", got.Name)
}

func TestFlattenLocalObjectReferences(t *testing.T) {
	got := FlattenLocalObjectReferences(v1.LocalObjectReference{Name: "my-cm"})
	require.Len(t, got, 1)
	m := got[0].(map[string]interface{})
	assert.Equal(t, "my-cm", m["name"])
}
