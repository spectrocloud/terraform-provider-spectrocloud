package spectrocloud

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-api-go/models"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestToTeam(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]interface{}
		expected *models.V1Team
	}{
		{
			name: "Valid Data",
			input: map[string]interface{}{
				"name":  "team-1",
				"uid":   "",
				"users": []interface{}{"user1", "user2"},
			},
			expected: &models.V1Team{
				Metadata: &models.V1ObjectMeta{
					Name: "team-1",
					UID:  "",
				},
				Spec: &models.V1TeamSpec{
					Users: []string{"user1", "user2"},
				},
			},
		},
		{
			name: "Missing Users",
			input: map[string]interface{}{
				"name": "team-2",
			},
			expected: &models.V1Team{
				Metadata: &models.V1ObjectMeta{
					Name: "team-2",
					UID:  "",
				},
				Spec: &models.V1TeamSpec{
					Users: []string{},
				},
			},
		},
		{
			name: "Empty Name",
			input: map[string]interface{}{
				"name":  "",
				"users": []interface{}{"user3"},
			},
			expected: &models.V1Team{
				Metadata: &models.V1ObjectMeta{
					Name: "",
					UID:  "",
				},
				Spec: &models.V1TeamSpec{
					Users: []string{"user3"},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := schema.TestResourceDataRaw(t, resourceTeam().Schema, tt.input)

			result := toTeam(d)
			assert.Equal(t, tt.expected, result) // Compare the expected and actual result
			assert.Equal(t, tt.expected.Metadata.Name, result.Metadata.Name)
			assert.Equal(t, tt.expected.Metadata.UID, result.Metadata.UID)
			assert.ElementsMatch(t, tt.expected.Spec.Users, result.Spec.Users) // Compare slices ignoring order
		})
	}
}

func TestToTeamProjectRoleMapping(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]interface{}
		expected *models.V1ProjectRolesPatch
	}{
		{
			name: "Valid Data",
			input: map[string]interface{}{
				"project_role_mapping": []interface{}{
					map[string]interface{}{
						"id":    "project1",
						"roles": []interface{}{"admin", "viewer"},
					},
					map[string]interface{}{
						"id":    "project2",
						"roles": []interface{}{"editor"},
					},
				},
			},
			expected: &models.V1ProjectRolesPatch{
				Projects: []*models.V1ProjectRolesPatchProjectsItems0{
					{
						ProjectUID: "project2",
						Roles:      []string{"editor"},
					},
					{
						ProjectUID: "project1",
						Roles:      []string{"admin", "viewer"},
					},
				},
			},
		},
		{
			name: "No Project Role Mappings",
			input: map[string]interface{}{
				"project_role_mapping": []interface{}{},
			},
			expected: &models.V1ProjectRolesPatch{
				Projects: []*models.V1ProjectRolesPatchProjectsItems0{},
			},
		},
		{
			name: "Empty Roles",
			input: map[string]interface{}{
				"project_role_mapping": []interface{}{
					map[string]interface{}{
						"id":    "project3",
						"roles": []interface{}{},
					},
				},
			},
			expected: &models.V1ProjectRolesPatch{
				Projects: []*models.V1ProjectRolesPatchProjectsItems0{
					{
						ProjectUID: "project3",
						Roles:      []string{},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Ensure the schema matches the expected format
			resourceSchema := resourceTeam().Schema
			d := schema.TestResourceDataRaw(t, resourceSchema, tt.input)

			// Run the function to test
			result := toTeamProjectRoleMapping(d)

			// Perform assertions
			assert.Equal(t, tt.expected, result)
			for i, project := range tt.expected.Projects {
				assert.Equal(t, project.ProjectUID, result.Projects[i].ProjectUID)
				assert.ElementsMatch(t, project.Roles, result.Projects[i].Roles)
			}
		})
	}
}

func TestToTeamTenantRoleMapping(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]interface{}
		expected *models.V1TeamTenantRolesUpdate
	}{
		{
			name: "Valid Data",
			input: map[string]interface{}{
				"tenant_role_mapping": []interface{}{"role1", "role2"},
			},
			expected: &models.V1TeamTenantRolesUpdate{
				Roles: []string{"role2", "role1"},
			},
		},
		{
			name: "No Tenant Role Mappings",
			input: map[string]interface{}{
				"tenant_role_mapping": []interface{}{},
			},
			expected: &models.V1TeamTenantRolesUpdate{
				Roles: []string{},
			},
		},
		{
			name: "Empty Roles",
			input: map[string]interface{}{
				"tenant_role_mapping": []interface{}{""},
			},
			expected: &models.V1TeamTenantRolesUpdate{
				Roles: []string{""},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Ensure the schema matches the expected format
			resourceSchema := resourceTeam().Schema
			d := schema.TestResourceDataRaw(t, resourceSchema, tt.input)

			// Run the function to test
			result := toTeamTenantRoleMapping(d)

			// Perform assertions
			assert.Equal(t, tt.expected, result)
			assert.ElementsMatch(t, tt.expected.Roles, result.Roles)
		})
	}
}

func TestToTeamWorkspaceRoleMapping(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]interface{}
		expected *models.V1WorkspacesRolesPatch
	}{
		{
			name: "Valid Data",
			input: map[string]interface{}{
				"workspace_role_mapping": []interface{}{
					map[string]interface{}{
						"workspace": []interface{}{
							map[string]interface{}{
								"id":    "workspace1",
								"roles": []interface{}{"role1", "role2"},
							},
						},
					},
				},
			},
			expected: &models.V1WorkspacesRolesPatch{
				Workspaces: []*models.V1WorkspaceRolesPatch{
					{
						UID: "workspace1",
						Roles: []string{
							"role2",
							"role1",
						},
					},
				},
			},
		},
		{
			name: "No Workspace Role Mappings",
			input: map[string]interface{}{
				"workspace_role_mapping": []interface{}{},
			},
			expected: &models.V1WorkspacesRolesPatch{
				Workspaces: []*models.V1WorkspaceRolesPatch{},
			},
		},
		{
			name: "Empty Workspace Role Mapping",
			input: map[string]interface{}{
				"workspace_role_mapping": []interface{}{
					map[string]interface{}{
						"workspace": []interface{}{
							map[string]interface{}{
								"id":    "workspace1",
								"roles": []interface{}{},
							},
						},
					},
				},
			},
			expected: &models.V1WorkspacesRolesPatch{
				Workspaces: []*models.V1WorkspaceRolesPatch{
					{
						UID:   "workspace1",
						Roles: []string{},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a schema.ResourceData instance
			d := schema.TestResourceDataRaw(t, resourceTeam().Schema, tt.input)

			// Call the function under test
			result := toTeamWorkspaceRoleMapping(d)

			// Perform assertions
			assert.Equal(t, tt.expected, result)
		})
	}
}
