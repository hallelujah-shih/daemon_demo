package config

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
)

type ProcessConfig struct {
	SoftLimit struct {
		LimitFileDescriptor int `json:"limit_file_descriptor" yaml:"limit_file_descriptor"`
	} `json:"soft_limit" yaml:"soft_limit"`
	UidGid      string   `json:"uid_gid" yaml:"uid_gid"`
	Envs        []string `json:"envs" yaml:"envs"`
	CommandName string   `json:"command" yaml:"command"`
	CommandArgs []string `json:"command_args" yaml:"command_args"`
	StderrPath  string   `json:"stderr_path" yaml:"stderr_path"`
	StdoutPath  string   `json:"stdout_path" yaml:"stdout_path"`
}

type ManagerConfig struct {
	ServeAddr string `json:"serve_addr" yaml:"serve_addr"`
}

type Config struct {
	MConfig        ManagerConfig             `json:"manager_config" yaml:"manager_config"`
	ProcessConfigs map[string]*ProcessConfig `json:"process_configs" yaml:"process_configs"`
}

func LoadConfig(filePath string) (*Config, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	datas, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}
	cfg := &Config{}
	err = yaml.Unmarshal(datas, cfg)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}
