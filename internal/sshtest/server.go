package sshtest

import (
	"fmt"
	"net"
	"regexp"
	"sync"

	"github.com/google/go-cmp/cmp"
	"golang.org/x/crypto/ssh"
)

type Server struct {
	Endpoint string
	Listener net.Listener
	Config   *ssh.ServerConfig
	Handler  func(ssh.Channel, <-chan *ssh.Request)

	t      TestLogger
	mu     sync.Mutex
	closed bool
}

func NewServerT(t TestLogger, hostKey ssh.Signer) *Server {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		if ln, err = net.Listen("tcp6", "[::1]:0"); err != nil {
			panic(fmt.Sprintf("sshtest: failed to listen on a port: %v", err))
		}
	}
	s := &Server{
		Listener: ln,
		Endpoint: ln.Addr().String(),
		Handler:  NullHandler,
		t:        t,
	}
	s.Config = &ssh.ServerConfig{NoClientAuth: true}
	s.Config.AddHostKey(hostKey)
	s.start()

	return s
}

func NewServer(hostKey ssh.Signer) *Server {
	return NewServerT(noopLogger{}, hostKey)
}

func (s *Server) start() {
	if s.Config == nil {
		s.t.Fatalf("sshtest: server config is not set")
	}

	go func() {
		for {
			serverConn, err := s.Listener.Accept()
			if err != nil {
				s.mu.Lock()
				if s.closed {
					s.mu.Unlock()
					return
				}
				s.mu.Unlock()
				continue
			}
			go func() {
				defer serverConn.Close()

				_, chans, reqs, err := ssh.NewServerConn(serverConn, s.Config)
				if err != nil {
					return
				}

				go ssh.DiscardRequests(reqs)
				for newCh := range chans {
					if newCh.ChannelType() != "session" {
						newCh.Reject(ssh.UnknownChannelType, "unknown channel type")
						continue
					}
					ch, inReqs, err := newCh.Accept()
					if err != nil {
						continue
					}
					if s.Handler == nil {
						NullHandler(ch, inReqs)
						continue
					}
					s.Handler(ch, inReqs)
				}
			}()
		}
	}()
}

func (s *Server) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return
	}
	s.closed = true
	s.Listener.Close()
}

func sendStatus(ch ssh.Channel, code uint32) error {
	var statusMsg = struct {
		Status uint32
	}{
		Status: code,
	}
	_, err := ch.SendRequest("exit-status", false, ssh.Marshal(&statusMsg))
	return err
}

func NullHandler(ch ssh.Channel, in <-chan *ssh.Request) {
	defer ch.Close()

	req, ok := <-in
	if !ok {
		return
	}
	req.Reply(true, nil)

	sendStatus(ch, 0)
}

// RequestReply is a single request-reply pair.
type RequestReply struct {
	// Request is a regular expression that matches the request.
	Request string `yaml:"request"`

	// Reply is a string that is sent back to the client.
	Reply string `yaml:"reply"`

	// Status is the ssh exit status code.
	Status uint32 `yaml:"status,omitempty"`

	rr *regexp.Regexp
}

// RequestReplyHandler returns a handler that replies to requests. It fails if
// the request does not match the expected request.
func RequestReplyHandler(t TestLogger, replies []RequestReply) func(ssh.Channel, <-chan *ssh.Request) {
	i := 0

	for i, r := range replies {
		replies[i].rr = regexp.MustCompile(r.Request)
	}

	return func(ch ssh.Channel, in <-chan *ssh.Request) {
		defer ch.Close()

		req, ok := <-in
		if !ok {
			return
		}

		switch req.Type {
		case "shell", "exec", "subsystem":
			var payload = struct{ Value string }{}
			ssh.Unmarshal(req.Payload, &payload)
			t.Logf("ssh %s: %s", req.Type, payload.Value)

			if i >= len(replies) {
				t.Fatalf("unexpected ssh request: %s, payload: %s", req.Type, payload.Value)
			}

			if !replies[i].rr.MatchString(payload.Value) {
				if diff := cmp.Diff(replies[i].Request, payload.Value); diff != "" {
					t.Fatalf("unexpected ssh request: %s, payload: %s, diff: %s", req.Type, payload.Value, diff)
				}
			}

			req.Reply(true, nil)
			sendStatus(ch, replies[i].Status)

			if replies[i].Reply != "" {
				ch.Write([]byte(replies[i].Reply))
			}
			i++
		default:
			t.Logf("ssh %s: ignored", req.Type)
			req.Reply(true, nil)
			sendStatus(ch, 0)
		}
	}
}
