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
		var boolVal bool
		switch v := val.(type) {
		case bool:
			boolVal = v
		case *bool:
			if v != nil {
				boolVal = *v
			}
		}
		buf.WriteString(fmt.Sprintf("%t-", boolVal))
	}
	if val, ok := nodePool["control_plane_as_worker"]; ok {
		var boolVal bool
		switch v := val.(type) {
		case bool:
			boolVal = v
		case *bool:
			if v != nil {
				boolVal = *v
			}
		}
		buf.WriteString(fmt.Sprintf("%t-", boolVal))
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
	// Hash override_scaling if present
	if overrideScaling, ok := nodePool["override_scaling"].([]interface{}); ok && len(overrideScaling) > 0 {
		scalingConfig := overrideScaling[0].(map[string]interface{})
		if maxSurge, ok := scalingConfig["max_surge"].(string); ok && maxSurge != "" {
			buf.WriteString(fmt.Sprintf("max_surge:%s-", maxSurge))
		}
		if maxUnavailable, ok := scalingConfig["max_unavailable"].(string); ok && maxUnavailable != "" {
			buf.WriteString(fmt.Sprintf("max_unavailable:%s-", maxUnavailable))
		}
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

	// Hash additional annotations and override_kubeadm_configuration
	if _, ok := m["additional_annotations"]; ok {
		buf.WriteString(HashStringMap(m["additional_annotations"]))
	}
	if val, ok := m["override_kubeadm_configuration"].(string); ok && val != "" {
		fmt.Fprintf(buf, "%s-", val)
	}

	if val, ok := m["instance_type"]; ok {
		fmt.Fprintf(buf, "%s-", val.(string))
	}
	if val, ok := m["is_system_node_pool"]; ok {
		var boolVal bool
		switch v := val.(type) {
		case bool:
			boolVal = v
		case *bool:
			if v != nil {
				boolVal = *v
			}
		}
		fmt.Fprintf(buf, "%t-", boolVal)
	}
	if val, ok := m["os_type"]; ok && val != "" {
		fmt.Fprintf(buf, "%s-", val.(string))
	}

	return int(hash(buf.String()))
}

func resourceMachinePoolAksHash(v interface{}) int {
	nodePool := v.(map[string]interface{})
	var buf bytes.Buffer

	// Hash additional annotations and override_kubeadm_configuration
	if _, ok := nodePool["additional_annotations"]; ok {
		buf.WriteString(HashStringMap(nodePool["additional_annotations"]))
	}
	if val, ok := nodePool["override_kubeadm_configuration"].(string); ok && val != "" {
		fmt.Fprintf(&buf, "%s-", val)
	}

	// Include all fields that should trigger a machine pool update
	if val, ok := nodePool["name"]; ok {
		buf.WriteString(fmt.Sprintf("%s-", val.(string)))
	}
	if val, ok := nodePool["count"]; ok {
		buf.WriteString(fmt.Sprintf("%d-", val.(int)))
	}
	if val, ok := nodePool["instance_type"]; ok {
		buf.WriteString(fmt.Sprintf("%s-", val.(string)))
	}
	if val, ok := nodePool["disk_size_gb"]; ok {
		buf.WriteString(fmt.Sprintf("%d-", val.(int)))
	}
	if val, ok := nodePool["is_system_node_pool"]; ok {
		var boolVal bool
		switch v := val.(type) {
		case bool:
			boolVal = v
		case *bool:
			if v != nil {
				boolVal = *v
			}
		}
		buf.WriteString(fmt.Sprintf("%t-", boolVal))
	}
	if val, ok := nodePool["storage_account_type"]; ok {
		buf.WriteString(fmt.Sprintf("%s-", val.(string)))
	}

	// Additional labels (map)
	if _, ok := nodePool["additional_labels"]; ok {
		buf.WriteString(HashStringMap(nodePool["additional_labels"]))
	}

	// Update strategy
	if val, ok := nodePool["update_strategy"]; ok {
		buf.WriteString(fmt.Sprintf("%s-", val.(string)))
	}

	// Hash override_scaling
	if overrideScaling, ok := nodePool["override_scaling"].([]interface{}); ok && len(overrideScaling) > 0 {
		scalingConfig := overrideScaling[0].(map[string]interface{})
		if maxSurge, ok := scalingConfig["max_surge"].(string); ok && maxSurge != "" {
			fmt.Fprintf(&buf, "max_surge:%s-", maxSurge)
		}
		if maxUnavailable, ok := scalingConfig["max_unavailable"].(string); ok && maxUnavailable != "" {
			fmt.Fprintf(&buf, "max_unavailable:%s-", maxUnavailable)
		}
	}

	// Min and Max for autoscaling
	if nodePool["min"] != nil {
		buf.WriteString(fmt.Sprintf("%d-", nodePool["min"].(int)))
	}
	if nodePool["max"] != nil {
		buf.WriteString(fmt.Sprintf("%d-", nodePool["max"].(int)))
	}

	// Node configuration (list of maps)
	if nodePool["node"] != nil {
		buf.WriteString(HashStringMapList(nodePool["node"]))
	}

	// Taints (list of maps)
	if _, ok := nodePool["taints"]; ok {
		buf.WriteString(HashStringMapList(nodePool["taints"]))
	}

	return int(hash(buf.String()))
}

func resourceMachinePoolGcpHash(v interface{}) int {
	m := v.(map[string]interface{})
	buf := CommonHash(m)

	// Hash additional annotations and override_kubeadm_configuration
	if _, ok := m["additional_annotations"]; ok {
		buf.WriteString(HashStringMap(m["additional_annotations"]))
	}
	if val, ok := m["override_kubeadm_configuration"].(string); ok && val != "" {
		fmt.Fprintf(buf, "%s-", val)
	}

	if _, ok := m["disk_size_gb"]; ok {
		fmt.Fprintf(buf, "%d-", m["disk_size_gb"].(int))
	}

	fmt.Fprintf(buf, "%s-", m["instance_type"].(string))
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
			fmt.Fprintf(buf, "%s-", azsStr)
		}
	}
	return int(hash(buf.String()))
}

func resourceMachinePoolAwsHash(v interface{}) int {
	m := v.(map[string]interface{})
	buf := CommonHash(m)

	// Hash additional annotations and override_kubeadm_configuration
	if _, ok := m["additional_annotations"]; ok {
		buf.WriteString(HashStringMap(m["additional_annotations"]))
	}
	if val, ok := m["override_kubeadm_configuration"].(string); ok && val != "" {
		fmt.Fprintf(buf, "%s-", val)
	}

	if m["min"] != nil {
		fmt.Fprintf(buf, "%d-", m["min"].(int))
	}
	if m["max"] != nil {
		fmt.Fprintf(buf, "%d-", m["max"].(int))
	}
	fmt.Fprintf(buf, "%s-", m["instance_type"].(string))
	fmt.Fprintf(buf, "%s-", m["capacity_type"].(string))
	fmt.Fprintf(buf, "%s-", m["max_price"].(string))
	if m["azs"] != nil {
		azsSet := m["azs"].(*schema.Set)
		azsList := azsSet.List()
		azsListStr := make([]string, len(azsList))
		for i, v := range azsList {
			azsListStr[i] = v.(string)
		}
		sort.Strings(azsListStr)
		azsStr := strings.Join(azsListStr, "-")
		fmt.Fprintf(buf, "%s-", azsStr)
	}
	fmt.Fprintf(buf, "%s-", m["azs"].(*schema.Set).GoString())
	buf.WriteString(HashStringMap(m["az_subnets"]))

	return int(hash(buf.String()))
}

func resourceMachinePoolEksHash(v interface{}) int {
	nodePool := v.(map[string]interface{})
	var buf bytes.Buffer

	// Hash additional annotations and override_kubeadm_configuration
	if _, ok := nodePool["additional_annotations"]; ok {
		buf.WriteString(HashStringMap(nodePool["additional_annotations"]))
	}
	if val, ok := nodePool["override_kubeadm_configuration"].(string); ok && val != "" {
		fmt.Fprintf(&buf, "%s-", val)
	}

	if val, ok := nodePool["count"]; ok {
		buf.WriteString(fmt.Sprintf("%d-", val.(int)))
	}
	if val, ok := nodePool["disk_size_gb"]; ok {
		buf.WriteString(fmt.Sprintf("%d-", val.(int)))
	}
	if val, ok := nodePool["instance_type"]; ok {
		buf.WriteString(fmt.Sprintf("%s-", val.(string)))
	}
	if val, ok := nodePool["name"]; ok {
		buf.WriteString(fmt.Sprintf("%s-", val.(string)))
	}

	if _, ok := nodePool["additional_labels"]; ok {
		buf.WriteString(HashStringMap(nodePool["additional_labels"]))
	}
	if val, ok := nodePool["ami_type"]; ok {
		buf.WriteString(fmt.Sprintf("%s-", val.(string)))
	}

	keys := make([]string, 0, len(nodePool["az_subnets"].(map[string]interface{})))
	for k := range nodePool["az_subnets"].(map[string]interface{}) {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		buf.WriteString(fmt.Sprintf("%s-%s", k, nodePool["az_subnets"].(map[string]interface{})[k].(string)))
	}

	if nodePool["azs"] != nil {
		azsList := nodePool["azs"].([]interface{})
		azsListStr := make([]string, len(azsList))
		for i, v := range azsList {
			azsListStr[i] = v.(string)
		}
		sort.Strings(azsListStr)
		azsStr := strings.Join(azsListStr, "-")
		buf.WriteString(fmt.Sprintf("%s-", azsStr))
	}

	if val, ok := nodePool["capacity_type"]; ok {
		buf.WriteString(fmt.Sprintf("%s-", val.(string)))
	}

	if nodePool["min"] != nil {
		buf.WriteString(fmt.Sprintf("%d-", nodePool["min"].(int)))
	}
	if nodePool["max"] != nil {
		buf.WriteString(fmt.Sprintf("%d-", nodePool["max"].(int)))
	}
	if nodePool["max_price"] != nil {
		buf.WriteString(fmt.Sprintf("%s-", nodePool["max_price"].(string)))
	}
	if nodePool["node"] != nil {
		buf.WriteString(HashStringMapList(nodePool["node"]))
	}
	if _, ok := nodePool["taints"]; ok {
		buf.WriteString(HashStringMapList(nodePool["taints"]))
	}
	if nodePool["eks_launch_template"] != nil {
		buf.WriteString(eksLaunchTemplate(nodePool["eks_launch_template"]))
	}
	if val, ok := nodePool["update_strategy"]; ok {
		buf.WriteString(fmt.Sprintf("%s-", val.(string)))
	}

	return int(hash(buf.String()))
}

func resourceMachinePoolGkeHash(v interface{}) int {
	nodePool := v.(map[string]interface{})
	var buf bytes.Buffer

	// Hash additional annotations and override_kubeadm_configuration
	if _, ok := nodePool["additional_annotations"]; ok {
		buf.WriteString(HashStringMap(nodePool["additional_annotations"]))
	}
	if val, ok := nodePool["override_kubeadm_configuration"].(string); ok && val != "" {
		fmt.Fprintf(&buf, "%s-", val)
	}

	// Include all fields that should trigger a machine pool update
	if val, ok := nodePool["name"]; ok {
		buf.WriteString(fmt.Sprintf("%s-", val.(string)))
	}
	if val, ok := nodePool["count"]; ok {
		buf.WriteString(fmt.Sprintf("%d-", val.(int)))
	}
	if val, ok := nodePool["disk_size_gb"]; ok {
		buf.WriteString(fmt.Sprintf("%d-", val.(int)))
	}
	if val, ok := nodePool["instance_type"]; ok {
		buf.WriteString(fmt.Sprintf("%s-", val.(string)))
	}

	// Additional labels (map)
	if _, ok := nodePool["additional_labels"]; ok {
		buf.WriteString(HashStringMap(nodePool["additional_labels"]))
	}

	// Update strategy
	if val, ok := nodePool["update_strategy"]; ok {
		buf.WriteString(fmt.Sprintf("%s-", val.(string)))
	}

	// Hash override_scaling
	if overrideScaling, ok := nodePool["override_scaling"].([]interface{}); ok && len(overrideScaling) > 0 {
		scalingConfig := overrideScaling[0].(map[string]interface{})
		if maxSurge, ok := scalingConfig["max_surge"].(string); ok && maxSurge != "" {
			fmt.Fprintf(&buf, "max_surge:%s-", maxSurge)
		}
		if maxUnavailable, ok := scalingConfig["max_unavailable"].(string); ok && maxUnavailable != "" {
			fmt.Fprintf(&buf, "max_unavailable:%s-", maxUnavailable)
		}
	}

	// Node configuration (list of maps)
	if nodePool["node"] != nil {
		buf.WriteString(HashStringMapList(nodePool["node"]))
	}

	// Taints (list of maps)
	if _, ok := nodePool["taints"]; ok {
		buf.WriteString(HashStringMapList(nodePool["taints"]))
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

func resourceMachinePoolVsphereHash(v interface{}) int {
	m := v.(map[string]interface{})
	buf := CommonHash(m)

	// Hash additional annotations and override_kubeadm_configuration
	if _, ok := m["additional_annotations"]; ok {
		buf.WriteString(HashStringMap(m["additional_annotations"]))
	}
	if val, ok := m["override_kubeadm_configuration"].(string); ok && val != "" {
		fmt.Fprintf(buf, "%s-", val)
	}

	if v, found := m["instance_type"]; found {
		if len(v.([]interface{})) > 0 {
			ins := v.([]interface{})[0].(map[string]interface{})
			fmt.Fprintf(buf, "%d-", ins["cpu"].(int))
			fmt.Fprintf(buf, "%d-", ins["disk_size_gb"].(int))
			fmt.Fprintf(buf, "%d-", ins["memory_mb"].(int))
		}
	}

	if placements, found := m["placement"]; found {
		for _, p := range placements.([]interface{}) {
			place := p.(map[string]interface{})
			fmt.Fprintf(buf, "%s-", place["cluster"].(string))
			fmt.Fprintf(buf, "%s-", place["resource_pool"].(string))
			fmt.Fprintf(buf, "%s-", place["datastore"].(string))
			fmt.Fprintf(buf, "%s-", place["network"].(string))
			fmt.Fprintf(buf, "%s-", place["static_ip_pool_id"].(string))
		}
	}
	return int(hash(buf.String()))
}

func resourceMachinePoolCustomCloudHash(v interface{}) int {
	m := v.(map[string]interface{})
	var buf bytes.Buffer

	// IMPORTANT: Only include user-provided fields in hash
	// Do NOT include computed fields (name, count, additional_labels) as they cause perpetual diffs

	if _, ok := m["taints"]; ok {
		buf.WriteString(HashStringMapList(m["taints"]))
	}
	if val, ok := m["control_plane"]; ok {
		var boolVal bool
		switch v := val.(type) {
		case bool:
			boolVal = v
		case *bool:
			if v != nil {
				boolVal = *v
			}
		}
		buf.WriteString(fmt.Sprintf("%t-", boolVal))
	}
	if val, ok := m["control_plane_as_worker"]; ok {
		var boolVal bool
		switch v := val.(type) {
		case bool:
			boolVal = v
		case *bool:
			if v != nil {
				boolVal = *v
			}
		}
		buf.WriteString(fmt.Sprintf("%t-", boolVal))
	}

	// Normalize YAML to match StateFunc behavior (critical for preventing perpetual diffs)
	if yamlContent, ok := m["node_pool_config"].(string); ok {
		normalizedYAML := NormalizeYamlContent(yamlContent)
		buf.WriteString(fmt.Sprintf("%s-", normalizedYAML))
	}

	// Include overrides in hash calculation for change detection
	if overrides, ok := m["overrides"]; ok {
		buf.WriteString(HashStringMap(overrides))
	}

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

	// Hash additional annotations and override_kubeadm_configuration
	if _, ok := m["additional_annotations"]; ok {
		buf.WriteString(HashStringMap(m["additional_annotations"]))
	}
	if val, ok := m["override_kubeadm_configuration"].(string); ok && val != "" {
		fmt.Fprintf(buf, "%s-", val)
	}

	if v, found := m["instance_type"]; found {
		if len(v.([]interface{})) > 0 {
			ins := v.([]interface{})[0].(map[string]interface{})
			fmt.Fprintf(buf, "%d-", ins["min_cpu"].(int))
			fmt.Fprintf(buf, "%d-", ins["min_memory_mb"].(int))
		}
	}
	if azs, ok := m["azs"]; ok && azs != nil {
		fmt.Fprintf(buf, "%s-", azs.(*schema.Set).GoString())
	}
	if nodeTags, ok := m["node_tags"]; ok && nodeTags != nil {
		fmt.Fprintf(buf, "%s-", nodeTags.(*schema.Set).GoString())
	}

	// Include placement fields if present
	if placementRaw, ok := m["placement"]; ok {
		placementList := placementRaw.([]interface{})
		if len(placementList) > 0 {
			place := placementList[0].(map[string]interface{})
			if rp, ok := place["resource_pool"]; ok && rp != nil {
				fmt.Fprintf(buf, "%s-", rp.(string))
			}
		}
	}

	// Include use_lxd_vm flag
	if v, ok := m["use_lxd_vm"]; ok {
		fmt.Fprintf(buf, "%t-", v.(bool))
	}

	// Include network settings if present
	if networkRaw, ok := m["network"]; ok {
		networkList := networkRaw.([]interface{})
		if len(networkList) > 0 {
			net := networkList[0].(map[string]interface{})
			if name, ok := net["network_name"]; ok && name != nil {
				fmt.Fprintf(buf, "%s-", name.(string))
			}
			if parent, ok := net["parent_pool_uid"]; ok && parent != nil {
				fmt.Fprintf(buf, "%s-", parent.(string))
			}
			if staticIP, ok := net["static_ip"]; ok {
				fmt.Fprintf(buf, "%t-", staticIP.(bool))
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

	// Hash additional annotations and override_kubeadm_configuration
	if _, ok := m["additional_annotations"]; ok {
		buf.WriteString(HashStringMap(m["additional_annotations"]))
	}
	if val, ok := m["override_kubeadm_configuration"].(string); ok && val != "" {
		fmt.Fprintf(buf, "%s-", val)
	}

	if edgeHosts, found := m["edge_host"]; found {
		var edgeHostList []interface{}
		if edgeHostSet, ok := edgeHosts.(*schema.Set); ok {
			edgeHostList = edgeHostSet.List()
		} else if edgeHostListRaw, ok := edgeHosts.([]interface{}); ok {
			// Fallback for backward compatibility
			edgeHostList = edgeHostListRaw
		}

		for _, host := range edgeHostList {
			hostMap := host.(map[string]interface{})

			if hostName, ok := hostMap["host_name"]; ok {
				fmt.Fprintf(buf, "host_name:%s-", hostName.(string))
			}

			if hostUID, ok := hostMap["host_uid"]; ok {
				fmt.Fprintf(buf, "host_uid:%s-", hostUID.(string))
			}

			if staticIP, ok := hostMap["static_ip"]; ok {
				fmt.Fprintf(buf, "static_ip:%s-", staticIP.(string))
			}

			if nicName, ok := hostMap["nic_name"]; ok {
				fmt.Fprintf(buf, "nic_name:%s-", nicName.(string))
			}

			if defaultGateway, ok := hostMap["default_gateway"]; ok {
				fmt.Fprintf(buf, "default_gateway:%s-", defaultGateway.(string))
			}

			if subnetMask, ok := hostMap["subnet_mask"]; ok {
				fmt.Fprintf(buf, "subnet_mask:%s-", subnetMask.(string))
			}

			if dnsServers, ok := hostMap["dns_servers"]; ok {
				var dns []string
				for _, v := range dnsServers.(*schema.Set).List() {
					dns = append(dns, v.(string))
				}
				fmt.Fprintf(buf, "dns_servers:%s-", strings.Join(dns, ","))
			}

			if twoNodeRole, ok := hostMap["two_node_role"]; ok {
				fmt.Fprintf(buf, "two_node_role:%s-", twoNodeRole.(string))
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

func resourceMachinePoolOpenStackHash(v interface{}) int {
	nodePool := v.(map[string]interface{})
	var buf bytes.Buffer

	// Use CommonHash for common fields: additional_labels, taints, control_plane,
	// control_plane_as_worker, name, count, update_strategy, node_repave_interval, node
	commonBuf := CommonHash(nodePool)
	buf.WriteString(commonBuf.String())

	// Hash additional annotations and override_kubeadm_configuration
	if _, ok := nodePool["additional_annotations"]; ok {
		buf.WriteString(HashStringMap(nodePool["additional_annotations"]))
	}
	if val, ok := nodePool["override_kubeadm_configuration"].(string); ok && val != "" {
		fmt.Fprintf(&buf, "%s-", val)
	}

	// Add OpenStack-specific fields
	if val, ok := nodePool["instance_type"]; ok {
		buf.WriteString(fmt.Sprintf("%s-", val.(string)))
	}

	// Handle azs (TypeSet) - sort for deterministic hash
	if nodePool["azs"] != nil {
		azsSet := nodePool["azs"].(*schema.Set)
		azsList := azsSet.List()
		azsListStr := make([]string, len(azsList))
		for i, v := range azsList {
			azsListStr[i] = v.(string)
		}
		sort.Strings(azsListStr)
		azsStr := strings.Join(azsListStr, "-")
		buf.WriteString(fmt.Sprintf("%s-", azsStr))
	}

	// Handle subnet_id (optional string field)
	if val, ok := nodePool["subnet_id"]; ok && val != nil && val.(string) != "" {
		buf.WriteString(fmt.Sprintf("%s-", val.(string)))
	}

	return int(hash(buf.String()))
}

// resourceEdgeHostHash creates a hash for edge_host TypeSet
func resourceEdgeHostHash(v interface{}) int {
	var buf bytes.Buffer
	host := v.(map[string]interface{})

	// Required field - always include
	if hostUID, ok := host["host_uid"]; ok && hostUID != nil {
		buf.WriteString(fmt.Sprintf("host_uid:%s-", hostUID.(string)))
	}

	// Optional fields
	if hostName, ok := host["host_name"]; ok && hostName != nil && hostName.(string) != "" {
		buf.WriteString(fmt.Sprintf("host_name:%s-", hostName.(string)))
	}

	if staticIP, ok := host["static_ip"]; ok && staticIP != nil && staticIP.(string) != "" {
		buf.WriteString(fmt.Sprintf("static_ip:%s-", staticIP.(string)))
	}

	if nicName, ok := host["nic_name"]; ok && nicName != nil && nicName.(string) != "" {
		buf.WriteString(fmt.Sprintf("nic_name:%s-", nicName.(string)))
	}

	if defaultGateway, ok := host["default_gateway"]; ok && defaultGateway != nil && defaultGateway.(string) != "" {
		buf.WriteString(fmt.Sprintf("default_gateway:%s-", defaultGateway.(string)))
	}

	if subnetMask, ok := host["subnet_mask"]; ok && subnetMask != nil && subnetMask.(string) != "" {
		buf.WriteString(fmt.Sprintf("subnet_mask:%s-", subnetMask.(string)))
	}

	// Handle dns_servers (TypeSet) - sort for deterministic hash
	if dnsServers, ok := host["dns_servers"]; ok && dnsServers != nil {
		if dnsSet, ok := dnsServers.(*schema.Set); ok {
			dnsList := dnsSet.List()
			dnsListStr := make([]string, len(dnsList))
			for i, v := range dnsList {
				dnsListStr[i] = v.(string)
			}
			sort.Strings(dnsListStr)
			buf.WriteString(fmt.Sprintf("dns_servers:%s-", strings.Join(dnsListStr, ",")))
		}
	}

	if twoNodeRole, ok := host["two_node_role"]; ok && twoNodeRole != nil && twoNodeRole.(string) != "" {
		buf.WriteString(fmt.Sprintf("two_node_role:%s-", twoNodeRole.(string)))
	}

	return int(hash(buf.String()))
}
