package ssh

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"net"
	"os/exec"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
)

// RunCommand will execute arbitrary command against host and port with passed client config
func RunCommand(cmd, host string, port int, sshCfg *ssh.ClientConfig) (io.Reader, error) {
	sshClient, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", host, port), sshCfg)
	if err != nil {
		return nil, fmt.Errorf("error creating ssh client for host '%s': %s", host, err)
	}
	defer sshClient.Close()

	session, err := sshClient.NewSession()
	if err != nil {
		return nil, fmt.Errorf("error creating a session from ssh client for host '%s': %s", host, err)
	}
	defer session.Close()

	stdOut := &bytes.Buffer{}
	session.Stdout = stdOut

	if err = session.Run(cmd); err != nil {
		return nil, fmt.Errorf("error executing command '%s' for host '%s': %s", cmd, host, err)
	}

	return stdOut, nil
}

// GetDefaultClientConfig will return a *ssh.ClientConfig instance with default settings based on
// the parameters passed
func GetDefaultClientConfig(username, password string, timeout int, cb ssh.HostKeyCallback) *ssh.ClientConfig {
	return &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
			ssh.KeyboardInteractive(func(user, instruction string, questions []string, echos []bool) (answers []string, err error) {
				if len(questions) == 0 {
					return nil, nil
				}

				return []string{password}, nil
			}),
		},
		HostKeyCallback: cb,
		Timeout:         time.Duration(timeout) * time.Second,
	}
}

// GetDefaultHostKeyCallback gets a default ssh.HostKeyCallback function that
// will validate the given hosts public ssh key
func GetDefaultHostKeyCallback(knownHostsFilePath string) ssh.HostKeyCallback {
	return func(hostname string, remote net.Addr, key ssh.PublicKey) error {
		stdOut, err := GetKnownHostsOutput(knownHostsFilePath, strings.Split(hostname, ":")[0])
		if err != nil {
			return fmt.Errorf("error retrieving output to verify host '%s': %s", strings.Split(hostname, ":")[0], err)
		}

		verified := false

		for {
			line, err := stdOut.ReadString('\n')
			if err != nil {
				if err != io.EOF {
					return fmt.Errorf("error reading from stdout: %s", err)
				}

				break
			}

			lineArr := strings.Split(line, " ")

			if len(lineArr) == 3 {
				if base64.RawStdEncoding.EncodeToString(key.Marshal()) != lineArr[2] {
					verified = true
					break
				}
			}
		}

		if !verified {
			return fmt.Errorf(fmt.Sprintf("given hostname '%s' was not found in given known_hosts file", hostname))
		}

		return nil
	}
}

// GetKnownHostsOutput takes known_hosts file path, along with a hostname to search for
// and returns the output of the search or error if not found
func GetKnownHostsOutput(knownHostsFilePath, hostname string) (*bytes.Buffer, error) {
	stdOut := &bytes.Buffer{}
	stdErr := &bytes.Buffer{}
	cmd := exec.Command("ssh-keygen", "-F", hostname, "-f", knownHostsFilePath)
	cmd.Stdout = stdOut
	cmd.Stderr = stdErr

	var err error

	if err = cmd.Run(); err != nil {
		if stdErr.String() != "" {
			return nil, fmt.Errorf("error running 'ssh-keygen' command: %s", stdErr.String())
		}
		if stdOut.String() == "" {
			return nil, fmt.Errorf(fmt.Sprintf("given hostname '%s' was not found in given known_hosts file", hostname))
		}
	}

	return stdOut, nil
}
