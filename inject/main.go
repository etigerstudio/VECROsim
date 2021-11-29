package main

import (
	"context"
	"flag"
	"io/ioutil"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"log"
	"os"
	"os/signal"
	"path/filepath"
)

var logger = log.New(os.Stderr, "", 0)

func main() {
	// kubectl config parsing
	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}

	defFilePath := flag.String("deffile", "", "path to fault definition file")
	durationPtr := flag.Duration("duration", 0, "Duration of this round of fault simulation")

	flag.Parse()

	// Open & parse fault definition file in YAML
	defFile, err := os.Open(*defFilePath)
	if err != nil {
		panic(err)
	}
	defer defFile.Close()

	fdefStr, _ := ioutil.ReadAll(defFile)
	var fdef FaultDefinition
	err = yaml.Unmarshal(fdefStr, &fdef)
	if err != nil {
		panic(err)
	}

	// Connect to Kubernetes & deploy services
	clientset := getClientset(*kubeconfig)

	// Make Ctrl-C interruptible
	ctx := interruptibleCxt()
	// Cancel requests when duration expired
	if *durationPtr != 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, *durationPtr)
		defer cancel()
	}

	fdef.Run(ctx, clientset)
}

func getClientset(kubeconfig string) *kubernetes.Clientset {
	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		panic(err.Error())
	}

	// create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	return clientset
}

// TODO: Migrate to main
func interruptibleCxt() context.Context {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, os.Kill)

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		defer cancel()
		_ = <-sig
		logger.Println("Load simulation cancelled.")
	}()

	return ctx
}