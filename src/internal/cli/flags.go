package cli

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func FlagError(cmd *cobra.Command, err error) error {
	hint := fmt.Sprintf(
		"Run '%s --help' to see the available options.",
		cmd.CommandPath(),
	)

	var unknown *pflag.NotExistError
	if errors.As(err, &unknown) {
		return NewError(
			fmt.Sprintf(
				"unknown option %q for %q",
				specifiedOption(unknown.GetSpecifiedName(), unknown.GetSpecifiedShortnames()),
				cmd.CommandPath(),
			),
			hint,
		)
	}

	var missing *pflag.ValueRequiredError
	if errors.As(err, &missing) {
		return NewError(
			fmt.Sprintf(
				"option %q requires a value",
				specifiedOption(missing.GetSpecifiedName(), missing.GetSpecifiedShortnames()),
			),
			hint,
		)
	}

	var invalid *pflag.InvalidValueError
	if errors.As(err, &invalid) {
		option := "--" + invalid.GetFlag().Name
		if invalid.GetFlag().Value.Type() == "bool" {
			return NewError(
				fmt.Sprintf("invalid value %q for option %q", invalid.GetValue(), option),
				fmt.Sprintf(
					"Use %s without a value to enable it, or %s=false to disable it.",
					option,
					option,
				),
			)
		}
		return NewError(
			fmt.Sprintf(
				"invalid value %q for option %q: %v",
				invalid.GetValue(),
				option,
				errors.Unwrap(invalid),
			),
			hint,
		)
	}

	var syntax *pflag.InvalidSyntaxError
	if errors.As(err, &syntax) {
		return NewError(
			fmt.Sprintf("invalid option syntax %q", syntax.GetSpecifiedFlag()),
			hint,
		)
	}

	return NewError(err.Error(), hint)
}

func specifiedOption(name, shorthands string) string {
	if shorthands != "" {
		return "-" + name
	}
	return "--" + name
}
