package server

import "time"

type Config struct {
	Name        string        `yaml:"name" default:"default-name"`
	Version     string        `yaml:"version" default:"0.0.0"`
	Env         string        `yaml:"environment" default:"development"`
	Port        string        `yaml:"port" default:"8080"`
	Timeout     time.Duration `yaml:"timeout" default:"10s"`
	IdleTimeout time.Duration `yaml:"idle_timeout" default:"60s"`
}
