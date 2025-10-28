package spectrocloud

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/spectrocloud/palette-sdk-go/client"
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

	resourceContext, profileIdentifier, err := ParseClusterProfileImportID(d.Id())
	if err != nil {
		return nil, err
	}

	c := getV1ClientWithResourceContext(m, resourceContext)

	// Try to get by ID first (backward compatibility)
	profile, err := c.GetClusterProfile(profileIdentifier)

	// If not found and identifier doesn't look like a UID, try searching by name
	if (err != nil || profile == nil) && !looksLikeUID(profileIdentifier) {
		profile, err = getClusterProfileByName(c, profileIdentifier, resourceContext)
	}

	if err != nil {
		return nil, fmt.Errorf("unable to retrieve cluster profile: %s", err)
	}
	if profile == nil {
		return nil, fmt.Errorf("cluster profile '%s' not found in context '%s'", profileIdentifier, resourceContext)
	}

	// Use the profile directly (both methods return *models.V1ClusterProfile)
	cp := profile

	err = d.Set("name", cp.Metadata.Name)
	if err != nil {
		return nil, err
	}
	err = d.Set("context", cp.Metadata.Annotations["scope"])
	if err != nil {
		return nil, err
	}

	// Set the ID of the resource in the state. This ID is used to track the
	// resource and must be set in the state during the import.
	d.SetId(cp.Metadata.UID)

	return c, nil
}

// ParseClusterProfileImportID parses the import ID which can be in formats:
// - {uid}:{context} (e.g., "12345:project") - backward compatible
// - {name}:{context} (e.g., "my-profile:project") - new name-based import
func ParseClusterProfileImportID(importID string) (string, string, error) {
	parts := strings.Split(importID, ":")

	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid import ID format '%s': expected format is '{id_or_name}:{context}' (e.g., 'profile-id:project' or 'profile-name:tenant')", importID)
	}

	identifier, context := parts[0], parts[1]

	// Validate context
	validContexts := map[string]bool{
		"project": true,
		"tenant":  true,
		"system":  true,
	}

	if !validContexts[context] {
		return "", "", fmt.Errorf("invalid context '%s': must be one of 'project', 'tenant', or 'system'", context)
	}

	if identifier == "" {
		return "", "", fmt.Errorf("profile identifier (ID or name) cannot be empty")
	}

	return context, identifier, nil
}

// looksLikeUID checks if a string looks like a UID (typically contains hyphens and alphanumeric characters)
// This is a heuristic to determine if we should try name-based lookup
func looksLikeUID(s string) bool {
	// UIDs are typically longer and contain specific patterns
	// If it's a simple name without special UID characteristics, try name lookup
	if len(s) < 10 {
		return false
	}

	hyphenCount := strings.Count(s, "-")

	// UUID format: 36 characters with 4 hyphens (8-4-4-4-12 pattern)
	if len(s) == 36 && hyphenCount == 4 {
		return true
	}

	// MongoDB ObjectId style: 24 hex characters, no hyphens
	if len(s) == 24 && hyphenCount == 0 && isHexString(s) {
		return true
	}

	// Long UID with many hyphens (4+) - likely a generated UID
	// But check if segments look like words or random characters
	if hyphenCount >= 3 {
		segments := strings.Split(s, "-")
		var shortSegmentCount int
		var longSegmentCount int

		for _, seg := range segments {
			segLen := len(seg)
			if segLen <= 3 {
				// Very short segments (1-3 chars) are more UID-like
				shortSegmentCount++
			} else if segLen > 10 {
				// Very long segments are more name-like
				longSegmentCount++
			}
		}

		// If most segments are very short (1-3 chars), likely a UID
		// Examples: "abc-def-ghi-jkl", "a1b2-c3d4-e5f6"
		if shortSegmentCount >= len(segments)/2 {
			return true
		}

		// If we have any very long segments and 4+ hyphens, likely a descriptive name
		// Examples: "very-long-profile-name-v2" (4 hyphens but word-like)
		if hyphenCount >= 4 && longSegmentCount > 0 {
			return true // Actually many hyphens usually means UID
		}

		// For exactly 3 hyphens, be conservative - assume it could be a name
		// unless segments are very short
		if hyphenCount == 3 {
			return shortSegmentCount >= 2 // At least half should be short for UID
		}

		// 4+ hyphens - likely UID
		return hyphenCount >= 4
	}

	return false
}

// isHexString checks if a string contains only hexadecimal characters
func isHexString(s string) bool {
	for _, c := range s {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return false
		}
	}
	return true
}

// getClusterProfileByName searches for a cluster profile by name in the given context
func getClusterProfileByName(c *client.V1Client, name string, context string) (*models.V1ClusterProfile, error) {
	// List all cluster profile metadata
	profiles, err := c.GetClusterProfiles()
	if err != nil {
		return nil, fmt.Errorf("failed to list cluster profiles: %w", err)
	}

	// Find profiles matching the name (metadata only has Name and UID)
	var candidateUIDs []string
	for _, profile := range profiles {
		if profile.Metadata != nil && profile.Metadata.Name == name {
			candidateUIDs = append(candidateUIDs, profile.Metadata.UID)
		}
	}

	if len(candidateUIDs) == 0 {
		return nil, fmt.Errorf("no cluster profile found with name '%s'", name)
	}

	// Now fetch full profiles to check context
	var matchingProfiles []*models.V1ClusterProfile
	for _, uid := range candidateUIDs {
		fullProfile, err := c.GetClusterProfile(uid)
		if err != nil {
			// Skip profiles we can't fetch
			continue
		}
		if fullProfile != nil && fullProfile.Metadata != nil &&
			fullProfile.Metadata.Annotations != nil &&
			fullProfile.Metadata.Annotations["scope"] == context {
			matchingProfiles = append(matchingProfiles, fullProfile)
		}
	}

	if len(matchingProfiles) == 0 {
		return nil, fmt.Errorf("no cluster profile found with name '%s' in context '%s'. Found %d profile(s) with this name but in different context(s). UIDs: %s",
			name, context, len(candidateUIDs), strings.Join(candidateUIDs, ", "))
	}

	// If multiple profiles found in the same context, return an error
	if len(matchingProfiles) > 1 {
		var profileUIDs []string
		for _, p := range matchingProfiles {
			profileUIDs = append(profileUIDs, p.Metadata.UID)
		}
		return nil, fmt.Errorf("multiple cluster profiles found with name '%s' in context '%s'. Please use the profile UID instead. Found UIDs: %s",
			name, context, strings.Join(profileUIDs, ", "))
	}

	return matchingProfiles[0], nil
}
