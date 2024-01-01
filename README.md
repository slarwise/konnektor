# Konnektor

Check whether a Service will connect to a Deployment.

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
  "service": "myapp",
  "deployment": "myapp",
  "selector_is_matching": true,
  "matching_target_ports": ["http"]
}
```
