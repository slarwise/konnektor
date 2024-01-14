# Konnektor

Check connections between service monitors, services and deployments, by parsing
the yaml files.

## Installation

```sh
go install github.com/slarwise/konnektor
```

## Usage

```sh
konnektor ./examples/*
kustomize build ./examples | konnektor
```

Example output:

```yaml
servicemonitors:
  - name: myapp
    services:
      - myapp
services:
  - name: myapp
    deployments:
      - myapp
```
