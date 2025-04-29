package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/exec"

	"golang.org/x/crypto/ssh"
)

func main() {
	// Path to server private key
	privateKeyPath := "id_rsa"

	// Read the private key
	privateBytes, err := os.ReadFile(privateKeyPath)
	if err != nil {
		log.Fatalf("Failed to load private key: %v", err)
	}

	// Create the Signer for this private key
	private, err := ssh.ParsePrivateKey(privateBytes)
	if err != nil {
		log.Fatalf("Failed to parse private key: %v", err)
	}

	// Configure SSH server
	config := &ssh.ServerConfig{
		PasswordCallback: func(conn ssh.ConnMetadata, password []byte) (*ssh.Permissions, error) {
			// Perform authentication for user
			if conn.User() == "test" && string(password) == "password" {
				return nil, nil
			}
			return nil, fmt.Errorf("password rejected for %q", conn.User())
		},
		PublicKeyCallback: func(conn ssh.ConnMetadata, key ssh.PublicKey) (*ssh.Permissions, error) {
			// Implement public key auth here if needed
			return nil, fmt.Errorf("public key auth not implemented")
		},
	}

	// Add private key to server config
	config.AddHostKey(private)

	// Listen on port 2222
	listener, err := net.Listen("tcp", "0.0.0.0:2222")
	if err != nil {
		log.Fatalf("Failed to listen on 2222: %v", err)
	}
	log.Printf("Listening on %s...", listener.Addr())

	// Accept connections
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Failed to accept incoming connection: %v", err)
			continue
		}
		go handleConnection(conn, config)
	}
}

func handleConnection(nConn net.Conn, config *ssh.ServerConfig) {
	defer nConn.Close()

	// Perform SSH handshake
	conn, chans, reqs, err := ssh.NewServerConn(nConn, config)
	if err != nil {
		log.Printf("Failed to handshake: %v", err)
		return
	}
	log.Printf("New SSH connection from %s (%s)", conn.RemoteAddr(), conn.ClientVersion())
	defer conn.Close()

	// Service the incoming SSH requests
	go ssh.DiscardRequests(reqs)

	// Service the incoming channels
	for newChannel := range chans {
		if newChannel.ChannelType() != "session" {
			newChannel.Reject(ssh.UnknownChannelType, "unknown channel type")
			continue
		}

		channel, requests, err := newChannel.Accept()
		if err != nil {
			log.Printf("Could not accept channel: %v", err)
			continue
		}

		go handleChannel(channel, requests)
	}
}

func handleChannel(channel ssh.Channel, requests <-chan *ssh.Request) {
	defer channel.Close()

	for req := range requests {
		switch req.Type {
		case "exec":
			// Extract command from payload
			var payload = struct{ Value string }{}
			if err := ssh.Unmarshal(req.Payload, &payload); err != nil {
				log.Printf("Failed to parse exec payload: %v", err)
				req.Reply(false, nil)
				continue
			}

			// Run the command
			cmd := exec.Command("sh", "-c", payload.Value)
			cmd.Stdout = channel
			cmd.Stderr = channel

			err := cmd.Run()
			if err != nil {
				log.Printf("Command execution error: %v", err)
			}

			// Return exit status
			var exitStatus = struct{ Status uint32 }{0}
			if err != nil {
				exitStatus.Status = 1
			}

			_, err = channel.SendRequest("exit-status", false, ssh.Marshal(&exitStatus))
			if err != nil {
				log.Printf("Failed to send exit status: %v", err)
			}

			return
		case "shell":
			// Start a shell
			cmd := exec.Command("sh")
			cmd.Env = os.Environ()
			
			stdin, err := cmd.StdinPipe()
			if err != nil {
				log.Printf("Failed to get stdin pipe: %v", err)
				req.Reply(false, nil)
				continue
			}
			
			stdout, err := cmd.StdoutPipe()
			if err != nil {
				log.Printf("Failed to get stdout pipe: %v", err)
				req.Reply(false, nil)
				continue
			}
			
			stderr, err := cmd.StderrPipe()
			if err != nil {
				log.Printf("Failed to get stderr pipe: %v", err)
				req.Reply(false, nil)
				continue
			}
			
			err = cmd.Start()
			if err != nil {
				log.Printf("Failed to start shell: %v", err)
				req.Reply(false, nil)
				continue
			}
			
			req.Reply(true, nil)
			
			// Pipe session to shell and vice versa
			go io.Copy(stdin, channel)
			go io.Copy(channel, stdout)
			go io.Copy(channel, stderr)
			
			if err := cmd.Wait(); err != nil {
				log.Printf("Shell exited with error: %v", err)
			}
			
			channel.SendRequest("exit-status", false, ssh.Marshal(&struct{ Status uint32 }{0}))
			return
		}

		// Reply to unhandled requests
		if req.WantReply {
			req.Reply(false, nil)
		}
	}
}