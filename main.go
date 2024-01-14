package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"slices"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"gopkg.in/yaml.v3"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes/scheme"
)

type Result struct {
	ServiceMonitors []ServiceMonitor `json:"service_monitors"`
	Services        []Service        `json:"service"`
}

type ServiceMonitor struct {
	Name     string   `json:"name"`
	Services []string `json:"services"`
}

type Service struct {
	Name        string   `json:"name"`
	Deployments []string `json:"deployments"`
}

func main() {
	asdf
	ignoreNamespace := flag.Bool("i", false, "Do not check whether namespaces match")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, `See how service monitors, services and deployments connect
Usage of %s:
	%s [flags] [FILE]

Provide yaml files by files or stdin
Flags:
`, os.Args[0], os.Args[0])
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, `EXAMPLES
    %s service-monitor.yaml service.yaml deployment.yaml
    kustomize build ./yaml | %s
`, os.Args[0], os.Args[0])
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

	monitoringv1.AddToScheme(scheme.Scheme)
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
	var services []*corev1.Service
	var deployments []*appsv1.Deployment
	var serviceMonitors []*monitoringv1.ServiceMonitor
	for _, o := range objects {
		switch o.GetObjectKind().GroupVersionKind().Kind {
		case "Service":
			services = append(services, o.(*corev1.Service))
		case "Deployment":
			deployments = append(deployments, o.(*appsv1.Deployment))
		case "ServiceMonitor":
			serviceMonitors = append(serviceMonitors, o.(*monitoringv1.ServiceMonitor))
		default:
			// Do nothing for other kubernetes objects
		}
	}

	result := Result{}
	for _, sm := range serviceMonitors {
		matchingServices := []string{}
		for _, s := range services {
			selectorMatch := labelsMatchSelector(sm.Spec.Selector.MatchLabels, s.Labels)
			nsMatch := true
			if !*ignoreNamespace {
				nsMatch = slices.Contains(sm.Spec.NamespaceSelector.MatchNames, s.Namespace)
			}
			if nsMatch && selectorMatch {
				matchingServices = append(matchingServices, s.Name)
			}
		}
		result.ServiceMonitors = append(result.ServiceMonitors, ServiceMonitor{
			Name:     sm.Name,
			Services: matchingServices,
		})
	}

	for _, s := range services {
		matchingDeployments := []string{}
		for _, servicePort := range s.Spec.Ports {
			for _, d := range deployments {
				for _, container := range d.Spec.Template.Spec.Containers {
					for _, containerPort := range container.Ports {
						if portsMatch(servicePort, containerPort) {
							matchingDeployments = append(matchingDeployments, d.Name)
						}
					}
				}
			}
		}
		result.Services = append(result.Services, Service{
			Name:        s.Name,
			Deployments: matchingDeployments,
		})
	}

	encoder := yaml.NewEncoder(os.Stdout)
	encoder.SetIndent(2)
	err := encoder.Encode(result)
	if err != nil {
		log.Fatalf("Failed to encode yaml result %s\n", err)
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
