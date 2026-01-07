package schemas

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

// ClusterTypeSchema returns the schema for the cluster_type field.
// This field specifies the type of cluster and can only be set during cluster creation.
// After creation, this field is read-only and any changes will be rejected.
func ClusterTypeSchema() *schema.Schema {
	return &schema.Schema{
		Type:         schema.TypeString,
		Optional:     true,
		Computed:     true,
		ValidateFunc: validation.StringInSlice([]string{"PureManage", "PureAttach"}, false),
		Description: "The cluster type. Valid values are `PureManage` and `PureAttach`. " +
			"This field can only be set during cluster creation and cannot be modified after the cluster is created. " +
			"If not specified, the cluster will use the default type determined by the system.",
	}
}
