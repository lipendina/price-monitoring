package config

import (
	"golang.org/x/xerrors"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
)

const confPath = "config/parameters.yaml"

type DBConfig struct {
	User     string `yaml:"db_user"`
	Password string `yaml:"db_password"`
	Host     string `yaml:"db_host"`
	Port     uint16 `yaml:"db_port"`
	DBName   string `yaml:"db_name"`
}

type ApplicationConfig struct {
	DB       DBConfig `yaml:",inline"`
	HTTPPort uint16   `yaml:"http_port"`
	Domain   string   `yaml:"domain"`
}

type EmailServer struct {
	Server   string `yaml:"email_server"`
	Name     string `yaml:"email_name"`
	Password string `yaml:"email_password"`
	Port     int    `yaml:"email_port"`
}

func ParseApplicationConfig() (*ApplicationConfig, error) {
	config := &ApplicationConfig{}
	err := ParseConfig(config)
	if err != nil {
		return nil, err
	}
	return config, nil
}

func ParseEmailServerConfig() (*EmailServer, error) {
	config := &EmailServer{}
	err := ParseConfig(config)
	if err != nil {
		return nil, err
	}
	return config, nil
}

func ParseConfig(config interface{}) error {
	confFile, err := ioutil.ReadFile(confPath)
	if err != nil {
		return xerrors.Errorf("Failed to read config file: %+v", err)
	}

	confFile = []byte(os.ExpandEnv(string(confFile)))

	if err = yaml.Unmarshal(confFile, config); err != nil {
		return xerrors.Errorf("Cannot unmarshal config: %+v", err)
	}

	return nil
}
