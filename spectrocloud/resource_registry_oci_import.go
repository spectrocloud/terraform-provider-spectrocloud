package spectrocloud

import (
	"context"
	"fmt"
	"strings"

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

	// The import ID can be either a registry UID or a registry name
	importID := d.Id()
	if importID == "" {
		return nil, fmt.Errorf("OCI registry import ID or Import Name is required")
	}

	var registryUID string
	var registryName string
	var registryType string
	var isPrivate bool

	// Step 1: Try to retrieve the registry by UID (maintains backward compatibility)
	// Try ECR first (most common)
	ecrRegistry, ecrErr := c.GetOciEcrRegistry(importID)
	if ecrErr == nil && ecrRegistry != nil {
		// Found by UID as ECR registry
		registryUID = importID
		registryName = ecrRegistry.Metadata.Name
		registryType = "ecr"
		if ecrRegistry.Spec.IsPrivate != nil {
			isPrivate = *ecrRegistry.Spec.IsPrivate
		}
	} else {
		// Try Basic registry
		basicRegistry, basicErr := c.GetOciBasicRegistry(importID)
		if basicErr == nil && basicRegistry != nil {
			// Found by UID as Basic registry
			registryUID = importID
			registryName = basicRegistry.Metadata.Name
			registryType = "basic"
			isPrivate = basicRegistry.Spec.Auth != nil
		} else {
			// Step 2: UID lookup failed, try to get by name
			registrySummary, nameErr := c.GetOciRegistryByName(importID)
			if nameErr != nil {
				return nil, fmt.Errorf("unable to retrieve OCI registry by UID or name. UID errors (ECR: %s, Basic: %s), Name error: %s", ecrErr, basicErr, nameErr)
			}
			if registrySummary == nil || registrySummary.Metadata == nil {
				return nil, fmt.Errorf("OCI registry '%s' not found", importID)
			}

			// Extract UID and type from the summary
			registryUID = registrySummary.Metadata.UID
			if registryUID == "" {
				return nil, fmt.Errorf("OCI registry with name '%s' found but has no UID", importID)
			}

			registryName = registrySummary.Metadata.Name

			// Determine registry type from the summary
			if registrySummary.Spec != nil && registrySummary.Spec.RegistryType != "" {
				registryType = strings.ToLower(registrySummary.Spec.RegistryType)
			} else {
				// If type is not in summary, try to determine by fetching full details
				// Try ECR first
				ecrRegistry, ecrErr = c.GetOciEcrRegistry(registryUID)
				if ecrErr == nil && ecrRegistry != nil {
					registryType = "ecr"
					if ecrRegistry.Spec.IsPrivate != nil {
						isPrivate = *ecrRegistry.Spec.IsPrivate
					}
				} else {
					// Try Basic
					basicRegistry, basicErr = c.GetOciBasicRegistry(registryUID)
					if basicErr == nil && basicRegistry != nil {
						registryType = "basic"
						isPrivate = basicRegistry.Spec.Auth != nil
					} else {
						return nil, fmt.Errorf("found registry by name but failed to retrieve full details: ECR error: %s, Basic error: %s", ecrErr, basicErr)
					}
				}
			}

			// Fetch full details if we haven't already (when type was determined from summary)
			// If type was determined by fetching (else block above), we already have the registry
			if ecrRegistry == nil && basicRegistry == nil {
				switch registryType {
				case "ecr":
					ecrRegistry, ecrErr = c.GetOciEcrRegistry(registryUID)
					if ecrErr != nil || ecrRegistry == nil {
						return nil, fmt.Errorf("found registry by name but failed to retrieve ECR registry details: %s", ecrErr)
					}
					if ecrRegistry.Spec.IsPrivate != nil {
						isPrivate = *ecrRegistry.Spec.IsPrivate
					}
				case "basic":
					basicRegistry, basicErr = c.GetOciBasicRegistry(registryUID)
					if basicErr != nil || basicRegistry == nil {
						return nil, fmt.Errorf("found registry by name but failed to retrieve Basic registry details: %s", basicErr)
					}
					isPrivate = basicRegistry.Spec.Auth != nil
				default:
					return nil, fmt.Errorf("unsupported registry type '%s' for registry '%s'. Supported types are 'ecr' and 'basic'", registryType, importID)
				}
			}
		}
	}

	// Set required fields
	if err := d.Set("name", registryName); err != nil {
		return nil, err
	}
	if err := d.Set("type", registryType); err != nil {
		return nil, err
	}
	if err := d.Set("is_private", isPrivate); err != nil {
		return nil, err
	}

	// Set the ID to the registry UID
	d.SetId(registryUID)

	return c, nil
}
