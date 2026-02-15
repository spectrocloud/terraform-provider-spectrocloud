package spectrocloud

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/spectrocloud/palette-sdk-go/client"
)

func flattenPacksWithRegistryMaps(c *client.V1Client, diagPacks []*models.V1PackManifestEntity, packs []*models.V1PackRef, manifestContent map[string]map[string]string, registryNameMap map[string]bool, registryUIDMap map[string]bool) ([]interface{}, error) {
	if packs == nil {
		return make([]interface{}, 0), nil
	}

	ps := make([]interface{}, len(packs))
	for i, pack := range packs {
		p := make(map[string]interface{})

		p["uid"] = pack.PackUID

		// Get the registry UID from the API response
		registryUID := c.GetPackRegistry(pack.PackUID, pack.Type)

		// Determine what the user originally provided in their config
		usesRegistryName := registryNameMap != nil && registryNameMap[*pack.Name]
		usesRegistryUID := registryUIDMap != nil && registryUIDMap[*pack.Name]

		if usesRegistryName {
			// User originally specified registry_name, resolve UID back to name
			if registryUID != "" {
				registryName, err := resolveRegistryUIDToName(c, registryUID)
				if err == nil && registryName != "" {
					p["registry_name"] = registryName
					// Do NOT set registry_uid - user didn't provide it
				} else {
					// Fallback to UID if name resolution fails
					p["registry_uid"] = registryUID
				}
			}
		} else if usesRegistryUID {
			// User originally specified registry_uid, set registry_uid
			if registryUID != "" {
				p["registry_uid"] = registryUID
			}
			// Do NOT set registry_name - user didn't provide it
		}
		// else: User didn't specify either registry_uid or registry_name
		// (they probably used uid directly), so don't set either in state

		p["name"] = *pack.Name
		p["tag"] = pack.Tag
		p["values"] = pack.Values
		p["type"] = pack.Type

		if _, ok := manifestContent[pack.PackUID]; ok {
			ma := make([]interface{}, len(pack.Manifests))
			for j, m := range pack.Manifests {
				mj := make(map[string]interface{})
				mj["name"] = m.Name
				mj["uid"] = m.UID
				mj["content"] = manifestContent[pack.PackUID][m.Name]

				ma[j] = mj
			}

			p["manifest"] = ma
		}
		ps[i] = p
	}

	return ps, nil
}

// buildPackRegistryNameMap creates a map indicating which packs use registry_name
// by directly checking the resource data
func buildPackRegistryNameMap(d *schema.ResourceData) map[string]bool {
	registryNameMap := make(map[string]bool)
	if packs, ok := d.GetOk("pack"); ok {
		for _, packInterface := range packs.([]interface{}) {
			pack := packInterface.(map[string]interface{})
			packName := pack["name"].(string)
			if registryName, ok := pack["registry_name"]; ok && registryName != nil && registryName.(string) != "" {
				registryNameMap[packName] = true
			}
		}
	}
	return registryNameMap
}

// buildPackRegistryUIDMap creates a map indicating which packs use registry_uid
// by directly checking the resource data
func buildPackRegistryUIDMap(d *schema.ResourceData) map[string]bool {
	registryUIDMap := make(map[string]bool)
	if packs, ok := d.GetOk("pack"); ok {
		for _, packInterface := range packs.([]interface{}) {
			pack := packInterface.(map[string]interface{})
			packName := pack["name"].(string)
			if registryUID, ok := pack["registry_uid"]; ok && registryUID != nil && registryUID.(string) != "" {
				registryUIDMap[packName] = true
			}
		}
	}
	return registryUIDMap
}

// resolveRegistryNameToUID resolves a registry name to its UID
func resolveRegistryNameToUID(c *client.V1Client, registryName string, registryType string) (string, error) {
	if registryName == "" {
		return "", nil
	}
	switch registryType {
	case "oci":
		registry, err := c.GetOciRegistryByName(registryName)
		if err != nil {
			return "", err
		}
		return registry.Metadata.UID, nil
	case "helm":
		registry, err := c.GetHelmRegistryByName(registryName)
		if err != nil {
			return "", err
		}
		return registry.Metadata.UID, nil
	case "spectro":
		registry, err := c.GetPackRegistryByName(registryName)
		if err != nil {
			return "", err
		}
		return registry.Metadata.UID, nil
	default:
		if registryType != "manifest" {
			registry, err := c.GetPackRegistryCommonByName(registryName)
			if err != nil {
				return "", err
			}
			return registry.UID, nil
		}
	}
	return "", nil
}

// resolveRegistryUIDToName resolves a registry UID to its name
func resolveRegistryUIDToName(c *client.V1Client, registryUID string) (string, error) {
	if registryUID == "" {
		return "", nil
	}
	registries, err := c.SearchPackRegistryCommon()
	if err != nil {
		return "", fmt.Errorf("failed to search registries: %w", err)
	}
	for _, registry := range registries {
		if registry.UID == registryUID {
			return registry.Name, nil
		}
	}
	return "", fmt.Errorf("registry with UID '%s' not found", registryUID)
}

// isRegistryPrivate returns true if the registry with the given UID is private.
// Private is determined from type-specific API fields so it matches backend classification:
//   - OCI: spec.registryMeta.isPrivate (or spec.isPrivate for ECR); for Basic OCI we use common metadata.
//   - Pack: spec.private
//   - Helm: spec.isPrivate
//
// Used to skip pack UID resolution for private registries (they do not expose packs via the same resolution API).
func isRegistryPrivate(c *client.V1Client, registryUID string) (bool, error) {
	if registryUID == "" {
		return false, nil
	}
	registries, err := c.SearchPackRegistryCommon()
	if err != nil {
		return false, fmt.Errorf("failed to search registries: %w", err)
	}
	var meta *models.V1RegistryMetadata
	for _, r := range registries {
		if r.UID == registryUID {
			meta = r
			break
		}
	}
	if meta == nil {
		return false, fmt.Errorf("registry with UID '%s' not found", registryUID)
	}
	kind := strings.ToLower(meta.Kind)
	switch kind {
	case "helm":
		reg, err := c.GetHelmRegistry(registryUID)
		if err != nil {
			return false, fmt.Errorf("failed to get helm registry %s: %w", registryUID, err)
		}
		if reg != nil && reg.Spec != nil {
			return reg.Spec.IsPrivate, nil
		}
		return meta.IsPrivate, nil
	case "pack":
		list, err := c.ListPackRegistries()
		if err != nil {
			return false, fmt.Errorf("failed to list pack registries: %w", err)
		}
		for _, s := range list {
			if s.Metadata != nil && s.Metadata.UID == registryUID {
				if s.Spec != nil {
					return s.Spec.Private, nil
				}
				return meta.IsPrivate, nil
			}
		}
		return false, fmt.Errorf("registry with UID '%s' not found", registryUID)
	case "oci":
		// OCI ECR: spec.isPrivate
		ecr, err := c.GetOciEcrRegistry(registryUID)
		if err == nil && ecr != nil && ecr.Spec != nil && ecr.Spec.IsPrivate != nil {
			return *ecr.Spec.IsPrivate, nil
		}
		// OCI Basic (Helm/Zarf/Pack provider): use common metadata (backend: spec.registryMeta.isPrivate)
		basic, err := c.GetOciBasicRegistry(registryUID)
		if err == nil && basic != nil {
			return meta.IsPrivate, nil
		}
		return meta.IsPrivate, nil
	default:
		return meta.IsPrivate, nil
	}
}

func getPacksContent(packs []*models.V1PackRef, c *client.V1Client, d *schema.ResourceData) (map[string]map[string]string, diag.Diagnostics, bool) {
	packManifests := make(map[string]map[string]string)
	for _, p := range packs {
		if len(p.Manifests) > 0 {
			content, err := c.GetClusterProfileManifestPack(d.Id(), *p.Name)
			if err != nil {
				return nil, diag.FromErr(err), true
			}

			if len(content) > 0 {
				c := make(map[string]string)
				for _, co := range content {
					c[co.Metadata.Name] = co.Spec.Published.Content
				}
				packManifests[p.PackUID] = c
			}
		}
	}
	return packManifests, nil, false
}
