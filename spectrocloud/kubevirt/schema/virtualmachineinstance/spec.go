package virtualmachineinstance

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	k8sv1 "k8s.io/api/core/v1"
	kubevirtapiv1 "kubevirt.io/api/core/v1"

	"github.com/spectrocloud/terraform-provider-spectrocloud/spectrocloud/kubevirt/schema/k8s"

	"github.com/spectrocloud/terraform-provider-spectrocloud/spectrocloud/kubevirt/utils"
)

func virtualMachineInstanceSpecFields() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"priority_class_name": {
			Type:        schema.TypeString,
			Description: "If specified, indicates the pod's priority. If not specified, the pod priority will be default or zero if there is no default.",
			Optional:    true,
		},
		"domain": domainSpecSchema(),
		"node_selector": {
			Type:        schema.TypeMap,
			Description: "NodeSelector is a selector which must be true for the vmi to fit on a node. Selector which must match a node's labels for the vmi to be scheduled on that node.",
			Optional:    true,
			Elem:        &schema.Schema{Type: schema.TypeString},
		},
		"affinity": k8s.AffinitySchema(),
		"scheduler_name": {
			Type:        schema.TypeString,
			Description: "If specified, the VMI will be dispatched by specified scheduler. If not specified, the VMI will be dispatched by default scheduler.",
			Optional:    true,
		},
		"tolerations": k8s.TolerationSchema(),
		"eviction_strategy": {
			Type:        schema.TypeString,
			Description: "EvictionStrategy can be set to \"LiveMigrate\" if the VirtualMachineInstance should be migrated instead of shut-off in case of a node drain.",
			Optional:    true,
			ValidateFunc: validation.StringInSlice([]string{
				"LiveMigrate",
			}, false),
		},
		"termination_grace_period_seconds": {
			Type:        schema.TypeInt,
			Description: "Grace period observed after signalling a VirtualMachineInstance to stop after which the VirtualMachineInstance is force terminated.",
			Optional:    true,
		},
		"volume":          VolumesSchema(),
		"liveness_probe":  ProbeSchema(),
		"readiness_probe": ProbeSchema(),
		"hostname": {
			Type:        schema.TypeString,
			Description: "Specifies the hostname of the vmi.",
			Optional:    true,
		},
		"subdomain": {
			Type:        schema.TypeString,
			Description: "If specified, the fully qualified vmi hostname will be \"<hostname>.<subdomain>.<pod namespace>.svc.<cluster domain>\".",
			Optional:    true,
		},
		"network": NetworksSchema(),
		"dns_policy": {
			Type:        schema.TypeString,
			Description: "DNSPolicy defines how a pod's DNS will be configured.",
			Optional:    true,
			ValidateFunc: validation.StringInSlice([]string{
				"ClusterFirstWithHostNet",
				"ClusterFirst",
				"Default",
				"None",
			}, false),
		},
		"pod_dns_config": k8s.PodDnsConfigSchema(),
	}
}

func virtualMachineInstanceSpecSchema() *schema.Schema {
	fields := virtualMachineInstanceSpecFields()

	return &schema.Schema{
		Type: schema.TypeList,

		Description: "Template is the direct specification of VirtualMachineInstance.",
		Optional:    true,
		MaxItems:    1,
		Elem: &schema.Resource{
			Schema: fields,
		},
	}

}

func expandVirtualMachineInstanceSpec(d *schema.ResourceData) (kubevirtapiv1.VirtualMachineInstanceSpec, error) {
	result := kubevirtapiv1.VirtualMachineInstanceSpec{}

	if v, ok := d.GetOk("priority_class_name"); ok {
		result.PriorityClassName = v.(string)
	}

	if domain, err := expandDomainSpec(d); err == nil {
		result.Domain = domain
	} else {
		return result, err
	}
	if v, ok := d.GetOk("node_selector"); ok && len(v.(map[string]interface{})) > 0 {
		result.NodeSelector = utils.ExpandStringMap(v.(map[string]interface{}))
	}
	if v, ok := d.GetOk("affinity"); ok {
		result.Affinity = k8s.ExpandAffinity(v.([]interface{}))
	}
	if v, ok := d.GetOk("scheduler_name"); ok {
		result.SchedulerName = v.(string)
	}
	if v, ok := d.GetOk("tolerations"); ok {
		tolerations, err := k8s.ExpandTolerations(v.([]interface{}))
		if err != nil {
			return result, err
		}
		result.Tolerations = tolerations
	}
	if v, ok := d.GetOk("eviction_strategy"); ok {
		if v.(string) != "" {
			evictionStrategy := kubevirtapiv1.EvictionStrategy(v.(string))
			result.EvictionStrategy = &evictionStrategy
		}
	}
	if v, ok := d.GetOk("termination_grace_period_seconds"); ok {
		seconds := int64(v.(int))
		result.TerminationGracePeriodSeconds = &seconds
	}
	if v, ok := d.GetOk("volume"); ok {
		result.Volumes = expandVolumes(v.([]interface{}))
	}
	if v, ok := d.GetOk("liveness_probe"); ok {
		result.LivenessProbe = expandProbe(v.([]interface{}))
	}
	if v, ok := d.GetOk("readiness_probe"); ok {
		result.ReadinessProbe = expandProbe(v.([]interface{}))
	}
	if v, ok := d.GetOk("hostname"); ok {
		result.Hostname = v.(string)
	}
	if v, ok := d.GetOk("subdomain"); ok {
		result.Subdomain = v.(string)
	}
	if v, ok := d.GetOk("network"); ok {
		result.Networks = expandNetworks(v.([]interface{}))
	}
	if v, ok := d.GetOk("dns_policy"); ok {
		result.DNSPolicy = k8sv1.DNSPolicy(v.(string))
	}
	if v, ok := d.GetOk("pod_dns_config"); ok {
		dnsConfig, err := k8s.ExpandPodDNSConfig(v.([]interface{}))
		if err != nil {
			return result, err
		}
		result.DNSConfig = dnsConfig
	}

	return result, nil
}

func flattenVirtualMachineInstanceSpec(in kubevirtapiv1.VirtualMachineInstanceSpec, resourceData *schema.ResourceData) []interface{} {
	att := make(map[string]interface{})

	att["priority_class_name"] = in.PriorityClassName
	att["domain"] = flattenDomainSpec(in.Domain)

	att["node_selector"] = utils.FlattenStringMap(in.NodeSelector)
	att["affinity"] = k8s.FlattenAffinity(in.Affinity)
	att["scheduler_name"] = in.SchedulerName
	att["tolerations"] = k8s.FlattenTolerations(in.Tolerations)
	if in.EvictionStrategy != nil {
		att["eviction_strategy"] = string(*in.EvictionStrategy)
	}
	if in.TerminationGracePeriodSeconds != nil {
		att["termination_grace_period_seconds"] = *in.TerminationGracePeriodSeconds
	}
	att["volume"] = flattenVolumes(in.Volumes)
	if in.LivenessProbe != nil {
		att["liveness_probe"] = flattenProbe(*in.LivenessProbe)
	}
	if in.ReadinessProbe != nil {
		att["readiness_probe"] = flattenProbe(*in.ReadinessProbe)
	}
	att["hostname"] = in.Hostname
	att["subdomain"] = in.Subdomain
	att["network"] = flattenNetworks(in.Networks)
	att["dns_policy"] = string(in.DNSPolicy)
	if in.DNSConfig != nil {
		att["pod_dns_config"] = k8s.FlattenPodDNSConfig(in.DNSConfig)
	}

	return []interface{}{att}
}
