package schemas

import (
	"bytes"
	"fmt"
	"hash/fnv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func VMVolumeSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"name": {
					Type:     schema.TypeString,
					Required: true,
				},
				"container_disk": {
					Type:     schema.TypeSet,
					Optional: true,
					Set:      resourceContainerDiskHash,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"image_url": {
								Type:        schema.TypeString,
								Required:    true,
								ForceNew:    true,
								Description: "The URL of the container image to use as the disk. This can be a local file path, a remote URL, or a registry URL.",
							},
						},
					},
					Description: "A container disk is a disk that is backed by a container image. The container image is expected to contain a disk image in a supported format. The disk image is extracted from the container image and used as the disk for the VM.",
				},
				"cloud_init_no_cloud": {
					Type:     schema.TypeSet,
					Optional: true,
					Set:      resourceCloudInitDiskHash,
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
				"data_volume": {
					Type:     schema.TypeSet,
					Optional: true,
					Set:      resourceDataVolumeHash,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"storage": {
								Type:        schema.TypeString,
								Required:    true,
								Description: "Storage size of the data volume.",
							},
						},
					},
					Description: "The name of the data volume to use as the disk.",
				},
			},
		},
	}
}

func resourceContainerDiskHash(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})

	buf.WriteString(fmt.Sprintf("%s-", m["image_url"].(string)))

	return int(hash(buf.String()))
}

func resourceCloudInitDiskHash(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})

	buf.WriteString(fmt.Sprintf("%s-", m["user_data"].(string)))

	return int(hash(buf.String()))
}

func resourceDataVolumeHash(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})

	buf.WriteString(fmt.Sprintf("%s-", m["storage"].(string)))

	return int(hash(buf.String()))
}

func hash(s string) uint32 {
	h := fnv.New32a()
	_, _ = h.Write([]byte(s))
	return h.Sum32()
}
