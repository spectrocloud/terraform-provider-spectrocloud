package virtualmachineinstance

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	kubevirtapiv1 "kubevirt.io/api/core/v1"

	"github.com/spectrocloud/terraform-provider-spectrocloud/spectrocloud/kubevirt/schema/k8s"
)

func volumesFields() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"name": {
			Type:        schema.TypeString,
			Description: "Volume's name.",
			Required:    true,
		},
		"volume_source": {
			Type:        schema.TypeList,
			Description: "VolumeSource represents the location and type of the mounted volume. Defaults to Disk, if no type is specified.",
			MaxItems:    1,
			Required:    true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"data_volume": {
						Type:        schema.TypeList,
						Description: "DataVolume represents the dynamic creation a PVC for this volume as well as the process of populating that PVC with a disk image.",
						MaxItems:    1,
						Optional:    true,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"name": {
									Type:        schema.TypeString,
									Description: "Name represents the name of the DataVolume in the same namespace.",
									Required:    true,
								},
							},
						},
					},
					"cloud_init_config_drive": {
						Type:        schema.TypeList,
						Description: "CloudInitConfigDrive represents a cloud-init Config Drive user-data source.",
						MaxItems:    1,
						Optional:    true,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"user_data_secret_ref": k8s.LocalObjectReferenceSchema("UserDataSecretRef references a k8s secret that contains config drive userdata."),
								"user_data_base64": {
									Type:        schema.TypeString,
									Description: "UserDataBase64 contains config drive cloud-init userdata as a base64 encoded string.",
									Optional:    true,
								},
								"user_data": {
									Type:        schema.TypeString,
									Description: "UserData contains config drive inline cloud-init userdata.",
									Optional:    true,
								},
								"network_data_secret_ref": k8s.LocalObjectReferenceSchema("NetworkDataSecretRef references a k8s secret that contains config drive networkdata."),
								"network_data_base64": {
									Type:        schema.TypeString,
									Description: "NetworkDataBase64 contains config drive cloud-init networkdata as a base64 encoded string.",
									Optional:    true,
								},
								"network_data": {
									Type:        schema.TypeString,
									Description: "NetworkData contains config drive inline cloud-init networkdata.",
									Optional:    true,
								},
							},
						},
					},
					"service_account": {
						Type:        schema.TypeList,
						Description: "ServiceAccountVolumeSource represents a reference to a service account.",
						MaxItems:    1,
						Optional:    true,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"service_account_name": {
									Type:        schema.TypeString,
									Description: "Name of the service account in the pod's namespace to use.",
									Required:    true,
								},
							},
						},
					},
					"container_disk": {
						Type:     schema.TypeSet,
						Optional: true,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"image_url": {
									Type:        schema.TypeString,
									Required:    true,
									Description: "The URL of the container image to use as the disk. This can be a local file path, a remote URL, or a registry URL.",
								},
							},
						},
						Description: "A container disk is a disk that is backed by a container image. The container image is expected to contain a disk image in a supported format. The disk image is extracted from the container image and used as the disk for the VM.",
					},
					"cloud_init_no_cloud": {
						Type:     schema.TypeSet,
						Optional: true,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"user_data": {
									Type:        schema.TypeString,
									Required:    true,
									Description: "The user data to use for the cloud-init no cloud disk. This can be a local file path, a remote URL, or a registry URL.",
								},
							},
						},
						Description: "Used to specify a cloud-init `noCloud` image. The image is expected to contain a disk image in a supported format. The disk image is extracted from the cloud-init `noCloud `image and used as the disk for the VM",
					},
					"ephemeral": {
						Type:        schema.TypeList,
						Description: "EphemeralVolumeSource represents a volume that is populated with the contents of a pod. Ephemeral volumes do not support ownership management or SELinux relabeling.",
						MaxItems:    1,
						Optional:    true,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"persistent_volume_claim": {
									Type:        schema.TypeList,
									Description: "PersistentVolumeClaimVolumeSource represents a reference to a PersistentVolumeClaim in the same namespace.",
									MaxItems:    1,
									Optional:    true,
									Elem: &schema.Resource{
										Schema: map[string]*schema.Schema{
											"claim_name": {
												Type:        schema.TypeString,
												Description: "ClaimName is the name of a PersistentVolumeClaim in the same namespace as the pod using this volume. More info: https://kubernetes.io/docs/concepts/storage/persistent-volumes#persistentvolumeclaims",
												Required:    true,
											},
											"read_only": {
												Type:        schema.TypeBool,
												Description: "Will force the ReadOnly setting in VolumeMounts. Default false.",
												Optional:    true,
											},
										},
									},
								},
							},
						},
					},
					"empty_disk": {
						Type:        schema.TypeList,
						Description: "EmptyDisk represents a temporary disk which shares the VM's lifecycle.",
						MaxItems:    1,
						Optional:    true,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"capacity": {
									Type:        schema.TypeString,
									Description: "Capacity of the sparse disk.",
									Required:    true,
								},
							},
						},
					},
					"persistent_volume_claim": {
						Type:        schema.TypeList,
						Description: "PersistentVolumeClaimVolumeSource represents a reference to a PersistentVolumeClaim in the same namespace.",
						MaxItems:    1,
						Optional:    true,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"claim_name": {
									Type:        schema.TypeString,
									Description: "ClaimName is the name of a PersistentVolumeClaim in the same namespace as the pod using this volume. More info: https://kubernetes.io/docs/concepts/storage/persistent-volumes#persistentvolumeclaims",
									Required:    true,
								},
								"read_only": {
									Type:        schema.TypeBool,
									Description: "Will force the ReadOnly setting in VolumeMounts. Default false.",
									Optional:    true,
								},
							},
						},
					},
					"host_disk": {
						Type:        schema.TypeList,
						Description: "HostDisk represents a disk created on the host.",
						MaxItems:    1,
						Optional:    true,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"path": {
									Type:        schema.TypeString,
									Description: "Path of the disk.",
									Required:    true,
								},
								"type": {
									Type:        schema.TypeString,
									Description: "Type of the disk, supported values are disk, directory, socket, char, block.",
									Required:    true,
								},
							},
						},
					},
					"config_map": {
						Type:        schema.TypeList,
						Description: "ConfigMapVolumeSource adapts a ConfigMap into a volume.",
						MaxItems:    1,
						Optional:    true,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"default_mode": {
									Type:        schema.TypeInt,
									Description: "Optional: mode bits to use on created files by default. Must be a value between 0 and 0777. Defaults to 0644. Directories within the path are not affected by this setting. This might be in conflict with other options that affect the file mode, like fsGroup, and the result can be other mode bits set.",
									Optional:    true,
								},
								"items": {
									Type:        schema.TypeList,
									Description: "If unspecified, each key-value pair in the Data field of the referenced ConfigMap will be projected into the volume as a file whose name is the key and content is the value. If specified, the listed keys will be projected into the specified paths, and unlisted keys will not be present. If a key is specified which is not present in the ConfigMap, the volume setup will error unless it is marked optional. Paths must be relative and may not contain the '..' path or start with '..'.",
									Optional:    true,
									Elem: &schema.Resource{
										Schema: map[string]*schema.Schema{
											"key": {
												Type:     schema.TypeString,
												Optional: true,
											},
										},
									},
								},
							},
						},
					},
					// TODO nargaman - Add other data volume source types
				},
			},
		},
	}
}

func volumesSchema() *schema.Schema {
	fields := volumesFields()

	return &schema.Schema{
		Type: schema.TypeList,

		Description: fmt.Sprintf("Specification of the desired behavior of the VirtualMachineInstance on the host."),
		Optional:    true,
		Elem: &schema.Resource{
			Schema: fields,
		},
	}

}

func expandVolumes(volumes []interface{}) []kubevirtapiv1.Volume {
	result := make([]kubevirtapiv1.Volume, len(volumes))

	if len(volumes) == 0 || volumes[0] == nil {
		return result
	}

	for i, condition := range volumes {
		in := condition.(map[string]interface{})

		if v, ok := in["name"].(string); ok {
			result[i].Name = v
		}
		if v, ok := in["volume_source"].([]interface{}); ok {
			result[i].VolumeSource = expandVolumeSource(v)
		}
	}

	return result
}

func expandVolumeSource(volumeSource []interface{}) kubevirtapiv1.VolumeSource {
	result := kubevirtapiv1.VolumeSource{}

	if len(volumeSource) == 0 || volumeSource[0] == nil {
		return result
	}

	in := volumeSource[0].(map[string]interface{})

	if v, ok := in["data_volume"].([]interface{}); ok {
		result.DataVolume = expandDataVolume(v)
	}
	if v, ok := in["cloud_init_config_drive"].([]interface{}); ok {
		result.CloudInitConfigDrive = expandCloudInitConfigDrive(v)
	}
	if v, ok := in["service_account"].([]interface{}); ok {
		result.ServiceAccount = expandServiceAccount(v)
	}
	if v, ok := in["container_disk"].(*schema.Set); ok {
		result.ContainerDisk = expandContainerDisk(v.List())
	}
	if v, ok := in["cloud_init_no_cloud"].(*schema.Set); ok {
		result.CloudInitNoCloud = expandCloudInitNoCloud(v.List())
	}
	if v, ok := in["ephemeral"].([]interface{}); ok {
		result.Ephemeral = expandEphemeral(v)
	}
	if v, ok := in["empty_disk"].([]interface{}); ok {
		result.EmptyDisk = expandEmptyDisk(v)
	}
	if v, ok := in["host_disk"].([]interface{}); ok {
		result.HostDisk = expandHostDisk(v)
	}
	if v, ok := in["config_map"].([]interface{}); ok {
		result.ConfigMap = expandConfigMap(v)
	}

	return result
}

func expandDataVolume(dataVolumeSource []interface{}) *kubevirtapiv1.DataVolumeSource {
	if len(dataVolumeSource) == 0 || dataVolumeSource[0] == nil {
		return nil
	}

	result := &kubevirtapiv1.DataVolumeSource{}
	in := dataVolumeSource[0].(map[string]interface{})

	if v, ok := in["name"].(string); ok {
		result.Name = v
	}

	return result
}

func expandCloudInitConfigDrive(cloudInitConfigDriveSource []interface{}) *kubevirtapiv1.CloudInitConfigDriveSource {
	if len(cloudInitConfigDriveSource) == 0 || cloudInitConfigDriveSource[0] == nil {
		return nil
	}

	result := &kubevirtapiv1.CloudInitConfigDriveSource{}
	in := cloudInitConfigDriveSource[0].(map[string]interface{})

	if v, ok := in["user_data_secret_ref"].([]interface{}); ok {
		result.UserDataSecretRef = k8s.ExpandLocalObjectReferences(v)
	}
	if v, ok := in["user_data_base64"].(string); ok {
		result.UserDataBase64 = v
	}
	if v, ok := in["user_data"].(string); ok {
		result.UserData = v
	}
	if v, ok := in["network_data_secret_ref"].([]interface{}); ok {
		result.NetworkDataSecretRef = k8s.ExpandLocalObjectReferences(v)
	}
	if v, ok := in["network_data_base64"].(string); ok {
		result.NetworkDataBase64 = v
	}
	if v, ok := in["network_data"].(string); ok {
		result.NetworkData = v
	}

	return result
}

func expandServiceAccount(serviceAccountSource []interface{}) *kubevirtapiv1.ServiceAccountVolumeSource {
	if len(serviceAccountSource) == 0 || serviceAccountSource[0] == nil {
		return nil
	}

	result := &kubevirtapiv1.ServiceAccountVolumeSource{}
	in := serviceAccountSource[0].(map[string]interface{})

	if v, ok := in["service_account_name"].(string); ok {
		result.ServiceAccountName = v
	}

	return result
}

func expandContainerDisk(containerDiskSource []interface{}) *kubevirtapiv1.ContainerDiskSource {
	if len(containerDiskSource) == 0 || containerDiskSource[0] == nil {
		return nil
	}

	result := &kubevirtapiv1.ContainerDiskSource{}
	in := containerDiskSource[0].(map[string]interface{})

	if v, ok := in["image_url"].(string); ok {
		result.Image = v
	}

	return result
}

func expandCloudInitNoCloud(cloudInitNoCloudSource []interface{}) *kubevirtapiv1.CloudInitNoCloudSource {
	if len(cloudInitNoCloudSource) == 0 || cloudInitNoCloudSource[0] == nil {
		return nil
	}

	result := &kubevirtapiv1.CloudInitNoCloudSource{}
	in := cloudInitNoCloudSource[0].(map[string]interface{})

	if v, ok := in["user_data"].(string); ok {
		result.UserData = v
	}

	return result
}

func expandEphemeral(ephemeralSource []interface{}) *kubevirtapiv1.EphemeralVolumeSource {
	if len(ephemeralSource) == 0 || ephemeralSource[0] == nil {
		return nil
	}

	result := &kubevirtapiv1.EphemeralVolumeSource{}
	in := ephemeralSource[0].(map[string]interface{})

	if v, ok := in["persistent_volume_claim"].([]interface{}); ok {
		result.PersistentVolumeClaim = expandPersistentVolumeClaim(v)
	}

	return result

}

func expandPersistentVolumeClaim(persistentVolumeClaimSource []interface{}) *v1.PersistentVolumeClaimVolumeSource {
	if len(persistentVolumeClaimSource) == 0 || persistentVolumeClaimSource[0] == nil {
		return nil
	}

	/*
		type PersistentVolumeClaimVolumeSource struct {
			v1.PersistentVolumeClaimVolumeSource `json:",inline"`
			// Hotpluggable indicates whether the volume can be hotplugged and hotunplugged.
			// +optional
			Hotpluggable bool `json:"hotpluggable,omitempty"`
		}

		type PersistentVolumeClaimVolumeSource struct {
			// claimName is the name of a PersistentVolumeClaim in the same namespace as the pod using this volume.
			// More info: https://kubernetes.io/docs/concepts/storage/persistent-volumes#persistentvolumeclaims
			ClaimName string `json:"claimName" protobuf:"bytes,1,opt,name=claimName"`
			// readOnly Will force the ReadOnly setting in VolumeMounts.
			// Default false.
			// +optional
			ReadOnly bool `json:"readOnly,omitempty" protobuf:"varint,2,opt,name=readOnly"`
		}

	*/

	result := &v1.PersistentVolumeClaimVolumeSource{}
	in := persistentVolumeClaimSource[0].(map[string]interface{})

	if v, ok := in["claim_name"].(string); ok {
		result.ClaimName = v
	}
	if v, ok := in["read_only"].(bool); ok {
		result.ReadOnly = v
	}

	return result
}

func expandEmptyDisk(emptyDiskSource []interface{}) *kubevirtapiv1.EmptyDiskSource {
	if len(emptyDiskSource) == 0 || emptyDiskSource[0] == nil {
		return nil
	}

	result := &kubevirtapiv1.EmptyDiskSource{}
	in := emptyDiskSource[0].(map[string]interface{})

	/* type EmptyDiskSource struct {
		// Capacity of the sparse disk.
		Capacity resource.Quantity `json:"capacity"`
	}
	type Quantity struct {
		// i is the quantity in int64 scaled form, if d.Dec == nil
		i int64Amount
		// d is the quantity in inf.Dec form if d.Dec != nil
		d infDecAmount
		// s is the generated value of this quantity to avoid recalculation
		s string

		// Change Format at will. See the comment for Canonicalize for
		// more details.
		Format
	}

	*/
	if v, ok := in["capacity"].(string); ok {
		result.Capacity = resource.MustParse(v)
	}

	return result
}

func expandHostDisk(hostDiskSource []interface{}) *kubevirtapiv1.HostDisk {
	if len(hostDiskSource) == 0 || hostDiskSource[0] == nil {
		return nil
	}

	result := &kubevirtapiv1.HostDisk{}
	in := hostDiskSource[0].(map[string]interface{})

	if v, ok := in["path"].(string); ok {
		result.Path = v
	}
	if v, ok := in["type"].(string); ok {
		result.Type = kubevirtapiv1.HostDiskType(v)
	}

	return result
}

func expandConfigMap(configMapSource []interface{}) *kubevirtapiv1.ConfigMapVolumeSource {
	if len(configMapSource) == 0 || configMapSource[0] == nil {
		return nil
	}

	result := &kubevirtapiv1.ConfigMapVolumeSource{}

	in := configMapSource[0].(map[string]interface{})
	if v, ok := in["name"].(string); ok {
		result.Name = v
	}
	if v, ok := in["optional"].(bool); ok {
		result.Optional = &v
	}
	if v, ok := in["volume_label"].(string); ok {
		result.VolumeLabel = v
	}

	return result
}

func flattenVolumes(in []kubevirtapiv1.Volume) []interface{} {
	att := make([]interface{}, len(in))

	for i, v := range in {
		c := make(map[string]interface{})

		c["name"] = v.Name
		c["volume_source"] = flattenVolumeSource(v.VolumeSource)

		att[i] = c
	}

	return att
}

func flattenVolumeSource(in kubevirtapiv1.VolumeSource) []interface{} {
	att := make(map[string]interface{})

	if in.DataVolume != nil {
		att["data_volume"] = flattenDataVolume(*in.DataVolume)
	}
	if in.CloudInitConfigDrive != nil {
		att["cloud_init_config_drive"] = flattenCloudInitConfigDrive(*in.CloudInitConfigDrive)
	}
	if in.ServiceAccount != nil {
		att["service_account"] = flattenServiceAccount(*in.ServiceAccount)
	}
	if in.ContainerDisk != nil {
		att["container_disk"] = flattenContainerDisk(*in.ContainerDisk)
	}
	if in.CloudInitNoCloud != nil {
		att["cloud_init_no_cloud"] = flattenCloudInitNoCloud(*in.CloudInitNoCloud)
	}
	if in.Ephemeral != nil {
		att["ephemeral"] = flattenEphemeral(*in.Ephemeral)
	}
	if in.EmptyDisk != nil {
		att["empty_disk"] = flattenEmptyDisk(*in.EmptyDisk)
	}
	if in.HostDisk != nil {
		att["host_disk"] = flattenHostDisk(*in.HostDisk)
	}
	if in.PersistentVolumeClaim != nil {
		//att["persistent_volume_claim"] = flattenPersistentVolumeClaim(in.PersistentVolumeClaim)
	}
	if in.ConfigMap != nil {
		att["config_map"] = flattenConfigMap(*in.ConfigMap)
	}

	return []interface{}{att}
}

func flattenDataVolume(in kubevirtapiv1.DataVolumeSource) []interface{} {
	att := make(map[string]interface{})

	att["name"] = in.Name

	return []interface{}{att}
}

func flattenCloudInitConfigDrive(in kubevirtapiv1.CloudInitConfigDriveSource) []interface{} {
	att := make(map[string]interface{})

	if in.UserDataSecretRef != nil {
		att["user_data_secret_ref"] = k8s.FlattenLocalObjectReferences(*in.UserDataSecretRef)
	}
	att["user_data_base64"] = in.UserDataBase64
	att["user_data"] = in.UserData
	if in.NetworkDataSecretRef != nil {
		att["network_data_secret_ref"] = k8s.FlattenLocalObjectReferences(*in.NetworkDataSecretRef)
	}
	att["network_data_base64"] = in.NetworkDataBase64
	att["network_data"] = in.NetworkData

	return []interface{}{att}
}

func flattenServiceAccount(in kubevirtapiv1.ServiceAccountVolumeSource) []interface{} {
	att := make(map[string]interface{})

	att["service_account_name"] = in.ServiceAccountName

	return []interface{}{att}
}

func flattenContainerDisk(in kubevirtapiv1.ContainerDiskSource) []interface{} {
	att := make(map[string]interface{})

	att["image_url"] = in.Image

	return []interface{}{att}
}

func flattenCloudInitNoCloud(in kubevirtapiv1.CloudInitNoCloudSource) []interface{} {
	att := make(map[string]interface{})

	att["user_data"] = in.UserData

	return []interface{}{att}
}

func flattenEphemeral(in kubevirtapiv1.EphemeralVolumeSource) []interface{} {
	att := make(map[string]interface{})

	if in.PersistentVolumeClaim != nil {
		att["persistent_volume_claim"] = flattenPersistentVolumeClaim(*in.PersistentVolumeClaim)
	}

	return []interface{}{att}
}

func flattenPersistentVolumeClaim(in v1.PersistentVolumeClaimVolumeSource) []interface{} {
	att := make(map[string]interface{})

	att["claim_name"] = in.ClaimName

	return []interface{}{att}
}

func flattenEmptyDisk(in kubevirtapiv1.EmptyDiskSource) []interface{} {
	att := make(map[string]interface{})

	att["capacity"] = in.Capacity

	return []interface{}{att}
}

func flattenHostDisk(in kubevirtapiv1.HostDisk) []interface{} {
	att := make(map[string]interface{})

	att["path"] = in.Path
	att["type"] = in.Type

	return []interface{}{att}
}

func flattenPVC(in kubevirtapiv1.PersistentVolumeClaimVolumeSource) []interface{} {
	att := make(map[string]interface{})

	att["claim_name"] = in.ClaimName

	return []interface{}{att}
}

func flattenConfigMap(in kubevirtapiv1.ConfigMapVolumeSource) []interface{} {
	att := make(map[string]interface{})

	att["name"] = in.Name

	return []interface{}{att}
}
