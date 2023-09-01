package config

import (
	"os"

	"github.com/labstack/gommon/log"
	"gopkg.in/yaml.v3"
)

type conf struct {
	REDIS struct {
		Protocol string `yaml:"protocol"`
		Host     string `yaml:"host"`
		Password string `yaml:"password"`
		Port     string `yaml:"port"`
		Expires  int    `yaml:"expires"`
		MaxIdle  int    `yaml:"max_idle"`
	}
}

var Config *conf

func Load(src string) error {
	if Config != nil {
		log.Info("Config already loaded")
		return nil
	}
	file, err := os.Open(src)
	if err != nil {
		return err
	}
	defer file.Close()
	d := yaml.NewDecoder(file)
	if err := d.Decode(&Config); err != nil {
		return err
	}
	return nil
}
