# {{.ProjectName}}

An example Upbound control plane project for Crossplane V2 with Core Kubernetes Resources.

A control plane project is a source-level representation of a Crossplane control
plane. It lets you treat your control plane configuration as a software project.
With a control plane project you can build your compositions using a language
like KCL or Python. This enables Crossplane schema-aware syntax highlighting,
autocompletion, and linting.

Read the [control plane project documentation][proj-docs] to learn more about
control plane projects.

This project defines a new `WebApp` API.

## Python editor support

If this project uses Python functions or tests, `up` builds them in a
container, so a local Python install (3.11-3.13) is only needed for editor
features. To resolve imports from the generated `models` package, run
`up project build`, then create a virtual environment in each function or test
directory:

```shell
cd functions/<function-name>
python3 -m venv .venv
source .venv/bin/activate
pip install -e .
pip install -e ../../.up/python
```

`pip install -e .` installs a non-editable copy of the models package, so
re-run `pip install -e ../../.up/python` afterwards to point your editor back
at the live models that `up` regenerates.

[proj-docs]: https://docs.upbound.io/core-concepts/projects/
