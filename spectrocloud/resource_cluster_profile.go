package spectrocloud

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/spectrocloud/terraform-provider-spectrocloud/util/ptr"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/spectrocloud/terraform-provider-spectrocloud/spectrocloud/schemas"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-sdk-go/api/models"
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
				Type:         schema.TypeString,
				Default:      "all",
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"all", "aws", "azure", "gcp", "vsphere", "openstack", "maas", "virtual", "baremetal", "eks", "aks", "edge", "edge-native", "tencent", "tke", "generic", "gke"}, false),
				ForceNew:     true,
				Description: "Specify the infrastructure provider the cluster profile is for. Only Palette supported infrastructure providers can be used. The  supported cloud types are - `all, aws, azure, gcp, vsphere, openstack, maas, virtual, baremetal, eks, aks, edge, edge-native, tencent, tke, generic, and gke`," +
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

	clusterProfile, err := toClusterProfileCreate(d)
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

	// if id contains colon - it's incorrect as the scope is not supported
	if strings.Contains(d.Id(), ":") {
		return diag.FromErr(fmt.Errorf("incorrect cluster profile id: %s, scope is not supported", d.Id()))
	}

	cp, err := c.GetClusterProfile(d.Id())
	if err != nil {
		return diag.FromErr(err)
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

	if d.HasChanges("name") || d.HasChanges("tags") || d.HasChanges("pack") {
		log.Printf("Updating packs")
		cp, err := c.GetClusterProfile(d.Id())
		if err != nil {
			return diag.FromErr(err)
		}
		cluster, err := toClusterProfileUpdate(d, cp)
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

func toClusterProfileCreate(d *schema.ResourceData) (*models.V1ClusterProfileEntity, error) {
	cp := toClusterProfileBasic(d)

	packs := make([]*models.V1PackManifestEntity, 0)
	for _, pack := range d.Get("pack").([]interface{}) {
		if p, e := toClusterProfilePackCreate(pack); e != nil {
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
				Type:      models.V1ProfileType(d.Get("type").(string)),
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

	switch pType {
	case models.V1PackTypeSpectro:
		if pTag == "" || pUID == "" {
			return nil, fmt.Errorf("pack %s needs to specify tag and/or uid", pName)
		}
	case models.V1PackTypeManifest:
		if pUID == "" {
			pUID = "spectro-manifest-pack"
		}
	}

	pack := &models.V1PackManifestEntity{
		Name:        ptr.To(pName),
		Tag:         p["tag"].(string),
		RegistryUID: pRegistryUID,
		UID:         pUID,
		Type:        pType,
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

func toClusterProfileUpdate(d *schema.ResourceData, cluster *models.V1ClusterProfile) (*models.V1ClusterProfileUpdateEntity, error) {
	cp := &models.V1ClusterProfileUpdateEntity{
		Metadata: &models.V1ObjectMeta{
			Name: d.Get("name").(string),
			UID:  d.Id(),
		},
		Spec: &models.V1ClusterProfileUpdateEntitySpec{
			Template: &models.V1ClusterProfileTemplateUpdate{
				Type: models.V1ProfileType(d.Get("type").(string)),
			},
			Version: d.Get("version").(string),
		},
	}
	packs := make([]*models.V1PackManifestUpdateEntity, 0)
	for _, pack := range d.Get("pack").([]interface{}) {
		if p, e := toClusterProfilePackUpdate(pack, cluster.Spec.Published.Packs); e != nil {
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

func toClusterProfilePackUpdate(pSrc interface{}, packs []*models.V1PackRef) (*models.V1PackManifestUpdateEntity, error) {
	p := pSrc.(map[string]interface{})

	pName := p["name"].(string)
	pTag := p["tag"].(string)
	pUID := p["uid"].(string)

	pRegistryUID := ""
	if p["registry_uid"] != nil {
		pRegistryUID = p["registry_uid"].(string)
	}
	pType := models.V1PackType(p["type"].(string))

	switch pType {
	case models.V1PackTypeSpectro:
		if pTag == "" || pUID == "" {
			return nil, fmt.Errorf("pack %s needs to specify tag", pName)
		}
	case models.V1PackTypeManifest:
		pUID = "spectro-manifest-pack"
	}

	pack := &models.V1PackManifestUpdateEntity{
		//Layer:  p["layer"].(string),
		Name:        ptr.To(pName),
		Tag:         p["tag"].(string),
		RegistryUID: pRegistryUID,
		UID:         pUID,
		Type:        pType,
		// UI strips a single newline, so we should do the same
		Values: strings.TrimSpace(p["values"].(string)),
	}

	manifests := make([]*models.V1ManifestRefUpdateEntity, 0)
	for _, manifest := range p["manifest"].([]interface{}) {
		m := manifest.(map[string]interface{})
		manifests = append(manifests, &models.V1ManifestRefUpdateEntity{
			Content: strings.TrimSpace(m["content"].(string)),
			Name:    ptr.To(m["name"].(string)),
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

		// Once the profile_Variables feature is extended to all cloud types, the following block should be removed.
		cloudType, _ := d.Get("cloud").(string)
		profileType, _ := d.Get("type").(string)
		if cloudType != "edge-native" {
			if profileType != "add-on" {
				err := errors.New("currently, `profile_variables` is only supported for the `add-on` profile type and other profile type is supported only for edge-native cloud type")
				return profileVariables, err
			}
		}

		if pVariables.([]interface{})[0] != nil {
			variables := pVariables.([]interface{})[0].(map[string]interface{})["variable"]
			for _, v := range variables.([]interface{}) {
				variable := v.(map[string]interface{})
				pv := &models.V1Variable{
					DefaultValue: variable["default_value"].(string),
					Description:  variable["description"].(string),
					DisplayName:  variable["display_name"].(string), // revisit
					Format:       models.V1VariableFormat(variable["format"].(string)),
					Hidden:       variable["hidden"].(bool),
					Immutable:    variable["immutable"].(bool),
					Name:         ptr.To(variable["name"].(string)),
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
	configVariables := d.Get("profile_variables").([]interface{})[0].(map[string]interface{})["variable"].([]interface{}) //([]interface{}) //(*schema.Set).List()
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
	for _, cv := range configVariables {
		mapV := cv.(map[string]interface{})
		for _, va := range variables {
			vs := va.(map[string]interface{})
			if mapV["name"].(string) == ptr.DeRef(vs["name"].(*string)) {
				sortedVariables = append(sortedVariables, va)
			}
		}
	}

	flattenProVariables := make([]interface{}, 1)
	flattenProVariables[0] = map[string]interface{}{
		"variable": sortedVariables,
	}
	return flattenProVariables, nil
}
