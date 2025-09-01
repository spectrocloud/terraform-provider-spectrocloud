package spectrocloud

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-sdk-go/client"
)

func resourceRegistryHelmImport(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	_, err := GetCommonRegistryHelm(d, m)
	if err != nil {
		return nil, err
	}

	// Read all registry data to populate the state
	diags := resourceRegistryHelmRead(ctx, d, m)
	if diags.HasError() {
		return nil, fmt.Errorf("could not read Helm registry for import: %v", diags)
	}

	return []*schema.ResourceData{d}, nil
}

func GetCommonRegistryHelm(d *schema.ResourceData, m interface{}) (*client.V1Client, error) {
	// Helm registries are tenant-level resources only
	c := getV1ClientWithResourceContext(m, "tenant")

	// The import ID should be the registry UID
	registryUID := d.Id()
	if registryUID == "" {
		return nil, fmt.Errorf("helm registry import ID is required")
	}

	// Validate that the registry exists and we can access it
	registry, err := c.GetHelmRegistry(registryUID)
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve Helm registry: %s", err)
	}
	if registry == nil {
		return nil, fmt.Errorf("helm registry with ID %s not found", registryUID)
	}

	// Set the required fields for the resource
	if err := d.Set("name", registry.Metadata.Name); err != nil {
		return nil, err
	}

	// Set the endpoint URL
	if registry.Spec != nil && registry.Spec.Endpoint != nil && *registry.Spec.Endpoint != "" {
		if err := d.Set("endpoint", *registry.Spec.Endpoint); err != nil {
			return nil, err
		}
	}

	// Set the is_private field from the registry specification
	if registry.Spec != nil {
		if err := d.Set("is_private", registry.Spec.IsPrivate); err != nil {
			return nil, err
		}
	}

	// Set the ID to the registry ID
	d.SetId(registryUID)

	return c, nil
}
