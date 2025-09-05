package spectrocloud

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceTeamImport(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	c := getV1ClientWithResourceContext(m, "tenant")

	// The import ID should be the team UID
	teamUID := d.Id()

	// Validate that the team exists and we can access it
	team, err := c.GetTeam(teamUID)
	if err != nil {
		return nil, fmt.Errorf("could not retrieve team for import: %s", err)
	}
	if team == nil {
		return nil, fmt.Errorf("team with ID %s not found", teamUID)
	}

	// Set the team name from the retrieved team
	if err := d.Set("name", team.Metadata.Name); err != nil {
		return nil, err
	}

	// Read all team data to populate the state
	diags := resourceTeamRead(ctx, d, m)
	if diags.HasError() {
		return nil, fmt.Errorf("could not read team for import: %v", diags)
	}

	return []*schema.ResourceData{d}, nil
}
