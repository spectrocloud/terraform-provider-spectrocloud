package spectrocloud

import (
	"context"

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
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"cloud": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"type": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "add-on",
				ForceNew: true,
			},
			"pack": {
				Type:     schema.TypeList,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"uid": {
							Type:     schema.TypeString,
							Required: true,
						},
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
						//"layer": {
						//	Type:     schema.TypeString,
						//	Required: true,
						//},
						"tag": {
							Type:     schema.TypeString,
							Required: true,
						},
						"values": {
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
		},
	}
}

func resourceClusterProfileCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.V1alpha1Client)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	clusterProfile := toClusterProfile(d)

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
	c := m.(*client.V1alpha1Client)

	var diags diag.Diagnostics

	cp, err := c.GetClusterProfile(d.Id())
	if err != nil {
		return diag.FromErr(err)
	} else if cp == nil {
		// Deleted - Terraform will recreate it
		d.SetId("")
		return diags
	}

	d.Set("name", cp.Metadata.Name)
	packs := flattenPacks(cp.Spec.Published.Packs)
	if err := d.Set("pack", packs); err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func flattenPacks(packs []*models.V1alpha1PackRef) []interface{} {
	if packs == nil {
		return make([]interface{}, 0)
	}

	ps := make([]interface{}, len(packs))
	for i, pack := range packs {
		p := make(map[string]interface{})

		p["uid"] = pack.PackUID
		p["name"] = *pack.Name
		p["tag"] = pack.Tag
		p["values"] = pack.Values

		ps[i] = p
	}

	return ps
}

func resourceClusterProfileUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.V1alpha1Client)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	if d.HasChanges("pack") {
		log.Printf("Updating packs")
		cluster := toClusterProfile(d)
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

func resourceClusterProfileDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.V1alpha1Client)

	var diags diag.Diagnostics

	err := c.DeleteClusterProfile(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func toClusterProfile(d *schema.ResourceData) *models.V1alpha1ClusterProfileEntity {
	// gnarly, I know! =/
	//cloudConfig := d.Get("cloud_config").([]interface{})[0].(map[string]interface{})
	//clientSecret := strfmt.Password(d.Get("azure_client_secret").(string))

	cluster := &models.V1alpha1ClusterProfileEntity{
		Metadata: &models.V1ObjectMeta{
			Name: d.Get("name").(string),
			UID:  d.Id(),
		},
		Spec: &models.V1alpha1ClusterProfileEntitySpec{
			Template: &models.V1alpha1ClusterProfileTemplateDraft{
				CloudType: models.V1alpha1CloudType(d.Get("cloud").(string)),
				Type:      models.V1alpha1ProfileType(d.Get("type").(string)),
			},
		},
	}

	packs := make([]*models.V1alpha1PackManifestEntity, 0)
	for _, pack := range d.Get("pack").([]interface{}) {
		p := toClusterProfilePack(pack)
		packs = append(packs, p)
	}
	cluster.Spec.Template.Packs = packs

	return cluster
}

func toClusterProfilePack(pSrc interface{}) *models.V1alpha1PackManifestEntity {
	p := pSrc.(map[string]interface{})

	pack := &models.V1alpha1PackManifestEntity{
		//Layer:  p["layer"].(string),
		Name: ptr.StringPtr(p["name"].(string)),
		Tag:  p["tag"].(string),
		UID:  ptr.StringPtr(p["uid"].(string)),
		// UI strips a single newline, so we should do the same
		Values: strings.TrimSpace(p["values"].(string)),
	}
	return pack
}
