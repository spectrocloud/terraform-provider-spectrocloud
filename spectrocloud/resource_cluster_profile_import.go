package spectrocloud

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/spectrocloud/palette-sdk-go/client"
)

// ParseClusterProfileImportID parses import ID in format:
//   - UID:CONTEXT or NAME:CONTEXT (2 parts)
//   - UID:CONTEXT:VERSION or NAME:CONTEXT:VERSION (3 parts, VERSION optional)
// Returns idOrName, context, version (version is "" if not provided).
func ParseClusterProfileImportID(importID string) (idOrName, resourceContext, version string, err error) {
	parts := strings.Split(importID, ":")
	if len(parts) < 2 || len(parts) > 3 {
		return "", "", "", fmt.Errorf("invalid cluster profile import ID format: expected UID_or_NAME:CONTEXT or UID_or_NAME:CONTEXT:VERSION, got %q", importID)
	}
	idOrName = strings.TrimSpace(parts[0])
	resourceContext = strings.TrimSpace(parts[1])
	if idOrName == "" || resourceContext == "" {
		return "", "", "", fmt.Errorf("invalid cluster profile import ID: id/name and context cannot be empty")
	}
	validContexts := map[string]bool{"project": true, "tenant": true, "system": true}
	if !validContexts[resourceContext] {
		return "", "", "", fmt.Errorf("invalid cluster profile import ID: context must be project, tenant, or system, got %q", resourceContext)
	}
	if len(parts) == 3 {
		version = strings.TrimSpace(parts[2])
	}
	return idOrName, resourceContext, version, nil
}

func resourceClusterProfileImport(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	_, err := GetCommonClusterProfile(d, m)
	if err != nil {
		return nil, err
	}

	diags := resourceClusterProfileRead(ctx, d, m)
	if diags.HasError() {
		return nil, fmt.Errorf("could not read cluster profile for import: %v", diags)
	}

	return []*schema.ResourceData{d}, nil
}

func GetCommonClusterProfile(d *schema.ResourceData, m interface{}) (*client.V1Client, error) {
	idOrName, resourceContext, requestedVersion, err := ParseClusterProfileImportID(d.Id())
	if err != nil {
		return nil, err
	}

	c := getV1ClientWithResourceContext(m, resourceContext)

	// 1) Try as UID first (works for UID:CONTEXT and UID:CONTEXT:VERSION)
	profile, err := c.GetClusterProfile(idOrName)
	if err == nil && profile != nil {
		scope := ""
		if profile.Metadata != nil && profile.Metadata.Annotations != nil {
			scope = profile.Metadata.Annotations["scope"]
		}
		if scope != resourceContext {
			return nil, fmt.Errorf("cluster profile with id %q not found in context %q (profile is in context %q)", idOrName, resourceContext, scope)
		}
		profileVersion := ""
		if profile.Spec != nil && profile.Spec.Published != nil {
			profileVersion = profile.Spec.Published.ProfileVersion
		}
		if requestedVersion != "" && profileVersion != requestedVersion {
			return nil, fmt.Errorf("cluster profile version mismatch: requested %q, profile has version %q", requestedVersion, profileVersion)
		}
		if err := setClusterProfileImportState(d, profile, resourceContext); err != nil {
			return nil, err
		}
		return c, nil
	}

	// 2) Treat as NAME: resolve by name (and optional version) from list
	profile, err = resolveClusterProfileByNameAndVersion(c, idOrName, resourceContext, requestedVersion)
	if err != nil {
		return nil, err
	}
	if err := setClusterProfileImportState(d, profile, resourceContext); err != nil {
		return nil, err
	}
	return c, nil
}

// resolveClusterProfileByNameAndVersion finds a profile by name in the given context.
// If requestedVersion is non-empty, returns that version or an error; otherwise returns latest version.
func resolveClusterProfileByNameAndVersion(c *client.V1Client, name, resourceContext, requestedVersion string) (*models.V1ClusterProfile, error) {
	profiles, err := c.GetClusterProfiles()
	if err != nil {
		return nil, fmt.Errorf("failed to list cluster profiles: %w", err)
	}
	if profiles == nil {
		return nil, fmt.Errorf("cluster profile with name %q not found in context %q", name, resourceContext)
	}

	var matches []*models.V1ClusterProfile
	for _, p := range profiles {
		if p == nil || p.Metadata == nil || p.Metadata.Name != name {
			continue
		}
		fullProfile, err := c.GetClusterProfile(p.Metadata.UID)
		if err != nil {
			continue
		}
		if fullProfile == nil || fullProfile.Metadata == nil || fullProfile.Metadata.Annotations == nil {
			continue
		}
		if fullProfile.Metadata.Annotations["scope"] != resourceContext {
			continue
		}
		matches = append(matches, fullProfile)
	}

	if len(matches) == 0 {
		return nil, fmt.Errorf("cluster profile with name %q not found in context %q", name, resourceContext)
	}

	if requestedVersion != "" {
		for _, p := range matches {
			v := getProfileVersion(p)
			if v == requestedVersion {
				return p, nil
			}
		}
		return nil, fmt.Errorf("cluster profile with name %q and version %q not found in context %q", name, requestedVersion, resourceContext)
	}

	// Latest: sort by version (semver) and take highest
	sort.Slice(matches, func(i, j int) bool {
		vi := getProfileVersion(matches[i])
		vj := getProfileVersion(matches[j])
		si, ei := semver.NewVersion(vi)
		sj, ej := semver.NewVersion(vj)
		if ei != nil && ej != nil {
			return vi < vj
		}
		if ei != nil {
			return true
		}
		if ej != nil {
			return false
		}
		return si.LessThan(sj)
	})
	return matches[len(matches)-1], nil
}

func getProfileVersion(p *models.V1ClusterProfile) string {
	if p != nil && p.Spec != nil && p.Spec.Published != nil {
		return p.Spec.Published.ProfileVersion
	}
	return "0.0.0"
}

func setClusterProfileImportState(d *schema.ResourceData, profile *models.V1ClusterProfile, resourceContext string) error {
	if err := d.Set("name", profile.Metadata.Name); err != nil {
		return err
	}
	if err := d.Set("context", resourceContext); err != nil {
		return err
	}
	if profile.Spec != nil && profile.Spec.Published != nil && profile.Spec.Published.ProfileVersion != "" {
		if err := d.Set("version", profile.Spec.Published.ProfileVersion); err != nil {
			return err
		}
	}
	d.SetId(profile.Metadata.UID)
	return nil
}
