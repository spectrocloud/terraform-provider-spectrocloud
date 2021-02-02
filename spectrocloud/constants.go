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

type CloudType string

const (
	cloud_type_aws     CloudType = "aws"
	cloud_type_gcp     CloudType = "gcp"
	cloud_type_azure   CloudType = "azure"
	cloud_type_vsphere CloudType = "vsphere"
)

//output
const (
	name            = "name"
	count           = "count"
	disk_size_in_gb = "disk_size_gb"

	pack   = "pack"
	tag    = "tag"
	values = "values"

	update_strategy          = "update_strategy"
	instance_type            = "instance_type"
	availability_zones       = "azs"
	rolling_update_scale_out = "rolling_update_scale_out"

	cluster_prrofile_id = "cluster_profile_id"
	cloud_account_id    = "cloud_account_id"

	cloud_config    = "cloud_config"
	cloud_config_id = "cloud_config_id"
	cloud           = "cloud"
	kubeconfig      = "kubeconfig"
	machine_pool    = "machine_pool"

	control_plane           = "control_plane"
	control_plane_as_worker = "control_plane_as_worker"

	cluster_import_manifest     = "cluster_import_manifest"
	cluster_import_manifest_url = "cluster_import_manifest_url"

	network = "network"
	project = "project"
	region  = "region"
)
