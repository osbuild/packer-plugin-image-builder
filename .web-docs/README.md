# Image Builder Packer plugin

HashiCorp [Packer](https://www.packer.io/) plugin for [image-builder-cli](https://github.com/osbuild/image-builder-cli) and [bootc-image-builder](https://github.com/osbuild/bootc-image-builder).

## Installation

To install this plugin, copy and paste this code into your Packer configuration, then run [`packer init`](https://www.packer.io/docs/commands/init).

```hcl
packer {
  required_plugins {
    image-builder = {
      source = "github.com/lzap/image-builder"
      version = ">= 0.0.1"
    }
  }
}
```

Alternatively, you can use `packer plugins install` to manage installation of this plugin.

```sh
$ packer plugins install github.com/lzap/image-builder
```

Visit the project page for more information and examples: github.com/lzap/packer-plugin-image-builder
