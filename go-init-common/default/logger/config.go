package logger

type Config struct {
	Level  string `yaml:"level" default:"INFO"`
	Format string `yaml:"format" default:"json"`
}
