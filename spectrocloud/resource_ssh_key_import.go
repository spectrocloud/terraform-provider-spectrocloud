package spectrocloud

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-sdk-go/client"
)

func resourceSSHKeyImport(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	_, err := GetCommonSSHKey(d, m)
	if err != nil {
		return nil, err
	}

	// Read all SSH key data to populate the state
	diags := resourceSSHKeyRead(ctx, d, m)
	if diags.HasError() {
		return nil, fmt.Errorf("could not read SSH key for import: %v", diags)
	}

	return []*schema.ResourceData{d}, nil
}

func GetCommonSSHKey(d *schema.ResourceData, m interface{}) (*client.V1Client, error) {
	// Parse the import ID which can be either:
	// 1. Simple format: ssh_key_id (defaults to project context)
	// 2. Context format: ssh_key_id:context (explicit context)
	importID := d.Id()
	if importID == "" {
		return nil, fmt.Errorf("SSH key import ID is required")
	}

	var context string
	var sshKeyID string

	// Check if the import ID contains context specification
	parts := strings.Split(importID, ":")
	if len(parts) == 2 {
		// Format: ssh_key_id:context
		sshKeyID = parts[0]
		context = parts[1]

		// Validate context
		if context != "project" && context != "tenant" {
			return nil, fmt.Errorf("invalid context '%s'. Expected 'project' or 'tenant'", context)
		}
	} else if len(parts) == 1 {
		// Format: ssh_key_id (default to project context)
		context = "project"
		sshKeyID = parts[0]
	} else {
		return nil, fmt.Errorf("invalid import ID format. Expected 'ssh_key_id' or 'ssh_key_id:context', got: %s", importID)
	}

	// Try the specified context first
	c := getV1ClientWithResourceContext(m, context)
	sshKey, err := c.GetSSHKey(sshKeyID)

	if err != nil || sshKey == nil {
		// If not found in specified context, try the other context
		otherContext := "tenant"
		if context == "tenant" {
			otherContext = "project"
		}

		c = getV1ClientWithResourceContext(m, otherContext)
		sshKey, err = c.GetSSHKey(sshKeyID)

		if err != nil {
			return nil, fmt.Errorf("unable to retrieve SSH key in either project or tenant context: %s", err)
		}
		if sshKey == nil {
			return nil, fmt.Errorf("SSH key with ID %s not found in either project or tenant context", sshKeyID)
		}

		// Update context to the one where we found the resource
		context = otherContext
	}

	// Set the required fields for the resource
	if err := d.Set("name", sshKey.Metadata.Name); err != nil {
		return nil, err
	}

	if err := d.Set("context", context); err != nil {
		return nil, err
	}

	// Note: We don't set the 'ssh_key' field during import because it's marked as sensitive.
	// This follows the same pattern as other sensitive fields (like credentials in cloud accounts).
	// The user will need to provide the ssh_key value in their Terraform configuration after import.
	// This is a security best practice to prevent sensitive data from being stored in state during import.

	// Set the ID to the SSH key ID
	d.SetId(sshKeyID)

	return c, nil
}
