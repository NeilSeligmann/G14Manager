package thermal

type Profile struct {
	Name             string    `json:"name"`
	WindowsPowerPlan string    `json:"windowsPowerPlan"`
	ThrottlePlan     uint32    `json:"throttlePlan"`
	CPUFanCurve      *FanTable `json:"cpuFanCurve"`
	GPUFanCurve      *FanTable `json:"gpuFanCurve"`
	FastSwitch       bool      `json:"fastSwitch"`
}

type ModifyProfileStruct struct {
	ProfileId        int    `json:"profileId"`
	Name             string `json:"name"`
	WindowsPowerPlan string `json:"windowsPowerPlan"`
	ThrottlePlan     uint32 `json:"throttlePlan"`
	CPUFanCurve      string `json:"cpuFanCurve"`
	GPUFanCurve      string `json:"gpuFanCurve"`
	FastSwitch       bool   `json:"fastSwitch"`
}

type MoveProfileStruct struct {
	FromId        int    `json:"fromId"`
	TargetId      int    `json:"targetId"`
}
