package ibk

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

// SSHTransportConfig is a configuration struct for creating a new SSHTransport.
type SSHTransportConfig struct {
	// Host is the hostname or IP address of the remote machine including optional port (e.g. "example.com:22").
	Host string

	// Username is the username to use for authentication.
	Username string

	// Password is the password to use for authentication. It is optional if private keys are provided.
	Password string

	// PrivateKeys is a list of private keys to use for authentication. If not provided, the default private keys will be used.
	PrivateKeys []*bytes.Buffer

	// KnownHosts is the path to the known hosts file. If not provided, the default known hosts file will be used.
	KnownHosts string

	// Timeout is the maximum amount of time a dial will wait for a connect to complete. The default is 10 seconds.
	Timeout time.Duration

	// Stdin is the standard input for the SSH session. The default is os.Stdin.
	Stdin io.Reader

	// Stdout is the standard output for the SSH session. The default is os.Stdout.
	Stdout io.Writer

	// Stderr is the standard error for the SSH session. The default is os.Stderr.
	Stderr io.Writer
}

var knownPrivateKeyPaths = []string{
	".ssh/id_rsa",
	".ssh/id_dsa",
	".ssh/id_ecdsa",
	".ssh/id_ed25519",
}

// ReadPrivateKeys reads the default private keys from the known paths.
func ReadPrivateKeys() ([]*bytes.Buffer, error) {
	var keys []*bytes.Buffer
	usr, _ := user.Current()
	homeDir := usr.HomeDir

	for _, path := range knownPrivateKeyPaths {
		f := filepath.Join(homeDir, path)
		if _, err := os.Stat(f); errors.Is(err, os.ErrNotExist) {
			continue
		}

		log.Printf("[DEBUG] Reading private key from %q", f)
		key, err := os.ReadFile(f)
		if err != nil {
			return nil, err
		}
		keys = append(keys, bytes.NewBuffer(key))
	}

	return keys, nil
}

// ReadPrivateKey reads a private key from the specified path.
func ReadPrivateKey(path string) (*bytes.Buffer, error) {
	key, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	return bytes.NewBuffer(key), nil
}

// SSHTransport is a struct that represents an SSH connection to a remote machine.
type SSHTransport struct {
	client   *ssh.Client
	stdin    io.Reader
	stdout   io.Writer
	stderr   io.Writer
	toDelete []string
}

var _ Pusher = (*SSHTransport)(nil)
var _ Executor = (*SSHTransport)(nil)
var _ Closer = (*SSHTransport)(nil)
var _ Transport = (*SSHTransport)(nil)

var ErrHostnameEmpty = errors.New("hostname is empty")
var ErrKnownHosts = errors.New("known hosts error")
var ErrSSHDial = errors.New("ssh dial error")
var ErrSSHNewSession = errors.New("ssh new session error")
var ErrCommand = errors.New("command error")
var ErrCopy = errors.New("copy error")

// NewSSHTransport creates a new SSHTransport with the given configuration.
// It immediatelly establishes a connection to the remote machine. Use Close to close the connection.
func NewSSHTransport(cfg SSHTransportConfig) (*SSHTransport, error) {
	clientConf := &ssh.ClientConfig{
		User: cfg.Username,
		Auth: make([]ssh.AuthMethod, 0),
	}

	if cfg.Timeout != 0 {
		clientConf.Timeout = cfg.Timeout
	} else {
		clientConf.Timeout = 10 * time.Second
	}

	if cfg.Password != "" {
		clientConf.Auth = append(clientConf.Auth, ssh.Password(cfg.Password))
	}

	if cfg.PrivateKeys == nil {
		keys, err := ReadPrivateKeys()
		if err != nil {
			return nil, fmt.Errorf("read private keys error: %w", err)
		}
		cfg.PrivateKeys = keys
	}

	for _, key := range cfg.PrivateKeys {
		signer, err := ssh.ParsePrivateKey(key.Bytes())
		if err != nil {
			return nil, fmt.Errorf("parse private key error: %w", err)
		}
		clientConf.Auth = append(clientConf.Auth, ssh.PublicKeys(signer))
	}

	if cfg.KnownHosts != "" {
		log.Printf("[DEBUG] Using known hosts file %q", cfg.KnownHosts)
		cb, err := knownhosts.New(cfg.KnownHosts)
		if err != nil {
			return nil, fmt.Errorf("known hosts '%s' error: %w", cfg.KnownHosts, err)
		}
		clientConf.HostKeyCallback = cb
	} else {
		clientConf.HostKeyCallback = noopCallback
	}

	if cfg.Host == "" {
		return nil, ErrHostnameEmpty
	}

	if cfg.Stdin == nil {
		cfg.Stdin = os.Stdin
	}

	if cfg.Stdout == nil {
		cfg.Stdout = os.Stdout
	}

	if cfg.Stderr == nil {
		cfg.Stderr = os.Stderr
	}

	if !strings.Contains(cfg.Host, ":") {
		cfg.Host = fmt.Sprintf("%s:22", cfg.Host)
	}

	log.Printf("[DEBUG] Connecting to %q", cfg.Host)
	client, err := ssh.Dial("tcp", cfg.Host, clientConf)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrSSHDial, err)
	}

	return &SSHTransport{
		client:   client,
		stdin:    cfg.Stdin,
		stdout:   cfg.Stdout,
		stderr:   cfg.Stderr,
		toDelete: make([]string, 0),
	}, nil
}

var noopCallback ssh.HostKeyCallback = func(_ string, _ net.Addr, _ ssh.PublicKey) error {
	return nil
}

// ExecuteOpt is a function that configures the SSH session for a command execution
type ExecuteOpt func(*ssh.Session)

// WithStdinWriter configures the SSH session to use the specified writer for standard input.
func WithCombinedWriter(w *CombinedWriter) ExecuteOpt {
	return func(s *ssh.Session) {
		s.Stdout = w
		s.Stderr = w
	}
}

// WithInputOutput configures the SSH session to use the specified reader and writers for standard input, output, and error.
func WithInputOutput(stdin io.Reader, stdout, stderr io.Writer) ExecuteOpt {
	return func(s *ssh.Session) {
		s.Stdin = stdin
		s.Stdout = stdout
		s.Stderr = stderr
	}
}

// Execute performs a command remotely via SSH session with standard input, output, and error configured
// as specified in the SSHTransportConfig. The command is executed in the remote machine.
// An optional arguments can be provided to override Stdout and Stderr config.
func (t *SSHTransport) Execute(ctx context.Context, cmd Command, opts ...ExecuteOpt) error {
	var err error

	s, err := t.client.NewSession()
	if err != nil {
		return fmt.Errorf("%w: %w", ErrSSHNewSession, err)
	}
	defer s.Close()

	s.Stdin = t.stdin
	s.Stdout = t.stdout
	s.Stderr = t.stderr

	for _, opt := range opts {
		opt(s)
	}

	command := cmd.Build()
	log.Printf("[DEBUG] Executing command %q", command)
	err = s.Start(command)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrCommand, err)
	}

	return Wait(ctx, func() error {
		return s.Wait()
	})
}

// Push copies the contents of a reader to a temporary file on the remote machine. Returns
// the path of the temporary file. The file(s) will be deleted when the SSH connection is closed.
func (t *SSHTransport) Push(ctx context.Context, contents, extension string) (string, error) {
	s, err := t.client.NewSession()
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrSSHNewSession, err)
	}
	defer s.Close()

	w, err := s.StdinPipe()
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrCopy, err)
	}

	if extension == "" {
		extension = "tmp"
	}
	targetDir := "/tmp"
	targetBaseFile := fmt.Sprintf("ibpacker-%s.%s", RandomString(13), extension)
	targetFile := filepath.Join(targetDir, targetBaseFile)
	t.toDelete = append(t.toDelete, targetFile)

	cmd := strings.Join([]string{"scp", "-t", targetDir}, " ")
	if err := s.Start(cmd); err != nil {
		w.Close()
		return "", fmt.Errorf("%w: %w", ErrCopy, err)
	}

	log.Printf("[DEBUG] Copying to temp file %q (size %d)", targetFile, len(contents))
	fmt.Fprintf(w, "C%#o %d %s\n", 0600, len(contents), targetBaseFile)
	io.Copy(w, strings.NewReader(contents))
	fmt.Fprint(w, "\x00")
	w.Close()

	err = Wait(ctx, func() error {
		return s.Wait()
	})

	return targetFile, err
}

// Close closes the SSH connection. Additionally, it deletes the temporary files created during the session.
func (t *SSHTransport) Close(ctx context.Context) error {
	s, err := t.client.NewSession()
	if err == nil {
		for _, file := range t.toDelete {
			log.Printf("[DEBUG] Deleting file %q", file)
			err := s.Run(fmt.Sprintf("rm -f %s", file))
			if err != nil {
				log.Printf("Failed to delete file %q: %v", file, err)
			}
		}

		s.Close()
	}

	if t.client != nil {
		return t.client.Close()
	}

	return nil
}
