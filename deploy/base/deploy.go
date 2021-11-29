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
const imageName = "ben-base:v1"
const baseListenPort = 8080

const nameEnvKey = "BEN_NAME"
const subsystemEnvKey = "BEN_SUBSYSTEM"
const listenAddressEnvKey = "BEN_LISTEN_ADDRESS"
const calleeEnvKey = "BEN_CALLS"
const calleeSeparator = " "

func prepareSystemDefinition(def *SystemDefinition) {
	// Prepare a map from names to services for faster searching
	def.serviceMap = make(map[string]*Service, len(def.Services))

	for i := 0; i < len(def.Services); i += 1{
		def.Services[i].id = i
		def.Services[i].Port = baseListenPort + i
		def.serviceMap[def.Services[i].Name] = &def.Services[i]
	}
}

func prepareDeployment(def SystemDefinition) *appsv1.Deployment {
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: def.Name,
			Namespace: def.Namespace,
			Labels: map[string]string{
				"app.kubernetes.io/name":       def.Name,
				"app.kubernetes.io/managed-by": labelManagedBy,
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32Ptr(def.Replicas),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app.kubernetes.io/name": def.Name,
				},
			},
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app.kubernetes.io/name":       def.Name,
						"app.kubernetes.io/managed-by": labelManagedBy,
					},
				},
				Spec: apiv1.PodSpec{
					Containers: make([]apiv1.Container, len(def.Services)),
				},
			},
		},
		Status: appsv1.DeploymentStatus{},
	}

	for i, svc := range def.Services {
		container := apiv1.Container{
			Name:  svc.Name,
			Image: imageName,
			Ports: []apiv1.ContainerPort{
				{
					// Currently there's no need to set a port name
					//Name:          svc.Name + "-port",
					ContainerPort: int32(svc.Port),
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
					Value: svc.Subsystem,
				},
				{
					Name:  calleeEnvKey,
					Value: assembleCalls(svc.Calls, def.Name, def.serviceMap),
				},
				{
					Name:  listenAddressEnvKey,
					Value: ":" + strconv.Itoa(svc.Port),
				},
			},
			Resources: apiv1.ResourceRequirements{
				Limits:   apiv1.ResourceList{
					apiv1.ResourceCPU: resource.MustParse("100m"),
				},
				Requests: apiv1.ResourceList{
					apiv1.ResourceCPU: resource.MustParse("10m"),
				},
			},
		}

		// Assemble service workload config to containers
		container.Env = append(container.Env, svc.toWorkloadEnvVar()...)
		deployment.Spec.Template.Spec.Containers[i] = container
	}

	return deployment
}

func commitDeployment(deploymentsClient clientappsv1.DeploymentInterface, deployment *appsv1.Deployment) (*appsv1.Deployment, error) {
	return deploymentsClient.Create(context.TODO(), deployment, metav1.CreateOptions{})
}

func createDeployment(clientset *kubernetes.Clientset, def SystemDefinition) {
	deployment := prepareDeployment(def)

	//fmt.Printf("%#v\n", deployment)
	deploymentsClient := clientset.AppsV1().Deployments(def.Namespace)
	result, err := commitDeployment(deploymentsClient, deployment)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Created deployment %q.\n", result.GetObjectMeta().GetName())
}

func prepareService(def SystemDefinition) *apiv1.Service {
	service := &apiv1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      def.Name,
			Namespace: def.Namespace,
			Annotations: map[string]string{
				"prometheus.io/scrape": "true", // For Prometheus to scrape metrics
			},
			Labels: map[string]string{
				"app.kubernetes.io/name":       def.Name,
				"app.kubernetes.io/managed-by": labelManagedBy,
			},
		},
		Spec: apiv1.ServiceSpec{
			Ports: make([]apiv1.ServicePort, len(def.Services)),
			Selector: map[string]string{
				"app.kubernetes.io/name": def.Name,
			},
			Type: "ClusterIP",
		},
	}

	for i, svc := range def.Services {
		service.Spec.Ports[i] = apiv1.ServicePort{
			// Service names serve as endpoint names
			Name:       svc.Name,
			Protocol:   "TCP",
			Port:       int32(def.serviceMap[svc.Name].Port),
			TargetPort: intstr.FromInt(def.serviceMap[svc.Name].Port),
		}
	}

	return service
}

func commitService(serviceClient clientcorev1.ServiceInterface, service *apiv1.Service) (*apiv1.Service, error) {
	return serviceClient.Create(context.TODO(), service, metav1.CreateOptions{})
}

func createService(clientset *kubernetes.Clientset, def SystemDefinition) {
	service := prepareService(def)

	//fmt.Printf("%#v\n", service)
	serviceClient := clientset.CoreV1().Services(def.Namespace)
	result, err := commitService(serviceClient, service)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Created service %q.\n", result.GetObjectMeta().GetName())
}

func assembleCalls(calls []string, systemName string, serviceMap map[string]*Service) string {
	if len(calls) == 0 {
		return ""
	}

	urls := make([]string, len(calls))
	for i, callee := range calls {
		//"http://info-service.app.svc.cluster.local/info"
		//"http://service-name.namespace.svc.cluster.local:port"
		urls[i] = fmt.Sprintf("http://%s:%d", systemName, serviceMap[callee].Port)
	}

	return strings.Join(urls, calleeSeparator)
}

func CreateResources(clientset *kubernetes.Clientset, def SystemDefinition) {
	prepareSystemDefinition(&def)
	fmt.Printf("Creating deployment...")
	createDeployment(clientset, def)
	fmt.Printf("Done.\nCreating service...")
	createService(clientset, def)
	fmt.Printf("Done.\n")

	// TODO: Create Prometheus Resource.
}

func int32Ptr(i int32) *int32 { return &i }
