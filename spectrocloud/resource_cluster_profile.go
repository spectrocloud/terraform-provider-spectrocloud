package spectrocloud

import (
	"context"
	"fmt"

	"github.com/spectrocloud/terraform-provider-spectrocloud/spectrocloud/schemas"
	"github.com/spectrocloud/terraform-provider-spectrocloud/types"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"log"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/spectrocloud/palette-sdk-go/client"
)

func resourceClusterProfile() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceClusterProfileCreate,
		ReadContext:   resourceClusterProfileRead,
		UpdateContext: resourceClusterProfileUpdate,
		DeleteContext: resourceClusterProfileDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceClusterProfileImport,
		},
		Description: "The Cluster Profile resource allows you to create and manage cluster profiles.",

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(20 * time.Second),
			Update: schema.DefaultTimeout(20 * time.Second),
			Delete: schema.DefaultTimeout(20 * time.Second),
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"version": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "1.0.0", // default as in UI
				Description: "Version of the cluster profile. Defaults to '1.0.0'.",
			},
			"context": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "project",
				ValidateFunc: validation.StringInSlice([]string{"", "project", "tenant", "system"}, false),
				Description: "The context of the cluster profile. Allowed values are `project` or `tenant`. " +
					"Default value is `project`. " + PROJECT_NAME_NUANCE,
			},
			"tags": {
				Type:     schema.TypeSet,
				Optional: true,
				Set:      schema.HashString,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Description: "A list of tags to be applied to the cluster. Tags must be in the form of `key:value`.",
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"cloud": {
				Type:     schema.TypeString,
				Default:  "all",
				Optional: true,
				// Removing validation to support custom clouds
				// ValidateFunc: validation.StringInSlice([]string{"all", "aws", "azure", "gcp", "vsphere", "openstack", "maas", "virtual", "baremetal", "eks", "aks", "edge", "edge-native", "generic", "gke"}, false),
				ForceNew: true,
				Description: "Specify the infrastructure provider the cluster profile is for. Only Palette supported infrastructure providers can be used. The supported cloud types are - `all, aws, azure, gcp, vsphere, openstack, maas, virtual, baremetal, eks, aks, edge-native, generic, and gke` or any custom cloud provider registered in Palette, e.g., `nutanix`." +
					"If the value is set to `all`, then the type must be set to `add-on`. Otherwise, the cluster profile may be incompatible with other providers. Default value is `all`.",
			},
			"type": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "add-on",
				ValidateFunc: validation.StringInSlice([]string{"add-on", "cluster", "infra", "system"}, false),
				Description:  "Specify the cluster profile type to use. Allowed values are `cluster`, `infra`, `add-on`, and `system`. These values map to the following User Interface (UI) labels. Use the value ' cluster ' for a **Full** cluster profile." + "For an Infrastructure cluster profile, use the value `infra`; for an Add-on cluster profile, use the value `add-on`." + "System cluster profiles can be specified using the value `system`. To learn more about cluster profiles, refer to the [Cluster Profile](https://docs.spectrocloud.com/cluster-profiles) documentation. Default value is `add-on`.",
				ForceNew:     true,
			},
			"profile_variables": schemas.ProfileVariables(),
			"pack":              schemas.PackSchema(),
		},
	}
}

func resourceClusterProfileCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	ProfileContext := d.Get("context").(string)
	c := getV1ClientWithResourceContext(m, ProfileContext)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	clusterProfile, err := toClusterProfileCreateWithResolution(d, c)
	if err != nil {
		return diag.FromErr(err)
	}

	// Create
	uid, err := c.CreateClusterProfile(clusterProfile)
	if err != nil {
		return diag.FromErr(err)
	}

	// And then publish
	if err = c.PublishClusterProfile(uid); err != nil {
		return diag.FromErr(err)
	}
	d.SetId(uid)
	resourceClusterProfileRead(ctx, d, m)
	return diags
}

func resourceClusterProfileRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	ProfileContext := d.Get("context").(string)
	c := getV1ClientWithResourceContext(m, ProfileContext)

	var diags diag.Diagnostics

	cp, err := c.GetClusterProfile(d.Id())
	if err != nil {
		return handleReadError(d, err, diags)
	} else if cp == nil {
		// Deleted - Terraform will recreate it
		d.SetId("")
		return diags
	}

	err = flattenClusterProfileCommon(d, cp)
	if err != nil {
		diag.FromErr(err)
	}

	tags := flattenTags(cp.Metadata.Labels)
	if tags != nil {
		if err := d.Set("tags", tags); err != nil {
			return diag.FromErr(err)
		}
	}

	packManifests, d2, done2 := getPacksContent(cp.Spec.Published.Packs, c, d)
	if done2 {
		return d2
	}

	err = d.Set("name", cp.Metadata.Name)
	if err != nil {
		return diag.FromErr(err)
	}

	// Profile variables
	profileVariables, err := c.GetProfileVariables(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	pv, err := flattenProfileVariables(d, profileVariables)
	if err != nil {
		return diag.FromErr(err)
	}
	err = d.Set("profile_variables", pv)
	if err != nil {
		return diag.FromErr(err)
	}

	diagPacks, diagnostics, done := GetDiagPacks(d, err)
	if done {
		return diagnostics
	}
	packs, err := flattenPacks(c, diagPacks, cp.Spec.Published.Packs, packManifests)
	if err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("pack", packs); err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func flattenClusterProfileCommon(d *schema.ResourceData, cp *models.V1ClusterProfile) error {
	// set cloud
	if err := d.Set("cloud", cp.Spec.Published.CloudType); err != nil {
		return err
	}
	// set type
	if err := d.Set("type", cp.Spec.Published.Type); err != nil {
		return err
	}
	// set version
	if err := d.Set("version", cp.Spec.Published.ProfileVersion); err != nil {
		return err
	}
	return nil
}

func resourceClusterProfileUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	ProfileContext := d.Get("context").(string)
	c := getV1ClientWithResourceContext(m, ProfileContext)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	if d.HasChanges("profile_variables") {
		pvs, err := toClusterProfileVariables(d)
		if err != nil {
			return diag.FromErr(err)
		}
		mVars := &models.V1Variables{
			Variables: pvs,
		}
		err = c.UpdateProfileVariables(mVars, d.Id())
		if err != nil {
			oldVariables, _ := d.GetChange("profile_variables")
			_ = d.Set("profile_variables", oldVariables)
			return diag.FromErr(err)
		}
	}

	if d.HasChanges("name") || d.HasChanges("tags") || d.HasChanges("pack") || d.HasChanges("description") {
		log.Printf("Updating packs")
		cp, err := c.GetClusterProfile(d.Id())
		if err != nil {
			return diag.FromErr(err)
		}
		cluster, err := toClusterProfileUpdateWithResolution(d, cp, c)
		if err != nil {
			return diag.FromErr(err)
		}
		metadata, err := toClusterProfilePatch(d)
		if err != nil {
			return diag.FromErr(err)
		}

		//ProfileContext := d.Get("context").(string)
		if err := c.UpdateClusterProfile(cluster); err != nil {
			return diag.FromErr(err)
		}
		if err := c.PatchClusterProfile(cluster, metadata); err != nil {
			return diag.FromErr(err)
		}
		if err := c.PublishClusterProfile(cluster.Metadata.UID); err != nil {
			return diag.FromErr(err)
		}
	}

	resourceClusterProfileRead(ctx, d, m)

	return diags
}

func resourceClusterProfileDelete(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	ProfileContext := d.Get("context").(string)
	c := getV1ClientWithResourceContext(m, ProfileContext)

	var diags diag.Diagnostics

	if err := c.DeleteClusterProfile(d.Id()); err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func toClusterProfileCreateWithResolution(d *schema.ResourceData, c *client.V1Client) (*models.V1ClusterProfileEntity, error) {
	cp := toClusterProfileBasic(d)

	packs := make([]*models.V1PackManifestEntity, 0)
	for _, pack := range d.Get("pack").([]interface{}) {
		if p, e := toClusterProfilePackCreateWithResolution(pack, c); e != nil {
			return nil, e
		} else {
			packs = append(packs, p)
		}
	}
	cp.Spec.Template.Packs = packs
	if profileVariable, err := toClusterProfileVariables(d); err == nil {
		cp.Spec.Variables = profileVariable
	} else {
		return cp, err
	}
	return cp, nil
}

func toClusterProfileBasic(d *schema.ResourceData) *models.V1ClusterProfileEntity {
	description := ""
	if d.Get("description") != nil {
		description = d.Get("description").(string)
	}
	cp := &models.V1ClusterProfileEntity{
		Metadata: &models.V1ObjectMeta{
			Name: d.Get("name").(string),
			UID:  d.Id(),
			Annotations: map[string]string{
				"description": description,
			},
			Labels: toTags(d),
		},
		Spec: &models.V1ClusterProfileEntitySpec{
			Template: &models.V1ClusterProfileTemplateDraft{
				CloudType: d.Get("cloud").(string),
				Type:      types.Ptr(models.V1ProfileType(d.Get("type").(string))),
			},
			Version: d.Get("version").(string),
		},
	}
	return cp
}

func toClusterProfilePackCreate(pSrc interface{}) (*models.V1PackManifestEntity, error) {
	p := pSrc.(map[string]interface{})

	pName := p["name"].(string)
	pTag := p["tag"].(string)
	pUID := p["uid"].(string)
	pRegistryUID := ""
	if p["registry_uid"] != nil {
		pRegistryUID = p["registry_uid"].(string)
	}
	pType := models.V1PackType(p["type"].(string))

	// Validate pack UID or resolution fields
	if err := schemas.ValidatePackUIDOrResolutionFields(p); err != nil {
		return nil, err
	}

	switch pType {
	case models.V1PackTypeSpectro:
		if pUID == "" {
			// UID not provided, validation already passed, so we have all resolution fields
			// This path should not be reached if validation is working correctly
			if pTag == "" || pRegistryUID == "" {
				return nil, fmt.Errorf("pack %s: internal error - validation should have caught missing resolution fields", pName)
			}
		}
	case models.V1PackTypeManifest:
		if pUID == "" {
			pUID = "spectro-manifest-pack"
		}
	}

	pack := &models.V1PackManifestEntity{
		Name:        types.Ptr(pName),
		Tag:         p["tag"].(string),
		RegistryUID: pRegistryUID,
		UID:         pUID,
		Type:        &pType,
		// UI strips a single newline, so we should do the same
		Values: strings.TrimSpace(p["values"].(string)),
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
	pack.Manifests = manifests

	return pack, nil
}

func toClusterProfilePackCreateWithResolution(pSrc interface{}, c *client.V1Client) (*models.V1PackManifestEntity, error) {
	p := pSrc.(map[string]interface{})

	pName := p["name"].(string)
	pTag := p["tag"].(string)
	pUID := p["uid"].(string)
	pRegistryUID := ""
	if p["registry_uid"] != nil {
		pRegistryUID = p["registry_uid"].(string)
	}
	pType := models.V1PackType(p["type"].(string))

	// Validate pack UID or resolution fields
	if err := schemas.ValidatePackUIDOrResolutionFields(p); err != nil {
		return nil, err
	}

	switch pType {
	case models.V1PackTypeSpectro:
		if pUID == "" {
			// UID not provided, validation already passed, so we have all resolution fields
			// Resolve the pack UID
			resolvedUID, err := resolvePackUID(c, pName, pTag, pRegistryUID)
			if err != nil {
				return nil, fmt.Errorf("failed to resolve pack UID for pack %s: %w", pName, err)
			}
			pUID = resolvedUID
		}
	case models.V1PackTypeHelm:
		if pUID == "" {
			// UID not provided, validation already passed, so we have all resolution fields
			// Resolve the pack UID
			resolvedUID, err := resolvePackUID(c, pName, pTag, pRegistryUID)
			if err != nil {
				return nil, fmt.Errorf("failed to resolve pack UID for pack %s: %w", pName, err)
			}
			pUID = resolvedUID
		}
	case models.V1PackTypeManifest:
		if pUID == "" {
			pUID = "spectro-manifest-pack"
		}
	}

	pack := &models.V1PackManifestEntity{
		Name:        types.Ptr(pName),
		Tag:         p["tag"].(string),
		RegistryUID: pRegistryUID,
		UID:         pUID,
		Type:        &pType,
		// UI strips a single newline, so we should do the same
		Values: strings.TrimSpace(p["values"].(string)),
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
	pack.Manifests = manifests

	return pack, nil
}

func toClusterProfileUpdateWithResolution(d *schema.ResourceData, cluster *models.V1ClusterProfile, c *client.V1Client) (*models.V1ClusterProfileUpdateEntity, error) {
	cp := &models.V1ClusterProfileUpdateEntity{
		Metadata: &models.V1ObjectMeta{
			Name: d.Get("name").(string),
			UID:  d.Id(),
		},
		Spec: &models.V1ClusterProfileUpdateEntitySpec{
			Template: &models.V1ClusterProfileTemplateUpdate{
				Type: types.Ptr(models.V1ProfileType(d.Get("type").(string))),
			},
			Version: d.Get("version").(string),
		},
	}
	packs := make([]*models.V1PackManifestUpdateEntity, 0)
	for _, pack := range d.Get("pack").([]interface{}) {
		if p, e := toClusterProfilePackUpdateWithResolution(pack, cluster.Spec.Published.Packs, c); e != nil {
			return nil, e
		} else {
			packs = append(packs, p)
		}
	}
	cp.Spec.Template.Packs = packs

	return cp, nil
}

func toClusterProfilePatch(d *schema.ResourceData) (*models.V1ProfileMetaEntity, error) {
	description := ""
	if d.Get("description") != nil {
		description = d.Get("description").(string)
	}
	metadata := &models.V1ProfileMetaEntity{
		Metadata: &models.V1ObjectMetaInputEntity{
			Name: d.Get("name").(string),
			Annotations: map[string]string{
				"description": description,
			},
			Labels: toTags(d),
		},
		Spec: &models.V1ClusterProfileSpecEntity{
			Version: d.Get("version").(string),
		},
	}

	return metadata, nil
}

func toClusterProfilePackUpdateWithResolution(pSrc interface{}, packs []*models.V1PackRef, c *client.V1Client) (*models.V1PackManifestUpdateEntity, error) {
	p := pSrc.(map[string]interface{})

	pName := p["name"].(string)
	pTag := p["tag"].(string)
	pUID := p["uid"].(string)

	pRegistryUID := ""
	if p["registry_uid"] != nil {
		pRegistryUID = p["registry_uid"].(string)
	}
	pType := models.V1PackType(p["type"].(string))

	// Validate pack UID or resolution fields
	if err := schemas.ValidatePackUIDOrResolutionFields(p); err != nil {
		return nil, err
	}

	switch pType {
	case models.V1PackTypeSpectro:
		if pUID == "" {
			// UID not provided, validation already passed, so we have all resolution fields
			// Resolve the pack UID
			resolvedUID, err := resolvePackUID(c, pName, pTag, pRegistryUID)
			if err != nil {
				return nil, fmt.Errorf("failed to resolve pack UID for pack %s: %w", pName, err)
			}
			pUID = resolvedUID
		}
	case models.V1PackTypeHelm:
		if pUID == "" {
			// UID not provided, validation already passed, so we have all resolution fields
			// Resolve the pack UID
			resolvedUID, err := resolvePackUID(c, pName, pTag, pRegistryUID)
			if err != nil {
				return nil, fmt.Errorf("failed to resolve pack UID for pack %s: %w", pName, err)
			}
			pUID = resolvedUID
		}
	case models.V1PackTypeManifest:
		pUID = "spectro-manifest-pack"
	}

	pack := &models.V1PackManifestUpdateEntity{
		//Layer:  p["layer"].(string),
		Name:        types.Ptr(pName),
		Tag:         p["tag"].(string),
		RegistryUID: pRegistryUID,
		UID:         pUID,
		Type:        &pType,
		// UI strips a single newline, so we should do the same
		Values: strings.TrimSpace(p["values"].(string)),
	}

	manifests := make([]*models.V1ManifestRefUpdateEntity, 0)
	for _, manifest := range p["manifest"].([]interface{}) {
		m := manifest.(map[string]interface{})
		manifests = append(manifests, &models.V1ManifestRefUpdateEntity{
			Content: strings.TrimSpace(m["content"].(string)),
			Name:    types.Ptr(m["name"].(string)),
			UID:     getManifestUID(m["name"].(string), packs),
		})
	}
	pack.Manifests = manifests

	return pack, nil
}

func getManifestUID(name string, packs []*models.V1PackRef) string {
	for _, pack := range packs {
		for _, manifest := range pack.Manifests {
			if manifest.Name == name {
				return manifest.UID
			}
		}
	}

	return ""
}

func toClusterProfileVariables(d *schema.ResourceData) ([]*models.V1Variable, error) {
	var profileVariables []*models.V1Variable
	if pVariables, ok := d.GetOk("profile_variables"); ok {

		if pVariables.([]interface{})[0] != nil {
			variables := pVariables.([]interface{})[0].(map[string]interface{})["variable"]
			for _, v := range variables.([]interface{}) {
				variable := v.(map[string]interface{})
				pv := &models.V1Variable{
					DefaultValue: variable["default_value"].(string),
					Description:  variable["description"].(string),
					DisplayName:  variable["display_name"].(string), // revisit
					Format:       types.Ptr(models.V1VariableFormat(variable["format"].(string))),
					Hidden:       variable["hidden"].(bool),
					Immutable:    variable["immutable"].(bool),
					Name:         StringPtr(variable["name"].(string)),
					Regex:        variable["regex"].(string),
					IsSensitive:  variable["is_sensitive"].(bool),
					Required:     variable["required"].(bool),
				}
				profileVariables = append(profileVariables, pv)
			}
		}
	}
	return profileVariables, nil
}

func flattenProfileVariables(d *schema.ResourceData, pv []*models.V1Variable) ([]interface{}, error) {
	if len(pv) == 0 {
		return make([]interface{}, 0), nil
	}
	var variables []interface{}
	for _, v := range pv {
		variable := make(map[string]interface{})
		variable["name"] = v.Name
		variable["display_name"] = v.DisplayName
		variable["description"] = v.Description
		variable["format"] = v.Format
		variable["default_value"] = v.DefaultValue
		variable["regex"] = v.Regex
		variable["required"] = v.Required
		variable["immutable"] = v.Immutable
		variable["hidden"] = v.Hidden
		variable["is_sensitive"] = v.IsSensitive
		variables = append(variables, variable)
	}
	// Sorting ordering the list per configuration this reference if we need to change profile_variables to TypeList
	var sortedVariables []interface{}
	var configVariables []interface{}
	if v, ok := d.GetOk("profile_variables"); ok {
		configVariables = v.([]interface{})[0].(map[string]interface{})["variable"].([]interface{})
		for _, cv := range configVariables {
			mapV := cv.(map[string]interface{})
			for _, va := range variables {
				vs := va.(map[string]interface{})
				if mapV["name"].(string) == String(vs["name"].(*string)) {
					sortedVariables = append(sortedVariables, va)
				}
			}
		}
	} else {
		sortedVariables = variables
	}

	flattenProVariables := make([]interface{}, 1)
	flattenProVariables[0] = map[string]interface{}{
		"variable": sortedVariables,
	}
	return flattenProVariables, nil
}

// resolvePackUID resolves the pack UID based on name, tag, and registry_uid
func resolvePackUID(c *client.V1Client, name, tag, registryUID string) (string, error) {
	if name == "" || tag == "" || registryUID == "" {
		return "", fmt.Errorf("name, tag, and registry_uid are all required for pack resolution")
	}

	// Get pack versions by name and registry
	packVersions, err := c.GetPacksByNameAndRegistry(name, registryUID)
	if err != nil {
		return "", fmt.Errorf("failed to get pack versions for name %s in registry %s: %w", name, registryUID, err)
	}

	if packVersions == nil || len(packVersions.Tags) == 0 {
		return "", fmt.Errorf("no pack found with name %s in registry %s", name, registryUID)
	}

	// Find the pack with matching tag/version
	for _, packTag := range packVersions.Tags {
		if packTag.Version == tag {
			return packTag.PackUID, nil
		}
	}

	return "", fmt.Errorf("no pack found with name %s, tag %s in registry %s", name, tag, registryUID)
}
