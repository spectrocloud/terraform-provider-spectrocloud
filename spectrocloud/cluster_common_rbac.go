package spectrocloud

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/spectrocloud/hapi/models"
	"github.com/spectrocloud/palette-sdk-go/client"
)

func toClusterRBACsInputEntities(d *schema.ResourceData) []*models.V1ClusterRbacInputEntity {
	rbacs := toRbacInputEntities(toClusterRbac(GetBindings(d)))
	return rbacs
}

func GetBindings(d *schema.ResourceData) []*models.V1ClusterRbacBinding {
	bindings := make([]*models.V1ClusterRbacBinding, 0)

	if d.Get("cluster_rbac_binding") == nil {
		return nil
	}
	for _, clusterRbac := range d.Get("cluster_rbac_binding").([]interface{}) {
		copy(bindings, toClusterRBACBindings(clusterRbac))
	}

	return bindings
}

func toClusterRbac(bindings []*models.V1ClusterRbacBinding) *models.V1ClusterRbac {
	return &models.V1ClusterRbac{
		Spec: &models.V1ClusterRbacSpec{
			Bindings: bindings,
		},
	}
}

func toRbacInputEntities(config *models.V1ClusterRbac) []*models.V1ClusterRbacInputEntity {
	rbacs := make([]*models.V1ClusterRbacInputEntity, 0)

	clusterRoleBindings := make([]*models.V1ClusterRbacBinding, 0)
	roleBindings := make([]*models.V1ClusterRbacBinding, 0)

	for _, binding := range config.Spec.Bindings {
		switch binding.Type {
		case "ClusterRoleBinding":
			clusterRoleBindings = append(clusterRoleBindings, binding)
		case "RoleBinding":
			roleBindings = append(roleBindings, binding)
		}

	}

	if len(clusterRoleBindings) > 0 {
		rbacs = append(rbacs, &models.V1ClusterRbacInputEntity{
			Spec: &models.V1ClusterRbacSpec{
				Bindings: clusterRoleBindings,
			},
		})
	}

	if len(roleBindings) > 0 {
		rbacs = append(rbacs, &models.V1ClusterRbacInputEntity{
			Spec: &models.V1ClusterRbacSpec{
				Bindings: roleBindings,
			},
		})
	}
	return rbacs
}

func toClusterRBACBindings(clusterRbacBinding interface{}) []*models.V1ClusterRbacBinding {
	m := clusterRbacBinding.(map[string]interface{})

	role, _ := m["role"].(map[string]interface{})

	namespace := m["namespace"].(string)
	bindings := make([]*models.V1ClusterRbacBinding, 0)
	subjects := make([]*models.V1ClusterRbacSubjects, 0)

	for _, val := range m["subjects"].([]interface{}) {
		subjectValue := val.(map[string]interface{})
		var subjectType string
		if subjectValue["type"] != nil {
			subjectType = subjectValue["type"].(string)
		}
		subject := &models.V1ClusterRbacSubjects{
			Name:      subjectValue["name"].(string),
			Namespace: subjectValue["namespace"].(string),
			Type:      subjectType,
		}
		subjects = append(subjects, subject)
	}

	bindings = append(bindings, &models.V1ClusterRbacBinding{
		Type: m["type"].(string),
		Role: &models.V1ClusterRoleRef{
			Kind: role["kind"].(string),
			Name: role["name"].(string),
		},
		Namespace: namespace,
		Subjects:  subjects,
	})

	return bindings

}

func flattenClusterRBAC(items []*models.V1ClusterRbac) []interface{} {
	result := make([]interface{}, 0)
	for _, rbac := range items {
		for _, binding := range rbac.Spec.Bindings {
			flattenRbac := make(map[string]interface{})
			flattenRbac["type"] = binding.Type
			flattenRbac["namespace"] = binding.Namespace

			flattenRole := make(map[string]interface{})
			flattenRole["kind"] = binding.Role.Kind
			flattenRole["name"] = binding.Role.Name
			flattenRbac["role"] = flattenRole

			subjects := make([]interface{}, 0)
			for _, subject := range binding.Subjects {
				flattenSubject := make(map[string]interface{})
				flattenSubject["type"] = subject.Type
				flattenSubject["name"] = subject.Name
				flattenSubject["namespace"] = subject.Namespace
				subjects = append(subjects, flattenSubject)
			}

			flattenRbac["subjects"] = subjects

			result = append(result, flattenRbac)
		}
	}
	return result
}

func updateClusterRBAC(c *client.V1Client, d *schema.ResourceData) error {
	if rbacs := toClusterRBACsInputEntities(d); rbacs != nil {
		return c.ApplyClusterRbacConfig(d.Id(), rbacs)
	}
	return nil
}
