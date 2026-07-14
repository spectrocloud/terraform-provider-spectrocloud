package spectrocloud

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/spectrocloud/palette-sdk-go/api/models"
)

// prepareSSHKeyResourceData is the fixture for both positive and negative
// SSH key tests. The name/ssh_key values match what the mock's Read route
// echoes back — see tests/mockApiServer/routes/mockSSHKey.go — so
// flattenSSHKey's writes reproduce the exact input, which keeps the
// Update leg of testResourceCRUD asserting "no drift" as intended.
func prepareSSHKeyResourceData() *schema.ResourceData {
	d := resourceSSHKey().TestResourceData()
	_ = d.Set("name", "test-ssh-key")
	_ = d.Set("ssh_key", "ssh-rsa AAAAB3NzaC1yc2ETEST test-ssh-key")
	_ = d.Set("context", "project")
	return d
}

func TestResourceSSHKeyCRUD(t *testing.T) {
	testResourceCRUD(t, prepareSSHKeyResourceData, unitTestMockAPIClient,
		resourceSSHKeyCreate, resourceSSHKeyRead, resourceSSHKeyUpdate, resourceSSHKeyDelete)
}

func TestResourceSSHKeyCRUDNegative(t *testing.T) {
	// Create → the mock returns 409, so no ID is set and Create returns
	// diags containing the conflict message.
	t.Run("Create conflict", func(t *testing.T) {
		testResourceCRUDNegative(t, "Create", prepareSSHKeyResourceData,
			unitTestMockAPINegativeClient,
			resourceSSHKeyCreate, resourceSSHKeyRead, resourceSSHKeyUpdate, resourceSSHKeyDelete,
			false, "already exists")
	})

	// Read → the mock returns 404. resourceSSHKeyRead routes 404s through
	// handleReadError, which clears d.Id() and returns nil diags rather
	// than an error — so the "diags contain msg" check would fail here.
	// Assert the clear-ID contract directly instead.
	t.Run("Read 404 clears id", func(t *testing.T) {
		d := prepareSSHKeyResourceData()
		d.SetId("stale-uid")

		diags := resourceSSHKeyRead(context.Background(), d, unitTestMockAPINegativeClient)
		assert.Empty(t, diags, "404 on Read should be swallowed by handleReadError")
		assert.Empty(t, d.Id(), "handleReadError should clear the ID on NotFound")
	})

	t.Run("Update failure", func(t *testing.T) {
		testResourceCRUDNegative(t, "Update", prepareSSHKeyResourceData,
			unitTestMockAPINegativeClient,
			resourceSSHKeyCreate, resourceSSHKeyRead, resourceSSHKeyUpdate, resourceSSHKeyDelete,
			true, "Invalid SSH key update")
	})

	t.Run("Delete failure", func(t *testing.T) {
		testResourceCRUDNegative(t, "Delete", prepareSSHKeyResourceData,
			unitTestMockAPINegativeClient,
			resourceSSHKeyCreate, resourceSSHKeyRead, resourceSSHKeyUpdate, resourceSSHKeyDelete,
			true, "not found")
	})
}

// TestToSSHKey exercises the pure schema→model converter in isolation so
// a break here fails one test with a clear signal, rather than showing
// up as an opaque Create failure.
func TestToSSHKey(t *testing.T) {
	d := prepareSSHKeyResourceData()

	got, err := toSSHKey(d)
	require.NoError(t, err)
	require.NotNil(t, got)
	require.NotNil(t, got.Metadata)
	require.NotNil(t, got.Spec)
	assert.Equal(t, "test-ssh-key", got.Metadata.Name)
	assert.Equal(t, "ssh-rsa AAAAB3NzaC1yc2ETEST test-ssh-key", got.Spec.PublicKey)
}

// TestFlattenSSHKey covers the reverse — model→schema. Fed a fixture that
// matches what the mock's Read route returns, the resource data should
// come back with the expected name/ssh_key.
func TestFlattenSSHKey(t *testing.T) {
	d := resourceSSHKey().TestResourceData()
	in := &models.V1UserAssetSSH{
		Metadata: &models.V1ObjectMeta{Name: "flatten-name"},
		Spec:     &models.V1UserAssetSSHSpec{PublicKey: "ssh-rsa flatten-key"},
	}
	require.NoError(t, flattenSSHKey(in, d))
	assert.Equal(t, "flatten-name", d.Get("name"))
	assert.Equal(t, "ssh-rsa flatten-key", d.Get("ssh_key"))
}
