package convert

import (
	"github.com/spectrocloud/hapi/models"
	kubevirtapiv1 "kubevirt.io/api/core/v1"
)

// TODO: implement clock.

func ToHapiVmClock(clock *kubevirtapiv1.Clock) *models.V1VMClock {
	if clock == nil {
		return nil
	}

	var Timezone string
	if clock.Timezone != nil {
		Timezone = string(*clock.Timezone)
	}

	return &models.V1VMClock{
		Timer:    ToHapiVmTimer(clock.Timer),
		Timezone: Timezone,
		Utc:      ToHapiVmClockOffest(clock.UTC),
	}
}

func ToHapiVmClockOffest(utc *kubevirtapiv1.ClockOffsetUTC) *models.V1VMClockOffsetUTC {
	if utc == nil {
		return nil
	}

	return &models.V1VMClockOffsetUTC{
		OffsetSeconds: int32(*utc.OffsetSeconds),
	}
}

func ToHapiVmTimer(timer *kubevirtapiv1.Timer) *models.V1VMTimer {
	if timer == nil {
		return nil
	}

	return &models.V1VMTimer{
		Hpet:   ToHapiVmHpet(timer.HPET),
		Hyperv: ToHapiVmHyperv(timer.Hyperv),
		Kvm:    ToHapiVmKvm(timer.KVM),
		Pit:    ToHapiVmPit(timer.PIT),
		Rtc:    ToHapiVmRtc(timer.RTC),
	}
}

func ToHapiVmRtc(rtc *kubevirtapiv1.RTCTimer) *models.V1VMRTCTimer {
	if rtc == nil {
		return nil
	}

	return &models.V1VMRTCTimer{
		// TODO
	}
}

func ToHapiVmPit(pit *kubevirtapiv1.PITTimer) *models.V1VMPITTimer {
	if pit == nil {
		return nil
	}

	return &models.V1VMPITTimer{
		// TODO
	}
}

func ToHapiVmKvm(kvm *kubevirtapiv1.KVMTimer) *models.V1VMKVMTimer {
	if kvm == nil {
		return nil
	}

	return &models.V1VMKVMTimer{
		Present: *kvm.Enabled,
	}
}

func ToHapiVmHyperv(hyperv *kubevirtapiv1.HypervTimer) *models.V1VMHypervTimer {
	if hyperv == nil {
		return nil
	}

	return &models.V1VMHypervTimer{
		Present: *hyperv.Enabled,
	}
}

func ToHapiVmHpet(hpet *kubevirtapiv1.HPETTimer) *models.V1VMHPETTimer {
	if hpet == nil {
		return nil
	}

	return &models.V1VMHPETTimer{
		Present:    *hpet.Enabled,
		TickPolicy: string(hpet.TickPolicy),
	}
}
