package spectrocloud

import (
	"github.com/spectrocloud/gomi/pkg/ptr"
	"github.com/spectrocloud/palette-sdk-go/client"
	"reflect"
	"testing"

	"github.com/spectrocloud/hapi/models"
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
	d.Set("context", "invalid")
	err := updateClusterRBAC(nil, d)
	if err == nil || err.Error() != "invalid Context set - invalid" {
		t.Errorf("Expected 'invalid Context set - invalid', got %v", err)
	}
}

func TestRepaveApprovalCheck(t *testing.T) {

	d := resourceClusterAws().TestResourceData()
	d.Set("review_repave_state", "Approved")
	d.Set("context", "tenant")
	d.SetId("TestclusterUID")

	m := &client.V1Client{
		ApproveClusterRepaveFn: func(context, clusterUID string) error {
			return nil
		},
		GetClusterFn: func(context, clusterUID string) (*models.V1SpectroCluster, error) {
			return &models.V1SpectroCluster{
				APIVersion: "",
				Kind:       "",
				Metadata:   nil,
				Spec:       nil,
				Status: &models.V1SpectroClusterStatus{
					Repave: &models.V1ClusterRepaveStatus{
						State: "Approved",
					},
				},
			}, nil
		},
		GetRepaveReasonsFn: func(context, clusterUID string) ([]string, error) {
			var reason []string
			reason = append(reason, "PackValuesUpdated")
			return reason, nil
		},
	}

	// Test case where repave state is pending and approve_system_repave is true
	err := validateSystemRepaveApproval(d, m)
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}

	// Test case where repave state is pending and approve_system_repave is false
	m = &client.V1Client{
		ApproveClusterRepaveFn: func(context, clusterUID string) error {
			return nil
		},
		GetClusterFn: func(context, clusterUID string) (*models.V1SpectroCluster, error) {
			return &models.V1SpectroCluster{
				APIVersion: "",
				Kind:       "",
				Metadata:   nil,
				Spec:       nil,
				Status: &models.V1SpectroClusterStatus{
					Repave: &models.V1ClusterRepaveStatus{
						State: "Pending",
					},
				},
			}, nil
		},
		GetRepaveReasonsFn: func(context, clusterUID string) ([]string, error) {
			var reason []string
			reason = append(reason, "PackValuesUpdated")
			return reason, nil
		},
	}

	d.Set("review_repave_state", "")
	err = validateSystemRepaveApproval(d, m)
	expectedErrMsg := "cluster repave state is pending. \nDue to the following reasons -  \nPackValuesUpdated\nKindly verify the cluster and set `review_repave_state` to `Approved` to continue the repave operation and day 2 operation on the cluster."
	if err == nil || err.Error() != expectedErrMsg {
		t.Errorf("Expected error message '%s', got '%s'", expectedErrMsg, err)
	}

}

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
			Namespace:             "dns-label",
			ResourceVersion:       "test-resource-version-01",
			SelfLink:              "",
			UID:                   "test-cluster-uid",
		},
		Spec: &models.V1SpectroClusterSpec{
			CloudConfigRef: &models.V1ObjectReference{
				APIVersion:      "V1",
				FieldPath:       "",
				Kind:            "",
				Name:            "spc-cluster-unit-tes",
				Namespace:       "test-namespace",
				ResourceVersion: "test-cloud-config-resource-version-01",
				UID:             "test-cloud-config-uid",
			},
			CloudType: "vsphere",
			ClusterConfig: &models.V1ClusterConfig{
				ClusterMetaAttribute: "test-cluster-meta-attributes",
				ClusterRbac:          nil,
				ClusterResources: &models.V1ClusterResources{
					Namespaces: []*models.V1ResourceReference{
						&models.V1ResourceReference{
							Kind: "",
							Name: "",
							UID:  ptr.StringPtr("test-cluster-resource"),
						},
					},
					Rbacs: []*models.V1ResourceReference{
						&models.V1ResourceReference{
							Kind: "",
							Name: "",
							UID:  ptr.StringPtr("test-cluster-rbac-resource"),
						},
					},
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
						APIVersion:      "",
						FieldPath:       "",
						Kind:            "",
						Name:            "",
						Namespace:       "",
						ResourceVersion: "",
						UID:             "test-cluster-group-uid",
					},
					HostCluster: &models.V1ObjectReference{
						APIVersion:      "",
						FieldPath:       "",
						Kind:            "",
						Name:            "",
						Namespace:       "",
						ResourceVersion: "",
						UID:             "test-host-cluster-uid",
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

func getClient() *client.V1Client {
	c := &client.V1Client{
		GetCloudConfigVsphereFn: func(uid string) (*models.V1VsphereCloudConfig, error) {
			return getCloudConfig(), nil
		},
		GetClusterFn: func(scope, uid string) (*models.V1SpectroCluster, error) {
			isHost := new(bool)
			*isHost = true
			cluster := &models.V1SpectroCluster{
				APIVersion: "v1",
				Metadata: &models.V1ObjectMeta{
					Annotations:       nil,
					CreationTimestamp: models.V1Time{},
					DeletionTimestamp: models.V1Time{},
					Labels: map[string]string{
						"owner": "siva",
					},
					LastModifiedTimestamp: models.V1Time{},
					Name:                  "test-vsphere-cluster-unit-test",
					Namespace:             "",
					ResourceVersion:       "",
					SelfLink:              "",
					UID:                   "vsphere-uid",
				},
				Spec: &models.V1SpectroClusterSpec{
					CloudConfigRef: &models.V1ObjectReference{
						APIVersion:      "",
						FieldPath:       "",
						Kind:            "",
						Name:            "",
						Namespace:       "",
						ResourceVersion: "",
						UID:             "test-cloud-config-uid",
					},
					CloudType: "vsphere",
					ClusterConfig: &models.V1ClusterConfig{
						ClusterRbac:                    nil,
						ClusterResources:               nil,
						ControlPlaneHealthCheckTimeout: "",
						HostClusterConfig: &models.V1HostClusterConfig{
							ClusterEndpoint: &models.V1HostClusterEndpoint{
								Config: nil,
								Type:   "LoadBalancer",
							},
							ClusterGroup:  nil,
							HostCluster:   nil,
							IsHostCluster: isHost,
						},
						LifecycleConfig:             nil,
						MachineHealthConfig:         nil,
						MachineManagementConfig:     nil,
						UpdateWorkerPoolsInParallel: false,
					},
					ClusterProfileTemplates: nil,
					ClusterType:             "",
				},
				Status: &models.V1SpectroClusterStatus{
					State: "running",
					Repave: &models.V1ClusterRepaveStatus{
						State: "",
					},
				},
			}
			return cluster, nil
		},
		GetClusterBackupConfigFn: func(uid string) (*models.V1ClusterBackup, error) {
			clusterBackup := &models.V1ClusterBackup{
				Metadata: nil,
				Spec: &models.V1ClusterBackupSpec{
					ClusterUID: "vsphere-cluster-uid",
					Config: &models.V1ClusterBackupConfig{
						BackupLocationUID:       "test-back-uid",
						BackupName:              "unit-back",
						BackupPrefix:            "vsphere",
						DurationInHours:         3,
						IncludeAllDisks:         false,
						IncludeClusterResources: false,
						LocationType:            "",
						Namespaces:              nil,
						Schedule: &models.V1ClusterFeatureSchedule{
							ScheduledRunTime: "daily",
						},
					},
				},
				Status: nil,
			}
			return clusterBackup, nil
		},
		GetClusterKubeConfigFn: func(uid string) (string, error) {
			return "testKubeConfig", nil
		},
		GetClusterAdminConfigFn: func(uid string) (string, error) {
			return "testAdminKubeConfig", nil
		},
		GetClusterScanConfigFn: func(uid string) (*models.V1ClusterComplianceScan, error) {
			clusterCom := &models.V1ClusterComplianceScan{
				Metadata: &models.V1ObjectMeta{
					Annotations:           nil,
					CreationTimestamp:     models.V1Time{},
					DeletionTimestamp:     models.V1Time{},
					Labels:                nil,
					LastModifiedTimestamp: models.V1Time{},
					Name:                  "vsphere-cluster",
					Namespace:             "",
					ResourceVersion:       "",
					SelfLink:              "",
					UID:                   "conpli-uid",
				},
				Spec: &models.V1ClusterComplianceScanSpec{
					ClusterUID: "vsphere-cluster-uid",
					DriverSpec: map[string]models.V1ComplianceScanDriverSpec{
						"kube-bench": {
							Config: &models.V1ComplianceScanConfig{
								Schedule: &models.V1ClusterFeatureSchedule{
									ScheduledRunTime: "daily",
								},
							},
							IsClusterConfig: false,
						},
						"kube-hunter": {
							Config: &models.V1ComplianceScanConfig{
								Schedule: &models.V1ClusterFeatureSchedule{
									ScheduledRunTime: "daily",
								},
							},
							IsClusterConfig: false,
						},
						"sonobuoy": {
							Config: &models.V1ComplianceScanConfig{
								Schedule: &models.V1ClusterFeatureSchedule{
									ScheduledRunTime: "daily",
								},
							},
							IsClusterConfig: false,
						},
					},
				},
			}
			return clusterCom, nil
		},
		GetClusterRbacConfigFn: func(uid string) (*models.V1ClusterRbacs, error) {
			var rbacs []*models.V1ClusterRbac
			var subject []*models.V1ClusterRbacSubjects
			var bindings []*models.V1ClusterRbacBinding
			subject = append(subject, &models.V1ClusterRbacSubjects{
				Name:      "test-subject",
				Namespace: "vsphere-test",
				Type:      "test-subject",
			})
			bindings = append(bindings, &models.V1ClusterRbacBinding{
				Namespace: "vsphere-unittest",
				Role: &models.V1ClusterRoleRef{
					Kind: "scan",
					Name: "test-kind",
				},
				Subjects: subject,
				Type:     "test",
			})
			rbacs = append(rbacs, &models.V1ClusterRbac{
				Metadata: nil,
				Spec: &models.V1ClusterRbacSpec{
					Bindings:      bindings,
					RelatedObject: nil,
				},
				Status: nil,
			})
			clusterRbac := &models.V1ClusterRbacs{
				Items: rbacs,
			}
			return clusterRbac, nil
		},
		GetClusterNamespaceConfigFn: func(uid string) (*models.V1ClusterNamespaceResources, error) {
			var nResources []*models.V1ClusterNamespaceResource
			nResources = append(nResources, &models.V1ClusterNamespaceResource{
				Metadata: &models.V1ObjectMeta{
					Annotations:           nil,
					CreationTimestamp:     models.V1Time{},
					DeletionTimestamp:     models.V1Time{},
					Labels:                nil,
					LastModifiedTimestamp: models.V1Time{},
					Name:                  "test-namespace-unit",
					Namespace:             "",
					ResourceVersion:       "",
					SelfLink:              "",
					UID:                   "",
				},
				Spec: &models.V1ClusterNamespaceSpec{
					IsRegex:       false,
					RelatedObject: nil,
					ResourceAllocation: &models.V1ClusterNamespaceResourceAllocation{
						CPUCores:  5,
						MemoryMiB: 1234,
					},
				},
				Status: nil,
			})
			namespaceResource := &models.V1ClusterNamespaceResources{
				Items: nResources,
			}
			return namespaceResource, nil
		},
		GetClusterWithoutStatusFn: func(uid string) (*models.V1SpectroCluster, error) {
			cluster := &models.V1SpectroCluster{
				Metadata: nil,
			}
			cluster.Status = &models.V1SpectroClusterStatus{
				AbortTimestamp: models.V1Time{},
				AddOnServices:  nil,
				APIEndpoints:   nil,
				ClusterImport:  nil,
				Conditions:     nil,
				Location: &models.V1ClusterLocation{
					CountryCode: "IN",
					CountryName: "India",
					GeoLoc: &models.V1GeolocationLatlong{
						Latitude:  12.4241231,
						Longitude: 1932.12312,
					},
					RegionCode: "12",
					RegionName: "Asia",
				},
				Packs:         nil,
				ProfileStatus: nil,
				Services:      nil,
				SpcApply:      nil,
				State:         "running",
				Upgrades:      nil,
				Virtual:       nil,
			}
			return cluster, nil
		},
	}
	return c
}
func TestReadCommonFieldsCluster(t *testing.T) {
	d := prepareClusterVsphereTestData()
	spc := prepareSpectroClusterModel()
	c := getClient()
	_, done := readCommonFields(c, d, spc)
	assert.Equal(t, false, done)
}

func TestReadCommonFieldsVirtualCluster(t *testing.T) {
	d := resourceClusterVirtual().TestResourceData()
	spc := prepareSpectroClusterModel()
	c := getClient()
	_, done := readCommonFields(c, d, spc)
	assert.Equal(t, false, done)
}
