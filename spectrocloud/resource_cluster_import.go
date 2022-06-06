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
			"tags": {
				Type:     schema.TypeSet,
				Optional: true,
				Set:      schema.HashString,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"cluster_profile": {
				Type:          schema.TypeList,
				Optional:      true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Required: true,
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
									"registry_uid": {
										Type:     schema.TypeString,
										Optional: true,
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
				},
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
		},
	}
}

func resourceCloudClusterImport(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.V1Client)
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

	c := m.(*client.V1Client)

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
		case "generic":
			return flattenCloudConfigGeneric(cluster.Spec.CloudConfigRef.UID, d, c)
		}
		return diag.FromErr(fmt.Errorf("failed to import cluster as cloud type '%s' is invalid", cloudType))
	}

	return diag.Diagnostics{}
}

func resourceCloudClusterImportManoifests(cluster *models.V1SpectroCluster, d *schema.ResourceData, c *client.V1Client) error {
	if cluster.Status != nil && cluster.Status.ClusterImport != nil && cluster.Status.ClusterImport.IsBrownfield {
		if err := d.Set("cluster_import_manifest_apply_command", cluster.Status.ClusterImport.ImportLink); err != nil {
			return err
		}

		//only if apply tag is true as downloading manifest from upstream changes cluster state to
		// Importing from Pending which isn't desired until intention is to apply the manifest locally
		if len(cluster.Metadata.Labels) > 0 {
			if v, ok := cluster.Metadata.Labels["apply"]; ok && v == "true" {
				importManifest, err := c.GetClusterImportManifest(cluster.Metadata.UID)
				if err != nil {
					return err
				}

				if err := d.Set("cluster_import_manifest", importManifest); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func cloudClusterImportFunc(c *client.V1Client, d *schema.ResourceData) (string, error) {
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
	case "generic":
		return c.ImportClusterGeneric(meta)
	}
	return "", fmt.Errorf("failed to find cloud type %s", cloudType)
}

func resourceCloudClusterUpdate(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.V1Client)
	var diags diag.Diagnostics

	err := c.UpdateClusterProfileValues(d.Id(), toCloudClusterProfiles(d))
	if err != nil {
		return diag.FromErr(err)
	}
	return diags
}

func toCloudClusterProfiles(d *schema.ResourceData) *models.V1SpectroClusterProfiles {
	if profiles := d.Get("cluster_profile").([]interface{}); len(profiles) > 0 {
		return &models.V1SpectroClusterProfiles{
			Profiles: toProfiles(d),
		}
	}
	return nil
}

func validateCloudType(data interface{}, path cty.Path) diag.Diagnostics {
	var diags diag.Diagnostics
	inCloudType := data.(string)
	for _, cloudType := range []string{"aws", "azure", "gcp", "vsphere", "generic"} {
		if cloudType == inCloudType {
			return diags
		}
	}
	return diag.FromErr(fmt.Errorf("cloud type '%s' is invalid. valid cloud types are %v", inCloudType, "cloud_types"))
}

func flattenCloudConfigGeneric(configUID string, d *schema.ResourceData, c *client.V1Client) diag.Diagnostics {
	d.Set("cloud_config_id", configUID)
	return diag.Diagnostics{}
}

func toClusterMeta(d *schema.ResourceData) *models.V1ObjectMetaInputEntity {
	return &models.V1ObjectMetaInputEntity{
		Name: d.Get("name").(string),
		Labels: toTags(d),
	}
}
