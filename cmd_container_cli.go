package ibk

import (
	"context"
	"fmt"
	"log"
	"strings"

	"al.essio.dev/pkg/shellescape"
)

// ContainerCliCommand builds an image-builder-cli command line via podman or docker
// which builds an image from a blueprint. The blueprint is pushed to the remote
// machine via SSH and the command is executed there. The output image is saved
// to the specified directory which must be created and cleanuped up beforehand.
//
// For more information see https://github.com/osbuild/image-builder-cli
type ContainerCliCommand struct {
	// Distro is the distribution name
	Distro string

	// Type is the image type
	Type string

	// Arch is the architecture, must be set to the architecture of the remote machine
	// since cross-compilation is not supported yet.
	Arch string

	// Blueprint is the full contents of a blueprint.
	Blueprint string

	// OutputDir is the directory where the output image is saved. When unset, a new directory will
	// be created in the remote machine home directory. The caller must cleanup the directory.
	OutputDir string

	// Common arguments for all container commands.
	Common CommonArgs

	containerCmd      string
	blueprintTempfile string
}

var _ Command = &ContainerCliCommand{}

func (c *ContainerCliCommand) Configure(ctx context.Context, t Executor) error {
	var err error

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
		
		log.Printf("Detected architecture %s", arch)
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
		log.Printf("[DEBUG] Created output directory %s", c.OutputDir)
	}

	return nil
}

func (c *ContainerCliCommand) Push(ctx context.Context, pusher Pusher) error {
	var err error

	// push blueprint
	c.blueprintTempfile, err = pusher.Push(ctx, c.Blueprint, "toml")
	log.Printf("[DEBUG] Pushed blueprint %s", c.blueprintTempfile)

	return err
}

func (c *ContainerCliCommand) Build() string {
	sb := strings.Builder{}

	if c.Common.DryRun {
		sb.WriteString("echo")
		sb.WriteRune(' ')
	}

	sb.WriteString("sudo")
	sb.WriteRune(' ')
	sb.WriteString(c.containerCmd)
	sb.WriteRune(' ')
	sb.WriteString("run --privileged --rm")
	sb.WriteRune(' ')
	if c.Common.Interactive {
		sb.WriteString("-i")
		sb.WriteRune(' ')
	}
	if c.Common.TTY {
		sb.WriteString("-t")
		sb.WriteRune(' ')
	}
	sb.WriteString("-v " + shellescape.Quote(c.OutputDir+":/output"))
	sb.WriteRune(' ')
	sb.WriteString("-v " + shellescape.Quote(c.blueprintTempfile+":"+c.blueprintTempfile))
	sb.WriteRune(' ')
	sb.WriteString("ghcr.io/osbuild/image-builder-cli:latest")
	sb.WriteRune(' ')
	sb.WriteString("build")
	sb.WriteRune(' ')
	sb.WriteString("--blueprint " + c.blueprintTempfile)
	sb.WriteRune(' ')
	sb.WriteString("--distro " + shellescape.Quote(c.Distro))
	sb.WriteRune(' ')
	sb.WriteString(shellescape.Quote(c.Type))

	if c.Common.TeeLog {
		sb.WriteRune(' ')
		sb.WriteString("2>&1 | tee " + c.OutputDir + "/build.log")
	}

	sb.WriteRune(' ')
	sb.WriteString("&& find " + c.OutputDir + " -type f")

	return sb.String()
}
