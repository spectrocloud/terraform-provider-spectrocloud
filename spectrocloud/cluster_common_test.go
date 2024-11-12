package spectrocloud

import (
	"fmt"
	"reflect"
	"sort"
	"testing"
	"time"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-sdk-go/client"
	"github.com/stretchr/testify/require"

	"github.com/spectrocloud/terraform-provider-spectrocloud/util/ptr"

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
					IsHostCluster: ptr.To((false),,
				},
				LifecycleConfig: &models.V1LifecycleConfig{
					Pause: ptr.To((false),,
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

func TestToNtpServers(t *testing.T) {
	data := map[string]interface{}{
		"ntp_servers": schema.NewSet(schema.HashString, []interface{}{"0.pool.ntp1.org"}),
	}

	servers := toNtpServers(data)

	expected := []string{"0.pool.ntp1.org"}
	assert.Equal(t, expected, servers)
}

func TestToClusterHostConfigs(t *testing.T) {
	d := schema.TestResourceDataRaw(t, map[string]*schema.Schema{
		"host_config": {
			Type:     schema.TypeList,
			Optional: true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"host_endpoint_type": {
						Type:     schema.TypeString,
						Optional: true,
					},
					"ingress_host": {
						Type:     schema.TypeString,
						Optional: true,
					},
					"external_traffic_policy": {
						Type:     schema.TypeString,
						Optional: true,
					},
					"load_balancer_source_ranges": {
						Type:     schema.TypeString,
						Optional: true,
					},
				},
			},
		},
	}, map[string]interface{}{
		"host_config": []interface{}{
			map[string]interface{}{
				"host_endpoint_type":          "LoadBalancer",
				"ingress_host":                "example.com",
				"external_traffic_policy":     "Cluster",
				"load_balancer_source_ranges": "10.0.0.0/24,192.168.1.0/24",
			},
		},
	})

	result := toClusterHostConfigs(d)

	expected := &models.V1HostClusterConfig{
		ClusterEndpoint: &models.V1HostClusterEndpoint{
			Type: "LoadBalancer",
			Config: &models.V1HostClusterEndpointConfig{
				IngressConfig: &models.V1IngressConfig{
					Host: "example.com",
				},
				LoadBalancerConfig: &models.V1LoadBalancerConfig{
					ExternalTrafficPolicy:    "Cluster",
					LoadBalancerSourceRanges: []string{"10.0.0.0/24", "192.168.1.0/24"},
				},
			},
		},
		IsHostCluster: ptr.To((true),,
	}

	assert.Equal(t, expected, result)
}

func TestToClusterHostConfigsNoHostConfig(t *testing.T) {
	d := schema.TestResourceDataRaw(t, map[string]*schema.Schema{
		"host_config": {
			Type:     schema.TypeList,
			Optional: true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"host_endpoint_type": {
						Type:     schema.TypeString,
						Optional: true,
					},
					"ingress_host": {
						Type:     schema.TypeString,
						Optional: true,
					},
					"external_traffic_policy": {
						Type:     schema.TypeString,
						Optional: true,
					},
					"load_balancer_source_ranges": {
						Type:     schema.TypeString,
						Optional: true,
					},
				},
			},
		},
	}, map[string]interface{}{})

	result := toClusterHostConfigs(d)

	expected := &models.V1HostClusterConfig{
		ClusterEndpoint: nil,
		IsHostCluster:   ptr.To((false),,
	}

	assert.Equal(t, expected, result)
}

func TestFlattenHostConfig(t *testing.T) {
	hostConfig := &models.V1HostClusterConfig{
		ClusterEndpoint: &models.V1HostClusterEndpoint{
			Type: "LoadBalancer",
			Config: &models.V1HostClusterEndpointConfig{
				IngressConfig: &models.V1IngressConfig{
					Host: "example.com",
				},
				LoadBalancerConfig: &models.V1LoadBalancerConfig{
					ExternalTrafficPolicy:    "Cluster",
					LoadBalancerSourceRanges: []string{"10.0.0.0/24", "192.168.1.0/24"},
				},
			},
		},
	}

	expected := []interface{}{
		map[string]interface{}{
			"host_endpoint_type":          "LoadBalancer",
			"ingress_host":                "example.com",
			"external_traffic_policy":     "Cluster",
			"load_balancer_source_ranges": "10.0.0.0/24,192.168.1.0/24",
		},
	}

	result := flattenHostConfig(hostConfig)

	assert.Equal(t, expected, result)
}

func TestFlattenHostConfigNil(t *testing.T) {
	hostConfig := &models.V1HostClusterConfig{}

	expected := []interface{}{}

	result := flattenHostConfig(hostConfig)

	assert.Equal(t, expected, result)
}

func TestFlattenSourceRanges(t *testing.T) {
	hostConfig := &models.V1HostClusterConfig{
		ClusterEndpoint: &models.V1HostClusterEndpoint{
			Config: &models.V1HostClusterEndpointConfig{
				LoadBalancerConfig: &models.V1LoadBalancerConfig{
					LoadBalancerSourceRanges: []string{"10.0.0.0/24", "192.168.1.0/24"},
				},
			},
		},
	}

	expected := "10.0.0.0/24,192.168.1.0/24"

	result := flattenSourceRanges(hostConfig)

	assert.Equal(t, expected, result)
}

func TestFlattenSourceRangesNil(t *testing.T) {
	hostConfig := &models.V1HostClusterConfig{
		ClusterEndpoint: &models.V1HostClusterEndpoint{
			Config: &models.V1HostClusterEndpointConfig{
				LoadBalancerConfig: &models.V1LoadBalancerConfig{
					LoadBalancerSourceRanges: []string{},
				},
			},
		},
	}

	expected := ""

	result := flattenSourceRanges(hostConfig)

	assert.Equal(t, expected, result)
}

func TestToClusterLocationConfigs(t *testing.T) {
	resourceData := schema.TestResourceDataRaw(t, map[string]*schema.Schema{
		"location_config": {
			Type:     schema.TypeList,
			Optional: true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"country_code": {
						Type:     schema.TypeString,
						Optional: true,
					},
					"country_name": {
						Type:     schema.TypeString,
						Optional: true,
					},
					"region_code": {
						Type:     schema.TypeString,
						Optional: true,
					},
					"region_name": {
						Type:     schema.TypeString,
						Optional: true,
					},
					"latitude": {
						Type:     schema.TypeFloat,
						Optional: true,
					},
					"longitude": {
						Type:     schema.TypeFloat,
						Optional: true,
					},
				},
			},
		},
	}, map[string]interface{}{
		"location_config": []interface{}{
			map[string]interface{}{
				"country_code": "US",
				"country_name": "United States",
				"region_code":  "CA",
				"region_name":  "California",
				"latitude":     37.7749,
				"longitude":    -122.4194,
			},
		},
	})

	expected := &models.V1ClusterLocation{
		CountryCode: "US",
		CountryName: "United States",
		RegionCode:  "CA",
		RegionName:  "California",
		GeoLoc: &models.V1GeolocationLatlong{
			Latitude:  37.7749,
			Longitude: -122.4194,
		},
	}

	result := toClusterLocationConfigs(resourceData)

	assert.Equal(t, expected, result)
}

func TestToClusterLocationConfig(t *testing.T) {
	config := map[string]interface{}{
		"country_code": "US",
		"country_name": "United States",
		"region_code":  "CA",
		"region_name":  "California",
		"latitude":     37.7749,
		"longitude":    -122.4194,
	}

	expected := &models.V1ClusterLocation{
		CountryCode: "US",
		CountryName: "United States",
		RegionCode:  "CA",
		RegionName:  "California",
		GeoLoc: &models.V1GeolocationLatlong{
			Latitude:  37.7749,
			Longitude: -122.4194,
		},
	}

	result := toClusterLocationConfig(config)

	assert.Equal(t, expected, result)
}

func TestToClusterGeoLoc(t *testing.T) {
	config := map[string]interface{}{
		"latitude":  37.7749,
		"longitude": -122.4194,
	}

	expected := &models.V1GeolocationLatlong{
		Latitude:  37.7749,
		Longitude: -122.4194,
	}

	result := toClusterGeoLoc(config)

	assert.Equal(t, expected, result)
}

func TestFlattenLocationConfig(t *testing.T) {
	location := &models.V1ClusterLocation{
		CountryCode: "US",
		CountryName: "United States",
		RegionCode:  "CA",
		RegionName:  "California",
		GeoLoc: &models.V1GeolocationLatlong{
			Latitude:  37.7749,
			Longitude: -122.4194,
		},
	}

	expected := []interface{}{
		map[string]interface{}{
			"country_code": "US",
			"country_name": "United States",
			"region_code":  "CA",
			"region_name":  "California",
			"latitude":     37.7749,
			"longitude":    -122.4194,
		},
	}

	result := flattenLocationConfig(location)

	assert.Equal(t, expected, result)
}

func TestFlattenLocationConfigNil(t *testing.T) {
	location := &models.V1ClusterLocation{}

	expected := []interface{}{
		map[string]interface{}{
			"country_code": "",
			"country_name": "",
			"region_code":  "",
			"region_name":  "",
		},
	}

	result := flattenLocationConfig(location)

	assert.Equal(t, expected, result)
}

func TestGetClusterMetadata(t *testing.T) {
	resourceData := resourceClusterAws().TestResourceData()
	_ = resourceData.Set("name", "test-cluster")
	_ = resourceData.Set("description", "A test cluster")
	_ = resourceData.Set("tags", []string{"env:prod", "team:devops"})

	resourceData.SetId("cluster-uid")

	expected := &models.V1ObjectMeta{
		Name:        "test-cluster",
		UID:         "cluster-uid",
		Labels:      map[string]string{"env": "prod", "team": "devops"},
		Annotations: map[string]string{"description": "A test cluster"},
	}

	result := getClusterMetadata(resourceData)

	assert.Equal(t, expected, result)
}

func TestToClusterMetadataUpdate(t *testing.T) {
	resourceData := resourceClusterAws().TestResourceData()
	_ = resourceData.Set("name", "test-cluster")
	_ = resourceData.Set("description", "A test cluster")
	_ = resourceData.Set("tags", []string{"env:prod", "team:devops"})

	expected := &models.V1ObjectMetaInputEntity{
		Name:        "test-cluster",
		Labels:      map[string]string{"env": "prod", "team": "devops"},
		Annotations: map[string]string{"description": "A test cluster"},
	}

	result := toClusterMetadataUpdate(resourceData)

	assert.Equal(t, expected, result)
}

func TestToUpdateClusterMetadata(t *testing.T) {
	resourceData := resourceClusterAws().TestResourceData()
	_ = resourceData.Set("name", "test-cluster")
	_ = resourceData.Set("description", "A test cluster")
	_ = resourceData.Set("tags", []string{"env:prod", "team:devops"})

	expected := &models.V1ObjectMetaInputEntitySchema{
		Metadata: &models.V1ObjectMetaInputEntity{
			Name:        "test-cluster",
			Labels:      map[string]string{"env": "prod", "team": "devops"},
			Annotations: map[string]string{"description": "A test cluster"},
		},
	}

	result := toUpdateClusterMetadata(resourceData)

	assert.Equal(t, expected, result)
}

func TestToUpdateClusterAdditionalMetadata(t *testing.T) {
	resourceData := resourceClusterAws().TestResourceData()
	_ = resourceData.Set("cluster_meta_attribute", "test-cluster-meta-attribute")
	expected := &models.V1ClusterMetaAttributeEntity{
		ClusterMetaAttribute: "test-cluster-meta-attribute",
	}
	result := toUpdateClusterAdditionalMetadata(resourceData)

	assert.Equal(t, expected, result)
}

func TestFlattenClusterNamespaces(t *testing.T) {
	namespaces := []*models.V1ClusterNamespaceResource{
		{
			Metadata: &models.V1ObjectMeta{
				Name: "namespace1",
			},
			Spec: &models.V1ClusterNamespaceSpec{
				ResourceAllocation: &models.V1ClusterNamespaceResourceAllocation{
					CPUCores:  2,
					MemoryMiB: 1024,
				},
			},
		},
		{
			Metadata: &models.V1ObjectMeta{
				Name: "namespace2",
			},
			Spec: &models.V1ClusterNamespaceSpec{
				ResourceAllocation: &models.V1ClusterNamespaceResourceAllocation{
					CPUCores:  4,
					MemoryMiB: 2048,
				},
			},
		},
	}

	expected := []interface{}{
		map[string]interface{}{
			"name": "namespace1",
			"resource_allocation": map[string]interface{}{
				"cpu_cores":  "2",
				"memory_MiB": "1024",
			},
		},
		map[string]interface{}{
			"name": "namespace2",
			"resource_allocation": map[string]interface{}{
				"cpu_cores":  "4",
				"memory_MiB": "2048",
			},
		},
	}

	result := flattenClusterNamespaces(namespaces)

	assert.Equal(t, expected, result)
}

func TestToClusterNamespace(t *testing.T) {
	clusterRbacBinding := map[string]interface{}{
		"name": "namespace1",
		"resource_allocation": map[string]interface{}{
			"cpu_cores":  "2",
			"memory_MiB": "1024",
		},
	}

	expected := &models.V1ClusterNamespaceResourceInputEntity{
		Metadata: &models.V1ObjectMetaUpdateEntity{
			Name: "namespace1",
		},
		Spec: &models.V1ClusterNamespaceSpec{
			IsRegex: false,
			ResourceAllocation: &models.V1ClusterNamespaceResourceAllocation{
				CPUCores:  2,
				MemoryMiB: 1024,
			},
		},
	}

	result := toClusterNamespace(clusterRbacBinding)

	assert.Equal(t, expected, result)
}

func TestToClusterNamespaces(t *testing.T) {
	resourceData := resourceClusterAws().TestResourceData()
	var ns []interface{}
	ns = append(ns, map[string]interface{}{
		"name": "namespace1",
		"resource_allocation": map[string]interface{}{
			"cpu_cores":  "2",
			"memory_MiB": "1024",
		},
	})
	ns = append(ns, map[string]interface{}{
		"name": "namespace2",
		"resource_allocation": map[string]interface{}{
			"cpu_cores":  "4",
			"memory_MiB": "2048",
		},
	})
	_ = resourceData.Set("namespaces", ns)
	expected := []*models.V1ClusterNamespaceResourceInputEntity{
		{
			Metadata: &models.V1ObjectMetaUpdateEntity{
				Name: "namespace1",
			},
			Spec: &models.V1ClusterNamespaceSpec{
				IsRegex: false,
				ResourceAllocation: &models.V1ClusterNamespaceResourceAllocation{
					CPUCores:  2,
					MemoryMiB: 1024,
				},
			},
		},
		{
			Metadata: &models.V1ObjectMetaUpdateEntity{
				Name: "namespace2",
			},
			Spec: &models.V1ClusterNamespaceSpec{
				IsRegex: false,
				ResourceAllocation: &models.V1ClusterNamespaceResourceAllocation{
					CPUCores:  4,
					MemoryMiB: 2048,
				},
			},
		},
	}

	result := toClusterNamespaces(resourceData)

	assert.Equal(t, expected, result)
}

func TestGetDefaultOsPatchConfig(t *testing.T) {
	expected := &models.V1MachineManagementConfig{
		OsPatchConfig: &models.V1OsPatchConfig{
			PatchOnBoot:      false,
			RebootIfRequired: false,
		},
	}

	result := getDefaultOsPatchConfig()

	assert.Equal(t, expected, result)
}

func TestToUpdateOsPatchEntityClusterRbac(t *testing.T) {
	config := &models.V1OsPatchConfig{
		PatchOnBoot:      true,
		RebootIfRequired: true,
	}

	expected := &models.V1OsPatchEntity{
		OsPatchConfig: config,
	}

	result := toUpdateOsPatchEntityClusterRbac(config)

	assert.Equal(t, expected, result)
}

func TestToOsPatchConfig(t *testing.T) {
	resourceData := schema.TestResourceDataRaw(t, map[string]*schema.Schema{
		"os_patch_on_boot": {
			Type:     schema.TypeBool,
			Optional: true,
		},
		"os_patch_schedule": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"os_patch_after": {
			Type:     schema.TypeString,
			Optional: true,
		},
	}, map[string]interface{}{
		"os_patch_on_boot":  true,
		"os_patch_schedule": "0 0 * * *",
		"os_patch_after":    "2024-01-01T00:00:00.000Z",
	})

	patchTime, _ := time.Parse(time.RFC3339, "2024-01-01T00:00:00.000Z")

	expected := &models.V1OsPatchConfig{
		PatchOnBoot:        true,
		Schedule:           "0 0 * * *",
		OnDemandPatchAfter: models.V1Time(patchTime),
	}

	result := toOsPatchConfig(resourceData)

	assert.Equal(t, expected, result)
}

func TestValidateOsPatchSchedule(t *testing.T) {
	validData := "0 0 * * *"
	invalidData := "invalid cron expression"

	validResult := validateOsPatchSchedule(validData, nil)
	invalidResult := validateOsPatchSchedule(invalidData, nil)

	assert.Empty(t, validResult)
	assert.NotEmpty(t, invalidResult)
	assert.Contains(t, invalidResult[0].Summary, "os patch schedule is invalid")
}

func TestValidateOsPatchOnDemandAfter(t *testing.T) {
	validData := time.Now().Add(20 * time.Minute).Format(time.RFC3339)
	invalidData := "invalid time format"
	pastData := time.Now().Add(-20 * time.Minute).Format(time.RFC3339)

	validResult := validateOsPatchOnDemandAfter(validData, nil)
	invalidResult := validateOsPatchOnDemandAfter(invalidData, nil)
	pastResult := validateOsPatchOnDemandAfter(pastData, nil)

	assert.Empty(t, validResult)
	assert.NotEmpty(t, invalidResult)
	assert.Contains(t, invalidResult[0].Summary, "time for 'os_patch_after' is invalid")

	assert.NotEmpty(t, pastResult)
	assert.Contains(t, pastResult[0].Summary, "valid timestamp is timestamp which is 10 mins ahead of current timestamp")
}

func TestToSpcApplySettings(t *testing.T) {
	// Test case when "apply_setting" is set
	resourceData := schema.TestResourceDataRaw(t, map[string]*schema.Schema{
		"apply_setting": {
			Type:     schema.TypeString,
			Optional: true,
		},
	}, map[string]interface{}{
		"apply_setting": "reboot",
	})

	expected := &models.V1SpcApplySettings{
		ActionType: "reboot",
	}

	result, err := toSpcApplySettings(resourceData)

	assert.Nil(t, err)
	assert.Equal(t, expected, result)

	// Test case when "apply_setting" is not set (empty string)
	resourceDataEmpty := schema.TestResourceDataRaw(t, map[string]*schema.Schema{
		"apply_setting": {
			Type:     schema.TypeString,
			Optional: true,
		},
	}, map[string]interface{}{
		"apply_setting": "",
	})

	resultEmpty, errEmpty := toSpcApplySettings(resourceDataEmpty)

	assert.Nil(t, errEmpty)
	assert.Nil(t, resultEmpty)

	// Test case when "apply_setting" is not present at all
	resourceDataNil := schema.TestResourceDataRaw(t, map[string]*schema.Schema{
		"apply_setting": {
			Type:     schema.TypeString,
			Optional: true,
		},
	}, map[string]interface{}{})

	resultNil, errNil := toSpcApplySettings(resourceDataNil)

	assert.Nil(t, errNil)
	assert.Nil(t, resultNil)
}

func TestGetNodeValue(t *testing.T) {
	// Test case 1: Standard input
	nodeId := "node1"
	action := "update"
	expected := map[string]interface{}{
		"node_id": nodeId,
		"action":  action,
	}
	result := getNodeValue(nodeId, action)
	assert.Equal(t, expected, result, "The returned map should match the expected map")

	// Test case 2: Different action
	nodeId = "node2"
	action = "reboot"
	expected = map[string]interface{}{
		"node_id": nodeId,
		"action":  action,
	}
	result = getNodeValue(nodeId, action)
	assert.Equal(t, expected, result, "The returned map should match the expected map")

	// Test case 3: Empty action
	nodeId = "node3"
	action = ""
	expected = map[string]interface{}{
		"node_id": nodeId,
		"action":  action,
	}
	result = getNodeValue(nodeId, action)
	assert.Equal(t, expected, result, "The returned map should match the expected map")

	// Test case 4: Empty nodeId
	nodeId = ""
	action = "update"
	expected = map[string]interface{}{
		"node_id": nodeId,
		"action":  action,
	}
	result = getNodeValue(nodeId, action)
	assert.Equal(t, expected, result, "The returned map should match the expected map")
}

func TestGetSpectroComponentsUpgrade(t *testing.T) {
	tests := []struct {
		name     string
		cluster  *models.V1SpectroCluster
		expected string
	}{
		{
			name: "Annotation is 'true'",
			cluster: &models.V1SpectroCluster{
				Metadata: &models.V1ObjectMeta{
					Annotations: map[string]string{
						"spectroComponentsUpgradeForbidden": "true",
					},
				},
			},
			expected: "lock",
		},
		{
			name: "Annotation is 'false'",
			cluster: &models.V1SpectroCluster{
				Metadata: &models.V1ObjectMeta{
					Annotations: map[string]string{
						"spectroComponentsUpgradeForbidden": "false",
					},
				},
			},
			expected: "unlock",
		},
		{
			name: "Annotation is not present",
			cluster: &models.V1SpectroCluster{
				Metadata: &models.V1ObjectMeta{
					Annotations: map[string]string{},
				},
			},
			expected: "unlock",
		},
		{
			name: "Annotations are nil",
			cluster: &models.V1SpectroCluster{
				Metadata: &models.V1ObjectMeta{
					Annotations: nil,
				},
			},
			expected: "unlock",
		},
		{
			name: "Different annotation key",
			cluster: &models.V1SpectroCluster{
				Metadata: &models.V1ObjectMeta{
					Annotations: map[string]string{
						"otherKey": "someValue",
					},
				},
			},
			expected: "unlock",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getSpectroComponentsUpgrade(tt.cluster)
			require.Equal(t, tt.expected, result)
		})
	}
}

func TestGetCommonCluster(t *testing.T) {
	tests := []struct {
		name          string
		resourceData  *schema.ResourceData
		expectedError string
		expectedID    string
		expectedName  string
		expectedCtx   string
	}{
		{
			name: "Successful cluster retrieval",

			resourceData: func() *schema.ResourceData {
				d := resourceClusterGcp().TestResourceData()
				d.SetId("cluster-id:project")
				return d
			}(),
			expectedError: "",
			expectedID:    "cluster-id",
			expectedName:  "cluster-name",
			expectedCtx:   "resource-context",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			_, err := GetCommonCluster(tt.resourceData, unitTestMockAPIClient)
			assert.NoError(t, err)

		})
	}
}

func TestValidateCloudType(t *testing.T) {
	tests := []struct {
		name          string
		resourceName  string
		cluster       *models.V1SpectroCluster
		expectedError string
	}{
		{
			name:         "Successful validation",
			resourceName: "spectrocloud_cluster_aws",
			cluster: &models.V1SpectroCluster{
				Metadata: &models.V1ObjectMeta{UID: "cluster-uid-123"},
				Spec:     &models.V1SpectroClusterSpec{CloudType: "aws"},
			},
			expectedError: "",
		},
		{
			name:         "Cluster spec is nil",
			resourceName: "spectrocloud_cluster_aws",
			cluster: &models.V1SpectroCluster{
				Metadata: &models.V1ObjectMeta{UID: "cluster-uid-123"},
				Spec:     nil,
			},
			expectedError: "cluster spec is nil in cluster cluster-uid-123",
		},
		{
			name:         "Cloud type mismatch",
			resourceName: "spectrocloud_cluster_aws",
			cluster: &models.V1SpectroCluster{
				Metadata: &models.V1ObjectMeta{UID: "cluster-uid-123"},
				Spec:     &models.V1SpectroClusterSpec{CloudType: "gcp"},
			},
			expectedError: "resource with id cluster-uid-123 is not of type spectrocloud_cluster_aws, need to correct resource type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateCloudType(tt.resourceName, tt.cluster)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUpdateAgentUpgradeSetting(t *testing.T) {
	resourceData := schema.TestResourceDataRaw(t, map[string]*schema.Schema{
		"pause_agent_upgrades": {
			Type:     schema.TypeString,
			Optional: true,
		},
	}, map[string]interface{}{})

	tests := []struct {
		name               string
		inputPauseUpgrades string
		mockError          error
		expectError        bool
	}{
		{
			name:               "Pause agent upgrades is set",
			inputPauseUpgrades: "true",
			mockError:          nil,
			expectError:        false,
		},
		{
			name:               "Pause agent upgrades is not set",
			inputPauseUpgrades: "",
			mockError:          nil,
			expectError:        false,
		},
		{
			name:               "Client returns an error",
			inputPauseUpgrades: "true",
			mockError:          assert.AnError,
			expectError:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resourceData.Set("pause_agent_upgrades", tt.inputPauseUpgrades)
			resourceData.SetId("test-cluster-id")

			err := updateAgentUpgradeSetting(getV1ClientWithResourceContext(unitTestMockAPIClient, "project"), resourceData)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

		})
	}
}

func TestValidateCloudTypeOne(t *testing.T) {
	tests := []struct {
		name           string
		input          interface{}
		expectedDiags  diag.Diagnostics
		expectedErrors bool
	}{
		{
			name:           "Valid cloud type: aws",
			input:          "aws",
			expectedDiags:  diag.Diagnostics{},
			expectedErrors: false,
		},
		{
			name:           "Valid cloud type: azure",
			input:          "azure",
			expectedDiags:  diag.Diagnostics{},
			expectedErrors: false,
		},
		{
			name:           "Invalid cloud type",
			input:          "invalid-cloud",
			expectedDiags:  diag.Diagnostics{diag.Diagnostic{Severity: diag.Error, Summary: fmt.Sprintf("cloud type '%s' is invalid. valid cloud types are %v", "invalid-cloud", "cloud_types")}},
			expectedErrors: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			diags := validateCloudType(tt.input, cty.Path{})

			if tt.expectedErrors {
				assert.Len(t, diags, 1)
				assert.Equal(t, tt.expectedDiags[0].Summary, diags[0].Summary)
			} else {
				assert.Empty(t, diags)
			}
		})
	}
}

func TestFlattenCloudConfigGeneric(t *testing.T) {
	resourceData := schema.TestResourceDataRaw(t, map[string]*schema.Schema{
		"cloud_config_id": {
			Type:     schema.TypeString,
			Optional: true,
		},
	}, map[string]interface{}{})

	client := &client.V1Client{}
	configUID := "test-config-uid"

	diags := flattenCloudConfigGeneric(configUID, resourceData, client)

	assert.Empty(t, diags)
	assert.Equal(t, configUID, resourceData.Get("cloud_config_id"))
}
