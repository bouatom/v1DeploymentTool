package policy

import "v1-sg-deployment-tool/internal/errors"

type FailureAction string

const (
	ActionRetrySame     FailureAction = "retry_same"
	ActionSwitchAuth    FailureAction = "switch_auth"
	ActionSwitchInstall FailureAction = "switch_install"
	ActionAbort         FailureAction = "abort"
)

func DecideNextAction(code errors.Code, remainingAuth bool, remainingInstall bool) FailureAction {
	switch code {
	case errors.CodeAuthDenied, errors.CodeAuthTimeout:
		if remainingAuth {
			return ActionSwitchAuth
		}
		return ActionAbort
	case errors.CodeInstallFailed:
		if remainingInstall {
			return ActionSwitchInstall
		}
		return ActionAbort
	case errors.CodeNetworkIssue:
		return ActionRetrySame
	default:
		return ActionAbort
	}
}
