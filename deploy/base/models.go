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

// TODO: support memory or cache workload
type Workload struct {
	CPU int `json:"cpu"` // CPU relative bogus operation number
	IO int `json:"io"` // IO relative bogus operation number
	Delay `json:"delay"` // Delay achieved by sleeping
	Net int `json:"net"` // Network egress data payload size
}

type Delay struct {
	Duration int `json:"duration"`
	Jitter int `json:"jitter"`
}