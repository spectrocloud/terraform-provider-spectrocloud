package spectrocloud

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/spectrocloud/terraform-provider-spectrocloud/spectrocloud/schemas"
	"github.com/spectrocloud/terraform-provider-spectrocloud/util/ptr"

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

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(20 * time.Second),
			Update: schema.DefaultTimeout(20 * time.Second),
			Delete: schema.DefaultTimeout(20 * time.Second),
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
		return diag.FromErr(err)
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

	diagPacks, diagnostics, done := GetDiagPacks(d, err)
	if done {
		return diagnostics
	}
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

func flattenAppPacks(c *client.V1Client, diagPacks []*models.V1PackManifestEntity, tiers []*models.V1AppTierRef, tierDet []*models.V1AppTier, d *schema.ResourceData, ctx context.Context) ([]interface{}, error) {
	if tiers == nil {
		return make([]interface{}, 0), nil
	}

	ps := make([]interface{}, len(tiers))
	for i, tier := range tierDet {
		p := make(map[string]interface{})
		p["uid"] = tier.Metadata.UID
		if isRegistryUID(diagPacks, tier.Metadata.Name) {
			p["registry_uid"] = c.GetPackRegistry(tier.Metadata.UID, string(tier.Spec.Type))
		}
		p["name"] = tier.Metadata.Name
		//p["tag"] = tier.Tag
		p["type"] = tier.Spec.Type
		p["registry_uid"] = tier.Spec.RegistryUID
		p["source_app_tier"] = tier.Spec.SourceAppTierUID
		prop := make(map[string]string)
		if len(tier.Spec.Properties) > 0 {
			for _, pt := range tier.Spec.Properties {
				if pt.Value != "********" {
					prop[pt.Name] = pt.Value
				} else {
					ogProp := d.Get("pack").([]interface{})[i].(map[string]interface{})["properties"]
					prop[pt.Name] = getValueInProperties(ogProp.(map[string]interface{}), pt.Name)
				}

			}
		}
		p["properties"] = prop
		if tier.Spec.Type == "container" {
			p["values"] = tier.Spec.Values
		}
		if tier.Spec.Type == "helm" || tier.Spec.Type == "manifest" {
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
			}
		}
		ps[i] = p
	}

	return ps, nil
}

func resourceApplicationProfileUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	ProfileContext := d.Get("context").(string)
	c := getV1ClientWithResourceContext(m, ProfileContext)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	if d.HasChanges("name") || d.HasChanges("tags") || d.HasChanges("pack") {
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
	for _, tier := range d.Get("pack").([]interface{}) {
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
	pType := models.V1AppTierType(p["type"].(string))

	tier := &models.V1AppTierEntity{
		Name:             ptr.To(pName),
		Version:          pVersion,
		SourceAppTierUID: source_app_tier,
		RegistryUID:      pRegistryUID,
		//UID:         pUID,
		Type: pType,
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
	for _, tier := range d.Get("pack").([]interface{}) {
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
			Name:    ptr.To(m["name"].(string)),
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
