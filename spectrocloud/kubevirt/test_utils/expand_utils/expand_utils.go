package expand_utils

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	test_entities "github.com/spectrocloud/terraform-provider-spectrocloud/spectrocloud/kubevirt/test_utils/entities"
	"github.com/spectrocloud/terraform-provider-spectrocloud/spectrocloud/kubevirt/utils"
)

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
				"name":       "test-vm-datavolumedisk1",
				"serial":     "serial",
				"boot_order": 1,
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
