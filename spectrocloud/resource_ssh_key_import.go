package spectrocloud

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceSSHKeyImport(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	c := getV1ClientWithResourceContext(m, "")

	// The import ID should be the SSH key UID
	sshKeyUID := d.Id()

	// Validate that the SSH key exists and we can access it
	sshKey, err := c.GetSSHKey(sshKeyUID)
	if err != nil {
		return nil, fmt.Errorf("could not retrieve SSH key for import: %s", err)
	}
	if sshKey == nil {
		return nil, fmt.Errorf("SSH key with ID %s not found", sshKeyUID)
	}

	// Set the SSH key name from the retrieved SSH key
	if err := d.Set("name", sshKey.Metadata.Name); err != nil {
		return nil, err
	}

	// Read all SSH key data to populate the state
	diags := resourceSSHKeyRead(ctx, d, m)
	if diags.HasError() {
		return nil, fmt.Errorf("could not read SSH key for import: %v", diags)
	}

	return []*schema.ResourceData{d}, nil
}
