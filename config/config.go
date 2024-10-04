package config

import (
	"log"
	"os"

	yaml "gopkg.in/yaml.v2"
)

type Config struct {
	Service struct {
		Master struct {
			Port            int `yaml:"port"`
			NodePortInitial int `yaml:"node_port_initial"`
		} `yaml:"master"`
		Nodes struct {
			MinCount int `yaml:"min_count"`
			MaxCount int `yaml:"max_count"`
		} `yaml:"nodes"`
		Logs struct {
			Dir string `yaml:"dir"`
		} `yaml:"logs"`
	} `yaml:"service"`
}

func ReadConfig() (*Config, error) {
	yamlFile, err := os.ReadFile("config/config.yaml")
	if err != nil {
		log.Fatalf("Error reading YAML file: %v", err)
	}

	var config Config
	err = yaml.Unmarshal(yamlFile, &config)
	if err != nil {
		log.Fatalf("Error parsing YAML: %v", err)
	}
	return &config, err
}
