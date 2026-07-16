package spectrocloud

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
)

//// Covers createWorkspaceBackupPolicy and updateWorkspaceBackupPolicy.

func prepareWorkspaceWithBackupPolicy() *schema.ResourceData {
	d := resourceWorkspace().TestResourceData()
	_ = d.Set("name", "ws-1")
	_ = d.Set("backup_policy", []interface{}{
		map[string]interface{}{
			"schedule":                  "0 0 * * *",
			"backup_location_id":        "loc-1",
			"prefix":                    "ws-backup",
			"expiry_in_hour":            24,
			"include_disks":             true,
			"include_cluster_resources": true,
			"include_all_clusters":      true,
			"cluster_uids":              schema.NewSet(schema.HashString, []interface{}{"c1", "c2"}),
			"namespaces":                schema.NewSet(schema.HashString, []interface{}{"default"}),
		},
	})
	d.SetId("test-ws-uid")
	return d
}

func TestCreateWorkspaceBackupPolicy(t *testing.T) {
	d := prepareWorkspaceWithBackupPolicy()
	c := getV1ClientWithResourceContext(unitTestMockAPIClient, "project")

	// Populated path — reaches CreateWorkspaceBackupConfig.
	err := createWorkspaceBackupPolicy(c, d)
	assert.NoError(t, err)

	// Empty path — no backup_policy set → the func returns nil early.
	d2 := resourceWorkspace().TestResourceData()
	err = createWorkspaceBackupPolicy(c, d2)
	assert.NoError(t, err)
}

func TestUpdateWorkspaceBackupPolicy(t *testing.T) {
	d := prepareWorkspaceWithBackupPolicy()
	c := getV1ClientWithResourceContext(unitTestMockAPIClient, "project")

	err := updateWorkspaceBackupPolicy(c, d)
	assert.NoError(t, err)

	// Empty policy → early nil return.
	d2 := resourceWorkspace().TestResourceData()
	err = updateWorkspaceBackupPolicy(c, d2)
	assert.NoError(t, err)
}
