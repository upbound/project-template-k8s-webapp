# project-template-k8s-webapp

This template can be used to initialize a new project using crossplane v2. By
default it comes with an `WebApp` XRD and a matching composition
function which creates an Deployment, optional Service and optional Ingress.
It also creates the corresponding unit and e2e tests.

## Usage

To use this template, run the following command:

```shell
up project init -t upbound/project-template-k8s-webapp --language=kcl <project-name>
```

This template supports the following languages:

- `kcl`
- `go`
- `python`
- `go-templating`
