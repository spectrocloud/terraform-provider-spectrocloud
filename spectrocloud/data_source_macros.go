package spectrocloud

// import (
// 	"context"

// 	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
// 	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

// 	"github.com/spectrocloud/palette-sdk-go/client"
// )

// func dataSourceMacros() *schema.Resource {
// 	return &schema.Resource{
// 		ReadContext: dataSourceProjectRead,

// 		Schema: map[string]*schema.Schema{
// 			"project": {
// 				Type:     schema.TypeString,
// 				Computed: true,
// 				Optional: true,
// 			},
// 		},
// 	}
// }

// func dataSourceMacrosRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
// 	c := m.(*client.V1Client)
// 	var diags diag.Diagnostics
// 	if v, ok := d.GetOk("project"); ok && v.(string) != "" {
// 		uid, err := c.GetProjectUID(v.(string))
// 		if err != nil {
// 			return diag.FromErr(err)
// 		}
// 		d.SetId(uid)
// 		d.Set("name", v.(string))
// 	}
// 	return diags
// }
