package main

import (
	"testing"
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
	targetPort := "http"
	containerPorts := []ContainerPort{
		{
			ContainerPort: 8080,
			Name:          "http",
		},
	}
	if len(getMatchingPorts(targetPort, containerPorts)) != 1 {
		t.Fail()
	}
}
