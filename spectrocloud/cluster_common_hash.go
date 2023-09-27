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
	if _, ok := nodePool["node"]; ok {
		buf.WriteString(HashStringMapList(nodePool["node"]))
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
	if v == nil {
		return ""
	}

	m, ok := v.(map[string]interface{})
	if !ok || len(m) == 0 {
		return ""
	}

	var b bytes.Buffer

	// Create and sort the keys
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Construct the string based on sorted keys
	for _, k := range keys {
		b.WriteString(fmt.Sprintf("%s-%s", k, m[k].(string)))
	}

	return b.String()
}

func hash(s string) uint32 {
	h := fnv.New32a()
	_, _ = h.Write([]byte(s))
	return h.Sum32()
}
