package spectrocloud

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/spectrocloud/palette-sdk-go/client"
)

func resourceApplicationImport(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	_, err := GetCommonApplication(d, m)
	if err != nil {
		return nil, err
	}

	diags := resourceApplicationRead(ctx, d, m)
	if diags.HasError() {
		return nil, fmt.Errorf("could not read application for import: %v", diags)
	}

	return []*schema.ResourceData{d}, nil
}

func GetCommonApplication(d *schema.ResourceData, m interface{}) (*client.V1Client, error) {
	idOrName := d.Id()
	if idOrName == "" {
		return nil, fmt.Errorf("application ID or name is required for import")
	}

	// Try project context first, then tenant — works for both UID and name lookups
	for _, resourceContext := range []string{"project", "tenant"} {
		c := getV1ClientWithResourceContext(m, resourceContext)

		// 1) Try as UID
		app, err := c.GetApplication(idOrName)
		if err == nil && app != nil {
			return setApplicationImportState(d, app, resourceContext, c)
		}

		// 2) Try as name via search
		app, err = getApplicationByName(c, idOrName)
		if err != nil {
			continue
		}
		if app != nil {
			return setApplicationImportState(d, app, resourceContext, c)
		}
	}

	return nil, fmt.Errorf("application %q not found in project or tenant context", idOrName)
}

func getApplicationByName(c *client.V1Client, name string) (*models.V1AppDeployment, error) {
	ignoreCase := false
	summaries, err := c.SearchAppDeploymentSummaries(
		&models.V1AppDeploymentFilterSpec{
			AppDeploymentName: &models.V1FilterString{
				Eq:         &name,
				IgnoreCase: &ignoreCase,
			},
		},
		nil,
	)
	if err != nil {
		return nil, err
	}

	var uid string
	for _, s := range summaries {
		if s != nil && s.Metadata != nil && s.Metadata.Name == name {
			if uid != "" {
				return nil, fmt.Errorf("multiple applications found with name %q; use UID to import instead", name)
			}
			uid = s.Metadata.UID
		}
	}

	if uid == "" {
		return nil, nil
	}
	return c.GetApplication(uid)
}

func setApplicationImportState(d *schema.ResourceData, app *models.V1AppDeployment, resourceContext string, c *client.V1Client) (*client.V1Client, error) {
	if err := d.Set("name", app.Metadata.Name); err != nil {
		return nil, err
	}
	if app.Spec != nil && app.Spec.Profile != nil && app.Spec.Profile.Metadata != nil {
		if err := d.Set("application_profile_uid", app.Spec.Profile.Metadata.UID); err != nil {
			return nil, err
		}
	}
	if app.Spec != nil && app.Spec.Config != nil && app.Spec.Config.Target != nil {
		config := map[string]interface{}{"cluster_context": resourceContext}
		if err := d.Set("config", []interface{}{config}); err != nil {
			return nil, err
		}
	}
	// Set placeholder config with required cluster_context
	// The resource context will be determined and set properly in the read function

	d.SetId(app.Metadata.UID)
	return c, nil
}
