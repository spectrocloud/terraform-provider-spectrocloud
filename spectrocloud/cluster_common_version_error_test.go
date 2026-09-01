package spectrocloud

import (
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/go-openapi/runtime"
)

// fakeClientResponse is a minimal runtime.ClientResponse whose Body() returns a
// caller-supplied JSON payload — mirroring the still-unconsumed body the generated
// go-swagger reader leaves on the *runtime.APIError at its default branch.
type fakeClientResponse struct {
	code int
	body string
}

func (f fakeClientResponse) Code() int                    { return f.code }
func (f fakeClientResponse) Message() string              { return "" }
func (f fakeClientResponse) GetHeader(_ string) string    { return "" }
func (f fakeClientResponse) GetHeaders(_ string) []string { return nil }
func (f fakeClientResponse) Body() io.ReadCloser {
	return io.NopCloser(strings.NewReader(f.body))
}

// newAPIError builds the opaque error the SDK returns for a non-204 profile PATCH.
func newAPIError(code int, body string) *runtime.APIError {
	return runtime.NewAPIError(
		"[PATCH /v1/spectroclusters/{uid}/profiles]",
		fakeClientResponse{code: code, body: body},
		code,
	)
}

const multiMinorBody = `{
  "code": "K8sMultiMinorUpgradeNotSupported",
  "message": "Kubernetes upgrades across multiple minor versions are not supported. Please update your cluster profile to sequentially upgrade the Kubernetes pack across each minor version.",
  "data": {"currentVersion": "1.30.4+k3s1", "targetVersion": "1.32.0+k3s1", "currentMinor": 30, "targetMinor": 32}
}`

const downgradeBody = `{
  "code": "K8sDowngradeAfterUpgradeNotSupported",
  "message": "Kubernetes downgrades are not supported after a successful upgrade.",
  "data": {"currentVersion": "1.32.0+k3s1", "targetVersion": "1.30.4+k3s1", "currentMinor": 32, "targetMinor": 30}
}`

func TestMapClusterProfileUpdateErr_MultiMinor(t *testing.T) {
	got := mapClusterProfileUpdateErr(newAPIError(400, multiMinorBody))
	if got == nil {
		t.Fatal("expected a mapped error, got nil")
	}
	msg := got.Error()
	if strings.Contains(msg, "status 400): {}") {
		t.Fatalf("mapped message is still opaque: %q", msg)
	}
	for _, want := range []string{
		codeK8sMultiMinorUpgradeNotSupported,
		"multiple minor versions",
		"one minor version at a time",
		"1.30.4+k3s1",
		"1.32.0+k3s1",
	} {
		if !strings.Contains(msg, want) {
			t.Errorf("multi-minor message %q missing %q", msg, want)
		}
	}
}

func TestMapClusterProfileUpdateErr_Downgrade(t *testing.T) {
	got := mapClusterProfileUpdateErr(newAPIError(400, downgradeBody))
	if got == nil {
		t.Fatal("expected a mapped error, got nil")
	}
	msg := got.Error()
	if strings.Contains(msg, "status 400): {}") {
		t.Fatalf("mapped message is still opaque: %q", msg)
	}
	for _, want := range []string{
		codeK8sDowngradeAfterUpgradeNotSupported,
		"downgrade",
		"1.32.0+k3s1",
		"1.30.4+k3s1",
	} {
		if !strings.Contains(msg, want) {
			t.Errorf("downgrade message %q missing %q", msg, want)
		}
	}
}

// TestMapClusterProfileUpdateErr_DistinguishesCodes is the DoD-named test proving
// the two machine codes yield different, non-opaque diagnostics.
func TestMapClusterProfileUpdateErr_DistinguishesCodes(t *testing.T) {
	multi := mapClusterProfileUpdateErr(newAPIError(400, multiMinorBody))
	down := mapClusterProfileUpdateErr(newAPIError(400, downgradeBody))
	if multi == nil || down == nil {
		t.Fatal("both codes must map to a non-nil error")
	}
	if multi.Error() == down.Error() {
		t.Fatalf("the two codes must produce distinct messages; both were %q", multi.Error())
	}
	if !strings.Contains(multi.Error(), codeK8sMultiMinorUpgradeNotSupported) {
		t.Errorf("multi-minor message missing its code: %q", multi.Error())
	}
	if !strings.Contains(down.Error(), codeK8sDowngradeAfterUpgradeNotSupported) {
		t.Errorf("downgrade message missing its code: %q", down.Error())
	}
}

// TestMapClusterProfileUpdateErr_DetailsFallback proves the edge/stylus framework
// shape (payload under `details`) is tolerated even though the mgmt API uses `data`.
func TestMapClusterProfileUpdateErr_DetailsFallback(t *testing.T) {
	body := `{"code":"K8sMultiMinorUpgradeNotSupported","message":"blocked","details":{"currentVersion":"1.28.0","targetVersion":"1.30.0","currentMinor":28,"targetMinor":30}}`
	got := mapClusterProfileUpdateErr(newAPIError(400, body))
	if got == nil {
		t.Fatal("expected mapped error")
	}
	for _, want := range []string{"1.28.0", "1.30.0"} {
		if !strings.Contains(got.Error(), want) {
			t.Errorf("details-fallback message %q missing %q", got.Error(), want)
		}
	}
}

func TestMapClusterProfileUpdateErr_Passthrough(t *testing.T) {
	tests := []struct {
		name string
		in   error
	}{
		{"nil", nil},
		{"plain error", errors.New("connection refused")},
		{"400 unknown code", newAPIError(400, `{"code":"SomethingElse","message":"nope"}`)},
		{"400 empty code", newAPIError(400, `{"message":"no code here"}`)},
		{"400 non-json body", newAPIError(400, `not json at all`)},
		{"non-400 with our code", newAPIError(500, multiMinorBody)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mapClusterProfileUpdateErr(tt.in)
			if tt.in == nil {
				if got != nil {
					t.Fatalf("nil in must map to nil out, got %v", got)
				}
				return
			}
			// Passthrough returns the identical error value, untouched.
			if got != tt.in {
				t.Fatalf("expected passthrough of original error %v, got %v", tt.in, got)
			}
		})
	}
}
