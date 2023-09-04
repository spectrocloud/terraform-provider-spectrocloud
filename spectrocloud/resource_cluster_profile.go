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
	"github.com/spectrocloud/palette-sdk-go/client"
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
				Description:  "The context of the cluster profile. Allowed values are `project` or `tenant`. Default value is `project`.",
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
				ValidateFunc: validation.StringInSlice([]string{"all", "aws", "azure", "gcp", "vsphere", "openstack", "maas", "nested", "baremetal", "eks", "aks", "edge", "edge-native", "libvirt", "tencent", "tke", "coxedge", "generic", "gke"}, false),
				ForceNew:     true,
				Description: "Specify the infrastructure provider the cluster profile is for. Only Palette supported infrastructure providers can be used. The  supported cloud types are - `all, aws, azure, gcp, vsphere, openstack, maas, nested, baremetal, eks, aks, edge, edge-native, libvirt, tencent, tke, coxedge, generic, and gke`," +
					"If the value is set to `all`, then the type must be set to `add-on`. Otherwise, the cluster profile may be incompatible with other providers. Default value is `all`.",
			},
			"type": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "add-on",
				ValidateFunc: validation.StringInSlice([]string{"add-on", "cluster", "infra", "system"}, false),
				Description:  "Specify the cluster profile type to use. Allowed values are `cluster`, infra`, `add-on`, and `system`. These values map to the following User Interface (UI) labels. Use the value ' cluster ' for a **Full** cluster profile." + "For an Infrastructure cluster profile, use the value `infra`; for an Add-on cluster profile, use the value `add-on`." + "System cluster profiles can be specified using the value `system`. To learn more about cluster profiles, refer to the [Cluster Profile](https://docs.spectrocloud.com/cluster-profiles) documentation. Default value is `add-on`.",
				ForceNew:     true,
			},
			"pack": schemas.PackSchema(),
		},
	}
}

func resourceClusterProfileCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.V1Client)
	clusterC, err := c.GetClusterClient()
	if err != nil {
		return diag.FromErr(err)
	}

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	clusterProfile, err := toClusterProfileCreate(d)
	if err != nil {
		return diag.FromErr(err)
	}

	// Create
	ProfileContext := d.Get("context").(string)
	uid, err := c.CreateClusterProfile(clusterC, clusterProfile, ProfileContext)
	if err != nil {
		return diag.FromErr(err)
	}

	// And then publish
	if err = c.PublishClusterProfile(clusterC, uid, ProfileContext); err != nil {
		return diag.FromErr(err)
	}
	d.SetId(uid)
	resourceClusterProfileRead(ctx, d, m)
	return diags
}

func resourceClusterProfileRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.V1Client)
	clusterC, err := c.GetClusterClient()
	if err != nil {
		return diag.FromErr(err)
	}

	var diags diag.Diagnostics

	cp, err := c.GetClusterProfile(clusterC, d.Id())
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
	clusterC, err := c.GetClusterClient()
	if err != nil {
		return diag.FromErr(err)
	}

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	if d.HasChanges("name") || d.HasChanges("tags") || d.HasChanges("pack") {
		log.Printf("Updating packs")
		cp, err := c.GetClusterProfile(clusterC, d.Id())
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
		if err := c.UpdateClusterProfile(clusterC, cluster, ProfileContext); err != nil {
			return diag.FromErr(err)
		}
		if err := c.PatchClusterProfile(clusterC, cluster, metadata, ProfileContext); err != nil {
			return diag.FromErr(err)
		}
		if err := c.PublishClusterProfile(clusterC, cluster.Metadata.UID, ProfileContext); err != nil {
			return diag.FromErr(err)
		}
	}

	resourceClusterProfileRead(ctx, d, m)

	return diags
}

func resourceClusterProfileDelete(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.V1Client)
	clusterC, err := c.GetClusterClient()
	if err != nil {
		return diag.FromErr(err)
	}

	var diags diag.Diagnostics

	if err := c.DeleteClusterProfile(clusterC, d.Id()); err != nil {
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
