package spectrocloud

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/spectrocloud/palette-sdk-go/client"
	"github.com/spectrocloud/palette-sdk-go/client/herr"
)

func resourceClusterProfileImport(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	_, err := GetCommonClusterProfile(d, m)
	if err != nil {
		return nil, err
	}

	diags := resourceClusterProfileRead(ctx, d, m)
	if diags.HasError() {
		return nil, fmt.Errorf("could not read cluster for import: %v", diags)
	}

	// Return the resource data. In most cases, this method is only used to
	// import one resource at a time, so you should return the resource data
	// in a slice with a single element.
	return []*schema.ResourceData{d}, nil
}

func GetCommonClusterProfile(d *schema.ResourceData, m interface{}) (*client.V1Client, error) {
	// Import ID format: id_or_name:context (e.g. "my-profile:project" or "uid-123:project")
	resourceContext, profileID, err := ParseResourceID(d)
	if err != nil {
		return nil, err
	}
	c := getV1ClientWithResourceContext(m, resourceContext)

	// Try by UID first, then by name (same pattern as EKS cluster import)
	profile, err := c.GetClusterProfile(profileID)
	if err == nil && profile != nil {
		return setClusterProfileImportState(d, c, profile)
	}
	if err != nil && !herr.IsNotFound(err) {
		return nil, fmt.Errorf("unable to retrieve cluster profile: %w", err)
	}

	// Resolve by name
	profiles, err := c.GetClusterProfiles()
	if err != nil {
		return nil, fmt.Errorf("failed to list cluster profiles: %w", err)
	}
	var match *models.V1ClusterProfileMetadata
	for _, p := range profiles {
		if p.Metadata != nil && p.Metadata.Name == profileID {
			match = p
			break
		}
	}
	if match == nil {
		return nil, fmt.Errorf("cluster profile with name '%s' not found in context %s", profileID, resourceContext)
	}
	profile, err = c.GetClusterProfile(match.Metadata.UID)
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve cluster profile: %w", err)
	}
	if profile == nil {
		return nil, fmt.Errorf("cluster profile '%s' not found", profileID)
	}
	// Verify scope matches (client may return profiles from other contexts)
	if profile.Metadata != nil && profile.Metadata.Annotations != nil {
		if profile.Metadata.Annotations["scope"] != resourceContext {
			return nil, fmt.Errorf("cluster profile with name '%s' not found in context %s", profileID, resourceContext)
		}
	}

	return setClusterProfileImportState(d, c, profile)
}

func setClusterProfileImportState(d *schema.ResourceData, c *client.V1Client, profile *models.V1ClusterProfile) (*client.V1Client, error) {
	if err := d.Set("name", profile.Metadata.Name); err != nil {
		return c, err
	}
	scope := ""
	if profile.Metadata != nil && profile.Metadata.Annotations != nil {
		scope = profile.Metadata.Annotations["scope"]
	}
	if err := d.Set("context", scope); err != nil {
		return c, err
	}
	d.SetId(profile.Metadata.UID)
	return c, nil
}
