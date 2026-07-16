package virtualmachineinstance

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// domain_spec_test.go — Batch 7.
// Covers the previously-untested expand*/flatten*/binding helpers in
// domain_spec.go. Most helpers accept []interface{} / map[string]interface{}
// directly, so we drive them with hand-built fixtures. ExpandDomainSpec
// and expandDevicesToVM take a *schema.ResourceData; we build a minimal
// schema.Resource with the top-level keys they read.

// ---------------------------------------------------------------------------
// Helpers that operate on []interface{} maps
// ---------------------------------------------------------------------------

func TestExpandResources(t *testing.T) {
	t.Run("empty input", func(t *testing.T) {
		got, err := expandResources(nil)
		require.NoError(t, err)
		require.NotNil(t, got)
	})

	t.Run("populated", func(t *testing.T) {
		got, err := expandResources([]interface{}{
			map[string]interface{}{
				"requests":                   map[string]interface{}{"memory": "1Gi", "cpu": "500m"},
				"limits":                     map[string]interface{}{"memory": "2Gi"},
				"over_commit_guest_overhead": true,
			},
		})
		require.NoError(t, err)
		assert.NotNil(t, got.Requests)
		assert.NotNil(t, got.Limits)
		assert.True(t, got.OvercommitGuestOverhead)
	})
}

func TestExpandCPUToVM(t *testing.T) {
	t.Run("empty map returns empty struct", func(t *testing.T) {
		got, err := expandCPUToVM(map[string]interface{}{})
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.EqualValues(t, 0, got.Cores)
	})

	t.Run("populated", func(t *testing.T) {
		got, err := expandCPUToVM(map[string]interface{}{
			"cores":   4,
			"sockets": 2,
			"threads": 1,
		})
		require.NoError(t, err)
		assert.EqualValues(t, 4, got.Cores)
		assert.EqualValues(t, 2, got.Sockets)
		assert.EqualValues(t, 1, got.Threads)
	})

	t.Run("negative cores rejected", func(t *testing.T) {
		_, err := expandCPUToVM(map[string]interface{}{"cores": -1})
		assert.Error(t, err)
	})

	t.Run("negative sockets rejected", func(t *testing.T) {
		_, err := expandCPUToVM(map[string]interface{}{"sockets": -1})
		assert.Error(t, err)
	})

	t.Run("negative threads rejected", func(t *testing.T) {
		_, err := expandCPUToVM(map[string]interface{}{"threads": -1})
		assert.Error(t, err)
	})
}

func TestExpandMemoryToVM(t *testing.T) {
	t.Run("empty input", func(t *testing.T) {
		got, err := expandMemoryToVM(nil)
		require.NoError(t, err)
		require.NotNil(t, got)
	})

	t.Run("populated", func(t *testing.T) {
		got, err := expandMemoryToVM([]interface{}{
			map[string]interface{}{"guest": "1Gi", "hugepages": "2Mi"},
		})
		require.NoError(t, err)
		assert.Equal(t, models.V1VMQuantity("1Gi"), got.Guest)
		require.NotNil(t, got.Hugepages)
		assert.Equal(t, "2Mi", got.Hugepages.PageSize)
	})
}

func TestExpandFirmwareToVM(t *testing.T) {
	t.Run("empty returns nil", func(t *testing.T) {
		got, err := expandFirmwareToVM(nil)
		require.NoError(t, err)
		assert.Nil(t, got)
	})

	t.Run("uuid + serial + bootloader", func(t *testing.T) {
		got, err := expandFirmwareToVM([]interface{}{
			map[string]interface{}{
				"uuid":   "abc-123",
				"serial": "S1",
				"bootloader": []interface{}{
					map[string]interface{}{
						"bios": []interface{}{map[string]interface{}{"use_serial": true}},
					},
				},
			},
		})
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, "abc-123", got.UUID)
		assert.Equal(t, "S1", got.Serial)
		require.NotNil(t, got.Bootloader)
		require.NotNil(t, got.Bootloader.Bios)
		assert.True(t, got.Bootloader.Bios.UseSerial)
	})
}

func TestExpandBootloaderToVM(t *testing.T) {
	assert.Nil(t, expandBootloaderToVM(nil))

	got := expandBootloaderToVM([]interface{}{
		map[string]interface{}{
			"bios": []interface{}{map[string]interface{}{"use_serial": true}},
			"efi":  []interface{}{map[string]interface{}{"secure_boot": true, "persistent": true}},
		},
	})
	require.NotNil(t, got)
	require.NotNil(t, got.Bios)
	require.NotNil(t, got.Efi)
	assert.True(t, got.Bios.UseSerial)
	require.NotNil(t, got.Efi.SecureBoot)
	assert.True(t, *got.Efi.SecureBoot)
}

func TestExpandBIOSToVM(t *testing.T) {
	assert.Nil(t, expandBIOSToVM(nil))
	got := expandBIOSToVM([]interface{}{map[string]interface{}{"use_serial": true}})
	require.NotNil(t, got)
	assert.True(t, got.UseSerial)
}

func TestExpandEFIToVM(t *testing.T) {
	assert.Nil(t, expandEFIToVM(nil))
	got := expandEFIToVM([]interface{}{
		map[string]interface{}{"secure_boot": false, "persistent": true},
	})
	require.NotNil(t, got)
	require.NotNil(t, got.SecureBoot)
	assert.False(t, *got.SecureBoot)
	require.NotNil(t, got.Persistent)
	assert.True(t, *got.Persistent)
}

func TestExpandFeaturesToVM(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		got, err := expandFeaturesToVM(nil)
		require.NoError(t, err)
		assert.Nil(t, got)
	})

	t.Run("all three sub-blocks", func(t *testing.T) {
		got, err := expandFeaturesToVM([]interface{}{
			map[string]interface{}{
				"acpi": []interface{}{map[string]interface{}{"enabled": true}},
				"apic": []interface{}{map[string]interface{}{"enabled": true}},
				"smm":  []interface{}{map[string]interface{}{"enabled": false}},
			},
		})
		require.NoError(t, err)
		require.NotNil(t, got.Acpi)
		require.NotNil(t, got.Apic)
		require.NotNil(t, got.Smm)
		assert.True(t, got.Acpi.Enabled)
		assert.False(t, got.Smm.Enabled)
	})
}

func TestExpandFeatureStateToVM(t *testing.T) {
	assert.Nil(t, expandFeatureStateToVM(nil))
	got := expandFeatureStateToVM([]interface{}{map[string]interface{}{"enabled": true}})
	require.NotNil(t, got)
	assert.True(t, got.Enabled)
}

func TestExpandFeatureAPICToVM(t *testing.T) {
	assert.Nil(t, expandFeatureAPICToVM(nil))
	got := expandFeatureAPICToVM([]interface{}{map[string]interface{}{"enabled": true}})
	require.NotNil(t, got)
	assert.True(t, got.Enabled)
}

// ---------------------------------------------------------------------------
// Disk & interface expand
// ---------------------------------------------------------------------------

func TestExpandDisksToVM(t *testing.T) {
	assert.Nil(t, expandDisksToVM(nil))
	assert.Nil(t, expandDisksToVM([]interface{}{nil}))

	got := expandDisksToVM([]interface{}{
		map[string]interface{}{
			"name":       "disk0",
			"serial":     "S1",
			"boot_order": 1,
			"disk_device": []interface{}{
				map[string]interface{}{
					"disk": []interface{}{map[string]interface{}{
						"bus":         "virtio",
						"read_only":   true,
						"pci_address": "0000:81:01.1",
					}},
				},
			},
		},
	})
	require.Len(t, got, 1)
	require.NotNil(t, got[0].Name)
	assert.Equal(t, "disk0", *got[0].Name)
	assert.Equal(t, "S1", got[0].Serial)
	assert.EqualValues(t, 1, got[0].BootOrder)
	require.NotNil(t, got[0].Disk)
	assert.Equal(t, "virtio", got[0].Disk.Bus)
	assert.True(t, got[0].Disk.Readonly)
}

func TestExpandDiskDeviceToVM_AllTargets(t *testing.T) {
	disk := &models.V1VMDisk{}
	// Empty input is a no-op.
	expandDiskDeviceToVM(nil, disk)
	assert.Nil(t, disk.Disk)

	// disk / cdrom / lun all in one — one call routes each subtype.
	expandDiskDeviceToVM([]interface{}{
		map[string]interface{}{
			"disk":  []interface{}{map[string]interface{}{"bus": "virtio"}},
			"cdrom": []interface{}{map[string]interface{}{"bus": "sata"}},
			"lun":   []interface{}{map[string]interface{}{"bus": "scsi", "read_only": true}},
		},
	}, disk)
	require.NotNil(t, disk.Disk)
	require.NotNil(t, disk.Cdrom)
	require.NotNil(t, disk.Lun)
	assert.Equal(t, "virtio", disk.Disk.Bus)
	assert.Equal(t, "sata", disk.Cdrom.Bus)
	assert.Equal(t, "scsi", disk.Lun.Bus)
	assert.True(t, disk.Lun.Readonly)
}

func TestExpandDiskTargetToVM(t *testing.T) {
	assert.Nil(t, expandDiskTargetToVM(nil))
	got := expandDiskTargetToVM([]interface{}{
		map[string]interface{}{"bus": "virtio", "read_only": true, "pci_address": "0000:81:01.1"},
	})
	require.NotNil(t, got)
	assert.Equal(t, "virtio", got.Bus)
	assert.True(t, got.Readonly)
	assert.Equal(t, "0000:81:01.1", got.PciAddress)
}

func TestExpandCDRomTargetToVM(t *testing.T) {
	assert.Nil(t, expandCDRomTargetToVM(nil))
	got := expandCDRomTargetToVM([]interface{}{map[string]interface{}{"bus": "sata"}})
	require.NotNil(t, got)
	assert.Equal(t, "sata", got.Bus)
}

func TestExpandLunTargetToVM(t *testing.T) {
	assert.Nil(t, expandLunTargetToVM(nil))
	got := expandLunTargetToVM([]interface{}{
		map[string]interface{}{"bus": "scsi", "read_only": true},
	})
	require.NotNil(t, got)
	assert.Equal(t, "scsi", got.Bus)
	assert.True(t, got.Readonly)
}

func TestExpandInterfacesToVM(t *testing.T) {
	assert.Nil(t, expandInterfacesToVM(nil))
	assert.Nil(t, expandInterfacesToVM([]interface{}{nil}))

	got := expandInterfacesToVM([]interface{}{
		map[string]interface{}{
			"name":                     "eth0",
			"model":                    "virtio",
			"interface_binding_method": "InterfaceBridge",
		},
	})
	require.Len(t, got, 1)
	require.NotNil(t, got[0].Name)
	assert.Equal(t, "eth0", *got[0].Name)
	assert.Equal(t, "virtio", got[0].Model)
	// After bridge is set, Bridge should be non-nil.
	assert.NotNil(t, got[0].Bridge)
}

// TestSetVMInterfaceBindingMethod pins that each accepted string flips a
// distinct interface field. Unknown strings are a no-op.
func TestSetVMInterfaceBindingMethod(t *testing.T) {
	cases := []struct {
		in    string
		check func(*models.V1VMInterface) bool
	}{
		{"InterfaceBridge", func(i *models.V1VMInterface) bool { return i.Bridge != nil }},
		{"bridge", func(i *models.V1VMInterface) bool { return i.Bridge != nil }},
		{"InterfaceMasquerade", func(i *models.V1VMInterface) bool { return i.Masquerade != nil }},
		{"masquerade", func(i *models.V1VMInterface) bool { return i.Masquerade != nil }},
		{"InterfaceSlirp", func(i *models.V1VMInterface) bool { return i.Slirp != nil }},
		{"slirp", func(i *models.V1VMInterface) bool { return i.Slirp != nil }},
		{"InterfaceSRIOV", func(i *models.V1VMInterface) bool { return i.Sriov != nil }},
		{"sriov", func(i *models.V1VMInterface) bool { return i.Sriov != nil }},
		{"macvtap", func(i *models.V1VMInterface) bool { return i.Macvtap != nil }},
		{"passt", func(i *models.V1VMInterface) bool { return i.Passt != nil }},
	}
	for _, c := range cases {
		t.Run(c.in, func(t *testing.T) {
			iface := &models.V1VMInterface{}
			setVMInterfaceBindingMethod(c.in, iface)
			assert.True(t, c.check(iface), "expected %q to set its binding", c.in)
		})
	}

	// Unknown string is a silent no-op.
	iface := &models.V1VMInterface{}
	setVMInterfaceBindingMethod("bogus", iface)
	assert.Nil(t, iface.Bridge)
	assert.Nil(t, iface.Masquerade)
}

// ---------------------------------------------------------------------------
// Flatten helpers
// ---------------------------------------------------------------------------

func TestFlattenResourcesFromVM(t *testing.T) {
	got := flattenResourcesFromVM(nil)
	require.Len(t, got, 1)

	got = flattenResourcesFromVM(&models.V1VMResourceRequirements{
		Requests:                map[string]interface{}{"cpu": "500m"},
		Limits:                  map[string]interface{}{"memory": "2Gi"},
		OvercommitGuestOverhead: true,
	})
	require.Len(t, got, 1)
	m := got[0].(map[string]interface{})
	assert.True(t, m["over_commit_guest_overhead"].(bool))
}

func TestFlattenCPUFromVM(t *testing.T) {
	// nil returns a slice with an empty map.
	got := flattenCPUFromVM(nil)
	require.Len(t, got, 1)

	// Populated preserves fields.
	got = flattenCPUFromVM(&models.V1VMCPU{Cores: 4, Sockets: 2, Threads: 1})
	m := got[0].(map[string]interface{})
	assert.EqualValues(t, 4, m["cores"])
	assert.EqualValues(t, 2, m["sockets"])
	assert.EqualValues(t, 1, m["threads"])
}

func TestFlattenMemoryFromVM(t *testing.T) {
	got := flattenMemoryFromVM(nil)
	require.Len(t, got, 1)

	got = flattenMemoryFromVM(&models.V1VMMemory{
		Guest:     "1Gi",
		Hugepages: &models.V1VMHugepages{PageSize: "2Mi"},
	})
	m := got[0].(map[string]interface{})
	assert.Equal(t, "1Gi", m["guest"])
	assert.Equal(t, "2Mi", m["hugepages"])
}

func TestFlattenFirmwareFromVM(t *testing.T) {
	assert.Empty(t, flattenFirmwareFromVM(nil))

	// Server-generated firmware (uuid/serial only, no bootloader) → drop.
	got := flattenFirmwareFromVM(&models.V1VMFirmware{UUID: "abc", Serial: "S1"})
	assert.Empty(t, got)

	// With bootloader → persist.
	got = flattenFirmwareFromVM(&models.V1VMFirmware{
		UUID:       "abc",
		Serial:     "S1",
		Bootloader: &models.V1VMBootloader{Bios: &models.V1VMBIOS{UseSerial: true}},
	})
	require.Len(t, got, 1)
	m := got[0].(map[string]interface{})
	assert.Equal(t, "abc", m["uuid"])
	assert.Equal(t, "S1", m["serial"])
	assert.Contains(t, m, "bootloader")
}

func TestFlattenBootloaderFromVM(t *testing.T) {
	assert.Empty(t, flattenBootloaderFromVM(nil))

	got := flattenBootloaderFromVM(&models.V1VMBootloader{
		Bios: &models.V1VMBIOS{UseSerial: true},
		Efi:  &models.V1VMEFI{SecureBoot: boolP(true)},
	})
	require.Len(t, got, 1)
	m := got[0].(map[string]interface{})
	assert.Contains(t, m, "bios")
	assert.Contains(t, m, "efi")
}

func TestFlattenBIOSFromVM(t *testing.T) {
	assert.Empty(t, flattenBIOSFromVM(nil))
	got := flattenBIOSFromVM(&models.V1VMBIOS{UseSerial: true})
	require.Len(t, got, 1)
	m := got[0].(map[string]interface{})
	assert.True(t, m["use_serial"].(bool))
}

func TestFlattenEFIFromVM(t *testing.T) {
	assert.Empty(t, flattenEFIFromVM(nil))

	// All nil sub-fields → empty map, which is dropped.
	got := flattenEFIFromVM(&models.V1VMEFI{})
	assert.Empty(t, got)

	sb, per := true, false
	got = flattenEFIFromVM(&models.V1VMEFI{SecureBoot: &sb, Persistent: &per})
	require.Len(t, got, 1)
	m := got[0].(map[string]interface{})
	assert.True(t, m["secure_boot"].(bool))
	assert.False(t, m["persistent"].(bool))
}

func TestFlattenFeaturesFromVM(t *testing.T) {
	assert.Empty(t, flattenFeaturesFromVM(nil))

	got := flattenFeaturesFromVM(&models.V1VMFeatures{
		Acpi: &models.V1VMFeatureState{Enabled: true},
		Apic: &models.V1VMFeatureAPIC{Enabled: false},
		Smm:  &models.V1VMFeatureState{Enabled: true},
	})
	require.Len(t, got, 1)
	m := got[0].(map[string]interface{})
	assert.Contains(t, m, "acpi")
	assert.Contains(t, m, "apic")
	assert.Contains(t, m, "smm")
}

func TestFlattenDevicesFromVM(t *testing.T) {
	got := flattenDevicesFromVM(nil)
	require.Len(t, got, 1)

	name := "disk0"
	got = flattenDevicesFromVM(&models.V1VMDevices{
		Disks: []*models.V1VMDisk{{Name: &name, Disk: &models.V1VMDiskTarget{Bus: "virtio"}}},
	})
	require.Len(t, got, 1)
	m := got[0].(map[string]interface{})
	assert.NotNil(t, m["disk"])
}

func TestFlattenDisksFromVM(t *testing.T) {
	assert.Nil(t, flattenDisksFromVM(nil))

	name := "disk0"
	got := flattenDisksFromVM([]*models.V1VMDisk{
		nil, // skipped
		{Name: &name, Serial: "S1", BootOrder: 2, Disk: &models.V1VMDiskTarget{Bus: "virtio"}},
	})
	// Both slots populated; nil entry becomes zero-value map.
	require.Len(t, got, 2)
}

func TestFlattenVMDiskDevice(t *testing.T) {
	got := flattenVMDiskDevice(nil)
	require.Len(t, got, 1)

	got = flattenVMDiskDevice(&models.V1VMDisk{
		Disk:  &models.V1VMDiskTarget{Bus: "virtio", Readonly: true, PciAddress: "0000:81:01.1"},
		Cdrom: &models.V1VMCDRomTarget{Bus: "sata"},
		Lun:   &models.V1VMLunTarget{Bus: "scsi"},
	})
	require.Len(t, got, 1)
	m := got[0].(map[string]interface{})
	assert.Contains(t, m, "disk")
	assert.Contains(t, m, "cdrom")
	assert.Contains(t, m, "lun")
}

func TestFlattenInterfacesFromVM(t *testing.T) {
	assert.Nil(t, flattenInterfacesFromVM(nil))

	name := "eth0"
	got := flattenInterfacesFromVM([]*models.V1VMInterface{
		nil, // skipped inside loop; leaves nil in slot
		{Name: &name, Model: "virtio", Bridge: struct{}{}},
	})
	require.Len(t, got, 2)
	// The second entry is the populated interface.
	m := got[1].(map[string]interface{})
	assert.Equal(t, "eth0", m["name"])
	assert.Equal(t, "virtio", m["model"])
	assert.Equal(t, "InterfaceBridge", m["interface_binding_method"])
}

// TestFlattenVMInterfaceBindingMethod exhaustively pins every branch of
// the interface binding switch.
func TestFlattenVMInterfaceBindingMethod(t *testing.T) {
	assert.Equal(t, "", flattenVMInterfaceBindingMethod(nil))
	assert.Equal(t, "InterfaceBridge", flattenVMInterfaceBindingMethod(&models.V1VMInterface{Bridge: struct{}{}}))
	assert.Equal(t, "InterfaceMasquerade", flattenVMInterfaceBindingMethod(&models.V1VMInterface{Masquerade: struct{}{}}))
	assert.Equal(t, "InterfaceSlirp", flattenVMInterfaceBindingMethod(&models.V1VMInterface{Slirp: struct{}{}}))
	assert.Equal(t, "InterfaceSRIOV", flattenVMInterfaceBindingMethod(&models.V1VMInterface{Sriov: struct{}{}}))
	assert.Equal(t, "macvtap", flattenVMInterfaceBindingMethod(&models.V1VMInterface{Macvtap: struct{}{}}))
	assert.Equal(t, "passt", flattenVMInterfaceBindingMethod(&models.V1VMInterface{Passt: struct{}{}}))
	assert.Equal(t, "", flattenVMInterfaceBindingMethod(&models.V1VMInterface{}))
}

// ---------------------------------------------------------------------------
// FlattenDomainSpecFromVM + ExpandDomainSpec — outer entry points.
// ---------------------------------------------------------------------------

func TestFlattenDomainSpecFromVM(t *testing.T) {
	t.Run("nil returns slice with empty map", func(t *testing.T) {
		got := FlattenDomainSpecFromVM(nil)
		require.Len(t, got, 1)
	})

	t.Run("fully populated", func(t *testing.T) {
		name := "disk0"
		got := FlattenDomainSpecFromVM(&models.V1VMDomainSpec{
			Resources: &models.V1VMResourceRequirements{OvercommitGuestOverhead: true},
			CPU:       &models.V1VMCPU{Cores: 2},
			Memory:    &models.V1VMMemory{Guest: "1Gi"},
			Firmware:  &models.V1VMFirmware{Bootloader: &models.V1VMBootloader{Bios: &models.V1VMBIOS{UseSerial: true}}},
			Features:  &models.V1VMFeatures{Acpi: &models.V1VMFeatureState{Enabled: true}},
			Devices:   &models.V1VMDevices{Disks: []*models.V1VMDisk{{Name: &name}}},
		})
		require.Len(t, got, 1)
		m := got[0].(map[string]interface{})
		assert.Contains(t, m, "resources")
		assert.Contains(t, m, "cpu")
		assert.Contains(t, m, "memory")
		assert.Contains(t, m, "firmware")
		assert.Contains(t, m, "features")
		assert.Contains(t, m, "devices")
	})
}

// domainSpecTestSchema is a minimal top-level schema that mirrors the
// keys ExpandDomainSpec / expandDevicesToVM read via d.GetOk. Fields are
// intentionally permissive — we only need Set() to succeed, not to
// enforce validation.
func domainSpecTestSchema() *schema.Resource {
	return &schema.Resource{Schema: map[string]*schema.Schema{
		"resources": {Type: schema.TypeList, Optional: true, Elem: &schema.Resource{Schema: map[string]*schema.Schema{
			"requests":                   {Type: schema.TypeMap, Optional: true, Elem: &schema.Schema{Type: schema.TypeString}},
			"limits":                     {Type: schema.TypeMap, Optional: true, Elem: &schema.Schema{Type: schema.TypeString}},
			"over_commit_guest_overhead": {Type: schema.TypeBool, Optional: true},
		}}},
		"cpu": {Type: schema.TypeList, Optional: true, Elem: &schema.Resource{Schema: map[string]*schema.Schema{
			"cores":   {Type: schema.TypeInt, Optional: true},
			"sockets": {Type: schema.TypeInt, Optional: true},
			"threads": {Type: schema.TypeInt, Optional: true},
		}}},
		"memory": {Type: schema.TypeList, Optional: true, Elem: &schema.Resource{Schema: map[string]*schema.Schema{
			"guest":     {Type: schema.TypeString, Optional: true},
			"hugepages": {Type: schema.TypeString, Optional: true},
		}}},
		"firmware": {Type: schema.TypeList, Optional: true, Elem: &schema.Resource{Schema: map[string]*schema.Schema{
			"uuid":       {Type: schema.TypeString, Optional: true},
			"serial":     {Type: schema.TypeString, Optional: true},
			"bootloader": {Type: schema.TypeList, Optional: true, Elem: &schema.Resource{Schema: map[string]*schema.Schema{}}},
		}}},
		"features": {Type: schema.TypeList, Optional: true, Elem: &schema.Resource{Schema: map[string]*schema.Schema{}}},
		"disk": {Type: schema.TypeList, Optional: true, Elem: &schema.Resource{Schema: map[string]*schema.Schema{
			"name":        {Type: schema.TypeString, Optional: true},
			"serial":      {Type: schema.TypeString, Optional: true},
			"boot_order":  {Type: schema.TypeInt, Optional: true},
			"disk_device": {Type: schema.TypeList, Optional: true, Elem: &schema.Resource{Schema: map[string]*schema.Schema{}}},
		}}},
		"interface": {Type: schema.TypeList, Optional: true, Elem: &schema.Resource{Schema: map[string]*schema.Schema{
			"name":                     {Type: schema.TypeString, Optional: true},
			"model":                    {Type: schema.TypeString, Optional: true},
			"interface_binding_method": {Type: schema.TypeString, Optional: true},
		}}},
	}}
}

func TestExpandDomainSpec(t *testing.T) {
	t.Run("empty ResourceData", func(t *testing.T) {
		d := domainSpecTestSchema().TestResourceData()
		got, err := ExpandDomainSpec(d)
		require.NoError(t, err)
		require.NotNil(t, got)
		// Devices is always initialized (empty).
		require.NotNil(t, got.Devices)
	})

	t.Run("populated ResourceData", func(t *testing.T) {
		d := domainSpecTestSchema().TestResourceData()
		require.NoError(t, d.Set("resources", []interface{}{
			map[string]interface{}{"over_commit_guest_overhead": true},
		}))
		require.NoError(t, d.Set("cpu", []interface{}{
			map[string]interface{}{"cores": 4, "sockets": 2, "threads": 1},
		}))
		require.NoError(t, d.Set("memory", []interface{}{
			map[string]interface{}{"guest": "1Gi"},
		}))
		require.NoError(t, d.Set("firmware", []interface{}{
			map[string]interface{}{"uuid": "abc", "serial": "S1"},
		}))
		require.NoError(t, d.Set("disk", []interface{}{
			map[string]interface{}{"name": "disk0", "boot_order": 1},
		}))
		require.NoError(t, d.Set("interface", []interface{}{
			map[string]interface{}{"name": "eth0", "interface_binding_method": "InterfaceBridge"},
		}))

		got, err := ExpandDomainSpec(d)
		require.NoError(t, err)
		require.NotNil(t, got.Resources)
		assert.True(t, got.Resources.OvercommitGuestOverhead)
		require.NotNil(t, got.CPU)
		assert.EqualValues(t, 4, got.CPU.Cores)
		require.NotNil(t, got.Memory)
		assert.Equal(t, models.V1VMQuantity("1Gi"), got.Memory.Guest)
		require.NotNil(t, got.Firmware)
		assert.Equal(t, "abc", got.Firmware.UUID)
		require.NotNil(t, got.Devices)
		require.Len(t, got.Devices.Disks, 1)
		require.Len(t, got.Devices.Interfaces, 1)
	})
}

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

func boolP(b bool) *bool { return &b }
