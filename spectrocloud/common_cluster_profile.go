package spectrocloud

import (
	"fmt"

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

// buildPackRegistryNameMapFromClusterProfiles creates a map of pack names that use registry_name
// from nested cluster_profile blocks on cluster resources.
func buildPackRegistryNameMapFromClusterProfiles(d *schema.ResourceData) map[string]bool {
	registryNameMap := make(map[string]bool)
	for _, profile := range normalizeInterfaceSliceFromListOrSet(d.Get("cluster_profile")) {
		profileMap := profile.(map[string]interface{})
		packs, ok := profileMap["pack"].([]interface{})
		if !ok {
			continue
		}
		for _, packInterface := range packs {
			pack := packInterface.(map[string]interface{})
			packName := pack["name"].(string)
			if registryName, ok := pack["registry_name"]; ok && registryName != nil && registryName.(string) != "" {
				registryNameMap[packName] = true
			}
		}
	}
	return registryNameMap
}

// buildPackRegistryUIDMapFromClusterProfiles creates a map of pack names that use registry_uid
// from nested cluster_profile blocks on cluster resources.
func buildPackRegistryUIDMapFromClusterProfiles(d *schema.ResourceData) map[string]bool {
	registryUIDMap := make(map[string]bool)
	for _, profile := range normalizeInterfaceSliceFromListOrSet(d.Get("cluster_profile")) {
		profileMap := profile.(map[string]interface{})
		packs, ok := profileMap["pack"].([]interface{})
		if !ok {
			continue
		}
		for _, packInterface := range packs {
			pack := packInterface.(map[string]interface{})
			packName := pack["name"].(string)
			if registryUID, ok := pack["registry_uid"]; ok && registryUID != nil && registryUID.(string) != "" {
				registryUIDMap[packName] = true
			}
		}
	}
	return registryUIDMap
}

// resolveProfileVariableValue picks the value to store in state after a read. When the variable is
// marked sensitive in the cluster variables API (IsSensitive), Palette returns a masked value and
// the prior state/config value must be preserved to avoid drift.
func resolveProfileVariableValue(prior, apiValue string, isSensitive bool) string {
	if isSensitive {
		if prior != "" {
			return prior
		}
		return ""
	}
	if apiValue != "" {
		return apiValue
	}
	return prior
}

// priorClusterProfileVariable returns the variable value from cluster_profile or nested cluster_template state.
func priorClusterProfileVariable(d *schema.ResourceData, profileUID, varName string) string {
	if v := priorClusterProfileVariableFromProfiles(d.Get("cluster_profile"), profileUID, varName); v != "" {
		return v
	}
	if raw := d.Get("cluster_template"); raw != nil {
		if templates, ok := raw.([]interface{}); ok && len(templates) > 0 {
			if template, ok := templates[0].(map[string]interface{}); ok {
				if nested, ok := template["cluster_profile"]; ok && nested != nil {
					return priorClusterProfileVariableFromProfiles(nested, profileUID, varName)
				}
			}
		}
	}
	return ""
}

func priorClusterProfileVariableFromProfiles(profilesRaw interface{}, profileUID, varName string) string {
	for _, profile := range normalizeInterfaceSliceFromListOrSet(profilesRaw) {
		profileMap, ok := profile.(map[string]interface{})
		if !ok {
			continue
		}
		id, ok := profileMap["id"].(string)
		if !ok || id != profileUID {
			continue
		}
		vars, ok := profileMap["variables"].(map[string]interface{})
		if !ok || vars == nil {
			return ""
		}
		if v, ok := vars[varName]; ok && v != nil {
			return v.(string)
		}
		return ""
	}
	return ""
}

// profileVariablesMapFromAPI builds a variables map for state from the cluster variables API response.
func profileVariablesMapFromAPI(d *schema.ResourceData, profileUID string, apiVars []*models.V1SpectroClusterVariableResponse) map[string]interface{} {
	vars := make(map[string]interface{})
	for _, v := range apiVars {
		if v.Name == nil {
			continue
		}
		name := *v.Name
		prior := priorClusterProfileVariable(d, profileUID, name)
		resolved := resolveProfileVariableValue(prior, v.Value, v.IsSensitive)
		if resolved != "" {
			vars[name] = resolved
			continue
		}
		// Import / first read: sensitive values are masked in the API and cannot be returned in cleartext.
		// Still record the variable name so import populates state; set the real value in Terraform config.
		if v.IsSensitive && prior == "" && v.Value != "" {
			vars[name] = ""
		}
	}
	return vars
}

// clusterProfileHasVariablesInConfig reports whether variables are declared for a profile in config.
func clusterProfileHasVariablesInConfig(d *schema.ResourceData, profileUID string) bool {
	for _, profile := range normalizeInterfaceSliceFromListOrSet(d.Get("cluster_profile")) {
		profileMap := profile.(map[string]interface{})
		id, ok := profileMap["id"].(string)
		if !ok || id != profileUID {
			continue
		}
		_, ok = profileMap["variables"]
		return ok
	}
	return false
}

// getClusterProfilePacksFromConfig returns pack blocks configured for a profile UID.
func getClusterProfilePacksFromConfig(d *schema.ResourceData, profileUID string) []interface{} {
	for _, profile := range normalizeInterfaceSliceFromListOrSet(d.Get("cluster_profile")) {
		profileMap := profile.(map[string]interface{})
		id, ok := profileMap["id"].(string)
		if !ok || id != profileUID {
			continue
		}
		if packs, ok := profileMap["pack"].([]interface{}); ok {
			return packs
		}
		return nil
	}
	return nil
}

func findConfigPackByName(configPacks []interface{}, name string) map[string]interface{} {
	for _, packRaw := range configPacks {
		pack, ok := packRaw.(map[string]interface{})
		if !ok {
			continue
		}
		if packName, ok := pack["name"].(string); ok && packName == name {
			return pack
		}
	}
	return nil
}

// alignPackStateWithConfig limits flattened pack state to fields the user declared in config.
// API read populates computed attributes (uid, type); omitting them from state when absent in
// config keeps cluster_profile TypeSet hashes aligned and avoids perpetual drift.
func alignPackStateWithConfig(pack map[string]interface{}, configPack map[string]interface{}) {
	if configPack == nil {
		delete(pack, "uid")
		delete(pack, "type")
		delete(pack, "registry_uid")
		delete(pack, "registry_name")
		delete(pack, "manifest")
		return
	}

	for _, field := range []string{"uid", "type", "registry_uid", "registry_name"} {
		if v, ok := configPack[field]; !ok || v == nil || v == "" {
			delete(pack, field)
		} else {
			pack[field] = v
		}
	}

	if _, ok := configPack["manifest"]; !ok {
		delete(pack, "manifest")
	}
}

// alignClusterProfilesStateWithConfig aligns refreshed profile state with config field presence.
func alignClusterProfilesStateWithConfig(d *schema.ResourceData, clusterProfiles []interface{}) {
	for i := range clusterProfiles {
		p := clusterProfiles[i].(map[string]interface{})
		uid, ok := p["id"].(string)
		if !ok || uid == "" {
			continue
		}

		if !clusterProfileHasVariablesInConfig(d, uid) {
			delete(p, "variables")
		}

		packs, ok := p["pack"].([]interface{})
		if !ok {
			continue
		}
		configPacks := getClusterProfilePacksFromConfig(d, uid)
		for j := range packs {
			pack, ok := packs[j].(map[string]interface{})
			if !ok {
				continue
			}
			name, _ := pack["name"].(string)
			alignPackStateWithConfig(pack, findConfigPackByName(configPacks, name))
			packs[j] = pack
		}
		p["pack"] = packs
	}
}

// clusterProfileHasPacksInConfig reports whether the given profile UID has pack blocks in config.
func clusterProfileHasPacksInConfig(d *schema.ResourceData, profileUID string) bool {
	for _, profile := range normalizeInterfaceSliceFromListOrSet(d.Get("cluster_profile")) {
		profileMap := profile.(map[string]interface{})
		id, ok := profileMap["id"].(string)
		if !ok || id != profileUID {
			continue
		}
		packs, ok := profileMap["pack"].([]interface{})
		return ok && len(packs) > 0
	}
	return false
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
