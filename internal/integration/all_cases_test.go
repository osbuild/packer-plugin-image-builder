package integration

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"text/template"

	"github.com/lzap/packer-plugin-image-builder/internal/sshtest"
	yaml "gopkg.in/yaml.v2"
)

type TestCase struct {
	Fixtures    []sshtest.RequestReply `yaml:"fixtures"`
	Template    string                 `yaml:"template"`
	Environment []string               `yaml:"environment"`
	Result      TestResult             `yaml:"result"`
}

type TestResult struct {
	Grep   string `yaml:"grep"`
	Status int    `yaml:"status"`
}

type TestVars struct {
	Hostname string
}

const PluginPath = "../../build"

func TestIntegration(t *testing.T) {
	// check if plugin binary was built
	if _, err := os.Stat(filepath.Join(PluginPath, "plugins")); os.IsNotExist(err) {
		t.Skip("./build/plugins does not exist - do 'make build' first, skipping integration tests")
	}

	packerBin, err := exec.LookPath("packer")
	if err != nil {
		t.Skip("packer binary not found in PATH, skipping integration tests")
	}

	// list all YAML files in the directory with this source code
	files, err := os.ReadDir(".")
	if err != nil {
		t.Fatal(err)
	}

	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".yaml") {
			continue
		}

		// run test for each file
		t.Run(file.Name(), func(t *testing.T) {
			// open the YAML file
			f, err := os.Open(file.Name())
			if err != nil {
				t.Fatal(err)
			}
			defer f.Close()

			// decode the YAML file
			var tc TestCase
			var vars TestVars
			if err := yaml.NewDecoder(f).Decode(&tc); err != nil {
				t.Fatal(err)
			}

			// start a mock SSH server on a random port
			server := sshtest.NewServer(sshtest.TestSigner(t))
			server.Handler = sshtest.RequestReplyHandler(t, tc.Fixtures)
			defer server.Close()

			// set test case variables
			vars.Hostname = server.Endpoint

			// prepare a temporary file
			tempFile, err := os.CreateTemp("", "packer-test-template-*.pkr.hcl")
			if err != nil {
				t.Fatal(err)
			}
			defer func() {
				tempFile.Close()
				os.Remove(tempFile.Name())
			}()

			// render the template to the temporary file
			tmpl, err := template.New(file.Name()).Parse(tc.Template)
			if err != nil {
				t.Fatal(err)
			}
			err = tmpl.Execute(tempFile, vars)
			if err != nil {
				t.Fatal(err)
			}

			// run packer build
			command := []string{"build", tempFile.Name()}
			t.Logf("Running: packer %s", strings.Join(command, " "))

			cmd := exec.Command(packerBin, command...)
			cmd.Env = append(cmd.Environ(), "PACKER_LOG=1", "PACKER_PLUGIN_PATH=../../build")
			cmd.Env = append(cmd.Env, tc.Environment...)

			out, err := cmd.CombinedOutput()
			if err != nil {
				t.Logf("Command returned error: %s", err)
			}

			if tc.Result.Grep != "" {
				if !strings.Contains(string(out), tc.Result.Grep) {
					t.Fatalf("Expected output to contain %q, got:\n%s", tc.Result.Grep, out)
				}
			}

			if cmd.ProcessState.ExitCode() != tc.Result.Status {
				t.Fatalf("Expected exit code %d, got %d", tc.Result.Status, cmd.ProcessState.ExitCode())
			}

			t.Log(string(out))
		})
	}

}
