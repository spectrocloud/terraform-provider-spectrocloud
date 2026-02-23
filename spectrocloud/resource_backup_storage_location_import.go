package spectrocloud

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/spectrocloud/palette-sdk-go/client"
)

func resourceBackupStorageLocationImport(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	_, err := GetCommonBackupStorageLocation(d, m)
	if err != nil {
		return nil, err
	}

	// Read all backup storage location data to populate the state
	diags := resourceBackupStorageLocationRead(ctx, d, m)
	if diags.HasError() {
		return nil, fmt.Errorf("could not read backup storage location for import: %v", diags)
	}

	return []*schema.ResourceData{d}, nil
}

func GetCommonBackupStorageLocation(d *schema.ResourceData, m interface{}) (*client.V1Client, error) {
	// Import ID format: id_or_name:context (e.g. "my-bsl:project" or "uid-123:project")
	// Or id_or_name only (defaults to project context)
	resourceContext, bslID, err := parseBackupStorageLocationImportID(d)
	if err != nil {
		return nil, err
	}
	c := getV1ClientWithResourceContext(m, resourceContext)

	// Try by UID first, then by name (EKS pattern)
	bsl, err := c.GetBackupStorageLocation(bslID)
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve backup storage location: %w", err)
	}
	if bsl != nil {
		if err := setBackupStorageLocationImportState(d, bsl, resourceContext); err != nil {
			return nil, err
		}
		return c, nil
	}

	// Resolve by name
	bsls, err := c.ListBackupStorageLocation()
	if err != nil {
		return nil, fmt.Errorf("failed to list backup storage locations: %w", err)
	}
	for _, a := range bsls {
		if a.Metadata != nil && a.Metadata.Name == bslID {
			if err := setBackupStorageLocationImportState(d, a, resourceContext); err != nil {
				return nil, err
			}
			return c, nil
		}
	}
	return nil, fmt.Errorf("backup storage location with name '%s' not found in context %s", bslID, resourceContext)
}

// parseBackupStorageLocationImportID returns (context, idOrName). Supports:
// - id_or_name (defaults to project context)
// - id_or_name:context
func parseBackupStorageLocationImportID(d *schema.ResourceData) (string, string, error) {
	importID := d.Id()
	if importID == "" {
		return "", "", fmt.Errorf("backup storage location import ID is required")
	}
	parts := strings.Split(importID, ":")
	if len(parts) == 2 {
		if parts[1] != "project" && parts[1] != "tenant" {
			return "", "", fmt.Errorf("invalid context '%s'. Expected 'project' or 'tenant'", parts[1])
		}
		return parts[1], parts[0], nil
	}
	if len(parts) == 1 {
		return "project", parts[0], nil
	}
	return "", "", fmt.Errorf("invalid import ID format. Expected 'id_or_name' or 'id_or_name:context', got: %s", importID)
}

func setBackupStorageLocationImportState(d *schema.ResourceData, bsl *models.V1UserAssetsLocation, resourceContext string) error {
	if err := d.Set("name", bsl.Metadata.Name); err != nil {
		return err
	}
	if err := d.Set("context", resourceContext); err != nil {
		return err
	}
	storageProvider := mapAPITypeToTerraformProvider(string(*bsl.Spec.Storage))
	if err := d.Set("storage_provider", storageProvider); err != nil {
		return err
	}
	d.SetId(bsl.Metadata.UID)
	return nil
}

// mapAPITypeToTerraformProvider maps API storage type values to Terraform provider constants
func mapAPITypeToTerraformProvider(apiType string) string {
	switch apiType {
	case "s3":
		return "aws" // API uses "s3" but Terraform uses "aws"
	case "gcp":
		return "gcp" // Same in both
	case "minio":
		return "minio" // Same in both
	case "azure":
		return "azure" // Same in both
	default:
		// Default to aws if unknown type
		return "aws"
	}
}
