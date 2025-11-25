package spectrocloud

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/spectrocloud/terraform-provider-spectrocloud/spectrocloud/schemas"
	"github.com/spectrocloud/terraform-provider-spectrocloud/types"

	"log"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/spectrocloud/palette-sdk-go/client"
)

// Added description to the resource application profile to avoid drift
func resourceApplicationProfile() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceApplicationProfileCreate,
		ReadContext:   resourceApplicationProfileRead,
		UpdateContext: resourceApplicationProfileUpdate,
		DeleteContext: resourceApplicationProfileDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceApplicationProfileImport,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(20 * time.Second),
			Update: schema.DefaultTimeout(20 * time.Second),
			Delete: schema.DefaultTimeout(20 * time.Second),
		},
		SchemaVersion: 3,
		StateUpgraders: []schema.StateUpgrader{
			{
				Type:    resourceApplicationProfileResourceV2().CoreConfigSchema().ImpliedType(),
				Upgrade: resourceApplicationProfileStateUpgradeV2,
				Version: 2,
			},
		},
		Description: "Provisions an Application Profile. App Profiles are templates created with preconfigured services. You can create as many profiles as required, with multiple tiers serving different functionalities per use case.",

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Description: "Name of the application profile",
				Required:    true,
			},
			"version": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "1.0.0", // default as in UI
				Description: "Version of the profile. Default value is 1.0.0.",
			},
			"context": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "project",
				ValidateFunc: validation.StringInSlice([]string{"", "project", "tenant", "system"}, false),
				Description: "Context of the profile. Allowed values are `project`, `cluster`, or `namespace`. " +
					"Default value is `project`." + PROJECT_NAME_NUANCE,
			},
			"tags": {
				Type:        schema.TypeSet,
				Optional:    true,
				Set:         schema.HashString,
				Description: "A list of tags to be applied to the application profile. Tags must be in the form of `key:value`.",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"description": {
				Type:        schema.TypeString,
				Description: "Description of the profile.",
				Optional:    true,
			},
			"cloud": {
				Type:        schema.TypeString,
				Default:     "all",
				Description: "The cloud provider the profile is eligible for. Default value is `all`.",
				Optional:    true,
			},
			"pack": schemas.AppPackSchema(),
		},
	}
}

func resourceApplicationProfileCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	ProfileContext := d.Get("context").(string)
	c := getV1ClientWithResourceContext(m, ProfileContext)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	applicationProfile, err := toApplicationProfileCreate(d)
	if err != nil {
		return diag.FromErr(err)
	}

	// Create

	uid, err := c.CreateApplicationProfile(applicationProfile)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(uid)
	resourceApplicationProfileRead(ctx, d, m)
	return diags
}

func resourceApplicationProfileRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	ProfileContext := d.Get("context").(string)
	c := getV1ClientWithResourceContext(m, ProfileContext)

	var diags diag.Diagnostics

	cp, err := c.GetApplicationProfile(d.Id())
	if err != nil {
		return handleReadError(d, err, diags)
	} else if cp == nil {
		// Deleted - Terraform will recreate it
		d.SetId("")
		return diags
	}

	tags := flattenTags(cp.Metadata.Labels)
	if tags != nil {
		if err := d.Set("tags", tags); err != nil {
			return diag.FromErr(err)
		}
	}

	tierDetails, d2, done2 := getAppTiersContent(c, d)
	if done2 {
		return d2
	}

	err = d.Set("name", cp.Metadata.Name)
	if err != nil {
		return diag.FromErr(err)
	}

	// Get description from annotations - always set it (even if empty) to avoid drift
	description := ""
	if cp.Metadata.Annotations != nil {
		if desc, found := cp.Metadata.Annotations["description"]; found {
			description = desc
		}
	}
	if err := d.Set("description", description); err != nil {
		return diag.FromErr(err)
	}

	// diagPacks is not used by flattenAppPacks for application profiles
	// It only uses the API response (tierDetails), so we pass an empty slice
	// Note: GetDiagPacks would fail here because pack is now TypeSet, not TypeList
	diagPacks := make([]*models.V1PackManifestEntity, 0)
	packs, err := flattenAppPacks(c, diagPacks, cp.Spec.Template.AppTiers, tierDetails, d, ctx)
	if err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("pack", packs); err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func getAppTiersContent(c *client.V1Client, d *schema.ResourceData) ([]*models.V1AppTier, diag.Diagnostics, bool) {
	tiersDetails, err := c.GetApplicationProfileTiers(d.Id())
	if err != nil {
		return nil, diag.FromErr(err), true
	}
	return tiersDetails, nil, false
}

func getValueInProperties(prop map[string]interface{}, key string) string {
	for k, v := range prop {
		if k == key {
			return v.(string)
		}
	}
	return ""
}

// getPackList converts pack from TypeSet or TypeList to []interface{}
func getPackList(d *schema.ResourceData) []interface{} {
	packRaw := d.Get("pack")
	if packSet, ok := packRaw.(*schema.Set); ok {
		return packSet.List()
	}
	if packList, ok := packRaw.([]interface{}); ok {
		return packList
	}
	return []interface{}{}
}

// getManifestList converts manifest from TypeSet or TypeList to []interface{}
func getManifestList(manifestRaw interface{}) []interface{} {
	if manifestSet, ok := manifestRaw.(*schema.Set); ok {
		return manifestSet.List()
	}
	if manifestList, ok := manifestRaw.([]interface{}); ok {
		return manifestList
	}
	return []interface{}{}
}
func flattenAppPacks(c *client.V1Client, _ []*models.V1PackManifestEntity, tiers []*models.V1AppTierRef, tierDet []*models.V1AppTier, d *schema.ResourceData, ctx context.Context) ([]interface{}, error) {
	// func flattenAppPacks(c *client.V1Client, diagPacks []*models.V1PackManifestEntity, tiers []*models.V1AppTierRef, tierDet []*models.V1AppTier, d *schema.ResourceData, ctx context.Context) ([]interface{}, error) {
	if tiers == nil {
		return make([]interface{}, 0), nil
	}

	// Build registry maps to track which packs use registry_name or registry_uid
	registryNameMap := buildPackRegistryNameMap(d)
	registryUIDMap := buildPackRegistryUIDMap(d)
	// Build tag map to track which packs have tag in user config
	tagMap := buildPackTagMap(d)

	ps := make([]interface{}, 0)
	for _, tier := range tierDet {
		if tier == nil || tier.Metadata == nil || tier.Spec == nil {
			continue
		}
		if tier.Metadata.Name == "" {
			continue // Skip tiers without a name (required field)
		}
		p := make(map[string]interface{})
		if tier.Metadata.UID != "" {
			p["uid"] = tier.Metadata.UID
		}

		// Get the registry UID from the API response
		registryUID := tier.Spec.RegistryUID
		if registryUID == "" && tier.Spec.Type != nil {
			registryUID = c.GetPackRegistry(tier.Metadata.UID, string(*tier.Spec.Type))
		}

		// Determine what the user originally provided in their config
		usesRegistryName := registryNameMap != nil && registryNameMap[tier.Metadata.Name]
		usesRegistryUID := registryUIDMap != nil && registryUIDMap[tier.Metadata.Name]

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

		p["name"] = tier.Metadata.Name
		usesTag := false
		if packList := getPackList(d); len(packList) > 0 {
			for _, packItem := range packList {
				if packMap, ok := packItem.(map[string]interface{}); ok {
					if packName, ok := packMap["name"].(string); ok && packName == tier.Metadata.Name {
						// Check if tag field exists (even if empty - indicates user may have provided it)
						if _, exists := packMap["tag"]; exists {
							usesTag = true
							break
						}
					}
				}
			}
		}

		// Also check tagMap (built from state with non-empty tags) as fallback
		if !usesTag {
			usesTag = tagMap != nil && tagMap[tier.Metadata.Name]
		}

		// If API returns a tag, user must have provided it (we send it during create/update)
		// This handles the case where state has empty tag but API has correct value
		if !usesTag && tier.Spec.Version != "" {
			usesTag = true
		}

		if usesTag {
			// User originally specified tag - use API value if available, otherwise preserve from config
			if tier.Spec.Version != "" {
				p["tag"] = tier.Spec.Version
			} else {
				// API didn't return tag - try to preserve from user's original config
				if packList := getPackList(d); len(packList) > 0 {
					for _, packItem := range packList {
						if packMap, ok := packItem.(map[string]interface{}); ok {
							if packName, ok := packMap["name"].(string); ok && packName == tier.Metadata.Name {
								if tagVal, exists := packMap["tag"]; exists && tagVal != nil {
									if tagStr, ok := tagVal.(string); ok && tagStr != "" {
										p["tag"] = tagStr
									}
								}
								break
							}
						}
					}
				}
			}
		}
		// else: User didn't specify tag, so don't set it in state (even if API returns it)

		// Set type from API if available, otherwise preserve from user config
		if tier.Spec.Type != nil {
			p["type"] = string(*tier.Spec.Type)
		} else {
			// API didn't return type - try to preserve from user's original config
			if packList := getPackList(d); len(packList) > 0 {
				for _, packItem := range packList {
					if packMap, ok := packItem.(map[string]interface{}); ok {
						if packName, ok := packMap["name"].(string); ok && packName == tier.Metadata.Name {
							if typeVal, exists := packMap["type"]; exists && typeVal != nil {
								if typeStr, ok := typeVal.(string); ok && typeStr != "" {
									p["type"] = typeStr
								}
							}
							break
						}
					}
				}
			}
		}
		if tier.Spec.SourceAppTierUID != "" {
			p["source_app_tier"] = tier.Spec.SourceAppTierUID
		}
		prop := make(map[string]string)
		if len(tier.Spec.Properties) > 0 {
			for _, pt := range tier.Spec.Properties {
				if pt.Value != "********" {
					prop[pt.Name] = pt.Value
				} else {
					if _, ok := d.GetOk("pack"); ok {
						packList := getPackList(d)
						// Match pack by name since TypeSet is unordered
						for _, packItem := range packList {
							if packMap, ok := packItem.(map[string]interface{}); ok {
								if packName, ok := packMap["name"].(string); ok && packName == tier.Metadata.Name {
									if ogProp, exists := packMap["properties"]; exists && ogProp != nil {
										if propMap, ok := ogProp.(map[string]interface{}); ok {
											prop[pt.Name] = getValueInProperties(propMap, pt.Name)
										}
									}
									break
								}
							}
						}
					}
				}

			}
		}
		// Always set properties to ensure consistent hashing
		// Try to preserve from user config first to avoid drift
		if len(prop) > 0 {
			// We have properties from API, use them
			p["properties"] = prop
		} else {
			// Properties is empty - try to preserve from user config/state to avoid drift
			preserved := false
			if packList := getPackList(d); len(packList) > 0 {
				for _, packItem := range packList {
					if packMap, ok := packItem.(map[string]interface{}); ok {
						if packName, ok := packMap["name"].(string); ok && packName == tier.Metadata.Name {
							if ogProp, exists := packMap["properties"]; exists && ogProp != nil {
								// Convert to map[string]string to ensure consistent type
								if propMap, ok := ogProp.(map[string]interface{}); ok {
									convertedProp := make(map[string]string)
									for k, v := range propMap {
										if strVal, ok := v.(string); ok {
											convertedProp[k] = strVal
										}
									}
									p["properties"] = convertedProp
									preserved = true
								} else if strMap, ok := ogProp.(map[string]string); ok {
									// Already the correct type
									p["properties"] = strMap
									preserved = true
								}
							}
							break
						}
					}
				}
			}
			// If not preserved from config, set empty map to ensure consistent hashing
			if !preserved {
				p["properties"] = make(map[string]string)
			}
		}
		if tier.Spec.Type != nil && string(*tier.Spec.Type) == "container" {
			if tier.Spec.Values != "" {
				p["values"] = tier.Spec.Values
			}
		}
		if tier.Spec.Type != nil && (*tier.Spec.Type == "helm" || *tier.Spec.Type == "manifest") {
			if len(tier.Spec.Manifests) > 0 {
				ma := make([]interface{}, 0)
				for _, m := range tier.Spec.Manifests {
					if m == nil {
						continue
					}
					mj := make(map[string]interface{})
					if m.Name != "" {
						mj["name"] = m.Name
					}
					if m.UID != "" {
						mj["uid"] = m.UID
					}
					cnt, err := c.GetApplicationProfileTierManifestContent(d.Id(), tier.Metadata.UID, m.UID)
					if err != nil {
						return nil, err
					}
					if cnt != "" {
						mj["content"] = cnt
					} else {
						mj["content"] = ""
					}
					// Only add manifest if it has a name (required field)
					if m.Name != "" {
						ma = append(ma, mj)
					}
				}
				if len(ma) > 0 {
					p["manifest"] = ma
				}
			}
		}
		// Set install_order from API response, or preserve from user config, or default to 0
		if tier.Spec.InstallOrder != 0 {
			p["install_order"] = int(tier.Spec.InstallOrder)
		} else {
			// Try to preserve from user config if API didn't return it
			if packList := getPackList(d); len(packList) > 0 {
				for _, packItem := range packList {
					if packMap, ok := packItem.(map[string]interface{}); ok {
						if packName, ok := packMap["name"].(string); ok && packName == tier.Metadata.Name {
							if installOrderVal, exists := packMap["install_order"]; exists && installOrderVal != nil {
								if installOrderInt, ok := installOrderVal.(int); ok {
									p["install_order"] = installOrderInt
								}
							}
							break
						}
					}
				}
			}
			// Default to 0 if not set
			if _, ok := p["install_order"]; !ok {
				p["install_order"] = 0
			}
		}
		// Only add pack if it has a name (required field) and all required fields are present
		if p["name"] != nil && p["name"].(string) != "" {
			// Ensure properties is always set (even if empty) for consistent hashing
			if _, ok := p["properties"]; !ok {
				p["properties"] = make(map[string]string)
			}
			// Ensure install_order is always set (even if 0) for consistent hashing
			if _, ok := p["install_order"]; !ok {
				p["install_order"] = 0
			}
			ps = append(ps, p)
		}
	}

	return ps, nil
}

func resourceApplicationProfileUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	ProfileContext := d.Get("context").(string)
	c := getV1ClientWithResourceContext(m, ProfileContext)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	if d.HasChanges("description") || d.HasChanges("name") || d.HasChanges("tags") || d.HasChanges("pack") {
		log.Printf("Updating packs")
		tiersCreate, tiersUpdateMap, tiersDeleteIds, err := toApplicationTiersUpdate(d, c)
		if err != nil {
			return diag.FromErr(err)
		}
		metadata, err := toApplicationProfilePatch(d)
		if err != nil {
			return diag.FromErr(err)
		}

		//ProfileContext := d.Get("context").(string)
		if err := c.CreateApplicationProfileTiers(d.Id(), tiersCreate); err != nil {
			return diag.FromErr(err)
		}
		for i, tier := range tiersUpdateMap {
			if err := c.UpdateApplicationProfileTiers(d.Id(), i, tier); err != nil {
				return diag.FromErr(err)
			}
		}
		if err := c.DeleteApplicationProfileTiers(d.Id(), tiersDeleteIds); err != nil {
			return diag.FromErr(err)
		}
		if err := c.PatchApplicationProfile(d.Id(), metadata); err != nil {
			return diag.FromErr(err)
		}
	}

	resourceApplicationProfileRead(ctx, d, m)

	return diags
}

func resourceApplicationProfileDelete(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	ProfileContext := d.Get("context").(string)
	c := getV1ClientWithResourceContext(m, ProfileContext)

	var diags diag.Diagnostics

	err := c.DeleteApplicationProfile(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func toApplicationProfileCreate(d *schema.ResourceData) (*models.V1AppProfileEntity, error) {
	cp := toApplicationProfileBasic(d)

	tiers := make([]*models.V1AppTierEntity, 0)
	for _, tier := range getPackList(d) {
		if t, e := toApplicationProfilePackCreate(tier); e != nil {
			return nil, e
		} else {
			tiers = append(tiers, t)
		}
	}
	cp.Spec.Template.AppTiers = tiers
	return cp, nil
}

func toApplicationProfileBasic(d *schema.ResourceData) *models.V1AppProfileEntity {
	description := ""
	if d.Get("description") != nil {
		description = d.Get("description").(string)
	}
	cp := &models.V1AppProfileEntity{
		Metadata: &models.V1ObjectMetaInputEntity{
			Name: d.Get("name").(string),
			Annotations: map[string]string{
				"description": description,
			},
			Labels: toTags(d),
		},
		Spec: &models.V1AppProfileEntitySpec{
			Template: &models.V1AppProfileTemplateEntity{
				AppTiers: toAppTiers(),
			},
			Version: d.Get("version").(string),
		},
	}
	return cp
}

func toAppTiers() []*models.V1AppTierEntity {
	ret := make([]*models.V1AppTierEntity, 0)
	return ret
}

func toApplicationProfilePackCreate(pSrc interface{}) (*models.V1AppTierEntity, error) {
	return toApplicationProfilePackCreateWithClient(pSrc, nil)
}

func toApplicationProfilePackCreateWithClient(pSrc interface{}, c *client.V1Client) (*models.V1AppTierEntity, error) {
	p, ok := pSrc.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid pack data: expected map[string]interface{}")
	}

	// Name is required field
	pName := ""
	if nameVal, ok := p["name"]; ok && nameVal != nil {
		if nameStr, ok := nameVal.(string); ok {
			pName = nameStr
		}
	}
	if pName == "" {
		return nil, fmt.Errorf("pack name is required but was not provided")
	}
	pVersion := ""
	if p["tag"] != nil {
		if tagStr, ok := p["tag"].(string); ok {
			pVersion = tagStr
		}
	}
	source_app_tier := ""
	if p["source_app_tier"] != nil {
		if sourceStr, ok := p["source_app_tier"].(string); ok {
			source_app_tier = sourceStr
		}
	}
	//pTag := p["tag"].(string)
	//pUID := p["uid"].(string)
	pRegistryUID := ""
	if p["registry_uid"] != nil {
		if uidStr, ok := p["registry_uid"].(string); ok {
			pRegistryUID = uidStr
		}
	}
	pRegistryName := ""
	if p["registry_name"] != nil {
		if nameStr, ok := p["registry_name"].(string); ok {
			pRegistryName = nameStr
		}
	}
	// Type handling: If user provided type, use it; otherwise leave as nil
	// Note: Type is optional in API, but if provided it must be a valid enum value
	var pType *models.V1AppTierType
	if p["type"] != nil {
		if typeStr, ok := p["type"].(string); ok && typeStr != "" {
			// Validate the type is one of the allowed values
			validTypes := map[string]bool{
				"container":         true,
				"helm":              true,
				"manifest":          true,
				"operator-instance": true,
			}
			if validTypes[typeStr] {
				tierType := models.V1AppTierType(typeStr)
				pType = &tierType
			} else {
				return nil, fmt.Errorf("pack %s: invalid type '%s'. Allowed values are: container, helm, manifest, operator-instance", pName, typeStr)
			}
		}
	}

	// Validate that both registry_uid and registry_name are not provided together
	if pRegistryUID != "" && pRegistryName != "" {
		return nil, fmt.Errorf("pack %s: only one of 'registry_uid' or 'registry_name' can be specified, not both", pName)
	}

	// If registry_name is provided and client is available, resolve it to registry_uid
	if pRegistryName != "" && pRegistryUID == "" && c != nil {
		// Need a type string for resolution - use "helm" as fallback if type not provided
		typeStrForResolution := "helm"
		if pType != nil {
			typeStrForResolution = string(*pType)
		}
		resolvedUID, err := resolveRegistryNameToUID(c, pRegistryName, typeStrForResolution)
		if err != nil {
			return nil, fmt.Errorf("pack %s: %w", pName, err)
		}
		pRegistryUID = resolvedUID
	}

	valuesStr := ""
	if p["values"] != nil {
		if valStr, ok := p["values"].(string); ok {
			valuesStr = strings.TrimSpace(valStr)
		}
	}

	installOrder := int32(0)
	if p["install_order"] != nil {
		if installOrderInt, ok := p["install_order"].(int); ok {
			installOrder = int32(installOrderInt)
		}
	}

	tier := &models.V1AppTierEntity{
		Name:             types.Ptr(pName),
		Version:          pVersion,
		SourceAppTierUID: source_app_tier,
		RegistryUID:      pRegistryUID,
		//UID:         pUID,
		Type:         pType, // Can be nil if not provided
		InstallOrder: installOrder,
		// UI strips a single newline, so we should do the same
		Values:     valuesStr,
		Properties: toPropertiesTier(p),
	}

	manifests := make([]*models.V1ManifestInputEntity, 0)
	if manifestRaw, ok := p["manifest"]; ok && manifestRaw != nil {
		manifestList := getManifestList(manifestRaw)
		if len(manifestList) > 0 {
			for _, manifest := range manifestList {
				if m, ok := manifest.(map[string]interface{}); ok {
					content := ""
					name := ""
					if contentVal, ok := m["content"]; ok && contentVal != nil {
						if contentStr, ok := contentVal.(string); ok {
							content = strings.TrimSpace(contentStr)
						}
					}
					if nameVal, ok := m["name"]; ok && nameVal != nil {
						if nameStr, ok := nameVal.(string); ok {
							name = nameStr
						}
					}
					if name != "" {
						manifests = append(manifests, &models.V1ManifestInputEntity{
							Content: content,
							Name:    name,
						})
					}
				}
			}
		}
	}
	tier.Manifests = manifests

	return tier, nil
}

// get update create delete separately based on previous version.
func toApplicationTiersUpdate(d *schema.ResourceData, c *client.V1Client) ([]*models.V1AppTierEntity, map[string]*models.V1AppTierUpdateEntity, []string, error) {
	previousTiers, err := c.GetApplicationProfileTiers(d.Id())
	if err != nil {
		return nil, nil, nil, err
	}

	previousTiersMap := map[string]*models.V1AppTier{}
	for _, tier := range previousTiers {
		previousTiersMap[tier.Metadata.Name] = tier
	}

	var createTiers []*models.V1AppTierEntity
	updateTiersMap := map[string]*models.V1AppTierUpdateEntity{}
	updateTiersMapId := map[string]*models.V1AppTierUpdateEntity{}
	var deleteTiers []string

	createTiersMap := map[string]*models.V1AppTierEntity{}
	for _, tier := range getPackList(d) {
		tierMap, ok := tier.(map[string]interface{})
		if !ok {
			continue
		}
		tierName := ""
		if nameVal, ok := tierMap["name"]; ok && nameVal != nil {
			if nameStr, ok := nameVal.(string); ok {
				tierName = nameStr
			}
		}
		if tierName == "" {
			continue
		}
		if _, found := previousTiersMap[tierName]; found {
			t := toApplicationProfilePackUpdate(tier)
			updateTiersMap[t.Name] = t
		} else {
			if t, e := toApplicationProfilePackCreate(tier); e != nil {
				return nil, nil, nil, e
			} else {
				createTiers = append(createTiers, t)
				createTiersMap[*t.Name] = t
			}
		}
	}

	for _, tier := range previousTiers {
		_, create := createTiersMap[tier.Metadata.Name]
		_, update := updateTiersMap[tier.Metadata.Name]
		if !create && !update {
			deleteTiers = append(deleteTiers, tier.Metadata.UID)
		}
		if update {
			updateTiersMapId[tier.Metadata.UID] = updateTiersMap[tier.Metadata.Name]
		}
	}

	return createTiers, updateTiersMapId, deleteTiers, nil

}

func toApplicationProfilePatch(d *schema.ResourceData) (*models.V1AppProfileMetaEntity, error) {
	description := ""
	if d.Get("description") != nil {
		description = d.Get("description").(string)
	}

	metadata := &models.V1AppProfileMetaEntity{
		Metadata: &models.V1AppProfileMetaUpdateEntity{
			//TODO name change?: Name: d.Get("name").(string),
			Annotations: map[string]string{
				"description": description,
			},
			Labels: toTags(d),
		},
		/*TODO: check profile version: Spec: &models.V1ClusterProfileSpecEntity{
			Version: d.Get("version").(string),
		},*/
	}

	return metadata, nil
}

func toPropertiesTier(prop map[string]interface{}) []*models.V1AppTierPropertyEntity {
	pProperties := make([]*models.V1AppTierPropertyEntity, 0)
	if prop["properties"] != nil {
		if propsMap, ok := prop["properties"].(map[string]interface{}); ok {
			for k, val := range propsMap {
				if valStr, ok := val.(string); ok {
					prop := &models.V1AppTierPropertyEntity{
						Name:  k,
						Value: valStr,
					}
					pProperties = append(pProperties, prop)
				}
			}
		}
	}
	return pProperties
}

func toApplicationProfilePackUpdate(pSrc interface{}) *models.V1AppTierUpdateEntity {
	p, ok := pSrc.(map[string]interface{})
	if !ok {
		// Return empty entity if invalid data
		return &models.V1AppTierUpdateEntity{}
	}

	pName := ""
	if nameVal, ok := p["name"]; ok && nameVal != nil {
		if nameStr, ok := nameVal.(string); ok {
			pName = nameStr
		}
	}
	if pName == "" {
		// Return empty entity if name is missing
		return &models.V1AppTierUpdateEntity{}
	}
	pTag := ""
	if p["tag"] != nil {
		if tagStr, ok := p["tag"].(string); ok {
			pTag = tagStr
		}
	}
	//pUID := p["uid"].(string)

	manifests := make([]*models.V1ManifestRefUpdateEntity, 0)
	if manifestRaw, ok := p["manifest"]; ok && manifestRaw != nil {
		manifestList := getManifestList(manifestRaw)
		for _, manifest := range manifestList {
			if m, ok := manifest.(map[string]interface{}); ok {
				content := ""
				name := ""
				if contentVal, ok := m["content"]; ok && contentVal != nil {
					if contentStr, ok := contentVal.(string); ok {
						content = strings.TrimSpace(contentStr)
					}
				}
				if nameVal, ok := m["name"]; ok && nameVal != nil {
					if nameStr, ok := nameVal.(string); ok {
						name = nameStr
					}
				}
				if name != "" {
					manifests = append(manifests, &models.V1ManifestRefUpdateEntity{
						Content: content,
						Name:    types.Ptr(name),
						//UID:     getManifestUID(m["name"].(string), packs),
					})
				}
			}
		}
	}

	valuesStr := ""
	if p["values"] != nil {
		if valStr, ok := p["values"].(string); ok {
			valuesStr = strings.TrimSpace(valStr)
		}
	}

	pack := &models.V1AppTierUpdateEntity{

		Name:      pName,
		Version:   pTag,
		Manifests: manifests,
		//RegistryUID: pRegistryUID,
		// UI strips a single newline, so we should do the same
		Values:     valuesStr,
		Properties: toPropertiesTier(p),
	}

	return pack
}

func buildPackTagMap(d *schema.ResourceData) map[string]bool {
	tagMap := make(map[string]bool)
	if packs, ok := d.GetOk("pack"); ok {
		var packList []interface{}
		if packSet, ok := packs.(*schema.Set); ok {
			packList = packSet.List()
		} else if packListInterface, ok := packs.([]interface{}); ok {
			packList = packListInterface
		}
		for _, packInterface := range packList {
			if pack, ok := packInterface.(map[string]interface{}); ok {
				if packName, ok := pack["name"].(string); ok && packName != "" {
					if tag, ok := pack["tag"]; ok && tag != nil && tag.(string) != "" {
						tagMap[packName] = true
					}
				}
			}
		}
	}
	return tagMap
}

func resourceApplicationProfileResourceV2() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Description: "Name of the application profile",
				Required:    true,
			},
			"version": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "1.0.0",
				Description: "Version of the profile. Default value is 1.0.0.",
			},
			"context": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "project",
				ValidateFunc: validation.StringInSlice([]string{"", "project", "tenant", "system"}, false),
				Description: "Context of the profile. Allowed values are `project`, `cluster`, or `namespace`. " +
					"Default value is `project`." + PROJECT_NAME_NUANCE,
			},
			"tags": {
				Type:        schema.TypeSet,
				Optional:    true,
				Set:         schema.HashString,
				Description: "A list of tags to be applied to the application profile. Tags must be in the form of `key:value`.",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"description": {
				Type:        schema.TypeString,
				Description: "Description of the profile.",
				Optional:    true,
			},
			"cloud": {
				Type:        schema.TypeString,
				Default:     "all",
				Description: "The cloud provider the profile is eligible for. Default value is `all`.",
				Optional:    true,
			},
			"pack": {
				Type:        schema.TypeList,
				Required:    true,
				Description: "A list of packs to be applied to the application profile.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "The type of Pack. Allowed values are `container`, `helm`, `manifest`, or `operator-instance`.",
							// No default for V2 schema to match V3
						},
						"source_app_tier": {
							Type:        schema.TypeString,
							Description: "The unique id of the pack to be used as the source for the pack.",
							Optional:    true,
						},
						"registry_uid": {
							Type:        schema.TypeString,
							Description: "The unique id of the registry to be used for the pack. Either `registry_uid` or `registry_name` can be specified, but not both.",
							Optional:    true,
						},
						"registry_name": {
							Type:        schema.TypeString,
							Description: "The name of the registry to be used for the pack. This can be used instead of `registry_uid` for better readability. Either `registry_uid` or `registry_name` can be specified, but not both.",
							Optional:    true,
						},
						"uid": {
							Type:        schema.TypeString,
							Description: "The unique id of the pack. This is a computed field and is not required to be set.",
							Computed:    true,
							Optional:    true,
						},
						"name": {
							Type:        schema.TypeString,
							Description: "The name of the specified pack.",
							Required:    true,
						},
						"properties": {
							Type:        schema.TypeMap,
							Optional:    true,
							Description: "The various properties required by different database tiers eg: `databaseName` and `databaseVolumeSize` size for Redis etc.",
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"install_order": {
							Type:        schema.TypeInt,
							Description: "The installation priority order of the app profile. The order of priority goes from lowest number to highest number. For example, a value of `-3` would be installed before an app profile with a higher number value. No upper and lower limits exist, and you may specify positive and negative integers. The default value is `0`. ",
							Default:     0,
							Optional:    true,
						},
						"manifest": {
							Type:        schema.TypeList,
							Optional:    true,
							Description: "The manifest of the pack.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"uid": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"name": {
										Type:        schema.TypeString,
										Description: "The name of the manifest.",
										Required:    true,
									},
									"content": {
										Type:        schema.TypeString,
										Required:    true,
										Description: "The content of the manifest.",
										DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
											// UI strips the trailing newline on save
											return strings.TrimSpace(old) == strings.TrimSpace(new)
										},
									},
								},
							},
						},
						"tag": {
							Type:        schema.TypeString,
							Description: "The identifier or version to label the pack.",
							Optional:    true,
						},
						"values": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "The values to be used for the pack. This is a stringified JSON object.",
							DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
								// UI strips the trailing newline on save
								return strings.TrimSpace(old) == strings.TrimSpace(new)
							},
						},
					},
				},
			},
		},
	}
}

func resourceApplicationProfileStateUpgradeV2(ctx context.Context, rawState map[string]interface{}, meta interface{}) (map[string]interface{}, error) {
	log.Printf("[DEBUG] Upgrading application profile state from version 2 to 3")

	// Convert pack from TypeList to TypeSet
	// Note: We keep the data as a list in rawState and let Terraform's schema processing
	// convert it to TypeSet during normal resource loading. This avoids JSON serialization
	// issues with schema.Set objects that contain hash functions.
	if packRaw, exists := rawState["pack"]; exists {
		var packList []interface{}

		// Handle both TypeSet (if already upgraded) and TypeList (from old state)
		if packSet, ok := packRaw.(*schema.Set); ok {
			// Already a Set, convert to list for state upgrade
			packList = packSet.List()
			log.Printf("[DEBUG] pack is already a Set, converting to list with %d items", len(packList))
		} else if packListRaw, ok := packRaw.([]interface{}); ok {
			// It's a list (old state format)
			packList = packListRaw
			log.Printf("[DEBUG] Keeping pack as list during state upgrade with %d items", len(packList))
		} else {
			log.Printf("[DEBUG] pack is neither Set nor list (type: %T), skipping conversion", packRaw)
			return rawState, nil
		}

		// Convert nested manifest from TypeList to TypeSet within each pack
		for i, packItem := range packList {
			if packMap, ok := packItem.(map[string]interface{}); ok {
				if manifestRaw, exists := packMap["manifest"]; exists {
					// Handle manifest as either Set or List
					var manifestList []interface{}
					if manifestSet, ok := manifestRaw.(*schema.Set); ok {
						manifestList = manifestSet.List()
					} else if manifestListRaw, ok := manifestRaw.([]interface{}); ok {
						manifestList = manifestListRaw
					}
					if len(manifestList) > 0 {
						log.Printf("[DEBUG] Keeping manifest as list during state upgrade for pack %d with %d items", i, len(manifestList))
						// Keep the manifest data as-is (as a list)
						// Terraform will convert it to TypeSet when loading the resource using the schema
						packMap["manifest"] = manifestList
					}
				}
			}
		}

		// Keep the pack data as-is (as a list)
		// Terraform will convert it to TypeSet when loading the resource using the schema
		rawState["pack"] = packList

		log.Printf("[DEBUG] Successfully prepared pack for TypeSet conversion")
	} else {
		log.Printf("[DEBUG] No pack found in state, skipping conversion")
	}

	return rawState, nil
}
