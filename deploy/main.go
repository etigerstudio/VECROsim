package main

import (
	"BenSim/deploy/base"
	"flag"
	"io/ioutil"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"os"
	"path/filepath"
)

func main() {
	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}

	defFilePath := flag.String("deffile", "", "path to system definition file")

	flag.Parse()

	// Open & parse system definition file in YAML
	defFile, err := os.Open(*defFilePath)
	if err != nil {
		panic(err)
	}
	defer defFile.Close()

	sysdefStr, _ := ioutil.ReadAll(defFile)
	var sysdef base.SystemDefinition
	err = yaml.Unmarshal(sysdefStr, &sysdef)
	if err != nil {
		panic(err)
	}

	// Connect to Kubernetes & deploy services
	clientset := getClientset(*kubeconfig)
	base.CreateResources(clientset, sysdef)
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
