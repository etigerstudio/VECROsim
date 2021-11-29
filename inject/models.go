package main

import (
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type FaultDefinition struct {
	Name string `json:"name"`
	Namespace string `json:"namespace"`
	Faults []Fault `json:"faults"`
}

type Fault struct {
	Name string `json:"name"`
	Target string `json:"target"`
	Start v1.Duration `json:"start"`
	Duration v1.Duration `json:"duration"`
	Behaviors Behaviors `json:"behaviors"`
}

type Behaviors struct {
	NetDelay `json:"net-delay"`
	NetLoss `json:"net-loss"`
	NetRate `json:"net-rate"`
	IOStress `json:"io-stress"`
	CPUStress `json:"cpu-stress"`
}

type NetDelay struct {
	Time v1.Duration `json:"time"`
	Jitter v1.Duration `json:"jitter"`
}

type NetLoss struct {
	Percent int `json:"percent"`
}

type NetRate struct {
	Rate string `json:"rate"`
}

type IOStress struct {
	Method string `json:"method"`
}

type CPUStress struct {
	Load int `json:"load"`
	Method string `json:"method"`
}