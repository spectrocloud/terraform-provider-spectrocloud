package flatten_utils

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	test_entities "github.com/spectrocloud/terraform-provider-spectrocloud/spectrocloud/kubevirt/test_utils/entities"
	"github.com/spectrocloud/terraform-provider-spectrocloud/types"
)

func GetBaseOutputForDataVolume() interface{} {
	return map[string]interface{}{
		"metadata": []interface{}{
			map[string]interface{}{
				"annotations":      interface{}(map[string]interface{}(nil)),
				"labels":           interface{}(map[string]interface{}(nil)),
				"name":             "test-vm-bootvolume",
				"resource_version": interface{}(""),
				"self_link":        interface{}(""),
				"uid":              interface{}(""),
				"generation":       interface{}(int64(0)),
				"namespace":        "tenantcluster",
				"generate_name":    "generate_name",
			},
		},
		"spec": []interface{}{
			map[string]interface{}{
				"pvc": []interface{}{
					map[string]interface{}{
						"access_modes": (func() *schema.Set {
							out := []interface{}{
								"ReadWriteOnce",
							}
							return schema.NewSet(schema.HashString, out)
						})(),
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
					},
				},
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
								"image_url": types.Ptr("docker://gcr.io/spectro-images-public/daily/os/ubuntu-container-disk:22.04"),
							},
						},
					},
				},
				"content_type": "content_type",
			},
		},
		"status": []interface{}{
			map[string]interface{}{
				"phase":    "",
				"progress": "",
			},
		},
	}
}
