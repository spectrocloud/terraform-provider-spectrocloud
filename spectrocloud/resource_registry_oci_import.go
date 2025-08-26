package spectrocloud

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceRegistryOciImport(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	c := getV1ClientWithResourceContext(m, "")

	// The import ID should be the registry UID
	registryUID := d.Id()

	// Validate that the registry exists and we can access it
	registry, err := c.GetOciEcrRegistry(registryUID)
	if err != nil {
		return nil, fmt.Errorf("could not retrieve OCI registry for import: %s", err)
	}
	if registry == nil {
		return nil, fmt.Errorf("OCI registry with ID %s not found", registryUID)
	}

	// Set the registry name from the retrieved registry
	if err := d.Set("name", registry.Metadata.Name); err != nil {
		return nil, err
	}

	// Read all registry data to populate the state
	diags := resourceRegistryEcrRead(ctx, d, m)
	if diags.HasError() {
		return nil, fmt.Errorf("could not read OCI registry for import: %v", diags)
	}

	return []*schema.ResourceData{d}, nil
}
