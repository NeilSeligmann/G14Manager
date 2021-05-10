package thermal

import "log"

// GetDefaultThermalProfiles will return the default list of Profiles
func GetDefaultThermalProfiles() []Profile {
	defaultProfiles := make([]Profile, 0, 3)
	defaults := []struct {
		name             string
		windowsPowerPlan string
		throttlePlan     uint32
		cpuFanCurve      string
		gpuFanCurve      string
	}{
		{
			name:             "Fanless",
			windowsPowerPlan: "Balanced",
			throttlePlan:     ThrottlePlanPerformance,
			cpuFanCurve:      "39c:0%,49c:0%,59c:0%,69c:0%,79c:31%,89c:49%,99c:56%,109c:56%",
			gpuFanCurve:      "39c:0%,49c:0%,59c:0%,69c:0%,79c:34%,89c:51%,99c:61%,109c:61%",
		},
		{
			name:             "Quiet",
			windowsPowerPlan: "Balanced",
			throttlePlan:     ThrottlePlanPerformance,
		},
		{
			name:             "Balanced",
			windowsPowerPlan: "Balanced",
			throttlePlan:     ThrottlePlanSilent,
			cpuFanCurve:      "20c:0%,50c:10%,55c:10%,60c:10%,65c:31%,70c:49%,75c:56%,98c:56%",
			gpuFanCurve:      "20c:0%,50c:10%,55c:10%,60c:10%,65c:34%,70c:51%,75c:61%,98c:61%",
		},
		{
			name:             "Performance",
			windowsPowerPlan: "High performance",
			throttlePlan:     ThrottlePlanPerformance,
			cpuFanCurve:      "20c:10%,50c:20%,55c:25%,60c:40%,65c:45%,70c:55%,75c:90%,98c:100%",
			gpuFanCurve:      "20c:10%,50c:20%,55c:25%,60c:40%,65c:45%,70c:55%,75c:90%,98c:100%",
		},
		{
			name:             "Turbo",
			windowsPowerPlan: "High performance",
			throttlePlan:     ThrottlePlanTurbo,
			cpuFanCurve:      "20c:10%,40c:25%,50c:30%,60c:80%,65c:90%,70c:100%,75c:100%,98c:100%",
			gpuFanCurve:      "20c:10%,40c:35%,50c:30%,60c:80%,65c:90%,70c:100%,75c:100%,98c:100%",
		},
	}
	for _, d := range defaults {
		var cpuTable, gpuTable *FanTable
		var err error
		profile := Profile{
			Name:             d.name,
			ThrottlePlan:     d.throttlePlan,
			WindowsPowerPlan: d.windowsPowerPlan,
		}

		if d.cpuFanCurve != "" {
			cpuTable, err = NewFanTable(d.cpuFanCurve)
			if err != nil {
				panic(err)
			}
			profile.CPUFanCurve = cpuTable
		}

		if d.gpuFanCurve != "" {
			gpuTable, err = NewFanTable(d.gpuFanCurve)
			if err != nil {
				panic(err)
			}
			profile.GPUFanCurve = gpuTable
		}
		defaultProfiles = append(defaultProfiles, profile)
	}

	log.Print(defaultProfiles)

	return defaultProfiles
}
