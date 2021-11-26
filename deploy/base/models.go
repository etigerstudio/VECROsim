package base

type Service struct {
	id int
	Name string `json:"name"`
	Subsystem string `json:"subsystem"`
	Workload `json:"workload"`
	Calls []string `json:"calls"`
	Port int
}

type SystemDefinition struct {
	Name string `json:"name"`
	Replicas int32 `json:"replicas"`
	Services []Service `json:"services"`
	serviceMap map[string]*Service
	Namespace string `json:"namespace"`
}

type Workload struct {
	CPU int `json:"cpu"`
	IO int `json:"io"`
	Delay `json:"delay"`
}

type Delay struct {
	Duration int `json:"duration"`
	Jitter int `json:"jitter"`
}