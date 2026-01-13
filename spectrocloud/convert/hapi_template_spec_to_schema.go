package convert

import (
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/spectrocloud/terraform-provider-spectrocloud/spectrocloud/kubevirt/utils"
)

// HapiTemplateSpecToSchema converts HAPI Template Spec to Terraform schema fields
func HapiTemplateSpecToSchema(templateSpec *models.V1VMVirtualMachineInstanceSpec, d *schema.ResourceData) error {
	if templateSpec == nil {
		return nil
	}

	var err error

	// PriorityClassName
	if templateSpec.PriorityClassName != "" {
		if err = d.Set("priority_class_name", templateSpec.PriorityClassName); err != nil {
			return fmt.Errorf("failed to set priority_class_name: %w", err)
		}
	}

	// Domain - CPU, Memory, Resources, Devices, Firmware, Features
	if templateSpec.Domain != nil {
		if err = HapiDomainSpecToSchema(templateSpec.Domain, d); err != nil {
			return fmt.Errorf("failed to convert domain spec: %w", err)
		}
	}

	// NodeSelector
	if templateSpec.NodeSelector != nil && len(templateSpec.NodeSelector) > 0 {
		if err = d.Set("node_selector", utils.FlattenStringMap(templateSpec.NodeSelector)); err != nil {
			return fmt.Errorf("failed to set node_selector: %w", err)
		}
	}

	// Affinity
	if templateSpec.Affinity != nil {
		affinity, setErr := HapiAffinityToSchema(templateSpec.Affinity)
		if setErr != nil {
			return fmt.Errorf("failed to convert affinity: %w", setErr)
		}
		if affinity != nil {
			if setErr = d.Set("affinity", affinity); setErr != nil {
				return fmt.Errorf("failed to set affinity: %w", setErr)
			}
		}
	}

	// SchedulerName
	if templateSpec.SchedulerName != "" {
		if err = d.Set("scheduler_name", templateSpec.SchedulerName); err != nil {
			return fmt.Errorf("failed to set scheduler_name: %w", err)
		}
	}

	// Tolerations
	if templateSpec.Tolerations != nil && len(templateSpec.Tolerations) > 0 {
		tolerations, err := HapiTolerationsToSchema(templateSpec.Tolerations)
		if err != nil {
			return fmt.Errorf("failed to convert tolerations: %w", err)
		}
		if err = d.Set("tolerations", tolerations); err != nil {
			return fmt.Errorf("failed to set tolerations: %w", err)
		}
	}

	// EvictionStrategy
	if templateSpec.EvictionStrategy != "" {
		if err = d.Set("eviction_strategy", templateSpec.EvictionStrategy); err != nil {
			return fmt.Errorf("failed to set eviction_strategy: %w", err)
		}
	}

	// TerminationGracePeriodSeconds
	if templateSpec.TerminationGracePeriodSeconds != 0 {
		if err = d.Set("termination_grace_period_seconds", int(templateSpec.TerminationGracePeriodSeconds)); err != nil {
			return fmt.Errorf("failed to set termination_grace_period_seconds: %w", err)
		}
	}

	// Volumes
	if templateSpec.Volumes != nil && len(templateSpec.Volumes) > 0 {
		volumes, err := HapiVolumesToSchema(templateSpec.Volumes)
		if err != nil {
			return fmt.Errorf("failed to convert volumes: %w", err)
		}
		if err = d.Set("volume", volumes); err != nil {
			return fmt.Errorf("failed to set volume: %w", err)
		}
	}

	// LivenessProbe
	if templateSpec.LivenessProbe != nil {
		probe, err := HapiProbeToSchema(templateSpec.LivenessProbe)
		if err != nil {
			return fmt.Errorf("failed to convert liveness probe: %w", err)
		}
		if err = d.Set("liveness_probe", probe); err != nil {
			return fmt.Errorf("failed to set liveness_probe: %w", err)
		}
	}

	// ReadinessProbe
	if templateSpec.ReadinessProbe != nil {
		probe, err := HapiProbeToSchema(templateSpec.ReadinessProbe)
		if err != nil {
			return fmt.Errorf("failed to convert readiness probe: %w", err)
		}
		if err = d.Set("readiness_probe", probe); err != nil {
			return fmt.Errorf("failed to set readiness_probe: %w", err)
		}
	}

	// Hostname
	if templateSpec.Hostname != "" {
		if err = d.Set("hostname", templateSpec.Hostname); err != nil {
			return fmt.Errorf("failed to set hostname: %w", err)
		}
	}

	// Subdomain
	if templateSpec.Subdomain != "" {
		if err = d.Set("subdomain", templateSpec.Subdomain); err != nil {
			return fmt.Errorf("failed to set subdomain: %w", err)
		}
	}

	// Networks
	if templateSpec.Networks != nil && len(templateSpec.Networks) > 0 {
		networks, err := HapiNetworksToSchema(templateSpec.Networks)
		if err != nil {
			return fmt.Errorf("failed to convert networks: %w", err)
		}
		if err = d.Set("network", networks); err != nil {
			return fmt.Errorf("failed to set network: %w", err)
		}
	}

	// DNSPolicy
	if templateSpec.DNSPolicy != "" {
		if err = d.Set("dns_policy", templateSpec.DNSPolicy); err != nil {
			return fmt.Errorf("failed to set dns_policy: %w", err)
		}
	}

	// DNSConfig
	if templateSpec.DNSConfig != nil {
		dnsConfig, err := HapiDNSConfigToSchema(templateSpec.DNSConfig)
		if err != nil {
			return fmt.Errorf("failed to convert DNS config: %w", err)
		}
		if err = d.Set("pod_dns_config", dnsConfig); err != nil {
			return fmt.Errorf("failed to set pod_dns_config: %w", err)
		}
	}

	return nil
}

// HapiDomainSpecToSchema converts HAPI Domain Spec to Terraform schema
func HapiDomainSpecToSchema(domain *models.V1VMDomainSpec, d *schema.ResourceData) error {
	if domain == nil {
		return nil
	}

	// Resources
	if domain.Resources != nil {
		resources, err := HapiResourcesToSchema(domain.Resources)
		if err != nil {
			return fmt.Errorf("failed to convert resources: %w", err)
		}
		if err = d.Set("resources", resources); err != nil {
			return fmt.Errorf("failed to set resources: %w", err)
		}
	}

	// Devices (Disks and Interfaces)
	if domain.Devices != nil {
		// Disks
		if domain.Devices.Disks != nil && len(domain.Devices.Disks) > 0 {
			disks, err := HapiDisksToSchema(domain.Devices.Disks)
			if err != nil {
				return fmt.Errorf("failed to convert disks: %w", err)
			}
			if err = d.Set("disk", disks); err != nil {
				return fmt.Errorf("failed to set disk: %w", err)
			}
		}

		// Interfaces
		if domain.Devices.Interfaces != nil && len(domain.Devices.Interfaces) > 0 {
			interfaces, err := HapiInterfacesToSchema(domain.Devices.Interfaces)
			if err != nil {
				return fmt.Errorf("failed to convert interfaces: %w", err)
			}
			if err = d.Set("interface", interfaces); err != nil {
				return fmt.Errorf("failed to set interface: %w", err)
			}
		}
	}

	// CPU
	if domain.CPU != nil {
		cpu, err := HapiCPUToSchema(domain.CPU)
		if err != nil {
			return fmt.Errorf("failed to convert CPU: %w", err)
		}
		if err = d.Set("cpu", cpu); err != nil {
			return fmt.Errorf("failed to set cpu: %w", err)
		}
	}

	// Memory
	if domain.Memory != nil {
		memory, err := HapiMemoryToSchema(domain.Memory)
		if err != nil {
			return fmt.Errorf("failed to convert memory: %w", err)
		}
		if err = d.Set("memory", memory); err != nil {
			return fmt.Errorf("failed to set memory: %w", err)
		}
	}

	// Firmware
	if domain.Firmware != nil {
		firmware, err := HapiFirmwareToSchema(domain.Firmware)
		if err != nil {
			return fmt.Errorf("failed to convert firmware: %w", err)
		}
		if err = d.Set("firmware", firmware); err != nil {
			return fmt.Errorf("failed to set firmware: %w", err)
		}
	}

	// Features
	if domain.Features != nil {
		features, err := HapiFeaturesToSchema(domain.Features)
		if err != nil {
			return fmt.Errorf("failed to convert features: %w", err)
		}
		if err = d.Set("features", features); err != nil {
			return fmt.Errorf("failed to set features: %w", err)
		}
	}

	return nil
}

// HapiResourcesToSchema converts HAPI ResourceRequirements to Terraform schema
func HapiResourcesToSchema(resources *models.V1VMResourceRequirements) ([]interface{}, error) {
	if resources == nil {
		return nil, nil
	}

	result := make(map[string]interface{})

	// Requests - handle as map[string]string or interface{}
	if resources.Requests != nil {
		requestsMap := make(map[string]interface{})
		switch v := resources.Requests.(type) {
		case map[string]interface{}:
			for k, val := range v {
				requestsMap[k] = val
			}
		case map[string]string:
			for k, val := range v {
				requestsMap[k] = val
			}
		default:
			// Try to convert via JSON
			requestsJSON, err := json.Marshal(resources.Requests)
			if err == nil {
				var requestsMapJSON map[string]interface{}
				if err := json.Unmarshal(requestsJSON, &requestsMapJSON); err == nil {
					requestsMap = requestsMapJSON
				}
			}
		}
		if len(requestsMap) > 0 {
			result["requests"] = requestsMap
		}
	}

	// Limits - handle as map[string]string or interface{}
	if resources.Limits != nil {
		limitsMap := make(map[string]interface{})
		switch v := resources.Limits.(type) {
		case map[string]interface{}:
			for k, val := range v {
				limitsMap[k] = val
			}
		case map[string]string:
			for k, val := range v {
				limitsMap[k] = val
			}
		default:
			// Try to convert via JSON
			limitsJSON, err := json.Marshal(resources.Limits)
			if err == nil {
				var limitsMapJSON map[string]interface{}
				if err := json.Unmarshal(limitsJSON, &limitsMapJSON); err == nil {
					limitsMap = limitsMapJSON
				}
			}
		}
		if len(limitsMap) > 0 {
			result["limits"] = limitsMap
		}
	}

	// OvercommitGuestOverhead
	result["over_commit_guest_overhead"] = resources.OvercommitGuestOverhead

	return []interface{}{result}, nil
}

// HapiCPUToSchema converts HAPI CPU to Terraform schema
func HapiCPUToSchema(cpu *models.V1VMCPU) ([]interface{}, error) {
	if cpu == nil {
		return nil, nil
	}

	result := make(map[string]interface{})

	if cpu.Cores != 0 {
		result["cores"] = int(cpu.Cores)
	}

	if cpu.Sockets != 0 {
		result["sockets"] = int(cpu.Sockets)
	}

	if cpu.Threads != 0 {
		result["threads"] = int(cpu.Threads)
	}

	return []interface{}{result}, nil
}

// HapiMemoryToSchema converts HAPI Memory to Terraform schema
func HapiMemoryToSchema(memory *models.V1VMMemory) ([]interface{}, error) {
	if memory == nil {
		return nil, nil
	}

	result := make(map[string]interface{})

	if memory.Guest != "" {
		result["guest"] = string(memory.Guest)
	}

	if memory.Hugepages != nil && memory.Hugepages.PageSize != "" {
		result["hugepages"] = memory.Hugepages.PageSize
	}

	return []interface{}{result}, nil
}

// HapiFirmwareToSchema converts HAPI Firmware to Terraform schema
func HapiFirmwareToSchema(firmware *models.V1VMFirmware) ([]interface{}, error) {
	if firmware == nil {
		return nil, nil
	}

	result := make(map[string]interface{})

	if firmware.UUID != "" {
		result["uuid"] = firmware.UUID
	}

	if firmware.Serial != "" {
		result["serial"] = firmware.Serial
	}

	if firmware.Bootloader != nil {
		bootloader, err := HapiBootloaderToSchema(firmware.Bootloader)
		if err != nil {
			return nil, fmt.Errorf("failed to convert bootloader: %w", err)
		}
		result["bootloader"] = bootloader
	}

	return []interface{}{result}, nil
}

// HapiBootloaderToSchema converts HAPI Bootloader to Terraform schema
func HapiBootloaderToSchema(bootloader *models.V1VMBootloader) ([]interface{}, error) {
	if bootloader == nil {
		return nil, nil
	}

	result := make(map[string]interface{})

	if bootloader.Bios != nil {
		bios, err := HapiBIOSToSchema(bootloader.Bios)
		if err != nil {
			return nil, fmt.Errorf("failed to convert BIOS: %w", err)
		}
		result["bios"] = bios
	}

	if bootloader.Efi != nil {
		efi, err := HapiEFIToSchema(bootloader.Efi)
		if err != nil {
			return nil, fmt.Errorf("failed to convert EFI: %w", err)
		}
		result["efi"] = efi
	}

	return []interface{}{result}, nil
}

// HapiBIOSToSchema converts HAPI BIOS to Terraform schema
func HapiBIOSToSchema(bios *models.V1VMBIOS) ([]interface{}, error) {
	if bios == nil {
		return []interface{}{map[string]interface{}{}}, nil
	}

	result := make(map[string]interface{})

	if bios.UseSerial {
		result["use_serial"] = bios.UseSerial
	}

	return []interface{}{result}, nil
}

// HapiEFIToSchema converts HAPI EFI to Terraform schema
func HapiEFIToSchema(efi *models.V1VMEFI) ([]interface{}, error) {
	if efi == nil {
		return []interface{}{map[string]interface{}{}}, nil
	}

	result := make(map[string]interface{})

	if efi.SecureBoot {
		result["secure_boot"] = efi.SecureBoot
	}

	if efi.Persistent {
		result["persistent"] = efi.Persistent
	}

	return []interface{}{result}, nil
}

// HapiFeaturesToSchema converts HAPI Features to Terraform schema
func HapiFeaturesToSchema(features *models.V1VMFeatures) ([]interface{}, error) {
	if features == nil {
		return nil, nil
	}

	result := make(map[string]interface{})

	// ACPI
	if features.Acpi != nil {
		acpi, err := HapiFeatureStateToSchema(features.Acpi)
		if err != nil {
			return nil, fmt.Errorf("failed to convert ACPI: %w", err)
		}
		result["acpi"] = acpi
	}

	// APIC
	if features.Apic != nil {
		apicFeatureState := &models.V1VMFeatureState{
			Enabled: features.Apic.Enabled,
		}
		apic, err := HapiFeatureStateToSchema(apicFeatureState)
		if err != nil {
			return nil, fmt.Errorf("failed to convert APIC: %w", err)
		}
		result["apic"] = apic
	}

	// SMM
	if features.Smm != nil {
		smm, err := HapiFeatureStateToSchema(features.Smm)
		if err != nil {
			return nil, fmt.Errorf("failed to convert SMM: %w", err)
		}
		result["smm"] = smm
	}

	return []interface{}{result}, nil
}

// HapiFeatureStateToSchema converts HAPI FeatureState to Terraform schema
func HapiFeatureStateToSchema(featureState *models.V1VMFeatureState) ([]interface{}, error) {
	if featureState == nil {
		return []interface{}{map[string]interface{}{}}, nil
	}

	result := make(map[string]interface{})

	result["enabled"] = featureState.Enabled

	return []interface{}{result}, nil
}

// HapiDisksToSchema converts HAPI Disks to Terraform schema
// Uses JSON marshaling as bridge since structures are compatible
func HapiDisksToSchema(disks []*models.V1VMDisk) ([]interface{}, error) {
	if len(disks) == 0 {
		return nil, nil
	}

	// Marshal HAPI disks to JSON and convert to Terraform schema format
	jsonBytes, err := json.Marshal(disks)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal disks to JSON: %w", err)
	}

	var disksJSON []map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &disksJSON); err != nil {
		return nil, fmt.Errorf("failed to unmarshal disks JSON: %w", err)
	}

	// Convert JSON structure to Terraform schema format
	result := make([]interface{}, len(disksJSON))
	for i, diskJSON := range disksJSON {
		diskMap := make(map[string]interface{})

		if name, ok := diskJSON["name"].(string); ok {
			diskMap["name"] = name
		}

		if serial, ok := diskJSON["serial"].(string); ok {
			diskMap["serial"] = serial
		}

		if bootOrder, ok := diskJSON["bootOrder"].(float64); ok {
			diskMap["boot_order"] = int(bootOrder)
		}

		// DiskDevice
		if diskDeviceJSON, ok := diskJSON["diskDevice"].(map[string]interface{}); ok {
			diskDeviceMap := make(map[string]interface{})

			if diskTargetJSON, ok := diskDeviceJSON["disk"].(map[string]interface{}); ok {
				diskTargetMap := make(map[string]interface{})

				if bus, ok := diskTargetJSON["bus"].(string); ok {
					diskTargetMap["bus"] = bus
				}

				if readOnly, ok := diskTargetJSON["readOnly"].(bool); ok {
					diskTargetMap["read_only"] = readOnly
				}

				if pciAddress, ok := diskTargetJSON["pciAddress"].(string); ok {
					diskTargetMap["pci_address"] = pciAddress
				}

				diskDeviceMap["disk"] = []interface{}{diskTargetMap}
			}

			diskMap["disk_device"] = []interface{}{diskDeviceMap}
		}

		result[i] = diskMap
	}

	return result, nil
}

// HapiInterfacesToSchema converts HAPI Interfaces to Terraform schema
// Uses JSON marshaling as bridge since structures are compatible
func HapiInterfacesToSchema(interfaces []*models.V1VMInterface) ([]interface{}, error) {
	if len(interfaces) == 0 {
		return nil, nil
	}

	// Marshal HAPI interfaces to JSON and convert to Terraform schema format
	jsonBytes, err := json.Marshal(interfaces)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal interfaces to JSON: %w", err)
	}

	var interfacesJSON []map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &interfacesJSON); err != nil {
		return nil, fmt.Errorf("failed to unmarshal interfaces JSON: %w", err)
	}

	// Convert JSON structure to Terraform schema format
	result := make([]interface{}, len(interfacesJSON))
	for i, ifaceJSON := range interfacesJSON {
		ifaceMap := make(map[string]interface{})

		if name, ok := ifaceJSON["name"].(string); ok {
			ifaceMap["name"] = name
		}

		if model, ok := ifaceJSON["model"].(string); ok {
			ifaceMap["model"] = model
		}

		// InterfaceBindingMethod
		if bindingMethodJSON, ok := ifaceJSON["interfaceBindingMethod"].(map[string]interface{}); ok {
			var bindingMethod string
			if _, ok := bindingMethodJSON["bridge"]; ok {
				bindingMethod = "InterfaceBridge"
			} else if _, ok := bindingMethodJSON["deprecatedSlirp"]; ok {
				bindingMethod = "InterfaceSlirp"
			} else if _, ok := bindingMethodJSON["masquerade"]; ok {
				bindingMethod = "InterfaceMasquerade"
			} else if _, ok := bindingMethodJSON["sriov"]; ok {
				bindingMethod = "InterfaceSRIOV"
			}
			ifaceMap["interface_binding_method"] = bindingMethod
		}

		result[i] = ifaceMap
	}

	return result, nil
}

// HapiVolumesToSchema converts HAPI Volumes to Terraform schema
// Uses JSON marshaling as bridge since structures are compatible
func HapiVolumesToSchema(volumes []*models.V1VMVolume) ([]interface{}, error) {
	if len(volumes) == 0 {
		return nil, nil
	}

	// Marshal HAPI volumes to JSON and convert to Terraform schema format
	jsonBytes, err := json.Marshal(volumes)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal volumes to JSON: %w", err)
	}

	var volumesJSON []map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &volumesJSON); err != nil {
		return nil, fmt.Errorf("failed to unmarshal volumes JSON: %w", err)
	}

	// Convert JSON structure to Terraform schema format
	result := make([]interface{}, len(volumesJSON))
	for i, volumeJSON := range volumesJSON {
		volumeMap := make(map[string]interface{})

		if name, ok := volumeJSON["name"].(string); ok {
			volumeMap["name"] = name
		}

		// VolumeSource - convert from JSON structure
		if volumeSourceJSON, ok := volumeJSON["volumeSource"].(map[string]interface{}); ok {
			volumeSourceMap := make(map[string]interface{})

			// Handle all volume source types
			if dataVolume, ok := volumeSourceJSON["dataVolume"].(map[string]interface{}); ok {
				volumeSourceMap["data_volume"] = []interface{}{dataVolume}
			}
			if cloudInitConfigDrive, ok := volumeSourceJSON["cloudInitConfigDrive"].(map[string]interface{}); ok {
				volumeSourceMap["cloud_init_config_drive"] = []interface{}{cloudInitConfigDrive}
			}
			if serviceAccount, ok := volumeSourceJSON["serviceAccount"].(map[string]interface{}); ok {
				volumeSourceMap["service_account"] = []interface{}{serviceAccount}
			}
			if containerDisk, ok := volumeSourceJSON["containerDisk"].(map[string]interface{}); ok {
				volumeSourceMap["container_disk"] = []interface{}{containerDisk}
			}
			if cloudInitNoCloud, ok := volumeSourceJSON["cloudInitNoCloud"].(map[string]interface{}); ok {
				volumeSourceMap["cloud_init_no_cloud"] = []interface{}{cloudInitNoCloud}
			}
			if ephemeral, ok := volumeSourceJSON["ephemeral"].(map[string]interface{}); ok {
				volumeSourceMap["ephemeral"] = []interface{}{ephemeral}
			}
			if emptyDisk, ok := volumeSourceJSON["emptyDisk"].(map[string]interface{}); ok {
				volumeSourceMap["empty_disk"] = []interface{}{emptyDisk}
			}
			if persistentVolumeClaim, ok := volumeSourceJSON["persistentVolumeClaim"].(map[string]interface{}); ok {
				volumeSourceMap["persistent_volume_claim"] = []interface{}{persistentVolumeClaim}
			}
			if hostDisk, ok := volumeSourceJSON["hostDisk"].(map[string]interface{}); ok {
				volumeSourceMap["host_disk"] = []interface{}{hostDisk}
			}
			if configMap, ok := volumeSourceJSON["configMap"].(map[string]interface{}); ok {
				volumeSourceMap["config_map"] = []interface{}{configMap}
			}

			volumeMap["volume_source"] = []interface{}{volumeSourceMap}
		}

		result[i] = volumeMap
	}

	return result, nil
}

// HapiNetworksToSchema converts HAPI Networks to Terraform schema
// Uses JSON marshaling as bridge since structures are compatible
func HapiNetworksToSchema(networks []*models.V1VMNetwork) ([]interface{}, error) {
	if len(networks) == 0 {
		return nil, nil
	}

	// Marshal HAPI networks to JSON and convert to Terraform schema format
	jsonBytes, err := json.Marshal(networks)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal networks to JSON: %w", err)
	}

	var networksJSON []map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &networksJSON); err != nil {
		return nil, fmt.Errorf("failed to unmarshal networks JSON: %w", err)
	}

	// Convert JSON structure to Terraform schema format
	result := make([]interface{}, len(networksJSON))
	for i, networkJSON := range networksJSON {
		networkMap := make(map[string]interface{})

		if name, ok := networkJSON["name"].(string); ok {
			networkMap["name"] = name
		}

		// NetworkSource
		if networkSourceJSON, ok := networkJSON["networkSource"].(map[string]interface{}); ok {
			networkSourceMap := make(map[string]interface{})

			if pod, ok := networkSourceJSON["pod"].(map[string]interface{}); ok {
				podMap := make(map[string]interface{})
				if vmCIDR, ok := pod["vmNetworkCIDR"].(string); ok {
					podMap["vm_network_cidr"] = vmCIDR
				}
				networkSourceMap["pod"] = []interface{}{podMap}
			}

			if multus, ok := networkSourceJSON["multus"].(map[string]interface{}); ok {
				multusMap := make(map[string]interface{})
				if networkName, ok := multus["networkName"].(string); ok {
					multusMap["network_name"] = networkName
				}
				if defaultNet, ok := multus["default"].(bool); ok {
					multusMap["default"] = defaultNet
				}
				networkSourceMap["multus"] = []interface{}{multusMap}
			}

			networkMap["network_source"] = []interface{}{networkSourceMap}
		}

		result[i] = networkMap
	}

	return result, nil
}

// HapiAffinityToSchema converts HAPI Affinity to Terraform schema
// Uses JSON marshaling as bridge via k8s types
func HapiAffinityToSchema(affinity *models.V1VMAffinity) ([]interface{}, error) {
	if affinity == nil {
		return nil, nil
	}

	// Marshal HAPI affinity to JSON and unmarshal to k8s types, then flatten
	jsonBytes, err := json.Marshal(affinity)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal affinity to JSON: %w", err)
	}

	// Use k8s.FlattenAffinity which expects k8s types
	// We'll unmarshal to a generic map first, then use k8s flattening
	var k8sAffinity map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &k8sAffinity); err != nil {
		return nil, fmt.Errorf("failed to unmarshal affinity JSON: %w", err)
	}

	// Return the JSON structure directly converted to schema format
	// This is a simplified approach - full implementation would convert to k8s types first
	return []interface{}{k8sAffinity}, nil
}

// HapiTolerationsToSchema converts HAPI Tolerations to Terraform schema
// Uses JSON marshaling as bridge via k8s types
func HapiTolerationsToSchema(tolerations []*models.V1VMToleration) ([]interface{}, error) {
	if len(tolerations) == 0 {
		return nil, nil
	}

	// Marshal HAPI tolerations to JSON
	jsonBytes, err := json.Marshal(tolerations)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal tolerations to JSON: %w", err)
	}

	// Convert to k8s types for flattening
	// Use k8s.FlattenTolerations which expects k8s types
	// For now, return JSON structure directly
	var tolerationsJSON []map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &tolerationsJSON); err != nil {
		return nil, fmt.Errorf("failed to unmarshal tolerations JSON: %w", err)
	}

	result := make([]interface{}, len(tolerationsJSON))
	for i, tolJSON := range tolerationsJSON {
		result[i] = tolJSON
	}

	return result, nil
}

// HapiProbeToSchema converts HAPI Probe to Terraform schema
func HapiProbeToSchema(probe *models.V1VMProbe) ([]interface{}, error) {
	if probe == nil {
		return nil, nil
	}

	// Marshal HAPI probe to JSON and convert to Terraform schema format
	jsonBytes, err := json.Marshal(probe)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal probe to JSON: %w", err)
	}

	var probeJSON map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &probeJSON); err != nil {
		return nil, fmt.Errorf("failed to unmarshal probe JSON: %w", err)
	}

	// Convert JSON keys to Terraform schema format
	probeMap := make(map[string]interface{})

	if initialDelaySeconds, ok := probeJSON["initialDelaySeconds"].(float64); ok {
		probeMap["initial_delay_seconds"] = int(initialDelaySeconds)
	}
	if timeoutSeconds, ok := probeJSON["timeoutSeconds"].(float64); ok {
		probeMap["timeout_seconds"] = int(timeoutSeconds)
	}
	if periodSeconds, ok := probeJSON["periodSeconds"].(float64); ok {
		probeMap["period_seconds"] = int(periodSeconds)
	}
	if successThreshold, ok := probeJSON["successThreshold"].(float64); ok {
		probeMap["success_threshold"] = int(successThreshold)
	}
	if failureThreshold, ok := probeJSON["failureThreshold"].(float64); ok {
		probeMap["failure_threshold"] = int(failureThreshold)
	}

	// Handler
	if httpGet, ok := probeJSON["httpGet"].(map[string]interface{}); ok {
		httpGetMap := make(map[string]interface{})
		if path, ok := httpGet["path"].(string); ok {
			httpGetMap["path"] = path
		}
		if port, ok := httpGet["port"].(float64); ok {
			httpGetMap["port"] = int(port)
		}
		if scheme, ok := httpGet["scheme"].(string); ok {
			httpGetMap["scheme"] = scheme
		}
		probeMap["http_get"] = []interface{}{httpGetMap}
	}

	if tcpSocket, ok := probeJSON["tcpSocket"].(map[string]interface{}); ok {
		tcpSocketMap := make(map[string]interface{})
		if port, ok := tcpSocket["port"].(float64); ok {
			tcpSocketMap["port"] = int(port)
		}
		probeMap["tcp_socket"] = []interface{}{tcpSocketMap}
	}

	if exec, ok := probeJSON["exec"].(map[string]interface{}); ok {
		execMap := make(map[string]interface{})
		if command, ok := exec["command"].([]interface{}); ok {
			execMap["command"] = command
		}
		probeMap["exec"] = []interface{}{execMap}
	}

	return []interface{}{probeMap}, nil
}

// HapiDNSConfigToSchema converts HAPI PodDNSConfig to Terraform schema
// Uses JSON marshaling as bridge via k8s types
func HapiDNSConfigToSchema(dnsConfig *models.V1VMPodDNSConfig) ([]interface{}, error) {
	if dnsConfig == nil {
		return nil, nil
	}

	// Marshal HAPI DNS config to JSON
	jsonBytes, err := json.Marshal(dnsConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal DNS config to JSON: %w", err)
	}

	// Convert to k8s types for flattening
	// Use k8s.FlattenPodDNSConfig which expects k8s types
	// For now, return JSON structure directly
	var dnsConfigJSON map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &dnsConfigJSON); err != nil {
		return nil, fmt.Errorf("failed to unmarshal DNS config JSON: %w", err)
	}

	return []interface{}{dnsConfigJSON}, nil
}
