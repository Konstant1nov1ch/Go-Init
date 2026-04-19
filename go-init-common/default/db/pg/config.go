package pg

import (
	"strings"
	"time"
)

type Config struct {
	Host        string `yaml:"host"`
	Port        string `yaml:"port"`
	Name        string `yaml:"database_name"`
	User        string `yaml:"user"`
	Password    string `yaml:"password"`
	SslMode     string `yaml:"ssl" default:"disable"`
	Timezone    string `yaml:"timezone" default:"Europe/Moscow"`
	AutoMigrate bool   `yaml:"auto_migrate"`
	Schema      string `yaml:"schema" default:"public"`
	// InitTimeout — ожидание первого подключения GORM (0 = 30s в коде ORM).
	InitTimeout time.Duration `yaml:"init_timeout"`
}

func BuildDsn(c *Config) string {
	var dsnParts []string

	dsnParts = append(dsnParts, "host="+c.Host)
	dsnParts = append(dsnParts, "port="+c.Port)
	dsnParts = append(dsnParts, "dbname="+c.Name)
	dsnParts = append(dsnParts, "user="+c.User)
	dsnParts = append(dsnParts, "password="+c.Password)
	dsnParts = append(dsnParts, "TimeZone="+c.Timezone)
	ssl := c.SslMode
	if ssl == "" {
		ssl = "disable"
	}
	dsnParts = append(dsnParts, "sslmode="+ssl)

	return strings.Join(dsnParts, " ")
}
