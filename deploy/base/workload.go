package base

import (
	v1 "k8s.io/api/core/v1"
	"strconv"
)

const (
	workloadCPUEnvKey           = "VECRO_WORKLOAD_CPU"
	workloadIOEnvKey            = "VECRO_WORKLOAD_IO"
	workloadDelayDurationEnvKey = "VECRO_WORKLOAD_DELAY_DURATION"
	workloadDelayJitterEnvKey   = "VECRO_WORKLOAD_DELAY_JITTER"
	workloadNetEnvKey           = "VECRO_WORKLOAD_NET"
	workloadMemoryEnvKey        = "VECRO_WORKLOAD_MEMORY"
	dbReadOpsEnvKey     		= "VECRO_DB_READ_OPS"
	dbWriteOpsEnvKey    		= "VECRO_DB_WRITE_OPS"
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
		{
			Name:  workloadMemoryEnvKey,
			Value: strconv.Itoa(w.Memory),
		},
		{
			Name:  dbReadOpsEnvKey,
			Value: strconv.Itoa(w.Read),
		},
		{
			Name:  dbWriteOpsEnvKey,
			Value: strconv.Itoa(w.Write),
		},
	}
}
