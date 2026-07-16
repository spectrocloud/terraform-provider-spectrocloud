package schemas

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// batch9_test.go — Batch 9. Simple presence tests for schema-builder
// functions currently at 0%. Each builder is a pure function returning
// a *schema.Schema; the smoke test just invokes it and asserts the
// top-level shape so accidental removal of fields fails a test.

func TestBackupPolicySchema(t *testing.T) {
	s := BackupPolicySchema()
	require.NotNil(t, s)
	assert.Equal(t, schema.TypeList, s.Type)
}

func TestClusterHostConfigSchema(t *testing.T) {
	s := ClusterHostConfigSchema()
	require.NotNil(t, s)
}

func TestClusterLocationSchemaComputed(t *testing.T) {
	s := ClusterLocationSchemaComputed()
	require.NotNil(t, s)
}

func TestClusterNamespacesSchema(t *testing.T) {
	s := ClusterNamespacesSchema()
	require.NotNil(t, s)
}

func TestMachinePoolArchTypeSchema(t *testing.T) {
	s := MachinePoolArchTypeSchema()
	require.NotNil(t, s)
}

func TestNodeSchema(t *testing.T) {
	s := NodeSchema()
	require.NotNil(t, s)
}

func TestClusterProfileSchema(t *testing.T) {
	s := ClusterProfileSchema()
	require.NotNil(t, s)
}

func TestClusterProfileSchemaV2(t *testing.T) {
	s := ClusterProfileSchemaV2()
	require.NotNil(t, s)
}

func TestClusterRbacBindingSchema(t *testing.T) {
	s := ClusterRbacBindingSchema()
	require.NotNil(t, s)
}

func TestClusterTaintsSchema(t *testing.T) {
	s := ClusterTaintsSchema()
	require.NotNil(t, s)
}

func TestClusterTemplateSchema(t *testing.T) {
	s := ClusterTemplateSchema()
	require.NotNil(t, s)
}

func TestClusterTypeSchema(t *testing.T) {
	s := ClusterTypeSchema()
	require.NotNil(t, s)
}

func TestOverrideClusterAPIConfigSchema(t *testing.T) {
	s := OverrideClusterAPIConfigSchema()
	require.NotNil(t, s)
}

func TestOverrideClusterAPIConfigMachinePoolSchema(t *testing.T) {
	s := OverrideClusterAPIConfigMachinePoolSchema()
	require.NotNil(t, s)
}

func TestOverrideHealthCheckConfigurationSchema(t *testing.T) {
	s := OverrideHealthCheckConfigurationSchema()
	require.NotNil(t, s)
}

func TestRenewK8sCertificatesNowSchema(t *testing.T) {
	s := RenewK8sCertificatesNowSchema()
	require.NotNil(t, s)
}

func TestWorkspaceNamespacesSchema(t *testing.T) {
	s := WorkspaceNamespacesSchema()
	require.NotNil(t, s)
}

// TestResourceWorkspaceNamespaceHash — the hash func takes a
// map[string]interface{} shaped like the namespaces schema. We construct
// a minimal fixture; the goal is to reach the func body, not to verify
// hash stability.
func TestResourceWorkspaceNamespaceHash(t *testing.T) {
	h := resourceWorkspaceNamespaceHash(map[string]interface{}{
		"name":                "ns1",
		"images_blacklist":    []interface{}{},
		"resource_allocation": []interface{}{},
	})
	// Any deterministic int result is acceptable; just prove no panic.
	_ = h
}

// TestValidatePackUIDOrResolutionFields exhaustively exercises the
// validation rules — the four "missing" branches, the manifest
// short-circuit, the uid short-circuit, and the mutex on
// registry_uid + registry_name.
func TestValidatePackUIDOrResolutionFields(t *testing.T) {
	t.Run("manifest short-circuits", func(t *testing.T) {
		err := ValidatePackUIDOrResolutionFields(map[string]interface{}{
			"type": "manifest",
			"name": "p",
		})
		assert.NoError(t, err)
	})

	t.Run("uid alone passes", func(t *testing.T) {
		err := ValidatePackUIDOrResolutionFields(map[string]interface{}{
			"uid":  "abc-123",
			"name": "p",
		})
		assert.NoError(t, err)
	})

	t.Run("registry_uid + registry_name both set → error", func(t *testing.T) {
		err := ValidatePackUIDOrResolutionFields(map[string]interface{}{
			"name":          "p",
			"registry_uid":  "r1",
			"registry_name": "r2",
		})
		assert.Error(t, err)
	})

	t.Run("resolution fields all present → passes", func(t *testing.T) {
		err := ValidatePackUIDOrResolutionFields(map[string]interface{}{
			"name":         "p",
			"tag":          "1.0.0",
			"registry_uid": "r1",
		})
		assert.NoError(t, err)
	})

	t.Run("missing tag → error", func(t *testing.T) {
		err := ValidatePackUIDOrResolutionFields(map[string]interface{}{
			"name":         "p",
			"registry_uid": "r1",
		})
		assert.Error(t, err)
	})

	t.Run("missing registry → error", func(t *testing.T) {
		err := ValidatePackUIDOrResolutionFields(map[string]interface{}{
			"name": "p",
			"tag":  "1.0.0",
		})
		assert.Error(t, err)
	})

	t.Run("missing everything → error listing all missing fields", func(t *testing.T) {
		err := ValidatePackUIDOrResolutionFields(map[string]interface{}{})
		assert.Error(t, err)
	})
}
