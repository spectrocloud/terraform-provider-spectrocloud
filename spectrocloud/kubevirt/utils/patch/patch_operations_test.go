package patch

import (
	"bytes"
	"fmt"
	"testing"
)

func TestDiffStringMap(t *testing.T) {
	testCases := []struct {
		Path        string
		Old         map[string]interface{}
		New         map[string]interface{}
		ExpectedOps PatchOperations
	}{
		{
			Path: "/parent/",
			Old: map[string]interface{}{
				"one": "111",
				"two": "222",
			},
			New: map[string]interface{}{
				"one":   "111",
				"two":   "222",
				"three": "333",
			},
			ExpectedOps: []PatchOperation{
				&AddOperation{
					Path:  "/parent/three",
					Value: "333",
				},
			},
		},
		{
			Path: "/parent/",
			Old: map[string]interface{}{
				"one": "111",
				"two": "222",
			},
			New: map[string]interface{}{
				"one": "111",
				"two": "abcd",
			},
			ExpectedOps: []PatchOperation{
				&ReplaceOperation{
					Path:  "/parent/two",
					Value: "abcd",
				},
			},
		},
		{
			Path: "/parent/",
			Old: map[string]interface{}{
				"one": "111",
				"two": "222",
			},
			New: map[string]interface{}{
				"two":   "abcd",
				"three": "333",
			},
			ExpectedOps: []PatchOperation{
				&RemoveOperation{Path: "/parent/one"},
				&ReplaceOperation{
					Path:  "/parent/two",
					Value: "abcd",
				},
				&AddOperation{
					Path:  "/parent/three",
					Value: "333",
				},
			},
		},
		{
			Path: "/parent/",
			Old: map[string]interface{}{
				"one": "111",
				"two": "222",
			},
			New: map[string]interface{}{
				"two": "222",
			},
			ExpectedOps: []PatchOperation{
				&RemoveOperation{Path: "/parent/one"},
			},
		},
		{
			Path: "/parent/",
			Old: map[string]interface{}{
				"one": "111",
				"two": "222",
			},
			New: map[string]interface{}{},
			ExpectedOps: []PatchOperation{
				&RemoveOperation{Path: "/parent/one"},
				&RemoveOperation{Path: "/parent/two"},
			},
		},
		{
			Path: "/parent/",
			Old:  map[string]interface{}{},
			New: map[string]interface{}{
				"one": "111",
				"two": "222",
			},
			ExpectedOps: []PatchOperation{
				&AddOperation{
					Path: "/parent",
					Value: map[string]interface{}{
						"one": "111",
						"two": "222",
					},
				},
			},
		},
		{
			Path: "/parent/",
			Old: map[string]interface{}{
				"two~with-tilde":           "220",
				"three/with/three/slashes": "330",
			},
			New: map[string]interface{}{
				"one/with-slash":           "111",
				"three/with/three/slashes": "333",
			},
			ExpectedOps: []PatchOperation{
				&AddOperation{
					Path:  "/parent/one~1with-slash",
					Value: "111",
				},
				&RemoveOperation{
					Path: "/parent/two~0with-tilde",
				},
				&ReplaceOperation{
					Path:  "/parent/three~1with~1three~1slashes",
					Value: "333",
				},
			},
		},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			ops := DiffStringMap(tc.Path, tc.Old, tc.New)
			if !tc.ExpectedOps.Equal(ops) {
				t.Fatalf("Operations don't match.\nExpected: %v\nGiven:    %v\n", tc.ExpectedOps, ops)
			}
		})
	}
}

func TestEscapeJsonPointer(t *testing.T) {
	testCases := []struct {
		Input          string
		ExpectedOutput string
	}{
		{"simple", "simple"},
		{"special-chars,but no escaping", "special-chars,but no escaping"},
		{"escape-this/forward-slash", "escape-this~1forward-slash"},
		{"escape-this~tilde", "escape-this~0tilde"},
	}
	for _, tc := range testCases {
		output := escapeJsonPointer(tc.Input)
		if output != tc.ExpectedOutput {
			t.Fatalf("Expected %q as after escaping %q, given: %q",
				tc.ExpectedOutput, tc.Input, output)
		}
	}
}

// TestPatchOperationsMarshalJSON — Batch 11. Covers the 7 previously-
// unreached type-methods: MarshalJSON + String on Replace/Add/Remove
// operations plus PatchOperations.MarshalJSON.
func TestPatchOperationsMarshalJSON(t *testing.T) {
	rep := &ReplaceOperation{Path: "/a", Value: "b"}
	repJSON, err := rep.MarshalJSON()
	if err != nil {
		t.Fatalf("ReplaceOperation.MarshalJSON: %v", err)
	}
	if !bytes.Contains(repJSON, []byte(`"op":"replace"`)) {
		t.Errorf("expected op=replace in %s", repJSON)
	}
	if s := rep.String(); s == "" || !bytes.Contains([]byte(s), []byte(`"replace"`)) {
		t.Errorf("ReplaceOperation.String: unexpected %q", s)
	}

	add := &AddOperation{Path: "/b", Value: 42}
	addJSON, err := add.MarshalJSON()
	if err != nil {
		t.Fatalf("AddOperation.MarshalJSON: %v", err)
	}
	if !bytes.Contains(addJSON, []byte(`"op":"add"`)) {
		t.Errorf("expected op=add in %s", addJSON)
	}
	if s := add.String(); s == "" || !bytes.Contains([]byte(s), []byte(`"add"`)) {
		t.Errorf("AddOperation.String: unexpected %q", s)
	}

	rem := &RemoveOperation{Path: "/c"}
	remJSON, err := rem.MarshalJSON()
	if err != nil {
		t.Fatalf("RemoveOperation.MarshalJSON: %v", err)
	}
	if !bytes.Contains(remJSON, []byte(`"op":"remove"`)) {
		t.Errorf("expected op=remove in %s", remJSON)
	}
	if s := rem.String(); s == "" || !bytes.Contains([]byte(s), []byte(`"remove"`)) {
		t.Errorf("RemoveOperation.String: unexpected %q", s)
	}

	// PatchOperations.MarshalJSON: emits an array of the individual ops.
	po := PatchOperations{rep, add, rem}
	poJSON, err := po.MarshalJSON()
	if err != nil {
		t.Fatalf("PatchOperations.MarshalJSON: %v", err)
	}
	if !bytes.HasPrefix(poJSON, []byte("[")) {
		t.Errorf("expected JSON array; got %s", poJSON)
	}
}
