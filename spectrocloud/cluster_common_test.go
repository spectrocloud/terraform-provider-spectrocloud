package spectrocloud

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/gomi/pkg/ptr"
	"reflect"
	"sort"
	"testing"

	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/stretchr/testify/assert"
)

func TestToAdditionalNodePoolLabels(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]interface{}
		expected map[string]string
	}{
		{
			name:     "Nil additional_labels",
			input:    map[string]interface{}{"additional_labels": nil},
			expected: map[string]string{},
		},
		{
			name:     "Empty additional_labels",
			input:    map[string]interface{}{"additional_labels": map[string]interface{}{}},
			expected: map[string]string{},
		},
		{
			name: "Valid additional_labels",
			input: map[string]interface{}{
				"additional_labels": map[string]interface{}{
					"label1": "value1",
					"label2": "value2",
				},
			},
			expected: map[string]string{
				"label1": "value1",
				"label2": "value2",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := toAdditionalNodePoolLabels(tt.input)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestToClusterTaints(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]interface{}
		expected []*models.V1Taint
	}{
		{
			name:     "Nil taints",
			input:    map[string]interface{}{"taints": nil},
			expected: nil,
		},
		{
			name:     "Empty taints",
			input:    map[string]interface{}{"taints": []interface{}{}},
			expected: []*models.V1Taint{},
		},
		{
			name: "Valid taints",
			input: map[string]interface{}{
				"taints": []interface{}{
					map[string]interface{}{
						"key":    "key1",
						"value":  "value1",
						"effect": "NoSchedule",
					},
					map[string]interface{}{
						"key":    "key2",
						"value":  "value2",
						"effect": "PreferNoSchedule",
					},
				},
			},
			expected: []*models.V1Taint{
				{
					Key:    "key1",
					Value:  "value1",
					Effect: "NoSchedule",
				},
				{
					Key:    "key2",
					Value:  "value2",
					Effect: "PreferNoSchedule",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := toClusterTaints(tt.input)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestToClusterTaint(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]interface{}
		expected *models.V1Taint
	}{
		{
			name: "Valid cluster taint",
			input: map[string]interface{}{
				"key":    "key1",
				"value":  "value1",
				"effect": "NoSchedule",
			},
			expected: &models.V1Taint{
				Key:    "key1",
				Value:  "value1",
				Effect: "NoSchedule",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := toClusterTaint(tt.input)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestFlattenClusterTaints(t *testing.T) {
	taint1 := &models.V1Taint{
		Key:    "key1",
		Value:  "value1",
		Effect: "NoSchedule",
	}
	taint2 := &models.V1Taint{
		Key:    "key2",
		Value:  "value2",
		Effect: "PreferNoSchedule",
	}

	tests := []struct {
		name     string
		input    []*models.V1Taint
		expected []interface{}
	}{
		{
			name:     "Empty items",
			input:    []*models.V1Taint{},
			expected: []interface{}{},
		},
		{
			name:  "Valid taints",
			input: []*models.V1Taint{taint1, taint2},
			expected: []interface{}{
				map[string]interface{}{
					"key":    "key1",
					"value":  "value1",
					"effect": "NoSchedule",
				},
				map[string]interface{}{
					"key":    "key2",
					"value":  "value2",
					"effect": "PreferNoSchedule",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := flattenClusterTaints(tt.input)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestFlattenAdditionalLabelsAndTaints(t *testing.T) {
	tests := []struct {
		name     string
		labels   map[string]string
		taints   []*models.V1Taint
		expected map[string]interface{}
	}{
		{
			name:     "Empty labels and taints",
			labels:   make(map[string]string),
			taints:   []*models.V1Taint{},
			expected: map[string]interface{}{"additional_labels": map[string]interface{}{}},
		},
		{
			name:   "Non-empty labels",
			labels: map[string]string{"label1": "value1", "label2": "value2"},
			taints: []*models.V1Taint{},
			expected: map[string]interface{}{
				"additional_labels": map[string]string{
					"label1": "value1",
					"label2": "value2",
				},
			},
		},
		{
			name:   "Non-empty labels and taints",
			labels: map[string]string{"label1": "value1", "label2": "value2"},
			taints: []*models.V1Taint{
				{
					Key:    "key1",
					Value:  "value1",
					Effect: "NoSchedule",
				},
				{
					Key:    "key2",
					Value:  "value2",
					Effect: "PreferNoSchedule",
				},
			},
			expected: map[string]interface{}{
				"additional_labels": map[string]string{
					"label1": "value1",
					"label2": "value2",
				},
				"taints": []interface{}{
					map[string]interface{}{
						"key":    "key1",
						"value":  "value1",
						"effect": "NoSchedule",
					},
					map[string]interface{}{
						"key":    "key2",
						"value":  "value2",
						"effect": "PreferNoSchedule",
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oi := make(map[string]interface{})
			FlattenAdditionalLabelsAndTaints(tt.labels, tt.taints, oi)
			if !reflect.DeepEqual(oi, tt.expected) {
				t.Logf("Expected: %#v\n", tt.expected)
				t.Logf("Actual: %#v\n", oi)
				t.Errorf("Test %s failed. Expected %#v, got %#v", tt.name, tt.expected, oi)
			}
		})
	}
}

func TestUpdateClusterRBAC(t *testing.T) {
	d := resourceClusterVsphere().TestResourceData()

	// Case 1: rbacs context is invalid
	err := d.Set("context", "invalid")
	if err != nil {
		return
	}
	err = updateClusterRBAC(nil, d)
	if err == nil || err.Error() != "invalid Context set - invalid" {
		t.Errorf("Expected 'invalid Context set - invalid', got %v", err)
	}
}

//func TestRepaveApprovalCheck(t *testing.T) {
//
//	d := resourceClusterAws().TestResourceData()
//	err := d.Set("review_repave_state", "Approved")
//	if err != nil {
//		return
//	}
//	err = d.Set("context", "tenant")
//	if err != nil {
//		return
//	}
//	d.SetId("TestclusterUID")
//
//	m := &client.V1Client{}
//
//	// Test case where repave state is pending and approve_system_repave is true
//	err = validateSystemRepaveApproval(d, m)
//	if err != nil {
//		t.Errorf("Unexpected error: %s", err)
//	}
//
//	// Test case where repave state is pending and approve_system_repave is false
//	m = &client.V1Client{}
//
//	err = d.Set("review_repave_state", "")
//	if err != nil {
//		return
//	}
//	err = validateSystemRepaveApproval(d, m)
//	expectedErrMsg := "cluster repave state is pending. \nDue to the following reasons -  \nPackValuesUpdated\nKindly verify the cluster and set `review_repave_state` to `Approved` to continue the repave operation and day 2 operation on the cluster."
//	if err == nil || err.Error() != expectedErrMsg {
//		t.Errorf("Expected error message '%s', got '%s'", expectedErrMsg, err)
//	}
//
//}

func prepareSpectroClusterModel() *models.V1SpectroCluster {

	scp := &models.V1SpectroCluster{
		APIVersion: "V1",
		Kind:       "",
		Metadata: &models.V1ObjectMeta{
			Annotations: map[string]string{
				"test_annotation": "tf",
				"scope":           "project",
			},
			CreationTimestamp: models.V1Time{},
			DeletionTimestamp: models.V1Time{},
			Labels: map[string]string{
				"test_label": "tf",
			},
			LastModifiedTimestamp: models.V1Time{},
			Name:                  "spc-cluster-unit-test",
			UID:                   "test-cluster-uid",
		},
		Spec: &models.V1SpectroClusterSpec{
			CloudConfigRef: &models.V1ObjectReference{
				Kind: "",
				Name: "spc-cluster-unit-tes",
				UID:  "test-cloud-config-uid",
			},
			CloudType: "vsphere",
			ClusterConfig: &models.V1ClusterConfig{
				ClusterMetaAttribute: "test-cluster-meta-attributes",
				ClusterResources: &models.V1ClusterResources{
					Namespaces: []*models.V1ResourceReference{},
					Rbacs:      []*models.V1ResourceReference{},
				},
				ControlPlaneHealthCheckTimeout: "",
				HostClusterConfig: &models.V1HostClusterConfig{
					ClusterEndpoint: &models.V1HostClusterEndpoint{
						Config: &models.V1HostClusterEndpointConfig{
							IngressConfig: &models.V1IngressConfig{
								Host: "121.1.1.0",
								Port: 9999,
							},
							LoadBalancerConfig: nil,
						},
						Type: "ingress",
					},
					ClusterGroup: &models.V1ObjectReference{
						Kind: "",
						Name: "",
						UID:  "test-cluster-group-uid",
					},
					HostCluster: &models.V1ObjectReference{
						Kind: "",
						Name: "",
						UID:  "test-host-cluster-uid",
					},
					IsHostCluster: ptr.BoolPtr(false),
				},
				LifecycleConfig: &models.V1LifecycleConfig{
					Pause: ptr.BoolPtr(false),
				},
				MachineHealthConfig: &models.V1MachineHealthCheckConfig{
					HealthCheckMaxUnhealthy:         "",
					NetworkReadyHealthCheckDuration: "",
					NodeReadyHealthCheckDuration:    "",
				},
				MachineManagementConfig: &models.V1MachineManagementConfig{
					OsPatchConfig: &models.V1OsPatchConfig{
						OnDemandPatchAfter: models.V1Time{},
						PatchOnBoot:        false,
						RebootIfRequired:   false,
						Schedule:           "",
					},
				},
				UpdateWorkerPoolsInParallel: false,
			},
			ClusterProfileTemplates: nil,
			ClusterType:             "full",
		},
		Status: &models.V1SpectroClusterStatus{
			AbortTimestamp: models.V1Time{},
			AddOnServices:  nil,
			APIEndpoints:   nil,
			ClusterImport:  nil,
			Conditions:     nil,
			Fips:           nil,
			Location:       nil,
			Packs:          nil,
			ProfileStatus:  nil,
			Repave:         nil,
			Services:       nil,
			SpcApply:       nil,
			State:          "",
			Upgrades:       nil,
			Virtual:        nil,
		},
	}
	return scp
}

//func TestReadCommonFieldsCluster(t *testing.T) {
//	d := prepareClusterVsphereTestData()
//	spc := prepareSpectroClusterModel()
//	c := getClientForCluster()
//	_, done := readCommonFields(c, d, spc)
//	assert.Equal(t, false, done)
//}

//func TestReadCommonFieldsVirtualCluster(t *testing.T) {
//	d := resourceClusterVirtual().TestResourceData()
//	spc := prepareSpectroClusterModel()
//	c := getClientForCluster()
//	_, done := readCommonFields(c, d, spc)
//	assert.Equal(t, false, done)
//}

func TestToSSHKeys(t *testing.T) {
	// Test case 1: When cloudConfig has "ssh_key" attribute
	cloudConfig1 := map[string]interface{}{
		"ssh_key": "ssh-key-1",
	}
	keys1, err1 := toSSHKeys(cloudConfig1)
	assert.NoError(t, err1)
	assert.Equal(t, []string{"ssh-key-1"}, keys1)

	// Test case 2: When cloudConfig has "ssh_keys" attribute
	cloudConfig2 := map[string]interface{}{
		"ssh_keys": schema.NewSet(schema.HashString, []interface{}{"ssh-key-2", "ssh-key-3"}),
	}
	keys2, err2 := toSSHKeys(cloudConfig2)
	assert.NoError(t, err2)
	assert.Equal(t, []string{"ssh-key-2", "ssh-key-3"}, keys2)

	// Test case 3: When cloudConfig has both "ssh_key" and "ssh_keys" attributes
	cloudConfig3 := map[string]interface{}{
		"ssh_key":  "ssh-key-4",
		"ssh_keys": schema.NewSet(schema.HashString, []interface{}{"ssh-key-5", "ssh-key-6"}),
	}

	keys3, err3 := toSSHKeys(cloudConfig3)
	sort.Strings(keys3)
	assert.NoError(t, err3)
	assert.Equal(t, []string{"ssh-key-4", "ssh-key-5", "ssh-key-6"}, keys3)

	// Test case 4: When cloudConfig has neither "ssh_key" nor "ssh_keys" attributes
	cloudConfig4 := map[string]interface{}{}
	keys4, err4 := toSSHKeys(cloudConfig4)
	assert.Error(t, err4)
	assert.Nil(t, keys4)
	assert.Equal(t, "validation ssh_key: Kindly specify any one attribute ssh_key or ssh_keys", err4.Error())
}

func TestValidateReviewRepaveValue(t *testing.T) {
	// Valid repave values
	validValues := []string{"", "Approved", "Pending"}

	for _, value := range validValues {
		warns, errs := validateReviewRepaveValue(value, "review_repave_state")
		assert.Empty(t, errs, "Expected no errors for valid repave value")
		if value == "Approved" {
			assert.NotEmpty(t, warns, "Expected warning for 'Approved' repave value")
		} else {
			assert.Empty(t, warns, "Expected no warnings for valid repave value")
		}
	}

	// Invalid repave value
	invalidValue := "InvalidStatus"
	warns, errs := validateReviewRepaveValue(invalidValue, "review_repave_state")
	assert.NotEmpty(t, errs, "Expected error for invalid repave value")
	assert.Empty(t, warns, "Expected no warnings for invalid repave value")
	expectedError := fmt.Sprintf("expected review_repave_state to be one of [``, `Pending`, `Approved`], got %s", invalidValue)
	assert.Equal(t, expectedError, errs[0].Error(), "Expected specific error message for invalid repave value")
}

func TestGeneralWarningForRepave(t *testing.T) {
	var diags diag.Diagnostics

	generalWarningForRepave(&diags)

	expectedMessage := "Please note that certain day 2 operations on a running cluster may trigger a node pool repave or a full repave of your cluster. This process might temporarily affect your cluster’s performance or configuration. For more details, please refer to the https://docs.spectrocloud.com/clusters/cluster-management/node-pool/"

	assert.Len(t, diags, 1)
	assert.Equal(t, diag.Warning, diags[0].Severity)
	assert.Equal(t, "Warning", diags[0].Summary)
	assert.Equal(t, expectedMessage, diags[0].Detail)
}

func TestToClusterRBACsInputEntities(t *testing.T) {
	d := resourceClusterGcp().TestResourceData()
	var clusterBinding []interface{}
	clusterBinding = append(clusterBinding, map[string]interface{}{
		"type":      "ClusterRoleBinding",
		"namespace": "default",
		"role": map[string]interface{}{
			"kind": "ClusterRole",
			"name": "admin",
		},
		"subjects": []interface{}{
			map[string]interface{}{
				"type":      "User",
				"name":      "admin-user",
				"namespace": "default",
			},
		},
	})
	clusterBinding = append(clusterBinding, map[string]interface{}{
		"type":      "RoleBinding",
		"namespace": "default",
		"role": map[string]interface{}{
			"kind": "Role",
			"name": "edit",
		},
		"subjects": []interface{}{
			map[string]interface{}{
				"type":      "Group",
				"name":      "editors",
				"namespace": "default",
			},
		},
	})
	err := d.Set("cluster_rbac_binding", clusterBinding)
	if err != nil {
		return
	}

	rbacs := toClusterRBACsInputEntities(d)

	assert.Len(t, rbacs, 2)

	assert.Equal(t, "ClusterRoleBinding", rbacs[0].Spec.Bindings[0].Type)
	assert.Equal(t, "ClusterRole", rbacs[0].Spec.Bindings[0].Role.Kind)
	assert.Equal(t, "admin", rbacs[0].Spec.Bindings[0].Role.Name)
	assert.Equal(t, "User", rbacs[0].Spec.Bindings[0].Subjects[0].Type)
	assert.Equal(t, "admin-user", rbacs[0].Spec.Bindings[0].Subjects[0].Name)

	assert.Equal(t, "RoleBinding", rbacs[1].Spec.Bindings[0].Type)
	assert.Equal(t, "Role", rbacs[1].Spec.Bindings[0].Role.Kind)
	assert.Equal(t, "edit", rbacs[1].Spec.Bindings[0].Role.Name)
	assert.Equal(t, "Group", rbacs[1].Spec.Bindings[0].Subjects[0].Type)
	assert.Equal(t, "editors", rbacs[1].Spec.Bindings[0].Subjects[0].Name)
}

func TestFlattenClusterRBAC(t *testing.T) {
	// Setup test data
	clusterRBACs := []*models.V1ClusterRbac{
		{
			Spec: &models.V1ClusterRbacSpec{
				Bindings: []*models.V1ClusterRbacBinding{
					{
						Type:      "ClusterRoleBinding",
						Namespace: "default",
						Role: &models.V1ClusterRoleRef{
							Kind: "ClusterRole",
							Name: "admin",
						},
						Subjects: []*models.V1ClusterRbacSubjects{
							{
								Type:      "User",
								Name:      "admin-user",
								Namespace: "default",
							},
						},
					},
				},
			},
		},
		{
			Spec: &models.V1ClusterRbacSpec{
				Bindings: []*models.V1ClusterRbacBinding{
					{
						Type:      "RoleBinding",
						Namespace: "kube-system",
						Role: &models.V1ClusterRoleRef{
							Kind: "Role",
							Name: "edit",
						},
						Subjects: []*models.V1ClusterRbacSubjects{
							{
								Type:      "Group",
								Name:      "editors",
								Namespace: "kube-system",
							},
						},
					},
				},
			},
		},
	}

	// Execute the function under test
	flattenedRBACs := flattenClusterRBAC(clusterRBACs)

	// Validate the results
	assert.Len(t, flattenedRBACs, 2)

	// First RBAC entry
	firstRBAC := flattenedRBACs[0].(map[string]interface{})
	assert.Equal(t, "ClusterRoleBinding", firstRBAC["type"])
	assert.Equal(t, "default", firstRBAC["namespace"])

	firstRole := firstRBAC["role"].(map[string]interface{})
	assert.Equal(t, "ClusterRole", firstRole["kind"])
	assert.Equal(t, "admin", firstRole["name"])

	firstSubjects := firstRBAC["subjects"].([]interface{})
	assert.Len(t, firstSubjects, 1)
	firstSubject := firstSubjects[0].(map[string]interface{})
	assert.Equal(t, "User", firstSubject["type"])
	assert.Equal(t, "admin-user", firstSubject["name"])
	assert.Equal(t, "default", firstSubject["namespace"])

	// Second RBAC entry
	secondRBAC := flattenedRBACs[1].(map[string]interface{})
	assert.Equal(t, "RoleBinding", secondRBAC["type"])
	assert.Equal(t, "kube-system", secondRBAC["namespace"])

	secondRole := secondRBAC["role"].(map[string]interface{})
	assert.Equal(t, "Role", secondRole["kind"])
	assert.Equal(t, "edit", secondRole["name"])

	secondSubjects := secondRBAC["subjects"].([]interface{})
	assert.Len(t, secondSubjects, 1)
	secondSubject := secondSubjects[0].(map[string]interface{})
	assert.Equal(t, "Group", secondSubject["type"])
	assert.Equal(t, "editors", secondSubject["name"])
	assert.Equal(t, "kube-system", secondSubject["namespace"])
}
