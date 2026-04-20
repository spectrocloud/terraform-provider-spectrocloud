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
								ValidateFunc: validation.StringInSlice([]string{"string", "number", "boolean", "ipv4", "ipv4cidr", "ipv6", "version", "base64"}, false),
								Description:  "The format of the variable. Default is `string`, `format` field can only be set during cluster profile creation. Allowed formats include `string`, `number`, `boolean`, `ipv4`, `ipv4cidr`, `ipv6`, `version`, `base64`.",
							},
							"description": {
								Type:        schema.TypeString,
								Optional:    true,
								Description: "The description of the variable.",
							},
							"default_value": {
								Type:        schema.TypeString,
								Optional:    true,
								Description: "The default value of the variable. If the format is `multiline`, then the default value should be a multi-line string. If the input type is `dropdown`, then the default value should be a option label.",
							},
							"regex": {
								Type:        schema.TypeString,
								Optional:    true,
								Description: "Regular expression pattern which the variable value must match.",
							},
							"required": {
								Type:        schema.TypeBool,
								Optional:    true,
								Description: "The `required` to specify if the variable is optional or mandatory. If it is mandatory then default value must be provided.",
							},
							"immutable": {
								Type:        schema.TypeBool,
								Optional:    true,
								Description: "If `immutable` is set to `true`, then variable value can't be editable. By default the `immutable` flag will be set to `false`.",
							},
							"is_sensitive": {
								Type:        schema.TypeBool,
								Optional:    true,
								Description: "If `is_sensitive` is set to `true`, then default value will be masked. By default the `is_sensitive` flag will be set to false.",
							},
							"hidden": {
								Type:        schema.TypeBool,
								Optional:    true,
								Description: "If `hidden` is set to `true`, then variable will be hidden for overriding the value. By default the `hidden` flag will be set to `false`.",
							},
							"input_type": {
								Type:         schema.TypeString,
								Optional:     true,
								Default:      "text",
								ValidateFunc: validation.StringInSlice([]string{"text", "dropdown", "multiline"}, false),
								Description:  "The input type of the variable. Defaults to `text` for backward compatibility. Allowed input types include `text`, `dropdown`, `multiline`.",
							},
							"options": {
								Type:        schema.TypeList,
								Optional:    true,
								Description: "The options of the variable. Only applicable for dropdown input type.",
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"default": {
											Type:        schema.TypeBool,
											Computed:    true,
											Description: "The default value of the option.",
										},
										"description": {
											Type:        schema.TypeString,
											Optional:    true,
											Description: "The description of the option.",
										},
										"label": {
											Type:        schema.TypeString,
											Required:    true,
											Description: "The label of the option.",
										},
										"value": {
											Type:        schema.TypeString,
											Required:    true,
											Description: "The value of the option.",
										},
									},
								},
							},
						},
					},
				},
			},
		},
		Description: "List of variables for the cluster profile.",
	}
}
