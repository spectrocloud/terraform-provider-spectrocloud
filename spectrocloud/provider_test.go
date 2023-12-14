package spectrocloud

import "testing"

func TestProvider(t *testing.T) {
	p := New("111.111.111")() // test version

	err := p.InternalValidate()

	if err != nil {
		t.Fatal(err)
	}
}
