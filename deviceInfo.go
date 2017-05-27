// deviceInfo.go
package main

import "fmt"

const (
	MAX_DEVICES = 0100
	// Standard device codes in octal, Priority Mask Bits in decimal
	// as per DG docs!
	// N.B. add to deviceToString() when new codes added here
	DEV_PSC  = 004
	PSC_PMB  = 13
	DEV_TTI  = 010
	TTI_PMB  = 14
	DEV_TTO  = 011
	TTO_PMB  = 15
	DEV_RTC  = 014
	RTC_PMB  = 13
	DEV_MTB  = 022
	MTB_PMB  = 10
	DEV_DSKP = 024
	DSKP_PMB = 7
	DEV_DPF  = 027
	DPF_PMB  = 7
	DEV_SCP  = 045
	SCP_PMB  = 15
	DEV_CPU  = 077
	CPU_PMB  = 0 // kinda!

	CPU_MODEL_NO = 0x224C
	UCODE_REV    = 0x04
)

func deviceToString(devNum int) string {
	var ds string
	switch devNum {
	case DEV_CPU:
		return "CPU"
	case DEV_DSKP:
		return "DSKP"
	case DEV_DPF:
		return "DPF"
	case DEV_MTB:
		return "MTB"
		//	case DEV_PIT:
		//		return "PIT"
	case DEV_PSC:
		return "PSC"
	case DEV_RTC:
		return "RTC"
	case DEV_SCP:
		return "SCP"
	case DEV_TTI:
		return "TTI"
	case DEV_TTO:
		return "TTO"
	default:
		ds = fmt.Sprintf("%#o", devNum)
	}
	return ds
}
