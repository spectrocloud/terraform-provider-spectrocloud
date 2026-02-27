package spectrocloud

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceApplicationProfileImport(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	_, err := GetCommonApplicationProfile(d, m)
	if err != nil {
		return nil, err
	}

	diags := resourceApplicationProfileRead(ctx, d, m)
	if diags.HasError() {
		return nil, fmt.Errorf("could not read application profile for import: %v", diags)
	}

	return []*schema.ResourceData{d}, nil
}

// GetCommonApplicationProfile resolves an application profile from the import ID and populates
// the ResourceData with name, version, context, and cloud. The import ID supports three formats:
//
//	NAME_or_UID                        — context defaults to "project", version defaults to "1.0.0"
//	NAME_or_UID:CONTEXT                — version defaults to "1.0.0"
//	NAME_or_UID:CONTEXT:VERSION
func GetCommonApplicationProfile(d *schema.ResourceData, m interface{}) (interface{}, error) {
	importID := d.Id()
	if importID == "" {
		return nil, fmt.Errorf("application profile import ID is required")
	}

	parts := strings.Split(importID, ":")
	if len(parts) > 3 {
		return nil, fmt.Errorf("invalid import ID format: expected NAME_or_UID, NAME_or_UID:CONTEXT, or NAME_or_UID:CONTEXT:VERSION, got %q", importID)
	}

	idOrName := strings.TrimSpace(parts[0])
	resourceContext := "project"
	version := "1.0.0"

	if len(parts) >= 2 && strings.TrimSpace(parts[1]) != "" {
		resourceContext = strings.TrimSpace(parts[1])
	}
	if len(parts) == 3 && strings.TrimSpace(parts[2]) != "" {
		version = strings.TrimSpace(parts[2])
	}

	validContexts := map[string]bool{"project": true, "tenant": true, "system": true}
	if !validContexts[resourceContext] {
		return nil, fmt.Errorf("invalid context %q: must be one of project, tenant, or system", resourceContext)
	}

	c := getV1ClientWithResourceContext(m, resourceContext)

	// 1) Try treating the ID as a UID first (backward-compatible with existing imports)
	appProfile, err := c.GetApplicationProfile(idOrName)
	if err == nil && appProfile != nil {
		resolvedVersion := appProfile.Spec.Version
		// If user explicitly provided a version, validate it matches
		if len(parts) == 3 && resolvedVersion != version {
			return nil, fmt.Errorf("application profile version mismatch: requested %q but profile UID %q has version %q", version, idOrName, resolvedVersion)
		}
		if err := setApplicationProfileImportState(d, appProfile.Metadata.Name, resolvedVersion, resourceContext, appProfile.Metadata.UID); err != nil {
			return nil, err
		}
		return c, nil
	}

	// 2) Treat as NAME: look up by name + version using the SDK helper
	profile, profileUID, resolvedVersion, err := c.GetApplicationProfileByNameAndVersion(idOrName, version)
	if err != nil {
		return nil, fmt.Errorf("application profile %q with version %q not found in context %q: %w", idOrName, version, resourceContext, err)
	}
	if profile == nil {
		return nil, fmt.Errorf("application profile %q not found in context %q", idOrName, resourceContext)
	}

	if err := setApplicationProfileImportState(d, profile.Metadata.Name, resolvedVersion, resourceContext, profileUID); err != nil {
		return nil, err
	}
	return c, nil
}

func setApplicationProfileImportState(d *schema.ResourceData, name, version, resourceContext, uid string) error {
	if err := d.Set("name", name); err != nil {
		return err
	}
	if err := d.Set("version", version); err != nil {
		return err
	}
	if err := d.Set("context", resourceContext); err != nil {
		return err
	}
	if err := d.Set("cloud", "all"); err != nil {
		return err
	}
	d.SetId(uid)
	return nil
}
