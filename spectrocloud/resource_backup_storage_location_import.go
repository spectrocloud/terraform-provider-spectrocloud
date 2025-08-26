package spectrocloud

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceBackupStorageLocationImport(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	c := getV1ClientWithResourceContext(m, "")

	// The import ID should be the backup storage location UID
	bslUID := d.Id()

	// Validate that the backup storage location exists and we can access it
	bsl, err := c.GetBackupStorageLocation(bslUID)
	if err != nil {
		return nil, fmt.Errorf("could not retrieve backup storage location for import: %s", err)
	}
	if bsl == nil {
		return nil, fmt.Errorf("backup storage location with ID %s not found", bslUID)
	}

	// Set the backup storage location name from the retrieved BSL
	if err := d.Set("name", bsl.Metadata.Name); err != nil {
		return nil, err
	}

	// Read all backup storage location data to populate the state
	diags := resourceBackupStorageLocationRead(ctx, d, m)
	if diags.HasError() {
		return nil, fmt.Errorf("could not read backup storage location for import: %v", diags)
	}

	return []*schema.ResourceData{d}, nil
}
