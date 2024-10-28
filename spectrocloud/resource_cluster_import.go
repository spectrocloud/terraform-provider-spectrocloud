package spectrocloud

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"github.com/spectrocloud/terraform-provider-spectrocloud/spectrocloud/schemas"

	"github.com/hashicorp/go-cty/cty"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/spectrocloud/palette-sdk-go/client"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceClusterImport() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceCloudClusterImport,
		ReadContext:   resourceCloudClusterRead,
		UpdateContext: resourceCloudClusterUpdate,
		DeleteContext: resourceClusterDelete,
		DeprecationMessage: "This resource is deprecated. To import a cluster deployed outside of the context of Palette, refer to the Import a Cluster guide at https://docs.spectrocloud.com/clusters/imported-clusters/cluster-import/) to learn more.",

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
			"context": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "project",
				ValidateFunc: validation.StringInSlice([]string{"", "project", "tenant"}, false),
				Description: "The context of the cluster. Can be `project` or `tenant`. " +
					"Default is `project`." + PROJECT_NAME_NUANCE,
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
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: "The description of the cluster. Default value is empty string.",
			},
			"cluster_profile": schemas.ClusterProfileSchema(),
			"cloud": {
				Type:             schema.TypeString,
				ForceNew:         true,
				Required:         true,
				ValidateDiagFunc: validateCloudType,
			},
			"cloud_config_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "ID of the cloud config used for the cluster. This cloud config must be of type `azure`.",
				Deprecated:  "This field is deprecated and will be removed in the future. Use `cloud_config` instead.",
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
	resourceContext := d.Get("context").(string)
	c := getV1ClientWithResourceContext(m, resourceContext)
	var diags diag.Diagnostics
	uid, err := cloudClusterImportFunc(c, d)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(uid)
	stateConf := &retry.StateChangeConf{
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

	profiles, err := toCloudClusterProfiles(c, d)
	if err != nil {
		return diag.FromErr(err)
	}
	if profiles != nil {
		if err := c.UpdateClusterProfileValues(uid, profiles); err != nil {
			return diag.FromErr(err)
		}
	}
	return diags
}

func resourceCloudClusterRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	resourceContext := d.Get("context").(string)
	c := getV1ClientWithResourceContext(m, resourceContext)

	var diags diag.Diagnostics

	cluster, err := resourceClusterRead(d, c, diags)
	if err != nil {
		return diag.FromErr(err)
	} else if cluster == nil {
		// Deleted - Terraform will recreate it
		d.SetId("")
		return diags
	}

	if err := resourceCloudClusterImportManoifests(cluster, d, c); err != nil {
		return diag.FromErr(err)
	}

	cloudType := d.Get("cloud").(string)
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
				//context := d.Get("context").(string)
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
	meta := toClusterMetadataUpdate(d)
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
	resourceContext := d.Get("context").(string)
	c := getV1ClientWithResourceContext(m, resourceContext)
	var diags diag.Diagnostics

	profiles, err := toCloudClusterProfiles(c, d)
	if err != nil {
		return diag.FromErr(err)
	}
	err = c.UpdateClusterProfileValues(d.Id(), profiles)
	if err != nil {
		return diag.FromErr(err)
	}
	return diags
}

func toCloudClusterProfiles(c *client.V1Client, d *schema.ResourceData) (*models.V1SpectroClusterProfiles, error) {
	if profiles := d.Get("cluster_profile").([]interface{}); len(profiles) > 0 {
		clusterContext := d.Get("context").(string)
		profiles, err := toProfiles(c, d, clusterContext)
		if err != nil {
			return nil, err
		}
		return &models.V1SpectroClusterProfiles{
			Profiles: profiles,
		}, nil
	}
	return nil, errors.New("failed to convert cluster profiles")
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
	if err := d.Set("cloud_config_id", configUID); err != nil {
		return diag.FromErr(err)
	}
	return diag.Diagnostics{}
}
