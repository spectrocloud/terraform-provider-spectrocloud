package virtualmachineinstance

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"k8s.io/apimachinery/pkg/api/resource"
	kubevirtapiv1 "kubevirt.io/api/core/v1"
)

func TestFlattenDomainSpec(t *testing.T) {
	testCases := []struct {
		input          kubevirtapiv1.DomainSpec
		expectedOutput []interface{}
	}{
		{
			input: kubevirtapiv1.DomainSpec{
				CPU: &kubevirtapiv1.CPU{
					Cores:   2,
					Sockets: 1,
					Threads: 1,
				},
			},
			expectedOutput: []interface{}{
				map[string]interface{}{
					"cpu": []interface{}{
						map[string]interface{}{
							"cores":   uint32(2),
							"sockets": uint32(1),
							"threads": uint32(1),
						},
					},
					"devices":   []interface{}{map[string]interface{}{"disk": []interface{}{}, "interface": []interface{}{}}},
					"resources": []interface{}{map[string]interface{}{"limits": map[string]interface{}{}, "over_commit_guest_overhead": false, "requests": map[string]interface{}{}}},
				},
			},
		},
		{
			input: kubevirtapiv1.DomainSpec{
				Memory: &kubevirtapiv1.Memory{
					Guest: resource.NewQuantity(2*1024*1024*1024, resource.BinarySI),
				},
			},
			expectedOutput: []interface{}{
				map[string]interface{}{
					"memory": []interface{}{
						map[string]interface{}{
							"guest": "2Gi",
						},
					},
					"devices":   []interface{}{map[string]interface{}{"disk": []interface{}{}, "interface": []interface{}{}}},
					"resources": []interface{}{map[string]interface{}{"limits": map[string]interface{}{}, "over_commit_guest_overhead": false, "requests": map[string]interface{}{}}},
				},
			},
		},
	}

	for _, tc := range testCases {
		output := flattenDomainSpec(tc.input)

		if diff := cmp.Diff(tc.expectedOutput, output, cmpopts.IgnoreUnexported(resource.Quantity{})); diff != "" {
			t.Errorf("Unexpected result (-want +got):\n%s", diff)
		}
	}
}
