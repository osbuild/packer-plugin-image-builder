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

Finally, install podman or docker and make sure scp is present as well:

    dnf -y install podman openssh-clients

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

## Usage

See [the builder documentation](.web-docs/) for more information on how to define and build an image. The same documentation is also available at the [Packer Integrations](https://developer.hashicorp.com/packer/integrations) page.

## Building without Packer

To test this library directly without packer, do:

    go run github.com/osbuild/packer-plugin-image-builder/cmd/ibpacker/ -help

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
git clone https://github.com/osbuild/packer-plugin-image-builder
go run ./cmd/ibpacker/ \
    -hostname example.com \
    -username builder \
    cli
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
