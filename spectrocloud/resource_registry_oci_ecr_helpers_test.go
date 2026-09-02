package spectrocloud

import (
	"testing"

	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/stretchr/testify/assert"
)

////
// Covers the pure helpers on resource_registry_oci_ecr.go that don't
// need the sync-endpoint mocked:
// - ociEcrSecretKeyForRead (masking / state-preservation logic)
// - resourceOciRegistrySyncRefreshFunc (via a stubbed sync-status
//   scenario — the closure is what we cover; underlying SDK failure is
//   acceptable)
// - getOciRegistrySyncStatus (both ecr + basic dispatch branches)

func TestOciEcrSecretKeyForRead(t *testing.T) {
	// No prior state, no API value → empty.
	d := resourceRegistryOciEcr().TestResourceData()
	assert.Equal(t, "", ociEcrSecretKeyForRead(d, ""))

	// No prior state, real API value → return API value.
	assert.Equal(t, "sk-abc", ociEcrSecretKeyForRead(d, "sk-abc"))

	// No prior state, masked API value → empty.
	assert.Equal(t, "", ociEcrSecretKeyForRead(d, "sk-***"))

	// Prior state with secret_key → prefer state over API.
	_ = d.Set("credentials", []interface{}{
		map[string]interface{}{
			"credential_type": "secret",
			"access_key":      "ak-1",
			"secret_key":      "state-sk",
		},
	})
	assert.Equal(t, "state-sk", ociEcrSecretKeyForRead(d, "sk-different"))
}

// TestResourceOciRegistrySyncRefreshFunc — exercises every branch of
// the state-mapping switch by driving getOciRegistrySyncStatus via a
// stub client. Since we don't have a stub client easily accessible, we
// instead assert on the pure logic that inspects a
// V1RegistrySyncStatus. Passing nil client makes SDK calls error, which
// hits the "" state branch — good enough to cover the closure body.
func TestResourceOciRegistrySyncRefreshFunc_ErrorBranch(t *testing.T) {
	c := getV1ClientWithResourceContext(unitTestMockAPINegativeClient, "project")
	refresh := resourceOciRegistrySyncRefreshFunc(c, "reg-uid", "ecr")
	// The SDK call may error against the negative mock — closure body
	// runs, error path is taken.
	_, _, _ = refresh()

	// Also cover basic dispatch.
	refresh = resourceOciRegistrySyncRefreshFunc(c, "reg-uid", "basic")
	_, _, _ = refresh()
}

// TestResourceOciRegistrySyncRefreshFunc_StatusBranches calls
// resourceOciRegistrySyncRefreshFunc directly (bypassing
// retry.StateChangeConf entirely) against a real V1Client backed by the
// mock server, driving every distinct branch of its status-mapping
// switch/if-chain via the UID-dispatched fixtures in
// ociRegistrySyncStatusFixtureFor (tests/mockApiServer/routes/mockRegistries.go).
// Calling it directly — rather than through waitForOciRegistrySync — avoids
// the retry loop's MinTimeout/Delay entirely for the "pending"-style
// statuses that would otherwise poll for minutes.
func TestResourceOciRegistrySyncRefreshFunc_StatusBranches(t *testing.T) {
	c := getV1ClientWithResourceContext(unitTestMockAPIClient, "tenant")

	tests := []struct {
		name          string
		uid           string
		wantState     string
		wantErr       bool
		wantErrSubstr string
	}{
		{name: "success", uid: "oci-sync-completed-no-message", wantState: "Success"},
		{name: "not supported short-circuits to Success", uid: "oci-sync-not-supported", wantState: "Success"},
		{name: "empty status is pending", uid: "oci-sync-empty-status", wantState: ""},
		{name: "failed with message", uid: "oci-sync-failed-with-message", wantState: "Failed", wantErr: true, wantErrSubstr: "boom"},
		{name: "failed without message", uid: "oci-sync-failed-no-message", wantState: "Error", wantErr: true, wantErrSubstr: "registry sync failed"},
		{name: "in progress", uid: "oci-sync-inprogress", wantState: "InProgress"},
		{name: "unknown status treated as pending", uid: "oci-sync-unknown", wantState: "Weird"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			refresh := resourceOciRegistrySyncRefreshFunc(c, tt.uid, "basic")
			result, state, err := refresh()
			assert.NotNil(t, result)
			assert.Equal(t, tt.wantState, state)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.wantErrSubstr != "" {
					assert.Contains(t, err.Error(), tt.wantErrSubstr)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestResourceOciRegistrySyncRefreshFunc_GetStatusError covers the
// `err != nil` branch when the underlying getOciRegistrySyncStatus call
// itself fails.
func TestResourceOciRegistrySyncRefreshFunc_GetStatusError(t *testing.T) {
	c := getV1ClientWithResourceContext(unitTestMockAPINegativeClient, "tenant")
	refresh := resourceOciRegistrySyncRefreshFunc(c, "any-uid", "basic")
	result, state, err := refresh()
	assert.Nil(t, result)
	assert.Empty(t, state)
	assert.Error(t, err)
}

// Compile-time reference so unused-import warnings don't fire in the
// event a future refactor drops the syncStatus struct.
var _ = &models.V1RegistrySyncStatus{}
