package thermal

type Profile struct {
	Name             string    `json:"name"`
	WindowsPowerPlan string    `json:"windowsPowerPlan"`
	ThrottlePlan     uint32    `json:"throttlePlan"`
	CPUFanCurve      *FanTable `json:"cpuFanCurve"`
	GPUFanCurve      *FanTable `json:"gpuFanCurve"`
}

// type Marshaler interface {
// 	MarshalJSON() ([]byte, error)
// }
