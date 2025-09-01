package spectrocloud

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-sdk-go/client"
)

func resourceRegistryOciImport(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	_, err := GetCommonRegistryOci(d, m)
	if err != nil {
		return nil, err
	}

	// Read all registry data to populate the state
	diags := resourceRegistryEcrRead(ctx, d, m)
	if diags.HasError() {
		return nil, fmt.Errorf("could not read OCI registry for import: %v", diags)
	}

	return []*schema.ResourceData{d}, nil
}

func GetCommonRegistryOci(d *schema.ResourceData, m interface{}) (*client.V1Client, error) {
	// OCI registries are tenant-level resources only
	c := getV1ClientWithResourceContext(m, "tenant")

	// The import ID should be the registry UID
	registryUID := d.Id()
	if registryUID == "" {
		return nil, fmt.Errorf("OCI registry import ID is required")
	}

	// Try to retrieve the registry as ECR first (most common)
	registry, err := c.GetOciEcrRegistry(registryUID)
	if err != nil {
		// If ECR retrieval fails, try basic OCI registry
		basicRegistry, basicErr := c.GetOciBasicRegistry(registryUID)
		if basicErr != nil {
			return nil, fmt.Errorf("unable to retrieve OCI registry as either ECR or basic type: ECR error: %s, Basic error: %s", err, basicErr)
		}
		if basicRegistry == nil {
			return nil, fmt.Errorf("OCI registry with ID %s not found", registryUID)
		}

		// Set required fields for basic registry
		if err := d.Set("name", basicRegistry.Metadata.Name); err != nil {
			return nil, err
		}
		if err := d.Set("type", "basic"); err != nil {
			return nil, err
		}
		// Basic registries are typically private if they have authentication
		isPrivate := basicRegistry.Spec.Auth != nil
		if err := d.Set("is_private", isPrivate); err != nil {
			return nil, err
		}
	} else if registry == nil {
		return nil, fmt.Errorf("OCI registry with ID %s not found", registryUID)
	} else {
		// Set required fields for ECR registry
		if err := d.Set("name", registry.Metadata.Name); err != nil {
			return nil, err
		}
		if err := d.Set("type", "ecr"); err != nil {
			return nil, err
		}
		if err := d.Set("is_private", registry.Spec.IsPrivate); err != nil {
			return nil, err
		}
	}

	// Set the ID to the registry ID
	d.SetId(registryUID)

	return c, nil
}
