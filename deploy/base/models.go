package base

type Service struct {
	id int
	Name string `json:"name"`
	Subsystem string `json:"subsystem"`
	Type string `json:"type"`
	Calls []string `json:"calls"`
	Port int
	//Cluster string `json:"cluster"`
	//Namespace string `json:"namespace"`
}

type SystemDefinition struct {
	Name string `json:"name"`
	Replicas int32 `json:"replicas"`
	Services []Service `json:"services"`
	serviceMap map[string]*Service
	Namespace string `json:"namespace"`
}
