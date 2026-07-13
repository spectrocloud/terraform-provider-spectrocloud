package spectrocloud

import (
	"bytes"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-sdk-go/api/models"
)

const overrideHealthCheckConfigurationRepaveWarning = "Changing the machine health check configuration may repave your nodes."

func expandOverrideHealthCheckConfiguration(m map[string]interface{}, poolConfig *models.V1MachinePoolConfigEntity) {
	if poolConfig == nil {
		return
	}
	if override, ok := m["override_health_check_configuration"].(string); ok && override != "" {
		poolConfig.OverrideHealthCheckConfiguration = override
	}
}

func flattenOverrideHealthCheckConfiguration(overrideHealthCheck string, oi map[string]interface{}) {
	if overrideHealthCheck != "" {
		oi["override_health_check_configuration"] = overrideHealthCheck
	}
}

func writeOverrideHealthCheckConfigurationHash(buf *bytes.Buffer, m map[string]interface{}) {
	if val, ok := m["override_health_check_configuration"].(string); ok && val != "" {
		fmt.Fprintf(buf, "%s-", val)
	}
}

func appendOverrideHealthCheckConfigurationCreateWarnings(d *schema.ResourceData, diags *diag.Diagnostics) {
	raw := d.Get("machine_pool")
	set, ok := raw.(*schema.Set)
	if !ok || set == nil {
		return
	}
	for _, item := range set.List() {
		m, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		if val, ok := m["override_health_check_configuration"].(string); ok && val != "" {
			name, _ := m["name"].(string)
			appendHealthCheckRepaveWarning(diags, name)
		}
	}
}

func appendOverrideHealthCheckConfigurationUpdateWarnings(d *schema.ResourceData, diags *diag.Diagnostics) {
	if !d.HasChange("machine_pool") {
		return
	}
	oldRaw, newRaw := d.GetChange("machine_pool")
	oldSet, ok := oldRaw.(*schema.Set)
	if !ok || oldSet == nil {
		oldSet = schema.NewSet(schema.HashString, nil)
	}
	newSet, ok := newRaw.(*schema.Set)
	if !ok || newSet == nil {
		return
	}

	oldByName := make(map[string]string)
	for _, item := range oldSet.List() {
		m, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		name, _ := m["name"].(string)
		oldByName[name], _ = m["override_health_check_configuration"].(string)
	}

	for _, item := range newSet.List() {
		m, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		name, _ := m["name"].(string)
		newVal, _ := m["override_health_check_configuration"].(string)
		oldVal := oldByName[name]
		if newVal != oldVal {
			appendHealthCheckRepaveWarning(diags, name)
		}
	}
}

func appendHealthCheckRepaveWarning(diags *diag.Diagnostics, poolName string) {
	detail := overrideHealthCheckConfigurationRepaveWarning
	if poolName != "" {
		detail = fmt.Sprintf("Machine pool %q: %s", poolName, overrideHealthCheckConfigurationRepaveWarning)
	}
	*diags = append(*diags, diag.Diagnostic{
		Severity: diag.Warning,
		Summary:  "Warning",
		Detail:   detail,
	})
}
