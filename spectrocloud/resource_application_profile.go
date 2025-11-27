package spectrocloud

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/spectrocloud/terraform-provider-spectrocloud/spectrocloud/schemas"
	"github.com/spectrocloud/terraform-provider-spectrocloud/types"

	"log"
	"sort"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/spectrocloud/palette-sdk-go/client"
)

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

		Description: "Provisions an Application Profile. App Profiles are templates created with preconfigured services. You can create as many profiles as required, with multiple tiers serving different functionalities per use case.",

		SchemaVersion: 1,
		StateUpgraders: []schema.StateUpgrader{
			{
				Type:    resourceApplicationProfileResourceV0().CoreConfigSchema().ImpliedType(),
				Upgrade: resourceApplicationProfileStateUpgradeV0,
				Version: 0,
			},
		},

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
			"pack": appPackSchemaSet(),
		},
	}
}

// appPackSchemaSet returns the TypeSet version of the pack schema for application profiles
func appPackSchemaSet() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeSet,
		Required:    true,
		Set:         resourceAppPackHash,
		Description: "A set of packs to be applied to the application profile. The order of packs in the set does not matter.",
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"type": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: "The type of Pack. Allowed values are `container`, `helm`, `manifest`, or `operator-instance`.",
					Default:     "spectro",
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
	}
}

// resourceAppPackHash creates a hash for the pack resource
// This hash includes all fields except computed ones (uid) to ensure proper identification
func resourceAppPackHash(v interface{}) int {
	var buf strings.Builder
	m := v.(map[string]interface{})

	// Include name (required field, most important identifier)
	if v, ok := m["name"].(string); ok {
		buf.WriteString(fmt.Sprintf("%s-", v))
	} else {
		log.Printf("[WARN] Pack hash: name is missing or not a string")
	}

	// Include type (optional but important for differentiation)
	if v, ok := m["type"].(string); ok && v != "" {
		buf.WriteString(fmt.Sprintf("%s-", v))
	}

	// Include tag/version (optional but critical for versioning)
	if v, ok := m["tag"].(string); ok && v != "" {
		buf.WriteString(fmt.Sprintf("%s-", v))
	}

	// Include source_app_tier (optional but important for cloning)
	if v, ok := m["source_app_tier"].(string); ok && v != "" {
		buf.WriteString(fmt.Sprintf("%s-", v))
	}

	// Include registry_uid (optional)
	if v, ok := m["registry_uid"].(string); ok && v != "" {
		buf.WriteString(fmt.Sprintf("%s-", v))
	}

	// Include registry_name (optional)
	if v, ok := m["registry_name"].(string); ok && v != "" {
		buf.WriteString(fmt.Sprintf("%s-", v))
	}

	// Include install_order (optional but affects deployment order)
	if v, ok := m["install_order"].(int); ok {
		buf.WriteString(fmt.Sprintf("%d-", v))
	}

	// Include properties (optional but critical for configuration)
	if v, ok := m["properties"].(map[string]interface{}); ok && len(v) > 0 {
		// Sort keys for consistent hashing
		keys := make([]string, 0, len(v))
		for k := range v {
			keys = append(keys, k)
		}
		sort.Strings(keys) // Ensure deterministic ordering
		for _, k := range keys {
			buf.WriteString(fmt.Sprintf("%s=%v;", k, v[k]))
		}
	}

	// Include values (optional but critical for helm/container configuration)
	if v, ok := m["values"].(string); ok && v != "" {
		// Normalize whitespace for consistent hashing
		normalized := strings.TrimSpace(v)
		buf.WriteString(fmt.Sprintf("%s-", normalized))
	}

	// Include manifests (optional but critical for manifest/helm types)
	if v, ok := m["manifest"].([]interface{}); ok && len(v) > 0 {
		for _, manifestItem := range v {
			if manifest, ok := manifestItem.(map[string]interface{}); ok {
				if name, ok := manifest["name"].(string); ok {
					buf.WriteString(fmt.Sprintf("manifest:%s;", name))
				}
				if content, ok := manifest["content"].(string); ok {
					// Normalize whitespace for consistent hashing
					normalized := strings.TrimSpace(content)
					buf.WriteString(fmt.Sprintf("content:%s;", normalized))
				}
			}
		}
	}

	// Note: We DO NOT include 'uid' in the hash as it is a computed field
	// The hash should be based on user input only, not computed values

	hashStr := buf.String()
	hash := schema.HashString(hashStr)
	log.Printf("[DEBUG] Pack hash for '%v': hash=%d, hashString='%s'", m["name"], hash, hashStr)

	return hash
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

	packs, err := flattenAppPacks(c, cp.Spec.Template.AppTiers, tierDetails, d, ctx)
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

// findPackPropertiesByName searches through a pack TypeSet and returns the properties
// of the pack with the matching name. This is used for retrieving masked property values.
func findPackPropertiesByName(packSet *schema.Set, packName string) map[string]interface{} {
	for _, packInterface := range packSet.List() {
		pack := packInterface.(map[string]interface{})
		if pack["name"].(string) == packName {
			if props, ok := pack["properties"].(map[string]interface{}); ok {
				return props
			}
			return nil
		}
	}
	return nil
}

func flattenAppPacks(c *client.V1Client, tiers []*models.V1AppTierRef, tierDet []*models.V1AppTier, d *schema.ResourceData, ctx context.Context) ([]interface{}, error) {
	if tiers == nil {
		return make([]interface{}, 0), nil
	}

	// Build registry maps to track which packs use registry_name or registry_uid
	registryNameMap := buildAppPackRegistryNameMap(d)
	registryUIDMap := buildAppPackRegistryUIDMap(d)

	// Use tierDet length instead of tiers length to avoid mismatch
	ps := make([]interface{}, len(tierDet))
	for i, tier := range tierDet {
		log.Printf("[DEBUG] Flattening pack %d: name=%s, uid=%s, version=%s", i, tier.Metadata.Name, tier.Metadata.UID, tier.Spec.Version)
		p := make(map[string]interface{})
		p["uid"] = tier.Metadata.UID

		// Get the registry UID from the API response
		registryUID := tier.Spec.RegistryUID
		if registryUID == "" {
			registryUID = c.GetPackRegistry(tier.Metadata.UID, string(*tier.Spec.Type))
		}

		// Determine what the user originally provided in their config
		usesRegistryName := registryNameMap != nil && registryNameMap[tier.Metadata.Name]
		usesRegistryUID := registryUIDMap != nil && registryUIDMap[tier.Metadata.Name]

		if usesRegistryName {
			// User originally specified registry_name, resolve UID back to name
			if registryUID != "" {
				registryName, err := resolveAppRegistryUIDToName(c, registryUID)
				if err == nil && registryName != "" {
					p["registry_name"] = registryName
					p["registry_uid"] = "" // Explicitly set to empty
				} else {
					// Fallback to UID if name resolution fails
					p["registry_uid"] = registryUID
					p["registry_name"] = ""
				}
			} else {
				p["registry_name"] = ""
				p["registry_uid"] = ""
			}
		} else if usesRegistryUID {
			// User originally specified registry_uid, set registry_uid
			if registryUID != "" {
				p["registry_uid"] = registryUID
			} else {
				p["registry_uid"] = ""
			}
			p["registry_name"] = "" // Explicitly set to empty
		} else {
			// User didn't specify either, set both to empty for consistent hashing
			p["registry_name"] = ""
			p["registry_uid"] = ""
		}

		p["name"] = tier.Metadata.Name
		// Set tag (version) from API response
		if tier.Spec.Version != "" {
			p["tag"] = tier.Spec.Version
		} else {
			p["tag"] = ""
		}
		p["type"] = tier.Spec.Type
		// Set source_app_tier - empty string if not provided
		if tier.Spec.SourceAppTierUID != "" {
			p["source_app_tier"] = tier.Spec.SourceAppTierUID
		} else {
			p["source_app_tier"] = ""
		}
		// Set install_order - default to 0 if not specified
		p["install_order"] = int(tier.Spec.InstallOrder)
		prop := make(map[string]string)
		if len(tier.Spec.Properties) > 0 {
			for _, pt := range tier.Spec.Properties {
				if pt.Value != "********" {
					prop[pt.Name] = pt.Value
				} else {
					// For masked properties, find the original pack by name from the TypeSet
					if packSet, ok := d.GetOk("pack"); ok {
						originalPackProps := findPackPropertiesByName(packSet.(*schema.Set), tier.Metadata.Name)
						if originalPackProps != nil {
							prop[pt.Name] = getValueInProperties(originalPackProps, pt.Name)
						}
					}
				}

			}
		}
		p["properties"] = prop

		// Set values - empty string if not provided
		if tier.Spec.Type != nil && string(*tier.Spec.Type) == "container" {
			p["values"] = tier.Spec.Values
		} else {
			p["values"] = ""
		}

		// Set manifest - empty list if not provided
		if tier.Spec.Type != nil && (*tier.Spec.Type == "helm" || *tier.Spec.Type == "manifest") {
			if len(tier.Spec.Manifests) > 0 {
				ma := make([]interface{}, len(tier.Spec.Manifests))
				for j, m := range tier.Spec.Manifests {
					mj := make(map[string]interface{})
					mj["name"] = m.Name
					mj["uid"] = m.UID
					cnt, err := c.GetApplicationProfileTierManifestContent(d.Id(), tier.Metadata.UID, m.UID)
					if err != nil {
						return nil, err
					}
					if cnt != "" {
						mj["content"] = cnt
					} else {
						mj["content"] = ""
					}
					ma[j] = mj
				}
				p["manifest"] = ma
			} else {
				p["manifest"] = make([]interface{}, 0)
			}
		} else {
			p["manifest"] = make([]interface{}, 0)
		}
		ps[i] = p

		// Debug log the flattened pack
		log.Printf("[DEBUG] Flattened pack %d complete: %+v", i, p)
	}

	log.Printf("[DEBUG] Total flattened packs: %d", len(ps))
	return ps, nil
}

func resourceApplicationProfileUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	ProfileContext := d.Get("context").(string)
	c := getV1ClientWithResourceContext(m, ProfileContext)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	if d.HasChange("name") || d.HasChange("tags") || d.HasChange("pack") {
		log.Printf("Updating application profile")
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
	// Get pack as TypeSet and convert to list
	packSet := d.Get("pack").(*schema.Set)
	for _, tier := range packSet.List() {
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
	p := pSrc.(map[string]interface{})

	pName := p["name"].(string)
	pVersion := ""
	if p["tag"] != nil {
		pVersion = p["tag"].(string)
	}
	source_app_tier := p["source_app_tier"].(string)
	//pTag := p["tag"].(string)
	//pUID := p["uid"].(string)
	pRegistryUID := ""
	if p["registry_uid"] != nil {
		pRegistryUID = p["registry_uid"].(string)
	}
	pRegistryName := ""
	if p["registry_name"] != nil {
		pRegistryName = p["registry_name"].(string)
	}
	pType := models.V1AppTierType(p["type"].(string))

	// Validate that both registry_uid and registry_name are not provided together
	if pRegistryUID != "" && pRegistryName != "" {
		return nil, fmt.Errorf("pack %s: only one of 'registry_uid' or 'registry_name' can be specified, not both", pName)
	}

	// If registry_name is provided and client is available, resolve it to registry_uid
	if pRegistryName != "" && pRegistryUID == "" && c != nil {
		resolvedUID, err := resolveAppRegistryNameToUID(c, pRegistryName, p["type"].(string))
		if err != nil {
			return nil, fmt.Errorf("pack %s: %w", pName, err)
		}
		pRegistryUID = resolvedUID
	}

	tier := &models.V1AppTierEntity{
		Name:             types.Ptr(pName),
		Version:          pVersion,
		SourceAppTierUID: source_app_tier,
		RegistryUID:      pRegistryUID,
		//UID:         pUID,
		Type: &pType,
		// UI strips a single newline, so we should do the same
		Values:     strings.TrimSpace(p["values"].(string)),
		Properties: toPropertiesTier(p),
	}

	manifests := make([]*models.V1ManifestInputEntity, 0)
	if len(p["manifest"].([]interface{})) > 0 {
		for _, manifest := range p["manifest"].([]interface{}) {
			m := manifest.(map[string]interface{})
			manifests = append(manifests, &models.V1ManifestInputEntity{
				Content: strings.TrimSpace(m["content"].(string)),
				Name:    m["name"].(string),
			})
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
	// Get pack as TypeSet and convert to list
	packSet := d.Get("pack").(*schema.Set)
	for _, tier := range packSet.List() {
		if _, found := previousTiersMap[tier.(map[string]interface{})["name"].(string)]; found {
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
		for k, val := range prop["properties"].(map[string]interface{}) {
			prop := &models.V1AppTierPropertyEntity{
				Name:  k,
				Value: val.(string),
			}
			pProperties = append(pProperties, prop)
		}
	}
	return pProperties
}

func toApplicationProfilePackUpdate(pSrc interface{}) *models.V1AppTierUpdateEntity {
	p := pSrc.(map[string]interface{})

	pName := p["name"].(string)
	pTag := p["tag"].(string)
	//pUID := p["uid"].(string)

	manifests := make([]*models.V1ManifestRefUpdateEntity, 0)
	for _, manifest := range p["manifest"].([]interface{}) {
		m := manifest.(map[string]interface{})
		manifests = append(manifests, &models.V1ManifestRefUpdateEntity{
			Content: strings.TrimSpace(m["content"].(string)),
			Name:    types.Ptr(m["name"].(string)),
			//UID:     getManifestUID(m["name"].(string), packs),
		})
	}

	pack := &models.V1AppTierUpdateEntity{

		Name:      pName,
		Version:   pTag,
		Manifests: manifests,
		//RegistryUID: pRegistryUID,
		// UI strips a single newline, so we should do the same
		Values:     strings.TrimSpace(p["values"].(string)),
		Properties: toPropertiesTier(p),
	}

	return pack
}

// ============================================================================
// Internal Helper Functions for Application Profile Resource
// ============================================================================
// These functions are copied from shared files to make the application profile
// resource more self-contained and independent. They are prefixed with "App"
// to distinguish them from the shared versions used by other resources.

// buildAppPackRegistryNameMap creates a map indicating which app packs use registry_name
// by directly checking the resource data. This is specific to application profiles.
func buildAppPackRegistryNameMap(d *schema.ResourceData) map[string]bool {
	registryNameMap := make(map[string]bool)
	if packs, ok := d.GetOk("pack"); ok {
		packSet := packs.(*schema.Set)
		for _, packInterface := range packSet.List() {
			pack := packInterface.(map[string]interface{})
			packName := pack["name"].(string)
			if registryName, ok := pack["registry_name"]; ok && registryName != nil && registryName.(string) != "" {
				registryNameMap[packName] = true
			}
		}
	}
	return registryNameMap
}

// buildAppPackRegistryUIDMap creates a map indicating which app packs use registry_uid
// by directly checking the resource data. This is specific to application profiles.
func buildAppPackRegistryUIDMap(d *schema.ResourceData) map[string]bool {
	registryUIDMap := make(map[string]bool)
	if packs, ok := d.GetOk("pack"); ok {
		packSet := packs.(*schema.Set)
		for _, packInterface := range packSet.List() {
			pack := packInterface.(map[string]interface{})
			packName := pack["name"].(string)
			if registryUID, ok := pack["registry_uid"]; ok && registryUID != nil && registryUID.(string) != "" {
				registryUIDMap[packName] = true
			}
		}
	}
	return registryUIDMap
}

// resolveAppRegistryNameToUID resolves a registry name to its UID for application profiles.
// This function handles different registry types: oci, helm, spectro, container, and generic.
func resolveAppRegistryNameToUID(c *client.V1Client, registryName string, registryType string) (string, error) {
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
	case "container":
		// Container type uses pack registry common
		registry, err := c.GetPackRegistryCommonByName(registryName)
		if err != nil {
			return "", err
		}
		return registry.UID, nil
	default:
		if registryType != "manifest" && registryType != "operator-instance" {
			registry, err := c.GetPackRegistryCommonByName(registryName)
			if err != nil {
				return "", err
			}
			return registry.UID, nil
		}
	}
	return "", nil
}

// resolveAppRegistryUIDToName resolves a registry UID to its name for application profiles.
// Used when displaying registry information back to the user in Terraform state.
func resolveAppRegistryUIDToName(c *client.V1Client, registryUID string) (string, error) {
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

// getAppDiagPacks converts the pack configuration from resource data to app tier entities
// for validation and diagnostic purposes. This is specific to application profiles.
func getAppDiagPacks(d *schema.ResourceData) ([]*models.V1AppTierEntity, diag.Diagnostics, error) {
	diagPacks := make([]*models.V1AppTierEntity, 0)
	// Get pack as TypeSet and convert to list
	packSet := d.Get("pack").(*schema.Set)
	for _, pack := range packSet.List() {
		p, err := toApplicationProfilePackCreate(pack)
		if err != nil {
			return nil, diag.FromErr(err), err
		}
		diagPacks = append(diagPacks, p)
	}
	return diagPacks, nil, nil
}

// ============================================================================
// State Upgrade Functions
// ============================================================================

// resourceApplicationProfileResourceV0 returns the schema for version 0 of the resource
// This is used for state migration from v0 (TypeList) to v1 (TypeSet)
func resourceApplicationProfileResourceV0() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"version": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "1.0.0",
			},
			"context": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "project",
			},
			"tags": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"cloud": {
				Type:     schema.TypeString,
				Default:  "all",
				Optional: true,
			},
			"pack": schemas.AppPackSchema(), // This was TypeList in v0
		},
	}
}

// resourceApplicationProfileStateUpgradeV0 migrates state from version 0 to version 1
// This handles the conversion of the "pack" field from TypeList to TypeSet
func resourceApplicationProfileStateUpgradeV0(ctx context.Context, rawState map[string]interface{}, meta interface{}) (map[string]interface{}, error) {
	log.Printf("[DEBUG] Upgrading application profile state from version 0 to 1")

	// Convert pack from TypeList to TypeSet
	// Note: We keep the data as a list in rawState and let Terraform's schema processing
	// convert it to TypeSet during normal resource loading. This avoids JSON serialization
	// issues with schema.Set objects that contain hash functions.
	if packRaw, exists := rawState["pack"]; exists {
		if packList, ok := packRaw.([]interface{}); ok {
			log.Printf("[DEBUG] Keeping pack as list during state upgrade with %d items", len(packList))

			// Keep the pack data as-is (as a list)
			// Terraform will convert it to TypeSet when loading the resource using the schema
			rawState["pack"] = packList

			log.Printf("[DEBUG] Successfully prepared pack for TypeSet conversion")
		} else {
			log.Printf("[DEBUG] pack is not a list, skipping conversion")
		}
	} else {
		log.Printf("[DEBUG] No pack found in state, skipping conversion")
	}

	return rawState, nil
}
