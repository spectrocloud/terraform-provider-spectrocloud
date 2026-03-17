package convert

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spectrocloud/palette-sdk-go/api/models"
)

// ToHapiVmFromTFJSON builds HAPI VM from JSON returned by VirtualMachineResourceDataToJSON.
// No kubevirt import: spec is KubeVirt-shaped JSON, then unmarshaled into HAPI (like ToHapiVm + ToHapiVmSpecM).
func ToHapiVmFromTFJSON(data []byte) (*models.V1ClusterVirtualMachine, error) {
	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("unmarshal TF JSON: %w", err)
	}

	meta := metadataFromTFMap(m)
	specMap := tfMapToVMSpecShape(m)
	specJSON, err := json.Marshal(specMap)
	if err != nil {
		return nil, fmt.Errorf("marshal spec map: %w", err)
	}
	var spec models.V1ClusterVirtualMachineSpec
	if err := json.Unmarshal(specJSON, &spec); err != nil {
		return nil, fmt.Errorf("unmarshal spec to HAPI: %w", err)
	}

	var status *models.V1ClusterVirtualMachineStatus
	if raw, ok := m["status"]; ok && raw != nil {
		status = &models.V1ClusterVirtualMachineStatus{}
		b, _ := json.Marshal(camelCaseJSONKeys(raw))
		_ = json.Unmarshal(b, status)
	}

	return &models.V1ClusterVirtualMachine{
		Metadata: meta,
		Spec:     &spec,
		Status:   status,
	}, nil
}

func metadataFromTFMap(m map[string]interface{}) *models.V1VMObjectMeta {
	meta := &models.V1VMObjectMeta{}
	if v, ok := m["name"].(string); ok {
		meta.Name = v
	}
	if v, ok := m["namespace"].(string); ok {
		meta.Namespace = v
	}
	if v, ok := m["generate_name"].(string); ok {
		meta.GenerateName = v
	}
	if v, ok := m["uid"].(string); ok {
		meta.UID = v
	}
	if v, ok := m["resource_version"].(string); ok {
		meta.ResourceVersion = v
	}
	if v, ok := m["generation"]; ok {
		meta.Generation = int64FromInterface(v)
	}
	if v, ok := m["labels"]; ok {
		meta.Labels = stringMapFromInterface(v)
	}
	if v, ok := m["annotations"]; ok {
		meta.Annotations = stringMapFromInterface(v)
	}
	if v, ok := m["finalizers"]; ok {
		meta.Finalizers = stringSliceFromInterface(v)
	}
	return meta
}

func tfMapToVMSpecShape(m map[string]interface{}) map[string]interface{} {
	out := map[string]interface{}{}
	if v, ok := m["run_strategy"].(string); ok && v != "" {
		out["runStrategy"] = v
	}
	if v, ok := m["run_on_launch"].(bool); ok {
		out["running"] = v
	}

	ts := map[string]interface{}{}
	if v, ok := m["node_selector"]; ok {
		ts["nodeSelector"] = v
	}
	if v, ok := m["affinity"]; ok {
		ts["affinity"] = camelCaseJSONKeys(v)
	}
	if v, ok := m["scheduler_name"]; ok {
		ts["schedulerName"] = v
	}
	if v, ok := m["hostname"]; ok {
		ts["hostname"] = v
	}
	if v, ok := m["subdomain"]; ok {
		ts["subdomain"] = v
	}
	if v, ok := m["dns_policy"]; ok {
		ts["dnsPolicy"] = v
	}
	if v, ok := m["priority_class_name"]; ok {
		ts["priorityClassName"] = v
	}
	if v, ok := m["network"]; ok {
		ts["networks"] = v
	}
	if v, ok := m["volume"]; ok {
		ts["volumes"] = v
	}
	if v, ok := m["tolerations"]; ok {
		ts["tolerations"] = camelCaseJSONKeys(v)
	}
	if v, ok := m["eviction_strategy"]; ok {
		ts["evictionStrategy"] = v
	}
	if v, ok := m["termination_grace_period_seconds"]; ok {
		ts["terminationGracePeriodSeconds"] = int64FromInterface(v)
	}
	if v, ok := m["liveness_probe"]; ok {
		ts["livenessProbe"] = camelCaseJSONKeys(v)
	}
	if v, ok := m["readiness_probe"]; ok {
		ts["readinessProbe"] = camelCaseJSONKeys(v)
	}
	if v, ok := m["pod_dns_config"]; ok {
		ts["podDNSConfig"] = camelCaseJSONKeys(v)
	}

	domain := map[string]interface{}{}
	if v, ok := m["resources"]; ok {
		domain["resources"] = firstListElemOrEmpty(v)
	}
	if v, ok := m["cpu"]; ok {
		domain["cpu"] = firstListElemOrEmpty(v)
	}
	if v, ok := m["memory"]; ok {
		domain["memory"] = firstListElemOrEmpty(v)
	}
	if v, ok := m["firmware"]; ok {
		domain["firmware"] = firstListElemOrEmpty(v)
	}
	if v, ok := m["features"]; ok {
		domain["features"] = firstListElemOrEmpty(v)
	}
	dev := map[string]interface{}{}
	if v, ok := m["disk"]; ok {
		dev["disks"] = v
	}
	if v, ok := m["interface"]; ok {
		dev["interfaces"] = v
	}
	if len(dev) > 0 {
		domain["devices"] = dev
	}
	if len(domain) > 0 {
		ts["domain"] = domain
	}

	out["template"] = map[string]interface{}{
		"metadata": map[string]interface{}{},
		"spec":     camelCaseJSONKeys(ts),
	}
	if v, ok := m["data_volume_templates"]; ok {
		out["dataVolumeTemplates"] = camelCaseJSONKeys(v)
	}
	return out
}

func camelCaseJSONKeys(v interface{}) interface{} {
	switch x := v.(type) {
	case map[string]interface{}:
		out := make(map[string]interface{}, len(x))
		for k, val := range x {
			out[toCamelKey(k)] = camelCaseJSONKeys(val)
		}
		return out
	case []interface{}:
		for i := range x {
			x[i] = camelCaseJSONKeys(x[i])
		}
		return x
	default:
		return v
	}
}

func toCamelKey(snake string) string {
	if snake == "" {
		return snake
	}
	parts := strings.Split(snake, "_")
	b := strings.Builder{}
	b.WriteString(parts[0])
	for i := 1; i < len(parts); i++ {
		if parts[i] == "" {
			continue
		}
		s := parts[i]
		b.WriteString(strings.ToUpper(s[:1]) + s[1:])
	}
	return b.String()
}

func firstListElemOrEmpty(v interface{}) interface{} {
	arr, ok := v.([]interface{})
	if !ok || len(arr) == 0 || arr[0] == nil {
		return map[string]interface{}{}
	}
	if m, ok := arr[0].(map[string]interface{}); ok {
		return m
	}
	return arr[0]
}

func stringMapFromInterface(v interface{}) map[string]string {
	out := map[string]string{}
	m, ok := v.(map[string]interface{})
	if !ok {
		return out
	}
	for k, val := range m {
		if s, ok := val.(string); ok {
			out[k] = s
		} else if val != nil {
			out[k] = fmt.Sprint(val)
		}
	}
	return out
}

func stringSliceFromInterface(v interface{}) []string {
	arr, ok := v.([]interface{})
	if !ok {
		return nil
	}
	out := make([]string, 0, len(arr))
	for _, e := range arr {
		if s, ok := e.(string); ok {
			out = append(out, s)
		}
	}
	return out
}

func int64FromInterface(v interface{}) int64 {
	switch n := v.(type) {
	case float64:
		return int64(n)
	case int:
		return int64(n)
	case int64:
		return n
	default:
		return 0
	}
}
