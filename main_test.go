package main

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

type Selector map[string]string
type Labels map[string]string

func TestLabelsMatchSelector(t *testing.T) {
	if !labelsMatchSelector(Selector{"app": "myapp"}, Labels{"app": "myapp"}) {
		t.Fail()
	}

	if !labelsMatchSelector(Selector{"app": "myapp"}, Labels{"app": "myapp", "color": "red"}) {
		t.Fail()
	}

	if labelsMatchSelector(Selector{"app": "myapp"}, Labels{}) {
		t.Fail()
	}

	if labelsMatchSelector(Selector{"app": "myapp"}, Labels{"app": "asdf"}) {
		t.Fail()
	}

	if labelsMatchSelector(Selector{"app": "myapp", "color": "red"}, Labels{"app": "myapp"}) {
		t.Fail()
	}

	if labelsMatchSelector(Selector{"app": "myapp", "color": "red"}, Labels{"app": "myapp", "color": "pink"}) {
		t.Fail()
	}
}

func TestGetMatchingPorts(t *testing.T) {
	servicePort := corev1.ServicePort{
		TargetPort: intstr.IntOrString{
			Type:   intstr.String,
			StrVal: "http",
		},
	}
	containerPort := corev1.ContainerPort{
		Name: "http",
	}
	if !portsMatch(servicePort, containerPort) {
		t.Fail()
	}

	containerPort.Name = "asdf"
	if portsMatch(servicePort, containerPort) {
		t.Fail()
	}

	servicePort = corev1.ServicePort{
		TargetPort: intstr.IntOrString{
			Type:   intstr.Int,
			IntVal: 8080,
		},
	}
	if portsMatch(servicePort, containerPort) {
		t.Fail()
	}

	containerPort.ContainerPort = 8080
	if !portsMatch(servicePort, containerPort) {
		t.Fail()
	}
}
