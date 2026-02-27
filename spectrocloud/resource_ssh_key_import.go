package spectrocloud

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-sdk-go/client"
	"github.com/spectrocloud/palette-sdk-go/client/herr"
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
	// 1. Simple format: id_or_name (defaults to project context)
	// 2. Context format: id_or_name:context (explicit context)
	// id_or_name can be the SSH key UID or the SSH key name (import by name).
	importID := d.Id()
	if importID == "" {
		return nil, fmt.Errorf("SSH key import ID is required")
	}

	var resCtx string
	var idOrName string

	parts := strings.Split(importID, ":")
	if len(parts) == 2 {
		idOrName = parts[0]
		resCtx = parts[1]
		if resCtx != "project" && resCtx != "tenant" {
			return nil, fmt.Errorf("invalid context '%s'. Expected 'project' or 'tenant'", resCtx)
		}
	} else if len(parts) == 1 {
		resCtx = "project"
		idOrName = parts[0]
	} else {
		return nil, fmt.Errorf("invalid import ID format. Expected 'id_or_name' or 'id_or_name:context', got: %s", importID)
	}

	// Try specified context first: by UID then by name
	c := getV1ClientWithResourceContext(m, resCtx)
	sshKey, err := c.GetSSHKey(idOrName)
	if err != nil && !herr.IsNotFound(err) {
		return nil, fmt.Errorf("unable to retrieve SSH key '%s': %w", idOrName, err)
	}
	if err != nil || sshKey == nil {
		sshKey, err = c.GetSSHKeyByName(idOrName)
		if err != nil {
			return nil, fmt.Errorf("unable to retrieve SSH key by name '%s': %w", idOrName, err)
		}
		if sshKey == nil {
			// Try other context: by UID then by name
			otherContext := "tenant"
			if resCtx == "tenant" {
				otherContext = "project"
			}
			c = getV1ClientWithResourceContext(m, otherContext)
			sshKey, err = c.GetSSHKey(idOrName)
			if err != nil && !herr.IsNotFound(err) {
				return nil, fmt.Errorf("unable to retrieve SSH key '%s': %w", idOrName, err)
			}
			if err == nil && sshKey != nil {
				resCtx = otherContext
			} else {
				sshKey, err = c.GetSSHKeyByName(idOrName)
				if err != nil {
					return nil, fmt.Errorf("unable to retrieve SSH key by name '%s': %w", idOrName, err)
				}
				if sshKey == nil {
					return nil, fmt.Errorf("SSH key with id or name '%s' not found in either project or tenant context", idOrName)
				}
				resCtx = otherContext
			}
		}
	}

	if err := d.Set("name", sshKey.Metadata.Name); err != nil {
		return nil, err
	}
	if err := d.Set("context", resCtx); err != nil {
		return nil, err
	}
	// Always set ID to UID so import by name results in stable resource id
	d.SetId(sshKey.Metadata.UID)

	// Note: We don't set the 'ssh_key' field during import because it's marked as sensitive.
	// The user will need to provide the ssh_key value in their Terraform configuration after import.

	return c, nil
}
