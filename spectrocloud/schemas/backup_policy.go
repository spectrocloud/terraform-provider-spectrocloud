package schemas

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func BackupPolicySchema() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeList,
		Optional:    true,
		MaxItems:    1,
		Description: "The backup policy for the cluster. If not specified, no backups will be taken.",
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"prefix": {
					Type:        schema.TypeString,
					Required:    true,
					Description: "Prefix for the backup name. The backup name will be of the format <prefix>-<cluster-name>-<timestamp>.",
				},
				"backup_location_id": {
					Type:        schema.TypeString,
					Required:    true,
					Description: "The ID of the backup location to use for the backup.",
				},
				"schedule": {
					Type:        schema.TypeString,
					Required:    true,
					Description: "The schedule for the backup. The schedule is specified in cron format. For example, to run the backup every day at 1:00 AM, the schedule should be set to `0 1 * * *`.",
				},
				"expiry_in_hour": {
					Type:        schema.TypeInt,
					Required:    true,
					Description: "The number of hours after which the backup will be deleted. For example, if the expiry is set to 24, the backup will be deleted after 24 hours.",
				},
				"include_disks": {
					Type:        schema.TypeBool,
					Optional:    true,
					Default:     true,
					Description: "Whether to include the disks in the backup. If set to false, only the cluster configuration will be backed up.",
				},
				"include_cluster_resources": {
					Type:        schema.TypeBool,
					Optional:    true,
					Default:     true,
					Description: "Indicates whether to include cluster resources in the backup. If set to false, only the cluster configuration and disks will be backed up. (Note: Starting with Palette version 4.6, the include_cluster_resources attribute will be deprecated, and a new attribute, include_cluster_resources_mode, will be introduced.)",
				},
				"include_cluster_resources_mode": {
					Type:          schema.TypeString,
					Optional:      true,
					ConflictsWith: []string{"include_cluster_resources"},
					Description:   "Specifies whether to include the cluster resources in the backup. Supported values are `always`, `never`, and `auto`.",
					ValidateFunc:  validation.StringInSlice([]string{"always", "never", "auto"}, false),
				},
				"namespaces": {
					Type:     schema.TypeSet,
					Optional: true,
					Set:      schema.HashString,
					Elem: &schema.Schema{
						Type: schema.TypeString,
					},
					Description: "The list of Kubernetes namespaces to include in the backup. If not specified, all namespaces will be included.",
				},
				"cluster_uids": {
					Type:     schema.TypeSet,
					Optional: true,
					Set:      schema.HashString,
					Elem: &schema.Schema{
						Type: schema.TypeString,
					},
					Description: "The list of cluster UIDs to include in the backup. If `include_all_clusters` is set to `true`, then all clusters will be included.",
				},
				"include_all_clusters": {
					Type:        schema.TypeBool,
					Optional:    true,
					Description: "Whether to include all clusters in the backup. If set to false, only the clusters specified in `cluster_uids` will be included.",
				},
			},
		},
	}
}
