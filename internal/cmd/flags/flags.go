package flags

import (
	"fmt"
	"slices"

	libnats "github.com/nats-io/nats.go"
	"github.com/urfave/cli/v3"
)

var validLogLevels = []string{"debug", "info", "warn", "error"}

var NATS_URL = &cli.StringFlag{
	Name:    "nats-url",
	Aliases: []string{"n"},
	Usage:   "The URL of the NATS server",
	Value:   libnats.DefaultURL,
	Sources: cli.EnvVars("NATS_URL"),
}

var INIT_NATS = &cli.BoolFlag{
	Name:        "nats-init",
	Aliases:     []string{"i"},
	Usage:       "Initialize the NATS server: create streams, consumers, etc.",
	DefaultText: "false",
	Value:       false,
	Sources:     cli.EnvVars("NATS_INIT"),
}

// TODO: extract custom EnumFlag
var LOG_LEVEL = &cli.StringFlag{
	Name:    "log-level",
	Aliases: []string{"l"},
	Usage:   "The level of the logs",
	Value:   "info",
	Validator: func(value string) error {
		if !slices.Contains(validLogLevels, value) {
			return fmt.Errorf("invalid log level: %s, allowed values are: %s", value, validLogLevels)
		}
		return nil
	},
	Sources: cli.EnvVars("LOG_LEVEL"),
}
