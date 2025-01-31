package sshtest

import (
	"bytes"
	"encoding/base64"

	"golang.org/x/crypto/ssh"
)

// notsecret
var PrivateKey = `-----BEGIN OPENSSH PRIVATE KEY-----
b3BlbnNzaC1rZXktdjEAAAAABG5vbmUAAAAEbm9uZQAAAAAAAAABAAAAMwAAAAtzc2gtZW
QyNTUxOQAAACCXwb4kU7ktwVZ+a3XvcUBa5spbB2/HCXaY7iSTVczCIwAAALBAwllAQMJZ
QAAAAAtzc2gtZWQyNTUxOQAAACCXwb4kU7ktwVZ+a3XvcUBa5spbB2/HCXaY7iSTVczCIw
AAAEDOtzrjmAq0+5qpnNLheYHdAVagfVoDBEazwrOqdpfiZJfBviRTuS3BVn5rde9xQFrm
ylsHb8cJdpjuJJNVzMIjAAAAKGJ1aWxkZXJAenp6YXAudHBiLmxhYi5lbmcuYnJxLnJlZG
hhdC5jb20BAgMEBQ==
-----END OPENSSH PRIVATE KEY-----` // notsecret

var PublicKey = `ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIJfBviRTuS3BVn5rde9xQFrmylsHb8cJdpjuJJNVzMIj user@example.com`

func TestPrivateKeyAsSlice() []byte {
	return []byte(PrivateKey)
}

func TestPublicKeyAsSlice() []byte {
	return []byte(PublicKey)
}

func TestSigner(t TestLogger) ssh.Signer {
	signer, err := ssh.ParsePrivateKey(TestPrivateKeyAsSlice())
	if err != nil {
		t.Fatalf("failed to parse private key: %v", err)
	}

	return signer
}

func TestPublicKey() ssh.PublicKey {
	parts := bytes.SplitN(TestPublicKeyAsSlice(), []byte(" "), 3)
	if len(parts) < 2 {
		return nil
	}

	decoded, err := base64.StdEncoding.DecodeString(string(parts[1]))
	if err != nil {
		return nil
	}

	pub, err := ssh.ParsePublicKey(decoded)
	if err != nil {
		return nil
	}

	return pub
}
