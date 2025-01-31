# Image Builder Packer plugin

HashiCorp [Packer](https://www.packer.io/) plugin for [image-builder-cli](https://github.com/osbuild/image-builder-cli) and [bootc-image-builder](https://github.com/osbuild/bootc-image-builder). Builds are happening on a remote linux machine over SSH.

## Preparing the environment

All building is done remotely over SSH connection, you need to create an instance or VM with a user dedicated to image building and sudo permission to start podman (or docker) without password or root account.

    adduser -m builder

Either setup a password

    passwd builder

or preferably deploy a public SSH key (execute from machine with packer)

    ssh-copy-id builder@host

Make sure the container runtime can be executed without password.

```
cat <<EOF >/etc/sudoers.d/builder
builder ALL=(ALL) NOPASSWD: /usr/bin/podman, /usr/bin/docker
EOF
```

Finally, install podman or docker:

    dnf -y install podman

Cross-architecture building is currently not supported so make sure the builder host architecture is correct.

## Install packer

Install packer *on your machine* not on the builder instance/VM, for example on Fedora:

```
sudo dnf install -y dnf-plugins-core
sudo dnf config-manager addrepo --from-repofile=https://rpm.releases.hashicorp.com/fedora/hashicorp.repo
sudo dnf -y install packer
```

On MacOS:

```
brew tap hashicorp/tap
brew install hashicorp/tap/packer
```

## Building using image-builder-cli

Create a packer template named `template.pkr.hcl`:

```
packer {
  required_plugins {
    image-builder = {
      source = "github.com/lzap/image-builder"
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
      source = "github.com/lzap/image-builder"
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

## Building without Packer

To test this library directly without packer, do:

    go run github.com/lzap/packer-plugin-image-builder/cmd/ibpacker/ -help

Use options to initiate a build:

```
Usage of ibpacker:
  -arch string
        architecture (default "x86_64")
  -blueprint string
        path to blueprint file
  -distro string
        distribution name (fedora, centos, rhel, ...) (default "fedora")
  -dry-run
        dry run
  -hostname string
        SSH hostname or IP with optional port (e.g. example.com:22)
  -type string
        image type (minimal-raw, qcow2, ...) (default "minimal-raw")
  -username string
        SSH username
```

For example:

```
git clone github.com/lzap/packer-plugin-image-builder
go run ./cmd/ibpacker/ \
    -hostname example.com \
    -username builder \
    -distro centos-9 \
    -type minimal-raw \
    -blueprint ./cmd/ibpacker/blueprint_example.toml
```

## Testing

To run unit and integration test against mock SSH server running on localhost:

    make test

Do not invoke tests directly via `go test` command because some tests require packer plugin binary to be present in the `./build` directory. To prevent these tests to be skipped, do this prior running tests directly:

    make build

Keys in `internal/sshtest/keys.go` are just dummy (test only) keys, you may receive false positives from security scanners about leaked keys when cloning the repo.

## Release process

The release process is fully automated, just push a new tag to the main branch and observe GitHub Actions to make a new release. To update plugin in an existing Terraform/Packer project:

    packer init --upgrade .

## LICENSE

Apache Version 2.0
