package base

import (
	"context"
	"fmt"
	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
	clientappsv1 "k8s.io/client-go/kubernetes/typed/apps/v1"
	clientcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"

	"strconv"
	"strings"
)

const labelManagedBy = "ben-sim"
const baseImageName = "ben-base:v1"
const mongoDBImageName = "ben-mongodb:v1"
const baseListeningPort = 8080
const baseExposedPort = 80

const nameEnvKey = "BEN_NAME"
const subsystemEnvKey = "BEN_SUBSYSTEM"
const listenAddressEnvKey = "BEN_LISTEN_ADDRESS"
const calleeEnvKey = "BEN_CALLS"
const calleeSeparator = " "

const benServiceName = "ben-sim/service-name"
const benServiceID = "ben-sim/service-id"

func prepareSystemDefinition(def *SystemDefinition) {
	// Currently there's no need for preparation
}

func prepareDeployments(def SystemDefinition) []*appsv1.Deployment {
	deployments := make([]*appsv1.Deployment, len(def.Services))
	for i, svc := range def.Services {
		deployment := &appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      def.Name + "-" + svc.Name,
				Namespace: def.Namespace,
				Labels: map[string]string{
					"app.kubernetes.io/name":       def.Name,
					"app.kubernetes.io/managed-by": labelManagedBy,
					benServiceName:                 svc.Name,
					benServiceID:                   strconv.Itoa(i),
				},
			},
			Spec: appsv1.DeploymentSpec{
				Replicas: int32Ptr(def.Replicas),
				Selector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"app.kubernetes.io/name":       def.Name,
						"app.kubernetes.io/managed-by": labelManagedBy,
						benServiceName:                 svc.Name,
					},
				},
				Template: apiv1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							"app.kubernetes.io/name":       def.Name,
							"app.kubernetes.io/managed-by": labelManagedBy,
							benServiceName:                 svc.Name,
							benServiceID:                   strconv.Itoa(i),
						},
					},
					Spec: apiv1.PodSpec{
						Containers: prepareContainers(svc, def.Name),
						Volumes: prepareVolumes(svc),
					},
				},
			},
			Status: appsv1.DeploymentStatus{},
		}

		deployments[i] = deployment
	}

	return deployments
}

func prepareVolumes(svc Service) []apiv1.Volume {
	volumes := make([]apiv1.Volume, 0)
	switch svc.Type {
	case "base":
		volume := apiv1.Volume{
			Name:         "tmp-io-dir",
			VolumeSource: apiv1.VolumeSource{
				EmptyDir: &apiv1.EmptyDirVolumeSource{},
			},
		}
		volumes = append(volumes, volume)

	case "mongodb":
		volume := apiv1.Volume{
			Name:         "init-script",
			VolumeSource: apiv1.VolumeSource{
				ConfigMap: &apiv1.ConfigMapVolumeSource{
					LocalObjectReference: apiv1.LocalObjectReference{
						Name: "mongo-initjs",
					},
				},
			},
		}
		volumes = append(volumes, volume)
	}
	
	return volumes
}

func prepareContainers(svc Service, sysName string) []apiv1.Container {
	containers := make([]apiv1.Container, 0)

	switch svc.Type {
	case "base":
		container := apiv1.Container{
			Name:  svc.Name,
			Image: baseImageName,
			Ports: []apiv1.ContainerPort{
				{
					Name:          svc.Name + "-port",
					ContainerPort: int32(baseListeningPort),
					Protocol:      apiv1.ProtocolTCP,
				},
			},
			Env: []apiv1.EnvVar{
				{
					Name:  nameEnvKey,
					Value: svc.Name,
				},
				{
					Name:  subsystemEnvKey,
					Value: sysName,
				},
				{
					Name:  calleeEnvKey,
					Value: assembleCalls(svc.Calls, sysName),
				},
				{
					Name:  listenAddressEnvKey,
					Value: ":" + strconv.Itoa(baseListeningPort),
				},
			},
			VolumeMounts: []apiv1.VolumeMount{
				{
					Name:      "tmp-io-dir",
					MountPath: "/tmp/ben-base-io",
				},
			},
			Resources: apiv1.ResourceRequirements{
				Limits: apiv1.ResourceList{
					apiv1.ResourceCPU: resource.MustParse("1000m"), // TODO: make cpu resource request & limit configurable
				},
				Requests: apiv1.ResourceList{
					apiv1.ResourceCPU: resource.MustParse("200m"),
				},
			},
		}

		// Assemble service workload config to base container
		container.Env = append(container.Env, svc.toWorkloadEnvVar()...)
		containers = append(containers, container)

	case "mongodb":
		baseContainer := apiv1.Container{
			Name:  svc.Name + "-agent",
			Image: mongoDBImageName,
			Ports: []apiv1.ContainerPort{
				{
					Name:          svc.Name + "-port",
					ContainerPort: int32(baseListeningPort),
					Protocol:      apiv1.ProtocolTCP,
				},
			},
			Env: []apiv1.EnvVar{
				{
					Name:  nameEnvKey,
					Value: svc.Name,
				},
				{
					Name:  subsystemEnvKey,
					Value: sysName,
				},
				{
					Name:  listenAddressEnvKey,
					Value: ":" + strconv.Itoa(baseListeningPort),
				},
			},
			Resources: apiv1.ResourceRequirements{
				Limits: apiv1.ResourceList{
					apiv1.ResourceCPU: resource.MustParse("1000m"), // TODO: make cpu resource request & limit configurable
				},
				Requests: apiv1.ResourceList{
					apiv1.ResourceCPU: resource.MustParse("100m"),
				},
			},
		}

		mongoDBContainer := apiv1.Container{
			Name:  svc.Name + "-mongodb",
			Image: "mongo",
			Ports: []apiv1.ContainerPort{
				{
					Name:          "mongodb-port",
					ContainerPort: 27017,
					Protocol:      apiv1.ProtocolTCP,
				},
			},
			Env: []apiv1.EnvVar{
				{
					Name:  "MONGO_INITDB_ROOT_USERNAME",
					Value: "root",
				},
				{
					Name:  "MONGO_INITDB_ROOT_PASSWORD",
					Value: "password",
				},
			},
			VolumeMounts: []apiv1.VolumeMount{
				{
					Name:      "init-script",
					MountPath: "/docker-entrypoint-initdb.d/mongo-init.js",
				},
			},
			Resources: apiv1.ResourceRequirements{
				Limits: apiv1.ResourceList{
					apiv1.ResourceCPU: resource.MustParse("1000m"), // TODO: make cpu resource request & limit configurable
				},
				Requests: apiv1.ResourceList{
					apiv1.ResourceCPU: resource.MustParse("250m"),
				},
			},
		}
		
		// Assemble service workload config to mongodb container
		baseContainer.Env = append(baseContainer.Env, svc.toWorkloadEnvVar()...)
		containers = append(containers, baseContainer, mongoDBContainer)
	}

	return containers
}

func commitDeployment(deploymentsClient clientappsv1.DeploymentInterface, deployment *appsv1.Deployment) (*appsv1.Deployment, error) {
	return deploymentsClient.Create(context.TODO(), deployment, metav1.CreateOptions{})
}

func createDeployment(clientset *kubernetes.Clientset, def SystemDefinition) {
	deployments := prepareDeployments(def)

	//fmt.Printf("%#v\n", deployments)
	deploymentsClient := clientset.AppsV1().Deployments(def.Namespace)
	for i, deployment := range deployments {
		result, err := commitDeployment(deploymentsClient, deployment)
		if err != nil {
			panic(err)
		}
		fmt.Printf("- Created deployment %d: %q.\n", i, result.GetObjectMeta().GetName())
	}
	fmt.Printf("Created deployments for %q.\n", def.Name)
}

func prepareServices(def SystemDefinition) []*apiv1.Service {
	services := make([]*apiv1.Service, len(def.Services))

	for i, svc := range def.Services {
		service := &apiv1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      def.Name + "-" + svc.Name,
				Namespace: def.Namespace,
				//Annotations: map[string]string{
				//	"prometheus.io/scrape": "true", // For Prometheus to scrape metrics
				//},
				Labels: map[string]string{
					"app.kubernetes.io/name":       def.Name,
					"app.kubernetes.io/managed-by": labelManagedBy,
				},
			},
			Spec: apiv1.ServiceSpec{
				Ports: []apiv1.ServicePort{
					{
						// Service names serve as endpoint names
						Name:       svc.Name,
						Protocol:   "TCP",
						Port:       int32(baseExposedPort),
						TargetPort: intstr.FromInt(baseListeningPort),
					},
				},
				Selector: map[string]string{
					"app.kubernetes.io/name":       def.Name,
					"app.kubernetes.io/managed-by": labelManagedBy,
					benServiceName:                 svc.Name,
				},
				Type: "ClusterIP",
			},
		}

		services[i] = service
	}

	return services
}

func commitService(serviceClient clientcorev1.ServiceInterface, service *apiv1.Service) (*apiv1.Service, error) {
	return serviceClient.Create(context.TODO(), service, metav1.CreateOptions{})
}

func createService(clientset *kubernetes.Clientset, def SystemDefinition) {
	services := prepareServices(def)

	//fmt.Printf("%#v\n", service)
	serviceClient := clientset.CoreV1().Services(def.Namespace)
	for i, service := range services {
		result, err := commitService(serviceClient, service)
		if err != nil {
			panic(err)
		}
		fmt.Printf("- Created service %d: %q.\n", i, result.GetObjectMeta().GetName())
	}
	fmt.Printf("Created services for %q.\n", def.Name)
}

func assembleCalls(calls []string, systemName string) string {
	if len(calls) == 0 {
		return ""
	}

	urls := make([]string, len(calls))
	for i, call := range calls {
		//"http://info-service.app.svc.cluster.local/info"
		//"http://service-name.namespace.svc.cluster.local:port"
		urls[i] = fmt.Sprintf("http://%s-%s", systemName, call)
	}

	return strings.Join(urls, calleeSeparator)
}

func CreateResources(clientset *kubernetes.Clientset, def SystemDefinition) {
	// TODO: Create k8s namespace

	prepareSystemDefinition(&def)
	fmt.Printf("Creating deployment...\n")
	createDeployment(clientset, def)
	fmt.Printf("Done.\nCreating service...\n")
	createService(clientset, def)
	fmt.Printf("Done.\n")

	// TODO: Create Prometheus Resource.
}

func int32Ptr(i int32) *int32 { return &i }
