package spectrocloud

import (
	"bytes"
	"fmt"
	"hash/fnv"
	"sort"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"gopkg.in/yaml.v3"
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
	if _, ok := m["disk_size_gb"]; ok {
		buf.WriteString(fmt.Sprintf("%d-", m["disk_size_gb"].(int)))
	}

	buf.WriteString(fmt.Sprintf("%s-", m["instance_type"].(string)))
	if _, ok := m["azs"]; ok {
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
	}
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

	// Include EKS-specific fields following MAAS pattern
	if val, ok := m["disk_size_gb"]; ok {
		buf.WriteString(fmt.Sprintf("%d-", val.(int)))
	}
	if val, ok := m["instance_type"]; ok {
		buf.WriteString(fmt.Sprintf("%s-", val.(string)))
	}
	if val, ok := m["capacity_type"]; ok {
		buf.WriteString(fmt.Sprintf("%s-", val.(string)))
	}
	if val, ok := m["max_price"]; ok && val != nil {
		buf.WriteString(fmt.Sprintf("%s-", val.(string)))
	}
	if val, ok := m["ami_type"]; ok && val != nil {
		buf.WriteString(fmt.Sprintf("%s-", val.(string)))
	}

	// Include AZ configuration
	if azSubnets, ok := m["az_subnets"].(map[string]interface{}); ok && len(azSubnets) > 0 {
		buf.WriteString(HashStringMap(azSubnets))
	}
	if azs, ok := m["azs"]; ok && azs != nil {
		azsList := azs.([]interface{})
		azsListStr := make([]string, len(azsList))
		for i, v := range azsList {
			azsListStr[i] = v.(string)
		}
		sort.Strings(azsListStr)
		azsStr := strings.Join(azsListStr, "-")
		buf.WriteString(fmt.Sprintf("%s-", azsStr))
	}

	// Include EKS launch template configuration
	if m["eks_launch_template"] != nil {
		buf.WriteString(eksLaunchTemplate(m["eks_launch_template"]))
	}

	return int(hash(buf.String()))
}

func resourceMachinePoolGkeHash(v interface{}) int {
	m := v.(map[string]interface{})
	buf := CommonHash(m)
	if _, ok := m["disk_size_gb"]; ok {
		buf.WriteString(fmt.Sprintf("%d-", m["disk_size_gb"].(int)))
	}
	buf.WriteString(fmt.Sprintf("%s-", m["instance_type"].(string)))
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

func resourceMachinePoolCustomCloudHash(v interface{}) int {
	m := v.(map[string]interface{})
	var buf bytes.Buffer
	if _, ok := m["name"]; ok {
		buf.WriteString(HashStringMap(m["name"]))
	}
	if _, ok := m["count"]; ok {
		buf.WriteString(HashStringMap(m["count"]))
	}
	if _, ok := m["additional_labels"]; ok {
		buf.WriteString(HashStringMap(m["additional_labels"]))
	}
	if _, ok := m["taints"]; ok {
		buf.WriteString(HashStringMapList(m["taints"]))
	}
	if val, ok := m["control_plane"]; ok {
		buf.WriteString(fmt.Sprintf("%t-", val.(bool)))
	}
	if val, ok := m["control_plane_as_worker"]; ok {
		buf.WriteString(fmt.Sprintf("%t-", val.(bool)))
	}
	buf.WriteString(fmt.Sprintf("%s-", m["node_pool_config"].(string)))

	// Include overrides in hash calculation for change detection
	if overrides, ok := m["overrides"]; ok {
		buf.WriteString(HashStringMap(overrides))
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
	if azs, ok := m["azs"]; ok && azs != nil {
		buf.WriteString(fmt.Sprintf("%s-", azs.(*schema.Set).GoString()))
	}
	if nodeTags, ok := m["node_tags"]; ok && nodeTags != nil {
		buf.WriteString(fmt.Sprintf("%s-", nodeTags.(*schema.Set).GoString()))
	}

	// Include placement fields if present
	if placementRaw, ok := m["placement"]; ok {
		placementList := placementRaw.([]interface{})
		if len(placementList) > 0 {
			place := placementList[0].(map[string]interface{})
			if rp, ok := place["resource_pool"]; ok && rp != nil {
				buf.WriteString(fmt.Sprintf("%s-", rp.(string)))
			}
		}
	}

	// Include use_lxd_vm flag
	if v, ok := m["use_lxd_vm"]; ok {
		buf.WriteString(fmt.Sprintf("%t-", v.(bool)))
	}

	// Include network settings if present
	if networkRaw, ok := m["network"]; ok {
		networkList := networkRaw.([]interface{})
		if len(networkList) > 0 {
			net := networkList[0].(map[string]interface{})
			if name, ok := net["network_name"]; ok && name != nil {
				buf.WriteString(fmt.Sprintf("%s-", name.(string)))
			}
			if parent, ok := net["parent_pool_uid"]; ok && parent != nil {
				buf.WriteString(fmt.Sprintf("%s-", parent.(string)))
			}
			if staticIP, ok := net["static_ip"]; ok {
				buf.WriteString(fmt.Sprintf("%t-", staticIP.(bool)))
			}
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

	if edgeHosts, found := m["edge_host"]; found {
		for _, host := range edgeHosts.([]interface{}) {
			hostMap := host.(map[string]interface{})

			if hostName, ok := hostMap["host_name"]; ok {
				buf.WriteString(fmt.Sprintf("host_name:%s-", hostName.(string)))
			}

			if hostUID, ok := hostMap["host_uid"]; ok {
				buf.WriteString(fmt.Sprintf("host_uid:%s-", hostUID.(string)))
			}

			if staticIP, ok := hostMap["static_ip"]; ok {
				buf.WriteString(fmt.Sprintf("static_ip:%s-", staticIP.(string)))
			}

			if nicName, ok := hostMap["nic_name"]; ok {
				buf.WriteString(fmt.Sprintf("nic_name:%s-", nicName.(string)))
			}

			if defaultGateway, ok := hostMap["default_gateway"]; ok {
				buf.WriteString(fmt.Sprintf("default_gateway:%s-", defaultGateway.(string)))
			}

			if subnetMask, ok := hostMap["subnet_mask"]; ok {
				buf.WriteString(fmt.Sprintf("subnet_mask:%s-", subnetMask.(string)))
			}

			if dnsServers, ok := hostMap["dns_servers"]; ok {
				var dns []string
				for _, v := range dnsServers.(*schema.Set).List() {
					dns = append(dns, v.(string))
				}
				buf.WriteString(fmt.Sprintf("dns_servers:%s-", strings.Join(dns, ",")))
			}

			if twoNodeRole, ok := hostMap["two_node_role"]; ok {
				buf.WriteString(fmt.Sprintf("two_node_role:%s-", twoNodeRole.(string)))
			}
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

// YamlContentHash creates a hash based on YAML semantic content, ignoring formatting
func YamlContentHash(yamlContent string) string {
	canonicalContent := yamlContentToCanonicalString(yamlContent)
	h := fnv.New64a()
	if _, err := h.Write([]byte(canonicalContent)); err != nil {
		// If hash writing fails, return a fallback hash
		return fmt.Sprintf("error_hash_%d", time.Now().UnixNano())
	}
	return fmt.Sprintf("%x", h.Sum64())
}

// yamlContentToCanonicalString converts YAML content to a canonical string for hashing
func yamlContentToCanonicalString(yamlContent string) string {
	if strings.TrimSpace(yamlContent) == "" {
		return ""
	}

	// Split multi-document YAML
	documents := strings.Split(yamlContent, "---")
	var canonicalDocs []string

	for _, doc := range documents {
		doc = strings.TrimSpace(doc)
		if doc == "" {
			continue
		}

		// Parse YAML document
		var yamlData interface{}
		if err := yaml.Unmarshal([]byte(doc), &yamlData); err != nil {
			// If parsing fails, use original doc for canonical form
			canonicalDocs = append(canonicalDocs, doc)
			continue
		}

		// Convert to canonical string representation
		canonical := toCanonicalString(yamlData)
		canonicalDocs = append(canonicalDocs, canonical)
	}

	if len(canonicalDocs) == 0 {
		return ""
	}

	return strings.Join(canonicalDocs, "|||") // Use ||| as document separator
}

// toCanonicalString converts a YAML structure to a deterministic string representation
func toCanonicalString(data interface{}) string {
	switch v := data.(type) {
	case map[string]interface{}:
		// Sort keys for deterministic output
		keys := make([]string, 0, len(v))
		for k := range v {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		var parts []string
		for _, k := range keys {
			value := toCanonicalString(v[k])
			parts = append(parts, fmt.Sprintf("%s:%s", k, value))
		}
		return "{" + strings.Join(parts, ",") + "}"

	case map[interface{}]interface{}:
		// Convert to string map and recurse
		stringMap := make(map[string]interface{})
		for key, value := range v {
			if keyStr, ok := key.(string); ok {
				stringMap[keyStr] = value
			}
		}
		return toCanonicalString(stringMap)

	case []interface{}:
		var parts []string
		for _, item := range v {
			parts = append(parts, toCanonicalString(item))
		}
		return "[" + strings.Join(parts, ",") + "]"

	case string:
		return fmt.Sprintf("\"%s\"", v)
	case int, int64, float64, bool:
		return fmt.Sprintf("%v", v)
	case nil:
		return "null"
	default:
		return fmt.Sprintf("%v", v)
	}
}

// NormalizeYamlContent parses YAML and re-serializes it in a consistent format for StateFunc
func NormalizeYamlContent(yamlContent string) string {
	if strings.TrimSpace(yamlContent) == "" {
		return ""
	}

	// Split multi-document YAML
	documents := strings.Split(yamlContent, "---")
	var normalizedDocs []string

	for _, doc := range documents {
		doc = strings.TrimSpace(doc)
		if doc == "" {
			continue
		}

		// Parse YAML document
		var yamlData interface{}
		if err := yaml.Unmarshal([]byte(doc), &yamlData); err != nil {
			// If parsing fails, return original (trimmed)
			normalizedDocs = append(normalizedDocs, doc)
			continue
		}

		// Re-serialize in consistent format
		normalizedYaml, err := yaml.Marshal(yamlData)
		if err != nil {
			// If marshaling fails, return original (trimmed)
			normalizedDocs = append(normalizedDocs, doc)
			continue
		}

		normalizedDocs = append(normalizedDocs, strings.TrimSpace(string(normalizedYaml)))
	}

	if len(normalizedDocs) == 0 {
		return ""
	}

	return strings.Join(normalizedDocs, "\n---\n")
}
