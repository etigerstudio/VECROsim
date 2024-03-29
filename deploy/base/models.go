package base

type Service struct {
	id int
	Name string `json:"name"`
	Workload `json:"workload"`
	Type string `json:"type"`
	Node string `json:"node"` // TODO: support multi-node
	Calls []string `json:"calls"`
}

type SystemDefinition struct {
	Name string `json:"name"`
	Replicas int32 `json:"replicas"`
	Services []Service `json:"services"`
	Namespace string `json:"namespace"`
}

// TODO: support memory or cpu cache workload
type Workload struct {
	CPU int `json:"cpu"` // CPU relative bogus operation number
	IO int `json:"io"` // IO relative bogus operation number
	Delay `json:"delay"` // Delay achieved by sleeping
	Net int `json:"net"` // Network egress data payload size
	Memory int `json:"memory"` // Memory object allocation size
	Read int `json:"read"` // Database read operation number
	Write int `json:"write"` // Database write operation number
}

type Delay struct {
	Duration int `json:"duration"`
	Jitter int `json:"jitter"`
}