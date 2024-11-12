package patch

import (
	"fmt"
	"testing"
)

func TestDiffStringMap(t *testing.T) {
	testCases := []struct {
		Path        string
		Old         map[string]interface{}
		New         map[string]interface{}
		ExpectedOps Operations
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
			ExpectedOps: Operations{
				{
					Path:  "/parent/three",
					Value: "333",
					Op:    opAdd,
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
			ExpectedOps: Operations{
				{
					Path:  "/parent/two",
					Value: "abcd",
					Op:    opReplace,
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
			ExpectedOps: Operations{
				{Path: "/parent/one", Op: opRemove},
				{
					Path:  "/parent/two",
					Value: "abcd",
					Op:    opReplace,
				},
				{
					Path:  "/parent/three",
					Value: "333",
					Op:    opAdd,
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
			ExpectedOps: Operations{
				{Path: "/parent/one", Op: opRemove},
			},
		},
		{
			Path: "/parent/",
			Old: map[string]interface{}{
				"one": "111",
				"two": "222",
			},
			New: map[string]interface{}{},
			ExpectedOps: Operations{
				{Path: "/parent/one", Op: opRemove},
				{Path: "/parent/two", Op: opRemove},
			},
		},
		{
			Path: "/parent/",
			Old:  map[string]interface{}{},
			New: map[string]interface{}{
				"one": "111",
				"two": "222",
			},
			ExpectedOps: Operations{
				{
					Path: "/parent",
					Value: map[string]interface{}{
						"one": "111",
						"two": "222",
					},
					Op: opAdd,
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
			ExpectedOps: Operations{
				{
					Path:  "/parent/one~1with-slash",
					Value: "111",
					Op:    opAdd,
				},
				{
					Path: "/parent/two~0with-tilde",
					Op:   opRemove,
				},
				{
					Path:  "/parent/three~1with~1three~1slashes",
					Value: "333",
					Op:    opReplace,
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

func TestEscapeJSONPointer(t *testing.T) {
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
		output := patchKeyEncoder.Replace(tc.Input)
		if output != tc.ExpectedOutput {
			t.Fatalf("Expected %q as after escaping %q, given: %q",
				tc.ExpectedOutput, tc.Input, output)
		}
	}
}
