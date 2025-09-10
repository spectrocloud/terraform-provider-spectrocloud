package spectrocloud

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceAlertImport(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {

	idParts := d.Id()
	context := "project"
	c := getV1ClientWithResourceContext(m, context)
	pjt, err := c.GetProject(ProviderInitProjectUid)
	if err != nil {
		return nil, err
	}
	if err := d.Set("project", pjt.Metadata.Name); err != nil {
		return nil, err
	}
	if err := d.Set("component", "ClusterHealth"); err != nil {
		return nil, err
	}
	d.SetId(idParts)
	// Read all alert data to populate the state
	diags := resourceAlertRead(ctx, d, m)
	if diags.HasError() {
		return nil, fmt.Errorf("could not read alert for import: %v", diags)
	}

	return []*schema.ResourceData{d}, nil
}
