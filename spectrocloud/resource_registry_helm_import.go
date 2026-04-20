package spectrocloud

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/spectrocloud/palette-sdk-go/client"
	"github.com/spectrocloud/palette-sdk-go/client/herr"
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

	// The import ID can be either a registry UID or a registry name
	importID := d.Id()
	if importID == "" {
		return nil, fmt.Errorf("helm registry import ID or name is required")
	}

	// Try to get by UID first
	registry, err := c.GetHelmRegistry(importID)
	if err != nil {
		// If not found by UID, try by name
		if !herr.IsNotFound(err) {
			return nil, fmt.Errorf("unable to retrieve Helm registry '%s': %s", importID, err)
		}
	} else if registry != nil {
		// Found by UID
		if err := setHelmRegistryState(d, registry, importID); err != nil {
			return nil, err
		}
		return c, nil
	}

	// Try to get by name
	registry, nameErr := c.GetHelmRegistryByName(importID)
	if nameErr != nil {
		return nil, fmt.Errorf("unable to retrieve Helm registry by name or id '%s': %s", importID, nameErr)
	}
	if registry == nil || registry.Metadata == nil {
		return nil, fmt.Errorf("helm registry '%s' not found", importID)
	}
	registryUID := registry.Metadata.UID
	if registryUID == "" {
		return nil, fmt.Errorf("helm registry with name '%s' found but has no UID", importID)
	}

	if err := setHelmRegistryState(d, registry, registryUID); err != nil {
		return nil, err
	}
	return c, nil
}

// setHelmRegistryState sets resource state from a Helm registry and the resolved UID.
func setHelmRegistryState(d *schema.ResourceData, registry *models.V1HelmRegistry, registryUID string) error {
	if err := d.Set("name", registry.Metadata.Name); err != nil {
		return err
	}
	if registry.Spec != nil && registry.Spec.Endpoint != nil && *registry.Spec.Endpoint != "" {
		if err := d.Set("endpoint", *registry.Spec.Endpoint); err != nil {
			return err
		}
	}
	if registry.Spec != nil {
		if err := d.Set("is_private", registry.Spec.IsPrivate); err != nil {
			return err
		}
	}
	d.SetId(registryUID)
	return nil
}
