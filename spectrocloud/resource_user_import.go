package spectrocloud

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceUserImport(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	c := getV1ClientWithResourceContext(m, "tenant")
	user, err := c.GetUserByID(d.Id())
	if err != nil {
		return nil, err
	}
	err = d.Set("email", user.Spec.EmailID)
	if err != nil {
		return nil, err
	}
	diags := resourceUserRead(ctx, d, m)
	if diags.HasError() {
		return nil, fmt.Errorf("could not read cluster for import: %v", diags)
	}
	return []*schema.ResourceData{d}, nil
}
