package base

import (
	v1 "k8s.io/api/core/v1"
	"strconv"
)

const (
	workloadCPUEnvKey           = "BEN_WORKLOAD_CPU"
	workloadIOEnvKey            = "BEN_WORKLOAD_IO"
	workloadDelayDurationEnvKey = "BEN_WORKLOAD_DELAY_DURATION"
	workloadDelayJitterEnvKey   = "BEN_WORKLOAD_DELAY_JITTER"
	workloadNetEnvKey           = "BEN_WORKLOAD_NET"
)

func (w Workload) toWorkloadEnvVar() []v1.EnvVar {
	return []v1.EnvVar{
		{
			Name:  workloadCPUEnvKey,
			Value: strconv.Itoa(w.CPU),
		},
		{
			Name:  workloadIOEnvKey,
			Value: strconv.Itoa(w.IO),
		},
		{
			Name:  workloadDelayDurationEnvKey,
			Value: strconv.Itoa(w.Delay.Duration),
		},
		{
			Name:  workloadDelayJitterEnvKey,
			Value: strconv.Itoa(w.Delay.Jitter),
		},
		{
			Name:  workloadNetEnvKey,
			Value: strconv.Itoa(w.Net),
		},
	}
}
