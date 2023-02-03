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
	"github.com/spectrocloud/hapi/models"
	"github.com/spectrocloud/terraform-provider-spectrocloud/pkg/client"
)

func resourceClusterProfile() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceClusterProfileCreate,
		ReadContext:   resourceClusterProfileRead,
		UpdateContext: resourceClusterProfileUpdate,
		DeleteContext: resourceClusterProfileDelete,

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
				Type:     schema.TypeString,
				Optional: true,
				Default:  "1.0.0", // default as in UI
			},
			"context": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "project",
				ValidateFunc: validation.StringInSlice([]string{"", "project", "tenant", "system"}, false),
			},
			"tags": {
				Type:     schema.TypeSet,
				Optional: true,
				Set:      schema.HashString,
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
				ForceNew: true,
			},
			"type": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "add-on",
				ValidateFunc: validation.StringInSlice([]string{"add-on", "cluster", "infra", "system"}, false),
				ForceNew:     true,
			},
			"pack": schemas.PackSchema(),
		},
	}
}

func resourceClusterProfileCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.V1Client)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	clusterProfile, err := toClusterProfileCreate(d)
	if err != nil {
		return diag.FromErr(err)
	}

	// Create
	ProfileContext := d.Get("context").(string)
	uid, err := c.CreateClusterProfile(clusterProfile, ProfileContext)
	if err != nil {
		return diag.FromErr(err)
	}

	// And then publish
	if err = c.PublishClusterProfile(uid, ProfileContext); err != nil {
		return diag.FromErr(err)
	}
	d.SetId(uid)
	resourceClusterProfileRead(ctx, d, m)
	return diags
}

func resourceClusterProfileRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.V1Client)

	var diags diag.Diagnostics

	cp, err := c.GetClusterProfile(d.Id())
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

	packManifests, d2, done2 := getPacksContent(cp.Spec.Published.Packs, c, d)
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
	packs, err := flattenPacks(c, diagPacks, cp.Spec.Published.Packs, packManifests)
	if err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("pack", packs); err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func resourceClusterProfileUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.V1Client)

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

		ProfileContext := d.Get("context").(string)
		if err := c.UpdateClusterProfile(cluster, ProfileContext); err != nil {
			return diag.FromErr(err)
		}
		if err := c.PatchClusterProfile(cluster, metadata, ProfileContext); err != nil {
			return diag.FromErr(err)
		}
		if err := c.PublishClusterProfile(cluster.Metadata.UID, ProfileContext); err != nil {
			return diag.FromErr(err)
		}
	}

	resourceClusterProfileRead(ctx, d, m)

	return diags
}

func resourceClusterProfileDelete(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.V1Client)

	var diags diag.Diagnostics

	err := c.DeleteClusterProfile(d.Id())
	if err != nil {
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
				CloudType: models.V1CloudType(d.Get("cloud").(string)),
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
		Name:        types.Ptr(pName),
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
		if pUID == "" {
			pUID = "spectro-manifest-pack"
		}
	}

	pack := &models.V1PackManifestUpdateEntity{
		//Layer:  p["layer"].(string),
		Name:        types.Ptr(pName),
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
