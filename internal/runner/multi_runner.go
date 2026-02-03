package runner

import "context"

type MultiRunner struct {
	SSH   SSHRunner
	WinRM WinRMRunner
}

func (runner MultiRunner) RunSSH(ctx context.Context, host string, commands []string, credentials SSHCredentials) (RunReport, error) {
	return runner.SSH.RunSSH(ctx, host, commands, credentials)
}

func (runner MultiRunner) RunWinRM(ctx context.Context, host string, commands []string, credentials WinRMCredentials) (RunReport, error) {
	return runner.WinRM.RunWinRM(ctx, host, commands, credentials)
}
