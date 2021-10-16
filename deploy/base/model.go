package base

type Service struct {
	//ID int `json:"id"`
	Name string `json:"name"`
	Subsystem string `json:"subsystem"`
	Type string `json:"type"`
	Calls []string `json:"calls"`
	//Cluster string `json:"cluster"`
	//Namespace string `json:"namespace"`
}

type SystemDefinition struct {
	Name string `json:"name"`
	Replicas int32 `json:"replicas"`
	Services []Service `json:"services"`
	Namespace string `json:"namespace"`
}
