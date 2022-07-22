package spectrocloud

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"strings"
)

func toTags(d *schema.ResourceData) map[string]string {
	tags := make(map[string]string)
	if d.Get("tags") != nil {
		for _, t := range d.Get("tags").(*schema.Set).List() {
			tag := t.(string)
			if strings.Contains(tag, ":") {
				tags[strings.Split(tag, ":")[0]] = strings.Split(tag, ":")[1]
			} else {
				tags[tag] = "spectro__tag"
			}
		}
		return tags
	} else {
		return nil
	}
}

func flattenTags(labels map[string]string) []interface{} {
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
