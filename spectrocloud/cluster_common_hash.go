package spectrocloud

import (
	"bytes"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"hash/fnv"
	"sort"
)

func resourceMachinePoolAzureHash(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})
	//d := m["disk"].([]interface{})[0].(map[string]interface{})
	buf.WriteString(fmt.Sprintf("%t-", m["control_plane"].(bool)))
	buf.WriteString(fmt.Sprintf("%t-", m["control_plane_as_worker"].(bool)))
	buf.WriteString(fmt.Sprintf("%s-", m["name"].(string)))
	buf.WriteString(fmt.Sprintf("%d-", m["count"].(int)))
	buf.WriteString(fmt.Sprintf("%s-", m["update_strategy"].(string)))

	buf.WriteString(fmt.Sprintf("%s-", m["instance_type"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["azs"].(*schema.Set).GoString()))

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
	//d := m["disk"].([]interface{})[0].(map[string]interface{})
	buf.WriteString(fmt.Sprintf("%t-", m["control_plane"].(bool)))
	buf.WriteString(fmt.Sprintf("%t-", m["control_plane_as_worker"].(bool)))
	buf.WriteString(fmt.Sprintf("%s-", m["name"].(string)))
	buf.WriteString(fmt.Sprintf("%d-", m["count"].(int)))
	buf.WriteString(fmt.Sprintf("%s-", m["update_strategy"].(string)))

	buf.WriteString(fmt.Sprintf("%s-", m["instance_type"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["azs"].(*schema.Set).GoString()))

	return int(hash(buf.String()))
}

func resourceMachinePoolAwsHash(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})
	buf.WriteString(fmt.Sprintf("%t-", m["control_plane"].(bool)))
	buf.WriteString(fmt.Sprintf("%t-", m["control_plane_as_worker"].(bool)))
	buf.WriteString(fmt.Sprintf("%s-", m["name"].(string)))
	buf.WriteString(fmt.Sprintf("%d-", m["count"].(int)))
	buf.WriteString(fmt.Sprintf("%s-", m["update_strategy"].(string)))

	buf.WriteString(fmt.Sprintf("%s-", m["instance_type"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["capacity_type"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["max_price"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["azs"].(*schema.Set).GoString()))

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
	buf.WriteString(fmt.Sprintf("%s-", m["update_strategy"].(string)))

	if v, found := m["instance_type"]; found {
		if len(v.([]interface{})) > 0 {
			ins := v.([]interface{})[0].(map[string]interface{})
			buf.WriteString(fmt.Sprintf("%d-", ins["cpu"].(int)))
			buf.WriteString(fmt.Sprintf("%d-", ins["disk_size_gb"].(int)))
			buf.WriteString(fmt.Sprintf("%d-", ins["memory_mb"].(int)))
		}
	}

	return int(hash(buf.String()))
}

func resourceMachinePoolOpenStackHash(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})

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

func resourceMachinePoolMaasHash(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})
	//d := m["disk"].([]interface{})[0].(map[string]interface{})
	buf.WriteString(fmt.Sprintf("%t-", m["control_plane"].(bool)))
	buf.WriteString(fmt.Sprintf("%t-", m["control_plane_as_worker"].(bool)))
	buf.WriteString(fmt.Sprintf("%s-", m["name"].(string)))
	buf.WriteString(fmt.Sprintf("%d-", m["count"].(int)))
	buf.WriteString(fmt.Sprintf("%s-", m["update_strategy"].(string)))

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

	if v, found := m["instance_type"]; found {
		if len(v.([]interface{})) > 0 {
			ins := v.([]interface{})[0].(map[string]interface{})
			buf.WriteString(fmt.Sprintf("%d-", ins["cpu"].(int)))
			buf.WriteString(fmt.Sprintf("%d-", ins["disk_size_gb"].(int)))
			buf.WriteString(fmt.Sprintf("%d-", ins["memory_mb"].(int)))
			buf.WriteString(fmt.Sprintf("%d-", ins["cpus_sets"].(string)))
			if ins["cache_passthrough"] != nil {
				buf.WriteString(fmt.Sprintf("%s-%s", "cache_passthrough", ins["cache_passthrough"].(bool)))
			}
			if ins["gpu_config"] != nil {
				config, _ := ins["gpu_config"].(map[string]interface{})
				if config != nil {
					buf.WriteString(fmt.Sprintf("%d-", config["num_gpus"].(int)))
					buf.WriteString(fmt.Sprintf("%d-", config["device_model"].(string)))
					buf.WriteString(fmt.Sprintf("%d-", config["vendor"].(string)))
				}
			}

			if ins["attached_disks"] != nil {
				for _, disk := range ins["attached_disks"].([]interface{}) {
					diskMap := disk.(map[string]interface{})
					if diskMap["managed"] != nil {
						buf.WriteString(fmt.Sprintf("%s-%s", "managed", diskMap["managed"].(bool)))
					}
					if diskMap["size_in_gb"] != nil {
						buf.WriteString(fmt.Sprintf("%s-%s", "size_in_gb", diskMap["size_in_gb"].(int)))
					}
				}
			}
		}
	}

	return int(hash(buf.String()))
}

func resourceMachinePoolEdgeHash(v interface{}) int {
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

func resourceClusterHash(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})

	buf.WriteString(fmt.Sprintf("%s-", m["uid"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["name"].(string)))

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

	sortedHashes := make([]string, len(hashes), len(hashes))
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

	sortedKeys := make([]string, len(keys), len(keys))
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
