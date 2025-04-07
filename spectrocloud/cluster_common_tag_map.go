package spectrocloud

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func toTagsMap(d *schema.ResourceData) map[string]string {
	tags := make(map[string]string)
	if d.Get("tags_map") != nil {
		for k, v := range d.Get("tags_map").(map[string]interface{}) {
			vStr := v.(string)
			if v != "" {
				tags[k] = vStr
			} else {
				tags[k] = "spectro__tag"
			}
		}
		return tags
	} else {
		return nil
	}
}

func flattenTagsMap(labels map[string]string) []interface{} {
	tags := make([]interface{}, 0)
	if len(labels) > 0 {
		for k, v := range labels {
			if v == "spectro__tag" {
				tags = append(tags, k)
			} else {
				tags = append(tags, fmt.Sprintf("%s:%s", k, v))
			}
		}
		return tags
	} else {
		return nil
	}
}
