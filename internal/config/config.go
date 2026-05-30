package config

type Config struct {
	RedisAddr string `flag:"redis-addr"`
	LogLevel  string `flag:"log-level"`
}
