package spectrocloud

type State string

const (
	Running      State = "Running"
	Pending      State = "Pending"
	Provisioning State = "Provisioning"
	Deleting     State = "Deleting"
	Importing    State = "Importing"
	Deleted      State = "Deleted"
)

type CloudType string

const (
	CloudTypeAWS     CloudType = "aws"
	CloudTypeGCP     CloudType = "gcp"
	CloudTypeAzure   CloudType = "azure"
	CloudTypeVsphere CloudType = "vsphere"
)

//output
const (
	Name         = "name"
	Count        = "count"
	DiskSizeInGb = "disk_size_gb"

	Pack   = "pack"
	Tag    = "tag"
	Values = "values"

	UpdateStrategy        = "update_strategy"
	InstanceType          = "instance_type"
	AvailabilityZones     = "azs"
	RollingUpdateScaleOut = "RollingUpdateScaleOut"

	ClusterProfileId = "cluster_profile_id"
	CloudAccountId   = "cloud_account_id"

	CloudConfig   = "cloud_config"
	CloudConfigId = "cloud_config_id"
	Cloud         = "cloud"
	Kubeconfig    = "kubeconfig"
	MachinePool   = "machine_pool"

	ControlPlane         = "control_plane"
	ControlPlaneAsWorker = "control_plane_as_worker"

	ClusterImportManifest    = "cluster_import_manifest"
	ClusterImportManifestUrl = "cluster_import_manifest_url"

	Network = "network"
	Project = "project"
	Region  = "region"
)
