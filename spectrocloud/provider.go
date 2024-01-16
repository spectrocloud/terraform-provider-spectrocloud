package spectrocloud

import (
	"context"
	"crypto/tls"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/spectrocloud/palette-sdk-go/client"
)

// make a constant string describing which project will be specified.
const (
	PROJECT_NAME_NUANCE = "If  the `project` context is specified, the project name will sourced from the provider configuration parameter " +
		"[`project_name`](https://registry.terraform.io/providers/spectrocloud/spectrocloud/latest/docs#schema)."
)

func New(_ string) func() *schema.Provider {
	return func() *schema.Provider {
		p := &schema.Provider{
			Schema: map[string]*schema.Schema{
				"host": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: "The Spectro Cloud API host url. Can also be set with the `SPECTROCLOUD_HOST` environment variable. Defaults to https://api.spectrocloud.com",
					DefaultFunc: schema.EnvDefaultFunc("SPECTROCLOUD_HOST", "api.spectrocloud.com"),
				},
				"api_key": {
					Type:        schema.TypeString,
					Optional:    true,
					Sensitive:   true,
					Description: "The Spectro Cloud API key. Can also be set with the `SPECTROCLOUD_APIKEY` environment variable.",
					DefaultFunc: schema.EnvDefaultFunc("SPECTROCLOUD_APIKEY", nil),
				},
				"trace": {
					Type:        schema.TypeBool,
					Optional:    true,
					Description: "Enable HTTP request tracing. Can also be set with the `SPECTROCLOUD_TRACE` environment variable. To enable Terraform debug logging, set `TF_LOG=DEBUG`. Visit the Terraform documentation to learn more about Terraform [debugging](https://developer.hashicorp.com/terraform/plugin/log/managing).",
					DefaultFunc: schema.EnvDefaultFunc("SPECTROCLOUD_TRACE", nil),
				},
				"retry_attempts": {
					Type:        schema.TypeInt,
					Optional:    true,
					Description: "Number of retry attempts. Can also be set with the `SPECTROCLOUD_RETRY_ATTEMPTS` environment variable. Defaults to 10.",
					DefaultFunc: schema.EnvDefaultFunc("SPECTROCLOUD_RETRY_ATTEMPTS", 10),
				},
				"project_name": {
					Type:     schema.TypeString,
					Optional: true,
					Default:  "Default",
					// cannot be empty
					ValidateFunc: validation.StringIsNotEmpty,
					Description:  "The Palette project the provider will target. If no value is provided, the `Default` Palette project is used. The default value is `Default`.",
				},
				"ignore_insecure_tls_error": {
					Type:        schema.TypeBool,
					Optional:    true,
					Description: "Ignore insecure TLS errors for Spectro Cloud API endpoints. Defaults to false.",
				},
			},
			ResourcesMap: map[string]*schema.Resource{
				"spectrocloud_team": resourceTeam(),

				"spectrocloud_project": resourceProject(),

				"spectrocloud_macro": resourceMacro(),

				"spectrocloud_macros": resourceMacros(),

				"spectrocloud_filter": resourceFilter(),

				"spectrocloud_application_profile":    resourceApplicationProfile(),
				"spectrocloud_cluster_profile":        resourceClusterProfile(),
				"spectrocloud_cluster_profile_import": resourceClusterProfileImportFeature(),

				"spectrocloud_cloudaccount_aws": resourceCloudAccountAws(),
				"spectrocloud_cluster_aws":      resourceClusterAws(),

				"spectrocloud_cloudaccount_maas": resourceCloudAccountMaas(),
				"spectrocloud_cluster_maas":      resourceClusterMaas(),

				"spectrocloud_cluster_eks": resourceClusterEks(),

				"spectrocloud_cloudaccount_coxedge": resourceCloudAccountCoxEdge(),
				"spectrocloud_cluster_coxedge":      resourceClusterCoxEdge(),

				"spectrocloud_cloudaccount_tencent": resourceCloudAccountTencent(),
				"spectrocloud_cluster_tke":          resourceClusterTke(),

				"spectrocloud_cloudaccount_azure": resourceCloudAccountAzure(),
				"spectrocloud_cluster_azure":      resourceClusterAzure(),

				"spectrocloud_cluster_aks": resourceClusterAks(),

				"spectrocloud_cloudaccount_gcp": resourceCloudAccountGcp(),
				"spectrocloud_cluster_gcp":      resourceClusterGcp(),

				"spectrocloud_cloudaccount_openstack": resourceCloudAccountOpenstack(),
				"spectrocloud_cluster_openstack":      resourceClusterOpenStack(),

				"spectrocloud_cloudaccount_vsphere": resourceCloudAccountVsphere(),
				"spectrocloud_cluster_vsphere":      resourceClusterVsphere(),

				"spectrocloud_cluster_libvirt": resourceClusterLibvirt(),

				"spectrocloud_cluster_edge_native": resourceClusterEdgeNative(),

				"spectrocloud_cluster_edge_vsphere": resourceClusterEdgeVsphere(),

				"spectrocloud_virtual_cluster": resourceClusterVirtual(),

				"spectrocloud_cluster_group": resourceClusterGroup(),

				"spectrocloud_cluster_import": resourceClusterImport(),

				"spectrocloud_addon_deployment": resourceAddonDeployment(),

				"spectrocloud_virtual_machine": resourceKubevirtVirtualMachine(),

				"spectrocloud_datavolume": resourceKubevirtDataVolume(),

				"spectrocloud_application": resourceApplication(),

				"spectrocloud_privatecloudgateway_ippool": resourcePrivateCloudGatewayIpPool(),

				"spectrocloud_backup_storage_location": resourceBackupStorageLocation(),

				"spectrocloud_registry_oci":  resourceRegistryOciEcr(),
				"spectrocloud_registry_helm": resourceRegistryHelm(),

				"spectrocloud_appliance": resourceAppliance(),

				"spectrocloud_workspace": resourceWorkspace(),
				"spectrocloud_alert":     resourceAlert(),
			},
			DataSourcesMap: map[string]*schema.Resource{
				"spectrocloud_user":    dataSourceUser(),
				"spectrocloud_project": dataSourceProject(),

				"spectrocloud_filter": dataSourceFilter(),

				"spectrocloud_role": dataSourceRole(),

				"spectrocloud_pack":        dataSourcePack(),
				"spectrocloud_pack_simple": dataSourcePackSimple(),

				"spectrocloud_cluster_profile": dataSourceClusterProfile(),

				"spectrocloud_cloudaccount_aws":       dataSourceCloudAccountAws(),
				"spectrocloud_cloudaccount_coxedge":   dataSourceCloudAccountCoxedge(),
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
				"spectrocloud_appliances":          dataSourceAppliances(),
				"spectrocloud_cluster":             dataSourceCluster(),
				"spectrocloud_cluster_group":       dataSourceClusterGroup(),
				"spectrocloud_application_profile": dataSourceApplicationProfile(),
				"spectrocloud_workspace":           dataSourceWorkspace(),

				"spectrocloud_private_cloud_gateway": dataSourcePCG(),
				"spectrocloud_ippool":                dataSourcePrivateCloudGatewayIpPool(),
			},
			ConfigureContextFunc: providerConfigure,
		}

		return p
	}
}

func providerConfigure(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	host := d.Get("host").(string)
	apiKey := ""
	transportDebug := false
	retryAttempts := 10

	if d.Get("trace") != nil {
		transportDebug = d.Get("trace").(bool)
	}

	if d.Get("retry_attempts") != nil {
		retryAttempts = d.Get("retry_attempts").(int)
	}

	if d.Get("api_key") != nil {
		apiKey = d.Get("api_key").(string)
	}
	projectName := d.Get("project_name").(string)
	ignoreTlsError := d.Get("ignore_insecure_tls_error").(bool)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	if apiKey == "" {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to create Spectro Cloud client",
			Detail:   "Unable to authenticate user for authenticated Spectro Cloud client",
		})
		return nil, diags
	}

	if ignoreTlsError {
		http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	c := client.New(
		client.WithHubbleURI(host),
		client.WithAPIKey(apiKey),
		client.WithRetries(retryAttempts))

	if transportDebug {
		client.WithTransportDebug()(c)
	}
	uid, err := c.GetProjectUID(projectName)
	if err != nil {
		return nil, diag.FromErr(err)
	}

	if uid != "" {
		client.WithProjectUID(uid)(c)
	}

	return c, diags

}
