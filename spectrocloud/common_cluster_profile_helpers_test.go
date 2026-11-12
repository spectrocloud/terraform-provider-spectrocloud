package spectrocloud

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

//// Covers pure-ish helpers in common_cluster_profile.go currently at 0%.

func TestResolveRegistryNameToUID_EmptyName(t *testing.T) {
	// Empty name short-circuits to ("", nil) regardless of registryType.
	c := getV1ClientWithResourceContext(unitTestMockAPIClient, "project")
	uid, err := resolveRegistryNameToUID(c, "", "oci")
	assert.NoError(t, err)
	assert.Equal(t, "", uid)
}

func TestResolveRegistryUIDToName_EmptyUID(t *testing.T) {
	c := getV1ClientWithResourceContext(unitTestMockAPIClient, "project")
	name, err := resolveRegistryUIDToName(c, "")
	assert.NoError(t, err)
	assert.Equal(t, "", name)
}
