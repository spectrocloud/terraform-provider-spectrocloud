package schema

import "testing"

func CompareErrors(t *testing.T, actual error, expected error) {
	if actual != nil && expected != nil {
		if actual.Error() != expected.Error() {
			t.Errorf("Unexpected error: %v, expected: %v", actual.Error(), expected.Error())
		}
	}
	if (actual == nil && expected != nil) || (actual != nil && expected == nil) {
		t.Errorf("One of errors is nil while another is not")
	}
}
