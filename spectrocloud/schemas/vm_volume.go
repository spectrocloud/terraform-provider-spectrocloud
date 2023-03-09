package schemas

import (
	"bytes"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"hash/fnv"
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
								Type:     schema.TypeString,
								Required: true,
							},
						},
					},
				},
				"cloud_init_no_cloud": {
					Type:     schema.TypeSet,
					Optional: true,
					Set:      resourceCloudInitDiskHash,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"user_data": {
								Type:     schema.TypeString,
								Required: true,
							},
						},
					},
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

func hash(s string) uint32 {
	h := fnv.New32a()
	_, _ = h.Write([]byte(s))
	return h.Sum32()
}
