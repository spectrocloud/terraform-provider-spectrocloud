package spectrocloud

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/hapi/models"
	"github.com/spectrocloud/terraform-provider-spectrocloud/pkg/client"
	"reflect"
)

type MyMap map[string]string

func dataSourceAppliances() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourcesApplianceRead,

		Schema: map[string]*schema.Schema{
			"ids": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"tags": {
				Type:     schema.TypeSet,
				Description: "A list of tags to filter the appliances."
				Optional: true,
				Set:      schema.HashString,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func dataSourcesApplianceRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.V1Client)
	labels := toTags(d)

	// read all appliances
	appliances, err := c.GetAppliances()
	if err != nil {
		return diag.FromErr(err)
	}

	// prepare filter
	check := func(edgeHostDevice *models.V1EdgeHostDevice) bool {
		return IsMapSubset(edgeHostDevice.Metadata.Labels, labels)
	}

	// apply filter
	output := filter(appliances.Payload.Items, check)

	//read back all ids
	var applianceIDs []string
	for _, appliance := range output {
		applianceIDs = append(applianceIDs, getEdgeHostDeviceUID(appliance))
	}

	id := toDatasourcesId("appliance", labels)

	d.SetId(id) //need to set some id
	err = d.Set("ids", applianceIDs)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func getEdgeHostDeviceUID[T any](appliance T) string { // T should be *models.V1EdgeHostDevice
	return reflect.ValueOf(appliance).Interface().(*models.V1EdgeHostDevice).Metadata.UID
}
