// Copyright 2025 Red Hat Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package main

// A utility for testing the code without packer integration

import (
	"context"
	"flag"
	"log"
	"os"
	"time"

	"github.com/hashicorp/logutils"
	ibk "github.com/lzap/packer-plugin-image-builder"
)

func cli(ctx context.Context, args []string) {
	flag := flag.NewFlagSet("ibpacker cli", flag.ExitOnError)
	var (
		distro        = flag.String("distro", "fedora", "distribution name (fedora, centos, rhel, ...)")
		imageType     = flag.String("type", "minimal-raw", "image type (minimal-raw, qcow2, ...)")
		arch          = flag.String("arch", "", "architecture")
		blueprintFile = flag.String("blueprint", "", "path to blueprint file")
	)
	flag.Parse(args)

	// open SSH connection
	cfg := ibk.SSHTransportConfig{
		Host:     *hostname,
		Username: *username,
		Timeout:  *connTimeout,
		Stderr:   os.Stdout,
	}
	c, err := ibk.NewSSHTransport(cfg)
	if err != nil {
		log.Panic(err)
	}
	defer c.Close(ctx)

	// load blueprint into a string
	blueprint, err := os.ReadFile(*blueprintFile)
	if err != nil {
		log.Panic(err)
	}

	// configure the command
	cmd := &ibk.ContainerCliCommand{
		Distro:    *distro,
		Type:      *imageType,
		Arch:      *arch,
		Blueprint: string(blueprint),
		Common: ibk.CommonArgs{
			DryRun:      *dryRun,
			Interactive: *interactive,
			TTY:         *tty,
			TeeLog:      *teeLog,
		},
	}

	// apply the command
	err = ibk.ApplyCommand(ctx, cmd, c)
	if err != nil {
		log.Panic(err)
	}
}

func bootc(ctx context.Context, args []string) {
	flag := flag.NewFlagSet("ibpacker bootc", flag.ExitOnError)
	var (
		repository    = flag.String("repository", "", "bootable container OCI/docker repository URL")
		imageType     = flag.String("type", "raw", "image type (ami, anaconda-iso, gce, iso, qcow2, raw, vhd, vmdk)")
		arch          = flag.String("arch", "", "architecture")
		blueprintFile = flag.String("blueprint", "", "path to blueprint file")
		rootFS        = flag.String("rootfs", "", "root file system (ext4, xfs, btrfs)")

		// ami specific
		awsAccessKeyID     = flag.String("aws-access-key-id", "", "AWS access key ID (required for ami type)")
		awsSecretAccessKey = flag.String("aws-secret-access-key", "", "AWS secret access key (required for ami type)")
		awsAmiName         = flag.String("aws-ami-name", "", "destination AMI name (required for ami type)")
		awsS3Bucket        = flag.String("aws-s3-bucket", "", "S3 bucket (required for ami type)")
		awsRegion          = flag.String("aws-region", "", "AWS region (required for ami type)")
	)
	flag.Parse(args)

	// open SSH connection
	cfg := ibk.SSHTransportConfig{
		Host:     *hostname,
		Username: *username,
		Timeout:  *connTimeout,
		Stderr:   os.Stdout,
	}
	c, err := ibk.NewSSHTransport(cfg)
	if err != nil {
		log.Panic(err)
	}
	defer c.Close(ctx)

	// load blueprint into a string
	blueprint, err := os.ReadFile(*blueprintFile)
	if err != nil {
		log.Panic(err)
	}

	// configure the command
	cmd := &ibk.ContainerBootCommand{
		Repository: *repository,
		Type:       *imageType,
		Arch:       *arch,
		Blueprint:  string(blueprint),
		RootFS:     *rootFS,
		Common: ibk.CommonArgs{
			DryRun:      *dryRun,
			Interactive: *interactive,
			TTY:         *tty,
			TeeLog:      *teeLog,
		},
	}
	if *imageType == "ami" {
		cmd.AWSUploadConfig = &ibk.AWSUploadConfig{
			AMIName:            *awsAmiName,
			S3Bucket:           *awsS3Bucket,
			Region:             *awsRegion,
			AWSAccessKeyID:     *awsAccessKeyID,
			AWSSecretAccessKey: *awsSecretAccessKey,
		}
	}

	// apply the command
	err = ibk.ApplyCommand(ctx, cmd, c)
	if err != nil {
		log.Panic(err)
	}
}

var (
	hostname    = flag.String("hostname", "", "SSH hostname or IP with optional port (e.g. example.com:22)")
	username    = flag.String("username", "", "SSH username")
	dryRun      = flag.Bool("dry-run", false, "dry run")
	debug       = flag.Bool("debug", false, "debug logging")
	interactive = flag.Bool("interactive", false, "pass --interactive mode to the container tool")
	tty         = flag.Bool("tty", false, "pass --tty mode to the container tool")
	connTimeout = flag.Duration("conn-timeout", 10*time.Second, "SSH connection timeout")
	timeout     = flag.Duration("timeout", 9999*time.Hour, "transaction timeout (overall build timeout)")
	teeLog      = flag.Bool("tee-log", true, "tee the output log to a file named build.log")
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()

	flag.Parse()
	if *interactive || *tty {
		*teeLog = false
	}

	level := logutils.LogLevel("WARN")
	if *debug {
		level = logutils.LogLevel("DEBUG")
	}
	filter := &logutils.LevelFilter{
		Levels:   []logutils.LogLevel{"DEBUG", "WARN", "ERROR"},
		MinLevel: level,
		Writer:   os.Stderr,
	}
	log.SetOutput(filter)
	log.SetFlags(0)

	args := flag.Args()
	if len(args) == 0 {
		log.Fatal("Please specify a subcommand.")
	}
	cmd, args := args[0], args[1:]

	switch cmd {
	case "cli":
		cli(ctx, args)
	case "bootc":
		bootc(ctx, args)
	default:
		log.Fatalf("Unrecognized command %q. Command must be one of: cli, bootc", cmd)
	}
}
