package spectrocloud

import (
	"context"
	"crypto/tls"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/spectrocloud/terraform-provider-spectrocloud/pkg/client"
)

func New(_ string) func() *schema.Provider {
	return func() *schema.Provider {
		p := &schema.Provider{
			Schema: map[string]*schema.Schema{
				"host": {
					Type:        schema.TypeString,
					Optional:    true,
					DefaultFunc: schema.EnvDefaultFunc("SPECTROCLOUD_HOST", "api.spectrocloud.com"),
				},
				"username": {
					Type:        schema.TypeString,
					Optional:    true,
					DefaultFunc: schema.EnvDefaultFunc("SPECTROCLOUD_USERNAME", nil),
				},
				"password": {
					Type:        schema.TypeString,
					Optional:    true,
					Sensitive:   true,
					DefaultFunc: schema.EnvDefaultFunc("SPECTROCLOUD_PASSWORD", nil),
				},
				"api_key": {
					Type:        schema.TypeString,
					Optional:    true,
					Sensitive:   true,
					DefaultFunc: schema.EnvDefaultFunc("SPECTROCLOUD_APIKEY", nil),
				},
				"trace": {
					Type:        schema.TypeBool,
					Optional:    true,
					DefaultFunc: schema.EnvDefaultFunc("SPECTROCLOUD_TRACE", nil),
				},
				"retry_attempts": {
					Type:        schema.TypeInt,
					Optional:    true,
					DefaultFunc: schema.EnvDefaultFunc("SPECTROCLOUD_RETRY_ATTEMPTS", 10),
				},
				"project_name": {
					Type:     schema.TypeString,
					Optional: true,
				},
				"ignore_insecure_tls_error": {
					Type:     schema.TypeBool,
					Optional: true,
				},
			},
			ResourcesMap: map[string]*schema.Resource{
				"spectrocloud_team": resourceTeam(),

				"spectrocloud_project": resourceProject(),

				"spectrocloud_macro": resourceMacro(),

				"spectrocloud_application_profile": resourceApplicationProfile(),
				"spectrocloud_cluster_profile":     resourceClusterProfile(),

				"spectrocloud_cloudaccount_aws": resourceCloudAccountAws(),
				"spectrocloud_cluster_aws":      resourceClusterAws(),

				"spectrocloud_cloudaccount_maas": resourceCloudAccountMaas(),
				"spectrocloud_cluster_maas":      resourceClusterMaas(),

				"spectrocloud_cluster_eks": resourceClusterEks(),

				"spectrocloud_cloudaccount_tencent": resourceCloudAccountTencent(),
				"spectrocloud_cluster_tke":          resourceClusterTke(),

				"spectrocloud_cloudaccount_azure": resourceCloudAccountAzure(),
				"spectrocloud_cluster_azure":      resourceClusterAzure(),

				"spectrocloud_cluster_aks": resourceClusterAks(),

				"spectrocloud_cloudaccount_gcp": resourceCloudAccountGcp(),
				"spectrocloud_cluster_gcp":      resourceClusterGcp(),

				"spectrocloud_cloudaccount_openstack": resourceCloudAccountOpenstack(),
				"spectrocloud_cluster_openstack":      resourceClusterOpenStack(),

				"spectrocloud_cluster_vsphere": resourceClusterVsphere(),

				"spectrocloud_cluster_libvirt": resourceClusterLibvirt(),

				"spectrocloud_cluster_edge_native": resourceClusterEdgeNative(),

				"spectrocloud_cluster_edge": resourceClusterEdge(),

				"spectrocloud_cluster_edge_vsphere": resourceClusterEdgeVsphere(),

				"spectrocloud_virtual_cluster": resourceClusterNested(),

				"spectrocloud_cluster_import": resourceClusterImport(),

				"spectrocloud_addon_deployment": resourceAddonDeployment(),

				"spectrocloud_application": resourceApplication(),

				"spectrocloud_privatecloudgateway_ippool": resourcePrivateCloudGatewayIpPool(),

				"spectrocloud_backup_storage_location": resourceBackupStorageLocation(),

				"spectrocloud_registry_oci":  resourceRegistryOciEcr(),
				"spectrocloud_registry_helm": resourceRegistryHelm(),

				"spectrocloud_appliance": resourceAppliance(),

				"spectrocloud_workspace": resourceWorkspace(),
			},
			DataSourcesMap: map[string]*schema.Resource{
				"spectrocloud_user":    dataSourceUser(),
				"spectrocloud_project": dataSourceProject(),
				"spectrocloud_role":    dataSourceRole(),

				"spectrocloud_pack": dataSourcePack(),

				"spectrocloud_cluster_profile": dataSourceClusterProfile(),

				"spectrocloud_cloudaccount_aws":       dataSourceCloudAccountAws(),
				"spectrocloud_cloudaccount_tencent":   dataSourceCloudAccountTencent(),
				"spectrocloud_cloudaccount_azure":     dataSourceCloudAccountAzure(),
				"spectrocloud_cloudaccount_gcp":       dataSourceCloudAccountGcp(),
				"spectrocloud_cloudaccount_vsphere":   dataSourceCloudAccountVsphere(),
				"spectrocloud_cloudaccount_openstack": dataSourceCloudAccountOpenStack(),
				"spectrocloud_cloudaccount_maas":      dataSourceCloudAccountMaas(),

				"spectrocloud_backup_storage_location": dataSourceBackupStorageLocation(),

				"spectrocloud_registry_pack": dataSourceRegistryPack(),
				"spectrocloud_registry_helm": dataSourceRegistryHelm(),
				"spectrocloud_registry_oci":  dataSourceRegistryOci(),
				"spectrocloud_registry":      dataSourceRegistry(), // registry datasource for all types.

				"spectrocloud_appliance":           dataSourceAppliance(),
				"spectrocloud_cluster":             dataSourceCluster(),
				"spectrocloud_cluster_group":       dataSourceClusterGroup(),
				"spectrocloud_application_profile": dataSourceApplicationProfile(),
				"spectrocloud_workspace":           dataSourceWorkspace(),
			},
			ConfigureContextFunc: providerConfigure,
		}

		return p
	}
}

func providerConfigure(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	host := d.Get("host").(string)
	username := ""
	password := ""
	apiKey := ""
	transportDebug := false
	retryAttempts := 10

	if d.Get("trace") != nil {
		transportDebug = d.Get("trace").(bool)
	}

	if d.Get("retry_attempts") != nil {
		retryAttempts = d.Get("retry_attempts").(int)
	}

	if d.Get("username") != nil && d.Get("password") != nil {
		username = d.Get("username").(string)
		password = d.Get("password").(string)
	}
	if d.Get("api_key") != nil {
		apiKey = d.Get("api_key").(string)
	}
	projectName := d.Get("project_name").(string)
	ignoreTlsError := d.Get("ignore_insecure_tls_error").(bool)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	if (apiKey == "") && ((username == "") || (password == "")) {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to create Spectro Cloud client",
			Detail:   "Unable to authenticate user for authenticated Spectro Cloud client",
		})
		// TODO(saamalik) verify this block "can" happen (e.g: does required guard this?)
		return nil, diags
	}

	if ignoreTlsError {
		http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	c := client.New(host, username, password, "", apiKey, transportDebug, retryAttempts)

	if projectName != "" {
		uid, err := c.GetProjectUID(projectName)
		if err != nil {
			return nil, diag.FromErr(err)
		}

		c = client.New(host, username, password, uid, apiKey, transportDebug, retryAttempts)
	}

	return c, diags

}
