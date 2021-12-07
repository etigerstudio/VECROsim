package main

import (
	"BenSim/inject/faults"
	"context"
	"errors"
	"fmt"
	batchv1 "k8s.io/api/batch/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	clientbatchv1 "k8s.io/client-go/kubernetes/typed/batch/v1"
	"time"
)

const labelManagedBy = "ben-sim"

func (fdef *FaultDefinition) Run(ctx context.Context, clientset *kubernetes.Clientset) {
	jobsClient := clientset.BatchV1().Jobs(fdef.Namespace)

	for _, f := range fdef.Faults {
		go singleFault(jobsClient, f)
	}

	select {
	case <-ctx.Done():
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			logger.Print("Fault injection completed successfully.")
		}
	}
}

func singleFault(jobsClient clientbatchv1.JobInterface, f Fault) {
	t := time.NewTimer(f.Duration.Duration)
	<-t.C

	createJob(jobsClient, f)
}

func createJob(jobsClient clientbatchv1.JobInterface, f Fault) {
	job := prepareJob(f)

	//fmt.Printf("%#v\n", job)
	result, err := jobsClient.Create(context.TODO(), job, metav1.CreateOptions{})
	if err != nil {
		panic(err)
	}

	fmt.Printf("Created job %q.\n", result.GetObjectMeta().GetName())
}

func prepareJob(f Fault) *batchv1.Job {
	pod := faults.NewPumbaPod()

	if f.Behaviors.NetDelay.Time.Milliseconds() > 0 {
		faults.AddNetDelay(pod,
			f.Name,
			f.Target,
			f.Behaviors.NetDelay.Time,
			f.Behaviors.NetDelay.Jitter,
			f.Duration)
	}

	if f.Behaviors.NetLoss.Percent > 0 {
		faults.AddNetLoss(pod,
			f.Name,
			f.Target,
			f.Behaviors.NetLoss.Percent,
			f.Duration)
	}

	if f.Behaviors.NetRate.Rate != "" {
		faults.AddNetRate(pod,
			f.Name,
			f.Target,
			f.Behaviors.NetRate.Rate,
			f.Duration)
	}

	if f.Behaviors.CPUStress.Load > 0 {
		method := f.Behaviors.CPUStress.Method
		if method == "" {
			method = "all" // Defaults to use all cpu stressing methods sequentially
		}
		faults.AddCPUStress(pod,
			f.Name,
			f.Target,
			f.Behaviors.CPUStress.Load,
			f.Behaviors.CPUStress.Method,
			f.Duration)
	}

	if f.Behaviors.IOStress.Method != "" {
		faults.AddIOStress(pod,
			f.Name,
			f.Target,
			f.Behaviors.IOStress.Method,
			f.Duration)
	}

	return &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			// Use GenerateName Field to make name unique for every job.
			GenerateName: fmt.Sprintf("%s-", f.Name),
			//Name:      f.Name,
			Labels: map[string]string{
				"app.kubernetes.io/name":       f.Name,
				"app.kubernetes.io/managed-by": labelManagedBy,
			},
		},
		Spec: batchv1.JobSpec{
			// Selector for a job is not necessary.
			//Selector: &metav1.LabelSelector{
			//	MatchLabels: map[string]string{
			//		"app.kubernetes.io/name": f.Name,
			//	},
			//},
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name:      f.Name,
					Labels: map[string]string{
						"app.kubernetes.io/name":       f.Name,
						"app.kubernetes.io/managed-by": labelManagedBy,
					},
				},
				Spec: *pod,
			},
		},
	}
}
