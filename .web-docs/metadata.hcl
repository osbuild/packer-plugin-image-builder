# For full specification on the configuration of this file visit:
# https://github.com/hashicorp/integration-template#metadata-configuration
#
integration {
  name = "Image Builder"
  description = "Integration for building OS images via osbuild aka Red Hat Image Builder remotely via SSH."
  identifier = "packer/osbuild/image-builder-packer-integration"
  component {
    type = "builder"
    name = "Image Builder"
    slug = "image-builder-builder"
  }
}
