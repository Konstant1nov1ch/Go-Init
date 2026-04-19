package config

import (
	"flag"
	"log"

	"gitlab.com/go-init/go-init-common/default/file"

	"gopkg.in/yaml.v2"
)

// OpenConfig загружает конфигурацию из YAML-файла.
func OpenConfig(cfg interface{}) {
	var filePath string
	flag.StringVar(&filePath, "config", "./config.yml", "Path to configuration file")
	flag.Parse()

	f, err := file.OpenFile(filePath)
	if err != nil {
		log.Fatalf("Failed to open config file %s: %v", filePath, err)
	}
	defer f.Close()

	decoder := yaml.NewDecoder(f)
	if err := decoder.Decode(cfg); err != nil {
		log.Fatalf("Invalid config file %s: %v", filePath, err)
	}
}
