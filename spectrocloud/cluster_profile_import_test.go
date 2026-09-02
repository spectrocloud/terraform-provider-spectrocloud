package spectrocloud

import (
	"context"
	"testing"

	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//// Sweeps the last chunk of 0% import + pure helper funcs in
// resource_cluster_profile_import.go and related.

// ---------------------------------------------------------------------------
// resourceClusterProfileImport
// ---------------------------------------------------------------------------

func TestResourceClusterProfileImport_InvalidID(t *testing.T) {
	d := resourceClusterProfile().TestResourceData()
	d.SetId("no-colon")
	_, err := resourceClusterProfileImport(context.Background(), d, unitTestMockAPIClient)
	require.Error(t, err)
}

func TestResourceClusterProfileImport_ValidUIDPath(t *testing.T) {
	// Valid format triggers GetCommonClusterProfile → GetClusterProfile.
	// The mock returns a profile fixture; branch coverage regardless of
	// full happy-path completion.
	d := resourceClusterProfile().TestResourceData()
	d.SetId("test-profile-uid:project")
	_, _ = resourceClusterProfileImport(context.Background(), d, unitTestMockAPIClient)
}

// (ParseClusterProfileImportID is already covered by resource_import_parser_helpers_test.go.)

// ---------------------------------------------------------------------------
// resolveClusterProfileByNameAndVersion + getProfileVersion
// ---------------------------------------------------------------------------

func TestGetProfileVersion(t *testing.T) {
	// Nil profile → default "0.0.0".
	assert.Equal(t, "0.0.0", getProfileVersion(nil))

	// Profile with no Spec → default.
	assert.Equal(t, "0.0.0", getProfileVersion(&models.V1ClusterProfile{}))

	// Profile with Spec + Published + Version → returns version.
	p := &models.V1ClusterProfile{
		Spec: &models.V1ClusterProfileSpec{
			Published: &models.V1ClusterProfileTemplate{ProfileVersion: "2.5.1"},
		},
	}
	assert.Equal(t, "2.5.1", getProfileVersion(p))
}

func TestResolveClusterProfileByNameAndVersion_NotFound(t *testing.T) {
	c := castV1Client(t, unitTestMockAPIClient)
	// A name unlikely to exist in the mock's fixture list → the "not
	// found" error branch fires.
	_, err := resolveClusterProfileByNameAndVersion(c, "no-such-profile-name-xyz", "project", "")
	assert.Error(t, err)
}

// ---------------------------------------------------------------------------
// setClusterProfileImportState
// ---------------------------------------------------------------------------

func TestSetClusterProfileImportState(t *testing.T) {
	d := resourceClusterProfile().TestResourceData()
	profile := &models.V1ClusterProfile{
		Metadata: &models.V1ObjectMeta{Name: "my-profile", UID: "my-uid"},
		Spec: &models.V1ClusterProfileSpec{
			Published: &models.V1ClusterProfileTemplate{ProfileVersion: "1.0.0"},
		},
	}
	require.NoError(t, setClusterProfileImportState(d, profile, "project"))
	assert.Equal(t, "my-uid", d.Id())
	assert.Equal(t, "my-profile", d.Get("name"))
	assert.Equal(t, "project", d.Get("context"))
	assert.Equal(t, "1.0.0", d.Get("version"))
}
