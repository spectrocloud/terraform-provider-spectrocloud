package spectrocloud

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/stretchr/testify/assert"
)

func TestConvertSummaryToIDS(t *testing.T) {
	tests := []struct {
		name     string
		input    []*models.V1UIDSummary
		expected []string
	}{
		{
			name: "Multiple UIDs",
			input: []*models.V1UIDSummary{
				{UID: "uid1"},
				{UID: "uid2"},
				{UID: "uid3"},
			},
			expected: []string{"uid1", "uid2", "uid3"},
		},
		{
			name:     "Empty input",
			input:    []*models.V1UIDSummary{},
			expected: []string(nil),
		},
		{
			name: "Single UID",
			input: []*models.V1UIDSummary{
				{UID: "singleUID"},
			},
			expected: []string{"singleUID"},
		},
		{
			name:     "Nil input",
			input:    nil,
			expected: []string(nil),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := convertSummaryToIDS(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestConvertToStrings(t *testing.T) {
	tests := []struct {
		name     string
		input    []interface{}
		expected []string
	}{
		{
			name:     "All strings",
			input:    []interface{}{"one", "two", "three"},
			expected: []string{"one", "two", "three"},
		},
		{
			name:     "Mixed types",
			input:    []interface{}{"one", 2, "three", 4.0, true},
			expected: []string{"one", "three"},
		},
		{
			name:     "No strings",
			input:    []interface{}{1, 2, 3, 4.5, false},
			expected: []string(nil),
		},
		{
			name:     "Empty input",
			input:    []interface{}{},
			expected: []string(nil),
		},
		{
			name:     "Nil input",
			input:    nil,
			expected: []string(nil),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := convertToStrings(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestToUser(t *testing.T) {
	resourceData := schema.TestResourceDataRaw(t, map[string]*schema.Schema{
		"first_name": {Type: schema.TypeString, Required: true},
		"last_name":  {Type: schema.TypeString, Required: true},
		"email":      {Type: schema.TypeString, Required: true},
		"team_ids":   {Type: schema.TypeList, Optional: true, Elem: &schema.Schema{Type: schema.TypeString}},
	}, map[string]interface{}{
		"first_name": "John",
		"last_name":  "Doe",
		"email":      "johndoe@example.com",
		"team_ids":   []interface{}{"team1", "team2"},
	})

	user := toUser(resourceData)

	expectedUser := &models.V1UserEntity{
		Metadata: &models.V1ObjectMeta{
			Name: "John Doe",
		},
		Spec: &models.V1UserSpecEntity{
			EmailID:   "johndoe@example.com",
			FirstName: "John",
			LastName:  "Doe",
			Teams:     []string{"team1", "team2"},
		},
	}

	assert.Equal(t, expectedUser, user)
}

func TestToUserNoTeams(t *testing.T) {
	resourceData := schema.TestResourceDataRaw(t, map[string]*schema.Schema{
		"first_name": {Type: schema.TypeString, Required: true},
		"last_name":  {Type: schema.TypeString, Required: true},
		"email":      {Type: schema.TypeString, Required: true},
		"team_ids":   {Type: schema.TypeList, Optional: true, Elem: &schema.Schema{Type: schema.TypeString}},
	}, map[string]interface{}{
		"first_name": "Alice",
		"last_name":  "Smith",
		"email":      "alice@example.com",
	})

	user := toUser(resourceData)

	expectedUser := &models.V1UserEntity{
		Metadata: &models.V1ObjectMeta{
			Name: "Alice Smith",
		},
		Spec: &models.V1UserSpecEntity{
			EmailID:   "alice@example.com",
			FirstName: "Alice",
			LastName:  "Smith",
			Teams:     nil,
		},
	}

	assert.Equal(t, expectedUser, user)
}

func TestSetToStringArray(t *testing.T) {
	// Create a schema.Set with some string values
	input := schema.NewSet(schema.HashString, []interface{}{"id1", "id2", "id3"})

	// Call the function with the set
	result := setToStringArray(input)

	// Define the expected output
	expected := []string{"id1", "id2", "id3"}

	// Assert that the result matches the expected output
	assert.ElementsMatch(t, expected, result)
}

func TestSetToStringArrayEmptySet(t *testing.T) {
	// Create an empty schema.Set
	input := schema.NewSet(schema.HashString, []interface{}{})

	// Call the function with the empty set
	result := setToStringArray(input)

	// Define the expected output for an empty set
	expected := []string{}

	// Assert that the result matches the expected output
	assert.Equal(t, expected, result)
}

func TestToUserWorkspaceRoleMappingEmpty(t *testing.T) {
	d := schema.TestResourceDataRaw(t, map[string]*schema.Schema{
		"workspace_role": {
			Type: schema.TypeSet,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"workspace": {
						Type: schema.TypeSet,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"id":       {Type: schema.TypeString},
								"role_ids": {Type: schema.TypeSet, Elem: &schema.Schema{Type: schema.TypeString}},
							},
						},
					},
				},
			},
		},
	}, map[string]interface{}{"workspace_role": []interface{}{}})

	result := toUserWorkspaceRoleMapping(d)
	expected := &models.V1WorkspacesRolesPatch{Workspaces: []*models.V1WorkspaceRolesPatch{}}

	assert.Equal(t, expected, result)
}

func TestToUserResourceRoleMapping(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() *schema.ResourceData
		expected []*models.V1ResourceRolesUpdateEntity
	}{
		{
			name: "Single resource role with all fields",
			setup: func() *schema.ResourceData {
				resourceData := schema.TestResourceDataRaw(t, map[string]*schema.Schema{
					"resource_role": {
						Type: schema.TypeSet,
						Set:  resourceUserResourceRoleMappingHash,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"project_ids": {
									Type:     schema.TypeSet,
									Set:      schema.HashString,
									Required: true,
									Elem:     &schema.Schema{Type: schema.TypeString},
								},
								"filter_ids": {
									Type:     schema.TypeSet,
									Set:      schema.HashString,
									Required: true,
									Elem:     &schema.Schema{Type: schema.TypeString},
								},
								"role_ids": {
									Type:     schema.TypeSet,
									Set:      schema.HashString,
									Required: true,
									Elem:     &schema.Schema{Type: schema.TypeString},
								},
							},
						},
					},
				}, map[string]interface{}{
					"resource_role": []interface{}{
						map[string]interface{}{
							"project_ids": []interface{}{"project1", "project2"},
							"filter_ids":  []interface{}{"filter1"},
							"role_ids":    []interface{}{"role1", "role2", "role3"},
						},
					},
				})
				return resourceData
			},
			expected: []*models.V1ResourceRolesUpdateEntity{
				{
					ProjectUids: []string{"project1", "project2"},
					FilterRefs:  []string{"filter1"},
					Roles:       []string{"role1", "role2", "role3"},
				},
			},
		},
		{
			name: "Multiple resource roles",
			setup: func() *schema.ResourceData {
				resourceData := schema.TestResourceDataRaw(t, map[string]*schema.Schema{
					"resource_role": {
						Type: schema.TypeSet,
						Set:  resourceUserResourceRoleMappingHash,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"project_ids": {
									Type:     schema.TypeSet,
									Set:      schema.HashString,
									Required: true,
									Elem:     &schema.Schema{Type: schema.TypeString},
								},
								"filter_ids": {
									Type:     schema.TypeSet,
									Set:      schema.HashString,
									Required: true,
									Elem:     &schema.Schema{Type: schema.TypeString},
								},
								"role_ids": {
									Type:     schema.TypeSet,
									Set:      schema.HashString,
									Required: true,
									Elem:     &schema.Schema{Type: schema.TypeString},
								},
							},
						},
					},
				}, map[string]interface{}{
					"resource_role": []interface{}{
						map[string]interface{}{
							"project_ids": []interface{}{"project1"},
							"filter_ids":  []interface{}{"filter1"},
							"role_ids":    []interface{}{"role1"},
						},
						map[string]interface{}{
							"project_ids": []interface{}{"project2", "project3"},
							"filter_ids":  []interface{}{"filter2", "filter3"},
							"role_ids":    []interface{}{"role2"},
						},
					},
				})
				return resourceData
			},
			expected: []*models.V1ResourceRolesUpdateEntity{
				{
					ProjectUids: []string{"project1"},
					FilterRefs:  []string{"filter1"},
					Roles:       []string{"role1"},
				},
				{
					ProjectUids: []string{"project2", "project3"},
					FilterRefs:  []string{"filter2", "filter3"},
					Roles:       []string{"role2"},
				},
			},
		},
		{
			name: "Empty sets within resource role",
			setup: func() *schema.ResourceData {
				resourceData := schema.TestResourceDataRaw(t, map[string]*schema.Schema{
					"resource_role": {
						Type: schema.TypeSet,
						Set:  resourceUserResourceRoleMappingHash,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"project_ids": {
									Type:     schema.TypeSet,
									Set:      schema.HashString,
									Required: true,
									Elem:     &schema.Schema{Type: schema.TypeString},
								},
								"filter_ids": {
									Type:     schema.TypeSet,
									Set:      schema.HashString,
									Required: true,
									Elem:     &schema.Schema{Type: schema.TypeString},
								},
								"role_ids": {
									Type:     schema.TypeSet,
									Set:      schema.HashString,
									Required: true,
									Elem:     &schema.Schema{Type: schema.TypeString},
								},
							},
						},
					},
				}, map[string]interface{}{})

				// Set empty sets using d.Set() to ensure they're properly created
				emptyResourceRole := []interface{}{
					map[string]interface{}{
						"project_ids": schema.NewSet(schema.HashString, []interface{}{}),
						"filter_ids":  schema.NewSet(schema.HashString, []interface{}{}),
						"role_ids":    schema.NewSet(schema.HashString, []interface{}{}),
					},
				}
				_ = resourceData.Set("resource_role", emptyResourceRole)

				return resourceData
			},
			expected: []*models.V1ResourceRolesUpdateEntity{
				{
					ProjectUids: []string{},
					FilterRefs:  []string{},
					Roles:       []string{},
				},
			},
		},
		{
			name: "No resource role field",
			setup: func() *schema.ResourceData {
				resourceData := schema.TestResourceDataRaw(t, map[string]*schema.Schema{
					"resource_role": {
						Type: schema.TypeSet,
						Set:  resourceUserResourceRoleMappingHash,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"project_ids": {
									Type:     schema.TypeSet,
									Set:      schema.HashString,
									Required: true,
									Elem:     &schema.Schema{Type: schema.TypeString},
								},
								"filter_ids": {
									Type:     schema.TypeSet,
									Set:      schema.HashString,
									Required: true,
									Elem:     &schema.Schema{Type: schema.TypeString},
								},
								"role_ids": {
									Type:     schema.TypeSet,
									Set:      schema.HashString,
									Required: true,
									Elem:     &schema.Schema{Type: schema.TypeString},
								},
							},
						},
					},
				}, map[string]interface{}{})
				return resourceData
			},
			expected: nil,
		},
		{
			name: "Empty resource role list",
			setup: func() *schema.ResourceData {
				resourceData := schema.TestResourceDataRaw(t, map[string]*schema.Schema{
					"resource_role": {
						Type: schema.TypeSet,
						Set:  resourceUserResourceRoleMappingHash,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"project_ids": {
									Type:     schema.TypeSet,
									Set:      schema.HashString,
									Required: true,
									Elem:     &schema.Schema{Type: schema.TypeString},
								},
								"filter_ids": {
									Type:     schema.TypeSet,
									Set:      schema.HashString,
									Required: true,
									Elem:     &schema.Schema{Type: schema.TypeString},
								},
								"role_ids": {
									Type:     schema.TypeSet,
									Set:      schema.HashString,
									Required: true,
									Elem:     &schema.Schema{Type: schema.TypeString},
								},
							},
						},
					},
				}, map[string]interface{}{
					"resource_role": []interface{}{},
				})
				return resourceData
			},
			expected: []*models.V1ResourceRolesUpdateEntity{},
		},
		{
			name: "Single project, multiple filters and roles",
			setup: func() *schema.ResourceData {
				resourceData := schema.TestResourceDataRaw(t, map[string]*schema.Schema{
					"resource_role": {
						Type: schema.TypeSet,
						Set:  resourceUserResourceRoleMappingHash,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"project_ids": {
									Type:     schema.TypeSet,
									Set:      schema.HashString,
									Required: true,
									Elem:     &schema.Schema{Type: schema.TypeString},
								},
								"filter_ids": {
									Type:     schema.TypeSet,
									Set:      schema.HashString,
									Required: true,
									Elem:     &schema.Schema{Type: schema.TypeString},
								},
								"role_ids": {
									Type:     schema.TypeSet,
									Set:      schema.HashString,
									Required: true,
									Elem:     &schema.Schema{Type: schema.TypeString},
								},
							},
						},
					},
				}, map[string]interface{}{
					"resource_role": []interface{}{
						map[string]interface{}{
							"project_ids": []interface{}{"project1"},
							"filter_ids":  []interface{}{"filter1", "filter2", "filter3"},
							"role_ids":    []interface{}{"role1", "role2"},
						},
					},
				})
				return resourceData
			},
			expected: []*models.V1ResourceRolesUpdateEntity{
				{
					ProjectUids: []string{"project1"},
					FilterRefs:  []string{"filter1", "filter2", "filter3"},
					Roles:       []string{"role1", "role2"},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resourceData := tt.setup()
			result := toUserResourceRoleMapping(resourceData)

			if tt.expected == nil {
				assert.Nil(t, result)
			} else {
				assert.Equal(t, len(tt.expected), len(result), "Result length should match expected length")

				// Helper function to check if two entities match
				entitiesMatch := func(e1, e2 *models.V1ResourceRolesUpdateEntity) bool {
					projectMatch := len(e1.ProjectUids) == len(e2.ProjectUids)
					if projectMatch {
						e1ProjMap := make(map[string]bool)
						for _, p := range e1.ProjectUids {
							e1ProjMap[p] = true
						}
						for _, p := range e2.ProjectUids {
							if !e1ProjMap[p] {
								projectMatch = false
								break
							}
						}
					}

					filterMatch := len(e1.FilterRefs) == len(e2.FilterRefs)
					if filterMatch {
						e1FilterMap := make(map[string]bool)
						for _, f := range e1.FilterRefs {
							e1FilterMap[f] = true
						}
						for _, f := range e2.FilterRefs {
							if !e1FilterMap[f] {
								filterMatch = false
								break
							}
						}
					}

					roleMatch := len(e1.Roles) == len(e2.Roles)
					if roleMatch {
						e1RoleMap := make(map[string]bool)
						for _, r := range e1.Roles {
							e1RoleMap[r] = true
						}
						for _, r := range e2.Roles {
							if !e1RoleMap[r] {
								roleMatch = false
								break
							}
						}
					}

					return projectMatch && filterMatch && roleMatch
				}

				// For multiple items, compare sets without relying on order
				if len(tt.expected) > 1 {
					// Create a map to track which expected entities have been matched
					matched := make([]bool, len(tt.expected))

					for _, resultEntity := range result {
						found := false
						for i, expectedEntity := range tt.expected {
							if !matched[i] && entitiesMatch(expectedEntity, resultEntity) {
								matched[i] = true
								found = true
								break
							}
						}
						assert.True(t, found, "Result entity should match one of the expected entities")
					}

					// Ensure all expected entities were matched
					for i, m := range matched {
						assert.True(t, m, "Expected entity at index %d should be matched", i)
					}
				} else {
					// For single item, compare directly
					if len(tt.expected) > 0 && len(result) > 0 {
						assert.True(t, entitiesMatch(tt.expected[0], result[0]), "Entities should match")
					}
				}
			}
		})
	}
}

func TestToUserProjectRoleMapping(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() *schema.ResourceData
		expected *models.V1ProjectRolesPatch
	}{
		{
			name: "Single project role with multiple roles",
			setup: func() *schema.ResourceData {
				resourceData := schema.TestResourceDataRaw(t, map[string]*schema.Schema{
					"project_role": {
						Type: schema.TypeSet,
						Set:  resourceUserProjectRoleMappingHash,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"project_id": {
									Type:     schema.TypeString,
									Required: true,
								},
								"role_ids": {
									Type:     schema.TypeSet,
									Set:      schema.HashString,
									Required: true,
									Elem:     &schema.Schema{Type: schema.TypeString},
								},
							},
						},
					},
				}, map[string]interface{}{
					"project_role": []interface{}{
						map[string]interface{}{
							"project_id": "project1",
							"role_ids":   []interface{}{"role1", "role2", "role3"},
						},
					},
				})
				return resourceData
			},
			expected: &models.V1ProjectRolesPatch{
				Projects: []*models.V1ProjectRolesPatchProjectsItems0{
					{
						ProjectUID: "project1",
						Roles:      []string{"role1", "role2", "role3"},
					},
				},
			},
		},
		{
			name: "Multiple project roles",
			setup: func() *schema.ResourceData {
				resourceData := schema.TestResourceDataRaw(t, map[string]*schema.Schema{
					"project_role": {
						Type: schema.TypeSet,
						Set:  resourceUserProjectRoleMappingHash,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"project_id": {
									Type:     schema.TypeString,
									Required: true,
								},
								"role_ids": {
									Type:     schema.TypeSet,
									Set:      schema.HashString,
									Required: true,
									Elem:     &schema.Schema{Type: schema.TypeString},
								},
							},
						},
					},
				}, map[string]interface{}{
					"project_role": []interface{}{
						map[string]interface{}{
							"project_id": "project1",
							"role_ids":   []interface{}{"role1"},
						},
						map[string]interface{}{
							"project_id": "project2",
							"role_ids":   []interface{}{"role2", "role3"},
						},
					},
				})
				return resourceData
			},
			expected: &models.V1ProjectRolesPatch{
				Projects: []*models.V1ProjectRolesPatchProjectsItems0{
					{
						ProjectUID: "project1",
						Roles:      []string{"role1"},
					},
					{
						ProjectUID: "project2",
						Roles:      []string{"role2", "role3"},
					},
				},
			},
		},
		{
			name: "Single project with single role",
			setup: func() *schema.ResourceData {
				resourceData := schema.TestResourceDataRaw(t, map[string]*schema.Schema{
					"project_role": {
						Type: schema.TypeSet,
						Set:  resourceUserProjectRoleMappingHash,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"project_id": {
									Type:     schema.TypeString,
									Required: true,
								},
								"role_ids": {
									Type:     schema.TypeSet,
									Set:      schema.HashString,
									Required: true,
									Elem:     &schema.Schema{Type: schema.TypeString},
								},
							},
						},
					},
				}, map[string]interface{}{
					"project_role": []interface{}{
						map[string]interface{}{
							"project_id": "project1",
							"role_ids":   []interface{}{"role1"},
						},
					},
				})
				return resourceData
			},
			expected: &models.V1ProjectRolesPatch{
				Projects: []*models.V1ProjectRolesPatchProjectsItems0{
					{
						ProjectUID: "project1",
						Roles:      []string{"role1"},
					},
				},
			},
		},
		{
			name: "Empty role_ids set",
			setup: func() *schema.ResourceData {
				resourceData := schema.TestResourceDataRaw(t, map[string]*schema.Schema{
					"project_role": {
						Type: schema.TypeSet,
						Set:  resourceUserProjectRoleMappingHash,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"project_id": {
									Type:     schema.TypeString,
									Required: true,
								},
								"role_ids": {
									Type:     schema.TypeSet,
									Set:      schema.HashString,
									Required: true,
									Elem:     &schema.Schema{Type: schema.TypeString},
								},
							},
						},
					},
				}, map[string]interface{}{})

				// Set empty role_ids using d.Set() to ensure they're properly created
				emptyProjectRole := []interface{}{
					map[string]interface{}{
						"project_id": "project1",
						"role_ids":   schema.NewSet(schema.HashString, []interface{}{}),
					},
				}
				_ = resourceData.Set("project_role", emptyProjectRole)

				return resourceData
			},
			expected: &models.V1ProjectRolesPatch{
				Projects: []*models.V1ProjectRolesPatchProjectsItems0{
					{
						ProjectUID: "project1",
						Roles:      []string{},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resourceData := tt.setup()
			result := toUserProjectRoleMapping(resourceData)

			if tt.expected == nil {
				assert.Nil(t, result)
			} else {
				assert.NotNil(t, result)
				assert.Equal(t, len(tt.expected.Projects), len(result.Projects), "Projects length should match")

				// Helper function to check if two project items match
				projectsMatch := func(p1, p2 *models.V1ProjectRolesPatchProjectsItems0) bool {
					if p1.ProjectUID != p2.ProjectUID {
						return false
					}
					if len(p1.Roles) != len(p2.Roles) {
						return false
					}
					p1RoleMap := make(map[string]bool)
					for _, r := range p1.Roles {
						p1RoleMap[r] = true
					}
					for _, r := range p2.Roles {
						if !p1RoleMap[r] {
							return false
						}
					}
					return true
				}

				// For multiple items, compare sets without relying on order
				if len(tt.expected.Projects) > 1 {
					matched := make([]bool, len(tt.expected.Projects))

					for _, resultProject := range result.Projects {
						found := false
						for i, expectedProject := range tt.expected.Projects {
							if !matched[i] && projectsMatch(expectedProject, resultProject) {
								matched[i] = true
								found = true
								break
							}
						}
						assert.True(t, found, "Result project should match one of the expected projects")
					}

					// Ensure all expected projects were matched
					for i, m := range matched {
						assert.True(t, m, "Expected project at index %d should be matched", i)
					}
				} else {
					// For single item, compare directly
					if len(tt.expected.Projects) > 0 && len(result.Projects) > 0 {
						assert.True(t, projectsMatch(tt.expected.Projects[0], result.Projects[0]), "Projects should match")
					}
				}
			}
		})
	}
}

func TestToUserTenantRoleMapping(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() *schema.ResourceData
		expected *models.V1UserRoleUIDs
	}{
		{
			name: "Single tenant role",
			setup: func() *schema.ResourceData {
				resourceData := schema.TestResourceDataRaw(t, map[string]*schema.Schema{
					"tenant_role": {
						Type:     schema.TypeSet,
						Set:      schema.HashString,
						Optional: true,
						Elem:     &schema.Schema{Type: schema.TypeString},
					},
				}, map[string]interface{}{
					"tenant_role": []interface{}{"role1"},
				})
				return resourceData
			},
			expected: &models.V1UserRoleUIDs{
				Roles: []string{"role1"},
			},
		},
		{
			name: "Multiple tenant roles",
			setup: func() *schema.ResourceData {
				resourceData := schema.TestResourceDataRaw(t, map[string]*schema.Schema{
					"tenant_role": {
						Type:     schema.TypeSet,
						Set:      schema.HashString,
						Optional: true,
						Elem:     &schema.Schema{Type: schema.TypeString},
					},
				}, map[string]interface{}{
					"tenant_role": []interface{}{"role1", "role2", "role3"},
				})
				return resourceData
			},
			expected: &models.V1UserRoleUIDs{
				Roles: []string{"role1", "role2", "role3"},
			},
		},
		{
			name: "Empty tenant_role set",
			setup: func() *schema.ResourceData {
				resourceData := schema.TestResourceDataRaw(t, map[string]*schema.Schema{
					"tenant_role": {
						Type:     schema.TypeSet,
						Set:      schema.HashString,
						Optional: true,
						Elem:     &schema.Schema{Type: schema.TypeString},
					},
				}, map[string]interface{}{})

				// Set empty set using d.Set() to ensure it's properly created
				emptyTenantRole := schema.NewSet(schema.HashString, []interface{}{})
				_ = resourceData.Set("tenant_role", emptyTenantRole)

				return resourceData
			},
			expected: &models.V1UserRoleUIDs{
				Roles: []string{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resourceData := tt.setup()
			result := toUserTenantRoleMapping(resourceData)

			assert.NotNil(t, result, "Result should not be nil")
			assert.Equal(t, len(tt.expected.Roles), len(result.Roles), "Roles length should match")

			// Compare roles as sets (order-independent)
			expectedRoleMap := make(map[string]bool)
			for _, r := range tt.expected.Roles {
				expectedRoleMap[r] = true
			}

			resultRoleMap := make(map[string]bool)
			for _, r := range result.Roles {
				resultRoleMap[r] = true
			}

			assert.Equal(t, len(expectedRoleMap), len(resultRoleMap), "Role maps should have same length")

			for role := range expectedRoleMap {
				assert.True(t, resultRoleMap[role], "Role %s should be present in result", role)
			}

			for role := range resultRoleMap {
				assert.True(t, expectedRoleMap[role], "Role %s should be present in expected", role)
			}
		})
	}
}

func TestToUserWorkspaceRoleMapping(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() *schema.ResourceData
		expected *models.V1WorkspacesRolesPatch
	}{
		{
			name: "Single workspace role with single workspace and single role",
			setup: func() *schema.ResourceData {
				resourceData := schema.TestResourceDataRaw(t, map[string]*schema.Schema{
					"workspace_role": {
						Type: schema.TypeSet,
						Set:  resourceUserWorkspaceRoleMappingHash,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"project_id": {
									Type:     schema.TypeString,
									Required: true,
								},
								"workspace": {
									Type:     schema.TypeSet,
									Set:      resourceUserWorkspaceRoleMappingHashInternal,
									Required: true,
									Elem: &schema.Resource{
										Schema: map[string]*schema.Schema{
											"id": {
												Type:     schema.TypeString,
												Required: true,
											},
											"role_ids": {
												Type:     schema.TypeSet,
												Set:      schema.HashString,
												Required: true,
												Elem:     &schema.Schema{Type: schema.TypeString},
											},
										},
									},
								},
							},
						},
					},
				}, map[string]interface{}{
					"workspace_role": []interface{}{
						map[string]interface{}{
							"project_id": "project1",
							"workspace": []interface{}{
								map[string]interface{}{
									"id":       "workspace1",
									"role_ids": []interface{}{"role1"},
								},
							},
						},
					},
				})
				return resourceData
			},
			expected: &models.V1WorkspacesRolesPatch{
				Workspaces: []*models.V1WorkspaceRolesPatch{
					{
						UID:   "workspace1",
						Roles: []string{"role1"},
					},
				},
			},
		},
		{
			name: "Single workspace role with single workspace and multiple roles",
			setup: func() *schema.ResourceData {
				resourceData := schema.TestResourceDataRaw(t, map[string]*schema.Schema{
					"workspace_role": {
						Type: schema.TypeSet,
						Set:  resourceUserWorkspaceRoleMappingHash,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"project_id": {
									Type:     schema.TypeString,
									Required: true,
								},
								"workspace": {
									Type:     schema.TypeSet,
									Set:      resourceUserWorkspaceRoleMappingHashInternal,
									Required: true,
									Elem: &schema.Resource{
										Schema: map[string]*schema.Schema{
											"id": {
												Type:     schema.TypeString,
												Required: true,
											},
											"role_ids": {
												Type:     schema.TypeSet,
												Set:      schema.HashString,
												Required: true,
												Elem:     &schema.Schema{Type: schema.TypeString},
											},
										},
									},
								},
							},
						},
					},
				}, map[string]interface{}{
					"workspace_role": []interface{}{
						map[string]interface{}{
							"project_id": "project1",
							"workspace": []interface{}{
								map[string]interface{}{
									"id":       "workspace1",
									"role_ids": []interface{}{"role1", "role2", "role3"},
								},
							},
						},
					},
				})
				return resourceData
			},
			expected: &models.V1WorkspacesRolesPatch{
				Workspaces: []*models.V1WorkspaceRolesPatch{
					{
						UID:   "workspace1",
						Roles: []string{"role1", "role2", "role3"},
					},
				},
			},
		},
		{
			name: "Single workspace role with multiple workspaces",
			setup: func() *schema.ResourceData {
				resourceData := schema.TestResourceDataRaw(t, map[string]*schema.Schema{
					"workspace_role": {
						Type: schema.TypeSet,
						Set:  resourceUserWorkspaceRoleMappingHash,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"project_id": {
									Type:     schema.TypeString,
									Required: true,
								},
								"workspace": {
									Type:     schema.TypeSet,
									Set:      resourceUserWorkspaceRoleMappingHashInternal,
									Required: true,
									Elem: &schema.Resource{
										Schema: map[string]*schema.Schema{
											"id": {
												Type:     schema.TypeString,
												Required: true,
											},
											"role_ids": {
												Type:     schema.TypeSet,
												Set:      schema.HashString,
												Required: true,
												Elem:     &schema.Schema{Type: schema.TypeString},
											},
										},
									},
								},
							},
						},
					},
				}, map[string]interface{}{
					"workspace_role": []interface{}{
						map[string]interface{}{
							"project_id": "project1",
							"workspace": []interface{}{
								map[string]interface{}{
									"id":       "workspace1",
									"role_ids": []interface{}{"role1"},
								},
								map[string]interface{}{
									"id":       "workspace2",
									"role_ids": []interface{}{"role2", "role3"},
								},
							},
						},
					},
				})
				return resourceData
			},
			expected: &models.V1WorkspacesRolesPatch{
				Workspaces: []*models.V1WorkspaceRolesPatch{
					{
						UID:   "workspace1",
						Roles: []string{"role1"},
					},
					{
						UID:   "workspace2",
						Roles: []string{"role2", "role3"},
					},
				},
			},
		},
		{
			name: "Multiple workspace roles (different projects)",
			setup: func() *schema.ResourceData {
				resourceData := schema.TestResourceDataRaw(t, map[string]*schema.Schema{
					"workspace_role": {
						Type: schema.TypeSet,
						Set:  resourceUserWorkspaceRoleMappingHash,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"project_id": {
									Type:     schema.TypeString,
									Required: true,
								},
								"workspace": {
									Type:     schema.TypeSet,
									Set:      resourceUserWorkspaceRoleMappingHashInternal,
									Required: true,
									Elem: &schema.Resource{
										Schema: map[string]*schema.Schema{
											"id": {
												Type:     schema.TypeString,
												Required: true,
											},
											"role_ids": {
												Type:     schema.TypeSet,
												Set:      schema.HashString,
												Required: true,
												Elem:     &schema.Schema{Type: schema.TypeString},
											},
										},
									},
								},
							},
						},
					},
				}, map[string]interface{}{
					"workspace_role": []interface{}{
						map[string]interface{}{
							"project_id": "project1",
							"workspace": []interface{}{
								map[string]interface{}{
									"id":       "workspace1",
									"role_ids": []interface{}{"role1"},
								},
							},
						},
						map[string]interface{}{
							"project_id": "project2",
							"workspace": []interface{}{
								map[string]interface{}{
									"id":       "workspace2",
									"role_ids": []interface{}{"role2"},
								},
							},
						},
					},
				})
				return resourceData
			},
			expected: &models.V1WorkspacesRolesPatch{
				Workspaces: []*models.V1WorkspaceRolesPatch{
					{
						UID:   "workspace1",
						Roles: []string{"role1"},
					},
					{
						UID:   "workspace2",
						Roles: []string{"role2"},
					},
				},
			},
		},
		{
			name: "Empty role_ids set",
			setup: func() *schema.ResourceData {
				resourceData := schema.TestResourceDataRaw(t, map[string]*schema.Schema{
					"workspace_role": {
						Type: schema.TypeSet,
						Set:  resourceUserWorkspaceRoleMappingHash,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"project_id": {
									Type:     schema.TypeString,
									Required: true,
								},
								"workspace": {
									Type:     schema.TypeSet,
									Set:      resourceUserWorkspaceRoleMappingHashInternal,
									Required: true,
									Elem: &schema.Resource{
										Schema: map[string]*schema.Schema{
											"id": {
												Type:     schema.TypeString,
												Required: true,
											},
											"role_ids": {
												Type:     schema.TypeSet,
												Set:      schema.HashString,
												Required: true,
												Elem:     &schema.Schema{Type: schema.TypeString},
											},
										},
									},
								},
							},
						},
					},
				}, map[string]interface{}{})

				// Set empty role_ids using d.Set() to ensure they're properly created
				emptyWorkspaceRole := []interface{}{
					map[string]interface{}{
						"project_id": "project1",
						"workspace": []interface{}{
							map[string]interface{}{
								"id":       "workspace1",
								"role_ids": schema.NewSet(schema.HashString, []interface{}{}),
							},
						},
					},
				}
				_ = resourceData.Set("workspace_role", emptyWorkspaceRole)

				return resourceData
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
			resourceData := tt.setup()
			result := toUserWorkspaceRoleMapping(resourceData)

			assert.NotNil(t, result, "Result should not be nil")
			assert.Equal(t, len(tt.expected.Workspaces), len(result.Workspaces), "Workspaces length should match")

			// Helper function to check if two workspace items match
			workspacesMatch := func(w1, w2 *models.V1WorkspaceRolesPatch) bool {
				if w1.UID != w2.UID {
					return false
				}
				if len(w1.Roles) != len(w2.Roles) {
					return false
				}
				w1RoleMap := make(map[string]bool)
				for _, r := range w1.Roles {
					w1RoleMap[r] = true
				}
				for _, r := range w2.Roles {
					if !w1RoleMap[r] {
						return false
					}
				}
				return true
			}

			// For multiple items, compare sets without relying on order
			if len(tt.expected.Workspaces) > 1 {
				matched := make([]bool, len(tt.expected.Workspaces))

				for _, resultWorkspace := range result.Workspaces {
					found := false
					for i, expectedWorkspace := range tt.expected.Workspaces {
						if !matched[i] && workspacesMatch(expectedWorkspace, resultWorkspace) {
							matched[i] = true
							found = true
							break
						}
					}
					assert.True(t, found, "Result workspace should match one of the expected workspaces")
				}

				// Ensure all expected workspaces were matched
				for i, m := range matched {
					assert.True(t, m, "Expected workspace at index %d should be matched", i)
				}
			} else {
				// For single item, compare directly
				if len(tt.expected.Workspaces) > 0 && len(result.Workspaces) > 0 {
					assert.True(t, workspacesMatch(tt.expected.Workspaces[0], result.Workspaces[0]), "Workspaces should match")
				}
			}
		})
	}
}
