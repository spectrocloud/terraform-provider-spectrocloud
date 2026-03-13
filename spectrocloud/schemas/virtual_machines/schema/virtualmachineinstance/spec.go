package virtualmachineinstance

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	k8sv1 "k8s.io/api/core/v1"
	kubevirtapiv1 "kubevirt.io/api/core/v1"

	"github.com/spectrocloud/terraform-provider-spectrocloud/spectrocloud/kubevirt/schema/k8s"

	"github.com/spectrocloud/terraform-provider-spectrocloud/spectrocloud/kubevirt/utils"
)

func expandVirtualMachineInstanceSpec(d *schema.ResourceData) (kubevirtapiv1.VirtualMachineInstanceSpec, error) {
	result := kubevirtapiv1.VirtualMachineInstanceSpec{}

	if v, ok := d.GetOk("priority_class_name"); ok {
		result.PriorityClassName = v.(string)
	}

	if domain, err := ExpandDomainSpec(d); err == nil {
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
	att["domain"] = FlattenDomainSpec(in.Domain)

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
