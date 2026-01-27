package spectrocloud

import (
	"context"
	"fmt"
	"strings"
	"unicode"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/spectrocloud/palette-sdk-go/client"
	"github.com/spectrocloud/palette-sdk-go/client/herr"
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

// isLikelyUID checks if the string looks like a UID (no spaces, alphanumeric with dashes/underscores)
// UIDs typically don't contain spaces, while names can
func isLikelyUID(s string) bool {
	// If it contains spaces, it's definitely not a UID
	if strings.Contains(s, " ") {
		return false
	}
	// UIDs are typically alphanumeric with dashes/underscores, no special characters
	// and usually longer than typical names
	for _, r := range s {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '-' && r != '_' {
			return false
		}
	}
	return true
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

	var ecrErr error
	var basicErr error
	var ecrRegistry *models.V1EcrRegistry
	var basicRegistry *models.V1BasicOciRegistry

	// Determine if importID looks like a UID or a name
	// If it contains spaces, it's definitely a name, skip UID lookup
	isUID := isLikelyUID(importID)

	// Step 1: Try to retrieve the registry by UID (only if it looks like a UID)
	if isUID {
		// Try ECR first (most common)
		ecrRegistry, ecrErr = c.GetOciEcrRegistry(importID)
		if ecrErr != nil {
			// Check if error is 404/ResourceNotFound - if so, continue to try Basic registry or name lookup
			if herr.IsNotFound(ecrErr) {
				ecrErr = nil
			}
		}
		if ecrErr == nil && ecrRegistry != nil {
			// Found by UID as ECR registry
			registryUID = importID
			registryName = ecrRegistry.Metadata.Name
			registryType = "ecr"
			// Set required fields immediately
			if err := d.Set("name", registryName); err != nil {
				return nil, err
			}
			if err := d.Set("type", registryType); err != nil {
				return nil, err
			}
			d.SetId(registryUID)
			return c, nil
		}

		// Try Basic registry
		basicRegistry, basicErr = c.GetOciBasicRegistry(importID)
		if basicErr != nil {
			// Check if error is 404/ResourceNotFound - if so, continue to try name lookup
			if herr.IsNotFound(basicErr) {
				basicErr = nil
			}
		}
		if basicErr == nil && basicRegistry != nil {
			// Found by UID as Basic registry
			registryUID = importID
			registryName = basicRegistry.Metadata.Name
			registryType = "basic"
			// Set required fields immediately
			if err := d.Set("name", registryName); err != nil {
				return nil, err
			}
			if err := d.Set("type", registryType); err != nil {
				return nil, err
			}
			d.SetId(registryUID)
			return c, nil
		}
	}

	// Step 2: Try to get by name (either UID lookup failed, or importID contains spaces/is a name)
	registrySummary, nameErr := c.GetOciRegistryByName(importID)
	if nameErr != nil {
		// If we tried UID lookup first, include those errors in the message
		if isUID {
			return nil, fmt.Errorf("unable to retrieve OCI registry by UID or name '%s'. UID errors (ECR: %v, Basic: %v), Name error: %s", importID, ecrErr, basicErr, nameErr)
		}
		return nil, fmt.Errorf("unable to retrieve OCI registry by name '%s': %s", importID, nameErr)
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
		if registryType != "basic" && registryType != "ecr" {
			return nil, fmt.Errorf("unsupported registry type '%s' for registry '%s'. API returned type '%s', but only 'basic' and 'ecr' are supported", registryType, importID, registrySummary.Spec.RegistryType)
		}
		// Set required fields after determining type from summary
		if err := d.Set("name", registryName); err != nil {
			return nil, err
		}
		if err := d.Set("type", registryType); err != nil {
			return nil, err
		}
		d.SetId(registryUID)
		return c, nil
	} else {
		// If type is not in summary, try to determine by fetching full details
		// Try ECR first
		ecrRegistry, ecrErr = c.GetOciEcrRegistry(registryUID)
		if ecrErr == nil && ecrRegistry != nil {
			registryType = "ecr"
		} else {
			// Try Basic
			basicRegistry, basicErr = c.GetOciBasicRegistry(registryUID)
			if basicErr == nil && basicRegistry != nil {
				registryType = "basic"
			} else {
				return nil, fmt.Errorf("found registry by name but failed to retrieve full details: ECR error: %s, Basic error: %s", ecrErr, basicErr)
			}
		}
		// Set required fields after determining type by fetching full details
		if err := d.Set("name", registryName); err != nil {
			return nil, err
		}
		if err := d.Set("type", registryType); err != nil {
			return nil, err
		}
		d.SetId(registryUID)
		return c, nil
	}
}
