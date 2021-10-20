package base

import (
	"context"
	"fmt"
	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
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
const serviceTypeEnvKey = "BEN_SERVICE_TYPE"
const listenAddressEnvKey = "BEN_LISTEN_ADDRESS"
const calleeEnvKey = "BEN_CALLS"
const calleeSeparator = " "

func prepareSystemDefinition(def *SystemDefinition) {
	// Prepare a map from names to services for faster searching
	def.serviceMap = make(map[string]*Service, len(def.Services))

	for i, svc := range def.Services {
		svc.id = i
		def.serviceMap[svc.Name] = &svc
	}
}

func prepareDeployment(def SystemDefinition) *appsv1.Deployment {
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: def.Name,
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
					ContainerPort: int32(baseListenPort + i),
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
					Name:  serviceTypeEnvKey,
					Value: svc.Type,
				},
				{
					Name:  calleeEnvKey,
					Value: assembleCalls(svc.Calls, def.Name, def.Namespace, def.serviceMap),
				},
				{
					Name:  listenAddressEnvKey,
					Value: ":" + strconv.Itoa(baseListenPort + i),
				},
			},
		}

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
				"name": def.Name,
			},
			Type: "ClusterIP",
		},
	}

	for i, svc := range def.Services {
		port := def.serviceMap[svc.Name].Port

		service.Spec.Ports[i] = apiv1.ServicePort{
			// Currently there's no need to set a port name
			//Name:       "",
			Protocol:   "TCP",
			Port:       int32(port),
			TargetPort: intstr.FromInt(port),
		}
	}

	return service
}

func commitService(serviceClient clientcorev1.ServiceInterface, service *apiv1.Service) (*apiv1.Service, error) {
	return serviceClient.Create(context.TODO(), service, metav1.CreateOptions{})
}

func createService(clientset *kubernetes.Clientset, def SystemDefinition) {
	service := prepareService(def)

	fmt.Printf("%#v\n", service)
	serviceClient := clientset.CoreV1().Services(def.Namespace)
	result, err := commitService(serviceClient, service)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Created service %q.\n", result.GetObjectMeta().GetName())
}

func assembleCalls(calls []string, systemName string, namespace string, serviceMap map[string]*Service) string {
	if len(calls) == 0 {
		return ""
	}

	urls := make([]string, len(calls))
	for i, callee := range calls {
		//"http://info-service.app.svc.cluster.local/info"
		//"http://service-name.namespace.svc.cluster.local:port"
		urls[i] = fmt.Sprintf("http://%s.%s.svc.cluster.local:%s", callee, systemName, namespace, serviceMap[callee].Port)
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
