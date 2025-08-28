package spectrocloud

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceAlertImport(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	c := getV1ClientWithResourceContext(m, "tenant")
	idParts := strings.Split(d.Id(), ":")
	if len(idParts) != 2 {
		return nil, fmt.Errorf("invalid import ID format, expected 'alertUid:component', got: %s", d.Id())
	}
	alertUid := idParts[0]
	component := idParts[1]
	pjt, err := c.GetProject(ProviderInitProjectUid)
	if err != nil {
		return nil, err
	}
	if err := d.Set("project", pjt.Metadata.Name); err != nil {
		return nil, err
	}
	if err := d.Set("component", component); err != nil {
		return nil, err
	}
	d.SetId(alertUid)
	// Read all alert data to populate the state
	diags := resourceAlertRead(ctx, d, m)
	if diags.HasError() {
		return nil, fmt.Errorf("could not read alert for import: %v", diags)
	}

	return []*schema.ResourceData{d}, nil
}
