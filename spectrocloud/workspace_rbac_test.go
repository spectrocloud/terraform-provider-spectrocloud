package spectrocloud

import (
	"testing"

	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/stretchr/testify/assert"
)

func TestWorkspaceRbacBindingType(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		rbac     *models.V1ClusterRbac
		expected string
	}{
		{
			name:     "nil rbac",
			rbac:     nil,
			expected: "",
		},
		{
			name: "cluster role binding",
			rbac: &models.V1ClusterRbac{
				Spec: &models.V1ClusterRbacSpec{
					Bindings: []*models.V1ClusterRbacBinding{
						{Type: "ClusterRoleBinding"},
					},
				},
			},
			expected: "ClusterRoleBinding",
		},
		{
			name: "role binding",
			rbac: &models.V1ClusterRbac{
				Spec: &models.V1ClusterRbacSpec{
					Bindings: []*models.V1ClusterRbacBinding{
						{Type: "RoleBinding", Namespace: "default"},
					},
				},
			},
			expected: "RoleBinding",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.expected, workspaceRbacBindingType(tt.rbac))
		})
	}
}
