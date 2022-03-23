package spectrocloud

import (
	"context"
	"crypto/tls"
	"net/http"

	"github.com/spectrocloud/terraform-provider-spectrocloud/pkg/client"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func New(_ string) func() *schema.Provider {
	return func() *schema.Provider {
		p := &schema.Provider{
			Schema: map[string]*schema.Schema{
				"host": &schema.Schema{
					Type:        schema.TypeString,
					Optional:    true,
					DefaultFunc: schema.EnvDefaultFunc("SPECTROCLOUD_HOST", "api.spectrocloud.com"),
				},
				"username": &schema.Schema{
					Type:        schema.TypeString,
					Optional:    true,
					DefaultFunc: schema.EnvDefaultFunc("SPECTROCLOUD_USERNAME", nil),
				},
				"password": &schema.Schema{
					Type:        schema.TypeString,
					Optional:    true,
					Sensitive:   true,
					DefaultFunc: schema.EnvDefaultFunc("SPECTROCLOUD_PASSWORD", nil),
				},
				"api_key": &schema.Schema{
					Type:        schema.TypeString,
					Optional:    true,
					Sensitive:   true,
					DefaultFunc: schema.EnvDefaultFunc("SPECTROCLOUD_APIKEY", nil),
				},
				"project_name": &schema.Schema{
					Type:     schema.TypeString,
					Optional: true,
				},
				"ignore_insecure_tls_error": &schema.Schema{
					Type:     schema.TypeBool,
					Optional: true,
				},
			},
			ResourcesMap: map[string]*schema.Resource{
				"spectrocloud_team": resourceTeam(),

				"spectrocloud_project": resourceProject(),

				"spectrocloud_cluster_profile": resourceClusterProfile(),

				"spectrocloud_cloudaccount_aws": resourceCloudAccountAws(),
				"spectrocloud_cluster_aws":      resourceClusterAws(),

				"spectrocloud_cloudaccount_maas": resourceCloudAccountMaas(),
				"spectrocloud_cluster_maas":      resourceClusterMaas(),

				"spectrocloud_cluster_eks": resourceClusterEks(),

				"spectrocloud_cloudaccount_azure": resourceCloudAccountAzure(),
				"spectrocloud_cluster_azure":      resourceClusterAzure(),

				"spectrocloud_cluster_aks": resourceClusterAks(),

				"spectrocloud_cloudaccount_gcp": resourceCloudAccountGcp(),
				"spectrocloud_cluster_gcp":      resourceClusterGcp(),

				"spectrocloud_cloudaccount_openstack": resourceCloudAccountOpenstack(),
				"spectrocloud_cluster_openstack":      resourceClusterOpenStack(),

				"spectrocloud_cluster_vsphere": resourceClusterVsphere(),

				"spectrocloud_cluster_libvirt": resourceClusterLibvirt(),

				"spectrocloud_cluster_edge": resourceClusterEdge(),

				"spectrocloud_cluster_import": resourceClusterImport(),

				"spectrocloud_privatecloudgateway_ippool": resourcePrivateCloudGatewayIpPool(),

				"spectrocloud_backup_storage_location": resourceBackupStorageLocation(),

				"spectrocloud_registry_oci": resourceRegistryOciEcr(),

				"spectrocloud_appliance": resourceAppliance(),
			},
			DataSourcesMap: map[string]*schema.Resource{
				"spectrocloud_user":    dataSourceUser(),
				"spectrocloud_project": dataSourceProject(),
				"spectrocloud_role":    dataSourceRole(),

				"spectrocloud_pack": dataSourcePack(),

				"spectrocloud_cluster_profile": dataSourceClusterProfile(),

				"spectrocloud_cloudaccount_aws":       dataSourceCloudAccountAws(),
				"spectrocloud_cloudaccount_azure":     dataSourceCloudAccountAzure(),
				"spectrocloud_cloudaccount_gcp":       dataSourceCloudAccountGcp(),
				"spectrocloud_cloudaccount_vsphere":   dataSourceCloudAccountVsphere(),
				"spectrocloud_cloudaccount_openstack": dataSourceCloudAccountOpenStack(),
				"spectrocloud_cloudaccount_maas":      dataSourceCloudAccountMaas(),

				"spectrocloud_backup_storage_location": dataSourceBackupStorageLocation(),

				"spectrocloud_registry_pack": dataSourceRegistryPack(),
				"spectrocloud_registry_helm": dataSourceRegistryHelm(),
				"spectrocloud_registry_oci":  dataSourceRegistryOci(),

				"spectrocloud_appliance": dataSourceAppliance(),
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

	if d.Get("transport_debug") != nil {
		transportDebug = d.Get("transport_debug").(bool)
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

	c := client.New(host, username, password, "", apiKey, transportDebug)

	if projectName != "" {
		uid, err := c.GetProjectUID(projectName)
		if err != nil {
			return nil, diag.FromErr(err)
		}

		c = client.New(host, username, password, uid, apiKey, transportDebug)
	}

	return c, diags

}
