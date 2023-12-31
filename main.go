package main

import (
	"flag"
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

type Service struct {
	Spec struct {
		Selector map[string]string `yaml:"selector"`
		Ports    []struct {
			TargetPort interface{} `yaml:"targetPort"`
		} `yaml:"ports"`
	} `yaml:"spec"`
}

type Deployment struct {
	Spec struct {
		Template PodTemplate `yaml:"template"`
	} `yaml:"spec"`
}

type PodTemplate struct {
	Metadata struct {
		Labels map[string]string `yaml:"labels"`
	} `yaml:"metadata"`
	Spec struct {
		Containers []struct {
			Ports []ContainerPort `yaml:"ports"`
		} `yamls:"containers"`
	} `yaml:"spec"`
}

type ContainerPort struct {
	ContainerPort int    `yaml:"containerPort"`
	Name          string `yaml:"name"`
}

// Create a CLI that
// Takes a Service and a Deployment file and
// checks if they match, i.e.
// all of the Service's spec.selector's labels are set on the Deployment's
// spec.template.spec.metadata.labels
// And
// The Service's spec.ports.targetPort matches one of the Deployment's
// spec.template.spec.containers[*].ports[*].[containerPort,name]
// Would be nice to print out if they do, and how they connect

// Take 2: Return what matches there are between the Service and Deployment.
// There could be more than 1

func main() {
	deploymentFlag := flag.String("deployment", "", "filepath to Deployment, such as deployment.yaml")
	serviceFlag := flag.String("service", "", "filepath to Service, such as service.yaml")
	flag.Parse()
	if *deploymentFlag == "" {
		log.Fatalln("-deployment must be provided")
	}
	if *serviceFlag == "" {
		log.Fatalln("-service must be provided")
	}

	serviceFile, err := os.ReadFile(*serviceFlag)
	if err != nil {
		log.Fatalf("Could not read service file: %s\n", err)
	}
	deploymentFile, err := os.ReadFile(*deploymentFlag)
	if err != nil {
		log.Fatalf("Could not read deployment: file %s\n", err)
	}

	service := Service{}
	err = yaml.Unmarshal(serviceFile, &service)
	if err != nil {
		log.Fatalf("Could not parse Service: %s\n", err)
	}
	log.Println(service.Spec.Selector)
	log.Println(service.Spec.Ports[0])
	deployment := Deployment{}
	err = yaml.Unmarshal(deploymentFile, &deployment)
	if err != nil {
		log.Fatalf("Could not parse Deployment %s\n", err)
	}
	log.Println(deployment.Spec.Template.Metadata.Labels)
}

func getMatches(service Service, pod PodTemplate) bool {
	return labelsMatchSelector(service.Spec.Selector, pod.Metadata.Labels)
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

func getMatchingPorts(targetPort string, containerPorts []ContainerPort) []ContainerPort {
	// Check if the name OR containerPort matches targetPort
	matches := []ContainerPort{}
	for _, cp := range containerPorts {
		if targetPort == cp.Name {
			matches = append(matches, cp)
		}
	}
	return matches
}
