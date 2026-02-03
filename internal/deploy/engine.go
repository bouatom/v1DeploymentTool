package deploy

import (
	"context"
	stdErrors "errors"
	"time"

	"v1-sg-deployment-tool/internal/auth"
	domainErrors "v1-sg-deployment-tool/internal/errors"
	"v1-sg-deployment-tool/internal/models"
	"v1-sg-deployment-tool/internal/policy"
	"v1-sg-deployment-tool/internal/runner"
)

type CredentialSet struct {
	SSH   runner.SSHCredentials
	WinRM runner.WinRMCredentials
}

type ExecutionResult struct {
	Method        auth.Method
	Report        runner.RunReport
	ErrorDetail   *domainErrors.Detail
}

type Engine struct {
	Runner runner.Runner
}

func (engine Engine) Execute(ctx context.Context, host string, os models.TargetOS, request InstallRequest, creds CredentialSet) (ExecutionResult, error) {
	if engine.Runner == nil {
		return ExecutionResult{}, stdErrors.New("runner is required")
	}

	plan, err := BuildPlan(request)
	if err != nil {
		return ExecutionResult{}, err
	}

	order := auth.OrderForOS(os)
	if len(order) == 0 {
		return ExecutionResult{}, stdErrors.New("unsupported os")
	}

	for index, method := range order {
		report, detail, err := engine.executeWithMethod(ctx, method, host, plan, creds)
		if err == nil {
			return ExecutionResult{
				Method: method,
				Report: report,
			}, nil
		}

		remainingAuth := index < len(order)-1

		if detail == nil {
			detail = &domainErrors.Detail{
				Code:        domainErrors.CodeInstallFailed,
				Message:     "deployment failed",
				Remediation: domainErrors.RemediationFor(domainErrors.CodeInstallFailed),
			}
		}

		action := policy.DecideNextAction(detail.Code, remainingAuth, false)
		if action == policy.ActionSwitchAuth {
			continue
		}

		return ExecutionResult{
			Method:      method,
			Report:      report,
			ErrorDetail: detail,
		}, err
	}

	return ExecutionResult{}, stdErrors.New("all auth methods failed")
}

func (engine Engine) executeWithMethod(ctx context.Context, method auth.Method, host string, plan DeployPlan, creds CredentialSet) (runner.RunReport, *domainErrors.Detail, error) {
	runCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	switch method {
	case auth.MethodSSHKey, auth.MethodSSHPassword:
		report, err := engine.Runner.RunSSH(runCtx, host, plan.Commands, creds.SSH)
		if err != nil {
			return report, buildDetailFromReport("SSH", report), err
		}
		return report, nil, nil
	case auth.MethodWinRMHTTPSCert, auth.MethodWinRMHTTPSUserPW:
		creds.WinRM.UseHTTPS = true
		report, err := engine.Runner.RunWinRM(runCtx, host, plan.Commands, creds.WinRM)
		if err != nil {
			return report, buildDetailFromReport("WinRM", report), err
		}
		return report, nil, nil
	default:
		return runner.RunReport{}, &domainErrors.Detail{
			Code:        domainErrors.CodeUnsupportedOS,
			Message:     "Unsupported auth method",
			Remediation: domainErrors.RemediationFor(domainErrors.CodeUnsupportedOS),
		}, stdErrors.New("unsupported auth method")
	}
}

func buildDetailFromReport(prefix string, report runner.RunReport) *domainErrors.Detail {
	if hasNonZeroExit(report.Results) {
		return &domainErrors.Detail{
			Code:        domainErrors.CodeInstallFailed,
			Message:     prefix + " command execution failed",
			Remediation: domainErrors.RemediationFor(domainErrors.CodeInstallFailed),
		}
	}

	return &domainErrors.Detail{
		Code:        domainErrors.CodeAuthDenied,
		Message:     prefix + " authentication failed",
		Remediation: domainErrors.RemediationFor(domainErrors.CodeAuthDenied),
	}
}

func hasNonZeroExit(results []runner.CommandResult) bool {
	for _, result := range results {
		if result.ExitCode != 0 {
			return true
		}
	}
	return false
}
