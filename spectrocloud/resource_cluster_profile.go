package spectrocloud

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"log"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/gomi/pkg/ptr"
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
			"pack": {
				Type:     schema.TypeList,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  "spectro",
						},
						"registry_uid": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"uid": {
							Type:     schema.TypeString,
							Computed: true,
							Optional: true,
						},
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"manifest": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"uid": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"name": {
										Type:     schema.TypeString,
										Required: true,
									},
									"content": {
										Type:     schema.TypeString,
										Required: true,
										DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
											// UI strips the trailing newline on save
											if strings.TrimSpace(old) == strings.TrimSpace(new) {
												return true
											}
											return false
										},
									},
								},
							},
						},
						//"layer": {
						//	Type:     schema.TypeString,
						//	Required: true,
						//},
						"tag": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"values": {
							Type:     schema.TypeString,
							Optional: true,
							DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
								// UI strips the trailing newline on save
								if strings.TrimSpace(old) == strings.TrimSpace(new) {
									return true
								}
								return false
							},
						},
					},
				},
			},
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

	if err := d.Set("tags", flattenTags(cp.Metadata.Labels)); err != nil {
		return diag.FromErr(err)
	}

	// make a map of all the content
	packManifests := make(map[string][]string)
	for _, p := range cp.Spec.Published.Packs {
		if len(p.Manifests) > 0 {
			content, err := c.GetClusterProfileManifestPack(d.Id(), *p.Name)
			if err != nil {
				return diag.FromErr(err)
			}

			if len(content) > 0 {
				// TODO at some point support multiple manifests... or I hope it's just returned in
				// the original call
				c := make([]string, len(content))
				for i, co := range content {
					c[i] = co.Spec.Published.Content
				}
				packManifests[p.PackUID] = c
			}
		}
	}

	_ = d.Set("name", cp.Metadata.Name)
	packs, err := flattenPacks(c, cp.Spec.Published.Packs, packManifests)
	if err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("pack", packs); err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func flattenPacks(c *client.V1Client, packs []*models.V1PackRef, manifestContent map[string][]string) ([]interface{}, error) {
	if packs == nil {
		return make([]interface{}, 0), nil
	}

	ps := make([]interface{}, len(packs))
	for i, pack := range packs {
		p := make(map[string]interface{})

		p["uid"] = pack.PackUID
		p["registry_uid"] = c.GetPackRegistry(pack.PackUID)
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
				mj["content"] = manifestContent[pack.PackUID][j]

				ma[j] = mj
			}

			p["manifest"] = ma
		}
		ps[i] = p
	}

	return ps, nil
}

func resourceClusterProfileUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.V1Client)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	if d.HasChanges("name") || d.HasChanges("pack") {
		log.Printf("Updating packs")
		cluster, err := toClusterProfileUpdate(d)
		if err != nil {
			return diag.FromErr(err)
		}
		if err := c.UpdateClusterProfile(cluster); err != nil {
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
	c := m.(*client.V1Client)

	var diags diag.Diagnostics

	err := c.DeleteClusterProfile(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func toClusterProfileCreate(d *schema.ResourceData) (*models.V1ClusterProfileEntity, error) {
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
		Name:        ptr.StringPtr(pName),
		Tag:         p["tag"].(string),
		RegistryUID: pRegistryUID,
		UID:         pUID,
		Type:        pType,
		// UI strips a single newline, so we should do the same
		Values: strings.TrimSpace(p["values"].(string)),
	}

	manifests := make([]*models.V1ManifestInputEntity, 0)
	for _, manifest := range p["manifest"].([]interface{}) {
		m := manifest.(map[string]interface{})
		manifests = append(manifests, &models.V1ManifestInputEntity{
			Content: strings.TrimSpace(m["content"].(string)),
			Name:    m["name"].(string),
		})
	}
	pack.Manifests = manifests

	return pack, nil
}

func toClusterProfileUpdate(d *schema.ResourceData) (*models.V1ClusterProfileUpdateEntity, error) {
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
		if p, e := toClusterProfilePackUpdate(pack); e != nil {
			return nil, e
		} else {
			packs = append(packs, p)
		}
	}
	cp.Spec.Template.Packs = packs

	return cp, nil
}

func toClusterProfilePackUpdate(pSrc interface{}) (*models.V1PackManifestUpdateEntity, error) {
	p := pSrc.(map[string]interface{})

	pName := p["name"].(string)
	pTag := p["tag"].(string)
	pUID := p["uid"].(string)
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
		Name: ptr.StringPtr(pName),
		Tag:  p["tag"].(string),
		UID:  pUID,
		Type: pType,
		// UI strips a single newline, so we should do the same
		Values: strings.TrimSpace(p["values"].(string)),
	}

	manifests := make([]*models.V1ManifestRefUpdateEntity, 0)
	for _, manifest := range p["manifest"].([]interface{}) {
		m := manifest.(map[string]interface{})
		manifests = append(manifests, &models.V1ManifestRefUpdateEntity{
			Content: strings.TrimSpace(m["content"].(string)),
			Name:    ptr.StringPtr(m["name"].(string)),
			UID:     m["uid"].(string),
		})
	}
	pack.Manifests = manifests

	return pack, nil
}
