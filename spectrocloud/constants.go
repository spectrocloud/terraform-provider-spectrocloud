package spectrocloud

type State string

const (
	running      State = "Running"
	pending      State = "Pending"
	provisioning State = "Provisioning"
	deleting     State = "Deleting"
	importing    State = "Importing"
	deleted      State = "Deleted"
)

var (
	cloud_types = []string{string(cloud_type_aws), string(cloud_type_gcp), string(cloud_type_azure), string(cloud_type_vsphere)}
)

type CloudType string

const (
	cloud_type_aws     CloudType = "aws"
	cloud_type_gcp     CloudType = "gcp"
	cloud_type_azure   CloudType = "azure"
	cloud_type_vsphere CloudType = "vsphere"
)

//cluster profile type
const (
	cluster_type = "cluster"
	infra_type   = "infra"
	add_on_type  = "add-on"
)

const (
	id           = "id"
	name         = "name"
	description  = "description"
	count        = "count"
	disk_size_gb = "disk_size_gb"
	size_gb      = "size_gb"
	memory_mb    = "memory_mb"
	cpu          = "cpu"
	master       = "master"
	cluster      = "cluster"

	pack          = "pack"
	tag           = "tag"
	values        = "values"
	resource_pool = "resource_pool"

	filters      = "filters"
	version      = "version"
	registry_uid = "registry_uid"
	cp_type      = "type"

	update_strategy          = "update_strategy"
	instance_type            = "instance_type"
	azs                      = "azs"
	rolling_update_scale_out = "RollingUpdateScaleOut"

	cluster_prrofile_id = "cluster_profile_id"
	cloud_account_id    = "cloud_account_id"

	cloud_config    = "cloud_config"
	cloud_config_id = "cloud_config_id"
	cloud           = "cloud"
	kubeconfig      = "kubeconfig"
	machine_pool    = "machine_pool"
	ssh_key_name    = "ssh_key_name"

	control_plane           = "control_plane"
	control_plane_as_worker = "control_plane_as_worker"

	cluster_import_manifest     = "cluster_import_manifest"
	cluster_import_manifest_url = "cluster_import_manifest_url"

	network = "network"
	project = "project"
	region  = "region"

	os_patch_on_boot  = "os_patch_on_boot"
	os_patch_schedule = "os_patch_schedule"
	os_patch_after    = "os_patch_after"

	//aws
	aws_access_key = "aws_access_key"
	aws_secret_key = "aws_secret_key"

	//azure
	disk                = "disk"
	disk_type           = "type"
	ssh_key             = "ssh_key"
	azure_tenant_id     = "azure_tenant_id"
	azure_client_id     = "azure_client_id"
	azure_client_secret = "azure_client_secret"
	subscription_id     = "subscription_id"
	resource_group      = "resource_group"

	//gcp
	gcp_json_credentials = "gcp_json_credentials"

	//vsphere
	private_cloud_gateway_id      = "private_cloud_gateway_id"
	vsphere_vcenter               = "vsphere_vcenter"
	vsphere_username              = "vsphere_username"
	vsphere_password              = "vsphere_password"
	vsphere_ignore_insecure_error = "vsphere_ignore_insecure_error"
	datastore                     = "datastore"
	static_ip_pool_id             = "static_ip_pool_id"
	placement                     = "placement"
	folder                        = "folder"
	network_type                  = "network_type"
	datacenter                    = "datacenter"
	static_ip                     = "static_ip"
	network_search_domain         = "network_search_domain"
)
