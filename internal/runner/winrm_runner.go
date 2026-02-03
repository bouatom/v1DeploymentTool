package runner

import (
	"context"
	"errors"
	"time"

	"github.com/masterzen/winrm"
)

type WinRMRunner struct {
	Timeout time.Duration
}

func (runner WinRMRunner) RunWinRM(ctx context.Context, host string, commands []string, credentials WinRMCredentials) (RunReport, error) {
	if host == "" {
		return RunReport{}, errors.New("host is required")
	}
	if credentials.Username == "" || credentials.Password == "" {
		return RunReport{}, errors.New("winrm username and password are required")
	}
	if len(commands) == 0 {
		return RunReport{}, errors.New("commands are required")
	}

	port := credentials.Port
	if port == 0 {
		if credentials.UseHTTPS {
			port = 5986
		} else {
			port = 5985
		}
	}

	timeout := runner.Timeout
	if timeout <= 0 {
		timeout = 8 * time.Second
	}

	endpoint := winrm.NewEndpoint(host, port, credentials.UseHTTPS, credentials.Insecure, nil, nil, nil, timeout)
	client, err := winrm.NewClient(endpoint, credentials.Username, credentials.Password)
	if err != nil {
		return RunReport{}, err
	}

	start := time.Now().UTC()
	var results []CommandResult

	for _, command := range commands {
		if err := ctx.Err(); err != nil {
			return RunReport{}, err
		}

		stdout, stderr, exitCode, err := client.RunWithContextWithString(ctx, command, "")
		results = append(results, CommandResult{
			Command:  command,
			Stdout:   stdout,
			Stderr:   stderr,
			ExitCode: exitCode,
		})
		if err != nil {
			return RunReport{
				Host:            host,
				Results:         results,
				DurationSeconds: int(time.Since(start).Seconds()),
			}, err
		}
		if exitCode != 0 {
			return RunReport{
				Host:            host,
				Results:         results,
				DurationSeconds: int(time.Since(start).Seconds()),
			}, errors.New("winrm command failed")
		}
	}

	return RunReport{
		Host:            host,
		Results:         results,
		DurationSeconds: int(time.Since(start).Seconds()),
	}, nil
}
