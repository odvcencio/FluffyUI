// Package ssh provides an SSH server for serving FluffyUI applications remotely.
package ssh

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"sync"

	"golang.org/x/crypto/ssh"
)

// Config configures the SSH server.
type Config struct {
	Addr        string   // Listen address (default ":2222")
	HostKeyPath string   // Path to host key file (auto-generated if empty)
	NoAuth      bool     // Allow any connection (for development)
	Password    string   // Password authentication (if set)
	Command     []string // Command to run for each session
}

// Server is an SSH server that launches FluffyUI apps for each connection.
type Server struct {
	config    Config
	sshConfig *ssh.ServerConfig
	listener  net.Listener
	mu        sync.Mutex
	sessions  int
}

// New creates a new SSH server with the given config.
func New(cfg Config) (*Server, error) {
	if cfg.Addr == "" {
		cfg.Addr = ":2222"
	}

	sshCfg := &ssh.ServerConfig{}

	if cfg.NoAuth {
		sshCfg.NoClientAuth = true
	} else if cfg.Password != "" {
		sshCfg.PasswordCallback = func(c ssh.ConnMetadata, pass []byte) (*ssh.Permissions, error) {
			if string(pass) == cfg.Password {
				return nil, nil
			}
			return nil, fmt.Errorf("password rejected for %q", c.User())
		}
	} else {
		sshCfg.NoClientAuth = true // Default to no auth for dev
	}

	// Generate or load host key
	key, err := loadOrGenerateHostKey(cfg.HostKeyPath)
	if err != nil {
		return nil, fmt.Errorf("host key: %w", err)
	}
	sshCfg.AddHostKey(key)

	return &Server{config: cfg, sshConfig: sshCfg}, nil
}

// ListenAndServe starts the SSH server.
func (s *Server) ListenAndServe() error {
	ln, err := net.Listen("tcp", s.config.Addr)
	if err != nil {
		return err
	}
	s.listener = ln
	log.Printf("SSH server listening on %s", s.config.Addr)

	for {
		conn, err := ln.Accept()
		if err != nil {
			return err
		}
		go s.handleConn(conn)
	}
}

// Close stops the server.
func (s *Server) Close() error {
	if s.listener != nil {
		return s.listener.Close()
	}
	return nil
}

func (s *Server) handleConn(netConn net.Conn) {
	defer netConn.Close()

	sshConn, chans, reqs, err := ssh.NewServerConn(netConn, s.sshConfig)
	if err != nil {
		log.Printf("SSH handshake failed: %v", err)
		return
	}
	defer sshConn.Close()
	log.Printf("New SSH connection from %s (%s)", sshConn.RemoteAddr(), sshConn.ClientVersion())

	go ssh.DiscardRequests(reqs)

	for newChan := range chans {
		if newChan.ChannelType() != "session" {
			newChan.Reject(ssh.UnknownChannelType, "unknown channel type")
			continue
		}
		channel, requests, err := newChan.Accept()
		if err != nil {
			continue
		}
		go s.handleSession(channel, requests)
	}
}

func (s *Server) handleSession(channel ssh.Channel, requests <-chan *ssh.Request) {
	defer channel.Close()

	// Handle session requests (pty, shell, exec)
	for req := range requests {
		switch req.Type {
		case "shell", "exec":
			if req.WantReply {
				req.Reply(true, nil)
			}
			s.runCommand(channel)
			return
		case "pty-req":
			if req.WantReply {
				req.Reply(true, nil)
			}
		case "window-change":
			if req.WantReply {
				req.Reply(true, nil)
			}
		default:
			if req.WantReply {
				req.Reply(false, nil)
			}
		}
	}
}

func (s *Server) runCommand(channel ssh.Channel) {
	if len(s.config.Command) == 0 {
		fmt.Fprintln(channel, "No command configured")
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cmd := exec.CommandContext(ctx, s.config.Command[0], s.config.Command[1:]...)
	cmd.Env = append(os.Environ(), "TERM=xterm-256color")
	cmd.Stdin = channel
	cmd.Stdout = channel
	cmd.Stderr = channel

	if err := cmd.Run(); err != nil {
		log.Printf("Command exited: %v", err)
	}
	channel.SendRequest("exit-status", false, ssh.Marshal(struct{ Status uint32 }{0}))
}

func loadOrGenerateHostKey(path string) (ssh.Signer, error) {
	if path != "" {
		data, err := os.ReadFile(path)
		if err == nil {
			return ssh.ParsePrivateKey(data)
		}
	}

	// Generate ephemeral key
	_, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}

	return ssh.NewSignerFromKey(priv)
}
