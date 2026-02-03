package runner

import "context"

type CommandResult struct {
	Command  string
	Stdout   string
	Stderr   string
	ExitCode int
}

type RunReport struct {
	Host     string
	Results  []CommandResult
	DurationSeconds int
}

type SSHCredentials struct {
	Username   string
	Password   string
	PrivateKey string
}

type WinRMCredentials struct {
	Username string
	Password string
	UseHTTPS bool
	Port     int
	Insecure bool
}

type Runner interface {
	RunSSH(ctx context.Context, host string, commands []string, credentials SSHCredentials) (RunReport, error)
	RunWinRM(ctx context.Context, host string, commands []string, credentials WinRMCredentials) (RunReport, error)
}
