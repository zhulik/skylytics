package flags

import (
	"fmt"
	"slices"

	"github.com/urfave/cli/v3"
)

var validLogLevels = []string{"debug", "info", "warn", "error"}

var RedisAddr = &cli.StringFlag{
	Name:    "redis-addr",
	Aliases: []string{"r"},
	Usage:   "The address of the Redis server",
	Value:   "localhost:6379",
	Sources: cli.EnvVars("REDIS_ADDR"),
}

// TODO: extract custom EnumFlag
var LogLevel = &cli.StringFlag{
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
