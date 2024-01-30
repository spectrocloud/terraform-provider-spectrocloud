package schemas

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func ProfileVariables() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		MaxItems: 1,
		Optional: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"variable": {
					Type:     schema.TypeList,
					Required: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"name": {
								Type:        schema.TypeString,
								Required:    true,
								Description: "The name of the variable should be unique among variables.",
							},
							"display_name": {
								Type:        schema.TypeString,
								Required:    true,
								Description: "The display name of the variable should be unique among variables.",
							},
							"format": {
								Type:         schema.TypeString,
								Optional:     true,
								Default:      "string",
								ValidateFunc: validation.StringInSlice([]string{"string", "number", "boolean", "password", "ipv4", "version"}, false),
								Description:  "The format of the variable. Default is `string`, `format` field can only be set during cluster profile creation. Subsequent day 2 operations on this field are blocked.",
							},
							"description": {
								Type:        schema.TypeString,
								Optional:    true,
								Description: "The description of the variable.",
							},
							"default_value": {
								Type:        schema.TypeString,
								Optional:    true,
								Description: "The default value of the variable.",
							},
							"regex": {
								Type:        schema.TypeString,
								Optional:    true,
								Description: "Regular expression pattern which the variable value must match. `regex` field can only be set during cluster profile creation. Subsequent day 2 operations on this field are blocked.",
							},
							"required": {
								Type:        schema.TypeBool,
								Optional:    true,
								Description: "Indicates whether the variable is required during cluster provisioning. `required` field can only be set during cluster profile creation. Subsequent day 2 operations on this field are blocked.",
							},
							"immutable": {
								Type:        schema.TypeBool,
								Optional:    true,
								Description: "Indicates if the variable is immutable.",
							},
							"hidden": {
								Type:        schema.TypeBool,
								Optional:    true,
								Description: "Indicates if the variable is hidden.",
							},
						},
					},
				},
			},
		},
		Description: "List of variables for the cluster profile.",
	}
}
