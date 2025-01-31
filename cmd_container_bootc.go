package ibk

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"

	"al.essio.dev/pkg/shellescape"
)

// ContainerBootCommand builds a bootc-image-builder command line via podman or docker
// which builds an image from a blueprint. The blueprint is pushed to the remote
// machine via SSH and the command is executed there. The output image is saved
// to the specified directory which must be created and cleanuped up beforehand.
//
// For more information see https://github.com/osbuild/bootc-image-builder
type ContainerBootCommand struct {
	// Source OCI (podman, docker) image repository URL.
	Repository string

	// Type is the image type: ami, anaconda-iso, gce, iso, qcow2, raw, vhd, vmdk (default [qcow2]
	// Maps to the argument named --type.
	Type string

	// Arch is the architecture, must be set to the architecture of the remote machine
	// since cross-compilation is not supported yet.
	// Maps to the argument named --target-arch.
	Arch string

	// RootFS is the root filesystem to use. Supported values: ext4, xfs, btrfs.
	// Maps to the argument named --rootfs.
	RootFS string

	// Blueprint is the full contents of a blueprint.
	Blueprint string

	// OutputDir is the directory where the output image is saved. When unset, a new directory will
	// be created in the remote machine home directory. The caller must cleanup the directory.
	OutputDir string

	// Common arguments for all container commands.
	Common CommonArgs

	// AWSUploadConfig is the configuration for uploading the image to AWS. Must be set when the
	// Type is set to "ami".
	AWSUploadConfig *AWSUploadConfig

	containerCmd       string
	blueprintTempfile  string
	awsSecretsTempfile string
}

var _ Command = &ContainerBootCommand{}

// AWSUploadCommand uploads the image to an S3 bucket and registers it as an AMI.
type AWSUploadConfig struct {
	// AWSAccessKeyID credential. Maps to the AWS_ACCESS_KEY_ID environment variable.
	AWSAccessKeyID string

	// AWSSecretAccessKey credential. Maps to the AWS_SECRET_ACCESS_KEY environment variable.
	AWSSecretAccessKey string

	// AMIName is the name of the AMI to register.
	AMIName string

	// S3Bucket is the name of a temporary S3 bucket to upload the image to.
	S3Bucket string

	// S3Region is the region of the S3 bucket and the resulting AMI.
	Region string
}

var ErrContainerPull = errors.New("error while pulling container")

func (c *ContainerBootCommand) Configure(ctx context.Context, t Executor) error {
	var err error

	// check configuration
	if c.Repository == "" {
		return fmt.Errorf("%w: repository is required", ErrConfigure)
	}

	if c.Type == "" {
		return fmt.Errorf("%w: type is required", ErrConfigure)
	}

	if c.Type == "ami" && c.AWSUploadConfig == nil {
		return fmt.Errorf("%w: aws upload config is required for type ami", ErrConfigure)
	}

	// detect container runtime
	c.containerCmd, err = which(ctx, t, "podman", "docker")
	if err != nil {
		return fmt.Errorf("%w: which: %w", ErrConfigure, err)
	}

	// detect architecture
	if c.Arch != "" {
		arch, err := tail1(ctx, t, "arch")
		if err != nil {
			return fmt.Errorf("%w: arch: %w", ErrConfigure, err)
		}

		log.Printf("[DEBUG] Found architecture %q", arch)
		if c.Arch != arch {
			return fmt.Errorf("%w architecture mismatch: %s", ErrConfigure, arch)
		}
	}

	// create output dir if not set
	if c.OutputDir == "" {
		c.OutputDir = fmt.Sprintf("./output-%s", RandomString(13))
		co, err := tail1(ctx, t, "mkdir "+c.OutputDir)
		if err != nil {
			return fmt.Errorf("%w mktemp: %w, output: %s", ErrConfigure, err, co)
		}
		log.Printf("[DEBUG] Created output dir %q", c.OutputDir)
	}

	// pull the container
	cmd := "sudo " + c.containerCmd + " pull " + shellescape.Quote(c.Repository)
	if c.Common.DryRun {
		cmd = "echo " + cmd
	}
	err = t.Execute(ctx, StringCommand(cmd))
	if err != nil {
		return fmt.Errorf("%w: %w", ErrContainerPull, err)
	}

	return nil
}

func (c *ContainerBootCommand) Push(ctx context.Context, pusher Pusher) error {
	var err error

	// push blueprint
	c.blueprintTempfile, err = pusher.Push(ctx, c.Blueprint, "toml")
	if err != nil {
		return fmt.Errorf("%w: blueprint: %w", ErrPush, err)
	}
	log.Printf("[DEBUG] Pushed blueprint %q", c.blueprintTempfile)

	// push aws.secrets env file
	if c.AWSUploadConfig != nil {
		awsSecrets := fmt.Sprintf("AWS_ACCESS_KEY_ID=%s\nAWS_SECRET_ACCESS_KEY=%s\n",
			c.AWSUploadConfig.AWSAccessKeyID,
			c.AWSUploadConfig.AWSSecretAccessKey,
		)
		c.awsSecretsTempfile, err = pusher.Push(ctx, awsSecrets, "env")
		if err != nil {
			return fmt.Errorf("%w: aws secrets: %w", ErrPush, err)
		}
	}

	return nil
}

func (c *ContainerBootCommand) Build() string {
	sb := strings.Builder{}

	if c.Common.DryRun {
		sb.WriteString("echo")
		sb.WriteRune(' ')
	}

	sb.WriteString("sudo")
	sb.WriteRune(' ')
	sb.WriteString(c.containerCmd)
	sb.WriteRune(' ')
	sb.WriteString("run --privileged --rm --pull=newer")
	sb.WriteRune(' ')
	if c.Common.Interactive {
		sb.WriteString("-i")
		sb.WriteRune(' ')
	}
	if c.Common.TTY {
		sb.WriteString("-t")
		sb.WriteRune(' ')
	}
	sb.WriteString("--security-opt label=type:unconfined_t")
	sb.WriteRune(' ')
	sb.WriteString("-v /var/lib/containers/storage:/var/lib/containers/storage")
	sb.WriteRune(' ')
	sb.WriteString("-v " + shellescape.Quote(c.OutputDir+":/output"))
	sb.WriteRune(' ')
	sb.WriteString("-v " + shellescape.Quote(c.blueprintTempfile+":/config.toml:ro"))
	sb.WriteRune(' ')

	if c.AWSUploadConfig != nil {
		sb.WriteString("--env-file " + c.awsSecretsTempfile)
		sb.WriteRune(' ')
	}

	sb.WriteString("quay.io/centos-bootc/bootc-image-builder:latest")
	sb.WriteRune(' ')
	sb.WriteString("--type " + shellescape.Quote(c.Type))
	sb.WriteRune(' ')
	sb.WriteString("--local")
	sb.WriteRune(' ')

	if c.RootFS != "" {
		sb.WriteString("--rootfs " + shellescape.Quote(c.RootFS))
		sb.WriteRune(' ')
	}

	if c.AWSUploadConfig != nil {
		sb.WriteString("--aws-ami-name " + shellescape.Quote(c.AWSUploadConfig.AMIName))
		sb.WriteRune(' ')
		sb.WriteString("--aws-s3-bucket " + shellescape.Quote(c.AWSUploadConfig.S3Bucket))
		sb.WriteRune(' ')
		sb.WriteString("--aws-s3-region " + shellescape.Quote(c.AWSUploadConfig.Region))
		sb.WriteRune(' ')
	}

	sb.WriteString(shellescape.Quote(c.Repository))

	if c.Common.TeeLog {
		sb.WriteRune(' ')
		sb.WriteString("2>&1 | tee " + c.OutputDir + "/build.log")
	}

	sb.WriteRune(' ')
	sb.WriteString("&& find " + c.OutputDir + " -type f")

	return sb.String()
}
