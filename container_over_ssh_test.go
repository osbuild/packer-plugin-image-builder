package ibk_test

import (
	"bytes"
	"context"
	"strings"
	"testing"

	ibk "github.com/lzap/packer-plugin-image-builder"
	"github.com/lzap/packer-plugin-image-builder/internal/sshtest"
)

func TestContainerOverSSH(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name    string
		cmd     ibk.Command
		session []sshtest.RequestReply
	}{
		{
			name: "fedora-minimal-raw-all-args",
			cmd: &ibk.ContainerCliCommand{
				Distro:    "fedora",
				Type:      "minimal-raw",
				Arch:      "x86_64",
				Blueprint: "blueprint",
				Common: ibk.CommonArgs{
					DryRun:      true,
					Interactive: true,
					TTY:         true,
					TeeLog:      true,
				},
			},
			session: []sshtest.RequestReply{
				{
					Request: "which podman",
					Reply:   "/usr/bin/podman\n",
					Status:  0,
				},
				{
					Request: "arch",
					Reply:   "x86_64\n",
					Status:  0,
				},
				{
					Request: "mkdir ./output-hehwuXP6NyGIr",
					Reply:   "",
					Status:  0,
				},
				{
					Request: "scp -t /tmp",
					Reply:   "",
					Status:  0,
				},
				{
					Request: "echo sudo /usr/bin/podman run --privileged --rm -i -t " +
						"-v ./output-hehwuXP6NyGIr:/output -v /tmp/ibpacker-o2rHJLEEkT68y.toml:/tmp/ibpacker-o2rHJLEEkT68y.toml " +
						"ghcr.io/osbuild/image-builder-cli:latest build " +
						"--blueprint /tmp/ibpacker-o2rHJLEEkT68y.toml " +
						"--distro fedora minimal-raw " +
						"2>&1 | tee ./output-hehwuXP6NyGIr/build.log && find ./output-hehwuXP6NyGIr -type f",
					Reply:  "Building...\nDone.\n",
					Status: 0,
				},
				{
					Request: "rm -f /tmp/ibpacker-o2rHJLEEkT68y.toml",
					Reply:   "",
					Status:  0,
				},
			},
		},
		{
			name: "fedora-minimal-raw",
			cmd: &ibk.ContainerCliCommand{
				Distro:    "fedora",
				Type:      "minimal-raw",
				Arch:      "x86_64",
				Blueprint: "blueprint",
			},
			session: []sshtest.RequestReply{
				{
					Request: "which podman",
					Reply:   "/usr/bin/podman\n",
					Status:  0,
				},
				{
					Request: "arch",
					Reply:   "x86_64\n",
					Status:  0,
				},
				{
					Request: "mkdir ./output-hehwuXP6NyGIr",
					Reply:   "",
					Status:  0,
				},
				{
					Request: "scp -t /tmp",
					Reply:   "",
					Status:  0,
				},
				{
					Request: "sudo /usr/bin/podman run --privileged --rm " +
						"-v ./output-hehwuXP6NyGIr:/output -v /tmp/ibpacker-o2rHJLEEkT68y.toml:/tmp/ibpacker-o2rHJLEEkT68y.toml " +
						"ghcr.io/osbuild/image-builder-cli:latest build " +
						"--blueprint /tmp/ibpacker-o2rHJLEEkT68y.toml " +
						"--distro fedora minimal-raw && find ./output-hehwuXP6NyGIr -type f",
					Reply:  "Building...\nDone.\n",
					Status: 0,
				},
				{
					Request: "rm -f /tmp/ibpacker-o2rHJLEEkT68y.toml",
					Reply:   "",
					Status:  0,
				},
			},
		},
		{
			name: "stream9-raw-all-args-docker",
			cmd: &ibk.ContainerBootCommand{
				Repository: "quay.io/centos-bootc/centos-bootc:stream9",
				Type:       "raw",
				Arch:       "x86_64",
				Blueprint:  "blueprint",
				RootFS:     "btrfs",
				Common: ibk.CommonArgs{
					DryRun:      true,
					Interactive: true,
					TTY:         true,
					TeeLog:      true,
				},
			},
			session: []sshtest.RequestReply{
				{
					Request: "which podman",
					Reply:   "/usr/bin/podman\n",
					Status:  1,
				},
				{
					Request: "which docker",
					Reply:   "/usr/bin/docker\n",
					Status:  0,
				},
				{
					Request: "arch",
					Reply:   "x86_64\n",
					Status:  0,
				},
				{
					Request: "mkdir ./output-hehwuXP6NyGIr",
					Reply:   "",
					Status:  0,
				},
				{
					Request: "echo sudo /usr/bin/docker pull quay.io/centos-bootc/centos-bootc:stream9",
					Reply:   "",
					Status:  0,
				},
				{
					Request: "scp -t /tmp",
					Reply:   "",
					Status:  0,
				},
				{
					Request: "echo sudo /usr/bin/docker run --privileged --rm --pull=newer -i -t " +
						"--security-opt label=type:unconfined_t " +
						"-v /var/lib/containers/storage:/var/lib/containers/storage " +
						"-v ./output-hehwuXP6NyGIr:/output -v /tmp/ibpacker-o2rHJLEEkT68y.toml:/config.toml:ro " +
						"quay.io/centos-bootc/bootc-image-builder:latest " +
						"--type raw --local --rootfs btrfs " +
						"quay.io/centos-bootc/centos-bootc:stream9 2>&1 | tee ./output-hehwuXP6NyGIr/build.log && " +
						"find ./output-hehwuXP6NyGIr -type f",
					Reply:  "Building...\nDone.\n",
					Status: 0,
				},
				{
					Request: "rm -f /tmp/ibpacker-o2rHJLEEkT68y.toml",
					Reply:   "",
					Status:  0,
				},
			},
		},
		{
			name: "stream9-raw",
			cmd: &ibk.ContainerBootCommand{
				Repository: "quay.io/centos-bootc/centos-bootc:stream9",
				Type:       "raw",
				Arch:       "x86_64",
				Blueprint:  "blueprint",
				RootFS:     "btrfs",
			},
			session: []sshtest.RequestReply{
				{
					Request: "which podman",
					Reply:   "/usr/bin/podman\n",
					Status:  0,
				},
				{
					Request: "arch",
					Reply:   "x86_64\n",
					Status:  0,
				},
				{
					Request: "mkdir ./output-hehwuXP6NyGIr",
					Reply:   "",
					Status:  0,
				},
				{
					Request: "sudo /usr/bin/podman pull quay.io/centos-bootc/centos-bootc:stream9",
					Reply:   "",
					Status:  0,
				},
				{
					Request: "scp -t /tmp",
					Reply:   "",
					Status:  0,
				},
				{
					Request: "sudo /usr/bin/podman run --privileged --rm --pull=newer " +
						"--security-opt label=type:unconfined_t " +
						"-v /var/lib/containers/storage:/var/lib/containers/storage " +
						"-v ./output-hehwuXP6NyGIr:/output -v /tmp/ibpacker-o2rHJLEEkT68y.toml:/config.toml:ro " +
						"quay.io/centos-bootc/bootc-image-builder:latest " +
						"--type raw --local --rootfs btrfs " +
						"quay.io/centos-bootc/centos-bootc:stream9 && " +
						"find ./output-hehwuXP6NyGIr -type f",
					Reply:  "Building...\nDone.\n",
					Status: 0,
				},
				{
					Request: "rm -f /tmp/ibpacker-o2rHJLEEkT68y.toml",
					Reply:   "",
					Status:  0,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ibk.RandSource.Seed(0)

			server := sshtest.NewServerT(t, sshtest.TestSigner(t))
			server.Handler = sshtest.RequestReplyHandler(t, tt.session)
			defer server.Close()

			buf := &ibk.CombinedWriter{}
			client, err := ibk.NewSSHTransport(ibk.SSHTransportConfig{
				Host:        server.Endpoint,
				Username:    "test",
				Password:    "unused",
				PrivateKeys: []*bytes.Buffer{bytes.NewBufferString(sshtest.PrivateKey)},
				Stdout:      buf,
				Stderr:      buf,
			})
			if err != nil {
				t.Fatal(err)
			}
			defer client.Close(ctx)

			err = ibk.ApplyCommand(context.Background(), tt.cmd, client)
			if err != nil {
				t.Fatal(err)
			}

			if !strings.Contains(buf.String(), "Building") {
				t.Fatalf("unexpected output: %s", buf.String())
			}
		})
	}
}
