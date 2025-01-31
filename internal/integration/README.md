## Integration tests

This directory contains a test that starts a mock SSH "honeypot" and loads fixtures and Packer template from YAML files. Then it executes `packer` command as a subprocess for each test, passes a template as a temporary file and check fixtures in the SSH mock.

Before integration tests can be executed, a plugin binary must be compiled in `./build` directory. When tests are invoked via Make, it will make sure the binary is up to date.

    make test

## YAML test cases

Example fixture YAML:

```yaml
---
fixtures:
  - request: which (podman|docker)
    reply: /usr/bin/podman
    status: 0

environment:
  - ENV_VAR=abc

template: |+
  source "image-builder" "example" {
      build_host {
          hostname = "{{ .Hostname }}"
      }
  }
  build {
      sources = [ "source.image-builder.example" ]
  }

result:
  grep: "Builds finished"
  status: 0
```

### Fixture

A list of fixtures for the SSH mock:

* `request` (required): Go regular expression capturing the SSH input
* `response`: optional output of a command
* `status`: optional exit status (defaults to 0)

### Environment

Optional environment variables to pass to the packer process.

### Template

A Go template which renders to valid Packer HCL template. Available variables:

* `Hostname`: `hostname:port` of running SSH mock (localhost with a random port)

### Result

Check of the packer process output and status:

* `grep` (required): string that must be found in the packer output
* `status`: optional exit status (defaults to 0)
