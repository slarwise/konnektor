package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes/scheme"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

type Result struct {
	Service             string               `json:"service"`
	Deployment          string               `json:"deployment"`
	SelectorIsMatching  bool                 `json:"selector_is_matching"`
	MatchingTargetPorts []intstr.IntOrString `json:"matching_target_ports"`
}

func main() {
	flag.Usage = func() {
		fmt.Println("Verify if a Service connects to a Deployment")
		fmt.Println("Provide yaml by files or stdin")
		fmt.Println("EXAMPLES")
		fmt.Printf("    %s service.yaml deployment.yaml\n", os.Args[0])
		fmt.Printf("    kustomize build ./yaml | %s\n", os.Args[0])
	}
	flag.Parse()
	args := flag.Args()
	var yamlData []byte
	if len(args) == 0 {
		bytes, err := io.ReadAll(os.Stdin)
		if err != nil {
			log.Fatalf("Could not read from stdin: %s\n", err)
		}
		yamlData = bytes
	} else {
		for _, f := range args {
			file, err := os.ReadFile(f)
			if err != nil {
				log.Fatalf("Could not read file: %s\n", err)
			}
			yamlData = append(yamlData, file...)
			newDocumentMarker := []byte("---\n")
			yamlData = append(yamlData, newDocumentMarker...)
		}
	}

	documents := bytes.Split(yamlData, []byte("---"))
	deserializer := scheme.Codecs.UniversalDeserializer()
	var objects []runtime.Object
	for _, d := range documents {
		if bytes.TrimSpace(d) == nil {
			continue
		}
		obj, _, err := deserializer.Decode(d, nil, nil)
		if err != nil {
			// Unknown object, skip it
			continue
		}
		objects = append(objects, obj)
	}
	var service *corev1.Service
	var deployment *appsv1.Deployment
	for _, o := range objects {
		switch o.GetObjectKind().GroupVersionKind().Kind {
		case "Service":
			if service != nil {
				log.Fatalln("Found more than one Service")
			}
			service = o.(*corev1.Service)
		case "Deployment":
			if deployment != nil {
				log.Fatalln("Found more than one Deployment")
			}
			deployment = o.(*appsv1.Deployment)
		default:
			// Do nothing for other kubernetes objects
		}
	}
	if service == nil {
		fmt.Fprintf(os.Stderr, "No Service found")
		os.Exit(1)
	}
	if deployment == nil {
		fmt.Fprintf(os.Stderr, "No Deployment found")
		os.Exit(1)
	}

	var result Result
	result.Service = service.Name
	result.Deployment = deployment.Name
	result.SelectorIsMatching = labelsMatchSelector(service.Spec.Selector, deployment.Spec.Template.Labels)
	for _, servicePort := range service.Spec.Ports {
		for _, container := range deployment.Spec.Template.Spec.Containers {
			_ = container.Name
			for _, containerPort := range container.Ports {
				if portsMatch(servicePort, containerPort) {
					result.MatchingTargetPorts = append(result.MatchingTargetPorts, servicePort.TargetPort)
				}
			}
		}
	}
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	err := encoder.Encode(result)
	if err != nil {
		log.Fatalf("Failed to encode json result %s\n", err)
	}
}

func labelsMatchSelector(selector map[string]string, podLabels map[string]string) bool {
	matches := 0
	for sl, sv := range selector {
		pv, ok := podLabels[sl]
		if ok && pv == sv {
			matches += 1
		}
	}
	return matches == len(selector)
}

func portsMatch(servicePort corev1.ServicePort, containerPort corev1.ContainerPort) bool {
	// Check if the name OR containerPort matches targetPort
	switch servicePort.TargetPort.Type {
	case intstr.Int:
		if servicePort.TargetPort.IntVal == containerPort.ContainerPort {
			return true
		}
	case intstr.String:
		if servicePort.TargetPort.StrVal == containerPort.Name {
			return true
		}
	}
	return false
}
