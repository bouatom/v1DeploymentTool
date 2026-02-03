package targets

import (
	"errors"
	"net"
	"strings"
)

type TargetKind string

const (
	TargetKindHostname TargetKind = "hostname"
	TargetKindIP       TargetKind = "ip"
	TargetKindCIDR     TargetKind = "cidr"
)

type TargetSpec struct {
	Kind     TargetKind
	Hostname string
	IP       net.IP
	CIDR     *net.IPNet
	Original string
}

func ParseInputs(values []string) ([]TargetSpec, []error) {
	if len(values) == 0 {
		return nil, []error{errors.New("no targets provided")}
	}

	var specs []TargetSpec
	var errs []error

	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			errs = append(errs, errors.New("empty target value"))
			continue
		}

		if ip := net.ParseIP(trimmed); ip != nil {
			specs = append(specs, TargetSpec{
				Kind:     TargetKindIP,
				IP:       ip,
				Original: trimmed,
			})
			continue
		}

		if _, ipNet, err := net.ParseCIDR(trimmed); err == nil {
			specs = append(specs, TargetSpec{
				Kind:     TargetKindCIDR,
				CIDR:     ipNet,
				Original: trimmed,
			})
			continue
		}

		if isHostname(trimmed) {
			specs = append(specs, TargetSpec{
				Kind:     TargetKindHostname,
				Hostname: trimmed,
				Original: trimmed,
			})
			continue
		}

		errs = append(errs, errors.New("invalid target: "+trimmed))
	}

	return specs, errs
}

func isHostname(value string) bool {
	if len(value) > 253 {
		return false
	}

	labels := strings.Split(value, ".")
	if len(labels) == 0 {
		return false
	}

	for _, label := range labels {
		if label == "" || len(label) > 63 {
			return false
		}
		for i, r := range label {
			isFirstOrLast := i == 0 || i == len(label)-1
			if r == '-' && isFirstOrLast {
				return false
			}
			if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' {
				continue
			}
			return false
		}
	}

	return true
}
