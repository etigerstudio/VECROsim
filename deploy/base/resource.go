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
	"strconv"
	"strings"
)

const labelManagedBy = "ben-sim"
const imageName = "ben-base:v1"
const appListenPort = 8080

const nameEnvKey = "BEN_NAME"
const subsystemEnvKey = "BEN_SUBSYSTEM"
const serviceTypeEnvKey = "BEN_SERVICE_TYPE"
const listenAddressEnvKey = "BEN_LISTEN_ADDRESS"
const calleeEnvKey = "BEN_CALLS"
const calleeSeparator = " "

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
					Name:          "http",
					ContainerPort: int32(appListenPort + i),
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
					Value: assembleCalls(svc.Calls, def.Name, def.Namespace),
				},
				{
					Name:  listenAddressEnvKey,
					Value: ":" + strconv.Itoa(appListenPort + i),
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

func CreateDeployment(clientset *kubernetes.Clientset, def SystemDefinition) {
	deployment := prepareDeployment(def)

	//fmt.Printf("%#v\n", deployment)
	deploymentsClient := clientset.AppsV1().Deployments(def.Namespace)
	result, err := commitDeployment(deploymentsClient, deployment)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Created deployment %q.\n", result.GetObjectMeta().GetName())
}

func assembleCalls(calls []string, systemName string, namespace string) string {
	if len(calls) == 0 {
		return ""
	}

	urls := make([]string, len(calls))
	for i, callee := range calls {
		//"http://info-service.app.svc.cluster.local/info"
		urls[i] = fmt.Sprintf("http://%s.%s.%s.svc.cluster.local", callee, systemName, namespace)
	}

	return strings.Join(urls, calleeSeparator)
}

func int32Ptr(i int32) *int32 { return &i }
