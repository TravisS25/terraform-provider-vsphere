package ssh

import (
	"bytes"
	"fmt"
	"io"
	"time"

	"golang.org/x/crypto/ssh"
)

func RunCommand(cmd, host, username, password string, port, timeout int) (io.Reader, error) {
	sshClient, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", host, port), &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{
			ssh.KeyboardInteractive(func(user, instruction string, questions []string, echos []bool) (answers []string, err error) {
				if len(questions) == 0 {
					return nil, nil
				}

				return []string{password}, nil
			}),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         time.Duration(timeout) * time.Second,
	})
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
