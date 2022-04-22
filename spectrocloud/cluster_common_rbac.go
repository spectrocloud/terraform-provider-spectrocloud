package spectrocloud

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/hapi/models"
	"github.com/spectrocloud/terraform-provider-spectrocloud/pkg/client"
)

func toClusterRBACs(d *schema.ResourceData) []*models.V1ClusterRbacInputEntity {
	clusterRbacs := make([]*models.V1ClusterRbacInputEntity, 0)
	clusterRbacBindings := make([]*models.V1ClusterRbacBinding, 0)
	rbacBindings := make([]*models.V1ClusterRbacBinding, 0)

	if d.Get("cluster_rbac_binding") == nil {
		return nil
	}
	for _, clusterRbac := range d.Get("cluster_rbac_binding").([]interface{}) {
		b := toClusterRBAC(clusterRbac)
		for _, binding := range b.Spec.Bindings {
			switch binding.Role.Kind {
			case "ClusterRole":
				clusterRbacBindings = append(clusterRbacBindings, binding)
				break
			case "Role":
				rbacBindings = append(rbacBindings, binding)
				break
			default:
				break
			}
		}
	}

	clusterRbacs = append(clusterRbacs, &models.V1ClusterRbacInputEntity{
		Spec: &models.V1ClusterRbacSpec{
			Bindings: clusterRbacBindings,
		},
	})

	clusterRbacs = append(clusterRbacs, &models.V1ClusterRbacInputEntity{
		Spec: &models.V1ClusterRbacSpec{
			Bindings: rbacBindings,
		},
	})

	return clusterRbacs
}

func toClusterRBAC(clusterRbacBinding interface{}) *models.V1ClusterRbacInputEntity {
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

	ret := &models.V1ClusterRbacInputEntity{
		Spec: &models.V1ClusterRbacSpec{
			Bindings: bindings,
		},
	}

	return ret

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
	if rbacs := toClusterRBACs(d); rbacs != nil {
		return c.ApplyClusterRbacConfig(d.Id(), toUpdateClusterRbac(rbacs))
	}
	return nil
}

func toUpdateClusterRbac(rbacs []*models.V1ClusterRbacInputEntity) *models.V1ClusterRbac {
	bindings := make([]*models.V1ClusterRbacBinding, 0)

	for _, rbac := range rbacs {
		for _, binding := range rbac.Spec.Bindings {
			bindings = append(bindings, binding)
		}
	}

	return &models.V1ClusterRbac{
		Spec: &models.V1ClusterRbacSpec{
			Bindings: bindings,
		},
	}
}
