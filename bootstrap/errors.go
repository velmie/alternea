package bootstrap

import (
	"strings"

	"github.com/pkg/errors"
)

type Error string

func (e Error) Error() string {
	return string(e)
}

const (
	ErrRequiredConfigurationMissed = Error("required configuration missed")
)

func errRequiredConfiguration(requester string, configName ...string) error {
	return errors.Wrapf(
		ErrRequiredConfigurationMissed,
		"%s requires '%s' configuration to be provided",
		requester,
		strings.Join(configName, "."),
	)
}
