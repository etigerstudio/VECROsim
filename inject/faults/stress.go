package faults

import (
	"fmt"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func AddCPUStress(pod *apiv1.PodSpec,
	name string,
	target string,
	load int,
	method string,
	duration metav1.Duration) {
	appendContainer(pod, *newPumbaContainer(name+"-cpu-stress", []string{
		"--log-level",
		"info",
		"--label",
		fmt.Sprintf("io.kubernetes.container.name=%s", target),
		"stress",
		"--duration",
		duration.String(),
		"--stressors",
		fmt.Sprintf("\"--cpus 1 --cpu-load %d --cpu-method %s\"", load, method),
	}, nil))
}

func AddIOStress(pod *apiv1.PodSpec,
	name string,
	target string,
	method string,
	duration metav1.Duration) {
	var stressors string
	if method == "sync" {
		stressors = fmt.Sprintf("\"--io 1\"")
	} else if method == "iomix" {
		stressors = fmt.Sprintf("\"--iomix 1\"")
	} else {
		panic("invalid IO stress method is specified.\nSupported options: sync, iomix.")
	}
	appendContainer(pod, *newPumbaContainer(name+"-io-stress", []string{
		"--log-level",
		"info",
		"--label",
		fmt.Sprintf("io.kubernetes.container.name=%s", target),
		"stress",
		"--duration",
		duration.String(),
		"--stressors",
		stressors,
	}, nil))
}