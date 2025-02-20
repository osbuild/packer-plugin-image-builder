# Image Builder Packer plugin

HashiCorp [Packer](https://www.packer.io/) plugin for [image-builder-cli](https://github.com/osbuild/image-builder-cli) and [bootc-image-builder](https://github.com/osbuild/bootc-image-builder).

## Installation

To install this plugin, copy and paste this code into your Packer configuration, then run [`packer init`](https://www.packer.io/docs/commands/init).

```hcl
packer {
  required_plugins {
    image-builder = {
      source = "github.com/osbuild/image-builder"
      version = ">= 0.0.1"
    }
  }
}
```

Alternatively, you can use `packer plugins install` to manage installation of this plugin.

```sh
$ packer plugins install github.com/osbuild/image-builder
```

Visit the project page for more information and examples: github.com/osbuild/packer-plugin-image-builder

## Building using image-builder-cli

Create a packer template named `template.pkr.hcl`:

```
packer {
  required_plugins {
    image-builder = {
      source = "github.com/osbuild/image-builder"
      version = ">= 0.0.2"
    }
  }
}

source "image-builder" "example" {
    build_host {
        hostname = "buildhost.example.com"
        username = "builder"
    }

    distro = "centos-9"

    blueprint = <<BLUEPRINT
[[customizations.user]]
name = "user"
password = "changeme"
groups = ["wheel"]
BLUEPRINT

    image_type = "minimal-raw"
}

build {
    sources = [ "source.image-builder.example" ]
}
```

See [osbuild blueprint reference](https://osbuild.org/docs/user-guide/blueprint-reference/) for more info about blueprint format.

Perform the build via:

      packer init template.pkr.hcl
      packer build template.pkr.hcl

The image builder plugin will print last several lines from the image builder output as an artifact. To see more detailed output:

      PACKER_LOG=1 packer build template.pkr.hcl

### Config mapping

For more info: https://github.com/osbuild/image-builder-cli

* **build_host.hostname** - IP or hostname with optional SSH port (required)
* **build_host.username** - either root or username with sudo permissions (required)
* **build_host.password** - SSH password when SSH keys are not available
* **distro** - maps to `--distro`
* **blueprint** - maps to `--blueprint`
* **image_type** - maps to image type argument

If there is an option missing, file an issue for us.

## Building using bootc-image-builder

Create a packer template named `template.pkr.hcl`:

```
packer {
  required_plugins {
    image-builder = {
      source = "github.com/osbuild/image-builder"
      version = ">= 0.0.2"
    }
  }
}

source "image-builder" "example" {
    build_host {
        hostname = "buildhost.example.com"
        username = "builder"
    }

    container_repository = "quay.io/centos-bootc/centos-bootc:stream9"

    blueprint = <<BLUEPRINT
[[customizations.user]]
name = "user"
password = "changeme"
groups = ["wheel"]
BLUEPRINT

    image_type = "raw"
}

build {
    sources = [ "source.image-builder.example" ]
}
```

See [osbuild blueprint reference](https://osbuild.org/docs/user-guide/blueprint-reference/) for more info about blueprint format.

Perform the build via:

      packer init template.pkr.hcl
      packer build template.pkr.hcl

The image builder plugin will print last several lines from the image builder output as an artifact. To see more detailed output:

      PACKER_LOG=1 packer build template.pkr.hcl

### Config mapping

For more info: https://github.com/osbuild/bootc-image-builder

* **build_host.hostname** - IP or hostname with optional SSH port (required)
* **build_host.username** - either root or username with sudo permissions (required)
* **build_host.password** - SSH password when SSH keys are not available
* **container_repository** - maps to container repository argument
* **blueprint** - maps to `--blueprint`
* **image_type** - maps to `--type`
* **rootfs** - maps to `--rootfs`
* **aws_upload.ami_name** - maps to AMI cloud uploader configuration
* **aws_upload.s3_bucket** - maps to AMI cloud uploader configuration
* **aws_upload.region** - maps to AMI cloud uploader configuration
* **aws_upload.access_key_id** - maps to AMI cloud uploader configuration
* **aws_upload.secret_access_key** - maps to AMI cloud uploader configuration

If there is an option missing, file an issue for us.

## Dry run

If you want to perform, for any reason, a dry run where the main build command is `echo`ed to the console rather than executed, just set `IMAGE_BUILDER_DRY_RUN=1` environment variable when executing packer. Good for demos or testing the integration.
