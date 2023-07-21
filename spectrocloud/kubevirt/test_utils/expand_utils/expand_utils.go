package expand_utils

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	test_entities "github.com/spectrocloud/terraform-provider-spectrocloud/spectrocloud/kubevirt/test_utils/entities"
	"github.com/spectrocloud/terraform-provider-spectrocloud/spectrocloud/kubevirt/utils"
	corev1 "k8s.io/api/core/v1"
	k8sv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubevirtapiv1 "kubevirt.io/api/core/v1"
	cdiv1 "kubevirt.io/containerized-data-importer-api/pkg/apis/core/v1beta1"
)

func getDataVolumeSpec() cdiv1.DataVolumeSpec {
	imgURL := "docker://gcr.io/spectro-images-public/daily/os/ubuntu-container-disk:22.04"
	limitStorage, _ := resource.ParseQuantity("20Gi")
	requestStorage, _ := resource.ParseQuantity("10Gi")
	storageClassName := "standard"
	volumeMode := "Block"
	return cdiv1.DataVolumeSpec{
		Source: &cdiv1.DataVolumeSource{
			HTTP: &cdiv1.DataVolumeSourceHTTP{
				URL:           "https://cloud.centos.org/centos/7/images/CentOS-7-x86_64-GenericCloud.qcow2",
				SecretRef:     "secret_ref",
				CertConfigMap: "cert_config_map",
			},
			Registry: &cdiv1.DataVolumeSourceRegistry{
				URL: &imgURL,
			},
			PVC: &cdiv1.DataVolumeSourcePVC{
				Namespace: "namespace",
				Name:      "name",
			},
			Blank: nil,
		},
		PVC: &corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{"ReadWriteOnce"},
			Resources: k8sv1.ResourceRequirements{
				Requests: k8sv1.ResourceList{
					"storage": requestStorage,
				},
				Limits: k8sv1.ResourceList{
					"storage": limitStorage,
				},
			},
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"anti-affinity-key": "anti-affinity-val",
				},
			},
			VolumeName:       "volume_name",
			StorageClassName: &storageClassName,
			VolumeMode:       (*k8sv1.PersistentVolumeMode)(&volumeMode),
		},
		ContentType: "content_type",
	}
}

func getBaseOutputForDataVolumeTemplateSpec() kubevirtapiv1.DataVolumeTemplateSpec {
	return kubevirtapiv1.DataVolumeTemplateSpec{
		ObjectMeta: v1.ObjectMeta{
			Name: "test-vm-bootvolume",
		},
		Spec: getDataVolumeSpec(),
	}
}

func GetBaseInputForDataVolume() interface{} {
	return []interface{}{map[string]interface{}{
		"metadata": []interface{}{
			map[string]interface{}{
				"name": "test-vm-bootvolume",
			},
		},
		"spec": []interface{}{
			map[string]interface{}{
				"source": []interface{}{
					map[string]interface{}{
						"blank": []interface{}{
							map[string]interface{}{},
						},
						"http": []interface{}{
							map[string]interface{}{
								"url":             "https://cloud.centos.org/centos/7/images/CentOS-7-x86_64-GenericCloud.qcow2",
								"secret_ref":      "secret_ref",
								"cert_config_map": "cert_config_map",
							},
						},
						"pvc": []interface{}{
							map[string]interface{}{
								"namespace": "namespace",
								"name":      "name",
							},
						},
						"registry": []interface{}{
							map[string]interface{}{
								"image_url": "docker://gcr.io/spectro-images-public/daily/os/ubuntu-container-disk:22.04",
							},
						},
					},
				},
				"pvc": []interface{}{
					map[string]interface{}{
						"access_modes": utils.NewStringSet(schema.HashString, []string{"ReadWriteOnce"}),
						"resources": []interface{}{
							map[string]interface{}{
								"requests": map[string]interface{}{
									"storage": "10Gi",
								},
								"limits": map[string]interface{}{
									"storage": "20Gi",
								},
							},
						},
						"selector":           test_entities.LabelSelectorTerraform,
						"volume_name":        "volume_name",
						"storage_class_name": "standard",
						"volume_mode":        "Block",
					},
				},
				"content_type": "content_type",
			},
		},
	}}
}

func GetBaseInputForVirtualMachine() interface{} {
	return map[string]interface{}{
		"data_volume_templates": GetBaseInputForDataVolume(),
		"run_strategy":          "Always",
		"annotations": map[string]interface{}{
			"annotation_key": "annotation_value",
		},
		"labels": map[string]interface{}{
			"kubevirt.io/vm": "test-vm",
		},
		"generate_name":       "generate_name",
		"name":                "name",
		"namespace":           "namespace",
		"priority_class_name": "priority_class_name",
		"resources": []interface{}{
			map[string]interface{}{
				"requests": map[string]interface{}{
					"cpu":    "4",
					"memory": "10G",
				},
				"limits": map[string]interface{}{
					"cpu":    "8",
					"memory": "20G",
				},
				"over_commit_guest_overhead": false,
			},
		},
		"disk": []interface{}{
			map[string]interface{}{
				"disk_device": []interface{}{
					map[string]interface{}{
						"disk": []interface{}{
							map[string]interface{}{
								"bus":         "virtio",
								"read_only":   true,
								"pci_address": "pci_address",
							},
						},
					},
				},
				"name":   "test-vm-datavolumedisk1",
				"serial": "serial",
			},
		},
		"interface": []interface{}{
			map[string]interface{}{
				"interface_binding_method": "InterfaceBridge",
				"name":                     "main",
			},
		},
		"node_selector": map[string]interface{}{
			"node_selector_key": "node_selector_value",
		},
		"affinity": []interface{}{
			map[string]interface{}{
				"node_affinity": []interface{}{
					map[string]interface{}{
						"required_during_scheduling_ignored_during_execution":  test_entities.NodeRequiredDuringSchedulingIgnoredDuringExecution,
						"preferred_during_scheduling_ignored_during_execution": test_entities.NodePreferredDuringSchedulingIgnoredDuringExecution,
					},
				},
				"pod_affinity": []interface{}{
					map[string]interface{}{
						"preferred_during_scheduling_ignored_during_execution": test_entities.PodPreferredDuringSchedulingIgnoredDuringExecutionTerraform,
						"required_during_scheduling_ignored_during_execution":  test_entities.PodRequiredDuringSchedulingIgnoredDuringExecutionTerraform,
					},
				},
				"pod_anti_affinity": []interface{}{
					map[string]interface{}{
						"preferred_during_scheduling_ignored_during_execution": test_entities.PodPreferredDuringSchedulingIgnoredDuringExecutionTerraform,
						"required_during_scheduling_ignored_during_execution":  test_entities.PodRequiredDuringSchedulingIgnoredDuringExecutionTerraform,
					},
				},
			},
		},
		"scheduler_name": "scheduler_name",
		"tolerations": []interface{}{
			map[string]interface{}{
				"effect":             "effect",
				"key":                "key",
				"operator":           "operator",
				"toleration_seconds": "60",
				"value":              "value",
			},
		},
		"eviction_strategy":                "eviction_strategy",
		"termination_grace_period_seconds": 120,
		"volume": []interface{}{
			map[string]interface{}{
				"name": "test-vm-datavolumedisk1",
				"volume_source": []interface{}{
					map[string]interface{}{
						"data_volume": []interface{}{
							map[string]interface{}{
								"name": "test-vm-bootvolume",
							},
						},
						"cloud_init_config_drive": []interface{}{
							map[string]interface{}{
								"user_data_secret_ref": []interface{}{
									map[string]interface{}{
										"name": "name",
									},
								},
								"user_data_base64": "user_data_base64",
								"user_data":        "user_data",
								"network_data_secret_ref": []interface{}{
									map[string]interface{}{
										"name": "name",
									},
								},
								"network_data_base64": "network_data_base64",
								"network_data":        "network_data",
							},
						},
						"service_account": []interface{}{
							map[string]interface{}{
								"service_account_name": "service_account_name",
							},
						},
					},
				},
			},
		},
		"hostname":  "hostname",
		"subdomain": "subdomain",
		"network": []interface{}{
			map[string]interface{}{
				"name": "main",
				"network_source": []interface{}{
					map[string]interface{}{
						"pod": []interface{}{
							map[string]interface{}{
								"vm_network_cidr": "vm_network_cidr",
							},
						},
						"multus": []interface{}{
							map[string]interface{}{
								"network_name": "tenantcluster",
							},
						},
					},
				},
			},
		},
		"dns_policy": "dns_policy",
	}
}

func GetBaseOutputForVirtualMachine() kubevirtapiv1.VirtualMachineSpec {
	return kubevirtapiv1.VirtualMachineSpec{
		RunStrategy: (func() *kubevirtapiv1.VirtualMachineRunStrategy {
			strategy := kubevirtapiv1.VirtualMachineRunStrategy("Always")
			return &strategy
		})(),
		DataVolumeTemplates: []kubevirtapiv1.DataVolumeTemplateSpec{
			getBaseOutputForDataVolumeTemplateSpec(),
		},
		Template: &kubevirtapiv1.VirtualMachineInstanceTemplateSpec{
			ObjectMeta: v1.ObjectMeta{},
			Spec: kubevirtapiv1.VirtualMachineInstanceSpec{
				PriorityClassName: "priority_class_name",
				Domain: kubevirtapiv1.DomainSpec{
					Resources: kubevirtapiv1.ResourceRequirements{
						Requests: k8sv1.ResourceList{
							"memory": (func() resource.Quantity { res, _ := resource.ParseQuantity("10G"); return res })(),
							"cpu":    *resource.NewQuantity(int64(4), resource.DecimalExponent),
						},
						Limits: k8sv1.ResourceList{
							"memory": (func() resource.Quantity { res, _ := resource.ParseQuantity("20G"); return res })(),
							"cpu":    *resource.NewQuantity(int64(8), resource.DecimalExponent),
						},
						OvercommitGuestOverhead: false,
					},
					Devices: kubevirtapiv1.Devices{
						Disks: []kubevirtapiv1.Disk{
							{
								Name:   "test-vm-datavolumedisk1",
								Serial: "serial",
								DiskDevice: kubevirtapiv1.DiskDevice{
									Disk: &kubevirtapiv1.DiskTarget{
										Bus:        "virtio",
										ReadOnly:   true,
										PciAddress: "pci_address",
									},
								},
							},
						},
						Interfaces: []kubevirtapiv1.Interface{
							{
								Name: "main",
								InterfaceBindingMethod: kubevirtapiv1.InterfaceBindingMethod{
									Bridge: &kubevirtapiv1.InterfaceBridge{},
								},
							},
						},
					},
				},
				NodeSelector: map[string]string{
					"node_selector_key": "node_selector_value",
				},
				Affinity:      nil,
				SchedulerName: "scheduler_name",
				Tolerations: []k8sv1.Toleration{
					{
						Effect:            k8sv1.TaintEffect("effect"),
						Key:               "key",
						Operator:          k8sv1.TolerationOperator("operator"),
						TolerationSeconds: utils.PtrToInt64(int64(60)),
						Value:             "value",
					},
				},
				EvictionStrategy: (func() *kubevirtapiv1.EvictionStrategy {
					retval := kubevirtapiv1.EvictionStrategy("eviction_strategy")
					return &retval
				})(),
				TerminationGracePeriodSeconds: utils.PtrToInt64(int64(120)),
				Volumes: []kubevirtapiv1.Volume{
					{
						Name: "test-vm-datavolumedisk1",
						VolumeSource: kubevirtapiv1.VolumeSource{
							DataVolume: &kubevirtapiv1.DataVolumeSource{
								Name: "test-vm-bootvolume",
							},
							CloudInitConfigDrive: &kubevirtapiv1.CloudInitConfigDriveSource{
								UserDataSecretRef: &k8sv1.LocalObjectReference{
									Name: "name",
								},
								UserDataBase64: "user_data_base64",
								UserData:       "user_data",
								NetworkDataSecretRef: &k8sv1.LocalObjectReference{
									Name: "name",
								},
								NetworkDataBase64: "network_data_base64",
								NetworkData:       "network_data",
							},
							ServiceAccount: &kubevirtapiv1.ServiceAccountVolumeSource{
								ServiceAccountName: "service_account_name",
							},
						},
					},
				},
				Hostname:  "hostname",
				Subdomain: "subdomain",
				Networks: []kubevirtapiv1.Network{
					{
						Name: "main",
						NetworkSource: kubevirtapiv1.NetworkSource{
							Pod: &kubevirtapiv1.PodNetwork{
								VMNetworkCIDR: "vm_network_cidr",
							},
							Multus: &kubevirtapiv1.MultusNetwork{
								NetworkName: "tenantcluster",
							},
						},
					},
				},
				DNSPolicy: k8sv1.DNSPolicy("dns_policy"),
				DNSConfig: nil,
			},
		},
	}
}
