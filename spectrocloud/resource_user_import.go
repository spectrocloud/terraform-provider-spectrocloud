package spectrocloud

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-sdk-go/client/herr"
)

// isUserNotFound returns true if the error indicates the user was not found,
// so the importer can fall back to name lookup. The user API returns
// Code:UserNotFound which herr.IsNotFound may not recognize.
func isUserNotFound(err error) bool {
	if err == nil {
		return false
	}
	if herr.IsNotFound(err) {
		return true
	}
	s := err.Error()
	return strings.Contains(s, "UserNotFound") || strings.Contains(s, "Specified user not found")
}

func resourceUserImport(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	c := getV1ClientWithResourceContext(m, "tenant")

	importID := d.Id()
	if importID == "" {
		return nil, fmt.Errorf("user import ID or name is required")
	}

	// Try by UID first
	user, err := c.GetUserByID(importID)
	if err != nil {
		if !isUserNotFound(err) {
			return nil, fmt.Errorf("unable to retrieve user '%s': %w", importID, err)
		}
		// Not found by UID — try by email (e.g. manimaran.audayappan@spc.com)
		user, err = c.GetUserByEmail(importID)
		if err != nil {
			if !isUserNotFound(err) {
				return nil, fmt.Errorf("unable to retrieve user by id or email '%s': %w", importID, err)
			}
			return nil, fmt.Errorf("user '%s' not found", importID)
		}
		if user == nil || user.Metadata == nil || user.Metadata.UID == "" {
			return nil, fmt.Errorf("user '%s' not found", importID)
		}
		d.SetId(user.Metadata.UID)
		if err := d.Set("email", user.Spec.EmailID); err != nil {
			return nil, err
		}
	} else if user != nil {
		// Found by UID
		if user.Metadata != nil && user.Metadata.UID != "" {
			d.SetId(user.Metadata.UID)
		} else {
			d.SetId(importID)
		}
		if err := d.Set("email", user.Spec.EmailID); err != nil {
			return nil, err
		}
	} else {
		return nil, fmt.Errorf("user '%s' not found", importID)
	}

	diags := resourceUserRead(ctx, d, m)
	if diags.HasError() {
		return nil, fmt.Errorf("could not read user for import: %v", diags)
	}
	return []*schema.ResourceData{d}, nil
}
