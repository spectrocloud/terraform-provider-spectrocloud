package virtualmachineinstance

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/spectrocloud/palette-sdk-go/api/models"
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
								"user_data_secret_ref": k8s.LocalObjectReferenceSchema("UserDataSecretRef references a k8s secret that contains cloud-init userdata."),
								"user_data_base64": {
									Type:        schema.TypeString,
									Description: "UserDataBase64 contains cloud-init userdata as a base64 encoded string.",
									Optional:    true,
								},
								"user_data": {
									Type:        schema.TypeString,
									Description: "UserData contains cloud-init inline userdata.",
									Optional:    true,
								},
								"network_data_secret_ref": k8s.LocalObjectReferenceSchema("NetworkDataSecretRef references a k8s secret that contains cloud-init networkdata."),
								"network_data_base64": {
									Type:        schema.TypeString,
									Description: "NetworkDataBase64 contains cloud-init networkdata as a base64 encoded string.",
									Optional:    true,
								},
								"network_data": {
									Type:        schema.TypeString,
									Optional:    true,
									Description: "NetworkData contains cloud-init inline network configuration data.",
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

func VolumesSchema() *schema.Schema {
	fields := volumesFields()

	return &schema.Schema{
		Type: schema.TypeList,

		Description: "Specification of the desired behavior of the VirtualMachineInstance on the host.",
		Optional:    true,
		Elem: &schema.Resource{
			Schema: fields,
		},
	}
}

func expandVolumes(volumes []interface{}) []*models.V1VMVolume {
	result := make([]*models.V1VMVolume, len(volumes))

	if len(volumes) == 0 || volumes[0] == nil {
		return result
	}

	for i, condition := range volumes {
		in := condition.(map[string]interface{})
		result[i] = &models.V1VMVolume{}
		if v, ok := in["name"].(string); ok {
			result[i].Name = &v
		}
		if vs, ok := in["volume_source"].([]interface{}); ok && len(vs) > 0 {
			expandVolumeSourceForHapi(vs[0].(map[string]interface{}), result[i])
		}
		// Hapi missing volume source
		// if v, ok := in["volume_source"].([]interface{}); ok {
		// 	// result[i].VolumeSource = expandVolumeSource(v)
		// }
	}

	return result
}

// expandVolumeSourceForHapi sets exactly one volume source on vol from Terraform volume_source map (container_disk, cloud_init_config_drive, etc.).
func expandVolumeSourceForHapi(m map[string]interface{}, vol *models.V1VMVolume) {
	if vol == nil {
		return
	}
	// container_disk (TypeSet) -> list of one map with image_url
	if v, ok := m["container_disk"]; ok && v != nil {
		var list []interface{}
		switch t := v.(type) {
		case *schema.Set:
			list = t.List()
		case []interface{}:
			list = t
		}
		if len(list) > 0 {
			cd := list[0].(map[string]interface{})
			if img, ok := cd["image_url"].(string); ok && img != "" {
				vol.ContainerDisk = &models.V1VMContainerDiskSource{Image: &img}
			}
		}
	}
	if vol.ContainerDisk != nil {
		return
	}

	// data_volume (TypeList, MaxItems 1) — references a DataVolume / dataVolumeTemplate by name
	if v, ok := m["data_volume"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		dv := v[0].(map[string]interface{})
		if name, ok := dv["name"].(string); ok && name != "" {
			n := name
			vol.DataVolume = &models.V1VMCoreDataVolumeSource{Name: &n}
			return
		}
	}

	// cloud_init_config_drive (TypeList)
	if v, ok := m["cloud_init_config_drive"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		cc := v[0].(map[string]interface{})
		src := &models.V1VMCloudInitConfigDriveSource{}
		if u, ok := cc["user_data"].(string); ok {
			src.UserData = u
		}
		if u, ok := cc["user_data_base64"].(string); ok {
			src.UserDataBase64 = u
		}
		if n, ok := cc["network_data"].(string); ok {
			src.NetworkData = n
		}
		if n, ok := cc["network_data_base64"].(string); ok {
			src.NetworkDataBase64 = n
		}
		vol.CloudInitConfigDrive = src
	}
	// cloud_init_no_cloud (TypeSet)
	// if v, ok := m["cloud_init_no_cloud"]; ok && v != nil {
	// 	var list []interface{}
	// 	switch t := v.(type) {
	// 	case *schema.Set:
	// 		list = t.List()
	// 	case []interface{}:
	// 		list = t
	// 	}
	// 	if len(list) > 0 && list[0] != nil {
	// 		nc := list[0].(map[string]interface{})
	// 		src := &models.V1VMCloudInitNoCloudSource{}
	// 		if u, ok := nc["user_data"].(string); ok {
	// 			src.UserData = u
	// 		}
	// 		if u, ok := nc["user_data_base64"].(string); ok {
	// 			src.UserDataBase64 = u
	// 		}
	// 		if n, ok := nc["network_data"].(string); ok {
	// 			src.NetworkData = n
	// 		}
	// 		if n, ok := nc["network_data_base64"].(string); ok {
	// 			src.NetworkDataBase64 = n
	// 		}
	// 		vol.CloudInitNoCloud = src
	// 	}
	// }

	// cloud_init_config_drive (TypeList)
	if v, ok := m["cloud_init_config_drive"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		cc := v[0].(map[string]interface{})
		src := &models.V1VMCloudInitConfigDriveSource{}
		if u, ok := cc["user_data"].(string); ok {
			src.UserData = u
		}
		if u, ok := cc["user_data_base64"].(string); ok {
			src.UserDataBase64 = u
		}
		if n, ok := cc["network_data"].(string); ok {
			src.NetworkData = n
		}
		if n, ok := cc["network_data_base64"].(string); ok {
			src.NetworkDataBase64 = n
		}
		vol.CloudInitConfigDrive = src
		return
	}
	// cloud_init_no_cloud (TypeSet)
	if v, ok := m["cloud_init_no_cloud"]; ok && v != nil {
		var list []interface{}
		switch t := v.(type) {
		case *schema.Set:
			list = t.List()
		case []interface{}:
			list = t
		}
		if len(list) > 0 && list[0] != nil {
			nc := list[0].(map[string]interface{})
			src := &models.V1VMCloudInitNoCloudSource{}
			if u, ok := nc["user_data"].(string); ok {
				src.UserData = u
			}
			if u, ok := nc["user_data_base64"].(string); ok {
				src.UserDataBase64 = u
			}
			if n, ok := nc["network_data"].(string); ok {
				src.NetworkData = n
			}
			if n, ok := nc["network_data_base64"].(string); ok {
				src.NetworkDataBase64 = n
			}
			vol.CloudInitNoCloud = src
		}
	}
}

// func expandCloudInitNoCloud(cloudInitNoCloudSource []interface{}) *kubevirtapiv1.CloudInitNoCloudSource {
// 	if len(cloudInitNoCloudSource) == 0 || cloudInitNoCloudSource[0] == nil {
// 		return nil
// 	}

// 	result := &kubevirtapiv1.CloudInitNoCloudSource{}
// 	in := cloudInitNoCloudSource[0].(map[string]interface{})

// 	if v, ok := in["user_data_secret_ref"].([]interface{}); ok {
// 		result.UserDataSecretRef = k8s.ExpandLocalObjectReferences(v)
// 	}
// 	if v, ok := in["user_data_base64"].(string); ok {
// 		result.UserDataBase64 = v
// 	}
// 	if v, ok := in["user_data"].(string); ok {
// 		result.UserData = v
// 	}
// 	if v, ok := in["network_data_secret_ref"].([]interface{}); ok {
// 		result.NetworkDataSecretRef = k8s.ExpandLocalObjectReferences(v)
// 	}
// 	if v, ok := in["network_data_base64"].(string); ok {
// 		result.NetworkDataBase64 = v
// 	}
// 	if v, ok := in["network_data"].(string); ok {
// 		result.NetworkData = v
// 	}

// 	return result
// }

// func flattenVolumes(in []kubevirtapiv1.Volume) []interface{} {
// 	att := make([]interface{}, len(in))

// 	for i, v := range in {
// 		c := make(map[string]interface{})

// 		c["name"] = v.Name
// 		c["volume_source"] = flattenVolumeSource(v.VolumeSource)

// 		att[i] = c
// 	}

// 	return att
// }

// flattenVolumesFromVM flattens []*models.V1VMVolume to the same shape as flattenVolumes.
func flattenVolumesFromVM(in []*models.V1VMVolume) []interface{} {
	if len(in) == 0 {
		return nil
	}
	att := make([]interface{}, len(in))
	for i, v := range in {
		if v == nil {
			continue
		}
		c := make(map[string]interface{})
		if v.Name != nil {
			c["name"] = *v.Name
		}
		c["volume_source"] = flattenVolumeSourceFromVM(v)
		att[i] = c
	}
	return att
}

func flattenVolumeSourceFromVM(in *models.V1VMVolume) []interface{} {
	if in == nil {
		return []interface{}{map[string]interface{}{}}
	}
	att := make(map[string]interface{})
	if in.DataVolume != nil {
		name := ""
		if in.DataVolume.Name != nil {
			name = *in.DataVolume.Name
		}
		att["data_volume"] = []interface{}{map[string]interface{}{"name": name}}
	}
	if in.ContainerDisk != nil {
		// att["container_disk"] = []interface{}{map[string]interface{}{"image": in.ContainerDisk.Image}}
		att["container_disk"] = []interface{}{map[string]interface{}{"image_url": in.ContainerDisk.Image}}
	}
	if in.CloudInitNoCloud != nil {
		att["cloud_init_no_cloud"] = []interface{}{map[string]interface{}{"user_data": in.CloudInitNoCloud.UserData, "user_data_base64": in.CloudInitNoCloud.UserDataBase64}}
	}
	if in.CloudInitConfigDrive != nil {
		c := in.CloudInitConfigDrive
		cc := map[string]interface{}{
			"user_data":           c.UserData,
			"user_data_base64":    c.UserDataBase64,
			"network_data":        c.NetworkData,
			"network_data_base64": c.NetworkDataBase64,
		}
		// Optional: flatten secret refs if your schema uses them and you have a helper:
		// if c.SecretRef != nil { cc["user_data_secret_ref"] = ... }
		// if c.NetworkDataSecretRef != nil { cc["network_data_secret_ref"] = ... }
		att["cloud_init_config_drive"] = []interface{}{cc}
	}
	if in.Ephemeral != nil {
		att["ephemeral"] = []interface{}{map[string]interface{}{}}
	}
	if in.EmptyDisk != nil {
		att["empty_disk"] = []interface{}{map[string]interface{}{"capacity": in.EmptyDisk.Capacity}}
	}
	if in.HostDisk != nil {
		att["host_disk"] = []interface{}{map[string]interface{}{"path": in.HostDisk.Path, "type": in.HostDisk.Type}}
	}
	return []interface{}{att}
}

// func flattenVolumeSource(in kubevirtapiv1.VolumeSource) []interface{} {
// 	att := make(map[string]interface{})

// 	if in.DataVolume != nil {
// 		att["data_volume"] = flattenDataVolume(*in.DataVolume)
// 	}
// 	if in.CloudInitConfigDrive != nil {
// 		att["cloud_init_config_drive"] = flattenCloudInitConfigDrive(*in.CloudInitConfigDrive)
// 	}
// 	if in.ServiceAccount != nil {
// 		att["service_account"] = flattenServiceAccount(*in.ServiceAccount)
// 	}
// 	if in.ContainerDisk != nil {
// 		att["container_disk"] = flattenContainerDisk(*in.ContainerDisk)
// 	}
// 	if in.CloudInitNoCloud != nil {
// 		att["cloud_init_no_cloud"] = flattenCloudInitNoCloud(*in.CloudInitNoCloud)
// 	}
// 	if in.Ephemeral != nil {
// 		att["ephemeral"] = flattenEphemeral(*in.Ephemeral)
// 	}
// 	if in.EmptyDisk != nil {
// 		att["empty_disk"] = flattenEmptyDisk(*in.EmptyDisk)
// 	}
// 	if in.HostDisk != nil {
// 		att["host_disk"] = flattenHostDisk(*in.HostDisk)
// 	}
// 	/*if in.PersistentVolumeClaim != nil {
// 		att["persistent_volume_claim"] = flattenPersistentVolumeClaim(in.PersistentVolumeClaim)
// 	}*/
// 	if in.ConfigMap != nil {
// 		att["config_map"] = flattenConfigMap(*in.ConfigMap)
// 	}

// 	return []interface{}{att}
// }

// func flattenDataVolume(in kubevirtapiv1.DataVolumeSource) []interface{} {
// 	att := make(map[string]interface{})

// 	att["name"] = in.Name

// 	return []interface{}{att}
// }

// func flattenCloudInitConfigDrive(in kubevirtapiv1.CloudInitConfigDriveSource) []interface{} {
// 	att := make(map[string]interface{})

// 	if in.UserDataSecretRef != nil {
// 		att["user_data_secret_ref"] = k8s.FlattenLocalObjectReferences(*in.UserDataSecretRef)
// 	}
// 	att["user_data_base64"] = in.UserDataBase64
// 	att["user_data"] = in.UserData
// 	if in.NetworkDataSecretRef != nil {
// 		att["network_data_secret_ref"] = k8s.FlattenLocalObjectReferences(*in.NetworkDataSecretRef)
// 	}
// 	att["network_data_base64"] = in.NetworkDataBase64
// 	att["network_data"] = in.NetworkData

// 	return []interface{}{att}
// }

// func flattenServiceAccount(in kubevirtapiv1.ServiceAccountVolumeSource) []interface{} {
// 	att := make(map[string]interface{})

// 	att["service_account_name"] = in.ServiceAccountName

// 	return []interface{}{att}
// }

// func flattenContainerDisk(in kubevirtapiv1.ContainerDiskSource) []interface{} {
// 	att := make(map[string]interface{})

// 	att["image_url"] = in.Image

// 	return []interface{}{att}
// }

// func flattenCloudInitNoCloud(in kubevirtapiv1.CloudInitNoCloudSource) []interface{} {
// 	att := make(map[string]interface{})

// 	if in.UserDataSecretRef != nil {
// 		att["user_data_secret_ref"] = k8s.FlattenLocalObjectReferences(*in.UserDataSecretRef)
// 	}
// 	att["user_data_base64"] = in.UserDataBase64
// 	att["user_data"] = in.UserData
// 	if in.NetworkDataSecretRef != nil {
// 		att["network_data_secret_ref"] = k8s.FlattenLocalObjectReferences(*in.NetworkDataSecretRef)
// 	}
// 	att["network_data_base64"] = in.NetworkDataBase64
// 	att["network_data"] = in.NetworkData

// 	return []interface{}{att}
// }

// func flattenEphemeral(in kubevirtapiv1.EphemeralVolumeSource) []interface{} {
// 	att := make(map[string]interface{})

// 	if in.PersistentVolumeClaim != nil {
// 		att["persistent_volume_claim"] = flattenPersistentVolumeClaim(*in.PersistentVolumeClaim)
// 	}

// 	return []interface{}{att}
// }

// func flattenPersistentVolumeClaim(in v1.PersistentVolumeClaimVolumeSource) []interface{} {
// 	att := make(map[string]interface{})

// 	att["claim_name"] = in.ClaimName

// 	return []interface{}{att}
// }

// func flattenEmptyDisk(in kubevirtapiv1.EmptyDiskSource) []interface{} {
// 	att := make(map[string]interface{})

// 	att["capacity"] = in.Capacity

// 	return []interface{}{att}
// }

// func flattenHostDisk(in kubevirtapiv1.HostDisk) []interface{} {
// 	att := make(map[string]interface{})

// 	att["path"] = in.Path
// 	att["type"] = in.Type

// 	return []interface{}{att}
// }

// func flattenConfigMap(in kubevirtapiv1.ConfigMapVolumeSource) []interface{} {
// 	att := make(map[string]interface{})

// 	att["name"] = in.Name

// 	return []interface{}{att}
// }
