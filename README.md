# Konnektor - Verify that a Deployment and a Service will connect

Compare yaml files describing a Deployment and a Service and check if they will
connect or not.

## Installation

```sh
go install github.com/slarwise/konnektor
```

## Usage

```sh
konnektor ./examples/service.yaml ./examples/deployment.yaml
kustomize build ./examples | konnektor
```

Example output:

```json
{
  "service": "kubeconform",
  "deployment": "kubeconform",
  "selector_is_matching": true,
  "matching_target_ports": ["http"]
}
```
