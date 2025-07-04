package spectrocloud

import (
	"context"
	"time"

	"github.com/spectrocloud/terraform-provider-spectrocloud/spectrocloud/schemas"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/spectrocloud/palette-sdk-go/client"
)

func resourceClusterGroup() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceClusterGroupCreate,
		ReadContext:   resourceClusterGroupRead,
		UpdateContext: resourceClusterGroupUpdate,
		DeleteContext: resourceClusterGroupDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Minute),
			Update: schema.DefaultTimeout(60 * time.Minute),
			Delete: schema.DefaultTimeout(60 * time.Minute),
		},

		SchemaVersion: 2,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of the cluster group",
			},
			"context": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "tenant",
				ValidateFunc: validation.StringInSlice([]string{"", "project", "tenant"}, false),
				Description: "The context of the Cluster group. Allowed values are `project` or `tenant`. " +
					"Defaults to `tenant`. " + PROJECT_NAME_NUANCE,
			},
			"tags": {
				Type:     schema.TypeSet,
				Optional: true,
				Set:      schema.HashString,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Description: "A list of tags to be applied to the cluster group. Tags must be in the form of `key:value`.",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: "The description of the cluster. Default value is empty string.",
			},
			"config": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"host_endpoint_type": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      "Ingress",
							ValidateFunc: validation.StringInSlice([]string{"Ingress", "LoadBalancer"}, false),
							Description:  "The host endpoint type. Allowed values are 'Ingress' or 'LoadBalancer'. Defaults to 'Ingress'.",
						},
						"cpu_millicore": {
							Type:        schema.TypeInt,
							Optional:    true,
							Description: "The CPU limit in millicores.",
						},
						"memory_in_mb": {
							Type:        schema.TypeInt,
							Optional:    true,
							Description: "The memory limit in megabytes (MB).",
						},
						"storage_in_gb": {
							Type:        schema.TypeInt,
							Optional:    true,
							Description: "The storage limit in gigabytes (GB).",
						},
						"oversubscription_percent": {
							Type:        schema.TypeInt,
							Optional:    true,
							Description: "The allowed oversubscription percentage.",
						},
						"values": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  "",
						},
						"k8s_distribution": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      "k3s",
							ForceNew:     true,
							ValidateFunc: validation.StringInSlice([]string{"k3s", "cncf_k8s"}, false),
							Description:  "The Kubernetes distribution, allowed values are `k3s` and `cncf_k8s`.",
						},
					},
				},
			},
			"cluster_profile": schemas.ClusterProfileSchema(),
			"clusters": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "A list of clusters to include in the cluster group.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"cluster_uid": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The UID of the host cluster.",
						},
						"host_dns": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "The host DNS wildcard for the cluster. i.e. `*.dev` or `*test.com`",
						},
					},
				},
			},
		},
	}
}

func resourceClusterGroupCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	resourceContext := d.Get("context").(string)
	c := getV1ClientWithResourceContext(m, resourceContext)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics
	cluster := toClusterGroup(c, d)

	uid, err := c.CreateClusterGroup(cluster)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(uid)

	resourceClusterGroupRead(ctx, d, m)

	return diags
}

//goland:noinspection GoUnhandledErrorResult
func resourceClusterGroupRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	resourceContext := d.Get("context").(string)
	c := getV1ClientWithResourceContext(m, resourceContext)

	var diags diag.Diagnostics
	uid := d.Id()
	clusterGroup, err := c.GetClusterGroup(uid)
	if err != nil {
		return handleReadError(d, err, diags)
	} else if clusterGroup == nil {
		// Deleted - Terraform will recreate it
		d.SetId("")
		return diags
	}

	return flattenClusterGroup(clusterGroup, d)
}

func flattenClusterGroup(clusterGroup *models.V1ClusterGroup, d *schema.ResourceData) diag.Diagnostics {

	if clusterGroup == nil {
		return diag.Diagnostics{}
	}

	d.SetId(clusterGroup.Metadata.UID)
	err := d.Set("name", clusterGroup.Metadata.Name)
	if err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("tags", flattenTags(clusterGroup.Metadata.Labels)); err != nil {
		return diag.FromErr(err)
	}

	clusterGroupSpec := clusterGroup.Spec
	if clusterGroupSpec != nil {
		clusterConfig := clusterGroupSpec.ClustersConfig
		if clusterConfig != nil {
			limitConfig := clusterConfig.LimitConfig
			if limitConfig != nil {
				err := d.Set("config", []map[string]interface{}{
					{
						"host_endpoint_type":       clusterConfig.EndpointType,
						"cpu_millicore":            limitConfig.CPUMilliCore,
						"memory_in_mb":             limitConfig.MemoryMiB,
						"storage_in_gb":            limitConfig.StorageGiB,
						"oversubscription_percent": limitConfig.OverSubscription,
						"k8s_distribution":         clusterConfig.KubernetesDistroType,
					},
				})
				if err != nil {
					return diag.FromErr(err)
				}
			}
		}
	}

	if clusterGroupSpec != nil {
		clusterConfig := clusterGroupSpec.ClustersConfig
		if clusterConfig != nil {
			hostConfig := clusterConfig.HostClustersConfig
			if hostConfig != nil {
				// set cluster uid and host
				clusters := make([]map[string]interface{}, 0)
				for _, cluster := range hostConfig {
					// if it's ingress config set ingress if it's loadbalancer set loadbalancer
					var host string
					if cluster.EndpointConfig.IngressConfig != nil {
						host = cluster.EndpointConfig.IngressConfig.Host
					}
					clusters = append(clusters, map[string]interface{}{
						"cluster_uid": cluster.ClusterUID,
						"host_dns":    host,
					})
				}
				// set clusters in the schema
				err = d.Set("clusters", clusters)
				if err != nil {
					return diag.FromErr(err)
				}
			}
		}
	}

	return diag.Diagnostics{}
}

func resourceClusterGroupUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	resourceContext := d.Get("context").(string)
	c := getV1ClientWithResourceContext(m, resourceContext)
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics
	// if there are changes in the name of  cluster group, update it using UpdateClusterGroupMeta()
	clusterGroup := toClusterGroup(c, d)
	err := c.UpdateClusterGroupMeta(clusterGroup)
	if err != nil {
		return diag.FromErr(err)
	}
	if d.HasChanges("config", "clusters") {
		clusterGroup := toClusterGroup(c, d)

		err := c.UpdateClusterGroup(clusterGroup.Metadata.UID, toClusterGroupUpdate(clusterGroup))
		if err != nil {
			return diag.FromErr(err)
		}
	}
	if d.HasChange("cluster_profile") {
		clusterGroupContext := d.Get("context").(string)
		profiles, _ := toProfilesCommon(c, d, "", clusterGroupContext)
		profilesBody := &models.V1SpectroClusterProfiles{
			Profiles: profiles,
		}
		err := c.UpdateClusterProfileInClusterGroup(d.Id(), profilesBody)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	resourceClusterGroupRead(ctx, d, m)

	return diags
}

func toClusterGroup(c *client.V1Client, d *schema.ResourceData) *models.V1ClusterGroupEntity {
	clusterRefs := make([]*models.V1ClusterGroupClusterRef, 0)
	clusterRefObj, ok := d.GetOk("clusters")
	clusterGroupContext := d.Get("context").(string)
	if ok {
		for i := range clusterRefObj.([]interface{}) {
			resources := clusterRefObj.([]interface{})[i].(map[string]interface{})
			mp := toClusterRef(resources)
			clusterRefs = append(clusterRefs, mp)
		}
	}

	var clusterGroupLimitConfig *models.V1ClusterGroupLimitConfig
	var values string
	resourcesObj, ok := d.GetOk("config")
	endpointType := "Ingress" // default endpoint type is ingress
	var k8Distro string
	if ok {
		resources := resourcesObj.([]interface{})[0].(map[string]interface{})
		clusterGroupLimitConfig = toClusterGroupLimitConfig(resources)
		if resources["values"] != nil {
			values = resources["values"].(string)
		}
		if resources["host_endpoint_type"] != nil {
			endpointType = resources["host_endpoint_type"].(string)
		}
		if resources["k8s_distribution"] != nil {
			k8Distro = resources["k8s_distribution"].(string)
		}
	}
	var hostClusterConfig []*models.V1ClusterGroupHostClusterConfig
	if endpointType == "Ingress" {
		hostClusterConfig = toHostClusterConfigs(clusterRefObj.([]interface{}))
	}

	ret := &models.V1ClusterGroupEntity{
		Metadata: getClusterMetadata(d),
		Spec: &models.V1ClusterGroupSpecEntity{
			Type:           "hostCluster",
			ClusterRefs:    clusterRefs,
			ClustersConfig: GetClusterGroupConfig(clusterGroupLimitConfig, hostClusterConfig, endpointType, values, k8Distro),
		},
	}
	profiles, _ := toProfilesCommon(c, d, "", clusterGroupContext)
	if len(profiles) > 0 {
		ret.Spec.Profiles = profiles
	}

	return ret
}

func GetClusterGroupConfig(clusterGroupLimitConfig *models.V1ClusterGroupLimitConfig, hostClusterConfig []*models.V1ClusterGroupHostClusterConfig, endpointType, values, k8Distro string) *models.V1ClusterGroupClustersConfig {
	if values != "" {
		return &models.V1ClusterGroupClustersConfig{
			EndpointType:         endpointType,
			LimitConfig:          clusterGroupLimitConfig,
			HostClustersConfig:   hostClusterConfig,
			Values:               values,
			KubernetesDistroType: models.V1ClusterKubernetesDistroType(k8Distro).Pointer(),
		}
	} else {
		return &models.V1ClusterGroupClustersConfig{
			EndpointType:         endpointType,
			LimitConfig:          clusterGroupLimitConfig,
			HostClustersConfig:   hostClusterConfig,
			KubernetesDistroType: models.V1ClusterKubernetesDistroType(k8Distro).Pointer(),
		}
	}
}

func toHostClusterConfigs(clusterConfig []interface{}) []*models.V1ClusterGroupHostClusterConfig {
	var hostClusterConfigs []*models.V1ClusterGroupHostClusterConfig
	for _, obj := range clusterConfig {
		resources := obj.(map[string]interface{})
		hostCluster := &models.V1ClusterGroupHostClusterConfig{
			ClusterUID: resources["cluster_uid"].(string),
			EndpointConfig: &models.V1HostClusterEndpointConfig{
				IngressConfig: &models.V1IngressConfig{
					Host: resources["host_dns"].(string),
				},
			},
		}
		hostClusterConfigs = append(hostClusterConfigs, hostCluster)
	}
	return hostClusterConfigs
}

func toClusterGroupUpdate(clusterGroupEntity *models.V1ClusterGroupEntity) *models.V1ClusterGroupHostClusterEntity {
	ret := &models.V1ClusterGroupHostClusterEntity{
		ClusterRefs:    clusterGroupEntity.Spec.ClusterRefs,
		ClustersConfig: clusterGroupEntity.Spec.ClustersConfig,
	}

	return ret
}

func toClusterRef(resources map[string]interface{}) *models.V1ClusterGroupClusterRef {
	cluster_uid := resources["cluster_uid"].(string)

	ret := &models.V1ClusterGroupClusterRef{
		ClusterUID: cluster_uid,
	}

	return ret
}

func toClusterGroupLimitConfig(resources map[string]interface{}) *models.V1ClusterGroupLimitConfig {
	cpu_milli := resources["cpu_millicore"].(int)
	mem_in_mb := resources["memory_in_mb"].(int)
	storage_in_gb := resources["storage_in_gb"].(int)
	oversubscription := resources["oversubscription_percent"].(int)

	ret := &models.V1ClusterGroupLimitConfig{

		CPUMilliCore:     int32(cpu_milli),
		MemoryMiB:        int32(mem_in_mb),
		StorageGiB:       int32(storage_in_gb),
		OverSubscription: int32(oversubscription),
	}

	return ret
}

func resourceClusterGroupDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	resourceContext := d.Get("context").(string)
	c := getV1ClientWithResourceContext(m, resourceContext)

	var diags diag.Diagnostics
	err := c.DeleteClusterGroup(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}
