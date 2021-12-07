package faults

import (
	"fmt"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strconv"
)

const tcImage = "gaiadocker/iproute2"

func AddNetDelay(pod *apiv1.PodSpec,
	name string,
	target string,
	delay metav1.Duration,
	jitter metav1.Duration,
	duration metav1.Duration) {
	appendContainer(pod, *newPumbaContainer(name+"-net-delay", []string{
		"--log-level",
		"info",
		"--label",
		fmt.Sprintf("io.kubernetes.container.name=%s", target),
		"netem",
		"--duration",
		duration.Duration.String(),
		"--tc-image",
		tcImage,
		"delay",
		"--time",
		strconv.FormatInt(delay.Milliseconds(), 10),
		"--jitter",
		strconv.FormatInt(jitter.Milliseconds(), 10),
	}, []apiv1.Capability{
		"NET_ADMIN",
	}))
}

func AddNetLoss(pod *apiv1.PodSpec,
	name string,
	target string,
	percent int,
	duration metav1.Duration) {
	appendContainer(pod, *newPumbaContainer(name+"-net-loss", []string{
		"--log-level",
		"info",
		"--label",
		fmt.Sprintf("io.kubernetes.container.name=%s", target),
		"netem",
		"--duration",
		duration.Duration.String(),
		"--tc-image",
		tcImage,
		"loss",
		"--percent",
		strconv.Itoa(percent),
	}, []apiv1.Capability{
		"NET_ADMIN",
	}))
}

func AddNetRate(pod *apiv1.PodSpec,
	name string,
	target string,
    rate string,
	duration metav1.Duration) {
	appendContainer(pod, *newPumbaContainer(name+"-net-rate", []string{
		"--log-level",
		"info",
		"--label",
		fmt.Sprintf("io.kubernetes.container.name=%s", target),
		"netem",
		"--duration",
		duration.Duration.String(),
		"--tc-image",
		tcImage,
		"rate",
		"--rate",
		rate,
	}, []apiv1.Capability{
		"NET_ADMIN",
	}))
}