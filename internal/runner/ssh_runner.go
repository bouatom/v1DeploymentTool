package runner

import (
	"bytes"
	"context"
	"errors"
	"net"
	"time"

	"golang.org/x/crypto/ssh"
)

type SSHRunner struct {
	Timeout time.Duration
}

func (runner SSHRunner) RunSSH(ctx context.Context, host string, commands []string, credentials SSHCredentials) (RunReport, error) {
	if host == "" {
		return RunReport{}, errors.New("host is required")
	}
	if credentials.Username == "" {
		return RunReport{}, errors.New("ssh username is required")
	}
	if len(commands) == 0 {
		return RunReport{}, errors.New("commands are required")
	}

	authMethods := []ssh.AuthMethod{}
	if credentials.PrivateKey != "" {
		signer, err := ssh.ParsePrivateKey([]byte(credentials.PrivateKey))
		if err != nil {
			return RunReport{}, err
		}
		authMethods = append(authMethods, ssh.PublicKeys(signer))
	}
	if credentials.Password != "" {
		authMethods = append(authMethods, ssh.Password(credentials.Password))
	}
	if len(authMethods) == 0 {
		return RunReport{}, errors.New("no ssh credentials provided")
	}

	timeout := runner.Timeout
	if timeout <= 0 {
		timeout = 8 * time.Second
	}

	config := &ssh.ClientConfig{
		User:            credentials.Username,
		Auth:            authMethods,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         timeout,
	}

	address := net.JoinHostPort(host, "22")
	conn, err := ssh.Dial("tcp", address, config)
	if err != nil {
		return RunReport{}, err
	}
	defer conn.Close()

	start := time.Now().UTC()
	var results []CommandResult
	for _, command := range commands {
		if err := ctx.Err(); err != nil {
			return RunReport{}, err
		}

		session, err := conn.NewSession()
		if err != nil {
			return RunReport{}, err
		}

		var stdout bytes.Buffer
		var stderr bytes.Buffer
		session.Stdout = &stdout
		session.Stderr = &stderr

		runErr := session.Run(command)
		exitCode := 0
		if runErr != nil {
			if exitError, ok := runErr.(*ssh.ExitError); ok {
				exitCode = exitError.ExitStatus()
			} else {
				exitCode = 1
			}
		}
		_ = session.Close()

		results = append(results, CommandResult{
			Command:  command,
			Stdout:   stdout.String(),
			Stderr:   stderr.String(),
			ExitCode: exitCode,
		})

		if runErr != nil {
			return RunReport{
				Host:             host,
				Results:          results,
				DurationSeconds:  int(time.Since(start).Seconds()),
			}, runErr
		}
	}

	return RunReport{
		Host:            host,
		Results:         results,
		DurationSeconds: int(time.Since(start).Seconds()),
	}, nil
}
