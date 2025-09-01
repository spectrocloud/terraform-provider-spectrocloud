package spectrocloud

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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
	// Parse the import ID which can be either:
	// 1. Simple format: bsl_id (defaults to project context)
	// 2. Context format: context:bsl_id (explicit context)
	importID := d.Id()
	if importID == "" {
		return nil, fmt.Errorf("backup storage location import ID is required")
	}

	var context string
	var bslID string

	// Check if the import ID contains context specification
	parts := strings.Split(importID, ":")
	if len(parts) == 2 {
		// Format: context:bsl_id
		bslID = parts[0]
		context = parts[1]

		// Validate context
		if context != "project" && context != "tenant" {
			return nil, fmt.Errorf("invalid context '%s'. Expected 'project' or 'tenant'", context)
		}
	} else if len(parts) == 1 {
		// Format: bsl_id (default to project context)
		context = "project"
		bslID = parts[0]
	} else {
		return nil, fmt.Errorf("invalid import ID format. Expected 'bsl_id' or 'context:bsl_id', got: %s", importID)
	}

	// Try the specified context first
	c := getV1ClientWithResourceContext(m, context)
	bsl, err := c.GetBackupStorageLocation(bslID)

	if err != nil || bsl == nil {
		if err != nil {
			return nil, fmt.Errorf("unable to retrieve backup storage location in either project or tenant context: %s", err)
		}
		if bsl == nil {
			return nil, fmt.Errorf("backup storage location with ID %s not found", bslID)
		}
	}

	// Set the required fields for the resource
	if err := d.Set("name", bsl.Metadata.Name); err != nil {
		return nil, err
	}

	if err := d.Set("context", context); err != nil {
		return nil, err
	}

	// Set the storage provider by mapping from API type to Terraform constants
	storageProvider := mapAPITypeToTerraformProvider(string(*bsl.Spec.Storage))
	if err := d.Set("storage_provider", storageProvider); err != nil {
		return nil, err
	}

	// Set the ID to the backup storage location ID
	d.SetId(bslID)

	return c, nil
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
