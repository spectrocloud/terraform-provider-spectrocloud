package spectrocloud

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/go-cty/cty"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/spectrocloud/hapi/models"
	"github.com/spectrocloud/terraform-provider-spectrocloud/pkg/client"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceClusterImport() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceCloudClusterImport,
		ReadContext:   resourceCloudClusterRead,
		UpdateContext: resourceCloudClusterUpdate,
		DeleteContext: resourceClusterDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		SchemaVersion: 2,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"cloud": {
				Type:             schema.TypeString,
				ForceNew:         true,
				Required:         true,
				ValidateDiagFunc: validateCloudType,
			},
			"cloud_config_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"cluster_import_manifest_apply_command": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"cluster_import_manifest": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"cluster_profile_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"pack": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"tag": {
							Type:     schema.TypeString,
							Required: true,
						},
						"values": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
		},
	}
}

func resourceCloudClusterImport(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.V1alpha1Client)
	var diags diag.Diagnostics
	uid, err := cloudClusterImportFunc(c, d)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(uid)
	stateConf := &resource.StateChangeConf{
		Target:     []string{"Pending"},
		Refresh:    resourceClusterStateRefreshFunc(c, d.Id()),
		Timeout:    d.Timeout(schema.TimeoutCreate) - 1*time.Minute,
		MinTimeout: 1 * time.Second,
		Delay:      5 * time.Second,
	}

	// Wait, catching any errors
	_, err = stateConf.WaitForStateContext(ctx)
	if err != nil {
		return diag.FromErr(err)
	}

	resourceCloudClusterRead(ctx, d, m)

	if profiles := toCloudClusterProfiles(d); profiles != nil {
		if err := c.UpdateClusterProfileValues(uid, profiles); err != nil {
			return diag.FromErr(err)
		}
	}
	return diags
}

func resourceCloudClusterRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	cloudType := d.Get("cloud").(string)

	c := m.(*client.V1alpha1Client)

	var diags diag.Diagnostics
	uid := d.Id()
	cluster, err := c.GetCluster(uid)
	if err != nil {
		return diag.FromErr(err)
	} else if cluster == nil {
		d.SetId("")
		return diags
	}

	if err := resourceCloudClusterImportManoifests(cluster, d, c); err != nil {
		return diag.FromErr(err)
	}

	if cluster.Status.State == "Running" {
		switch cloudType {
		case "aws":
			return flattenCloudConfigAws(cluster.Spec.CloudConfigRef.UID, d, c)
		case "azure":
			return flattenCloudConfigAzure(cluster.Spec.CloudConfigRef.UID, d, c)
		case "gcp":
			return flattenCloudConfigGcp(cluster.Spec.CloudConfigRef.UID, d, c)
		case "vsphere":
			return flattenCloudConfigVsphere(cluster.Spec.CloudConfigRef.UID, d, c)
		}
		return diag.FromErr(fmt.Errorf("failed to import cluster as cloud type '%s' is invalid", cloudType))
	}

	return diag.Diagnostics{}
}

func resourceCloudClusterImportManoifests(cluster *models.V1alpha1SpectroCluster, d *schema.ResourceData, c *client.V1alpha1Client) error {
	if cluster.Status != nil && cluster.Status.ClusterImport != nil && cluster.Status.ClusterImport.IsBrownfield {
		if err := d.Set("cluster_import_manifest_apply_command", cluster.Status.ClusterImport.ImportLink); err != nil {
			return err
		}

		importManifest, err := c.GetClusterImportManifest(cluster.Metadata.UID)
		if err != nil {
			return err
		}
		if err := d.Set("cluster_import_manifest", importManifest); err != nil {
			return err
		}
	}
	return nil
}

func cloudClusterImportFunc(c *client.V1alpha1Client, d *schema.ResourceData) (string, error) {
	meta := toClusterMeta(d)
	cloudType := d.Get("cloud").(string)
	switch cloudType {
	case "aws":
		return c.ImportClusterAws(meta)
	case "azure":
		return c.ImportClusterAzure(meta)
	case "gcp":
		return c.ImportClusterGcp(meta)
	case "vsphere":
		return c.ImportClusterVsphere(meta)
	}
	return "", fmt.Errorf("failed to find cloud type %s", cloudType)
}

func resourceCloudClusterUpdate(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.V1alpha1Client)
	var diags diag.Diagnostics

	clusterProfileId := d.Get("cluster_profile_id").(string)
	profiles := make([]*models.V1alpha1SpectroClusterProfileEntity, 0)
	packValues := make([]*models.V1alpha1PackValuesEntity, 0)
	for _, pack := range d.Get("pack").([]interface{}) {
		p := toPack(pack)
		packValues = append(packValues, p)
	}

	profiles = append(profiles, &models.V1alpha1SpectroClusterProfileEntity{
		PackValues: packValues,
		UID:        clusterProfileId,
	})

	err := c.UpdateClusterProfileValues(d.Id(), &models.V1alpha1SpectroClusterProfiles{
		Profiles: profiles,
	})
	if err != nil {
		return diag.FromErr(err)
	}
	return diags
}

func toCloudClusterProfiles(d *schema.ResourceData) *models.V1alpha1SpectroClusterProfiles {
	if clusterProfileUid := d.Get("cluster_profile_id"); clusterProfileUid != nil {
		profileEntities := make([]*models.V1alpha1SpectroClusterProfileEntity, 0)
		packValues := make([]*models.V1alpha1PackValuesEntity, 0)
		for _, pack := range d.Get("pack").([]interface{}) {
			p := toPack(pack)
			packValues = append(packValues, p)
		}

		profileEntities = append(profileEntities, &models.V1alpha1SpectroClusterProfileEntity{
			PackValues: packValues,
			UID:        clusterProfileUid.(string),
		})
		return &models.V1alpha1SpectroClusterProfiles{
			Profiles: profileEntities,
		}
	}
	return nil
}

func validateCloudType(data interface{}, path cty.Path) diag.Diagnostics {
	var diags diag.Diagnostics
	inCloudType := data.(string)
	for _, cloudType := range []string{"aws", "azure", "gcp", "vsphere"} {
		if cloudType == inCloudType {
			return diags
		}
	}
	return diag.FromErr(fmt.Errorf("cloud type '%s' is invalid. valid cloud types are %v", inCloudType, "cloud_types"))
}

func toClusterMeta(d *schema.ResourceData) *models.V1ObjectMetaInputEntity {
	return &models.V1ObjectMetaInputEntity{
		Name: d.Get("name").(string),
	}
}
