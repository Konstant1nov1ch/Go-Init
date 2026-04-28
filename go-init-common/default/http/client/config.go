package client

import "time"

// Config содержит настройки HTTP-клиента.
type Config struct {
	EndpointUrl         string        `yaml:"url"`
	UserAgent           string        `yaml:"user_agent"`
	Timeout             time.Duration `yaml:"timeout" default:"3s"`
	MaxIdleConns        int           `yaml:"max_idle_conns" default:"100"`
	MaxConnsPerHost     int           `yaml:"max_conns_per_host" default:"100"`
	MaxIdleConnsPerHost int           `yaml:"max_idle_conns_per_host" default:"100"`
	IdleConnTimeout     time.Duration `yaml:"idle_conn_timeout" default:"60s"`
}
