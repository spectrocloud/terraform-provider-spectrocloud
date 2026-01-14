package convert

import (
	"encoding/json"
	"fmt"
	"math"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/spectrocloud/terraform-provider-spectrocloud/spectrocloud/kubevirt/schema/k8s"
	"github.com/spectrocloud/terraform-provider-spectrocloud/spectrocloud/kubevirt/utils"
)

// SchemaToHapiTemplateSpec converts Terraform schema template spec fields to HAPI VM Instance Spec
// This is a comprehensive conversion covering all template spec fields
func SchemaToHapiTemplateSpec(d *schema.ResourceData) (*models.V1VMVirtualMachineInstanceSpec, error) {
	spec := &models.V1VMVirtualMachineInstanceSpec{}

	// PriorityClassName
	if v, ok := d.GetOk("priority_class_name"); ok {
		spec.PriorityClassName = v.(string)
	}

	// Domain - CPU, Memory, Resources, Devices (Disks, Interfaces), Firmware, Features
	domain, err := SchemaToHapiDomainSpec(d)
	if err != nil {
		return nil, fmt.Errorf("failed to convert domain spec: %w", err)
	}
	spec.Domain = domain

	// NodeSelector
	if v, ok := d.GetOk("node_selector"); ok && len(v.(map[string]interface{})) > 0 {
		spec.NodeSelector = utils.ExpandStringMap(v.(map[string]interface{}))
	}

	// Affinity
	if v, ok := d.GetOk("affinity"); ok {
		affinity, err := SchemaToHapiAffinity(v.([]interface{}))
		if err != nil {
			return nil, fmt.Errorf("failed to convert affinity: %w", err)
		}
		spec.Affinity = affinity
	}

	// SchedulerName
	if v, ok := d.GetOk("scheduler_name"); ok {
		spec.SchedulerName = v.(string)
	}

	// Tolerations
	if v, ok := d.GetOk("tolerations"); ok {
		tolerations, err := SchemaToHapiTolerations(v.([]interface{}))
		if err != nil {
			return nil, fmt.Errorf("failed to convert tolerations: %w", err)
		}
		spec.Tolerations = tolerations
	}

	// EvictionStrategy
	if v, ok := d.GetOk("eviction_strategy"); ok {
		if v.(string) != "" {
			spec.EvictionStrategy = v.(string)
		}
	}

	// TerminationGracePeriodSeconds
	if v, ok := d.GetOk("termination_grace_period_seconds"); ok {
		spec.TerminationGracePeriodSeconds = int64(v.(int))
	}

	// Volumes
	if v, ok := d.GetOk("volume"); ok {
		volumes, err := SchemaToHapiVolumes(v.([]interface{}))
		if err != nil {
			return nil, fmt.Errorf("failed to convert volumes: %w", err)
		}
		spec.Volumes = volumes
	}

	// LivenessProbe
	if v, ok := d.GetOk("liveness_probe"); ok {
		probe, err := SchemaToHapiProbe(v.([]interface{}))
		if err != nil {
			return nil, fmt.Errorf("failed to convert liveness probe: %w", err)
		}
		spec.LivenessProbe = probe
	}

	// ReadinessProbe
	if v, ok := d.GetOk("readiness_probe"); ok {
		probe, err := SchemaToHapiProbe(v.([]interface{}))
		if err != nil {
			return nil, fmt.Errorf("failed to convert readiness probe: %w", err)
		}
		spec.ReadinessProbe = probe
	}

	// Hostname
	if v, ok := d.GetOk("hostname"); ok {
		spec.Hostname = v.(string)
	}

	// Subdomain
	if v, ok := d.GetOk("subdomain"); ok {
		spec.Subdomain = v.(string)
	}

	// Networks
	if v, ok := d.GetOk("network"); ok {
		networks, err := SchemaToHapiNetworks(v.([]interface{}))
		if err != nil {
			return nil, fmt.Errorf("failed to convert networks: %w", err)
		}
		spec.Networks = networks
	}

	// DNSPolicy
	if v, ok := d.GetOk("dns_policy"); ok {
		spec.DNSPolicy = v.(string)
	}

	// DNSConfig
	if v, ok := d.GetOk("pod_dns_config"); ok {
		dnsConfig, err := SchemaToHapiDNSConfig(v.([]interface{}))
		if err != nil {
			return nil, fmt.Errorf("failed to convert DNS config: %w", err)
		}
		spec.DNSConfig = dnsConfig
	}

	return spec, nil
}

// SchemaToHapiDomainSpec converts Terraform schema domain fields to HAPI Domain Spec
func SchemaToHapiDomainSpec(d *schema.ResourceData) (*models.V1VMDomainSpec, error) {
	domain := &models.V1VMDomainSpec{}

	// Resources
	if v, ok := d.GetOk("resources"); ok {
		resources, err := SchemaToHapiResources(v.([]interface{}))
		if err != nil {
			return nil, fmt.Errorf("failed to convert resources: %w", err)
		}
		domain.Resources = resources
	}

	// Devices (Disks and Interfaces)
	devices, err := SchemaToHapiDevices(d)
	if err != nil {
		return nil, fmt.Errorf("failed to convert devices: %w", err)
	}
	domain.Devices = devices

	// CPU
	if v, ok := d.GetOk("cpu"); ok {
		cpuList := v.([]interface{})
		if len(cpuList) > 0 && cpuList[0] != nil {
			cpu, err := SchemaToHapiCPU(cpuList[0].(map[string]interface{}))
			if err != nil {
				return nil, fmt.Errorf("failed to convert CPU: %w", err)
			}
			domain.CPU = cpu
		}
	}

	// Memory
	if v, ok := d.GetOk("memory"); ok {
		memory, err := SchemaToHapiMemory(v.([]interface{}))
		if err != nil {
			return nil, fmt.Errorf("failed to convert memory: %w", err)
		}
		domain.Memory = memory
	}

	// Firmware
	if v, ok := d.GetOk("firmware"); ok {
		firmware, err := SchemaToHapiFirmware(v.([]interface{}))
		if err != nil {
			return nil, fmt.Errorf("failed to convert firmware: %w", err)
		}
		domain.Firmware = firmware
	}

	// Features
	if v, ok := d.GetOk("features"); ok {
		features, err := SchemaToHapiFeatures(v.([]interface{}))
		if err != nil {
			return nil, fmt.Errorf("failed to convert features: %w", err)
		}
		domain.Features = features
	}

	return domain, nil
}

// SchemaToHapiResources converts Terraform schema resources to HAPI ResourceRequirements
func SchemaToHapiResources(resources []interface{}) (*models.V1VMResourceRequirements, error) {
	if len(resources) == 0 || resources[0] == nil {
		return nil, nil
	}

	result := &models.V1VMResourceRequirements{}
	in := resources[0].(map[string]interface{})

	// Requests
	if v, ok := in["requests"].(map[string]interface{}); ok && len(v) > 0 {
		requests, err := utils.ExpandMapToResourceList(v)
		if err != nil {
			return nil, fmt.Errorf("failed to expand requests: %w", err)
		}
		// Convert k8s ResourceList to HAPI format
		requestsMap := make(map[string]string)
		for k, q := range *requests {
			requestsMap[string(k)] = q.String()
		}
		result.Requests = requestsMap
	}

	// Limits
	if v, ok := in["limits"].(map[string]interface{}); ok && len(v) > 0 {
		limits, err := utils.ExpandMapToResourceList(v)
		if err != nil {
			return nil, fmt.Errorf("failed to expand limits: %w", err)
		}
		// Convert k8s ResourceList to HAPI format
		limitsMap := make(map[string]string)
		for k, q := range *limits {
			limitsMap[string(k)] = q.String()
		}
		result.Limits = limitsMap
	}

	// OvercommitGuestOverhead
	if v, ok := in["over_commit_guest_overhead"].(bool); ok {
		result.OvercommitGuestOverhead = v
	}

	return result, nil
}

// SchemaToHapiDevices converts Terraform schema devices (disks, interfaces) to HAPI Devices
func SchemaToHapiDevices(d *schema.ResourceData) (*models.V1VMDevices, error) {
	devices := &models.V1VMDevices{}

	// Disks
	if v, ok := d.GetOk("disk"); ok {
		disks, err := SchemaToHapiDisks(v.([]interface{}))
		if err != nil {
			return nil, fmt.Errorf("failed to convert disks: %w", err)
		}
		devices.Disks = disks
	}

	// Interfaces
	if v, ok := d.GetOk("interface"); ok {
		interfaces, err := SchemaToHapiInterfaces(v.([]interface{}))
		if err != nil {
			return nil, fmt.Errorf("failed to convert interfaces: %w", err)
		}
		devices.Interfaces = interfaces
	}

	return devices, nil
}

// SchemaToHapiCPU converts Terraform schema CPU to HAPI CPU
func SchemaToHapiCPU(cpu map[string]interface{}) (*models.V1VMCPU, error) {
	if len(cpu) == 0 {
		return nil, nil
	}

	result := &models.V1VMCPU{}

	if v, ok := cpu["cores"].(int); ok {
		if v < 0 {
			return nil, fmt.Errorf("cores value %d cannot be negative", v)
		}
		if v > math.MaxInt {
			return nil, fmt.Errorf("cores value %d is out of range for int64", v)
		}
		result.Cores = int64(v)
	}

	if v, ok := cpu["sockets"].(int); ok {
		if v < 0 {
			return nil, fmt.Errorf("sockets value %d cannot be negative", v)
		}
		if v > math.MaxInt {
			return nil, fmt.Errorf("sockets value %d is out of range for int64", v)
		}
		result.Sockets = int64(v)
	}

	if v, ok := cpu["threads"].(int); ok {
		if v < 0 {
			return nil, fmt.Errorf("threads value %d cannot be negative", v)
		}
		if v > math.MaxInt {
			return nil, fmt.Errorf("threads value %d is out of range for int64", v)
		}
		result.Threads = int64(v)
	}

	return result, nil
}

// SchemaToHapiMemory converts Terraform schema memory to HAPI Memory
func SchemaToHapiMemory(memory []interface{}) (*models.V1VMMemory, error) {
	if len(memory) == 0 || memory[0] == nil {
		return nil, nil
	}

	result := &models.V1VMMemory{}
	in := memory[0].(map[string]interface{})

	if v, ok := in["guest"].(string); ok && v != "" {
		// HAPI uses V1VMQuantity which is a string type
		result.Guest = models.V1VMQuantity(v)
	}

	if v, ok := in["hugepages"].(string); ok && v != "" {
		hugepages := &models.V1VMHugepages{
			PageSize: v,
		}
		result.Hugepages = hugepages
	}

	return result, nil
}

// SchemaToHapiFirmware converts Terraform schema firmware to HAPI Firmware
func SchemaToHapiFirmware(firmware []interface{}) (*models.V1VMFirmware, error) {
	if len(firmware) == 0 || firmware[0] == nil {
		return nil, nil
	}

	result := &models.V1VMFirmware{}
	in := firmware[0].(map[string]interface{})

	if v, ok := in["uuid"].(string); ok && v != "" {
		result.UUID = v
	}

	if v, ok := in["serial"].(string); ok && v != "" {
		result.Serial = v
	}

	if v, ok := in["bootloader"].([]interface{}); ok && len(v) > 0 {
		bootloader, err := SchemaToHapiBootloader(v)
		if err != nil {
			return nil, fmt.Errorf("failed to convert bootloader: %w", err)
		}
		result.Bootloader = bootloader
	}

	return result, nil
}

// SchemaToHapiBootloader converts Terraform schema bootloader to HAPI Bootloader
func SchemaToHapiBootloader(bootloader []interface{}) (*models.V1VMBootloader, error) {
	if len(bootloader) == 0 || bootloader[0] == nil {
		return nil, nil
	}

	result := &models.V1VMBootloader{}
	in := bootloader[0].(map[string]interface{})

	if v, ok := in["bios"].([]interface{}); ok && len(v) > 0 {
		bios, err := SchemaToHapiBIOS(v)
		if err != nil {
			return nil, fmt.Errorf("failed to convert BIOS: %w", err)
		}
		result.Bios = bios
	}

	if v, ok := in["efi"].([]interface{}); ok && len(v) > 0 {
		efi, err := SchemaToHapiEFI(v)
		if err != nil {
			return nil, fmt.Errorf("failed to convert EFI: %w", err)
		}
		result.Efi = efi
	}

	return result, nil
}

// SchemaToHapiBIOS converts Terraform schema BIOS to HAPI BIOS
func SchemaToHapiBIOS(bios []interface{}) (*models.V1VMBIOS, error) {
	if len(bios) == 0 || bios[0] == nil {
		return &models.V1VMBIOS{}, nil
	}

	result := &models.V1VMBIOS{}
	in := bios[0].(map[string]interface{})

	if v, ok := in["use_serial"].(bool); ok {
		result.UseSerial = v
	}

	return result, nil
}

// SchemaToHapiEFI converts Terraform schema EFI to HAPI EFI
func SchemaToHapiEFI(efi []interface{}) (*models.V1VMEFI, error) {
	if len(efi) == 0 || efi[0] == nil {
		return &models.V1VMEFI{}, nil
	}

	result := &models.V1VMEFI{}
	in := efi[0].(map[string]interface{})

	if v, ok := in["secure_boot"].(bool); ok {
		result.SecureBoot = v
	}

	if v, ok := in["persistent"].(bool); ok {
		result.Persistent = v
	}

	return result, nil
}

// SchemaToHapiFeatures converts Terraform schema features to HAPI Features
func SchemaToHapiFeatures(features []interface{}) (*models.V1VMFeatures, error) {
	if len(features) == 0 || features[0] == nil {
		return nil, nil
	}

	result := &models.V1VMFeatures{}
	in := features[0].(map[string]interface{})

	if v, ok := in["acpi"].([]interface{}); ok && len(v) > 0 {
		acpi, err := SchemaToHapiFeatureState(v)
		if err != nil {
			return nil, fmt.Errorf("failed to convert ACPI feature: %w", err)
		}
		result.Acpi = acpi
	}

	if v, ok := in["apic"].([]interface{}); ok && len(v) > 0 {
		apic, err := SchemaToHapiFeatureState(v)
		if err != nil {
			return nil, fmt.Errorf("failed to convert APIC feature: %w", err)
		}
		result.Apic = &models.V1VMFeatureAPIC{
			Enabled: apic.Enabled,
		}
	}

	if v, ok := in["smm"].([]interface{}); ok && len(v) > 0 {
		smm, err := SchemaToHapiFeatureState(v)
		if err != nil {
			return nil, fmt.Errorf("failed to convert SMM feature: %w", err)
		}
		result.Smm = smm
	}

	return result, nil
}

// SchemaToHapiFeatureState converts Terraform schema feature state to HAPI FeatureState
func SchemaToHapiFeatureState(featureState []interface{}) (*models.V1VMFeatureState, error) {
	if len(featureState) == 0 || featureState[0] == nil {
		return &models.V1VMFeatureState{}, nil
	}

	result := &models.V1VMFeatureState{}
	in := featureState[0].(map[string]interface{})

	if v, ok := in["enabled"].(bool); ok {
		result.Enabled = v
	}

	return result, nil
}

// SchemaToHapiDisks converts Terraform schema disks to HAPI Disks
// Uses JSON marshaling as bridge since structures are compatible
func SchemaToHapiDisks(disks []interface{}) ([]*models.V1VMDisk, error) {
	if len(disks) == 0 {
		return nil, nil
	}

	// Convert Terraform schema to JSON-compatible structure
	disksJSON := make([]map[string]interface{}, 0, len(disks))
	for _, disk := range disks {
		if disk == nil {
			continue
		}
		diskMap, ok := disk.(map[string]interface{})
		if !ok {
			continue
		}
		diskJSON := make(map[string]interface{})

		if v, ok := diskMap["name"].(string); ok {
			diskJSON["name"] = v
		}

		if v, ok := diskMap["serial"].(string); ok {
			diskJSON["serial"] = v
		}

		if v, ok := diskMap["boot_order"].(int); ok && v > 0 {
			diskJSON["bootOrder"] = uint(v)
		}

		// DiskDevice - always set if disk_device exists, even if disk doesn't (matches HapiDisksToSchema pattern)
		if v, ok := diskMap["disk_device"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
			diskDeviceMap, ok := v[0].(map[string]interface{})
			if !ok {
				continue
			}
			diskDeviceJSON := make(map[string]interface{})

			if diskList, ok := diskDeviceMap["disk"].([]interface{}); ok && len(diskList) > 0 && diskList[0] != nil {
				diskTargetMap, ok := diskList[0].(map[string]interface{})
				if !ok {
					continue
				}
				diskTargetJSON := make(map[string]interface{})

				if bus, ok := diskTargetMap["bus"].(string); ok {
					diskTargetJSON["bus"] = bus
				}
				if readOnly, ok := diskTargetMap["read_only"].(bool); ok {
					diskTargetJSON["readOnly"] = readOnly
				}
				if pciAddress, ok := diskTargetMap["pci_address"].(string); ok {
					diskTargetJSON["pciAddress"] = pciAddress
				}

				diskDeviceJSON["disk"] = diskTargetJSON
			}
			// Always set diskDevice if disk_device exists, even if disk doesn't exist inside it
			// This matches the pattern in HapiDisksToSchema where we always set disk_device
			diskJSON["diskDevice"] = diskDeviceJSON
		}

		disksJSON = append(disksJSON, diskJSON)
	}

	// Marshal to JSON and unmarshal to HAPI models
	jsonBytes, err := json.Marshal(disksJSON)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal disks to JSON: %w", err)
	}

	var hapiDisks []*models.V1VMDisk
	if err := json.Unmarshal(jsonBytes, &hapiDisks); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON to HAPI disks: %w", err)
	}

	return hapiDisks, nil
}

// SchemaToHapiInterfaces converts Terraform schema interfaces to HAPI Interfaces
// Uses JSON marshaling as bridge since structures are compatible
func SchemaToHapiInterfaces(interfaces []interface{}) ([]*models.V1VMInterface, error) {
	if len(interfaces) == 0 {
		return nil, nil
	}

	// Convert Terraform schema to JSON-compatible structure
	interfacesJSON := make([]map[string]interface{}, 0, len(interfaces))
	for _, iface := range interfaces {
		if iface == nil {
			continue
		}
		ifaceMap, ok := iface.(map[string]interface{})
		if !ok {
			continue
		}
		ifaceJSON := make(map[string]interface{})

		if v, ok := ifaceMap["name"].(string); ok {
			ifaceJSON["name"] = v
		}

		if v, ok := ifaceMap["model"].(string); ok {
			ifaceJSON["model"] = v
		}

		// InterfaceBindingMethod
		if v, ok := ifaceMap["interface_binding_method"].(string); ok {
			bindingMethodJSON := make(map[string]interface{})
			switch v {
			case "InterfaceBridge":
				bindingMethodJSON["bridge"] = map[string]interface{}{}
			case "InterfaceSlirp":
				bindingMethodJSON["deprecatedSlirp"] = map[string]interface{}{}
			case "InterfaceMasquerade":
				bindingMethodJSON["masquerade"] = map[string]interface{}{}
			case "InterfaceSRIOV":
				bindingMethodJSON["sriov"] = map[string]interface{}{}
			}
			ifaceJSON["interfaceBindingMethod"] = bindingMethodJSON
		}

		interfacesJSON = append(interfacesJSON, ifaceJSON)
	}

	// Marshal to JSON and unmarshal to HAPI models
	jsonBytes, err := json.Marshal(interfacesJSON)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal interfaces to JSON: %w", err)
	}

	var hapiInterfaces []*models.V1VMInterface
	if err := json.Unmarshal(jsonBytes, &hapiInterfaces); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON to HAPI interfaces: %w", err)
	}

	return hapiInterfaces, nil
}

// SchemaToHapiVolumes converts Terraform schema volumes to HAPI Volumes
// Uses JSON marshaling as bridge since structures are compatible
func SchemaToHapiVolumes(volumes []interface{}) ([]*models.V1VMVolume, error) {
	if len(volumes) == 0 {
		return nil, nil
	}

	// Use existing expandVolumes function pattern but convert via JSON
	// First, expand to a JSON-compatible structure
	volumesJSON := make([]map[string]interface{}, 0, len(volumes))
	for _, volume := range volumes {
		if volume == nil {
			continue
		}
		volumeMap, ok := volume.(map[string]interface{})
		if !ok {
			continue
		}
		volumeJSON := make(map[string]interface{})

		if v, ok := volumeMap["name"].(string); ok {
			volumeJSON["name"] = v
		}

		// VolumeSource - complex nested structure, use JSON marshaling
		if v, ok := volumeMap["volume_source"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
			volumeSourceMap, ok := v[0].(map[string]interface{})
			if !ok {
				continue
			}
			volumeSourceJSON := make(map[string]interface{})

			// Handle all volume source types
			if dv, ok := volumeSourceMap["data_volume"].([]interface{}); ok && len(dv) > 0 && dv[0] != nil {
				if dvMap, ok := dv[0].(map[string]interface{}); ok {
					volumeSourceJSON["dataVolume"] = map[string]interface{}{
						"name": dvMap["name"],
					}
				}
			}
			if cicd, ok := volumeSourceMap["cloud_init_config_drive"].([]interface{}); ok && len(cicd) > 0 {
				volumeSourceJSON["cloudInitConfigDrive"] = expandCloudInitSource(cicd)
			}
			if sa, ok := volumeSourceMap["service_account"].([]interface{}); ok && len(sa) > 0 && sa[0] != nil {
				if saMap, ok := sa[0].(map[string]interface{}); ok {
					volumeSourceJSON["serviceAccount"] = map[string]interface{}{
						"serviceAccountName": saMap["service_account_name"],
					}
				}
			}
			// Handle schema.Set for container_disk and cloud_init_no_cloud
			if cdSet, ok := volumeSourceMap["container_disk"].(*schema.Set); ok && cdSet.Len() > 0 {
				volumeSourceJSON["containerDisk"] = expandContainerDiskSource(cdSet.List())
			} else if cdList, ok := volumeSourceMap["container_disk"].([]interface{}); ok && len(cdList) > 0 {
				volumeSourceJSON["containerDisk"] = expandContainerDiskSource(cdList)
			}
			if cinSet, ok := volumeSourceMap["cloud_init_no_cloud"].(*schema.Set); ok && cinSet.Len() > 0 {
				volumeSourceJSON["cloudInitNoCloud"] = expandCloudInitSource(cinSet.List())
			} else if cinList, ok := volumeSourceMap["cloud_init_no_cloud"].([]interface{}); ok && len(cinList) > 0 {
				volumeSourceJSON["cloudInitNoCloud"] = expandCloudInitSource(cinList)
			}
			if ep, ok := volumeSourceMap["ephemeral"].([]interface{}); ok && len(ep) > 0 {
				volumeSourceJSON["ephemeral"] = expandEphemeralSource(ep)
			}
			if ed, ok := volumeSourceMap["empty_disk"].([]interface{}); ok && len(ed) > 0 && ed[0] != nil {
				volumeSourceJSON["emptyDisk"] = expandEmptyDiskSource(ed)
			}
			if pvc, ok := volumeSourceMap["persistent_volume_claim"].([]interface{}); ok && len(pvc) > 0 && pvc[0] != nil {
				volumeSourceJSON["persistentVolumeClaim"] = expandPVCSource(pvc)
			}
			if hd, ok := volumeSourceMap["host_disk"].([]interface{}); ok && len(hd) > 0 && hd[0] != nil {
				volumeSourceJSON["hostDisk"] = expandHostDiskSource(hd)
			}
			if cm, ok := volumeSourceMap["config_map"].([]interface{}); ok && len(cm) > 0 && cm[0] != nil {
				volumeSourceJSON["configMap"] = expandConfigMapSource(cm)
			}

			// Flatten volume source fields directly onto volumeJSON (not nested under "volumeSource")
			// HAPI model expects volume source fields directly on the volume object
			for k, v := range volumeSourceJSON {
				volumeJSON[k] = v
			}
		}

		volumesJSON = append(volumesJSON, volumeJSON)
	}

	// Marshal to JSON and unmarshal to HAPI models
	jsonBytes, err := json.Marshal(volumesJSON)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal volumes to JSON: %w", err)
	}

	var hapiVolumes []*models.V1VMVolume
	if err := json.Unmarshal(jsonBytes, &hapiVolumes); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON to HAPI volumes: %w", err)
	}

	return hapiVolumes, nil
}

// Helper functions for volume source expansion
func expandCloudInitSource(source []interface{}) map[string]interface{} {
	if len(source) == 0 {
		return nil
	}
	sourceMap := source[0].(map[string]interface{})
	result := make(map[string]interface{})
	if v, ok := sourceMap["user_data_secret_ref"].([]interface{}); ok {
		result["userDataSecretRef"] = expandLocalObjectRef(v)
	}
	if v, ok := sourceMap["user_data_base64"].(string); ok {
		result["userDataBase64"] = v
	}
	if v, ok := sourceMap["user_data"].(string); ok {
		result["userData"] = v
	}
	if v, ok := sourceMap["network_data_secret_ref"].([]interface{}); ok {
		result["networkDataSecretRef"] = expandLocalObjectRef(v)
	}
	if v, ok := sourceMap["network_data_base64"].(string); ok {
		result["networkDataBase64"] = v
	}
	if v, ok := sourceMap["network_data"].(string); ok {
		result["networkData"] = v
	}
	return result
}

func expandContainerDiskSource(source []interface{}) map[string]interface{} {
	if len(source) == 0 {
		return nil
	}
	sourceMap := source[0].(map[string]interface{})
	return map[string]interface{}{
		"image": sourceMap["image_url"],
	}
}

func expandEphemeralSource(source []interface{}) map[string]interface{} {
	if len(source) == 0 {
		return nil
	}
	sourceMap := source[0].(map[string]interface{})
	if pvc, ok := sourceMap["persistent_volume_claim"].([]interface{}); ok && len(pvc) > 0 {
		return map[string]interface{}{
			"persistentVolumeClaim": expandPVCSource(pvc),
		}
	}
	return nil
}

func expandEmptyDiskSource(source []interface{}) map[string]interface{} {
	if len(source) == 0 {
		return nil
	}
	sourceMap := source[0].(map[string]interface{})
	return map[string]interface{}{
		"capacity": sourceMap["capacity"],
	}
}

func expandPVCSource(source []interface{}) map[string]interface{} {
	if len(source) == 0 {
		return nil
	}
	sourceMap := source[0].(map[string]interface{})
	result := map[string]interface{}{
		"claimName": sourceMap["claim_name"],
	}
	if v, ok := sourceMap["read_only"].(bool); ok {
		result["readOnly"] = v
	}
	return result
}

func expandHostDiskSource(source []interface{}) map[string]interface{} {
	if len(source) == 0 {
		return nil
	}
	sourceMap := source[0].(map[string]interface{})
	return map[string]interface{}{
		"path": sourceMap["path"],
		"type": sourceMap["type"],
	}
}

func expandConfigMapSource(source []interface{}) map[string]interface{} {
	if len(source) == 0 {
		return nil
	}
	sourceMap := source[0].(map[string]interface{})
	result := make(map[string]interface{})
	if v, ok := sourceMap["name"].(string); ok {
		result["name"] = v
	}
	if v, ok := sourceMap["optional"].(bool); ok {
		result["optional"] = v
	}
	if v, ok := sourceMap["volume_label"].(string); ok {
		result["volumeLabel"] = v
	}
	return result
}

func expandLocalObjectRef(ref []interface{}) map[string]interface{} {
	if len(ref) == 0 {
		return nil
	}
	refMap := ref[0].(map[string]interface{})
	return map[string]interface{}{
		"name": refMap["name"],
	}
}

// SchemaToHapiNetworks converts Terraform schema networks to HAPI Networks
// Uses JSON marshaling as bridge since structures are compatible
func SchemaToHapiNetworks(networks []interface{}) ([]*models.V1VMNetwork, error) {
	if len(networks) == 0 {
		return nil, nil
	}

	// Convert Terraform schema to JSON-compatible structure
	networksJSON := make([]map[string]interface{}, 0, len(networks))
	for _, network := range networks {
		if network == nil {
			continue
		}
		networkMap, ok := network.(map[string]interface{})
		if !ok {
			continue
		}
		networkJSON := make(map[string]interface{})

		if v, ok := networkMap["name"].(string); ok {
			networkJSON["name"] = v
		}

		// NetworkSource
		if v, ok := networkMap["network_source"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
			networkSourceMap, ok := v[0].(map[string]interface{})
			if !ok {
				continue
			}
			networkSourceJSON := make(map[string]interface{})

			if pod, ok := networkSourceMap["pod"].([]interface{}); ok && len(pod) > 0 && pod[0] != nil {
				podMap, ok := pod[0].(map[string]interface{})
				if ok {
					podJSON := make(map[string]interface{})
					if vmCIDR, ok := podMap["vm_network_cidr"].(string); ok {
						podJSON["vmNetworkCIDR"] = vmCIDR
					}
					networkSourceJSON["pod"] = podJSON
				}
			}

			if multus, ok := networkSourceMap["multus"].([]interface{}); ok && len(multus) > 0 && multus[0] != nil {
				multusMap, ok := multus[0].(map[string]interface{})
				if ok {
					multusJSON := make(map[string]interface{})
					if networkName, ok := multusMap["network_name"].(string); ok {
						multusJSON["networkName"] = networkName
					}
					if defaultNet, ok := multusMap["default"].(bool); ok {
						multusJSON["default"] = defaultNet
					}
					networkSourceJSON["multus"] = multusJSON
				}
			}

			if len(networkSourceJSON) > 0 {
				networkJSON["networkSource"] = networkSourceJSON
			}
		}

		networksJSON = append(networksJSON, networkJSON)
	}

	// Marshal to JSON and unmarshal to HAPI models
	jsonBytes, err := json.Marshal(networksJSON)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal networks to JSON: %w", err)
	}

	var hapiNetworks []*models.V1VMNetwork
	if err := json.Unmarshal(jsonBytes, &hapiNetworks); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON to HAPI networks: %w", err)
	}

	return hapiNetworks, nil
}

// SchemaToHapiAffinity converts Terraform schema affinity to HAPI Affinity
// Uses JSON marshaling as bridge via k8s types
func SchemaToHapiAffinity(affinity []interface{}) (*models.V1VMAffinity, error) {
	if len(affinity) == 0 {
		return nil, nil
	}

	// Use existing k8s.ExpandAffinity to get k8s types, then convert via JSON
	k8sAffinity := k8s.ExpandAffinity(affinity)

	// Marshal k8s affinity to JSON and unmarshal to HAPI
	jsonBytes, err := json.Marshal(k8sAffinity)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal affinity to JSON: %w", err)
	}

	var hapiAffinity models.V1VMAffinity
	if err := json.Unmarshal(jsonBytes, &hapiAffinity); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON to HAPI affinity: %w", err)
	}

	return &hapiAffinity, nil
}

// SchemaToHapiTolerations converts Terraform schema tolerations to HAPI Tolerations
// Uses JSON marshaling as bridge via k8s types
func SchemaToHapiTolerations(tolerations []interface{}) ([]*models.V1VMToleration, error) {
	if len(tolerations) == 0 {
		return nil, nil
	}

	// Use existing k8s.ExpandTolerations to get k8s types, then convert via JSON
	k8sTolerations, err := k8s.ExpandTolerations(tolerations)
	if err != nil {
		return nil, fmt.Errorf("failed to expand tolerations: %w", err)
	}

	// Marshal k8s tolerations to JSON and unmarshal to HAPI
	jsonBytes, err := json.Marshal(k8sTolerations)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal tolerations to JSON: %w", err)
	}

	var hapiTolerations []*models.V1VMToleration
	if err := json.Unmarshal(jsonBytes, &hapiTolerations); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON to HAPI tolerations: %w", err)
	}

	return hapiTolerations, nil
}

// SchemaToHapiProbe converts Terraform schema probe to HAPI Probe
// Uses JSON marshaling as bridge
func SchemaToHapiProbe(probe []interface{}) (*models.V1VMProbe, error) {
	if len(probe) == 0 || probe[0] == nil {
		return nil, nil
	}

	// Use existing expandProbe pattern but convert via JSON
	// For now, use JSON marshaling as bridge
	probeMap := probe[0].(map[string]interface{})
	probeJSON := make(map[string]interface{})

	// Handle probe fields (initialDelaySeconds, timeoutSeconds, periodSeconds, successThreshold, failureThreshold, handler)
	if v, ok := probeMap["initial_delay_seconds"].(int); ok {
		probeJSON["initialDelaySeconds"] = int64(v)
	}
	if v, ok := probeMap["timeout_seconds"].(int); ok {
		probeJSON["timeoutSeconds"] = int64(v)
	}
	if v, ok := probeMap["period_seconds"].(int); ok {
		probeJSON["periodSeconds"] = int64(v)
	}
	if v, ok := probeMap["success_threshold"].(int); ok {
		probeJSON["successThreshold"] = int32(v)
	}
	if v, ok := probeMap["failure_threshold"].(int); ok {
		probeJSON["failureThreshold"] = int32(v)
	}

	// Handler (httpGet, tcpSocket, exec)
	if v, ok := probeMap["http_get"].([]interface{}); ok && len(v) > 0 {
		probeJSON["httpGet"] = expandHTTPGetAction(v)
	}
	if v, ok := probeMap["tcp_socket"].([]interface{}); ok && len(v) > 0 {
		probeJSON["tcpSocket"] = expandTCPSocketAction(v)
	}
	if v, ok := probeMap["exec"].([]interface{}); ok && len(v) > 0 {
		probeJSON["exec"] = expandExecAction(v)
	}

	// Marshal to JSON and unmarshal to HAPI
	jsonBytes, err := json.Marshal(probeJSON)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal probe to JSON: %w", err)
	}

	var hapiProbe models.V1VMProbe
	if err := json.Unmarshal(jsonBytes, &hapiProbe); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON to HAPI probe: %w", err)
	}

	return &hapiProbe, nil
}

// Helper functions for probe actions
func expandHTTPGetAction(httpGet []interface{}) map[string]interface{} {
	if len(httpGet) == 0 {
		return nil
	}
	actionMap := httpGet[0].(map[string]interface{})
	result := make(map[string]interface{})
	if v, ok := actionMap["path"].(string); ok {
		result["path"] = v
	}
	if v, ok := actionMap["port"].(int); ok {
		result["port"] = int32(v)
	}
	if v, ok := actionMap["scheme"].(string); ok {
		result["scheme"] = v
	}
	return result
}

func expandTCPSocketAction(tcpSocket []interface{}) map[string]interface{} {
	if len(tcpSocket) == 0 {
		return nil
	}
	actionMap := tcpSocket[0].(map[string]interface{})
	result := make(map[string]interface{})
	if v, ok := actionMap["port"].(int); ok {
		result["port"] = int32(v)
	}
	return result
}

func expandExecAction(exec []interface{}) map[string]interface{} {
	if len(exec) == 0 {
		return nil
	}
	actionMap := exec[0].(map[string]interface{})
	result := make(map[string]interface{})
	if v, ok := actionMap["command"].([]interface{}); ok {
		commands := make([]string, len(v))
		for i, cmd := range v {
			commands[i] = cmd.(string)
		}
		result["command"] = commands
	}
	return result
}

// SchemaToHapiDNSConfig converts Terraform schema DNS config to HAPI PodDNSConfig
// Uses JSON marshaling as bridge via k8s types
func SchemaToHapiDNSConfig(dnsConfig []interface{}) (*models.V1VMPodDNSConfig, error) {
	if len(dnsConfig) == 0 {
		return nil, nil
	}

	// Use existing k8s.ExpandPodDNSConfig to get k8s types, then convert via JSON
	k8sDNSConfig, err := k8s.ExpandPodDNSConfig(dnsConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to expand DNS config: %w", err)
	}

	// Marshal k8s DNS config to JSON and unmarshal to HAPI
	jsonBytes, err := json.Marshal(k8sDNSConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal DNS config to JSON: %w", err)
	}

	var hapiDNSConfig models.V1VMPodDNSConfig
	if err := json.Unmarshal(jsonBytes, &hapiDNSConfig); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON to HAPI DNS config: %w", err)
	}

	return &hapiDNSConfig, nil
}
