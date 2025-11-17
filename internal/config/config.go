package config

type Config struct {
	NATSURL  string `flag:"nats-url"`
	NATSInit bool   `flag:"nats-init"`
	LogLevel string `flag:"log-level"`
}
