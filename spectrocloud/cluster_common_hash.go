package spectrocloud

import (
	"bytes"
	"fmt"
	"hash/fnv"
	"sort"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func CommonHash(nodePool map[string]interface{}) *bytes.Buffer {
	var buf bytes.Buffer

	if _, ok := nodePool["additional_labels"]; ok {
		buf.WriteString(HashStringMap(nodePool["additional_labels"]))
	}
	if _, ok := nodePool["taints"]; ok {
		buf.WriteString(HashStringMapList(nodePool["taints"]))
	}
	if val, ok := nodePool["control_plane"]; ok {
		buf.WriteString(fmt.Sprintf("%t-", val.(bool)))
	}
	if val, ok := nodePool["control_plane_as_worker"]; ok {
		buf.WriteString(fmt.Sprintf("%t-", val.(bool)))
	}
	if val, ok := nodePool["name"]; ok {
		buf.WriteString(fmt.Sprintf("%s-", val.(string)))
	}
	if val, ok := nodePool["count"]; ok {
		buf.WriteString(fmt.Sprintf("%d-", val.(int)))
	}
	if val, ok := nodePool["update_strategy"]; ok {
		buf.WriteString(fmt.Sprintf("%s-", val.(string)))
	}
	if val, ok := nodePool["node_repave_interval"]; ok {
		buf.WriteString(fmt.Sprintf("%d-", val.(int)))
	}
	/*if val, ok := nodePool["instance_type"]; ok {
		buf.WriteString(fmt.Sprintf("%s-", val.(string)))
	}
	if val, ok := nodePool["azs"]; ok {
		buf.WriteString(fmt.Sprintf("%s-", val.(string)))
	}*/
	if val, ok := nodePool["min"]; ok {
		buf.WriteString(fmt.Sprintf("%d-", val.(int)))
	}
	if val, ok := nodePool["max"]; ok {
		buf.WriteString(fmt.Sprintf("%d-", val.(int)))
	}

	return &buf
}

func resourceMachinePoolAzureHash(v interface{}) int {
	m := v.(map[string]interface{})
	buf := CommonHash(m)

	if val, ok := m["instance_type"]; ok {
		buf.WriteString(fmt.Sprintf("%s-", val.(string)))
	}
	if val, ok := m["is_system_node_pool"]; ok {
		buf.WriteString(fmt.Sprintf("%t-", val.(bool)))
	}
	if val, ok := m["os_type"]; ok && val != "" {
		buf.WriteString(fmt.Sprintf("%s-", val.(string)))
	}

	return int(hash(buf.String()))
}

func resourceMachinePoolAksHash(v interface{}) int {
	m := v.(map[string]interface{})
	buf := CommonHash(m)

	if val, ok := m["instance_type"]; ok {
		buf.WriteString(fmt.Sprintf("%s-", val.(string)))
	}
	if val, ok := m["disk_size_gb"]; ok {
		buf.WriteString(fmt.Sprintf("%d-", val.(int)))
	}
	if val, ok := m["is_system_node_pool"]; ok {
		buf.WriteString(fmt.Sprintf("%t-", val.(bool)))
	}
	if val, ok := m["storage_account_type"]; ok {
		buf.WriteString(fmt.Sprintf("%s-", val.(string)))
	}

	return int(hash(buf.String()))
}

func resourceMachinePoolGcpHash(v interface{}) int {
	m := v.(map[string]interface{})
	buf := CommonHash(m)

	return int(hash(buf.String()))
}

func resourceMachinePoolAwsHash(v interface{}) int {
	m := v.(map[string]interface{})
	buf := CommonHash(m)

	if m["min"] != nil {
		buf.WriteString(fmt.Sprintf("%d-", m["min"].(int)))
	}
	if m["max"] != nil {
		buf.WriteString(fmt.Sprintf("%d-", m["max"].(int)))
	}
	buf.WriteString(fmt.Sprintf("%s-", m["instance_type"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["capacity_type"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["max_price"].(string)))
	if m["azs"] != nil {
		azsSet := m["azs"].(*schema.Set)
		azsList := azsSet.List()
		azsListStr := make([]string, len(azsList))
		for i, v := range azsList {
			azsListStr[i] = v.(string)
		}
		sort.Strings(azsListStr)
		azsStr := strings.Join(azsListStr, "-")
		buf.WriteString(fmt.Sprintf("%s-", azsStr))
	}
	buf.WriteString(fmt.Sprintf("%s-", m["azs"].(*schema.Set).GoString()))
	buf.WriteString(HashStringMap(m["az_subnets"]))

	return int(hash(buf.String()))
}

func resourceMachinePoolEksHash(v interface{}) int {
	m := v.(map[string]interface{})
	buf := CommonHash(m)

	buf.WriteString(fmt.Sprintf("%d-", m["disk_size_gb"].(int)))
	if m["min"] != nil {
		buf.WriteString(fmt.Sprintf("%d-", m["min"].(int)))
	}
	if m["max"] != nil {
		buf.WriteString(fmt.Sprintf("%d-", m["max"].(int)))
	}
	buf.WriteString(fmt.Sprintf("%s-", m["instance_type"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["capacity_type"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["max_price"].(string)))

	for i, j := range m["az_subnets"].(map[string]interface{}) {
		buf.WriteString(fmt.Sprintf("%s-%s", i, j.(string)))
	}

	if m["eks_launch_template"] != nil {
		buf.WriteString(eksLaunchTemplate(m["eks_launch_template"]))
	}

	return int(hash(buf.String()))
}

func eksLaunchTemplate(v interface{}) string {
	var buf bytes.Buffer
	if len(v.([]interface{})) > 0 {
		m := v.([]interface{})[0].(map[string]interface{})

		if m["ami_id"] != nil {
			buf.WriteString(fmt.Sprintf("%s-", m["ami_id"].(string)))
		}
		if m["root_volume_type"] != nil {
			buf.WriteString(fmt.Sprintf("%s-", m["root_volume_type"].(string)))
		}
		if m["root_volume_iops"] != nil {
			buf.WriteString(fmt.Sprintf("%d-", m["root_volume_iops"].(int)))
		}
		if m["root_volume_throughput"] != nil {
			buf.WriteString(fmt.Sprintf("%d-", m["root_volume_throughput"].(int)))
		}
		if m["additional_security_groups"] != nil {
			for _, sg := range m["additional_security_groups"].(*schema.Set).List() {
				buf.WriteString(fmt.Sprintf("%s-", sg.(string)))
			}
		}
	}

	return buf.String()
}

func resourceMachinePoolCoxEdgeHash(v interface{}) int {
	m := v.(map[string]interface{})
	buf := CommonHash(m)

	return int(hash(buf.String()))
}

func resourceMachinePoolTkeHash(v interface{}) int {
	m := v.(map[string]interface{})
	buf := CommonHash(m)

	for i, j := range m["az_subnets"].(map[string]interface{}) {
		buf.WriteString(fmt.Sprintf("%s-%s", i, j.(string)))
	}

	return int(hash(buf.String()))
}

func resourceMachinePoolVsphereHash(v interface{}) int {
	m := v.(map[string]interface{})
	buf := CommonHash(m)

	if v, found := m["instance_type"]; found {
		if len(v.([]interface{})) > 0 {
			ins := v.([]interface{})[0].(map[string]interface{})
			buf.WriteString(fmt.Sprintf("%d-", ins["cpu"].(int)))
			buf.WriteString(fmt.Sprintf("%d-", ins["disk_size_gb"].(int)))
			buf.WriteString(fmt.Sprintf("%d-", ins["memory_mb"].(int)))
		}
	}

	if placements, found := m["placement"]; found {
		for _, p := range placements.([]interface{}) {
			place := p.(map[string]interface{})
			buf.WriteString(fmt.Sprintf("%s-", place["cluster"].(string)))
			buf.WriteString(fmt.Sprintf("%s-", place["resource_pool"].(string)))
			buf.WriteString(fmt.Sprintf("%s-", place["datastore"].(string)))
			buf.WriteString(fmt.Sprintf("%s-", place["network"].(string)))
			buf.WriteString(fmt.Sprintf("%s-", place["static_ip_pool_id"].(string)))
		}
	}
	return int(hash(buf.String()))
}

func resourceMachinePoolOpenStackHash(v interface{}) int {
	m := v.(map[string]interface{})
	buf := CommonHash(m)

	buf.WriteString(fmt.Sprintf("%s-", m["instance_type"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["subnet_id"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["update_strategy"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["azs"].(*schema.Set).GoString()))

	return int(hash(buf.String()))
}

func resourceMachinePoolVirtualHash(v interface{}) int {
	m := v.(map[string]interface{})
	buf := CommonHash(m)

	return int(hash(buf.String()))
}

func resourceMachinePoolMaasHash(v interface{}) int {
	m := v.(map[string]interface{})
	buf := CommonHash(m)

	if v, found := m["instance_type"]; found {
		if len(v.([]interface{})) > 0 {
			ins := v.([]interface{})[0].(map[string]interface{})
			buf.WriteString(fmt.Sprintf("%d-", ins["min_cpu"].(int)))
			buf.WriteString(fmt.Sprintf("%d-", ins["min_memory_mb"].(int)))
		}
	}
	buf.WriteString(fmt.Sprintf("%s-", m["azs"].(*schema.Set).GoString()))

	return int(hash(buf.String()))
}

func resourceMachinePoolLibvirtHash(v interface{}) int {
	m := v.(map[string]interface{})
	buf := CommonHash(m)

	if v, found := m["xsl_template"]; found {
		buf.WriteString(fmt.Sprintf("%s-", v.(string)))
	}

	if v, found := m["instance_type"]; found {
		if len(v.([]interface{})) > 0 {
			ins := v.([]interface{})[0].(map[string]interface{})
			buf.WriteString(InstanceTypeHash(ins))
		}
	}

	return int(hash(buf.String()))
}

func InstanceTypeHash(ins map[string]interface{}) string {
	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf("%d-", ins["cpu"].(int)))
	buf.WriteString(fmt.Sprintf("%d-", ins["disk_size_gb"].(int)))
	buf.WriteString(fmt.Sprintf("%d-", ins["memory_mb"].(int)))
	buf.WriteString(fmt.Sprintf("%s-", ins["cpus_sets"].(string)))
	if ins["cache_passthrough"] != nil {
		buf.WriteString(fmt.Sprintf("%s-%t", "cache_passthrough", ins["cache_passthrough"].(bool)))
	}
	if ins["gpu_config"] != nil {
		config, _ := ins["gpu_config"].(map[string]interface{})
		if config != nil {
			buf.WriteString(GpuConfigHash(config))
		}
	}

	if ins["attached_disks"] != nil {
		for _, disk := range ins["attached_disks"].([]interface{}) {
			diskMap := disk.(map[string]interface{})
			if diskMap["managed"] != nil {
				buf.WriteString(fmt.Sprintf("%s-%t", "managed", diskMap["managed"].(bool)))
			}
			if diskMap["size_in_gb"] != nil {
				buf.WriteString(fmt.Sprintf("%s-%d", "size_in_gb", diskMap["size_in_gb"].(int)))
			}
		}
	}
	return buf.String()
}

func GpuConfigHash(config map[string]interface{}) string {
	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf("%d-", config["num_gpus"].(int)))
	buf.WriteString(fmt.Sprintf("%s-", config["device_model"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", config["vendor"].(string)))
	buf.WriteString(HashStringMap(config["addresses"]))
	return buf.String()
}

func resourceMachinePoolEdgeNativeHash(v interface{}) int {
	m := v.(map[string]interface{})
	buf := CommonHash(m)

	if _, found := m["host_uids"]; found {
		for _, host := range m["host_uids"].([]interface{}) {
			buf.WriteString(fmt.Sprintf("%s-", host.(string)))
		}
	}

	return int(hash(buf.String()))
}

/*func resourceMachinePoolAzureHash(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})

	buf.WriteString(HashStringMap(m["additional_labels"]))
	buf.WriteString(HashStringMapList(m["taints"]))

	buf.WriteString(fmt.Sprintf("%t-", m["control_plane"].(bool)))
	buf.WriteString(fmt.Sprintf("%t-", m["control_plane_as_worker"].(bool)))
	buf.WriteString(fmt.Sprintf("%s-", m["name"].(string)))
	buf.WriteString(fmt.Sprintf("%d-", m["count"].(int)))
	buf.WriteString(fmt.Sprintf("%s-", m["update_strategy"].(string)))

	buf.WriteString(fmt.Sprintf("%s-", m["instance_type"].(string)))
	buf.WriteString(fmt.Sprintf("%t-", m["is_system_node_pool"].(bool)))

	buf.WriteString(fmt.Sprintf("%s-", m["azs"].(*schema.Set).GoString()))
	if m["node_repave_interval"] != nil {
		buf.WriteString(fmt.Sprintf("%d-", m["node_repave_interval"].(int)))
	}
	if m["os_type"] != "" {
		buf.WriteString(fmt.Sprintf("%s-", m["os_type"].(string)))
	}

	// TODO(saamalik) fix for disk
	//buf.WriteString(fmt.Sprintf("%d-", d["size_gb"].(int)))
	//buf.WriteString(fmt.Sprintf("%s-", d["type"].(string)))

	//d2 := m["disk"].([]interface{})
	//d := d2[0].(map[string]interface{})

	return int(hash(buf.String()))
}

func resourceMachinePoolAksHash(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})

	buf.WriteString(HashStringMap(m["additional_labels"]))
	buf.WriteString(HashStringMapList(m["taints"]))

	buf.WriteString(fmt.Sprintf("%s-", m["name"].(string)))
	buf.WriteString(fmt.Sprintf("%d-", m["count"].(int)))
	buf.WriteString(fmt.Sprintf("%s-", m["update_strategy"].(string)))

	if m["min"] != nil {
		buf.WriteString(fmt.Sprintf("%d-", m["min"].(int)))
	}
	if m["max"] != nil {
		buf.WriteString(fmt.Sprintf("%d-", m["max"].(int)))
	}

	buf.WriteString(fmt.Sprintf("%s-", m["instance_type"].(string)))
	buf.WriteString(fmt.Sprintf("%d-", m["disk_size_gb"].(int)))
	buf.WriteString(fmt.Sprintf("%t-", m["is_system_node_pool"].(bool)))
	buf.WriteString(fmt.Sprintf("%s-", m["storage_account_type"].(string)))
	return int(hash(buf.String()))
}

func resourceMachinePoolGcpHash(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})

	buf.WriteString(HashStringMap(m["additional_labels"]))
	buf.WriteString(HashStringMapList(m["taints"]))

	buf.WriteString(fmt.Sprintf("%t-", m["control_plane"].(bool)))
	buf.WriteString(fmt.Sprintf("%t-", m["control_plane_as_worker"].(bool)))
	buf.WriteString(fmt.Sprintf("%s-", m["name"].(string)))
	buf.WriteString(fmt.Sprintf("%d-", m["count"].(int)))
	buf.WriteString(fmt.Sprintf("%s-", m["update_strategy"].(string)))

	buf.WriteString(fmt.Sprintf("%s-", m["instance_type"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["azs"].(*schema.Set).GoString()))
	if m["node_repave_interval"] != nil {
		buf.WriteString(fmt.Sprintf("%d-", m["node_repave_interval"].(int)))
	}

	return int(hash(buf.String()))
}

func resourceMachinePoolAwsHash(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})

	buf.WriteString(HashStringMap(m["additional_labels"]))
	buf.WriteString(HashStringMapList(m["taints"]))

	buf.WriteString(fmt.Sprintf("%t-", m["control_plane"].(bool)))
	buf.WriteString(fmt.Sprintf("%t-", m["control_plane_as_worker"].(bool)))
	buf.WriteString(fmt.Sprintf("%s-", m["name"].(string)))
	buf.WriteString(fmt.Sprintf("%d-", m["count"].(int)))

	if m["min"] != nil {
		buf.WriteString(fmt.Sprintf("%d-", m["min"].(int)))
	}
	if m["max"] != nil {
		buf.WriteString(fmt.Sprintf("%d-", m["max"].(int)))
	}
	if m["node_repave_interval"] != nil {
		buf.WriteString(fmt.Sprintf("%d-", m["node_repave_interval"].(int)))
	}
	buf.WriteString(fmt.Sprintf("%s-", m["update_strategy"].(string)))

	buf.WriteString(fmt.Sprintf("%s-", m["instance_type"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["capacity_type"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["max_price"].(string)))
	if m["azs"] != nil {
		azsSet := m["azs"].(*schema.Set)
		azsList := azsSet.List()
		azsListStr := make([]string, len(azsList))
		for i, v := range azsList {
			azsListStr[i] = v.(string)
		}
		sort.Strings(azsListStr)
		azsStr := strings.Join(azsListStr, "-")
		buf.WriteString(fmt.Sprintf("%s-", azsStr))
	}
	buf.WriteString(fmt.Sprintf("%s-", m["azs"].(*schema.Set).GoString()))
	buf.WriteString(HashStringMap(m["az_subnets"]))

	return int(hash(buf.String()))
}

func resourceMachinePoolEksHash(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})

	buf.WriteString(HashStringMap(m["additional_labels"]))
	buf.WriteString(HashStringMapList(m["taints"]))

	buf.WriteString(fmt.Sprintf("%s-", m["name"].(string)))
	buf.WriteString(fmt.Sprintf("%d-", m["disk_size_gb"].(int)))
	buf.WriteString(fmt.Sprintf("%d-", m["count"].(int)))
	buf.WriteString(fmt.Sprintf("%s-", m["update_strategy"].(string)))

	if m["min"] != nil {
		buf.WriteString(fmt.Sprintf("%d-", m["min"].(int)))
	}
	if m["max"] != nil {
		buf.WriteString(fmt.Sprintf("%d-", m["max"].(int)))
	}
	buf.WriteString(fmt.Sprintf("%s-", m["instance_type"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["capacity_type"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["max_price"].(string)))

	for i, j := range m["az_subnets"].(map[string]interface{}) {
		buf.WriteString(fmt.Sprintf("%s-%s", i, j.(string)))
	}

	if m["eks_launch_template"] != nil {
		buf.WriteString(eksLaunchTemplate(m["eks_launch_template"]))
	}

	return int(hash(buf.String()))
}

func eksLaunchTemplate(v interface{}) string {
	var buf bytes.Buffer
	if len(v.([]interface{})) > 0 {
		m := v.([]interface{})[0].(map[string]interface{})

		if m["ami_id"] != nil {
			buf.WriteString(fmt.Sprintf("%s-", m["ami_id"].(string)))
		}
		if m["root_volume_type"] != nil {
			buf.WriteString(fmt.Sprintf("%s-", m["root_volume_type"].(string)))
		}
		if m["root_volume_iops"] != nil {
			buf.WriteString(fmt.Sprintf("%d-", m["root_volume_iops"].(int)))
		}
		if m["root_volume_throughput"] != nil {
			buf.WriteString(fmt.Sprintf("%d-", m["root_volume_throughput"].(int)))
		}
		if m["additional_security_groups"] != nil {
			for _, sg := range m["additional_security_groups"].(*schema.Set).List() {
				buf.WriteString(fmt.Sprintf("%s-", sg.(string)))
			}
		}
	}

	return buf.String()
}

func resourceMachinePoolCoxEdgeHash(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})

	buf.WriteString(HashStringMap(m["additional_labels"]))
	buf.WriteString(HashStringMapList(m["taints"]))

	buf.WriteString(fmt.Sprintf("%s-", m["name"].(string)))
	buf.WriteString(fmt.Sprintf("%d-", m["count"].(int)))

	if m["minSize"] != nil {
		buf.WriteString(fmt.Sprintf("%d-", m["min"].(int)))
	}
	if m["maxSize"] != nil {
		buf.WriteString(fmt.Sprintf("%d-", m["max"].(int)))
	}
	buf.WriteString(fmt.Sprintf("%t-", m["control_plane_as_worker"].(bool)))
	buf.WriteString(fmt.Sprintf("%t-", m["control_plane"].(bool)))

	return int(hash(buf.String()))
}

func resourceMachinePoolTkeHash(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})

	buf.WriteString(HashStringMap(m["additional_labels"]))
	buf.WriteString(HashStringMapList(m["taints"]))

	buf.WriteString(fmt.Sprintf("%s-", m["name"].(string)))
	buf.WriteString(fmt.Sprintf("%d-", m["disk_size_gb"].(int)))
	buf.WriteString(fmt.Sprintf("%d-", m["count"].(int)))
	buf.WriteString(fmt.Sprintf("%s-", m["update_strategy"].(string)))

	if m["min"] != nil {
		buf.WriteString(fmt.Sprintf("%d-", m["min"].(int)))
	}
	if m["max"] != nil {
		buf.WriteString(fmt.Sprintf("%d-", m["max"].(int)))
	}
	buf.WriteString(fmt.Sprintf("%s-", m["instance_type"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["capacity_type"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["max_price"].(string)))

	for i, j := range m["az_subnets"].(map[string]interface{}) {
		buf.WriteString(fmt.Sprintf("%s-%s", i, j.(string)))
	}

	return int(hash(buf.String()))
}

func resourceMachinePoolVsphereHash(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})

	buf.WriteString(HashStringMap(m["additional_labels"]))
	buf.WriteString(HashStringMapList(m["taints"]))

	//d := m["disk"].([]interface{})[0].(map[string]interface{})
	buf.WriteString(fmt.Sprintf("%t-", m["control_plane"].(bool)))
	buf.WriteString(fmt.Sprintf("%t-", m["control_plane_as_worker"].(bool)))
	buf.WriteString(fmt.Sprintf("%s-", m["name"].(string)))
	buf.WriteString(fmt.Sprintf("%d-", m["count"].(int)))
	if m["node_repave_interval"] != nil {
		buf.WriteString(fmt.Sprintf("%d-", m["node_repave_interval"].(int)))
	}
	buf.WriteString(fmt.Sprintf("%s-", m["update_strategy"].(string)))

	if v, found := m["instance_type"]; found {
		if len(v.([]interface{})) > 0 {
			ins := v.([]interface{})[0].(map[string]interface{})
			buf.WriteString(fmt.Sprintf("%d-", ins["cpu"].(int)))
			buf.WriteString(fmt.Sprintf("%d-", ins["disk_size_gb"].(int)))
			buf.WriteString(fmt.Sprintf("%d-", ins["memory_mb"].(int)))
		}
	}

	if placements, found := m["placement"]; found {
		for _, p := range placements.([]interface{}) {
			place := p.(map[string]interface{})
			buf.WriteString(fmt.Sprintf("%s-", place["cluster"].(string)))
			buf.WriteString(fmt.Sprintf("%s-", place["resource_pool"].(string)))
			buf.WriteString(fmt.Sprintf("%s-", place["datastore"].(string)))
			buf.WriteString(fmt.Sprintf("%s-", place["network"].(string)))
			buf.WriteString(fmt.Sprintf("%s-", place["static_ip_pool_id"].(string)))
		}
	}
	return int(hash(buf.String()))
}

func resourceMachinePoolOpenStackHash(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})

	buf.WriteString(HashStringMap(m["additional_labels"]))
	buf.WriteString(HashStringMapList(m["taints"]))

	buf.WriteString(fmt.Sprintf("%t-", m["control_plane"].(bool)))
	buf.WriteString(fmt.Sprintf("%t-", m["control_plane_as_worker"].(bool)))
	buf.WriteString(fmt.Sprintf("%s-", m["name"].(string)))
	buf.WriteString(fmt.Sprintf("%d-", m["count"].(int)))
	buf.WriteString(fmt.Sprintf("%s-", m["update_strategy"].(string)))

	buf.WriteString(fmt.Sprintf("%s-", m["instance_type"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["subnet_id"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["update_strategy"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["azs"].(*schema.Set).GoString()))

	return int(hash(buf.String()))
}

func resourceMachinePoolVirtualHash(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})

	buf.WriteString(HashStringMap(m["additional_labels"]))
	buf.WriteString(HashStringMapList(m["taints"]))

	buf.WriteString(fmt.Sprintf("%t-", m["control_plane"].(bool)))
	buf.WriteString(fmt.Sprintf("%t-", m["control_plane_as_worker"].(bool)))
	buf.WriteString(fmt.Sprintf("%s-", m["name"].(string)))
	buf.WriteString(fmt.Sprintf("%d-", m["count"].(int)))
	buf.WriteString(fmt.Sprintf("%s-", m["update_strategy"].(string)))

	return int(hash(buf.String()))
}

func resourceMachinePoolMaasHash(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})

	buf.WriteString(HashStringMap(m["additional_labels"]))
	buf.WriteString(HashStringMapList(m["taints"]))

	buf.WriteString(fmt.Sprintf("%t-", m["control_plane"].(bool)))
	buf.WriteString(fmt.Sprintf("%t-", m["control_plane_as_worker"].(bool)))
	buf.WriteString(fmt.Sprintf("%s-", m["name"].(string)))
	buf.WriteString(fmt.Sprintf("%d-", m["count"].(int)))
	buf.WriteString(fmt.Sprintf("%s-", m["update_strategy"].(string)))
	if m["node_repave_interval"] != nil {
		buf.WriteString(fmt.Sprintf("%d-", m["node_repave_interval"].(int)))
	}
	if m["min"] != nil {
		buf.WriteString(fmt.Sprintf("%d-", m["min"].(int)))
	}
	if m["max"] != nil {
		buf.WriteString(fmt.Sprintf("%d-", m["max"].(int)))
	}

	if v, found := m["instance_type"]; found {
		if len(v.([]interface{})) > 0 {
			ins := v.([]interface{})[0].(map[string]interface{})
			buf.WriteString(fmt.Sprintf("%d-", ins["min_cpu"].(int)))
			buf.WriteString(fmt.Sprintf("%d-", ins["min_memory_mb"].(int)))
		}
	}
	buf.WriteString(fmt.Sprintf("%s-", m["azs"].(*schema.Set).GoString()))

	return int(hash(buf.String()))
}

func resourceMachinePoolLibvirtHash(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})

	buf.WriteString(HashStringMap(m["additional_labels"]))
	buf.WriteString(HashStringMapList(m["taints"]))

	//d := m["disk"].([]interface{})[0].(map[string]interface{})
	buf.WriteString(fmt.Sprintf("%t-", m["control_plane"].(bool)))
	buf.WriteString(fmt.Sprintf("%t-", m["control_plane_as_worker"].(bool)))
	buf.WriteString(fmt.Sprintf("%s-", m["name"].(string)))
	buf.WriteString(fmt.Sprintf("%d-", m["count"].(int)))
	buf.WriteString(fmt.Sprintf("%s-", m["update_strategy"].(string)))
	if m["node_repave_interval"] != nil {
		buf.WriteString(fmt.Sprintf("%d-", m["node_repave_interval"].(int)))
	}
	if v, found := m["xsl_template"]; found {
		buf.WriteString(fmt.Sprintf("%s-", v.(string)))
	}

	if v, found := m["instance_type"]; found {
		if len(v.([]interface{})) > 0 {
			ins := v.([]interface{})[0].(map[string]interface{})
			buf.WriteString(fmt.Sprintf("%d-", ins["cpu"].(int)))
			buf.WriteString(fmt.Sprintf("%d-", ins["disk_size_gb"].(int)))
			buf.WriteString(fmt.Sprintf("%d-", ins["memory_mb"].(int)))
			buf.WriteString(fmt.Sprintf("%s-", ins["cpus_sets"].(string)))
			if ins["cache_passthrough"] != nil {
				buf.WriteString(fmt.Sprintf("%s-%t", "cache_passthrough", ins["cache_passthrough"].(bool)))
			}
			if ins["gpu_config"] != nil {
				config, _ := ins["gpu_config"].(map[string]interface{})
				if config != nil {
					buf.WriteString(fmt.Sprintf("%d-", config["num_gpus"].(int)))
					buf.WriteString(fmt.Sprintf("%s-", config["device_model"].(string)))
					buf.WriteString(fmt.Sprintf("%s-", config["vendor"].(string)))
					buf.WriteString(HashStringMap(config["addresses"]))
				}
			}

			if ins["attached_disks"] != nil {
				for _, disk := range ins["attached_disks"].([]interface{}) {
					diskMap := disk.(map[string]interface{})
					if diskMap["managed"] != nil {
						buf.WriteString(fmt.Sprintf("%s-%t", "managed", diskMap["managed"].(bool)))
					}
					if diskMap["size_in_gb"] != nil {
						buf.WriteString(fmt.Sprintf("%s-%d", "size_in_gb", diskMap["size_in_gb"].(int)))
					}
				}
			}
		}
	}

	return int(hash(buf.String()))
}

func resourceMachinePoolEdgeNativeHash(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})

	buf.WriteString(HashStringMap(m["additional_labels"]))
	buf.WriteString(HashStringMapList(m["taints"]))

	buf.WriteString(fmt.Sprintf("%t-", m["control_plane"].(bool)))
	buf.WriteString(fmt.Sprintf("%t-", m["control_plane_as_worker"].(bool)))
	buf.WriteString(fmt.Sprintf("%s-", m["name"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["update_strategy"].(string)))
	if m["node_repave_interval"] != nil {
		buf.WriteString(fmt.Sprintf("%d-", m["node_repave_interval"].(int)))
	}

	if _, found := m["host_uids"]; found {
		for _, host := range m["host_uids"].([]interface{}) {
			buf.WriteString(fmt.Sprintf("%s-", host.(string)))
		}
	}

	return int(hash(buf.String()))
}*/

func resourceClusterHash(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})

	buf.WriteString(fmt.Sprintf("%s-", m["uid"].(string)))

	return int(hash(buf.String()))
}

func HashStringMapList(v interface{}) string {
	var b bytes.Buffer
	m := v.([]interface{})

	if len(m) == 0 {
		return ""
	}

	hashes := make([]string, 0)

	for _, i := range m {
		hashes = append(hashes, HashStringMap(i))
	}

	sortedHashes := make([]string, len(hashes))
	copy(sortedHashes, hashes)
	sort.Strings(sortedHashes)

	for _, i := range sortedHashes {
		b.WriteString(i)
	}

	return b.String()
}

func HashStringMap(v interface{}) string {
	if v == nil || len(v.(map[string]interface{})) == 0 {
		return ""
	}

	var b bytes.Buffer
	m := v.(map[string]interface{})

	keys := make([]string, 0)
	for k := range m {
		keys = append(keys, k)
	}

	sortedKeys := make([]string, len(keys))
	copy(sortedKeys, keys)
	sort.Strings(sortedKeys)

	for _, k := range sortedKeys {
		b.WriteString(fmt.Sprintf("%s-%s", k, m[k].(string)))
	}

	return b.String()
}

func hash(s string) uint32 {
	h := fnv.New32a()
	_, _ = h.Write([]byte(s))
	return h.Sum32()
}
