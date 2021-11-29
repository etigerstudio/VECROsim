package faults

import apiv1 "k8s.io/api/core/v1"

func NewPumbaPod() *apiv1.PodSpec {
	return &apiv1.PodSpec{
		RestartPolicy: apiv1.RestartPolicyNever,
		Volumes: []apiv1.Volume{
			{
				Name: "dockersocket",
				VolumeSource: apiv1.VolumeSource{
					HostPath: &apiv1.HostPathVolumeSource{
						Path: "/var/run/docker.sock",
					},
				},
			},
		},
	}
}

func newPumbaContainer(containerName string,
	args []string,
	addCapabilities []apiv1.Capability) *apiv1.Container {
	return &apiv1.Container{
		Name:            containerName,
		Image:           "gaiaadm/pumba",
		ImagePullPolicy: apiv1.PullIfNotPresent,
		Args:            args,
		VolumeMounts: []apiv1.VolumeMount{
			{
				Name:      "dockersocket",
				MountPath: "/var/run/docker.sock",
			},
		},
		SecurityContext: &apiv1.SecurityContext{
			Capabilities: &apiv1.Capabilities{
				Add: addCapabilities,
			},
		},
	}
}

func appendContainer(pod *apiv1.PodSpec, container apiv1.Container) {
	pod.Containers = append(pod.Containers, container)
}
