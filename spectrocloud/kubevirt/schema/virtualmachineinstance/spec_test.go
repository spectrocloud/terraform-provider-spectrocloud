package virtualmachineinstance

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/spectrocloud/terraform-provider-spectrocloud/spectrocloud/kubevirt/schema/k8s"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// vmiSpecTestSchema is a top-level schema mirroring the keys
// expandVirtualMachineInstanceSpec reads via d.GetOk, layered on top of
// domainSpecTestSchema (resources/cpu/memory/firmware/features/disk/interface)
// and the real k8s/virtualmachineinstance sub-schemas for the rest.
func vmiSpecTestSchema() map[string]*schema.Schema {
	fields := domainSpecTestSchema().Schema
	fields["priority_class_name"] = &schema.Schema{Type: schema.TypeString, Optional: true}
	fields["node_selector"] = &schema.Schema{Type: schema.TypeMap, Optional: true, Elem: &schema.Schema{Type: schema.TypeString}}
	fields["affinity"] = k8s.AffinitySchema()
	fields["scheduler_name"] = &schema.Schema{Type: schema.TypeString, Optional: true}
	fields["tolerations"] = k8s.TolerationSchema()
	fields["eviction_strategy"] = &schema.Schema{Type: schema.TypeString, Optional: true}
	fields["termination_grace_period_seconds"] = &schema.Schema{Type: schema.TypeInt, Optional: true}
	fields["volume"] = VolumesSchema()
	fields["liveness_probe"] = ProbeSchema()
	fields["readiness_probe"] = ProbeSchema()
	fields["hostname"] = &schema.Schema{Type: schema.TypeString, Optional: true}
	fields["subdomain"] = &schema.Schema{Type: schema.TypeString, Optional: true}
	fields["network"] = NetworksSchema()
	fields["dns_policy"] = &schema.Schema{Type: schema.TypeString, Optional: true}
	fields["pod_dns_config"] = k8s.PodDnsConfigSchema()
	return fields
}

func TestExpandVirtualMachineInstanceSpec(t *testing.T) {
	t.Run("empty ResourceData succeeds with zero-value result", func(t *testing.T) {
		d := schema.TestResourceDataRaw(t, vmiSpecTestSchema(), map[string]interface{}{})
		got, err := expandVirtualMachineInstanceSpec(d)
		require.NoError(t, err)
		require.NotNil(t, got)
		require.NotNil(t, got.Domain)
	})

	t.Run("populated ResourceData", func(t *testing.T) {
		d := schema.TestResourceDataRaw(t, vmiSpecTestSchema(), map[string]interface{}{
			"priority_class_name": "high-priority",
			"node_selector": map[string]interface{}{
				"disktype": "ssd",
			},
			"affinity": []interface{}{
				map[string]interface{}{},
			},
			"scheduler_name": "custom-scheduler",
			"tolerations": []interface{}{
				map[string]interface{}{"key": "k1", "operator": "Exists"},
			},
			"eviction_strategy":                "LiveMigrate",
			"termination_grace_period_seconds": 30,
			"volume": []interface{}{
				map[string]interface{}{
					"name": "vol1",
					"volume_source": []interface{}{
						map[string]interface{}{
							"data_volume": []interface{}{
								map[string]interface{}{"name": "dv1"},
							},
						},
					},
				},
			},
			"liveness_probe":  []interface{}{map[string]interface{}{}},
			"readiness_probe": []interface{}{map[string]interface{}{}},
			"hostname":        "my-host",
			"subdomain":       "my-subdomain",
			"network": []interface{}{
				map[string]interface{}{"name": "net1"},
			},
			"dns_policy": "ClusterFirst",
			"pod_dns_config": []interface{}{
				map[string]interface{}{
					"nameservers": []interface{}{"8.8.8.8"},
				},
			},
		})

		got, err := expandVirtualMachineInstanceSpec(d)
		require.NoError(t, err)
		require.NotNil(t, got)

		assert.Equal(t, "high-priority", got.PriorityClassName)
		assert.Equal(t, map[string]string{"disktype": "ssd"}, got.NodeSelector)
		assert.NotNil(t, got.Affinity)
		assert.Equal(t, "custom-scheduler", got.SchedulerName)
		require.Len(t, got.Tolerations, 1)
		assert.Equal(t, "LiveMigrate", got.EvictionStrategy)
		assert.EqualValues(t, 30, got.TerminationGracePeriodSeconds)
		require.Len(t, got.Volumes, 1)
		// probeFields() has no schema fields defined yet (TODO in probe.go),
		// so the raw []interface{"{}"} value diffs down to a single nil
		// element and expandProbeToVM's nil-element guard returns nil; the
		// GetOk true-branch is still exercised.
		assert.Nil(t, got.LivenessProbe)
		assert.Nil(t, got.ReadinessProbe)
		assert.Equal(t, "my-host", got.Hostname)
		assert.Equal(t, "my-subdomain", got.Subdomain)
		require.Len(t, got.Networks, 1)
		assert.Equal(t, "ClusterFirst", got.DNSPolicy)
		assert.NotNil(t, got.DNSConfig)
	})
}

func TestFlattenVirtualMachineInstanceSpecFromVM(t *testing.T) {
	t.Run("nil input", func(t *testing.T) {
		got := flattenVirtualMachineInstanceSpecFromVM(nil, nil)
		require.Len(t, got, 1)
		m := got[0].(map[string]interface{})
		assert.Empty(t, m)
	})

	t.Run("populated input", func(t *testing.T) {
		name := "disk0"
		in := &models.V1VMVirtualMachineInstanceSpec{
			PriorityClassName: "high-priority",
			Domain: &models.V1VMDomainSpec{
				Devices: &models.V1VMDevices{Disks: []*models.V1VMDisk{{Name: &name}}},
			},
			NodeSelector:                  map[string]string{"disktype": "ssd"},
			SchedulerName:                 "custom-scheduler",
			EvictionStrategy:              "LiveMigrate",
			TerminationGracePeriodSeconds: 30,
			Hostname:                      "my-host",
			Subdomain:                     "my-subdomain",
			DNSPolicy:                     "ClusterFirst",
			DNSConfig:                     &models.V1VMPodDNSConfig{},
			LivenessProbe:                 &models.V1VMProbe{},
			ReadinessProbe:                &models.V1VMProbe{},
		}

		got := flattenVirtualMachineInstanceSpecFromVM(in, nil)
		require.Len(t, got, 1)
		m := got[0].(map[string]interface{})
		assert.Equal(t, "high-priority", m["priority_class_name"])
		assert.Contains(t, m, "domain")
		assert.Equal(t, "custom-scheduler", m["scheduler_name"])
		assert.Equal(t, "LiveMigrate", m["eviction_strategy"])
		assert.EqualValues(t, 30, m["termination_grace_period_seconds"])
		assert.Equal(t, "my-host", m["hostname"])
		assert.Equal(t, "my-subdomain", m["subdomain"])
		assert.Equal(t, "ClusterFirst", m["dns_policy"])
		assert.Contains(t, m, "pod_dns_config")
		assert.Contains(t, m, "liveness_probe")
		assert.Contains(t, m, "readiness_probe")
	})
}
